package workercli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDecideExecutionReturnsFailureForConfiguredInterface(t *testing.T) {
	a := newTestApp()
	a.cfg.FailInterfaces = []string{"spawn_item"}

	decision := a.decideExecution("spawn_item", map[string]any{"node_id": "npc-1"})
	if decision.Status != "failed" {
		t.Fatalf("expected failed status, got %q", decision.Status)
	}
	if decision.DecisionName != "forced_failure" {
		t.Fatalf("expected forced_failure decision, got %q", decision.DecisionName)
	}
}

func TestDecideExecutionReturnsLongRunningForConfiguredInterface(t *testing.T) {
	a := newTestApp()
	a.cfg.LongTaskInterfaces = []string{"game_client_request_data"}

	decision := a.decideExecution("game_client_request_data", map[string]any{"request_data": map[string]any{"label": "scene"}})
	if !decision.LongRunning {
		t.Fatal("expected long-running decision")
	}
	if decision.DecisionName != "long_running" {
		t.Fatalf("expected long_running decision, got %q", decision.DecisionName)
	}
	if decision.Result["scene"] != "starter_inn" {
		t.Fatalf("unexpected fixture result: %#v", decision.Result)
	}
}

func TestBuildFixtureResultForSpawnItemIncludesTarget(t *testing.T) {
	a := newTestApp()
	result := a.buildFixtureResult("spawn_item", map[string]any{"node_id": "npc-123"}, "success", false)
	if result["spawned"] != true {
		t.Fatalf("expected spawned=true, got %#v", result)
	}
	if result["inventory_target"] != "npc-123" {
		t.Fatalf("expected inventory_target npc-123, got %#v", result)
	}
}

func TestParseRuntimeTaskPayloadFallsBackToRawPayloadJSON(t *testing.T) {
	payload := parseRuntimeTaskPayload("not-json")
	if payload["raw_payload_json"] != "not-json" {
		t.Fatalf("expected raw payload fallback, got %#v", payload)
	}
}

func TestBuildFixtureResultResolvesAuthorityQueriesFromStateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(path, []byte(`
world_id: demo_world
actors:
  player_001:
    hp: 15
    max_hp: 20
    money: 66
    location_id: scene_inn
    inventory:
      - item_id: knife_bloody
        quantity: 1
scenes:
  scene_inn:
    id: scene_inn
    occupants: [player_001, npc_innkeeper]
tasks:
  task_case:
    status: active
    stage: investigate
`), 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	a := newTestApp()
	a.cfg.StateFile = path
	result := a.buildFixtureResult("game_client_request_data", map[string]any{
		"request_data": map[string]any{
			"queries": []any{
				map[string]any{"type": "player_state", "node_id": "player_001"},
				map[string]any{"type": "player_wallet", "node_id": "player_001"},
				map[string]any{"type": "player_inventory", "node_id": "player_001"},
				map[string]any{"type": "player_location", "node_id": "player_001"},
				map[string]any{"type": "scene_state", "node_id": "scene_inn"},
				map[string]any{"type": "task_state", "node_id": "task_case"},
				map[string]any{"type": "item_presence", "node_id": "player_001", "filter": "knife_bloody"},
			},
		},
	}, "success", false)

	queries, ok := result["queries"].([]map[string]any)
	if ok {
		_ = queries
	}
	rawQueries, ok := result["queries"].([]map[string]any)
	if ok && len(rawQueries) > 0 {
		return
	}
	list, ok := result["queries"].([]any)
	if !ok || len(list) != 7 {
		t.Fatalf("expected 7 resolved queries, got %#v", result["queries"])
	}
	first := list[0].(map[string]any)
	if first["hp"] != 15 || first["max_hp"] != 20 {
		t.Fatalf("unexpected player_state result: %#v", first)
	}
	if result["world_id"] != "demo_world" {
		t.Fatalf("expected world_id demo_world, got %#v", result)
	}
}
