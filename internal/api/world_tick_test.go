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

type tickStubProvider struct{}

func (tickStubProvider) Chat(req *engine.LLMChatRequest) (*engine.LLMResult, error) {
	return &engine.LLMResult{Content: `{"reply":"ok","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"ok","world_events":[],"proposed_actions":[]}}`, Model: "stub", Tokens: 9}, nil
}

func (tickStubProvider) ModelName() string { return "stub" }

func initWorldTickTestDB(t *testing.T) string {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "Tick World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	raw, err := engine.EncodeWorldTimeSettings(&engine.WorldTimeSettings{
		TickScaleMode: engine.TickScaleModeFixed,
		TickMinUnit:   "时辰",
		TickStep:      1,
		TickUnits:     []string{"日", "时辰"},
		TimeScaleCarry: []engine.WorldTimeCarryRule{{
			From: "时辰",
			To:   "日",
			Base: 12,
		}},
	})
	if err != nil {
		t.Fatalf("encode world time settings: %v", err)
	}
	if _, err := store.UpsertWorldSettingsWithMask(world.UUID, &store.WorldSettingsModel{WorldTimeSettingsJSON: raw}, &store.WorldSettingsUpdateMask{WorldTimeSettings: true}); err != nil {
		t.Fatalf("upsert world settings: %v", err)
	}
	return world.UUID
}

func TestMakeTickAdvanceHandlerRejectsFixedScaleRequestedTicks(t *testing.T) {
	worldID := initWorldTickTestDB(t)
	handler := MakeTickAdvanceHandler(engine.NewPipeline(tickStubProvider{}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/worlds/"+worldID+"/ticks/advance", strings.NewReader(`{"tick_type":"manual","game_time":"day-1","requested_ticks":2}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_world_tick_request") {
		t.Fatalf("expected invalid_world_tick_request code, got %s", w.Body.String())
	}
}
