package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initPolicyTestDB(t *testing.T) string {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "Policy World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	return world.UUID
}

func TestSetWorldPolicyHandlerReturnsWorldUUID(t *testing.T) {
	worldID := initPolicyTestDB(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/policy", strings.NewReader(`{"blocked_actions":["spawn_item"],"safe_actions":["inspect_map"]}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldPolicyHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		WorldID        string   `json:"world_id"`
		BlockedActions []string `json:"blocked_actions"`
		SafeActions    []string `json:"safe_actions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.WorldID != worldID {
		t.Fatalf("expected world_id %q, got %q", worldID, body.WorldID)
	}
	if len(body.BlockedActions) != 1 || body.BlockedActions[0] != "spawn_item" {
		t.Fatalf("unexpected blocked actions: %#v", body.BlockedActions)
	}
	if len(body.SafeActions) != 1 || body.SafeActions[0] != "inspect_map" {
		t.Fatalf("unexpected safe actions: %#v", body.SafeActions)
	}
}
