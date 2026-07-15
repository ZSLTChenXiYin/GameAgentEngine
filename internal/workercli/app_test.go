package workercli

import "testing"

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
