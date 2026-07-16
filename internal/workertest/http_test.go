package workertest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientEngineJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "dev-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer server.Close()
	client := &Client{BaseURL: server.URL, APIKey: "dev-key"}
	var resp map[string]any
	if err := client.EngineJSON("GET", "/health", nil, nil, &resp); err != nil {
		t.Fatalf("EngineJSON returned error: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestWaitHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	if err := WaitHealthy(server.URL, time.Second); err != nil {
		t.Fatalf("WaitHealthy returned error: %v", err)
	}
}
