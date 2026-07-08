package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetStateComponentsUsesWorldStateEndpoint(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"world_id": "world-1",
			"components": []map[string]any{{
				"component_type": "world_state",
				"data":           map[string]any{"summary": "vault breach"},
			}},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.GetStateComponents("world-1")
	if err != nil {
		t.Fatalf("get state components: %v", err)
	}
	if gotPath != "/api/v1/worlds/world-1/state-components" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if result.WorldID != "world-1" || len(result.Components) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	data, ok := result.Components[0].Data.(map[string]any)
	if !ok || data["summary"] != "vault breach" {
		t.Fatalf("unexpected state payload: %#v", result.Components[0].Data)
	}
}

func TestPutStateComponentSendsStructuredPayload(t *testing.T) {
	var gotPath string
	var gotMethod string
	var payload map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"world_id": "world-1",
			"state_component": map[string]any{
				"component_type": "tick_policy",
				"data":           payload,
			},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.PutStateComponent("world-1", "tick_policy", map[string]any{
		"continuity_rules": []string{"Do not erase established underground structures."},
	})
	if err != nil {
		t.Fatalf("put state component: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/worlds/world-1/state-components/tick_policy" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if result.StateComponent.ComponentType != "tick_policy" {
		t.Fatalf("unexpected response: %#v", result)
	}
	items, ok := payload["continuity_rules"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("unexpected request payload: %#v", payload)
	}
}

func TestGetTimelinesUsesTimelinesEndpoint(t *testing.T) {
	var gotPath string
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"world_id": "world-1",
			"timelines": []map[string]any{{
				"tick_number":    7,
				"tick_type":      "daily",
				"advanced_ticks": 3,
				"data":           map[string]any{"future_outline": "watch the ridge", "advanced_ticks": 3},
				"timeline":       map[string]any{"id": "tick-7", "world_id": "world-1", "tick_number": 7, "tick_type": "daily", "created_at": "2026-01-01T00:00:00Z"},
			}},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.GetTimelines("world-1", 15)
	if err != nil {
		t.Fatalf("get timelines: %v", err)
	}
	if gotPath != "/api/v1/worlds/world-1/timelines" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery != "limit=15" {
		t.Fatalf("unexpected query: %s", gotQuery)
	}
	if len(result.Timelines) != 1 || result.Timelines[0].TickNumber != 7 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.Timelines[0].AdvancedTicks != 3 {
		t.Fatalf("expected advanced_ticks=3, got %#v", result.Timelines[0])
	}
}
