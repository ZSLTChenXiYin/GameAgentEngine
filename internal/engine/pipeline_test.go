package engine

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"sync"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type stubProvider struct {
	response string
	err      error
	calls    int
}

type barrierProvider struct {
	response string
	ready    chan struct{}
	release  chan struct{}
	mu       sync.Mutex
	calls    int
}

type captureProvider struct {
	response      string
	lastPrompt    string
	lastMsgs      []ChatMessage
	lastTools     []LLMToolDefinition
	supportsTools bool
	toolsSet      bool
}

type sequenceProvider struct {
	responses []string
	calls     int
	lastPrompt string
}

func (s *stubProvider) Chat(req *LLMChatRequest) (*LLMResult, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return &LLMResult{Content: s.response, Model: "stub", Tokens: 7}, nil
}

func (s *stubProvider) ModelName() string { return "stub" }

func (b *barrierProvider) Chat(req *LLMChatRequest) (*LLMResult, error) {
	b.mu.Lock()
	b.calls++
	current := b.calls
	b.mu.Unlock()
	if current == 2 {
		close(b.ready)
	}
	<-b.release
	return &LLMResult{Content: b.response, Model: "barrier", Tokens: 9}, nil
}

func (b *barrierProvider) ModelName() string { return "barrier" }

func (c *captureProvider) Chat(req *LLMChatRequest) (*LLMResult, error) {
	if req != nil {
		c.lastPrompt = req.SystemPrompt
		c.lastMsgs = req.Messages
		c.lastTools = req.Tools
	}
	return &LLMResult{Content: c.response, Model: "capture", Tokens: 11}, nil
}

func (c *captureProvider) ModelName() string { return "capture" }

func (c *captureProvider) SupportsStructuredTools() bool {
	if !c.toolsSet {
		return true
	}
	return c.supportsTools
}

func (s *sequenceProvider) Chat(req *LLMChatRequest) (*LLMResult, error) {
	if req != nil {
		s.lastPrompt = req.SystemPrompt
	}
	if s.calls >= len(s.responses) {
		return &LLMResult{Content: `{"reply":"done","action_calls":[],"memory_updates":[]}`, Model: "sequence", Tokens: 5}, nil
	}
	resp := s.responses[s.calls]
	s.calls++
	return &LLMResult{Content: resp, Model: "sequence", Tokens: 5}, nil
}

func (s *sequenceProvider) ModelName() string { return "sequence" }

func initTestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func createWorldAndNode(t *testing.T) (string, string) {
	t.Helper()
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("update world id: %v", err)
	}
	node := &store.NodeModel{UUID: store.NewUUID(), Name: "NPC", NodeType: "npc", WorldID: world.ID}
	if err := store.CreateNode(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	return world.UUID, node.UUID
}

func TestExecuteVerticalRespectsPipelineModeAndSkipsExtraRounds(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{PipelineMode: "vertical", MaxAnalysisRounds: 3}, &store.WorldSettingsUpdateMask{PipelineMode: true, MaxAnalysisRounds: true}); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.Reply != "ok" {
		t.Fatalf("unexpected reply: %q", resp.Reply)
	}
	if provider.calls != 1 {
		t.Fatalf("expected 1 llm call in vertical mode, got %d", provider.calls)
	}
	if resp.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if resp.Metadata.ConfiguredPipelineMode != "vertical" {
		t.Fatalf("expected configured pipeline mode vertical, got %q", resp.Metadata.ConfiguredPipelineMode)
	}
	if resp.Metadata.EffectivePipelineMode != "vertical" {
		t.Fatalf("expected effective pipeline mode vertical, got %q", resp.Metadata.EffectivePipelineMode)
	}
	if resp.Metadata.MaxAnalysisRounds != 3 {
		t.Fatalf("expected max analysis rounds 3, got %d", resp.Metadata.MaxAnalysisRounds)
	}
	if resp.Metadata.RoundsUsed != 1 {
		t.Fatalf("expected rounds used 1, got %d", resp.Metadata.RoundsUsed)
	}
}

func TestExecuteReturnsLLMErrorInsteadOfPanicking(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	pipeline := NewPipeline(&stubProvider{err: fmt.Errorf("boom")})

	_, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "llm chat: boom" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestExecuteKeepsWorldPoliciesIsolatedAcrossConcurrentRequests(t *testing.T) {
	initTestDB(t)
	worldA, nodeA := createWorldAndNode(t)
	worldB, nodeB := createWorldAndNode(t)
	if _, err := store.UpsertWorldPolicy(worldA, []string{"spawn_item"}, nil); err != nil {
		t.Fatalf("policy A: %v", err)
	}
	if _, err := store.UpsertWorldPolicy(worldB, nil, nil); err != nil {
		t.Fatalf("policy B: %v", err)
	}

	provider := &barrierProvider{
		response: `{"reply":"ok","action_calls":[{"action_id":"spawn_item","args":{}}],"memory_updates":[]}`,
		ready:    make(chan struct{}),
		release:  make(chan struct{}),
	}
	pipeline := NewPipeline(provider)

	type result struct {
		resp *InvokeResponse
		err  error
	}
	results := make(chan result, 2)
	go func() {
		resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldA, NodeID: nodeA, TaskType: TaskCustom})
		results <- result{resp: resp, err: err}
	}()
	go func() {
		resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldB, NodeID: nodeB, TaskType: TaskCustom})
		results <- result{resp: resp, err: err}
	}()

	<-provider.ready
	close(provider.release)

	var got []*InvokeResponse
	for i := 0; i < 2; i++ {
		res := <-results
		if res.err != nil {
			t.Fatalf("execute: %v", res.err)
		}
		got = append(got, res.resp)
	}

	allowedCount := 0
	blockedCount := 0
	for _, resp := range got {
		if len(resp.ActionCalls) == 0 {
			blockedCount++
			continue
		}
		if len(resp.ActionCalls) == 1 && resp.ActionCalls[0].ActionID == "spawn_item" {
			allowedCount++
		}
	}
	if allowedCount != 1 || blockedCount != 1 {
		t.Fatalf("expected one allowed and one blocked response, got allowed=%d blocked=%d", allowedCount, blockedCount)
	}
}

func TestExecuteDebugTraceIncludesPipelineObservability(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{PipelineMode: "full", MaxAnalysisRounds: 4}, &store.WorldSettingsUpdateMask{PipelineMode: true, MaxAnalysisRounds: true}); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	previousMode := config.Global.Engine.ExecutionMode
	config.Global.Engine.ExecutionMode = "debug"
	defer func() { config.Global.Engine.ExecutionMode = previousMode }()
	GlobalTraceRing = NewTraceRing(1000)

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context:  &InvokeContext{PipelineMode: PipelinePolling},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if resp.Metadata.ConfiguredPipelineMode != "full" {
		t.Fatalf("expected configured pipeline mode full, got %q", resp.Metadata.ConfiguredPipelineMode)
	}
	if resp.Metadata.EffectivePipelineMode != "polling" {
		t.Fatalf("expected effective pipeline mode polling, got %q", resp.Metadata.EffectivePipelineMode)
	}

	traces := GlobalTraceRing.List(10)
	if len(traces) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(traces))
	}
	trace := traces[0]
	if trace.ConfiguredPipelineMode != "full" {
		t.Fatalf("expected trace configured pipeline mode full, got %q", trace.ConfiguredPipelineMode)
	}
	if trace.EffectivePipelineMode != "polling" {
		t.Fatalf("expected trace effective pipeline mode polling, got %q", trace.EffectivePipelineMode)
	}
	if trace.MaxAnalysisRounds != 4 {
		t.Fatalf("expected trace max analysis rounds 4, got %d", trace.MaxAnalysisRounds)
	}
	if trace.RoundsUsed != 1 {
		t.Fatalf("expected trace rounds used 1, got %d", trace.RoundsUsed)
	}
}

func TestExecutePersistsStructuredPipelineLogs(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp == nil || resp.Metadata == nil {
		t.Fatal("expected response metadata")
	}

	logs, err := store.GetInferenceLogs(worldID, 10, 0, string(TaskCustom))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	if len(logs) < 3 {
		t.Fatalf("expected at least 3 pipeline logs, got %d", len(logs))
	}
	seen := map[string]store.InferenceLogModel{}
	for _, item := range logs {
		seen[item.EventName] = item
	}
	if _, ok := seen["request_started"]; !ok {
		t.Fatalf("expected request_started log, got %#v", seen)
	}
	if _, ok := seen["context_built"]; !ok {
		t.Fatalf("expected context_built log, got %#v", seen)
	}
	completed, ok := seen["response_completed"]
	if !ok {
		t.Fatalf("expected response_completed log, got %#v", seen)
	}
	if completed.Category != "pipeline" {
		t.Fatalf("expected pipeline category, got %q", completed.Category)
	}
	if completed.ExecutionMode != string(ModeProduction) {
		t.Fatalf("expected production execution mode, got %q", completed.ExecutionMode)
	}
	if completed.ResponseData == "" {
		t.Fatal("expected response data in completed log")
	}
}

func TestExecuteDebugModePersistsFullRoundDetails(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	previousMode := config.Global.Engine.ExecutionMode
	config.Global.Engine.ExecutionMode = "debug"
	defer func() { config.Global.Engine.ExecutionMode = previousMode }()

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom}); err != nil {
		t.Fatalf("execute: %v", err)
	}

	logs, err := store.GetInferenceLogs(worldID, 20, 0, string(TaskCustom))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	var found bool
	for _, item := range logs {
		if item.EventName == "llm_response_received" {
			found = true
			if item.DetailData == "" {
				t.Fatal("expected full detail data in debug mode")
			}
			if !strings.Contains(item.DetailData, "raw_response") {
				t.Fatalf("expected raw_response in detail data, got %s", item.DetailData)
			}
		}
	}
	if !found {
		t.Fatal("expected llm_response_received log in debug mode")
	}
}

