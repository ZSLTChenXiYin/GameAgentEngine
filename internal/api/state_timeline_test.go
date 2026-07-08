package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initStateTimelineTestDB(t *testing.T) string {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "State World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	return world.UUID
}

func TestGetStateComponentsHandlerReturnsRecognizedEntries(t *testing.T) {
	worldID := initStateTimelineTestDB(t)
	if _, err := service.UpsertStateComponent(worldID, engine.CompWorldState, engine.WorldStateComponent{Summary: "storm front moving inland"}); err != nil {
		t.Fatalf("upsert state component: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/worlds/"+worldID+"/state-components", nil)
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	GetStateComponentsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		WorldID    string `json:"world_id"`
		Components []struct {
			ComponentType string         `json:"component_type"`
			Data          map[string]any `json:"data"`
		} `json:"components"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.WorldID != worldID {
		t.Fatalf("expected world_id %q, got %q", worldID, body.WorldID)
	}
	if len(body.Components) != len(engine.StateComponentTypes()) {
		t.Fatalf("expected %d components, got %d", len(engine.StateComponentTypes()), len(body.Components))
	}
	if body.Components[0].ComponentType != string(engine.CompWorldState) {
		t.Fatalf("expected first component %q, got %q", engine.CompWorldState, body.Components[0].ComponentType)
	}
	if body.Components[0].Data["summary"] != "storm front moving inland" {
		t.Fatalf("unexpected state data: %#v", body.Components[0].Data)
	}
}

func TestPutStateComponentHandlerPersistsStructuredPayload(t *testing.T) {
	worldID := initStateTimelineTestDB(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/state-components/world_state", strings.NewReader(`{"summary":"vault breach","canonical_facts":["地下52米量子谐振腔"]}`))
	req.SetPathValue("world_id", worldID)
	req.SetPathValue("component_type", string(engine.CompWorldState))
	w := httptest.NewRecorder()

	PutStateComponentHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	component, err := service.GetStateComponent(worldID, engine.CompWorldState)
	if err != nil {
		t.Fatalf("load component: %v", err)
	}
	if component == nil || !strings.Contains(component.Data, "地下52米量子谐振腔") {
		t.Fatalf("expected persisted component data, got %#v", component)
	}
	var body struct {
		StateComponent struct {
			ComponentType string         `json:"component_type"`
			Data          map[string]any `json:"data"`
		} `json:"state_component"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.StateComponent.ComponentType != string(engine.CompWorldState) {
		t.Fatalf("unexpected component type: %q", body.StateComponent.ComponentType)
	}
	if body.StateComponent.Data["summary"] != "vault breach" {
		t.Fatalf("unexpected response payload: %#v", body.StateComponent.Data)
	}
}

func TestGetTimelinesHandlerReturnsParsedTimelinePayload(t *testing.T) {
	worldID := initStateTimelineTestDB(t)
	worldInt := store.ResolveWorldUUID(worldID)
	if err := store.CreateTimelineTick(&store.TimelineModel{
		UUID:          store.NewUUID(),
		WorldID:       worldInt,
		WorldUUID:     worldID,
		TickNumber:    3,
		TickType:      "daily",
		GameTime:      "Day 3",
		Summary:       "reactor stabilized",
		Data:          `{"reply":"ok","future_outline":"watch the western ridge"}`,
		FutureOutline: "watch the western ridge",
	}); err != nil {
		t.Fatalf("create timeline: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/worlds/"+worldID+"/timelines?limit=5", nil)
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	GetTimelinesHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Timelines []struct {
			TickNumber int            `json:"tick_number"`
			Data       map[string]any `json:"data"`
		} `json:"timelines"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.Timelines) != 1 {
		t.Fatalf("expected one timeline, got %d", len(body.Timelines))
	}
	if body.Timelines[0].TickNumber != 3 {
		t.Fatalf("unexpected tick number: %d", body.Timelines[0].TickNumber)
	}
	if body.Timelines[0].Data["future_outline"] != "watch the western ridge" {
		t.Fatalf("unexpected timeline data: %#v", body.Timelines[0].Data)
	}
}
