package sdk

import "testing"

func TestNewDynamicDataRequestBuilder(t *testing.T) {
	di := NewDynamicDataRequest(
		"scene_facts",
		"game_client_request_data",
		WithDescription("query visible scene facts"),
		WithQueryTypes("node_detail", "visible_entities"),
		WithArgsSchema(map[string]any{"type": "object"}),
		WithMaxQueries(2),
	)

	if di.ID != "scene_facts" {
		t.Fatalf("unexpected id: %#v", di)
	}
	if di.Kind != DynamicInterfaceDataRequest {
		t.Fatalf("unexpected kind: %#v", di)
	}
	if di.ExternalInterface != "game_client_request_data" {
		t.Fatalf("unexpected external interface: %#v", di)
	}
	if di.Description != "query visible scene facts" {
		t.Fatalf("unexpected description: %#v", di)
	}
	if len(di.QueryTypes) != 2 || di.QueryTypes[0] != "node_detail" || di.QueryTypes[1] != "visible_entities" {
		t.Fatalf("unexpected query types: %#v", di.QueryTypes)
	}
	if di.ArgsSchema["type"] != "object" {
		t.Fatalf("unexpected args schema: %#v", di.ArgsSchema)
	}
	if di.MaxQueries != 2 {
		t.Fatalf("unexpected max queries: %#v", di)
	}
}

func TestNewDynamicActionBuilder(t *testing.T) {
	di := NewDynamicAction(
		"merchant_ops",
		"npc_trade_action",
		WithActionDescription("execute merchant trade operations"),
		WithActionArgsSchema(map[string]any{"type": "object", "properties": map[string]any{"item_id": map[string]any{"type": "string"}}}),
		WithMaxCalls(1),
	)

	if di.Kind != DynamicInterfaceAction {
		t.Fatalf("unexpected kind: %#v", di)
	}
	if di.Description != "execute merchant trade operations" {
		t.Fatalf("unexpected description: %#v", di)
	}
	if di.MaxCalls != 1 {
		t.Fatalf("unexpected max calls: %#v", di)
	}
	properties, ok := di.ArgsSchema["properties"].(map[string]any)
	if !ok || properties["item_id"] == nil {
		t.Fatalf("unexpected args schema: %#v", di.ArgsSchema)
	}
}

func TestInvokeRequestAddDynamicInterfacesInitializesContext(t *testing.T) {
	req := &InvokeRequest{WorldID: "world-1", NodeID: "npc-1", TaskType: "npc_dialogue"}
	req.AddDynamicInterfaces(
		NewDynamicDataRequest("scene_facts", "game_client_request_data"),
		NewDynamicAction("merchant_ops", "npc_trade_action"),
	)

	if req.Context == nil {
		t.Fatal("expected context to be initialized")
	}
	if len(req.Context.DynamicInterfaces) != 2 {
		t.Fatalf("expected 2 dynamic interfaces, got %#v", req.Context.DynamicInterfaces)
	}
	if req.Context.DynamicInterfaces[0].Kind != DynamicInterfaceDataRequest {
		t.Fatalf("unexpected first interface: %#v", req.Context.DynamicInterfaces[0])
	}
	if req.Context.DynamicInterfaces[1].Kind != DynamicInterfaceAction {
		t.Fatalf("unexpected second interface: %#v", req.Context.DynamicInterfaces[1])
	}
}

func TestInvokeContextAddDynamicInterfacesAppends(t *testing.T) {
	ctx := NewInvokeContext().AddDynamicInterfaces(NewDynamicDataRequest("scene", "game_client_request_data"))
	ctx.AddDynamicInterfaces(NewDynamicAction("trade", "npc_trade_action"))

	if len(ctx.DynamicInterfaces) != 2 {
		t.Fatalf("expected appended interfaces, got %#v", ctx.DynamicInterfaces)
	}
}
