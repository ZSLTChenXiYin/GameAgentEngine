package workercli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/playerintent"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
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
	a.cfg.LongTaskInterfaces = []string{sdk.AuthorityInterfaceGameClientRequestData}

	decision := a.decideExecution(sdk.AuthorityInterfaceGameClientRequestData, map[string]any{"request_data": map[string]any{"label": "scene"}})
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
	payload := sdk.ParseRuntimeTaskPayloadJSON("not-json")
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
	result := a.buildFixtureResult(sdk.AuthorityInterfaceGameClientRequestData, map[string]any{
		"request_data": map[string]any{
			"queries": []any{
				map[string]any{"type": sdk.AuthorityQueryPlayerState, "node_id": "player_001"},
				map[string]any{"type": sdk.AuthorityQueryPlayerWallet, "node_id": "player_001"},
				map[string]any{"type": sdk.AuthorityQueryPlayerInventory, "node_id": "player_001"},
				map[string]any{"type": sdk.AuthorityQueryPlayerLocation, "node_id": "player_001"},
				map[string]any{"type": sdk.AuthorityQuerySceneState, "node_id": "scene_inn"},
				map[string]any{"type": sdk.AuthorityQueryTaskState, "node_id": "task_case"},
				map[string]any{"type": sdk.AuthorityQueryItemPresence, "node_id": "player_001", "filter": "knife_bloody"},
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

func TestBuildFixtureResultReturnsRequestErrorForMalformedAuthorityPayload(t *testing.T) {
	a := newTestApp()
	a.setAuthorityState(&workerstate.WorldState{WorldID: "world_1"})
	result := a.buildFixtureResult(sdk.AuthorityInterfaceGameClientRequestData, map[string]any{
		"request_data": map[string]any{"queries": "bad-shape"},
	}, "success", false)
	if result["request_error"] == nil {
		t.Fatalf("expected request_error, got %#v", result)
	}
	if result["world_id"] != "world_1" {
		t.Fatalf("expected world_id world_1, got %#v", result)
	}
}

func TestParsePlayCommand(t *testing.T) {
	cmd := parsePlayCommand("/talk innkeeper")
	if cmd.Name != "talk" || cmd.Args != "innkeeper" {
		t.Fatalf("unexpected play command: %#v", cmd)
	}
}

func TestParsePlayCommandSupportsPlusPrefix(t *testing.T) {
	cmd := parsePlayCommand("/+talk innkeeper")
	if cmd.Name != "talk" || cmd.Args != "innkeeper" {
		t.Fatalf("unexpected plus-prefixed play command: %#v", cmd)
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

func TestResolveDirectDialogueTargetAutoSelectsOnlyNPC(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_a"},
			"npc_1":    {ID: "npc_1", Name: "innkeeper", Kind: "npc", LocationID: "scene_a"},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_a"}
	targetID, err := s.resolveDirectDialogueTarget()
	if err != nil {
		t.Fatalf("resolveDirectDialogueTarget returned error: %v", err)
	}
	if targetID != "npc_1" {
		t.Fatalf("expected npc_1, got %q", targetID)
	}
}

func TestResolveDirectDialogueTargetRequiresExplicitChoiceWhenMultipleNPCs(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_a"},
			"npc_1":    {ID: "npc_1", Name: "innkeeper", Kind: "npc", LocationID: "scene_a"},
			"npc_2":    {ID: "npc_2", Name: "guard", Kind: "npc", LocationID: "scene_a"},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_a"}
	_, err := s.resolveDirectDialogueTarget()
	if err == nil {
		t.Fatal("expected explicit target selection error")
	}
	if !strings.Contains(err.Error(), "multiple dialogue targets are available") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSceneDialogueTargetsAndCyclePlayTarget(t *testing.T) {
	a := newTestApp()
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_a"},
			"npc_2":    {ID: "npc_2", Name: "guard", Kind: "npc", LocationID: "scene_a"},
			"npc_1":    {ID: "npc_1", Name: "innkeeper", Kind: "npc", LocationID: "scene_a"},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_a"}
	targets := s.sceneDialogueTargets()
	if len(targets) != 2 || targets[0] != "npc_2" || targets[1] != "npc_1" {
		t.Fatalf("unexpected dialogue targets order: %#v", targets)
	}
	if err := a.cyclePlayTarget(s, 1); err != nil {
		t.Fatalf("cyclePlayTarget returned error: %v", err)
	}
	if s.currentTargetID != "npc_2" {
		t.Fatalf("expected first cycled target npc_2, got %q", s.currentTargetID)
	}
	if err := a.cyclePlayTarget(s, 1); err != nil {
		t.Fatalf("cyclePlayTarget returned error: %v", err)
	}
	if s.currentTargetID != "npc_1" {
		t.Fatalf("expected second cycled target npc_1, got %q", s.currentTargetID)
	}
	if err := a.cyclePlayTarget(s, -1); err != nil {
		t.Fatalf("cyclePlayTarget returned error: %v", err)
	}
	if s.currentTargetID != "npc_2" {
		t.Fatalf("expected cycled-back target npc_2, got %q", s.currentTargetID)
	}
}

func TestRunPlayMoveUpdatesAuthorityStateAndScene(t *testing.T) {
	a := newTestApp()
	a.setAuthorityState(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_inn"},
		},
		Scenes: map[string]*workerstate.SceneState{
			"scene_inn":    {ID: "scene_inn", Name: "Inn", Occupants: []string{"player_1"}},
			"scene_square": {ID: "scene_square", Name: "Square"},
		},
	})
	s := &playSession{view: a.authorityView(), playerNodeID: "player_1", currentSceneID: "scene_inn"}
	if err := a.runPlayMove(s, "Square"); err != nil {
		t.Fatalf("runPlayMove returned error: %v", err)
	}
	if s.currentSceneID != "scene_square" {
		t.Fatalf("expected current scene scene_square, got %q", s.currentSceneID)
	}
	view := a.authorityView()
	if locationID, ok := view.ActorLocation("player_1"); !ok || locationID != "scene_square" {
		t.Fatalf("expected player moved to scene_square, got %q ok=%v", locationID, ok)
	}
}

func TestRenderInspectionDefaultsToSceneSummary(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_inn"},
		},
		Scenes: map[string]*workerstate.SceneState{
			"scene_inn": {ID: "scene_inn", Name: "Inn", Description: "Warm light."},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_inn"}
	text, err := s.renderInspection("")
	if err != nil {
		t.Fatalf("renderInspection returned error: %v", err)
	}
	if !strings.Contains(text, "当前场景: Inn (scene_inn)") {
		t.Fatalf("unexpected inspection text: %q", text)
	}
}

func TestRenderInspectionSupportsActorAndVisibleItem(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_inn"},
			"npc_1":    {ID: "npc_1", Name: "innkeeper", Kind: "npc", LocationID: "scene_inn", Inventory: []workerstate.InventoryEntry{{ItemID: "knife_bloody", Quantity: 1}}},
		},
		Scenes: map[string]*workerstate.SceneState{
			"scene_inn": {ID: "scene_inn", Name: "Inn"},
		},
		Items: map[string]*workerstate.ItemState{
			"knife_bloody": {ID: "knife_bloody", Name: "Bloody Knife", OwnerID: "npc_1"},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_inn"}
	actorText, err := s.renderInspection("innkeeper")
	if err != nil {
		t.Fatalf("renderInspection actor returned error: %v", err)
	}
	if !strings.Contains(actorText, "角色: innkeeper (npc_1)") {
		t.Fatalf("unexpected actor inspection: %q", actorText)
	}
	itemText, err := s.renderInspection("Bloody Knife")
	if err != nil {
		t.Fatalf("renderInspection item returned error: %v", err)
	}
	if !strings.Contains(itemText, "物品: Bloody Knife (knife_bloody)") {
		t.Fatalf("unexpected item inspection: %q", itemText)
	}
}

func TestRunPlayUseItemRequiresOwnedItem(t *testing.T) {
	a := newTestApp()
	a.setAuthorityState(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_inn", Inventory: []workerstate.InventoryEntry{{ItemID: "apple", Quantity: 1}}},
		},
		Scenes: map[string]*workerstate.SceneState{
			"scene_inn": {ID: "scene_inn", Name: "Inn"},
		},
		Items: map[string]*workerstate.ItemState{
			"apple": {ID: "apple", Name: "Apple", OwnerID: "player_1"},
		},
	})
	s := &playSession{view: a.authorityView(), playerNodeID: "player_1", currentSceneID: "scene_inn"}
	if err := a.runPlayUseItem(s, "Apple"); err != nil {
		t.Fatalf("runPlayUseItem returned error: %v", err)
	}
	if err := a.runPlayUseItem(s, "Knife"); err == nil {
		t.Fatal("expected missing item error")
	}
}

func TestRenderInventoryShowsDetailedEntries(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Name: "Hero", Kind: "player", Inventory: []workerstate.InventoryEntry{{ItemID: "apple", Quantity: 2}, {ItemID: "knife_bloody", Quantity: 1, Equipped: true}}},
		},
		Items: map[string]*workerstate.ItemState{
			"apple":        {ID: "apple", Name: "Apple"},
			"knife_bloody": {ID: "knife_bloody", Name: "Bloody Knife"},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1"}
	text := s.renderInventory()
	for _, want := range []string{"背包: Hero", "- Apple (apple) x2", "- Bloody Knife (knife_bloody) x1 [equipped]"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected inventory text to contain %q, got %q", want, text)
		}
	}
}

func TestRenderSceneSummaryIncludesPromptFlagsAndTarget(t *testing.T) {
	view := workerstate.NewStateView(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_inn"},
			"npc_1":    {ID: "npc_1", Name: "innkeeper", Kind: "npc", LocationID: "scene_inn"},
		},
		Scenes: map[string]*workerstate.SceneState{
			"scene_inn": {ID: "scene_inn", Name: "Inn", Description: "Warm light and wooden tables.", Flags: map[string]any{"open": true}},
		},
	})
	s := &playSession{view: view, playerNodeID: "player_1", currentSceneID: "scene_inn", currentTargetID: "npc_1"}
	text := s.renderSceneSummary()
	for _, want := range []string{"当前场景: Inn (scene_inn)", "当前目标: innkeeper", "场景状态: open=true", "直接输入文本可与 innkeeper 对话", "同场角色:"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected scene summary to contain %q, got %q", want, text)
		}
	}
}

