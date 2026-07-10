package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
)

type WebSocketAdapter struct{}

func (a *WebSocketAdapter) Dispatch(ctx context.Context, integration config.ExternalIntegrationConfig, req DispatchRequest) (*DispatchResult, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(integration.BaseURL), "/")
	path := strings.TrimSpace(integration.Path)
	if baseURL == "" {
		return nil, fmt.Errorf("websocket adapter base_url required")
	}
	if !strings.HasPrefix(baseURL, "ws://") && !strings.HasPrefix(baseURL, "wss://") {
		return nil, fmt.Errorf("websocket adapter base_url must use ws:// or wss://")
	}
	if path == "" {
		path = "/ws/runtime/dispatch"
	}
	timeout := integration.TimeoutMs
	if timeout <= 0 {
		timeout = 5000
	}
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	for k, v := range integration.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		headers.Set(k, v)
	}
	switch strings.ToLower(strings.TrimSpace(integration.Auth.Mode)) {
	case "bearer":
		if token := strings.TrimSpace(integration.Auth.Token); token != "" {
			headers.Set("Authorization", "Bearer "+token)
		}
	case "header":
		headerName := strings.TrimSpace(integration.Auth.HeaderName)
		if headerName == "" {
			headerName = "X-External-Auth"
		}
		if token := strings.TrimSpace(integration.Auth.Token); token != "" {
			headers.Set(headerName, token)
		}
	}
	if headerName := strings.TrimSpace(integration.IdempotencyHeader); headerName != "" && strings.TrimSpace(req.IdempotencyKey) != "" {
		headers.Set(headerName, req.IdempotencyKey)
	}
	dialer := websocket.Dialer{HandshakeTimeout: time.Duration(timeout) * time.Millisecond}
	conn, _, err := dialer.DialContext(ctx, baseURL+path, headers)
	if err != nil {
		return nil, fmt.Errorf("dispatch websocket request: %w", err)
	}
	defer conn.Close()
	deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
	_ = conn.SetWriteDeadline(deadline)
	if err := conn.WriteJSON(req); err != nil {
		return nil, fmt.Errorf("write websocket dispatch request: %w", err)
	}
	_ = conn.SetReadDeadline(deadline)
	_, msg, err := conn.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("read websocket dispatch response: %w", err)
	}
	result := &DispatchResult{
		Transport: strings.TrimSpace(req.PrimaryTransport),
		Status:    http.StatusOK,
		Body:      string(msg),
	}
	var payload map[string]any
	if err := json.Unmarshal(msg, &payload); err == nil {
		result.Metadata = payload
		if status, ok := parseDispatchStatus(payload["status"]); ok {
			result.Status = status
		}
	}
	if result.Status >= 400 {
		return result, fmt.Errorf("dispatch websocket request returned status %d", result.Status)
	}
	return result, nil
}

func parseDispatchStatus(value any) (int, bool) {
	switch v := value.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}
