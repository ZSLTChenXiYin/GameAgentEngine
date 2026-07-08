package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initLogsTestDB(t *testing.T) (string, string) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "Logs World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	node := &store.NodeModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, Name: "Observer", NodeType: "npc"}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	return world.UUID, node.UUID
}

func TestGetLogsHandlerSupportsStructuredFilters(t *testing.T) {
	worldID, nodeID := initLogsTestDB(t)
	if err := store.CreateInferenceLog(&store.InferenceLogModel{
		WorldUUID:     worldID,
		NodeUUID:      nodeID,
		TaskType:      "world_tick",
		Category:      "pipeline",
		EventName:     "raw_llm_response_received",
		ExecutionMode: "debug",
		RequestID:     "req-1",
		Round:         2,
		Message:       "keep",
	}); err != nil {
		t.Fatalf("create filtered log: %v", err)
	}
	if err := store.CreateInferenceLog(&store.InferenceLogModel{
		WorldUUID:     worldID,
		NodeUUID:      nodeID,
		TaskType:      "world_tick",
		Category:      "pipeline",
		EventName:     "context_built",
		ExecutionMode: "review",
		RequestID:     "req-2",
		Round:         1,
		Message:       "skip",
	}); err != nil {
		t.Fatalf("create non-matching log: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs?world_id="+worldID+"&node_id="+nodeID+"&category=pipeline&event_name=raw_llm_response_received&execution_mode=debug&request_id=req-1&round=2", nil)
	w := httptest.NewRecorder()

	GetLogsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body []struct {
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected one log, got %d", len(body))
	}
	if body[0].Message != "keep" || body[0].RequestID != "req-1" {
		t.Fatalf("unexpected filtered log: %#v", body[0])
	}
}