func TestExecuteDebugModeLogsStructuredToolNegotiationDetails(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	previousMode := config.Global.Engine.ExecutionMode
	config.Global.Engine.ExecutionMode = "debug"
	defer func() { config.Global.Engine.ExecutionMode = previousMode }()

	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`, supportsTools: false, toolsSet: true}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskNPCDialogue,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"node_detail"},
		}}},
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}

	logs, err := store.GetInferenceLogs(worldID, 20, 0, string(TaskNPCDialogue))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	var found bool
	for _, item := range logs {
		if item.EventName == "prompt_prepared" {
			found = true
			if !strings.Contains(item.DetailData, "provider_supports_tools") {
				t.Fatalf("expected provider_supports_tools in detail data, got %s", item.DetailData)
			}
			if !strings.Contains(item.DetailData, "planned_tools") {
				t.Fatalf("expected planned_tools in detail data, got %s", item.DetailData)
			}
			if !strings.Contains(item.DetailData, "exposed_tools") {
				t.Fatalf("expected exposed_tools in detail data, got %s", item.DetailData)
			}
		}
	}
	if !found {
		t.Fatal("expected prompt_prepared log in debug mode")
	}
}

func TestExecuteReviewModePersistsPendingPlanLog(t *testing.T) {
	initTestDB(t)
	worldID, _ := createWorldAndNode(t)
	previousMode := config.Global.Engine.ExecutionMode
	config.Global.Engine.ExecutionMode = "review"
	defer func() { config.Global.Engine.ExecutionMode = previousMode }()
	GlobalPlanReview = NewPlanReviewStore()

	provider := &stubProvider{response: `{"reply":"plan","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"major","summary":"需要审批","world_events":[],"proposed_actions":[]}}`}
	pipeline := NewPipeline(provider)
	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: worldID, TaskType: TaskWorldTick})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.ExecutionMode != ModeReview {
		t.Fatalf("expected review execution mode, got %q", resp.ExecutionMode)
	}

	logs, err := store.GetInferenceLogs(worldID, 20, 0, string(TaskWorldTick))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	var found bool
	for _, item := range logs {
		if item.EventName == "plan_pending_review" {
			found = true
			if item.DetailData == "" {
				t.Fatal("expected pending plan detail data")
			}
		}
	}
	if !found {
		t.Fatal("expected plan_pending_review log")
	}
}

func TestExecuteWorldTickIncludesPersistentContinuityState(t *testing.T) {
	initTestDB(t)
	worldID, _ := createWorldAndNode(t)
	world := store.ResolveNodeUUID(worldID)
	if world == 0 {
		t.Fatal("expected world id")
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: world, NodeUUID: worldID, ComponentType: string(CompStoryState), Data: `{"current_situation":"地下52米量子谐振腔已经暴露"}`}); err != nil {
		t.Fatalf("create story_state: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: world, NodeUUID: worldID, ComponentType: string(CompTickPolicy), Data: `{"continuity_rules":["保持地点和关键设施连续"]}`}); err != nil {
		t.Fatalf("create tick_policy: %v", err)
	}
	provider := &captureProvider{response: `{"reply":"tick","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"推进","world_events":[],"proposed_actions":[]}}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: worldID, TaskType: TaskWorldTick}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(provider.lastPrompt, "地下52米量子谐振腔已经暴露") {
		t.Fatalf("expected story state in prompt, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, "保持地点和关键设施连续") {
		t.Fatalf("expected tick policy in prompt, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, "不要无故重置") {
		t.Fatalf("expected continuity guard in prompt, got %s", provider.lastPrompt)
	}
}

func TestExecuteDialogueUsesLocatedAtEnvironmentChain(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	if worldInt == 0 || nodeInt == 0 {
		t.Fatal("expected resolved world and node ids")
	}
	location := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Tavern", NodeType: string(NodeTypeLocation), ParentID: &worldInt}
	if err := store.CreateNode(location); err != nil {
		t.Fatalf("create location: %v", err)
	}
	store.ResolveNodeParentUUID(location)
	if err := store.CreateComponent(&store.ComponentModel{NodeID: location.ID, NodeUUID: location.UUID, ComponentType: string(CompLore), Data: `{"lighting":"warm lanterns"}`}); err != nil {
		t.Fatalf("create location component: %v", err)
	}
	if err := store.CreateMemory(&store.MemoryModel{NodeID: location.ID, NodeUUID: location.UUID, Content: "酒馆里弥漫着麦酒和木烟味。", Level: string(MemShared)}); err != nil {
		t.Fatalf("create location memory: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: location.ID, TargetUUID: location.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create located_at relation: %v", err)
	}

	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskNPCDialogue}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(provider.lastPrompt, "当前环境链：Tavern(location) > World(world)") {
		t.Fatalf("expected environment chain in prompt, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, "warm lanterns") {
		t.Fatalf("expected location component in prompt, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, "酒馆里弥漫着麦酒和木烟味。") {
		t.Fatalf("expected location memory in prompt, got %s", provider.lastPrompt)
	}
}

func TestExecuteDialogueWithoutLocatedAtDoesNotInjectEnvironmentChain(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskNPCDialogue}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if strings.Contains(provider.lastPrompt, "当前环境链：") {
		t.Fatalf("did not expect environment chain in prompt, got %s", provider.lastPrompt)
	}
	if strings.Contains(provider.lastPrompt, "环境信息：") {
		t.Fatalf("did not expect environment block in prompt, got %s", provider.lastPrompt)
	}
}

func TestExecuteDialogueInjectsDynamicInterfacePromptBlock(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskNPCDialogue,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"scene_state", "visible_entities"},
			MaxQueries:        2,
		}}},
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"Dynamic Interfaces", "scene_facts: Query visible scene state", "query_types: scene_state, visible_entities", "max_queries: 2"} {
		if !strings.Contains(provider.lastPrompt, want) {
			t.Fatalf("expected prompt to contain %q, got %s", want, provider.lastPrompt)
		}
	}
}

func TestExecuteAutonomousActInjectsDynamicInterfacePromptBlock(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	nodeInt := store.ResolveNodeUUID(nodeID)
	if err := store.CreateComponent(&store.ComponentModel{NodeID: nodeInt, NodeUUID: nodeID, ComponentType: string(CompAutonomous), Data: `{"enabled":true,"trigger":"manual","capabilities":[{"id":"send_dialogue"}]}`}); err != nil {
		t.Fatalf("create autonomous component: %v", err)
	}

	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskAutonomousAct,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "merchant_ops",
			Kind:              DynamicInterfaceAction,
			ExternalInterface: "npc_trade_action",
			Description:       "Perform trade-related external actions",
			MaxCalls:          1,
		}}},
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"Dynamic Interfaces", "Action interfaces:", "merchant_ops: Perform trade-related external actions", "max_calls: 1"} {
		if !strings.Contains(provider.lastPrompt, want) {
			t.Fatalf("expected prompt to contain %q, got %s", want, provider.lastPrompt)
		}
	}
}

func TestExecuteDialogueBuildsStructuredToolsFromDynamicInterfaces(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskNPCDialogue,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{
			{
				ID:                "scene_facts",
				Kind:              DynamicInterfaceDataRequest,
				ExternalInterface: "game_client_request_data",
				Description:       "Query visible scene state",
				QueryTypes:        []string{"node_detail", "visible_entities"},
				MaxQueries:        2,
			},
			{
				ID:                "merchant_ops",
				Kind:              DynamicInterfaceAction,
				ExternalInterface: "npc_trade_action",
				Description:       "Perform trade-related external actions",
				ArgsSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"intent": map[string]any{"type": "string"},
					},
				},
				MaxCalls: 1,
			},
		}},
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	seen := map[string]LLMToolDefinition{}
	for _, tool := range provider.lastTools {
		seen[tool.Name] = tool
	}
	dataTool, ok := seen["scene_facts"]
	if !ok {
		t.Fatalf("expected scene_facts tool in %+v", provider.lastTools)
	}
	if dataTool.Invocation != LLMToolInvocationDataRequest || dataTool.DataRequest == nil {
		t.Fatalf("expected scene_facts to be data_request, got %+v", dataTool)
	}
	if dataTool.DataRequest.ExternalInterface != "game_client_request_data" {
		t.Fatalf("expected normalized external interface, got %+v", dataTool.DataRequest)
	}
	actionTool, ok := seen["merchant_ops"]
	if !ok {
		t.Fatalf("expected merchant_ops tool in %+v", provider.lastTools)
	}
	if actionTool.Invocation != LLMToolInvocationAction || actionTool.ActionID != "merchant_ops" {
		t.Fatalf("expected merchant_ops to be action tool, got %+v", actionTool)
	}
	if actionTool.Parameters["type"] != "object" {
		t.Fatalf("expected action args schema to survive, got %+v", actionTool.Parameters)
	}
}

func TestExecuteDialogueIncludesBuiltinToolsAlongsideDynamicInterfaces(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskNPCDialogue,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"node_detail"},
		}}},
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	seen := map[string]LLMToolDefinition{}
	for _, tool := range provider.lastTools {
		seen[tool.Name] = tool
	}
	for _, name := range []string{"request_store_data", "add_memory", "update_mood", "send_dialogue", "adjust_relation", "spawn_item", "scene_facts"} {
		if _, ok := seen[name]; !ok {
			t.Fatalf("expected tool %q in %+v", name, provider.lastTools)
		}
	}
}

func TestExecuteAutonomousIncludesOnlyCapabilityBuiltinTools(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	nodeInt := store.ResolveNodeUUID(nodeID)
	if err := store.CreateComponent(&store.ComponentModel{NodeID: nodeInt, NodeUUID: nodeID, ComponentType: string(CompAutonomous), Data: `{"enabled":true,"trigger":"manual","capabilities":[{"id":"send_dialogue"}]}`}); err != nil {
		t.Fatalf("create autonomous component: %v", err)
	}
	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskAutonomousAct}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	seen := map[string]struct{}{}
	for _, tool := range provider.lastTools {
		seen[tool.Name] = struct{}{}
	}
	if _, ok := seen["request_store_data"]; !ok {
		t.Fatalf("expected request_store_data tool, got %+v", provider.lastTools)
	}
	if _, ok := seen["send_dialogue"]; !ok {
		t.Fatalf("expected send_dialogue tool, got %+v", provider.lastTools)
	}
	for _, blocked := range []string{"add_memory", "update_mood", "adjust_relation", "spawn_item"} {
		if _, ok := seen[blocked]; ok {
			t.Fatalf("did not expect autonomous tool %q in %+v", blocked, provider.lastTools)
		}
	}
}

func TestExecuteDialogueFallsBackWhenProviderDoesNotSupportStructuredTools(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`, supportsTools: false, toolsSet: true}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskNPCDialogue,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"node_detail"},
		}}},
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if provider.lastTools != nil {
		t.Fatalf("expected tools to be omitted when provider lacks support, got %+v", provider.lastTools)
	}
	if !strings.Contains(provider.lastPrompt, "Dynamic Interfaces") {
		t.Fatalf("expected prompt fallback to retain dynamic interface instructions, got %s", provider.lastPrompt)
	}
}

