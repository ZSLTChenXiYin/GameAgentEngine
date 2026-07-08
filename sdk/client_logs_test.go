package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetLogsByQueryUsesStructuredFilterQuery(t *testing.T) {
	var gotPath string
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id": "log-1",
		}})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	logs, err := client.GetLogsByQuery(InferenceLogQuery{
		WorldID:       "world-1",
		NodeID:        "node-1",
		TaskType:      "world_tick",
		Category:      "pipeline",
		EventName:     "raw_llm_response_received",
		ExecutionMode: "debug",
		RequestID:     "req-1",
		Round:         2,
		Limit:         10,
		Offset:        5,
	})
	if err != nil {
		t.Fatalf("get logs by query: %v", err)
	}
	if len(logs) != 1 || logs[0].ID != "log-1" {
		t.Fatalf("unexpected logs: %#v", logs)
	}
	if gotPath != "/api/v1/logs" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	for _, want := range []string{"world_id=world-1", "node_id=node-1", "task_type=world_tick", "category=pipeline", "event_name=raw_llm_response_received", "execution_mode=debug", "request_id=req-1", "round=2", "limit=10", "offset=5"} {
		if !containsQueryPair(gotQuery, want) {
			t.Fatalf("missing %q in query %q", want, gotQuery)
		}
	}
}

func containsQueryPair(query, expected string) bool {
	parts := splitQuery(query)
	for _, part := range parts {
		if part == expected {
			return true
		}
	}
	return false
}

func splitQuery(query string) []string {
	if query == "" {
		return nil
	}
	return strings.Split(query, "&")
}
