package external

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
)

type Route struct {
	DeliveryMode     string
	PrimaryTransport string
	Consumer         string
	TimeoutMs        int
}

type DispatchRequest struct {
	TaskID            string         `json:"task_id"`
	IdempotencyKey    string         `json:"idempotency_key,omitempty"`
	Category          string         `json:"category"`
	InterfaceName     string         `json:"interface_name"`
	DeliveryMode      string         `json:"delivery_mode"`
	PrimaryTransport  string         `json:"primary_transport,omitempty"`
	Consumer          string         `json:"consumer,omitempty"`
	WorldID           string         `json:"world_id,omitempty"`
	NodeID            string         `json:"node_id,omitempty"`
	RequestID         string         `json:"request_id,omitempty"`
	CallbackID        string         `json:"callback_id,omitempty"`
	ResumeExecutionID string         `json:"resume_execution_id,omitempty"`
	ResumePolicy      string         `json:"resume_policy,omitempty"`
	Payload           map[string]any `json:"payload,omitempty"`
	RawPayloadJSON    string         `json:"raw_payload_json,omitempty"`
}

type DispatchResult struct {
	Transport string         `json:"transport"`
	Status    int            `json:"status,omitempty"`
	Body      string         `json:"body,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Adapter interface {
	Dispatch(ctx context.Context, integration config.ExternalIntegrationConfig, req DispatchRequest) (*DispatchResult, error)
}

type Dispatcher struct {
	adapters map[string]Adapter
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{adapters: map[string]Adapter{
		"http_adapter":      &HTTPAdapter{},
		"rpc_adapter":       &RPCAdapter{},
		"websocket_adapter": &WebSocketAdapter{},
	}}
}

func (d *Dispatcher) Dispatch(ctx context.Context, route Route, req DispatchRequest) (*DispatchResult, error) {
	transport := strings.TrimSpace(route.PrimaryTransport)
	if transport == "" {
		return nil, fmt.Errorf("primary transport required for push dispatch")
	}
	integration, ok := config.Global.ExternalIntegrations[transport]
	if !ok {
		return nil, fmt.Errorf("external integration %q not configured", transport)
	}
	adapterType := strings.TrimSpace(integration.Type)
	if adapterType == "" {
		return nil, fmt.Errorf("external integration %q missing type", transport)
	}
	adapter, ok := d.adapters[adapterType]
	if !ok {
		return nil, fmt.Errorf("external integration type %q not supported", adapterType)
	}
	maxAttempts := integration.RetryMaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	backoffMs := integration.RetryBackoffMs
	if backoffMs <= 0 {
		backoffMs = 100
	}
	var lastResult *DispatchResult
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := adapter.Dispatch(ctx, integration, req)
		lastResult = result
		lastErr = err
		if result != nil {
			if result.Metadata == nil {
				result.Metadata = map[string]any{}
			}
			result.Metadata["dispatch_attempt"] = attempt
			result.Metadata["dispatch_max_attempts"] = maxAttempts
		}
		if err == nil {
			if result != nil && strings.TrimSpace(result.Transport) == "" {
				result.Transport = transport
			}
			return result, nil
		}
		if attempt >= maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return lastResult, ctx.Err()
		case <-time.After(time.Duration(backoffMs) * time.Millisecond):
		}
	}
	if lastResult != nil && strings.TrimSpace(lastResult.Transport) == "" {
		lastResult.Transport = transport
	}
	return lastResult, lastErr
}

func NormalizeRoute(deliveryMode string, primaryTransport string, consumer string, timeoutMs int) Route {
	mode := strings.ToLower(strings.TrimSpace(deliveryMode))
	switch mode {
	case "push", "pull", "hybrid":
	default:
		if strings.TrimSpace(primaryTransport) != "" {
			mode = "push"
		} else {
			mode = "pull"
		}
	}
	return Route{
		DeliveryMode:     mode,
		PrimaryTransport: strings.TrimSpace(primaryTransport),
		Consumer:         strings.TrimSpace(consumer),
		TimeoutMs:        timeoutMs,
	}
}

func (r Route) ShouldPush() bool {
	return r.DeliveryMode == "push" || r.DeliveryMode == "hybrid"
}

func (r Route) ShouldQueuePullTask() bool {
	return r.DeliveryMode == "pull" || r.DeliveryMode == "hybrid"
}

func (r Route) IsStrictPush() bool {
	return r.DeliveryMode == "push"
}