func TestExecuteDialogueIncludesStructuredToolsWhenProviderSupportsThem(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`, supportsTools: true, toolsSet: true}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskNPCDialogue,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"node_detail"},
		}}},
	}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(provider.lastTools) == 0 {
		t.Fatalf("expected structured tools to be forwarded, got %+v", provider.lastTools)
	}
}

func TestHandleDataRequestFiltersNodeRelationsByRelationType(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	ally := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Ally", NodeType: string(NodeTypeNPC)}
	location := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Square", NodeType: string(NodeTypeLocation)}
	if err := store.CreateNode(ally); err != nil {
		t.Fatalf("create ally: %v", err)
	}
	if err := store.CreateNode(location); err != nil {
		t.Fatalf("create location: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: ally.ID, TargetUUID: ally.UUID, RelationType: string(RelAlly), Weight: 2}); err != nil {
		t.Fatalf("create ally relation: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: location.ID, TargetUUID: location.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create located_at relation: %v", err)
	}

	pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`})
	result := pipeline.handleDataRequest(nil, &DataRequest{Queries: []DataQuery{{Type: "node_relations", NodeID: nodeID, Filter: string(RelLocatedAt)}}})
	if !strings.Contains(result, string(RelLocatedAt)) {
		t.Fatalf("expected located_at relation, got %q", result)
	}
	if strings.Contains(result, string(RelAlly)) {
		t.Fatalf("did not expect ally relation, got %q", result)
	}
}

func TestHandleDataRequestFiltersNodeMemoriesByLevel(t *testing.T) {
	initTestDB(t)
	_, nodeID := createWorldAndNode(t)
	nodeInt := store.ResolveNodeUUID(nodeID)
	if err := store.CreateMemory(&store.MemoryModel{NodeID: nodeInt, NodeUUID: nodeID, Content: "short memory", Level: string(MemShortTerm)}); err != nil {
		t.Fatalf("create short memory: %v", err)
	}
	if err := store.CreateMemory(&store.MemoryModel{NodeID: nodeInt, NodeUUID: nodeID, Content: "shared memory", Level: string(MemShared)}); err != nil {
		t.Fatalf("create shared memory: %v", err)
	}

	pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`})
	result := pipeline.handleDataRequest(nil, &DataRequest{Queries: []DataQuery{{Type: "node_memories", NodeID: nodeID, Filter: string(MemShared)}}})
	if !strings.Contains(result, "shared memory") {
		t.Fatalf("expected shared memory, got %q", result)
	}
	if strings.Contains(result, "short memory") {
		t.Fatalf("did not expect short memory, got %q", result)
	}
}

func TestResumePausedExecutionContinuesAfterGameClientCallback(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"after-resume","action_calls":[],"memory_updates":[]}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first.DataRequest == nil {
		t.Fatal("expected paused data request")
	}
	if len(first.ActionCalls) != 1 || first.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected callback action, got %+v", first.ActionCalls)
	}
	callbackID := first.ActionCalls[0].CallbackID
	runtimeTask, err := store.GetRuntimeTaskByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get runtime task by callback: %v", err)
	}
	if runtimeTask.Status != store.RuntimeTaskStatusPending {
		t.Fatalf("expected pending runtime task, got %q", runtimeTask.Status)
	}
	if runtimeTask.Consumer != "game_client" || runtimeTask.InterfaceName != "game_client_request_data" {
		t.Fatalf("unexpected runtime task routing: %+v", runtimeTask)
	}

	resumed, err := pipeline.ResumePausedExecution(callbackID, map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed == nil || resumed.Reply != "after-resume" {
		t.Fatalf("unexpected resumed response: %+v", resumed)
	}
	paused, err := store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get paused execution: %v", err)
	}
	if paused.Status != "completed" {
		t.Fatalf("expected completed paused execution, got %q", paused.Status)
	}
	record, err := store.GetAsyncCallbackRecord(callbackID)
	if err != nil {
		t.Fatalf("get callback record: %v", err)
	}
	if record.Status != "pending" && record.Status != "success" && record.Status != "completed" && record.Status != "ok" {
		t.Fatalf("unexpected callback status: %q", record.Status)
	}
	if provider.calls != 2 {
		t.Fatalf("expected 2 llm calls, got %d", provider.calls)
	}
	logs, err := store.GetInferenceLogs(worldID, 50, 0, string(TaskCustom))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	var foundResume bool
	for _, item := range logs {
		if item.EventName == "resume_completed" {
			foundResume = true
			break
		}
	}
	if !foundResume {
		t.Fatal("expected resume_completed log")
	}
}

func TestResumePausedExecutionContinuesAfterSubTaskGameClientCallback(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"append"}]}`,
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"child-after-resume","action_calls":[],"memory_updates":[]}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom, Context: &InvokeContext{PipelineMode: PipelineFull}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first.DataRequest == nil {
		t.Fatal("expected paused data request from sub-task")
	}
	if len(first.ActionCalls) != 1 || first.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected callback action, got %+v", first.ActionCalls)
	}
	callbackID := first.ActionCalls[0].CallbackID

	paused, err := store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get paused execution: %v", err)
	}
	if paused.RequestID == "" || paused.WorldUUID != worldID {
		t.Fatalf("expected parent paused execution snapshot, got %+v", paused)
	}

	resumed, err := pipeline.ResumePausedExecution(callbackID, map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed == nil {
		t.Fatal("expected resumed response")
	}
	if !strings.Contains(resumed.Reply, "child-after-resume") {
		t.Fatalf("expected resumed reply to include child result, got %+v", resumed)
	}
	if !strings.Contains(provider.lastPrompt, `"scene":"tavern"`) {
		t.Fatalf("expected resumed child prompt to include callback payload, got %s", provider.lastPrompt)
	}
	if len(resumed.SubTasks) != 1 || resumed.SubTasks[0].Label != "fetch_scene" {
		t.Fatalf("expected resumed sub-task metadata, got %+v", resumed.SubTasks)
	}
	paused, err = store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get paused execution after resume: %v", err)
	}
	if paused.Status != "completed" {
		t.Fatalf("expected completed paused execution, got %q", paused.Status)
	}
	if provider.calls != 3 {
		t.Fatalf("expected 3 llm calls, got %d", provider.calls)
	}
}

func TestResumePausedExecutionSupportsRepeatedSubTaskGameClientCallbacks(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"append"}]}`,
		`{"reply":"need-client-1","request_data":{"label":"fetch-scene-1","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"need-client-2","request_data":{"label":"fetch-scene-2","target":"game_client","queries":[{"type":"node_relations","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"child-after-second-resume","action_calls":[],"memory_updates":[]}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom, Context: &InvokeContext{PipelineMode: PipelineFull}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first.DataRequest == nil || len(first.ActionCalls) != 1 || first.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected first paused sub-task request, got %+v", first)
	}
	firstCallbackID := first.ActionCalls[0].CallbackID

	second, err := pipeline.ResumePausedExecution(firstCallbackID, map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("first resume: %v", err)
	}
	if second == nil || second.DataRequest == nil || len(second.ActionCalls) != 1 || second.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected second paused sub-task request, got %+v", second)
	}
	secondCallbackID := second.ActionCalls[0].CallbackID
	if secondCallbackID == firstCallbackID {
		t.Fatalf("expected a new callback id after repeated sub-task pause, got %q", secondCallbackID)
	}

	resumed, err := pipeline.ResumePausedExecution(secondCallbackID, map[string]any{"relations": []string{"guard", "merchant"}})
	if err != nil {
		t.Fatalf("second resume: %v", err)
	}
	if resumed == nil || !strings.Contains(resumed.Reply, "child-after-second-resume") {
		t.Fatalf("expected final repeated sub-task reply, got %+v", resumed)
	}
	if !strings.Contains(provider.lastPrompt, `"relations":["guard","merchant"]`) {
		t.Fatalf("expected repeated sub-task prompt to include second callback payload, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, `"scene":"tavern"`) {
		t.Fatalf("expected repeated sub-task prompt to retain first callback payload, got %s", provider.lastPrompt)
	}
	if provider.calls != 4 {
		t.Fatalf("expected 4 llm calls, got %d", provider.calls)
	}
	paused, err := store.GetPausedExecutionByCallbackID(secondCallbackID)
	if err != nil {
		t.Fatalf("get second paused execution: %v", err)
	}
	if paused.Status != "completed" {
		t.Fatalf("expected completed paused execution after second sub-task resume, got %q", paused.Status)
	}
}

func TestResumePausedExecutionReusesResolvedDataRequestInsteadOfRequerying(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"still-thinking","request_data":{"label":"fetch-scene","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"final-after-reuse","action_calls":[],"memory_updates":[]}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first.DataRequest == nil || len(first.ActionCalls) != 1 {
		t.Fatalf("expected first data request pause, got %+v", first)
	}
	callbackID := first.ActionCalls[0].CallbackID

	resumed, err := pipeline.ResumePausedExecution(callbackID, map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed == nil || resumed.Reply != "final-after-reuse" {
		t.Fatalf("unexpected resumed response: %+v", resumed)
	}
	if provider.calls != 3 {
		t.Fatalf("expected 3 llm calls, got %d", provider.calls)
	}
	tasks, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{WorldUUID: worldID, Limit: 20})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected only one runtime task, got %+v", tasks)
	}
	logs, err := store.GetInferenceLogs(worldID, 50, 0, string(TaskCustom))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	var foundReuse bool
	for _, item := range logs {
		if item.EventName == "data_request_reused" {
			foundReuse = true
			break
		}
	}
	if !foundReuse {
		t.Fatal("expected data_request_reused log")
	}
}

func TestResumePausedExecutionContinuesAfterSubTaskSummarizeGameClientCallback(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"summarize"}]}`,
		`{"reply":"child-branch","action_calls":[],"memory_updates":[]}`,
		`{"request_data":{"label":"summarize-scene","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"summary-after-resume"}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom, Context: &InvokeContext{PipelineMode: PipelineFull}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first.DataRequest == nil {
		t.Fatal("expected paused summarize data request")
	}
	if len(first.ActionCalls) != 1 || first.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected callback action, got %+v", first.ActionCalls)
	}
	callbackID := first.ActionCalls[0].CallbackID

	paused, err := store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get paused execution: %v", err)
	}
	if paused.RequestID == "" || paused.WorldUUID != worldID {
		t.Fatalf("expected parent paused execution snapshot, got %+v", paused)
	}

	resumed, err := pipeline.ResumePausedExecution(callbackID, map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed == nil {
		t.Fatal("expected resumed response")
	}
	if !strings.Contains(resumed.Reply, "summary-after-resume") {
		t.Fatalf("expected resumed summary reply, got %+v", resumed)
	}
	if !strings.Contains(provider.lastPrompt, `"scene":"tavern"`) {
		t.Fatalf("expected resumed summarize prompt to include callback payload, got %s", provider.lastPrompt)
	}
	if len(resumed.SubTasks) != 1 || resumed.SubTasks[0].Label != "fetch_scene" {
		t.Fatalf("expected resumed sub-task metadata, got %+v", resumed.SubTasks)
	}
	paused, err = store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get paused execution after resume: %v", err)
	}
	if paused.Status != "completed" {
		t.Fatalf("expected completed paused execution, got %q", paused.Status)
	}
	if provider.calls != 4 {
		t.Fatalf("expected 4 llm calls, got %d", provider.calls)
	}
}

