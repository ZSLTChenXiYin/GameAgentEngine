package engine

import (
	"strings"
	"testing"
)

func TestBuildDynamicInterfaceTools(t *testing.T) {
	tools := buildDynamicInterfaceTools([]DynamicInterface{
		{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"node_detail", "visible_entities"},
			MaxQueries:        2,
		},
		{
			ID:                "merchant_ops",
			Kind:              DynamicInterfaceAction,
			ExternalInterface: "npc_trade_action",
			Description:       "Perform trade-related external actions",
			ArgsSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"intent": map[string]any{"type": "string"},
				},
			},
		},
	})
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %+v", tools)
	}
	if tools[0].Invocation != LLMToolInvocationDataRequest || tools[0].DataRequest == nil {
		t.Fatalf("expected first tool to be data_request, got %+v", tools[0])
	}
	if tools[0].DataRequest.Target != "game_client" || tools[0].DataRequest.ExternalInterface != "game_client_request_data" {
		t.Fatalf("unexpected data request template: %+v", tools[0].DataRequest)
	}
	queries := tools[0].Parameters["properties"].(map[string]any)["queries"].(map[string]any)
	if queries["maxItems"] != 2 {
		t.Fatalf("expected maxItems 2, got %+v", queries)
	}
	if tools[1].Invocation != LLMToolInvocationAction || tools[1].ActionID != "merchant_ops" {
		t.Fatalf("expected second tool to be action, got %+v", tools[1])
	}
	if tools[1].Parameters["type"] != "object" {
		t.Fatalf("expected args schema to survive, got %+v", tools[1].Parameters)
	}
}

func TestBuildDialogueModeInstructionIncludesShowItemConstraint(t *testing.T) {
	text := buildDialogueModeInstruction("direct_dialogue", &InteractionContext{
		Event: &InteractionEvent{Type: "show_item", ItemID: "knife_bloody"},
	})
	if !strings.Contains(text, "展示物品") {
		t.Fatalf("expected show_item guidance, got %q", text)
	}
}

func TestBuildDialogueModeInstructionIncludesThreatenConstraint(t *testing.T) {
	text := buildDialogueModeInstruction("direct_dialogue", &InteractionContext{
		Event: &InteractionEvent{Type: "threaten"},
	})
	if !strings.Contains(text, "威胁施压") {
		t.Fatalf("expected threaten guidance, got %q", text)
	}
}

func TestBuildPlayerIntentPromptIncludesProposalConstraints(t *testing.T) {
	text := buildPlayerIntentPrompt("system context", "player_1", &InteractionContext{
		Mode:         "direct_dialogue",
		TargetNodeID: "npc_innkeeper",
		SceneNodeID:  "scene_inn",
	})
	for _, want := range []string{"行为意图提案", "player_intent", "missing_facts", "suggested_interaction", "request_data", "后续由权威接口验证后才能当真", "composite"} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected prompt to contain %q, got %q", want, text)
		}
	}
}
