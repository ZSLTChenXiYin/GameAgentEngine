package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateWorldSettingsSendsOnlyChangedFields(t *testing.T) {
	var payload map[string]any
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"world_id":                   "world-1",
			"memory_limit":               88,
			"max_analysis_rounds":        5,
			"max_context_depth":          3,
			"auto_apply":                 true,
			"require_review_above":       "critical",
			"propagation_max_depth":      2,
			"sub_task_max_retries":       2,
			"sub_task_timeout_secs":      60,
			"enable_propagation_machine": false,
			"pipeline_mode":              "polling",
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dev-key")
	memoryLimit := 88
	pipelineMode := "polling"
	result, err := client.UpdateWorldSettings("world-1", &WorldSettingsUpdate{
		MemoryLimit:  &memoryLimit,
		PipelineMode: &pipelineMode,
	})
	if err != nil {
		t.Fatalf("update world settings: %v", err)
	}
	if result.MemoryLimit != 88 || result.PipelineMode != "polling" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if gotPath != "/api/v1/worlds/world-1/settings" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if len(payload) != 2 {
		t.Fatalf("expected 2 fields in payload, got %#v", payload)
	}
	if payload["memory_limit"] != float64(88) {
		t.Fatalf("unexpected memory_limit payload: %#v", payload)
	}
	if payload["pipeline_mode"] != "polling" {
		t.Fatalf("unexpected pipeline_mode payload: %#v", payload)
	}
}

func TestUpdateWorldSettingsSendsExplicitZeroAndFalseValues(t *testing.T) {
	var payload map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"world_id":                   "world-1",
			"memory_limit":               50,
			"max_analysis_rounds":        5,
			"max_context_depth":          3,
			"auto_apply":                 false,
			"require_review_above":       "critical",
			"propagation_max_depth":      0,
			"sub_task_max_retries":       0,
			"sub_task_timeout_secs":      0,
			"enable_propagation_machine": false,
			"pipeline_mode":              "full",
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dev-key")
	propagationDepth := 0
	retries := 0
	timeout := 0
	autoApply := false
	enableMachine := false
	_, err := client.UpdateWorldSettings("world-1", &WorldSettingsUpdate{
		PropagationMaxDepth:      &propagationDepth,
		SubTaskMaxRetries:        &retries,
		SubTaskTimeoutSecs:       &timeout,
		AutoApply:                &autoApply,
		EnablePropagationMachine: &enableMachine,
	})
	if err != nil {
		t.Fatalf("update world settings: %v", err)
	}
	if len(payload) != 5 {
		t.Fatalf("expected 5 fields in payload, got %#v", payload)
	}
	if payload["propagation_max_depth"] != float64(0) {
		t.Fatalf("unexpected propagation_max_depth payload: %#v", payload)
	}
	if payload["sub_task_max_retries"] != float64(0) {
		t.Fatalf("unexpected sub_task_max_retries payload: %#v", payload)
	}
	if payload["sub_task_timeout_secs"] != float64(0) {
		t.Fatalf("unexpected sub_task_timeout_secs payload: %#v", payload)
	}
	if payload["auto_apply"] != false {
		t.Fatalf("unexpected auto_apply payload: %#v", payload)
	}
	if payload["enable_propagation_machine"] != false {
		t.Fatalf("unexpected enable_propagation_machine payload: %#v", payload)
	}
}
