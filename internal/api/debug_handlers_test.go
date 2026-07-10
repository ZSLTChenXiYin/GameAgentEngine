package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
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

type callbackSequenceProvider struct {
	responses []string
	calls     int
}

type singleResponseProvider struct {
	response string
}

func (s *callbackSequenceProvider) Chat(systemPrompt string, messages []engine.ChatMessage) (*engine.LLMResult, error) {
	resp := s.responses[s.calls]
	s.calls++
	return &engine.LLMResult{Content: resp, Model: "callback-seq", Tokens: 6}, nil
}

func (s *callbackSequenceProvider) ModelName() string { return "callback-seq" }

func (s *singleResponseProvider) Chat(systemPrompt string, messages []engine.ChatMessage) (*engine.LLMResult, error) {
	return &engine.LLMResult{Content: s.response, Model: "single-response", Tokens: 5}, nil
}

func (s *singleResponseProvider) ModelName() string { return "single-response" }

func TestMakeActionCallbackHandlerAutoResumesPausedExecution(t *testing.T) {
	if err := store.Init("sqlite", "file:callback_resume_api?mode=memory&cache=shared"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world id: %v", err)
	}
	node := &store.NodeModel{UUID: store.NewUUID(), Name: "NPC", NodeType: "npc", WorldID: world.ID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	provider := &callbackSequenceProvider{responses: []string{
		`{"reply":"wait","request_data":{"label":"fetch-client","target":"game_client","queries":[{"type":"node_detail","node_id":"` + node.UUID + `"}]}}`,
		`{"reply":"resumed-final","action_calls":[],"memory_updates":[]}`,
	}}
	pipeline := engine.NewPipeline(provider)
	first, err := pipeline.Execute(&engine.InvokeRequest{WorldID: world.UUID, NodeID: node.UUID, TaskType: engine.TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	callbackID := first.ActionCalls[0].CallbackID
	runtimeTask, err := store.GetRuntimeTaskByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if runtimeTask.Status != store.RuntimeTaskStatusPending {
		t.Fatalf("expected pending runtime task before callback, got %q", runtimeTask.Status)
	}
	h := MakeActionCallbackHandler(pipeline)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/actions/callback", strings.NewReader(`{"callback_id":"`+callbackID+`","status":"success","result":{"scene":"tavern"}}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	resumed, ok := body["resumed"].(map[string]any)
	if !ok {
		t.Fatalf("expected resumed payload, got %+v", body)
	}
	if resumed["reply"] != "resumed-final" {
		t.Fatalf("unexpected resumed reply: %+v", resumed)
	}
	runtimeTask, err = store.GetRuntimeTaskByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get runtime task after callback: %v", err)
	}
	if runtimeTask.Status != store.RuntimeTaskStatusSucceeded {
		t.Fatalf("expected succeeded runtime task, got %q", runtimeTask.Status)
	}
	if runtimeTask.ResultJSON == "" {
		t.Fatalf("expected runtime task result payload, got %+v", runtimeTask)
	}
}

func TestMakeActionCallbackHandlerCompletesAsyncActionRuntimeTask(t *testing.T) {
	if err := store.Init("sqlite", "file:async_action_runtime_task_api?mode=memory&cache=shared"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world id: %v", err)
	}
	node := &store.NodeModel{UUID: store.NewUUID(), Name: "NPC", NodeType: "npc", WorldID: world.ID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	if _, err := store.UpsertWorldPolicy(world.UUID, nil, []string{"spawn_item"}); err != nil {
		t.Fatalf("policy: %v", err)
	}
	pipeline := engine.NewPipeline(&singleResponseProvider{response: `{"reply":"ok","action_calls":[{"action_id":"spawn_item","args":{"item_name":"potion","consumer":"bridge"}}],"memory_updates":[]}`})
	resp, err := pipeline.Execute(&engine.InvokeRequest{WorldID: world.UUID, NodeID: node.UUID, TaskType: engine.TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	callbackID := resp.ActionCalls[0].CallbackID
	runtimeTask, err := store.GetRuntimeTaskByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if runtimeTask.Status != store.RuntimeTaskStatusPending {
		t.Fatalf("expected pending runtime task, got %q", runtimeTask.Status)
	}
	h := MakeActionCallbackHandler(pipeline)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/actions/callback", strings.NewReader(`{"callback_id":"`+callbackID+`","status":"success","result":{"spawned":true}}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	runtimeTask, err = store.GetRuntimeTaskByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get runtime task after callback: %v", err)
	}
	if runtimeTask.Status != store.RuntimeTaskStatusSucceeded {
		t.Fatalf("expected succeeded runtime task, got %q", runtimeTask.Status)
	}
	if runtimeTask.ResultJSON == "" {
		t.Fatalf("expected runtime task result payload, got %+v", runtimeTask)
	}
}

func TestMakeActionCallbackHandlerSkipsAutoResumeWhenResumePolicyIsNone(t *testing.T) {
	if err := store.Init("sqlite", "file:callback_resume_policy_none?mode=memory&cache=shared"); err != nil {
		t.Fatalf("init db: %v", err)
	}
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world id: %v", err)
	}
	node := &store.NodeModel{UUID: store.NewUUID(), Name: "NPC", NodeType: "npc", WorldID: world.ID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	previousInterfaces := config.Global.ExternalInterfaces
	config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
		"game_client_request_data": {Category: "external_query", DeliveryMode: "pull", Consumer: "game_client", ResumePolicy: "none"},
	}
	defer func() { config.Global.ExternalInterfaces = previousInterfaces }()
	provider := &callbackSequenceProvider{responses: []string{
		`{"reply":"wait","request_data":{"label":"fetch-client","target":"game_client","queries":[{"type":"node_detail","node_id":"` + node.UUID + `"}]}}`,
		`{"reply":"should-not-run","action_calls":[],"memory_updates":[]}`,
	}}
	pipeline := engine.NewPipeline(provider)
	first, err := pipeline.Execute(&engine.InvokeRequest{WorldID: world.UUID, NodeID: node.UUID, TaskType: engine.TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	callbackID := first.ActionCalls[0].CallbackID
	h := MakeActionCallbackHandler(pipeline)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/actions/callback", strings.NewReader(`{"callback_id":"`+callbackID+`","status":"success","result":{"scene":"tavern"}}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if _, ok := body["resumed"]; ok {
		t.Fatalf("expected no resumed payload when resume_policy=none, got %+v", body)
	}
	paused, err := store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get paused execution: %v", err)
	}
	if paused.Status != "paused" {
		t.Fatalf("expected paused execution to remain paused, got %q", paused.Status)
	}
}