func TestResumePausedExecutionPreservesStoreContextBeforeSubTaskSummarizeGameClientCallback(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	if err := store.CreateComponent(&store.ComponentModel{NodeUUID: nodeID, NodeID: store.ResolveNodeUUID(nodeID), ComponentType: string(CompLore), Data: `{"scene":"lantern hall"}`}); err != nil {
		t.Fatalf("create component: %v", err)
	}
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"summarize"}]}`,
		`{"reply":"child-branch","action_calls":[],"memory_updates":[]}`,
		`{"request_data":{"label":"fetch-store","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"request_data":{"label":"summarize-scene","target":"game_client","queries":[{"type":"node_relations","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"summary-after-mixed-resume"}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom, Context: &InvokeContext{PipelineMode: PipelineFull}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first.DataRequest == nil {
		t.Fatal("expected paused summarize data request")
	}
	callbackID := first.ActionCalls[0].CallbackID
	resumed, err := pipeline.ResumePausedExecution(callbackID, map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed == nil || !strings.Contains(resumed.Reply, "summary-after-mixed-resume") {
		t.Fatalf("expected resumed mixed summary reply, got %+v", resumed)
	}
	if !strings.Contains(provider.lastPrompt, `"scene":"tavern"`) {
		t.Fatalf("expected resumed summarize prompt to include callback payload, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, "[summarize request_data] fetch-store") {
		t.Fatalf("expected resumed summarize prompt to retain store query markers, got %s", provider.lastPrompt)
	}
	if provider.calls != 5 {
		t.Fatalf("expected 5 llm calls, got %d", provider.calls)
	}
}

func TestResumePausedExecutionSupportsRepeatedSubTaskSummarizeGameClientCallbacks(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"summarize"}]}`,
		`{"reply":"child-branch","action_calls":[],"memory_updates":[]}`,
		`{"request_data":{"label":"summarize-pass-1","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"request_data":{"label":"summarize-pass-2","target":"game_client","queries":[{"type":"node_relations","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"summary-after-second-resume"}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom, Context: &InvokeContext{PipelineMode: PipelineFull}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first.DataRequest == nil || len(first.ActionCalls) != 1 || first.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected first paused summarize request, got %+v", first)
	}
	firstCallbackID := first.ActionCalls[0].CallbackID

	second, err := pipeline.ResumePausedExecution(firstCallbackID, map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("first resume: %v", err)
	}
	if second == nil || second.DataRequest == nil || len(second.ActionCalls) != 1 || second.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected second paused summarize request, got %+v", second)
	}
	secondCallbackID := second.ActionCalls[0].CallbackID
	if secondCallbackID == firstCallbackID {
		t.Fatalf("expected new callback id for second summarize pause, got %q", secondCallbackID)
	}

	resumed, err := pipeline.ResumePausedExecution(secondCallbackID, map[string]any{"relations": []string{"guard", "merchant"}})
	if err != nil {
		t.Fatalf("second resume: %v", err)
	}
	if resumed == nil || !strings.Contains(resumed.Reply, "summary-after-second-resume") {
		t.Fatalf("expected final resumed summary reply, got %+v", resumed)
	}
	if !strings.Contains(provider.lastPrompt, `"relations":["guard","merchant"]`) {
		t.Fatalf("expected final summarize prompt to include second callback payload, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, `"scene":"tavern"`) {
		t.Fatalf("expected final summarize prompt to retain first callback payload, got %s", provider.lastPrompt)
	}
	if provider.calls != 5 {
		t.Fatalf("expected 5 llm calls, got %d", provider.calls)
	}
	paused, err := store.GetPausedExecutionByCallbackID(secondCallbackID)
	if err != nil {
		t.Fatalf("get second paused execution: %v", err)
	}
	if paused.Status != "completed" {
		t.Fatalf("expected completed paused execution after second resume, got %q", paused.Status)
	}
}

func TestSubTaskSummarizePersistsStructuredDataRequestLogs(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	previousMode := config.Global.Engine.ExecutionMode
	config.Global.Engine.ExecutionMode = "debug"
	defer func() { config.Global.Engine.ExecutionMode = previousMode }()
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"summarize"}]}`,
		`{"reply":"child-branch","action_calls":[],"memory_updates":[]}`,
		`{"request_data":{"label":"fetch-store","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"request_data":{"label":"summarize-scene","target":"game_client","queries":[{"type":"node_relations","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"summary-after-resume"}`,
	}}
	pipeline := NewPipeline(provider)

	first, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom, Context: &InvokeContext{PipelineMode: PipelineFull}})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if first == nil || first.DataRequest == nil || len(first.ActionCalls) != 1 {
		t.Fatalf("expected summarize pause response, got %+v", first)
	}
	callbackID := first.ActionCalls[0].CallbackID
	resumed, err := pipeline.ResumePausedExecution(callbackID, map[string]any{"relations": []string{"guard", "merchant"}})
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if resumed == nil || !strings.Contains(resumed.Reply, "summary-after-resume") {
		t.Fatalf("expected resumed summarize reply, got %+v", resumed)
	}

	logs, err := store.GetInferenceLogs(worldID, 100, 0, string(TaskCustom))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	counts := map[string]int{}
	for _, item := range logs {
		counts[item.EventName]++
	}
	for _, eventName := range []string{
		"summarize_data_request_emitted",
		"summarize_data_request_resolved",
		"summarize_data_request_paused_for_client",
		"summarize_data_request_resolved_from_client",
	} {
		if counts[eventName] == 0 {
			t.Fatalf("expected summarize log event %q, got %+v", eventName, counts)
		}
	}

	pausedLogs, err := store.GetInferenceLogsByQuery(store.InferenceLogQuery{WorldUUID: worldID, TaskType: string(TaskCustom), EventName: "summarize_data_request_paused_for_client", Limit: 10})
	if err != nil {
		t.Fatalf("get paused summarize logs: %v", err)
	}
	if len(pausedLogs) == 0 {
		t.Fatal("expected paused summarize log rows")
	}
	if pausedLogs[0].ResponseData == "" {
		t.Fatal("expected paused summarize response data")
	}

	resolvedLogs, err := store.GetInferenceLogsByQuery(store.InferenceLogQuery{WorldUUID: worldID, TaskType: string(TaskCustom), EventName: "summarize_data_request_resolved_from_client", Limit: 10})
	if err != nil {
		t.Fatalf("get resolved summarize logs: %v", err)
	}
	if len(resolvedLogs) == 0 {
		t.Fatal("expected summarize resolved-from-client log rows")
	}
	if resolvedLogs[0].DetailData == "" {
		t.Fatal("expected summarize resolved-from-client detail data in debug mode")
	}
	if !strings.Contains(resolvedLogs[0].DetailData, "guard") {
		t.Fatalf("expected callback payload in summarize resolved detail, got %s", resolvedLogs[0].DetailData)
	}
}

func TestSubTaskSummarizeDynamicDataRequestUsesAllowedInterface(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"summarize"}]}`,
		`{"reply":"child-branch","action_calls":[],"memory_updates":[]}`,
		`{"request_data":{"label":"summarize-scene","target":"game_client","external_interface":"scene_facts","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{
			PipelineMode: PipelineFull,
			DynamicInterfaces: []DynamicInterface{{
				ID:                "scene_facts",
				Kind:              DynamicInterfaceDataRequest,
				ExternalInterface: "game_client_request_data",
				Description:       "Query visible scene state",
				QueryTypes:        []string{"node_detail"},
				MaxQueries:        1,
			}},
		},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.DataRequest == nil || resp.DataRequest.ExternalInterface != "game_client_request_data" {
		t.Fatalf("expected normalized summarize external interface, got %+v", resp.DataRequest)
	}
	if len(resp.ActionCalls) != 1 || resp.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected summarize callback action, got %+v", resp.ActionCalls)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if task.InterfaceName != "game_client_request_data" {
		t.Fatalf("expected summarize runtime task interface game_client_request_data, got %+v", task)
	}
}

func TestSubTaskSummarizeDynamicDataRequestBlocksDisallowedQueryType(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"parent","sub_tasks":[{"label":"fetch_scene","task_type":"custom","node_id":"` + nodeID + `","merge_mode":"summarize"}]}`,
		`{"reply":"child-branch","action_calls":[],"memory_updates":[]}`,
		`{"request_data":{"label":"summarize-scene","target":"game_client","external_interface":"scene_facts","queries":[{"type":"node_relations","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"summary-after-block"}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{
			PipelineMode: PipelineFull,
			DynamicInterfaces: []DynamicInterface{{
				ID:                "scene_facts",
				Kind:              DynamicInterfaceDataRequest,
				ExternalInterface: "game_client_request_data",
				Description:       "Query visible scene state",
				QueryTypes:        []string{"node_detail"},
				MaxQueries:        1,
			}},
		},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.DataRequest != nil {
		t.Fatalf("expected summarize dynamic request to be blocked, got %+v", resp.DataRequest)
	}
	if len(resp.ActionCalls) != 0 {
		t.Fatalf("expected no summarize callback action, got %+v", resp.ActionCalls)
	}
	if !strings.Contains(resp.Reply, "summary-after-block") {
		t.Fatalf("expected summarize loop to continue after blocked request, got %+v", resp)
	}
	tasks, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{WorldUUID: worldID, Limit: 10})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected no runtime tasks for blocked summarize request, got %+v", tasks)
	}
	if !strings.Contains(provider.lastPrompt, "[summarize request_data blocked]") {
		t.Fatalf("expected summarize prompt to record blocked dynamic request, got %s", provider.lastPrompt)
	}
	logs, err := store.GetInferenceLogsByQuery(store.InferenceLogQuery{WorldUUID: worldID, TaskType: string(TaskCustom), EventName: "summarize_data_request_blocked", Limit: 10})
	if err != nil {
		t.Fatalf("get blocked summarize logs: %v", err)
	}
	if len(logs) == 0 {
		t.Fatal("expected blocked summarize data request log")
	}
	if !strings.Contains(logs[0].Message, "query type") {
		t.Fatalf("expected blocked summarize log message to mention query type, got %q", logs[0].Message)
	}
}

func TestExecuteAsyncActionEnqueuesRuntimeTask(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	if _, err := store.UpsertWorldPolicy(worldID, nil, []string{"spawn_item"}); err != nil {
		t.Fatalf("policy: %v", err)
	}
	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"spawn_item","args":{"item_name":"potion","consumer":"bridge"}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.ActionCalls) != 1 || resp.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected async action callback, got %+v", resp.ActionCalls)
	}
	callbackID := resp.ActionCalls[0].CallbackID
	runtimeTask, err := store.GetRuntimeTaskByCallbackID(callbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if runtimeTask.Category != "external_action" {
		t.Fatalf("expected external_action task, got %+v", runtimeTask)
	}
	if runtimeTask.InterfaceName != "spawn_item" || runtimeTask.Consumer != "bridge" {
		t.Fatalf("unexpected runtime task routing: %+v", runtimeTask)
	}
	if runtimeTask.Status != store.RuntimeTaskStatusPending {
		t.Fatalf("expected pending runtime task, got %q", runtimeTask.Status)
	}
}

func TestExecuteIncludeRelatedNodesSkipsSocialRelationsByDefault(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	ally := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Ally", NodeType: string(NodeTypeNPC)}
	if err := store.CreateNode(ally); err != nil {
		t.Fatalf("create ally: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: ally.ID, NodeUUID: ally.UUID, ComponentType: string(CompLore), Data: `{"social":"allied contact"}`}); err != nil {
		t.Fatalf("create ally component: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: ally.ID, TargetUUID: ally.UUID, RelationType: string(RelAlly), Weight: 1}); err != nil {
		t.Fatalf("create ally relation: %v", err)
	}

	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskNPCDialogue, Context: &InvokeContext{IncludeRelatedNodes: true}}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if strings.Contains(provider.lastPrompt, "allied contact") {
		t.Fatalf("did not expect social ally component in prompt, got %s", provider.lastPrompt)
	}
}

func TestExecuteIncludeRelatedNodesSkipsExternalParentByDefault(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	extraScope := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "SecretScope", NodeType: string(NodeTypeFaction)}
	if err := store.CreateNode(extraScope); err != nil {
		t.Fatalf("create extra scope: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: extraScope.ID, NodeUUID: extraScope.UUID, ComponentType: string(CompLore), Data: `{"scope":"external parent only"}`}); err != nil {
		t.Fatalf("create extra scope component: %v", err)
	}
	if err := store.CreateMemory(&store.MemoryModel{NodeID: extraScope.ID, NodeUUID: extraScope.UUID, Content: "只允许显式作用域管线使用的额外挂接信息。", Level: string(MemShared)}); err != nil {
		t.Fatalf("create extra scope memory: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: extraScope.ID, TargetUUID: extraScope.UUID, RelationType: string(RelExternalParent), Weight: 1}); err != nil {
		t.Fatalf("create external_parent relation: %v", err)
	}

	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskNPCDialogue, Context: &InvokeContext{IncludeRelatedNodes: true}}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if strings.Contains(provider.lastPrompt, "external parent only") {
		t.Fatalf("did not expect external_parent component in prompt, got %s", provider.lastPrompt)
	}
	if strings.Contains(provider.lastPrompt, "额外挂接信息") {
		t.Fatalf("did not expect external_parent memory in prompt, got %s", provider.lastPrompt)
	}
}

func TestPropagateUpwardIgnoresExternalParentByDefault(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	mainParent := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "MainParent", NodeType: string(NodeTypeFaction), ParentID: &worldInt}
	if err := store.CreateNode(mainParent); err != nil {
		t.Fatalf("create main parent: %v", err)
	}
	store.ResolveNodeParentUUID(mainParent)
	if err := store.DB.Model(&store.NodeModel{}).Where("id = ?", nodeInt).Update("parent_id", mainParent.ID).Error; err != nil {
		t.Fatalf("attach main parent: %v", err)
	}
	extraScope := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "ExtraScope", NodeType: string(NodeTypeFaction), ParentID: &worldInt}
	if err := store.CreateNode(extraScope); err != nil {
		t.Fatalf("create extra scope: %v", err)
	}
	store.ResolveNodeParentUUID(extraScope)
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: extraScope.ID, TargetUUID: extraScope.UUID, RelationType: string(RelExternalParent), Weight: 1}); err != nil {
		t.Fatalf("create external_parent relation: %v", err)
	}

	pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`})
	req := &InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom}
	runtime := &executionConfig{memoryLimit: 50, maxRounds: 1, configuredPipelineMode: PipelineFull, pipelineMode: PipelineFull}
	pipeline.PropagateUpward(req, runtime, ModeProduction, nodeID, "主父链传播测试", MemLongTerm, 1, false)

	mainMems, err := store.GetNodeMemories(mainParent.UUID, 10)
	if err != nil {
		t.Fatalf("get main parent memories: %v", err)
	}
	extraMems, err := store.GetNodeMemories(extraScope.UUID, 10)
	if err != nil {
		t.Fatalf("get extra scope memories: %v", err)
	}
	if len(mainMems) != 1 {
		t.Fatalf("expected 1 propagated memory on main parent, got %d", len(mainMems))
	}
	if mainMems[0].Content != "主父链传播测试" || mainMems[0].Level != string(MemShared) {
		t.Fatalf("unexpected propagated memory on main parent: %#v", mainMems[0])
	}
	if len(extraMems) != 0 {
		t.Fatalf("did not expect propagated memory on external_parent scope, got %#v", extraMems)
	}
}

