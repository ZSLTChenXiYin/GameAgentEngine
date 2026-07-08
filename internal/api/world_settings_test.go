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

func initWorldSettingsTestDB(t *testing.T) string {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "Settings World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world_id: %v", err)
	}
	return world.UUID
}

func TestSetWorldSettingsHandlerRejectsInvalidNumbers(t *testing.T) {
	worldID := initWorldSettingsTestDB(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/settings", strings.NewReader(`{"memory_limit":0}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldSettingsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_world_setting") {
		t.Fatalf("expected invalid_world_setting code, got %s", w.Body.String())
	}
}

func TestSetWorldSettingsHandlerAppliesPartialUpdate(t *testing.T) {
	worldID := initWorldSettingsTestDB(t)
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{MemoryLimit: 50, PipelineMode: "full"}, &store.WorldSettingsUpdateMask{MemoryLimit: true, PipelineMode: true}); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/settings", strings.NewReader(`{"pipeline_mode":"polling"}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldSettingsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		WorldID string `json:"world_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.WorldID != worldID {
		t.Fatalf("expected world_id %q, got %q", worldID, body.WorldID)
	}
	settings, err := store.GetWorldSettings(worldID)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.PipelineMode != "polling" {
		t.Fatalf("expected pipeline mode polling, got %q", settings.PipelineMode)
	}
	if settings.MemoryLimit != 50 {
		t.Fatalf("expected memory limit to remain 50, got %d", settings.MemoryLimit)
	}
}

func TestSetWorldSettingsHandlerPersistsExplicitZeroValues(t *testing.T) {
	worldID := initWorldSettingsTestDB(t)
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{
		PropagationMaxDepth: 3,
		SubTaskMaxRetries:   4,
		SubTaskTimeoutSecs:  90,
	}, &store.WorldSettingsUpdateMask{
		PropagationMaxDepth: true,
		SubTaskMaxRetries:   true,
		SubTaskTimeoutSecs:  true,
	}); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/settings", strings.NewReader(`{"propagation_max_depth":0,"sub_task_max_retries":0,"sub_task_timeout_secs":0}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldSettingsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	settings, err := store.GetWorldSettings(worldID)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.PropagationMaxDepth != 0 {
		t.Fatalf("expected propagation max depth 0, got %d", settings.PropagationMaxDepth)
	}
	if settings.SubTaskMaxRetries != 0 {
		t.Fatalf("expected sub task max retries 0, got %d", settings.SubTaskMaxRetries)
	}
	if settings.SubTaskTimeoutSecs != 0 {
		t.Fatalf("expected sub task timeout secs 0, got %d", settings.SubTaskTimeoutSecs)
	}
}

func TestSetWorldSettingsHandlerPersistsFalseBooleanUpdate(t *testing.T) {
	worldID := initWorldSettingsTestDB(t)
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{
		AutoApply:                true,
		EnablePropagationMachine: true,
	}, &store.WorldSettingsUpdateMask{
		AutoApply:                true,
		EnablePropagationMachine: true,
	}); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/settings", strings.NewReader(`{"auto_apply":false,"enable_propagation_machine":false}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldSettingsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	settings, err := store.GetWorldSettings(worldID)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.AutoApply {
		t.Fatal("expected auto_apply false")
	}
	if settings.EnablePropagationMachine {
		t.Fatal("expected enable_propagation_machine false")
	}
}

func TestSetWorldSettingsHandlerRejectsEmptyRequireReviewAbove(t *testing.T) {
	worldID := initWorldSettingsTestDB(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/settings", strings.NewReader(`{"require_review_above":""}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldSettingsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_world_setting") {
		t.Fatalf("expected invalid_world_setting code, got %s", w.Body.String())
	}
}

func TestSetWorldSettingsHandlerPersistsWorldTimeSettings(t *testing.T) {
	worldID := initWorldSettingsTestDB(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/settings", strings.NewReader(`{
		"world_time_settings": {
			"tick_scale_mode": "fixed",
			"tick_min_unit": "时辰",
			"tick_step": 2,
			"tick_units": ["年", "月", "日", "时辰"],
			"time_scale_carry": [
				{"from": "时辰", "to": "日", "base": 12},
				{"from": "日", "to": "月", "base": 30},
				{"from": "月", "to": "年", "base": 12}
			],
			"time_calendar": {
				"enabled": true,
				"calendar_name": "太阴",
				"units": [
					{"unit": "年", "value": "8"},
					{"unit": "月", "value": "7"},
					{"unit": "日", "value": "20"},
					{"unit": "时辰", "value": "卯"}
				]
			},
			"unit_value_sequences": [
				{"unit": "时辰", "values": ["子", "丑", "寅", "卯"]}
			]
		}
	}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldSettingsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	settings, err := store.GetWorldSettings(worldID)
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if !strings.Contains(settings.WorldTimeSettingsJSON, `"tick_scale_mode":"fixed"`) {
		t.Fatalf("expected world_time_settings_json to contain tick_scale_mode, got %s", settings.WorldTimeSettingsJSON)
	}
	var body struct {
		WorldTimeSettings struct {
			TickMinUnit  string `json:"tick_min_unit"`
			TickStep     int    `json:"tick_step"`
			TimeCalendar struct {
				CalendarName string `json:"calendar_name"`
			} `json:"time_calendar"`
		} `json:"world_time_settings"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.WorldTimeSettings.TickMinUnit != "时辰" || body.WorldTimeSettings.TickStep != 2 || body.WorldTimeSettings.TimeCalendar.CalendarName != "太阴" {
		t.Fatalf("unexpected world_time_settings response: %#v", body.WorldTimeSettings)
	}
}

func TestSetWorldSettingsHandlerRejectsInvalidWorldTimeSettings(t *testing.T) {
	worldID := initWorldSettingsTestDB(t)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/worlds/"+worldID+"/settings", strings.NewReader(`{
		"world_time_settings": {
			"tick_scale_mode": "fixed",
			"tick_min_unit": "时辰",
			"tick_step": 2,
			"tick_units": ["年", "月", "日", "时辰"],
			"time_scale_carry": [
				{"from": "时辰", "to": "日", "base": 12},
				{"from": "日", "to": "月", "base": 30},
				{"from": "月", "to": "年", "base": 12}
			],
			"time_calendar": {
				"enabled": true,
				"units": [
					{"unit": "年", "value": "8"},
					{"unit": "月", "value": "7"},
					{"unit": "日", "value": "20"},
					{"unit": "时辰", "value": "卯"}
				]
			}
		}
	}`))
	req.SetPathValue("world_id", worldID)
	w := httptest.NewRecorder()

	SetWorldSettingsHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_world_time_settings") {
		t.Fatalf("expected invalid_world_time_settings code, got %s", w.Body.String())
	}
}
