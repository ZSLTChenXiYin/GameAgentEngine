package api

import (
	"sync"
	"time"
	"strings"
	"fmt"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// writeJSON 统一写出 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("write json error: %v", err)
	}
}

// errorJSON 统一写出错误响应体。
func errorJSON(w http.ResponseWriter, status int, msg string) {
	errorJSONCode(w, status, defaultErrorCode(status), msg)
}

// errorJSONCode 统一写出包含稳定错误码的错误响应体。
func errorJSONCode(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{"error": msg, "code": code})
}

func defaultErrorCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	default:
		return "internal_error"
	}
}

// RequestAuth 校验请求携带的 API key 或外部执行面专用 token。
func RequestAuth(next http.Handler, auth config.AuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/api/v1/actions/callback" && auth.CallbackToken != "" {
			if r.Header.Get("X-Callback-Token") == auth.CallbackToken {
				next.ServeHTTP(w, r)
				return
			}
		}
		if len(r.URL.Path) >= len("/api/v1/runtime/tasks") && r.URL.Path[:len("/api/v1/runtime/tasks")] == "/api/v1/runtime/tasks" && auth.RuntimeTaskToken != "" {
			if r.Header.Get("X-Runtime-Task-Token") == auth.RuntimeTaskToken {
				next.ServeHTTP(w, r)
				return
			}
		}

		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}
		if key != auth.APIKey {
			errorJSONCode(w, http.StatusUnauthorized, "invalid_api_key", "invalid api key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// APIKeyAuth 保留旧接口，内部转调新的 RequestAuth。
func APIKeyAuth(next http.Handler, validKey string) http.Handler {
	return RequestAuth(next, config.AuthConfig{APIKey: validKey})
}

// CORSMiddleware 为前端工具补充跨域响应头。
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key, Idempotency-Key, X-Callback-Token, X-Callback-Request-Id, X-Runtime-Task-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// captureWriter 用于捕获响应状态码和 body，供幂等中间件缓存。
type captureWriter struct {
	http.ResponseWriter
	statusCode int
	buf        *bytes.Buffer
}

func newCaptureWriter(w http.ResponseWriter) *captureWriter {
	return &captureWriter{ResponseWriter: w, statusCode: http.StatusOK, buf: &bytes.Buffer{}}
}

func (w *captureWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *captureWriter) Write(data []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	w.buf.Write(data)
	return w.ResponseWriter.Write(data)
}

func requestFingerprint(r *http.Request, body []byte) string {
	bodyHash := sha256.Sum256(body)
	raw := r.Method + "\n" + r.URL.Path + "\n" + r.URL.RawQuery + "\n" + hex.EncodeToString(bodyHash[:])
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// IdempotencyMiddleware 检查 Idempotency-Key 请求头，对已处理的请求直接返回缓存结果。
// 仅对 POST/PUT/DELETE 方法生效；成功后缓存响应 body。
func IdempotencyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			next(w, r)
			return
		}
		if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodDelete {
			next(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			errorJSON(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		fingerprint := requestFingerprint(r, body)

		// 检查是否已处理
		if result, err := store.GetIdempotencyResult(key); err == nil {
			if result.Fingerprint != fingerprint {
				errorJSONCode(w, http.StatusConflict, "idempotency_key_conflict", "idempotency key reused with different request payload")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotency-Replayed", "true")
			w.WriteHeader(result.StatusCode)
			w.Write([]byte(result.Result))
			return
		} else if !store.IsRecordNotFound(err) {
			log.Printf("load idempotency result: %v", err)
		}

		// 捕获响应
		cw := newCaptureWriter(w)
		next(cw, r)

		// 成功的请求才缓存
		if cw.statusCode >= 200 && cw.statusCode < 300 {
			if err := store.SetIdempotencyResult(key, fingerprint, cw.statusCode, cw.buf.String()); err != nil {
				log.Printf("save idempotency result: %v", err)
			}
		}
	}
}

// CallbackReplayMiddleware protects callback requests from duplicate replay when request IDs are supplied.
func CallbackReplayMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Callback-Request-Id")
		if requestID == "" {
			if config.Global.Auth.CallbackRequireRequestID {
				errorJSONCode(w, http.StatusBadRequest, "invalid_callback_request_id", "X-Callback-Request-Id required")
				return
			}
			next(w, r)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			errorJSON(w, http.StatusBadRequest, "failed to read request body")
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		fingerprint := requestFingerprint(r, body)
		key := "callback:" + requestID
		record, created, err := store.AcquireIdempotencyKey(key, fingerprint)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		if !created {
			if record.Fingerprint != fingerprint {
				errorJSONCode(w, http.StatusConflict, "callback_request_conflict", "callback request id reused with different payload")
				return
			}
			if record.StatusCode > 0 {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Callback-Replayed", "true")
				w.WriteHeader(record.StatusCode)
				w.Write([]byte(record.Result))
				return
			}
			errorJSONCode(w, http.StatusConflict, "callback_request_in_progress", "callback request id is already being processed")
			return
		}
		cw := newCaptureWriter(w)
		next(cw, r)
		if cw.statusCode >= 200 && cw.statusCode < 300 {
			if err := store.SetIdempotencyResult(key, fingerprint, cw.statusCode, cw.buf.String()); err != nil {
				log.Printf("save callback replay result: %v", err)
			}
		}
	}
}

// ValidateWorldAccess checks whether the given world_id exists in the store.
// Returns an error if the world is not found or access is denied.
// This enforces basic world isolation: every world-scoped request must pass
// through a valid world ID.

// ============ Rate Limiter (P1) ============
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64
	burst    int
	stopCh   chan struct{}
}

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		stopCh:   make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, ok := rl.visitors[ip]
	if !ok {
		v = &visitor{tokens: float64(rl.burst), lastSeen: time.Now()}
		rl.visitors[ip] = v
	}
	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > float64(rl.burst) { v.tokens = float64(rl.burst) }
	v.lastSeen = now
	if v.tokens < 1 { return false }
	v.tokens--
	return true
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

func (rl *RateLimiter) Stop() { close(rl.stopCh) }

func RateLimitMiddleware(rl *RateLimiter) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			if !rl.Allow(ip) {
				w.Header().Set("Retry-After", "1")
				errorJSONCode(w, 429, "rate_limited", "too many requests")
				return
			}
			next(w, r)
		}
	}
}

func ValidateWorldAccess(worldID string, r *http.Request) error {
	if strings.TrimSpace(worldID) == "" {
		return fmt.Errorf("world_id required")
	}
	// Basic isolation: verify world node exists
	node, err := store.GetNode(worldID)
	if err != nil {
		return fmt.Errorf("world %s not found: %w", worldID, err)
	}
	if node == nil {
		return fmt.Errorf("world %s not found", worldID)
	}
	// World isolation: the node must be of type "world"
	if node.NodeType != "world" {
		return fmt.Errorf("node %s is not a world", worldID)
	}
	return nil
}