func TestParseMemoryUpdatesIncludesPropagationRule(t *testing.T) {
	pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`})
	updates := pipeline.parseMemoryUpdates(`[
		{
			"node_id":"npc-1",
			"content":"环境告警",
			"level":"long_term",
			"tags":"alarm",
			"propagation":{"mode":"environment_scope","max_depth":2}
		}
	]`)
	if len(updates) != 1 {
		t.Fatalf("expected 1 memory update, got %d", len(updates))
	}
	if updates[0].Propagation == nil {
		t.Fatal("expected propagation rule to be parsed")
	}
	if updates[0].Propagation.Mode != PropModeEnvironment {
		t.Fatalf("expected environment_scope mode, got %q", updates[0].Propagation.Mode)
	}
	if updates[0].Propagation.MaxDepth != 2 {
		t.Fatalf("expected max_depth 2, got %d", updates[0].Propagation.MaxDepth)
	}
}

func TestParseMemoryUpdatesDropsUnsupportedPropagationRule(t *testing.T) {
	pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`})
	updates := pipeline.parseMemoryUpdates(`[
		{
			"node_id":"npc-1",
			"content":"错误传播模式",
			"level":"long_term",
			"propagation":{"mode":"sideways"}
		}
	]`)
	if len(updates) != 1 {
		t.Fatalf("expected 1 memory update, got %d", len(updates))
	}
	if updates[0].Propagation != nil {
		t.Fatalf("expected invalid propagation rule to be dropped, got %#v", updates[0].Propagation)
	}
}

func TestPropagateEnvironmentScopeUsesLocatedAtAndLocationAncestors(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	region := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "NorthRegion", NodeType: string(NodeTypeLocation), ParentID: &worldInt}
	if err := store.CreateNode(region); err != nil {
		t.Fatalf("create region: %v", err)
	}
	store.ResolveNodeParentUUID(region)
	location := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Camp", NodeType: string(NodeTypeLocation), ParentID: &region.ID}
	if err := store.CreateNode(location); err != nil {
		t.Fatalf("create location: %v", err)
	}
	store.ResolveNodeParentUUID(location)
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: location.ID, TargetUUID: location.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create located_at relation: %v", err)
	}

	pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`})
	req := &InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom}
	runtime := &executionConfig{memoryLimit: 50, maxRounds: 1, configuredPipelineMode: PipelineFull, pipelineMode: PipelineFull}
	pipeline.PropagateMemoryByRule(req, runtime, ModeProduction, MemoryUpdate{NodeID: nodeID, Content: "环境异动", Level: MemLongTerm, Propagation: &PropagationRule{Mode: PropModeEnvironment, MaxDepth: 2}}, nodeID)

	locationMems, err := store.GetNodeMemories(location.UUID, 10)
	if err != nil {
		t.Fatalf("get location memories: %v", err)
	}
	regionMems, err := store.GetNodeMemories(region.UUID, 10)
	if err != nil {
		t.Fatalf("get region memories: %v", err)
	}
	worldMems, err := store.GetNodeMemories(worldID, 10)
	if err != nil {
		t.Fatalf("get world memories: %v", err)
	}
	if len(locationMems) != 1 || locationMems[0].Level != string(MemShared) {
		t.Fatalf("expected shared memory on location, got %#v", locationMems)
	}
	if len(regionMems) != 1 || regionMems[0].Level != string(MemWorld) {
		t.Fatalf("expected world memory on region ancestor, got %#v", regionMems)
	}
	if len(worldMems) != 1 || worldMems[0].Level != string(MemWorld) {
		t.Fatalf("expected world memory on world ancestor, got %#v", worldMems)
	}
}

func TestPropagateOrganizationScopeUsesBelongsToAndSubordinateTargets(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	faction := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "IronLegion", NodeType: string(NodeTypeFaction), ParentID: &worldInt}
	if err := store.CreateNode(faction); err != nil {
		t.Fatalf("create faction: %v", err)
	}
	store.ResolveNodeParentUUID(faction)
	commander := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Commander", NodeType: string(NodeTypeNPC), ParentID: &faction.ID}
	if err := store.CreateNode(commander); err != nil {
		t.Fatalf("create commander: %v", err)
	}
	store.ResolveNodeParentUUID(commander)
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: faction.ID, TargetUUID: faction.UUID, RelationType: string(RelBelongsTo), Weight: 1}); err != nil {
		t.Fatalf("create belongs_to relation: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: commander.ID, TargetUUID: commander.UUID, RelationType: string(RelSubordinate), Weight: 1}); err != nil {
		t.Fatalf("create subordinate relation: %v", err)
	}

	pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`})
	req := &InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom}
	runtime := &executionConfig{memoryLimit: 50, maxRounds: 1, configuredPipelineMode: PipelineFull, pipelineMode: PipelineFull}
	pipeline.PropagateMemoryByRule(req, runtime, ModeProduction, MemoryUpdate{NodeID: nodeID, Content: "组织通报", Level: MemLongTerm, Propagation: &PropagationRule{Mode: PropModeOrganization, MaxDepth: 1}}, nodeID)

	factionMems, err := store.GetNodeMemories(faction.UUID, 10)
	if err != nil {
		t.Fatalf("get faction memories: %v", err)
	}
	commanderMems, err := store.GetNodeMemories(commander.UUID, 10)
	if err != nil {
		t.Fatalf("get commander memories: %v", err)
	}
	worldMems, err := store.GetNodeMemories(worldID, 10)
	if err != nil {
		t.Fatalf("get world memories: %v", err)
	}
	if len(factionMems) != 1 || factionMems[0].Level != string(MemShared) {
		t.Fatalf("expected shared memory on belongs_to target, got %#v", factionMems)
	}
	if len(commanderMems) != 1 || commanderMems[0].Level != string(MemShared) {
		t.Fatalf("expected shared memory on subordinate target, got %#v", commanderMems)
	}
	if len(worldMems) != 1 || worldMems[0].Level != string(MemWorld) {
		t.Fatalf("expected one deduped world-level organization memory, got %#v", worldMems)
	}
}

