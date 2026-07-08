package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdvanceTickWithOptionsSendsRequestedTicks(t *testing.T) {
	var payload map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/worlds/world-1/ticks/advance" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tick":   map[string]any{"id": "tick-1", "world_id": "world-1", "tick_number": 1, "tick_type": "manual", "game_time": "day-1", "created_at": "2026-01-01T00:00:00Z"},
			"invoke": map[string]any{"reply": "ok"},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dev-key")
	requestedTicks := 4
	autonomousLimit := 2
	result, err := client.AdvanceTickWithOptions("world-1", "manual", "day-1", &requestedTicks, &autonomousLimit)
	if err != nil {
		t.Fatalf("advance tick: %v", err)
	}
	if result.Tick == nil || result.Tick.TickNumber != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if payload["requested_ticks"] != float64(4) {
		t.Fatalf("expected requested_ticks=4, got %#v", payload)
	}
	if payload["autonomous_limit"] != float64(2) {
		t.Fatalf("expected autonomous_limit=2, got %#v", payload)
	}
}
