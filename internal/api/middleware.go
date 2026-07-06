package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

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

// APIKeyAuth 校验请求携带的 API Key。
// 当前仅健康检查接口跳过认证，其余能力统一按 API 访问控制。
func APIKeyAuth(next http.Handler, validKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		key := r.Header.Get("X-API-Key")
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}
		if key != validKey {
			errorJSONCode(w, http.StatusUnauthorized, "invalid_api_key", "invalid api key")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware 为前端工具补充跨域响应头。
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
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