func TestExecuteAutonomousActUsesLocatedAtEnvironmentChain(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	location := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Gate", NodeType: string(NodeTypeLocation), ParentID: &worldInt}
	if err := store.CreateNode(location); err != nil {
		t.Fatalf("create location: %v", err)
	}
	store.ResolveNodeParentUUID(location)
	if err := store.CreateComponent(&store.ComponentModel{NodeID: location.ID, NodeUUID: location.UUID, ComponentType: string(CompLore), Data: `{"terrain":"stone rampart"}`}); err != nil {
		t.Fatalf("create location component: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: location.ID, TargetUUID: location.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create located_at relation: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: nodeInt, NodeUUID: nodeID, ComponentType: string(CompAutonomous), Data: `{"enabled":true,"trigger":"manual","capabilities":[{"id":"send_dialogue"}]}`}); err != nil {
		t.Fatalf("create autonomous component: %v", err)
	}

	provider := &captureProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskAutonomousAct}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(provider.lastPrompt, "当前环境链：Gate(location) > World(world)") {
		t.Fatalf("expected environment chain in autonomous prompt, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, "stone rampart") {
		t.Fatalf("expected location component in autonomous prompt, got %s", provider.lastPrompt)
	}
}

func TestExecuteWorldEventUsesLocatedAtEnvironmentForNPCScope(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	nodeInt := store.ResolveNodeUUID(nodeID)
	location := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Harbor", NodeType: string(NodeTypeLocation), ParentID: &worldInt}
	if err := store.CreateNode(location); err != nil {
		t.Fatalf("create location: %v", err)
	}
	store.ResolveNodeParentUUID(location)
	if err := store.CreateComponent(&store.ComponentModel{NodeID: location.ID, NodeUUID: location.UUID, ComponentType: string(CompLore), Data: `{"weather":"salt fog"}`}); err != nil {
		t.Fatalf("create location component: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: nodeInt, SourceUUID: nodeID, TargetID: location.ID, TargetUUID: location.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create located_at relation: %v", err)
	}

	provider := &captureProvider{response: `{"reply":"impact","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"impact","world_events":[],"proposed_actions":[]}}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskWorldEvent, Event: &WorldEvent{EventType: "storm", ScopeID: nodeID, Description: "storm front", Severity: "high"}}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(provider.lastPrompt, "当前环境链：Harbor(location) > World(world)") {
		t.Fatalf("expected environment chain in world event prompt, got %s", provider.lastPrompt)
	}
	if !strings.Contains(provider.lastPrompt, "salt fog") {
		t.Fatalf("expected location component in world event prompt, got %s", provider.lastPrompt)
	}
}

func TestExecuteFallsBackToDefaultSettingsWhenStoredValuesAreInvalid(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	settings, err := store.GetOrCreateWorldSettings(worldID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	settings.MemoryLimit = 0
	settings.MaxAnalysisRounds = 0
	settings.SubTaskTimeoutSecs = 0
	settings.SubTaskMaxRetries = -1
	settings.PipelineMode = "mystery"
	if err := store.DB.Save(settings).Error; err != nil {
		t.Fatalf("save invalid settings: %v", err)
	}

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if resp.Metadata.ConfiguredPipelineMode != "full" {
		t.Fatalf("expected configured pipeline mode fallback full, got %q", resp.Metadata.ConfiguredPipelineMode)
	}
	if resp.Metadata.EffectivePipelineMode != "full" {
		t.Fatalf("expected effective pipeline mode fallback full, got %q", resp.Metadata.EffectivePipelineMode)
	}
	if resp.Metadata.MaxAnalysisRounds != 5 {
		t.Fatalf("expected max analysis rounds fallback 5, got %d", resp.Metadata.MaxAnalysisRounds)
	}
}

func TestExecuteWorldTickIncludesWorldTimeConstraintsAndParsesAdvancedTicks(t *testing.T) {
	initTestDB(t)
	worldID, _ := createWorldAndNode(t)
	raw, err := EncodeWorldTimeSettings(&WorldTimeSettings{
		TickScaleMode: TickScaleModeFlexible,
		TickMinUnit:   "时辰",
		TickStep:      1,
		TickUnits:     []string{"日", "时辰"},
		TimeScaleCarry: []WorldTimeCarryRule{{
			From: "时辰",
			To:   "日",
			Base: 12,
		}},
	})
	if err != nil {
		t.Fatalf("encode world time settings: %v", err)
	}
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{WorldTimeSettingsJSON: raw}, &store.WorldSettingsUpdateMask{WorldTimeSettings: true}); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	worldInt := store.ResolveNodeUUID(worldID)
	if err := store.CreateComponent(&store.ComponentModel{NodeID: worldInt, NodeUUID: worldID, ComponentType: string(CompWorldTimeState), Data: `{"tick_scale_mode":"flexible","tick_min_unit":"时辰","tick_step":1,"tick_units":["日","时辰"],"current_units":[{"unit":"日","value":"20"},{"unit":"时辰","value":"卯"}],"current_time_label":"太阴历 8年 7月 20日 卯时辰"}`}); err != nil {
		t.Fatalf("create world_time_state: %v", err)
	}
	provider := &captureProvider{response: `{"reply":"tick","advanced_ticks":3,"action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"推进","world_events":[],"proposed_actions":[]}}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: worldID, TaskType: TaskWorldTick})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.AdvancedTicks != 3 {
		t.Fatalf("expected advanced_ticks=3, got %d", resp.AdvancedTicks)
	}
	for _, want := range []string{"tick_scale_mode: flexible", "current_time_label: 太阴历 8年 7月 20日 卯时辰", "you must return advanced_ticks"} {
		if !strings.Contains(provider.lastPrompt, want) {
			t.Fatalf("expected prompt to contain %q, got %s", want, provider.lastPrompt)
		}
	}
}

func TestExecuteWorldTickIncludesHighValueRelationSummary(t *testing.T) {
	initTestDB(t)
	worldID, _ := createWorldAndNode(t)
	worldInt := store.ResolveNodeUUID(worldID)
	if worldInt == 0 {
		t.Fatal("expected resolved world id")
	}
	city := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "IronCity", NodeType: string(NodeTypeLocation), ParentID: &worldInt}
	if err := store.CreateNode(city); err != nil {
		t.Fatalf("create city: %v", err)
	}
	store.ResolveNodeParentUUID(city)
	faction := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Council", NodeType: string(NodeTypeFaction), ParentID: &worldInt}
	if err := store.CreateNode(faction); err != nil {
		t.Fatalf("create faction: %v", err)
	}
	store.ResolveNodeParentUUID(faction)
	npc := &store.NodeModel{UUID: store.NewUUID(), WorldID: worldInt, Name: "Scout", NodeType: string(NodeTypeNPC), ParentID: &worldInt}
	if err := store.CreateNode(npc); err != nil {
		t.Fatalf("create npc: %v", err)
	}
	store.ResolveNodeParentUUID(npc)
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: npc.ID, SourceUUID: npc.UUID, TargetID: city.ID, TargetUUID: city.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create located_at relation: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: worldInt, WorldUUID: worldID, SourceID: npc.ID, SourceUUID: npc.UUID, TargetID: faction.ID, TargetUUID: faction.UUID, RelationType: string(RelBelongsTo), Weight: 1}); err != nil {
		t.Fatalf("create belongs_to relation: %v", err)
	}

	provider := &captureProvider{response: `{"reply":"tick","advanced_ticks":1,"action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"推进","world_events":[],"proposed_actions":[]}}`}
	pipeline := NewPipeline(provider)
	if _, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: worldID, TaskType: TaskWorldTick}); err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, want := range []string{"当前 scope 的高价值关系摘要：", "子节点分布: World 下有 1 个 location（样本: IronCity）", "子节点分布: World 下有 1 个 faction（样本: Council）", "子节点分布: World 下有 1 个 npc（样本: Scout）", "位置锚点: Scout(npc) 位于 IronCity(location)", "归属结构: Scout(npc) 属于 Council(faction)"} {
		if !strings.Contains(provider.lastPrompt, want) {
			t.Fatalf("expected prompt to contain %q, got %s", want, provider.lastPrompt)
		}
	}
	if strings.Contains(provider.lastPrompt, string(RelAlly)) {
		t.Fatalf("did not expect social relation names in world tick summary, got %s", provider.lastPrompt)
	}
}

func TestExecutePushesGameClientRequestDataThroughHTTPAdapter(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	attempts := 0
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			http.Error(w, "temporary error", http.StatusBadGateway)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	previousIntegrations := config.Global.ExternalIntegrations
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_http": {Type: "http_adapter", BaseURL: server.URL, Path: "/dispatch", RetryMaxAttempts: 2, RetryBackoffMs: 1, IdempotencyHeader: "Idempotency-Key"},
	}
	defer func() { config.Global.ExternalIntegrations = previousIntegrations }()

	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","delivery_mode":"push","primary_transport":"game_http","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp == nil || len(resp.ActionCalls) != 1 || resp.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected async callback response, got %+v", resp)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task by callback id: %v", err)
	}
	if task.Status != store.RuntimeTaskStatusDispatched {
		t.Fatalf("expected dispatched task status, got %q", task.Status)
	}
	if attempts != 2 || task.DispatchAttempts != 2 {
		t.Fatalf("expected 2 dispatch attempts, got server=%d task=%d", attempts, task.DispatchAttempts)
	}
	if task.LastDispatchDecision != "dispatched" || task.LastTransitionReason != "push_dispatch_succeeded" {
		t.Fatalf("expected dispatched decision metadata, got %+v", task)
	}
	if task.IdempotencyKey == "" {
		t.Fatalf("expected idempotency key to be recorded, got %+v", task)
	}
	if gotBody["task_id"] == "" {
		t.Fatalf("expected task_id in dispatched payload, got %+v", gotBody)
	}
	if gotBody["interface_name"] != "game_client_request_data" {
		t.Fatalf("unexpected interface_name in payload: %+v", gotBody)
	}
}

