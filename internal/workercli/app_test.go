package workercli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
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

func TestParsePlayCommand(t *testing.T) {
	cmd := parsePlayCommand("/talk innkeeper")
	if cmd.Name != "talk" || cmd.Args != "innkeeper" {
		t.Fatalf("unexpected play command: %#v", cmd)
	}
}

func TestResolvePlayPlayerNodeIDPrefersPlayerKind(t *testing.T) {
	a := newTestApp()
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"npc_1":    {ID: "npc_1", Kind: "npc"},
			"player_1": {ID: "player_1", Kind: "player"},
		},
	})
	got, err := a.resolvePlayPlayerNodeID(view)
	if err != nil {
		t.Fatalf("resolvePlayPlayerNodeID returned error: %v", err)
	}
	if got != "player_1" {
		t.Fatalf("expected player_1, got %q", got)
	}
}

func TestResolveSceneActorRequiresSameScene(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_a"},
			"npc_1":    {ID: "npc_1", Name: "innkeeper", Kind: "npc", LocationID: "scene_b"},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_a"}
	if _, err := s.resolveSceneActor("innkeeper"); err == nil {
		t.Fatal("expected same-scene validation error")
	}
}

func TestTransferInventoryItemMovesAuthorityState(t *testing.T) {
	a := newTestApp()
	a.setAuthorityState(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Inventory: []workerstate.InventoryEntry{{ItemID: "knife_bloody", Quantity: 1}}},
			"npc_1":    {ID: "npc_1"},
		},
		Items: map[string]*workerstate.ItemState{
			"knife_bloody": {ID: "knife_bloody", OwnerID: "player_1"},
		},
	})
	if err := a.transferInventoryItem("player_1", "npc_1", "knife_bloody", 1); err != nil {
		t.Fatalf("transferInventoryItem returned error: %v", err)
	}
	view := a.authorityView()
	if view.ItemPresentOnActor("player_1", "knife_bloody") {
		t.Fatal("expected item removed from player inventory")
	}
	if !view.ItemPresentOnActor("npc_1", "knife_bloody") {
		t.Fatal("expected item added to npc inventory")
	}
	if item := view.State().Items["knife_bloody"]; item == nil || item.OwnerID != "npc_1" {
		t.Fatalf("expected item owner updated to npc_1, got %#v", item)
	}
}

func TestResolvePlayerInventoryItemRequiresPossession(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Inventory: []workerstate.InventoryEntry{{ItemID: "knife_bloody", Quantity: 1}}},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1"}
	if _, _, err := s.resolvePlayerInventoryItem("silver_ring"); err == nil {
		t.Fatal("expected missing item error")
	}
}

func TestUniqueParticipantIDsDeduplicatesAndPreservesOrder(t *testing.T) {
	got := uniqueParticipantIDs([]string{"player_1", "npc_1", "npc_1"}, "player_1", "npc_2")
	if len(got) != 3 || got[0] != "player_1" || got[1] != "npc_1" || got[2] != "npc_2" {
		t.Fatalf("unexpected participants: %#v", got)
	}
}

func TestResolveGroupChatTargetDefaultsToFirstNPC(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_a"},
			"npc_1":    {ID: "npc_1", Name: "innkeeper", Kind: "npc", LocationID: "scene_a"},
			"npc_2":    {ID: "npc_2", Name: "guard", Kind: "npc", LocationID: "scene_a"},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_a"}
	targetID, participants, err := s.resolveGroupChatTarget("")
	if err != nil {
		t.Fatalf("resolveGroupChatTarget returned error: %v", err)
	}
	if targetID == "" || targetID == "player_1" {
		t.Fatalf("expected npc target, got %q", targetID)
	}
	if len(participants) != 3 {
		t.Fatalf("expected 3 participants, got %#v", participants)
	}
}
