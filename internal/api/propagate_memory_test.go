package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type propagateStubProvider struct{}

func (propagateStubProvider) Chat(req *engine.LLMChatRequest) (*engine.LLMResult, error) {
	return &engine.LLMResult{Content: `{"reply":"ok","action_calls":[],"memory_updates":[]}`, Model: "stub", Tokens: 3}, nil
}

func (propagateStubProvider) ModelName() string { return "stub" }

func initPropagateMemoryTestDB(t *testing.T) (string, string) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "Prop World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	node := &store.NodeModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, Name: "NPC", NodeType: "npc", ParentID: &world.ID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	store.ResolveNodeParentUUID(node)
	memory := &store.MemoryModel{UUID: store.NewUUID(), NodeID: node.ID, NodeUUID: node.UUID, Content: "Rumor", Level: string(engine.MemLongTerm), Tags: "rumor"}
	if err := store.CreateMemory(memory); err != nil {
		t.Fatalf("create memory: %v", err)
	}
	return world.UUID, memory.UUID
}

func TestMakePropagateMemoryHandlerRejectsUnsupportedMode(t *testing.T) {
	worldID, memoryID := initPropagateMemoryTestDB(t)
	handler := MakePropagateMemoryHandler(engine.NewPipeline(propagateStubProvider{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memories/propagate", strings.NewReader(`{"memory_id":"`+memoryID+`","mode":"sideways"}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "unsupported propagation mode") {
		t.Fatalf("expected unsupported propagation mode error, got %s", w.Body.String())
	}
}