func TestPrintPlayExecutionResultShowsOutcomeSummaries(t *testing.T) {
	a := newTestApp()
	s := &playSession{view: workerstate.NewStateView(&workerstate.WorldState{}), playerNodeID: "player_1"}
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}
	os.Stdout = w
	a.printPlayExecutionResult(s, &playerintent.ExecutionResult{Outcomes: []playerintent.StepOutcome{{Type: "move", Applied: true, Summary: "moved player_1 to scene_square"}}})
	_ = w.Close()
	os.Stdout = stdout
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "moved player_1 to scene_square") {
		t.Fatalf("expected printed execution summary, got %q", buf.String())
	}
}

func TestPrintPlayExecutionResultRefreshesSceneAfterMove(t *testing.T) {
	a := newTestApp()
	a.setAuthorityState(&workerstate.WorldState{
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_square"},
		},
		Scenes: map[string]*workerstate.SceneState{
			"scene_inn":    {ID: "scene_inn", Name: "Inn"},
			"scene_square": {ID: "scene_square", Name: "Square", Occupants: []string{"player_1"}},
		},
	})
	s := &playSession{
		view:           a.authorityView(),
		playerNodeID:   "player_1",
		currentSceneID: "scene_inn",
	}
	stdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe returned error: %v", err)
	}
	os.Stdout = w
	a.printPlayExecutionResult(s, &playerintent.ExecutionResult{Outcomes: []playerintent.StepOutcome{{Type: "move", Applied: true, Summary: "moved player_1 to scene_square"}}})
	_ = w.Close()
	os.Stdout = stdout
	if s.currentSceneID != "scene_square" {
		t.Fatalf("expected current scene to refresh to scene_square, got %q", s.currentSceneID)
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("ReadFrom returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "scene_square") {
		t.Fatalf("expected refreshed scene output, got %q", buf.String())
	}
}

