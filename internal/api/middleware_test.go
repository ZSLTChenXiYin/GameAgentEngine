package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type capturePlayerInputProvider struct {
	lastReq *engine.LLMChatRequest
}

func (c *capturePlayerInputProvider) Chat(req *engine.LLMChatRequest) (*engine.LLMResult, error) {
	c.lastReq = req
	return &engine.LLMResult{Content: `{"reply":"ok","action_calls":[],"memory_updates":[]}`, Model: "capture-player-input", Tokens: 5}, nil
}

func (c *capturePlayerInputProvider) ModelName() string { return "capture-player-input" }

func initMiddlewareTestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func TestIdempotencyMiddlewareReplaysMatchingRequest(t *testing.T) {
	initMiddlewareTestDB(t)
	count := 0
	h := IdempotencyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		count++
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "count": count})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":1}`))
	req1.Header.Set("Idempotency-Key", "same")
	w1 := httptest.NewRecorder()
	h(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":1}`))
	req2.Header.Set("Idempotency-Key", "same")
	w2 := httptest.NewRecorder()
	h(w2, req2)

	if count != 1 {
		t.Fatalf("expected handler to run once, got %d", count)
	}
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected replayed status 201, got %d", w2.Code)
	}
	if replayed := w2.Header().Get("X-Idempotency-Replayed"); replayed != "true" {
		t.Fatalf("expected replay header, got %q", replayed)
	}
}

func TestIdempotencyMiddlewareRejectsConflictingPayload(t *testing.T) {
	initMiddlewareTestDB(t)
	h := IdempotencyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":1}`))
	req1.Header.Set("Idempotency-Key", "same")
	w1 := httptest.NewRecorder()
	h(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":2}`))
	req2.Header.Set("Idempotency-Key", "same")
	w2 := httptest.NewRecorder()
	h(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "idempotency_key_conflict") {
		t.Fatalf("expected conflict code in body, got %s", w2.Body.String())
	}
}

func TestRequestAuthAllowsCallbackTokenOnCallbackEndpoint(t *testing.T) {
	h := RequestAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}), config.AuthConfig{APIKey: "dev-key", CallbackToken: "cb-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/actions/callback", nil)
	req.Header.Set("X-Callback-Token", "cb-token")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestRequestAuthAllowsRuntimeTaskTokenOnRuntimeTaskEndpoints(t *testing.T) {
	h := RequestAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}), config.AuthConfig{APIKey: "dev-key", RuntimeTaskToken: "rt-token"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtime/tasks/pending", nil)
	req.Header.Set("X-Runtime-Task-Token", "rt-token")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestCallbackReplayMiddlewareReplaysMatchingRequestID(t *testing.T) {
	initMiddlewareTestDB(t)
	previousAuth := config.Global.Auth
	config.Global.Auth = config.AuthConfig{}
	defer func() { config.Global.Auth = previousAuth }()
	count := 0
	h := CallbackReplayMiddleware(func(w http.ResponseWriter, r *http.Request) {
		count++
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "count": count})
	})
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/actions/callback", strings.NewReader(`{"callback_id":"cb-1","status":"success"}`))
	req1.Header.Set("X-Callback-Request-Id", "rid-1")
	w1 := httptest.NewRecorder()
	h(w1, req1)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/actions/callback", strings.NewReader(`{"callback_id":"cb-1","status":"success"}`))
	req2.Header.Set("X-Callback-Request-Id", "rid-1")
	w2 := httptest.NewRecorder()
	h(w2, req2)
	if count != 1 {
		t.Fatalf("expected handler to run once, got %d", count)
	}
	if replayed := w2.Header().Get("X-Callback-Replayed"); replayed != "true" {
		t.Fatalf("expected callback replay header, got %q", replayed)
	}
}

func TestCallbackReplayMiddlewareRequiresRequestIDWhenConfigured(t *testing.T) {
	initMiddlewareTestDB(t)
	previousAuth := config.Global.Auth
	config.Global.Auth = config.AuthConfig{CallbackRequireRequestID: true}
	defer func() { config.Global.Auth = previousAuth }()
	h := CallbackReplayMiddleware(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/actions/callback", strings.NewReader(`{"callback_id":"cb-1","status":"success"}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_callback_request_id") {
		t.Fatalf("expected invalid_callback_request_id, got %s", w.Body.String())
	}
}

func TestInvokeHandlerRejectsInvalidPipelineMode(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakeInvokeHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoke", strings.NewReader(`{
		"world_id":"w1",
		"node_id":"n1",
		"task_type":"custom",
		"context":{"pipeline_mode":"turbo"}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid_pipeline_mode") {
		t.Fatalf("expected invalid_pipeline_mode response, got %s", w.Body.String())
	}
}

func TestInvokeHandlerRejectsInvalidDynamicInterfaces(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakeInvokeHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoke", strings.NewReader(`{
		"world_id":"w1",
		"node_id":"n1",
		"task_type":"custom",
		"context":{
			"dynamic_interfaces":[
				{"id":"bad","kind":"data_request","external_interface":"game_client_request_data"}
			]
		}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid_dynamic_interfaces") {
		t.Fatalf("expected invalid_dynamic_interfaces response, got %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "query_types required") {
		t.Fatalf("expected query_types validation error, got %s", w.Body.String())
	}
}

