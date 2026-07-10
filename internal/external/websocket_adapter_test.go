package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
)

func TestWebSocketAdapterDispatchWritesJSONAndReadsResponse(t *testing.T) {
	var gotAuth string
	var gotBody DispatchRequest
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade websocket: %v", err)
		}
		defer conn.Close()
		if err := conn.ReadJSON(&gotBody); err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := conn.WriteJSON(map[string]any{"status": 202, "accepted": true}); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	adapter := &WebSocketAdapter{}
	result, err := adapter.Dispatch(context.Background(), config.ExternalIntegrationConfig{
		Type:    "websocket_adapter",
		BaseURL: wsURL,
		Path:    "/dispatch",
		Auth: config.ExternalIntegrationAuthConfig{
			Mode:  "bearer",
			Token: "ws-secret",
		},
	}, DispatchRequest{TaskID: "task-ws-1", Category: "external_query", InterfaceName: "game_client_request_data"})
	if err != nil {
		t.Fatalf("dispatch websocket: %v", err)
	}
	if gotAuth != "Bearer ws-secret" {
		t.Fatalf("unexpected auth header: %q", gotAuth)
	}
	if gotBody.TaskID != "task-ws-1" {
		t.Fatalf("unexpected request body: %+v", gotBody)
	}
	if result == nil || result.Status != 202 {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
	if result.Metadata == nil || result.Metadata["accepted"] != true {
		t.Fatalf("expected metadata to contain accepted=true, got %+v", result)
	}
}

func TestDispatcherDispatchUsesConfiguredWebSocketIntegration(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade websocket: %v", err)
		}
		defer conn.Close()
		var body map[string]any
		if err := conn.ReadJSON(&body); err != nil {
			t.Fatalf("read request body: %v", err)
		}
		resp, _ := json.Marshal(map[string]any{"status": 200, "task_id": body["task_id"]})
		if err := conn.WriteMessage(websocket.TextMessage, resp); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	previous := config.Global.ExternalIntegrations
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_ws": {
			Type:    "websocket_adapter",
			BaseURL: wsURL,
			Path:    "/dispatch",
		},
	}
	defer func() { config.Global.ExternalIntegrations = previous }()

	dispatcher := NewDispatcher()
	result, err := dispatcher.Dispatch(context.Background(), Route{DeliveryMode: "push", PrimaryTransport: "game_ws"}, DispatchRequest{TaskID: "task-1"})
	if err != nil {
		t.Fatalf("dispatch websocket: %v", err)
	}
	if result == nil || result.Transport != "game_ws" {
		t.Fatalf("unexpected dispatch result: %+v", result)
	}
}
