package engine

import "testing"

func TestValidateDynamicActionArgsAllowsReservedInjectedFields(t *testing.T) {
	err := validateDynamicActionArgs(
		map[string]any{
			"intent":             "quote",
			"node_id":            "npc-1",
			"external_interface": "npc_trade_action",
		},
		map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"intent": map[string]any{"type": "string"},
			},
			"required": []string{"intent"},
		},
	)
	if err != nil {
		t.Fatalf("expected reserved fields to be ignored, got %v", err)
	}
}

func TestValidateDynamicInterfacesRejectsActionArgsSchemaWithoutObjectType(t *testing.T) {
	err := ValidateDynamicInterfaces([]DynamicInterface{{
		ID:                "merchant_ops",
		Kind:              DynamicInterfaceAction,
		ExternalInterface: "npc_trade_action",
		ArgsSchema: map[string]any{
			"type": "string",
		},
	}})
	if err == nil {
		t.Fatal("expected args_schema validation error")
	}
	if got := err.Error(); got != "dynamic_interfaces[0].args_schema.type must be object when provided" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestValidateInteractionContextRejectsInvalidEventType(t *testing.T) {
	err := ValidateInteractionContext(&InteractionContext{
		Mode:          "direct_dialogue",
		SpeakerNodeID: "player_1",
		TargetNodeID:  "npc_1",
		Event:         &InteractionEvent{Type: "teleport"},
	})
	if err == nil {
		t.Fatal("expected interaction validation error")
	}
	if got := err.Error(); got != "interaction.event.type must be one of: speech, gift, show_item, trade_request, threaten" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestValidateInteractionContextAcceptsValidPayload(t *testing.T) {
	err := ValidateInteractionContext(&InteractionContext{
		Mode:               "group_chat",
		SpeakerNodeID:      "player_1",
		TargetNodeID:       "npc_1",
		SceneNodeID:        "scene_tavern",
		RoomID:             "room_mainhall",
		ParticipantNodeIDs: []string{"player_1", "npc_1", "npc_2"},
		AudienceScope:      "public",
		TurnIndex:          1,
		Event:              &InteractionEvent{Type: "speech"},
	})
	if err != nil {
		t.Fatalf("expected valid interaction payload, got %v", err)
	}
}

func TestValidateInteractionContextRejectsInvalidAudienceScope(t *testing.T) {
	err := ValidateInteractionContext(&InteractionContext{
		Mode:          "direct_dialogue",
		SpeakerNodeID: "player_1",
		TargetNodeID:  "npc_1",
		AudienceScope: "secret",
	})
	if err == nil {
		t.Fatal("expected interaction validation error")
	}
	if got := err.Error(); got != "interaction.audience_scope must be one of: public, private, whisper" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestValidateInteractionContextRejectsDuplicateParticipants(t *testing.T) {
	err := ValidateInteractionContext(&InteractionContext{
		Mode:               "group_chat",
		SpeakerNodeID:      "player_1",
		TargetNodeID:       "npc_1",
		ParticipantNodeIDs: []string{"player_1", "npc_1", "player_1"},
	})
	if err == nil {
		t.Fatal("expected interaction validation error")
	}
	if got := err.Error(); got != "interaction.participant_node_ids[2] duplicated: player_1" {
		t.Fatalf("unexpected error: %s", got)
	}
}
