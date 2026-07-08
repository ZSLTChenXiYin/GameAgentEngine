package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetDebugTracesUsesDebugEndpoint(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"traces": []map[string]any{{
				"id":                       "trace-1",
				"world_id":                 "world-1",
				"task_type":                "npc_dialogue",
				"configured_pipeline_mode": "full",
				"effective_pipeline_mode":  "polling",
				"max_analysis_rounds":      4,
				"rounds_used":              2,
			}},
			"count": 1,
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	payload, err := client.GetDebugTraces("world-1", 10)
	if err != nil {
		t.Fatalf("get debug traces: %v", err)
	}
	if payload.Count != 1 || len(payload.Traces) != 1 {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	if payload.Traces[0].ID != "trace-1" {
		t.Fatalf("unexpected trace: %#v", payload.Traces[0])
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/debug/traces" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery != "limit=10&world_id=world-1" && gotQuery != "world_id=world-1&limit=10" {
		t.Fatalf("unexpected query: %s", gotQuery)
	}
}
