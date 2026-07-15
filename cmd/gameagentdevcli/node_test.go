package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLegacyNodesCommandSharesListFlags(t *testing.T) {
	previousServerURL := serverURL
	previousAPIKey := apiKey
	defer func() {
		serverURL = previousServerURL
		apiKey = previousAPIKey
	}()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/nodes" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("world_id") != "world-1" {
			t.Fatalf("expected world_id query, got %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":        "node-1",
			"world_id":  "world-1",
			"name":      "Guard",
			"node_type": "npc",
		}})
	}))
	defer ts.Close()

	serverURL = ts.URL
	apiKey = "test-key"

	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = stdout
	}()

	nodeLegacyListCmd.SetArgs([]string{"--world", "world-1"})
	defer nodeLegacyListCmd.SetArgs(nil)
	if err := nodeLegacyListCmd.Execute(); err != nil {
		t.Fatalf("execute legacy nodes command: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(bytes.TrimSpace(output)), `"id": "node-1"`) {
		t.Fatalf("unexpected output: %s", output)
	}
}