func TestInvokeHandlerAcceptsValidDynamicInterfaces(t *testing.T) {
	initMiddlewareTestDB(t)
	world := &store.NodeModel{UUID: "w1", Name: "World", NodeType: "world", WorldUUID: "w1"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	node := &store.NodeModel{UUID: "n1", Name: "NPC", NodeType: "npc", WorldUUID: "w1", ParentUUID: &world.UUID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	h := MakeInvokeHandler(engine.NewPipeline(&singleResponseProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoke", strings.NewReader(`{
		"world_id":"w1",
		"node_id":"n1",
		"task_type":"custom",
		"context":{
			"dynamic_interfaces":[
				{"id":"scene_facts","kind":"data_request","external_interface":"game_client_request_data","query_types":["scene_state"]},
				{"id":"merchant_ops","kind":"action","external_interface":"npc_trade_action","max_calls":1}
			]
		}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected handler to pass validation and complete invoke, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "invalid_dynamic_interfaces") {
		t.Fatalf("did not expect invalid_dynamic_interfaces response, got %s", w.Body.String())
	}
}

func TestInvokeHandlerRejectsInvalidInteraction(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakeInvokeHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoke", strings.NewReader(`{
		"world_id":"w1",
		"node_id":"n1",
		"task_type":"custom",
		"context":{
			"interaction":{
				"mode":"group_chat",
				"speaker_node_id":"player_1",
				"target_node_id":"npc_1",
				"participant_node_ids":["player_1", ""]
			}
		}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid_interaction") {
		t.Fatalf("expected invalid_interaction response, got %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "participant_node_ids[1] required") {
		t.Fatalf("expected participant validation error, got %s", w.Body.String())
	}
}

func TestInvokeHandlerAcceptsValidInteraction(t *testing.T) {
	initMiddlewareTestDB(t)
	world := &store.NodeModel{UUID: "w1", Name: "World", NodeType: "world", WorldUUID: "w1"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	node := &store.NodeModel{UUID: "n1", Name: "NPC", NodeType: "npc", WorldUUID: "w1", ParentUUID: &world.UUID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	h := MakeInvokeHandler(engine.NewPipeline(&singleResponseProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoke", strings.NewReader(`{
		"world_id":"w1",
		"node_id":"n1",
		"task_type":"custom",
		"context":{
			"interaction":{
				"mode":"direct_dialogue",
				"speaker_node_id":"player_1",
				"target_node_id":"npc_1",
				"scene_node_id":"scene_inn",
				"participant_node_ids":["player_1", "npc_1"],
				"audience_scope":"public",
				"turn_index":2,
				"event":{"type":"speech"}
			}
		}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected handler to pass validation and complete invoke, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "invalid_interaction") {
		t.Fatalf("did not expect invalid_interaction response, got %s", w.Body.String())
	}
}

func TestInvokeHandlerRejectsNPCDialogueNodeTargetMismatch(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakeInvokeHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoke", strings.NewReader(`{
		"world_id":"w1",
		"node_id":"npc_a",
		"task_type":"npc_dialogue",
		"context":{
			"interaction":{
				"mode":"direct_dialogue",
				"speaker_node_id":"player_1",
				"target_node_id":"npc_b",
				"participant_node_ids":["player_1","npc_b"],
				"event":{"type":"speech"}
			}
		}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_interaction_contract") {
		t.Fatalf("expected invalid_interaction_contract, got %s", w.Body.String())
	}
}

func TestPlayerInputInterpretHandlerRejectsMissingFields(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakePlayerInputInterpretHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/player/input/interpret", strings.NewReader(`{
		"world_id":"w1",
		"message":"我把刀拍在柜台上"
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "player_node_id") {
		t.Fatalf("expected missing player_node_id error, got %s", w.Body.String())
	}
}

func TestPlayerInputInterpretHandlerAcceptsValidRequest(t *testing.T) {
	initMiddlewareTestDB(t)
	world := &store.NodeModel{UUID: "w1", Name: "World", NodeType: "world", WorldUUID: "w1"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	player := &store.NodeModel{UUID: "player_1", Name: "Player Mirror", NodeType: "npc", WorldUUID: "w1", ParentUUID: &world.UUID}
	if err := store.CreateNode(player); err != nil {
		t.Fatalf("create player node: %v", err)
	}
	provider := &capturePlayerInputProvider{}
	h := MakePlayerInputInterpretHandler(engine.NewPipeline(provider))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/player/input/interpret", strings.NewReader(`{
		"world_id":"w1",
		"player_node_id":"player_1",
		"target_node_id":"npc_1",
		"scene_node_id":"scene_inn",
		"message":"我把刀拍在柜台上"
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "invalid_interaction") {
		t.Fatalf("did not expect invalid_interaction, got %s", w.Body.String())
	}
	if provider.lastReq == nil {
		t.Fatal("expected provider request to be captured")
	}
	if !strings.Contains(provider.lastReq.SystemPrompt, "行为意图提案") {
		t.Fatalf("expected player intent prompt, got %s", provider.lastReq.SystemPrompt)
	}
	if len(provider.lastReq.Messages) != 1 || strings.Contains(provider.lastReq.Messages[0].Content, "[player_input_interpret]") {
		t.Fatalf("expected sanitized player message, got %#v", provider.lastReq.Messages)
	}
	var body struct {
		PlayerIntent *json.RawMessage `json:"player_intent,omitempty"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
}

func TestPlayerInputInterpretHandlerRejectsSpeakerNodeMismatchInExplicitInteraction(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakePlayerInputInterpretHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/player/input/interpret", strings.NewReader(`{
		"world_id":"w1",
		"player_node_id":"player_1",
		"message":"hello",
		"context":{
			"interaction":{
				"mode":"direct_dialogue",
				"speaker_node_id":"player_2",
				"target_node_id":"npc_1",
				"participant_node_ids":["player_2","npc_1"],
				"event":{"type":"speech"}
			}
		}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_interaction_contract") {
		t.Fatalf("expected invalid_interaction_contract, got %s", w.Body.String())
	}
}

func TestExecuteInteractionHandlerRejectsMissingFields(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakeExecuteInteractionHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/interactions/execute", strings.NewReader(`{
		"world_id":"w1",
		"message":"hello"
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "actor_node_id") {
		t.Fatalf("expected missing actor_node_id error, got %s", w.Body.String())
	}
}

func TestExecuteInteractionHandlerAcceptsValidRequest(t *testing.T) {
	initMiddlewareTestDB(t)
	world := &store.NodeModel{UUID: "w1", Name: "World", NodeType: "world", WorldUUID: "w1"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	actor := &store.NodeModel{UUID: "player_1", Name: "Player", NodeType: "npc", WorldUUID: "w1", ParentUUID: &world.UUID}
	if err := store.CreateNode(actor); err != nil {
		t.Fatalf("create actor node: %v", err)
	}
	target := &store.NodeModel{UUID: "npc_1", Name: "Innkeeper", NodeType: "npc", WorldUUID: "w1", ParentUUID: &world.UUID}
	if err := store.CreateNode(target); err != nil {
		t.Fatalf("create target node: %v", err)
	}
	h := MakeExecuteInteractionHandler(engine.NewPipeline(&singleResponseProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/interactions/execute", strings.NewReader(`{
		"world_id":"w1",
		"actor_node_id":"player_1",
		"target_node_id":"npc_1",
		"scene_node_id":"scene_inn",
		"message":"老板，今晚见过这把刀的主人吗？",
		"participant_node_ids":["player_1","npc_1"],
		"mode":"direct_dialogue",
		"audience_scope":"private",
		"turn_index":3,
		"event":{"type":"speech"}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "invalid_interaction") {
		t.Fatalf("did not expect invalid_interaction, got %s", w.Body.String())
	}
}

func TestBuildCanonicalInteractionContextDefaults(t *testing.T) {
	ctx, err := buildCanonicalInteractionContext(interactionContractInput{
		ActorNodeID:           "player_1",
		ParticipantNodeIDs:    []string{"player_1", "npc_1", "npc_2"},
		FallbackTargetToActor: true,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Mode != "group_chat" {
		t.Fatalf("expected group_chat mode, got %#v", ctx)
	}
	if ctx.AudienceScope != "public" {
		t.Fatalf("expected public audience, got %#v", ctx)
	}
	if ctx.TargetNodeID != "player_1" {
		t.Fatalf("expected fallback target to actor, got %#v", ctx)
	}
	if ctx.Event == nil || ctx.Event.Type != "speech" {
		t.Fatalf("expected default speech event, got %#v", ctx)
	}
	if got := ctx.Event.Args["input_source"]; got != nil {
		t.Fatalf("did not expect input_source by default, got %#v", ctx.Event.Args)
	}
}

func TestBuildCanonicalInteractionContextAddsInputSource(t *testing.T) {
	ctx, err := buildCanonicalInteractionContext(interactionContractInput{
		ActorNodeID:           "player_1",
		TargetNodeID:          "npc_1",
		InputSource:           "player_input_interpret",
		FallbackTargetToActor: true,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ctx.Event == nil {
		t.Fatalf("expected event, got %#v", ctx)
	}
	if got := ctx.Event.Args["input_source"]; got != "player_input_interpret" {
		t.Fatalf("expected input_source to be injected, got %#v", ctx.Event.Args)
	}
}

func TestBuildCanonicalInteractionContextRejectsExplicitSpeakerConflict(t *testing.T) {
	_, err := buildCanonicalInteractionContext(interactionContractInput{
		ActorNodeID:           "player_1",
		FallbackTargetToActor: true,
	}, &engine.InteractionContext{
		SpeakerNodeID: "player_2",
	})
	if err == nil {
		t.Fatal("expected explicit speaker conflict error")
	}
	if !strings.Contains(err.Error(), "interaction.speaker_node_id conflicts") {
		t.Fatalf("unexpected error: %v", err)
	}
}
