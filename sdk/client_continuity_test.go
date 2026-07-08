package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetContinuityBundleAggregatesTimelineStateLogsAndTraces(t *testing.T) {
	requested := map[string]int{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested[r.URL.Path]++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/worlds/world-1/timelines/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"world_id": "world-1",
				"timeline": map[string]any{
					"tick_number": 7,
					"tick_type": "daily",
					"timeline": map[string]any{"id": "tick-7", "world_id": "world-1", "tick_number": 7, "tick_type": "daily", "created_at": "2026-01-01T00:00:00Z"},
				},
			})
		case "/api/v1/worlds/world-1/state-components":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"world_id": "world-1",
				"components": []map[string]any{{"component_type": "world_state", "data": map[string]any{"summary": "vault breach"}}},
			})
		case "/api/v1/logs":
			if r.URL.Query().Get("task_type") != "world_tick" {
				t.Fatalf("expected task_type world_tick, got %q", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{{"id": "log-1", "task_type": "world_tick"}})
		case "/debug/traces":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"traces": []map[string]any{{"id": "trace-1", "world_id": "world-1", "task_type": "world_tick"}},
				"count": 1,
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	bundle, err := client.GetContinuityBundle("world-1", nil)
	if err != nil {
		t.Fatalf("get continuity bundle: %v", err)
	}
	if bundle.WorldID != "world-1" {
		t.Fatalf("unexpected world id: %#v", bundle)
	}
	if bundle.LatestTimeline == nil || bundle.LatestTimeline.TickNumber != 7 {
		t.Fatalf("unexpected timeline: %#v", bundle.LatestTimeline)
	}
	if len(bundle.StateComponents) != 1 || bundle.StateComponents[0].ComponentType != "world_state" {
		t.Fatalf("unexpected state components: %#v", bundle.StateComponents)
	}
	if len(bundle.Logs) != 1 || bundle.Logs[0].ID != "log-1" {
		t.Fatalf("unexpected logs: %#v", bundle.Logs)
	}
	if len(bundle.Traces) != 1 || bundle.Traces[0].ID != "trace-1" {
		t.Fatalf("unexpected traces: %#v", bundle.Traces)
	}
	for _, path := range []string{"/api/v1/worlds/world-1/timelines/latest", "/api/v1/worlds/world-1/state-components", "/api/v1/logs", "/debug/traces"} {
		if requested[path] != 1 {
			t.Fatalf("expected one request to %s, got %d", path, requested[path])
		}
	}
}

func TestGetContinuityBundleRespectsIncludeFlags(t *testing.T) {
	requested := map[string]int{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested[r.URL.Path]++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/worlds/world-1/timelines/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{"world_id": "world-1", "timeline": map[string]any{"tick_number": 1, "tick_type": "manual", "timeline": map[string]any{"id": "tick-1", "world_id": "world-1", "tick_number": 1, "tick_type": "manual", "created_at": "2026-01-01T00:00:00Z"}}})
		case "/api/v1/worlds/world-1/state-components":
			_ = json.NewEncoder(w).Encode(map[string]any{"world_id": "world-1", "components": []map[string]any{}})
		case "/api/v1/logs":
			_ = json.NewEncoder(w).Encode([]map[string]any{})
		case "/debug/traces":
			_ = json.NewEncoder(w).Encode(map[string]any{"traces": []map[string]any{}, "count": 0})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	_, err := client.GetContinuityBundle("world-1", &ContinuityBundleOptions{SkipLogs: true, SkipTraces: true})
	if err != nil {
		t.Fatalf("get continuity bundle: %v", err)
	}
	if requested["/api/v1/logs"] != 0 || requested["/debug/traces"] != 0 {
		t.Fatalf("expected logs and traces to be skipped, got %#v", requested)
	}
	if requested["/api/v1/worlds/world-1/timelines/latest"] != 1 || requested["/api/v1/worlds/world-1/state-components"] != 1 {
		t.Fatalf("expected timeline/state requests, got %#v", requested)
	}
}
