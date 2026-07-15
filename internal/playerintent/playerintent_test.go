package playerintent

import (
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func sampleState() *workerstate.WorldState {
	return &workerstate.WorldState{
		WorldID: "world_1",
		Actors: map[string]*workerstate.ActorState{
			"player_1": {ID: "player_1", Kind: "player", LocationID: "scene_inn", Money: 30, Inventory: []workerstate.InventoryEntry{{ItemID: "knife_bloody", Quantity: 1}, {ItemID: "apple", Quantity: 2}}},
			"npc_1":    {ID: "npc_1", Kind: "npc", LocationID: "scene_inn"},
			"npc_2":    {ID: "npc_2", Kind: "npc", LocationID: "scene_square"},
		},
		Scenes: map[string]*workerstate.SceneState{
			"scene_inn":    {ID: "scene_inn", Occupants: []string{"player_1", "npc_1"}, Flags: map[string]any{"open": true}},
			"scene_square": {ID: "scene_square", Occupants: []string{"npc_2"}},
		},
		Items: map[string]*workerstate.ItemState{
			"knife_bloody": {ID: "knife_bloody", OwnerID: "player_1"},
			"apple":        {ID: "apple", OwnerID: "player_1"},
		},
		Tasks: map[string]*workerstate.QuestState{
			"task_1": {ID: "task_1", Status: "active", Stage: "investigate"},
		},
	}
}

func TestValidateShowItemIntentRequiresPossessionAndSameScene(t *testing.T) {
	view := workerstate.NewStateView(sampleState())
	payload := &sdk.PlayerIntentInterpretation{Intent: &sdk.PlayerIntent{
		Type:         "show_item",
		ActorNodeID:  "player_1",
		TargetNodeID: "npc_1",
		SceneNodeID:  "scene_inn",
		RiskLevel:    "low",
		Steps: []sdk.PlayerIntentStep{{
			Type:          "show_item",
			TargetNodeID:  "npc_1",
			ItemID:        "knife_bloody",
			Preconditions: []sdk.PlayerIntentPrecondition{{Type: "same_scene"}, {Type: "item_present", ItemID: "knife_bloody"}},
		}},
	}}
	result := Validate(view, payload)
	if !result.OK {
		t.Fatalf("expected validation success, got %#v", result)
	}
}

func TestValidateRejectsMissingItem(t *testing.T) {
	view := workerstate.NewStateView(sampleState())
	payload := &sdk.PlayerIntentInterpretation{Intent: &sdk.PlayerIntent{
		Type:         "show_item",
		ActorNodeID:  "player_1",
		TargetNodeID: "npc_1",
		SceneNodeID:  "scene_inn",
		RiskLevel:    "low",
		Steps:        []sdk.PlayerIntentStep{{Type: "show_item", TargetNodeID: "npc_1", ItemID: "ring_missing"}},
	}}
	result := Validate(view, payload)
	if result.OK {
		t.Fatalf("expected validation failure, got %#v", result)
	}
}

func TestExecuteGiftMovesInventoryOwnership(t *testing.T) {
	state := sampleState()
	payload := &sdk.PlayerIntentInterpretation{Intent: &sdk.PlayerIntent{
		Type:         "gift",
		ActorNodeID:  "player_1",
		TargetNodeID: "npc_1",
		SceneNodeID:  "scene_inn",
		RiskLevel:    "low",
		Steps: []sdk.PlayerIntentStep{{
			Type:          "gift",
			TargetNodeID:  "npc_1",
			ItemID:        "apple",
			Preconditions: []sdk.PlayerIntentPrecondition{{Type: "same_scene"}, {Type: "item_present", ItemID: "apple"}},
		}},
	}}
	result, err := Execute(state, payload)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if len(result.Outcomes) != 1 || !result.Outcomes[0].Applied {
		t.Fatalf("unexpected execution result: %#v", result)
	}
	view := workerstate.NewStateView(state)
	if entry, ok := view.ActorInventoryEntry("player_1", "apple"); !ok || entry == nil || entry.Quantity != 1 {
		t.Fatalf("expected player apple quantity to decrease to 1, got %#v", entry)
	}
	if !view.ItemPresentOnActor("npc_1", "apple") {
		t.Fatal("expected apple added to npc inventory")
	}
}

func TestExecuteCompositeStopsOnValidationFailure(t *testing.T) {
	state := sampleState()
	payload := &sdk.PlayerIntentInterpretation{Intent: &sdk.PlayerIntent{
		Type:         "composite",
		ActorNodeID:  "player_1",
		TargetNodeID: "npc_2",
		SceneNodeID:  "scene_inn",
		RiskLevel:    "medium",
		Steps: []sdk.PlayerIntentStep{
			{Type: "show_item", TargetNodeID: "npc_2", ItemID: "knife_bloody", Preconditions: []sdk.PlayerIntentPrecondition{{Type: "same_scene"}}},
			{Type: "speech", TargetNodeID: "npc_2", Content: "look at this"},
		},
	}}
	if _, err := Execute(state, payload); err == nil {
		t.Fatal("expected validation failure")
	}
}

func TestBuildInteractionSpecUsesSuggestedInteraction(t *testing.T) {
	payload := &sdk.PlayerIntentInterpretation{
		Intent: &sdk.PlayerIntent{
			Type:         "speech",
			ActorNodeID:  "player_1",
			TargetNodeID: "npc_1",
			Summary:      "ask question",
			RiskLevel:    "low",
			Confidence:   0.8,
			Steps:        []sdk.PlayerIntentStep{{Type: "speech", TargetNodeID: "npc_1", Content: "where did he go?"}},
		},
		SuggestedInteraction: &sdk.SuggestedInteraction{Mode: "group_chat", EventType: "speech", AudienceScope: "public", TargetNodeID: "npc_1"},
	}
	spec, err := BuildInteractionSpec(payload, "player_1", "scene_inn")
	if err != nil {
		t.Fatalf("BuildInteractionSpec returned error: %v", err)
	}
	if spec.Mode != "group_chat" || spec.AudienceScope != "public" || spec.TargetNodeID != "npc_1" {
		t.Fatalf("unexpected interaction spec: %#v", spec)
	}
}
