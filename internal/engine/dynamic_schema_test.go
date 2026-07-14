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
