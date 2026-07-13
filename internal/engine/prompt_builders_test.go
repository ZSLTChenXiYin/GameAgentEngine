package engine

import "testing"

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
