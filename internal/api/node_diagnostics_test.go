package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func TestGetNodeHandlerIncludesRelationDiagnostics(t *testing.T) {
	if err := store.Init("sqlite", "file:node-diagnostics-test?mode=memory&cache=shared"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	location := &store.NodeModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, Name: "Dock", NodeType: "location", ParentID: &world.ID}
	if err := store.CreateNode(location); err != nil {
		t.Fatalf("create location: %v", err)
	}
	store.ResolveNodeParentUUID(location)
	npc := &store.NodeModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, Name: "Guard", NodeType: "npc", ParentID: &world.ID}
	if err := store.CreateNode(npc); err != nil {
		t.Fatalf("create npc: %v", err)
	}
	store.ResolveNodeParentUUID(npc)
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, SourceID: npc.ID, SourceUUID: npc.UUID, TargetID: location.ID, TargetUUID: location.UUID, RelationType: "located_at", Weight: 1}); err != nil {
		t.Fatalf("create located_at: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, SourceID: npc.ID, SourceUUID: npc.UUID, TargetID: world.ID, TargetUUID: world.UUID, RelationType: "external_parent", Weight: 1}); err != nil {
		t.Fatalf("create external_parent: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/"+npc.UUID, nil)
	req.SetPathValue("id", npc.UUID)
	w := httptest.NewRecorder()

	GetNodeHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	issues, ok := payload["relation_validation_issues"].([]any)
	if !ok || len(issues) == 0 {
		t.Fatalf("expected relation validation issues, got %#v", payload["relation_validation_issues"])
	}
	preview, ok := payload["graph_context_preview"].(map[string]any)
	if !ok {
		t.Fatalf("expected graph_context_preview, got %#v", payload["graph_context_preview"])
	}
	summary, ok := preview["summary"].([]any)
	if !ok || len(summary) == 0 {
		t.Fatalf("expected graph summary, got %#v", preview)
	}
}