func TestResolvePlayResponseReturnsResumedResponse(t *testing.T) {
	a := newTestApp()
	a.cfg.PlayAutoWorker = true
	a.cfg.Consumer = "game_client"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/runtime/tasks/pending":
			_ = json.NewEncoder(w).Encode(map[string]any{"tasks": []map[string]any{{"task_id": "task_1", "interface_name": sdk.AuthorityInterfaceGameClientRequestData, "callback_id": "cb_1", "payload_json": `{"request_data":{"queries":[{"type":"scene_state","node_id":"scene_inn"}]}}`, "status": "pending"}}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/runtime/tasks/claim":
			_ = json.NewEncoder(w).Encode(map[string]any{"task": map[string]any{"task_id": "task_1", "interface_name": sdk.AuthorityInterfaceGameClientRequestData, "callback_id": "cb_1", "payload_json": `{"request_data":{"queries":[{"type":"scene_state","node_id":"scene_inn"}]}}`, "status": "claimed", "lease_token": "lease_1"}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/runtime/tasks/start":
			_ = json.NewEncoder(w).Encode(map[string]any{"task": map[string]any{"task_id": "task_1", "interface_name": sdk.AuthorityInterfaceGameClientRequestData, "callback_id": "cb_1", "payload_json": `{"request_data":{"queries":[{"type":"scene_state","node_id":"scene_inn"}]}}`, "status": "running", "lease_token": "lease_1"}})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/actions/callback":
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "resumed": map[string]any{"request_id": "req_2", "task_type": "npc_dialogue", "execution_mode": "production", "reply": "Innkeeper responds.", "action_calls": []any{}, "memory_updates": []any{}}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	a.cfg.EngineBaseURL = server.URL
	a.setAuthorityState(&workerstate.WorldState{WorldID: "world_1", Scenes: map[string]*workerstate.SceneState{"scene_inn": {ID: "scene_inn"}}})
	resp, err := a.resolvePlayResponse(&sdk.InvokeResponse{ActionCalls: []sdk.ActionCall{{ActionID: "data_request", Mode: "async"}}})
	if err != nil {
		t.Fatalf("resolvePlayResponse returned error: %v", err)
	}
	if resp == nil || resp.Reply != "Innkeeper responds." {
		t.Fatalf("unexpected resumed response: %#v", resp)
	}
}

func TestHasPendingDataRequestDetectsAsyncCallback(t *testing.T) {
	if !(&sdk.InvokeResponse{ActionCalls: []sdk.ActionCall{{ActionID: sdk.ActionIDDataRequest, Mode: sdk.ActionModeAsync}}}).HasPendingDataRequest() {
		t.Fatal("expected pending data request detection")
	}
	if (&sdk.InvokeResponse{ActionCalls: []sdk.ActionCall{{ActionID: "spawn_item", Mode: sdk.ActionModeAsync}}}).HasPendingDataRequest() {
		t.Fatal("did not expect non-data_request action to be treated as pending data request")
	}
}

func TestNewTestCommandRegistersExpectedScenarios(t *testing.T) {
	got := SupportedTestScenarios()
	sort.Strings(got)
	want := []string{"all", "base-data", "callback-resume", "continuity", "machine-scenario", "runtime-tasks", "tooling-smoke"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected test scenarios: got=%v want=%v", got, want)
	}
	a := newTestApp()
	flags := pflag.NewFlagSet("base-data", pflag.ContinueOnError)
	a.bindTestFlags(flags)
	for _, name := range []string{"engine-exe", "devcli-exe", "worker-exe", "tests-dir", "out", "engine-port", "push-port", "keep-temp", "json"} {
		if flags.Lookup(name) == nil {
			t.Fatalf("expected flag %q on base-data subcommand", name)
		}
	}
}

func TestRunNamedTestScenarioReturnsNotImplemented(t *testing.T) {
	a := newTestApp()
	err := a.runNamedTestScenario("unknown-scenario")
	if err == nil {
		t.Fatal("expected not implemented error")
	}
	if got := err.Error(); got != "worker test scenario \"unknown-scenario\" is not implemented yet" {
		t.Fatalf("unexpected error: %s", got)
	}
}