func TestExecuteAsyncActionHybridFallsBackToPendingTaskWhenPushFails(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	previousIntegrations := config.Global.ExternalIntegrations
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_http": {Type: "http_adapter", BaseURL: server.URL, Path: "/dispatch"},
	}
	defer func() { config.Global.ExternalIntegrations = previousIntegrations }()

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"spawn_item","args":{"delivery_mode":"hybrid","primary_transport":"game_http","consumer":"bridge","item_name":"apple"}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp == nil || len(resp.ActionCalls) != 1 || resp.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected async action call, got %+v", resp)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task by callback id: %v", err)
	}
	if task.Status != store.RuntimeTaskStatusPending {
		t.Fatalf("expected hybrid dispatch failure to keep pending task, got %q", task.Status)
	}
	if task.LastDispatchDecision != "pending_retry" || task.LastDispatchFailureClass != "upstream_5xx" {
		t.Fatalf("expected pending_retry decision metadata, got %+v", task)
	}
	if !strings.Contains(task.ErrorMessage, "status 502") {
		t.Fatalf("expected dispatch failure to be recorded, got %q", task.ErrorMessage)
	}
}

func TestExecutePushesGameClientRequestDataThroughWebSocketAdapter(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	var gotBody map[string]any
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade websocket: %v", err)
		}
		defer conn.Close()
		if err := conn.ReadJSON(&gotBody); err != nil {
			t.Fatalf("read websocket dispatch payload: %v", err)
		}
		if err := conn.WriteJSON(map[string]any{"status": 200, "accepted": true}); err != nil {
			t.Fatalf("write websocket response: %v", err)
		}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	previousIntegrations := config.Global.ExternalIntegrations
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_ws": {Type: "websocket_adapter", BaseURL: wsURL, Path: "/dispatch"},
	}
	defer func() { config.Global.ExternalIntegrations = previousIntegrations }()

	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","delivery_mode":"push","primary_transport":"game_ws","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp == nil || len(resp.ActionCalls) != 1 || resp.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected async callback response, got %+v", resp)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task by callback id: %v", err)
	}
	if task.Status != store.RuntimeTaskStatusDispatched {
		t.Fatalf("expected dispatched task status, got %q", task.Status)
	}
	if gotBody["primary_transport"] != "game_ws" {
		t.Fatalf("unexpected primary_transport in payload: %+v", gotBody)
	}
}

type pipelineRPCDispatchService struct{}

func (s *pipelineRPCDispatchService) Dispatch(args map[string]any, reply *map[string]any) error {
	*reply = map[string]any{
		"transport": "game_rpc",
		"status":    200,
		"body":      "rpc accepted",
		"metadata": map[string]any{
			"task_id": args["request"].(map[string]any)["task_id"],
		},
	}
	return nil
}

func startPipelineRPCServer(t *testing.T) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen rpc server: %v", err)
	}
	server := rpc.NewServer()
	if err := server.RegisterName("Runtime", &pipelineRPCDispatchService{}); err != nil {
		listener.Close()
		t.Fatalf("register rpc service: %v", err)
	}
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go server.ServeCodec(jsonrpc.NewServerCodec(conn))
		}
	}()
	return "tcp://" + listener.Addr().String(), func() {
		_ = listener.Close()
		<-stopped
	}
}

func TestExecutePushesGameClientRequestDataThroughRPCAdapter(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	baseURL, stop := startPipelineRPCServer(t)
	defer stop()

	previousIntegrations := config.Global.ExternalIntegrations
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_rpc": {Type: "rpc_adapter", BaseURL: baseURL, Path: "Runtime.Dispatch"},
	}
	defer func() { config.Global.ExternalIntegrations = previousIntegrations }()

	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","delivery_mode":"push","primary_transport":"game_rpc","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp == nil || len(resp.ActionCalls) != 1 || resp.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected async callback response, got %+v", resp)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task by callback id: %v", err)
	}
	if task.Status != store.RuntimeTaskStatusDispatched {
		t.Fatalf("expected dispatched task status, got %q", task.Status)
	}
	if task.Transport != "game_rpc" {
		t.Fatalf("expected transport game_rpc, got %q", task.Transport)
	}
}

func TestExecuteUsesExternalInterfaceConfigForGameClientRequest(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	previousIntegrations := config.Global.ExternalIntegrations
	previousInterfaces := config.Global.ExternalInterfaces
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_http": {Type: "http_adapter", BaseURL: server.URL, Path: "/dispatch"},
	}
	config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
		"game_client_request_data": {Category: "external_query", DeliveryMode: "push", PrimaryTransport: "game_http", Consumer: "game_client", ResumePolicy: "resume_paused_execution"},
	}
	defer func() {
		config.Global.ExternalIntegrations = previousIntegrations
		config.Global.ExternalInterfaces = previousInterfaces
	}()

	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if task.Transport != "game_http" || task.Status != store.RuntimeTaskStatusDispatched {
		t.Fatalf("expected configured route to dispatch via game_http, got %+v", task)
	}
	if gotBody["primary_transport"] != "game_http" {
		t.Fatalf("unexpected dispatch payload: %+v", gotBody)
	}
}

func TestExecuteDynamicDataRequestAliasUsesConfiguredInterface(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	}))
	defer server.Close()

	previousIntegrations := config.Global.ExternalIntegrations
	previousInterfaces := config.Global.ExternalInterfaces
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_http": {Type: "http_adapter", BaseURL: server.URL, Path: "/dispatch"},
	}
	config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
		"game_client_request_data": {Category: "external_query", DeliveryMode: "push", PrimaryTransport: "game_http", Consumer: "game_client", ResumePolicy: "resume_paused_execution"},
	}
	defer func() {
		config.Global.ExternalIntegrations = previousIntegrations
		config.Global.ExternalInterfaces = previousInterfaces
	}()

	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","external_interface":"scene_facts","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"node_detail"},
			MaxQueries:        1,
		}}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.DataRequest == nil || resp.DataRequest.ExternalInterface != "game_client_request_data" {
		t.Fatalf("expected normalized external interface, got %+v", resp.DataRequest)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if task.InterfaceName != "game_client_request_data" {
		t.Fatalf("expected task interface game_client_request_data, got %+v", task)
	}
	if gotBody["interface_name"] != "game_client_request_data" {
		t.Fatalf("unexpected dispatch payload: %+v", gotBody)
	}
}

func TestExecuteDynamicDataRequestBlocksDisallowedQueryType(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &sequenceProvider{responses: []string{
		`{"reply":"need-client","request_data":{"label":"fetch-scene","target":"game_client","external_interface":"scene_facts","queries":[{"type":"node_relations","node_id":"` + nodeID + `"}]}}`,
	}}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "scene_facts",
			Kind:              DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data",
			Description:       "Query visible scene state",
			QueryTypes:        []string{"node_detail"},
			MaxQueries:        1,
		}}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.DataRequest != nil {
		t.Fatalf("expected blocked data request to be dropped, got %+v", resp.DataRequest)
	}
	if len(resp.ActionCalls) != 0 {
		t.Fatalf("expected no async callback action, got %+v", resp.ActionCalls)
	}
	tasks, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{WorldUUID: worldID, Limit: 10})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected no runtime tasks, got %+v", tasks)
	}
}

func TestExecuteUsesExternalInterfaceConfigForAsyncAction(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`ok`))
	}))
	defer server.Close()

	previousIntegrations := config.Global.ExternalIntegrations
	previousInterfaces := config.Global.ExternalInterfaces
	autoRequeue := true
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_http": {Type: "http_adapter", BaseURL: server.URL, Path: "/dispatch"},
	}
	config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
		"spawn_item": {Category: "external_action", DeliveryMode: "push", PrimaryTransport: "game_http", Consumer: "bridge", ResumePolicy: "none", CallbackPostProcess: "write_memory", CallbackMemoryLevel: "long_term", CallbackMemoryTemplate: "spawn callback {status}: {result_json}", MaxAttempts: 3, HeartbeatTimeoutAutoRequeue: &autoRequeue, HeartbeatTimeoutRequeueDelayMs: 2500, HeartbeatTimeoutReason: "interface timeout requeue"},
	}
	defer func() {
		config.Global.ExternalIntegrations = previousIntegrations
		config.Global.ExternalInterfaces = previousInterfaces
	}()

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"spawn_item","args":{"item_name":"apple"}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if task.Transport != "game_http" || task.Status != store.RuntimeTaskStatusDispatched {
		t.Fatalf("expected configured async route to dispatch via game_http, got %+v", task)
	}
	if task.MaxAttempts != 3 {
		t.Fatalf("expected task max_attempts from interface config, got %+v", task)
	}
	var taskPayload struct {
		MaxAttempts            int `json:"max_attempts"`
		HeartbeatTimeoutPolicy struct {
			AutoRequeue    bool   `json:"auto_requeue"`
			RequeueDelayMs int    `json:"requeue_delay_ms"`
			Reason         string `json:"reason"`
		} `json:"heartbeat_timeout_policy"`
		CallbackPostProcess struct {
			Mode           string `json:"mode"`
			MemoryLevel    string `json:"memory_level"`
			MemoryTemplate string `json:"memory_template"`
		} `json:"callback_post_process"`
	}
	if err := json.Unmarshal([]byte(task.PayloadJSON), &taskPayload); err != nil {
		t.Fatalf("unmarshal runtime task payload: %v", err)
	}
	if taskPayload.CallbackPostProcess.Mode != "write_memory" {
		t.Fatalf("expected callback post process mode write_memory, got %+v", taskPayload.CallbackPostProcess)
	}
	if taskPayload.CallbackPostProcess.MemoryLevel != "long_term" {
		t.Fatalf("expected callback memory level long_term, got %+v", taskPayload.CallbackPostProcess)
	}
	if taskPayload.CallbackPostProcess.MemoryTemplate != "spawn callback {status}: {result_json}" {
		t.Fatalf("unexpected callback memory template: %+v", taskPayload.CallbackPostProcess)
	}
	if taskPayload.MaxAttempts != 3 {
		t.Fatalf("expected payload max_attempts 3, got %+v", taskPayload)
	}
	if !taskPayload.HeartbeatTimeoutPolicy.AutoRequeue || taskPayload.HeartbeatTimeoutPolicy.RequeueDelayMs != 2500 || taskPayload.HeartbeatTimeoutPolicy.Reason != "interface timeout requeue" {
		t.Fatalf("unexpected heartbeat timeout policy snapshot: %+v", taskPayload.HeartbeatTimeoutPolicy)
	}
	dispatchPayload := gotBody["payload"].(map[string]any)
	if dispatchPayload["action_id"] != "spawn_item" {
		t.Fatalf("unexpected async dispatch payload: %+v", gotBody)
	}
}

