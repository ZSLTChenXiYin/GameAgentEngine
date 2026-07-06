package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

func TestMakeDebugTracesHandlerReturnsPipelineObservabilityFields(t *testing.T) {
	engine.GlobalTraceRing = engine.NewTraceRing(16)
	engine.GlobalTraceRing.Push(&engine.Trace{
		ID:                     "t1",
		WorldID:                "world-1",
		RequestID:              "req-1",
		TaskType:               engine.TaskCustom,
		NodeID:                 "node-1",
		ConfiguredPipelineMode: "full",
		EffectivePipelineMode:  "polling",
		MaxAnalysisRounds:      4,
		RoundsUsed:             2,
		Timestamp:              time.Now(),
	})

	h := MakeDebugTracesHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/debug/traces?world_id=world-1&limit=5", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body struct {
		Traces []struct {
			ConfiguredPipelineMode string `json:"configured_pipeline_mode"`
			EffectivePipelineMode  string `json:"effective_pipeline_mode"`
			MaxAnalysisRounds      int    `json:"max_analysis_rounds"`
			RoundsUsed             int    `json:"rounds_used"`
		} `json:"traces"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Count != 1 || len(body.Traces) != 1 {
		t.Fatalf("expected one trace, got %+v", body)
	}
	trace := body.Traces[0]
	if trace.ConfiguredPipelineMode != "full" || trace.EffectivePipelineMode != "polling" {
		t.Fatalf("unexpected pipeline fields: %+v", trace)
	}
	if trace.MaxAnalysisRounds != 4 || trace.RoundsUsed != 2 {
		t.Fatalf("unexpected round fields: %+v", trace)
	}
}
