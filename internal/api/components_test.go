package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initComponentsTestDB(t *testing.T) (string, string) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "Component World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	node := &store.NodeModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, Name: "Guard", NodeType: "npc", ParentID: &world.ID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	store.ResolveNodeParentUUID(node)
	return world.UUID, node.UUID
}

func TestAddComponentHandlerSucceedsWhenWorldSettingsAlreadyExist(t *testing.T) {
	worldID, nodeID := initComponentsTestDB(t)
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{PipelineMode: "full"}, &store.WorldSettingsUpdateMask{PipelineMode: true}); err != nil {
		t.Fatalf("seed world settings: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/components", strings.NewReader(`{"node_id":"`+nodeID+`","component_type":"profile","data":"{\"role\":\"gatekeeper\"}"}`))
	w := httptest.NewRecorder()

	AddComponentHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
	components, err := store.GetNodeComponents(nodeID)
	if err != nil {
		t.Fatalf("get node components: %v", err)
	}
	if len(components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(components))
	}
	var count int64
	if err := store.DB.Model(&store.WorldSettingsModel{}).Where("world_uuid = ?", worldID).Count(&count).Error; err != nil {
		t.Fatalf("count world settings: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 world settings row, got %d", count)
	}
}