func TestExecuteDynamicActionAliasUsesConfiguredInterface(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	previousInterfaces := config.Global.ExternalInterfaces
	config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
		"npc_trade_action": {Category: "external_action", DeliveryMode: "pull", Consumer: "bridge", ResumePolicy: "none"},
	}
	defer func() {
		config.Global.ExternalInterfaces = previousInterfaces
	}()

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"merchant_ops","args":{"intent":"quote"}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "merchant_ops",
			Kind:              DynamicInterfaceAction,
			ExternalInterface: "npc_trade_action",
			Description:       "Perform trade-related external actions",
			MaxCalls:          1,
		}}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.ActionCalls) != 1 || resp.ActionCalls[0].Mode != "async" || resp.ActionCalls[0].CallbackID == "" {
		t.Fatalf("expected one async dynamic action, got %+v", resp.ActionCalls)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if task.Category != "external_action" || task.InterfaceName != "npc_trade_action" {
		t.Fatalf("expected pull runtime task for npc_trade_action, got %+v", task)
	}
	var payload struct {
		ExternalInterface string         `json:"external_interface"`
		ActionID          string         `json:"action_id"`
		Args              map[string]any `json:"args"`
	}
	if err := json.Unmarshal([]byte(task.PayloadJSON), &payload); err != nil {
		t.Fatalf("unmarshal runtime task payload: %v", err)
	}
	if payload.ExternalInterface != "npc_trade_action" {
		t.Fatalf("expected payload external interface npc_trade_action, got %+v", payload)
	}
	if payload.Args["external_interface"] != "npc_trade_action" {
		t.Fatalf("expected args.external_interface to be normalized, got %+v", payload.Args)
	}
}

func TestExecuteDynamicActionBlocksCallsBeyondMaxCalls(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	previousInterfaces := config.Global.ExternalInterfaces
	config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
		"npc_trade_action": {Category: "external_action", DeliveryMode: "pull", Consumer: "bridge", ResumePolicy: "none"},
	}
	defer func() {
		config.Global.ExternalInterfaces = previousInterfaces
	}()

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"merchant_ops","args":{"intent":"quote"}},{"action_id":"merchant_ops","args":{"intent":"buy"}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "merchant_ops",
			Kind:              DynamicInterfaceAction,
			ExternalInterface: "npc_trade_action",
			Description:       "Perform trade-related external actions",
			MaxCalls:          1,
		}}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.ActionCalls) != 1 {
		t.Fatalf("expected only first dynamic action to pass, got %+v", resp.ActionCalls)
	}
	tasks, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{WorldUUID: worldID, Limit: 10})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected one runtime task, got %+v", tasks)
	}
}

func TestExecuteDynamicActionBlocksUnknownUnapprovedAction(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"forbidden_ops","args":{"intent":"quote"}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "merchant_ops",
			Kind:              DynamicInterfaceAction,
			ExternalInterface: "npc_trade_action",
			Description:       "Perform trade-related external actions",
			MaxCalls:          1,
		}}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.ActionCalls) != 0 {
		t.Fatalf("expected blocked dynamic action to be dropped, got %+v", resp.ActionCalls)
	}
	tasks, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{WorldUUID: worldID, Limit: 10})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected no runtime tasks, got %+v", tasks)
	}
}

func TestExecuteDynamicActionBlocksInvalidArgsBySchema(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"merchant_ops","args":{"intent":123}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "merchant_ops",
			Kind:              DynamicInterfaceAction,
			ExternalInterface: "npc_trade_action",
			Description:       "Perform trade-related external actions",
			ArgsSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"intent": map[string]any{"type": "string"},
				},
				"required": []string{"intent"},
			},
		}}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.ActionCalls) != 0 {
		t.Fatalf("expected invalid dynamic action to be dropped, got %+v", resp.ActionCalls)
	}
	tasks, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{WorldUUID: worldID, Limit: 10})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected no runtime tasks, got %+v", tasks)
	}
}

func TestExecuteDynamicActionBlocksUnknownArgsWhenSchemaDisallowsAdditionalProperties(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"merchant_ops","args":{"intent":"quote","price":10}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{
		WorldID:  worldID,
		NodeID:   nodeID,
		TaskType: TaskCustom,
		Context: &InvokeContext{DynamicInterfaces: []DynamicInterface{{
			ID:                "merchant_ops",
			Kind:              DynamicInterfaceAction,
			ExternalInterface: "npc_trade_action",
			Description:       "Perform trade-related external actions",
			ArgsSchema: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"intent": map[string]any{"type": "string"},
				},
				"required": []string{"intent"},
			},
		}}},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.ActionCalls) != 0 {
		t.Fatalf("expected invalid dynamic action to be dropped, got %+v", resp.ActionCalls)
	}
	tasks, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{WorldUUID: worldID, Limit: 10})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected no runtime tasks, got %+v", tasks)
	}
}

func TestExecuteHybridFallbackTransportMovesAsyncTaskToReleased(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	previousIntegrations := config.Global.ExternalIntegrations
	previousInterfaces := config.Global.ExternalInterfaces
	config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
		"game_http": {Type: "http_adapter", BaseURL: server.URL, Path: "/dispatch"},
	}
	config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
		"spawn_item": {Category: "external_action", DeliveryMode: "hybrid", PrimaryTransport: "game_http", FallbackTransport: "task_pull", Consumer: "bridge", ResumePolicy: "none"},
	}
	defer func() {
		config.Global.ExternalIntegrations = previousIntegrations
		config.Global.ExternalInterfaces = previousInterfaces
	}()

	provider := &stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"spawn_item","args":{"item_name":"apple"}}],"memory_updates":[]}`}
	pipeline := NewPipeline(provider)

	resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
	if err != nil {
		t.Fatalf("get runtime task: %v", err)
	}
	if task.Status != store.RuntimeTaskStatusReleased {
		t.Fatalf("expected released fallback task, got %q", task.Status)
	}
	if task.Transport != "task_pull" {
		t.Fatalf("expected fallback transport task_pull, got %q", task.Transport)
	}
	if task.LastDispatchDecision != "fallback_to_pull" || task.LastDispatchFailureClass != "upstream_5xx" {
		t.Fatalf("expected fallback decision metadata, got %+v", task)
	}
	if task.FallbackFromTransport != "game_http" || task.LastTransitionReason != "push_dispatch_failed_then_fallback" {
		t.Fatalf("expected fallback transition metadata, got %+v", task)
	}
}

func TestExecuteAsyncActionDeliveryMatrix(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	if _, err := store.UpsertWorldPolicy(worldID, nil, []string{"spawn_item"}); err != nil {
		t.Fatalf("policy: %v", err)
	}

	prevIntegrations := config.Global.ExternalIntegrations
	prevInterfaces := config.Global.ExternalInterfaces
	defer func() {
		config.Global.ExternalIntegrations = prevIntegrations
		config.Global.ExternalInterfaces = prevInterfaces
	}()

	type matrixCase struct {
		name               string
		interfaceConfig    config.ExternalInterfaceConfig
		serverStatus       int
		expectStatus       string
		expectTransport    string
		expectMaxAttempts  int
		expectErrorContain string
	}

	cases := []matrixCase{
		{
			name:              "push_success",
			interfaceConfig:   config.ExternalInterfaceConfig{Category: "external_action", DeliveryMode: "push", PrimaryTransport: "game_http", Consumer: "bridge", ResumePolicy: "none", MaxAttempts: 2},
			serverStatus:      http.StatusOK,
			expectStatus:      store.RuntimeTaskStatusDispatched,
			expectTransport:   "game_http",
			expectMaxAttempts: 2,
		},
		{
			name:              "pull_pending",
			interfaceConfig:   config.ExternalInterfaceConfig{Category: "external_action", DeliveryMode: "pull", Consumer: "bridge", ResumePolicy: "none", MaxAttempts: 4},
			expectStatus:      store.RuntimeTaskStatusPending,
			expectTransport:   "",
			expectMaxAttempts: 4,
		},
		{
			name:               "hybrid_fallback_released",
			interfaceConfig:    config.ExternalInterfaceConfig{Category: "external_action", DeliveryMode: "hybrid", PrimaryTransport: "game_http", FallbackTransport: "task_pull", Consumer: "bridge", ResumePolicy: "none", MaxAttempts: 3},
			serverStatus:       http.StatusBadGateway,
			expectStatus:       store.RuntimeTaskStatusReleased,
			expectTransport:    "task_pull",
			expectMaxAttempts:  3,
			expectErrorContain: "status 502",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config.Global.ExternalInterfaces = map[string]config.ExternalInterfaceConfig{
				"spawn_item": tc.interfaceConfig,
			}
			config.Global.ExternalIntegrations = nil
			var server *httptest.Server
			if tc.interfaceConfig.PrimaryTransport != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if tc.serverStatus >= 400 {
						http.Error(w, "upstream error", tc.serverStatus)
						return
					}
					_, _ = w.Write([]byte(`{"status":"accepted"}`))
				}))
				defer server.Close()
				config.Global.ExternalIntegrations = map[string]config.ExternalIntegrationConfig{
					"game_http": {Type: "http_adapter", BaseURL: server.URL, Path: "/dispatch"},
				}
			}

			pipeline := NewPipeline(&stubProvider{response: `{"reply":"ok","action_calls":[{"action_id":"spawn_item","args":{"item_name":"apple"}}],"memory_updates":[]}`})
			resp, err := pipeline.Execute(&InvokeRequest{WorldID: worldID, NodeID: nodeID, TaskType: TaskCustom})
			if err != nil {
				t.Fatalf("execute: %v", err)
			}
			task, err := store.GetRuntimeTaskByCallbackID(resp.ActionCalls[0].CallbackID)
			if err != nil {
				t.Fatalf("get runtime task: %v", err)
			}
			if task.Status != tc.expectStatus {
				t.Fatalf("expected status %q, got %+v", tc.expectStatus, task)
			}
			if task.Transport != tc.expectTransport {
				t.Fatalf("expected transport %q, got %+v", tc.expectTransport, task)
			}
			if task.MaxAttempts != tc.expectMaxAttempts {
				t.Fatalf("expected max_attempts %d, got %+v", tc.expectMaxAttempts, task)
			}
			if tc.expectErrorContain != "" && !strings.Contains(task.ErrorMessage, tc.expectErrorContain) {
				t.Fatalf("expected error message to contain %q, got %+v", tc.expectErrorContain, task)
			}
		})
	}
}
