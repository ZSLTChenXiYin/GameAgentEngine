package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
)

func TestHTTPAdapterDispatchPostsJSONWithBearerAuth(t *testing.T) {
	var gotMethod string
	var gotAuth string
	var gotContentType string
	var gotBody DispatchRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	adapter := &HTTPAdapter{}
	result, err := adapter.Dispatch(context.Background(), config.ExternalIntegrationConfig{
		Type:    "http_adapter",
		BaseURL: server.URL,
		Path:    "/dispatch",
		Auth: config.ExternalIntegrationAuthConfig{
			Mode:  "bearer",
			Token: "secret-token",
		},
	}, DispatchRequest{TaskID: "task-1", Category: "external_query", InterfaceName: "game_client_request_data"})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotAuth != "Bearer secret-token" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Fatalf("unexpected content type: %q", gotContentType)
	}
	if gotBody.TaskID != "task-1" {
		t.Fatalf("unexpected task id in request body: %+v", gotBody)
	}
	if result == nil || result.Status != http.StatusOK {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}

func TestDispatcherDispatchUsesConfiguredIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`ok`))
	}))
	defer server.Close()

	previous := config.Global.ExternalIntegrations
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_http": {
			Type:    "http_adapter",
			BaseURL: server.URL,
			Path:    "/dispatch",
		},
	}
	defer func() { config.Global.ExternalIntegrations = previous }()

	dispatcher := NewDispatcher()
	result, err := dispatcher.Dispatch(context.Background(), Route{DeliveryMode: "push", PrimaryTransport: "game_http"}, DispatchRequest{TaskID: "task-1"})
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if result == nil || result.Transport != "game_http" {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}
