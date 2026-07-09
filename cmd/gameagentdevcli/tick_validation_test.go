package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateRequestedTicksForWorldRejectsFixedMode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"world_id": "world-1",
			"memory_limit": 50,
			"max_analysis_rounds": 5,
			"max_context_depth": 3,
			"auto_apply": true,
			"require_review_above": "critical",
			"pipeline_mode": "full",
			"world_time_settings": map[string]any{
				"tick_scale_mode": "fixed",
				"tick_min_unit": "时辰",
				"tick_step": 1,
			},
		})
	}))
	defer ts.Close()

	serverURL = ts.URL
	apiKey = "dev-key"
	requested := 2
	err := validateRequestedTicksForWorld("world-1", &requested)
	if err == nil {
		t.Fatal("expected fixed-mode validation error")
	}
}
