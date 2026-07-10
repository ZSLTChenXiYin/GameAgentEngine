package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
)

type HTTPAdapter struct{}

func (a *HTTPAdapter) Dispatch(ctx context.Context, integration config.ExternalIntegrationConfig, req DispatchRequest) (*DispatchResult, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(integration.BaseURL), "/")
	path := strings.TrimSpace(integration.Path)
	if baseURL == "" {
		return nil, fmt.Errorf("http adapter base_url required")
	}
	if path == "" {
		path = "/api/v1/runtime/dispatch"
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal dispatch request: %w", err)
	}
	timeout := integration.TimeoutMs
	if timeout <= 0 {
		timeout = 5000
	}
	client := &http.Client{Timeout: time.Duration(timeout) * time.Millisecond}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build dispatch request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	for k, v := range integration.Headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		httpReq.Header.Set(k, v)
	}
	switch strings.ToLower(strings.TrimSpace(integration.Auth.Mode)) {
	case "bearer":
		if token := strings.TrimSpace(integration.Auth.Token); token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+token)
		}
	case "header":
		headerName := strings.TrimSpace(integration.Auth.HeaderName)
		if headerName == "" {
			headerName = "X-External-Auth"
		}
		if token := strings.TrimSpace(integration.Auth.Token); token != "" {
			httpReq.Header.Set(headerName, token)
		}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("dispatch http request: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	result := &DispatchResult{
		Transport: strings.TrimSpace(req.PrimaryTransport),
		Status:    resp.StatusCode,
		Body:      string(respBody),
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, fmt.Errorf("dispatch http request returned status %d", resp.StatusCode)
	}
	return result, nil
}
