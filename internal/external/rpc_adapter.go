package external

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
)

type RPCAdapter struct{}

type RPCDispatchEnvelope struct {
	Request DispatchRequest                      `json:"request"`
	Auth    config.ExternalIntegrationAuthConfig `json:"auth,omitempty"`
	Headers map[string]string                    `json:"headers,omitempty"`
}

func (a *RPCAdapter) Dispatch(ctx context.Context, integration config.ExternalIntegrationConfig, req DispatchRequest) (*DispatchResult, error) {
	baseURL := strings.TrimSpace(integration.BaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("rpc adapter base_url required")
	}
	network, address, err := parseRPCBaseURL(baseURL)
	if err != nil {
		return nil, err
	}
	method := strings.TrimSpace(integration.Path)
	if method == "" {
		method = "Runtime.Dispatch"
	}
	timeout := integration.TimeoutMs
	if timeout <= 0 {
		timeout = 5000
	}
	dialer := net.Dialer{Timeout: time.Duration(timeout) * time.Millisecond}
	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("dial rpc endpoint: %w", err)
	}
	deadline := time.Now().Add(time.Duration(timeout) * time.Millisecond)
	_ = conn.SetDeadline(deadline)
	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))
	defer client.Close()
	var result DispatchResult
	err = client.Call(method, RPCDispatchEnvelope{Request: req, Auth: integration.Auth, Headers: integration.Headers}, &result)
	if err != nil {
		return nil, fmt.Errorf("call rpc dispatch method: %w", err)
	}
	if strings.TrimSpace(result.Transport) == "" {
		result.Transport = strings.TrimSpace(req.PrimaryTransport)
	}
	if result.Status == 0 {
		result.Status = 200
	}
	if result.Status >= 400 {
		return &result, fmt.Errorf("rpc dispatch returned status %d", result.Status)
	}
	return &result, nil
}

func parseRPCBaseURL(baseURL string) (string, string, error) {
	parts := strings.SplitN(baseURL, "://", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("rpc adapter base_url must use tcp:// or unix://")
	}
	network := strings.ToLower(strings.TrimSpace(parts[0]))
	address := strings.TrimSpace(parts[1])
	if address == "" {
		return "", "", fmt.Errorf("rpc adapter address required")
	}
	switch network {
	case "tcp", "unix":
		return network, address, nil
	default:
		return "", "", fmt.Errorf("rpc adapter unsupported network %q", network)
	}
}
