package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type stateComponentEnvelope struct {
	ComponentType string                `json:"component_type"`
	Component     *store.ComponentModel `json:"component,omitempty"`
	Data          any                   `json:"data,omitempty"`
}

type timelineEnvelope struct {
	TickNumber    int    `json:"tick_number"`
	TickType      string `json:"tick_type"`
	GameTime      string `json:"game_time,omitempty"`
	Summary       string `json:"summary,omitempty"`
	FutureOutline string `json:"future_outline,omitempty"`
	Timeline      any    `json:"timeline"`
	Data          any    `json:"data,omitempty"`
}

func decodeComponentJSON(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return nil
	}
	return data
}

func writeStateComponentEnvelope(componentType engine.ComponentType, component *store.ComponentModel) stateComponentEnvelope {
	envelope := stateComponentEnvelope{ComponentType: string(componentType)}
	if component == nil {
		return envelope
	}
	envelope.Component = component
	envelope.Data = decodeComponentJSON(component.Data)
	return envelope
}

func writeTimelineEnvelope(item store.TimelineModel) timelineEnvelope {
	return timelineEnvelope{
		TickNumber:    item.TickNumber,
		TickType:      item.TickType,
		GameTime:      item.GameTime,
		Summary:       item.Summary,
		FutureOutline: item.FutureOutline,
		Timeline:      item,
		Data:          decodeComponentJSON(item.Data),
	}
}

// GetStateComponentsHandler returns all engine-recognized continuity state components for a world.
func GetStateComponentsHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	items := make([]stateComponentEnvelope, 0, len(engine.StateComponentTypes()))
	for _, componentType := range engine.StateComponentTypes() {
		component, err := service.GetStateComponent(worldID, componentType)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		items = append(items, writeStateComponentEnvelope(componentType, component))
	}
	writeJSON(w, http.StatusOK, map[string]any{"world_id": worldID, "components": items})
}

// GetStateComponentHandler returns one continuity state component for a world.
func GetStateComponentHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	componentType := r.PathValue("component_type")
	if !engine.IsStateComponentType(componentType) {
		errorJSONCode(w, http.StatusBadRequest, "invalid_component_type", "component_type must be one of: world_state, story_state, story_history, tick_policy, state_snapshot")
		return
	}
	component, err := service.GetStateComponent(worldID, engine.ComponentType(componentType))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"world_id": worldID, "state_component": writeStateComponentEnvelope(engine.ComponentType(componentType), component)})
}

// PutStateComponentHandler creates or updates one continuity state component for a world.
func PutStateComponentHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	componentType := r.PathValue("component_type")
	if !engine.IsStateComponentType(componentType) {
		errorJSONCode(w, http.StatusBadRequest, "invalid_component_type", "component_type must be one of: world_state, story_state, story_history, tick_policy, state_snapshot")
		return
	}
	var payload any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errorJSONCode(w, http.StatusBadRequest, "invalid_component_data", "invalid json: "+err.Error())
		return
	}
	component, err := service.UpsertStateComponent(worldID, engine.ComponentType(componentType), payload)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"world_id": worldID, "state_component": writeStateComponentEnvelope(engine.ComponentType(componentType), component)})
}

// GetTimelinesHandler returns recent world tick history with parsed payloads.
func GetTimelinesHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	limit := 20
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	items, err := store.GetTimelineTicks(worldID, limit)
	if err != nil {
		errorJSONCode(w, http.StatusInternalServerError, "timeline_query_failed", err.Error())
		return
	}
	result := make([]timelineEnvelope, 0, len(items))
	for _, item := range items {
		result = append(result, writeTimelineEnvelope(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"world_id": worldID, "timelines": result})
}

// GetLatestTimelineHandler returns the latest world tick history entry.
func GetLatestTimelineHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	item, err := store.GetLatestTick(worldID)
	if err != nil {
		errorJSONCode(w, http.StatusNotFound, "timeline_not_found", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"world_id": worldID, "timeline": writeTimelineEnvelope(*item)})
}
