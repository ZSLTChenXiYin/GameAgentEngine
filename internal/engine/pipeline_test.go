package engine

import (
	"fmt"
	"strings"
	"sync"
	"testing"

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
	response   string
	lastPrompt string
	lastMsgs   []ChatMessage
}

type sequenceProvider struct {
	responses []string
	calls     int
}

func (s *stubProvider) Chat(systemPrompt string, messages []ChatMessage) (*LLMResult, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return &LLMResult{Content: s.response, Model: "stub", Tokens: 7}, nil
}

func (s *stubProvider) ModelName() string { return "stub" }

func (b *barrierProvider) Chat(systemPrompt string, messages []ChatMessage) (*LLMResult, error) {
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

func (c *captureProvider) Chat(systemPrompt string, messages []ChatMessage) (*LLMResult, error) {
	c.lastPrompt = systemPrompt
	c.lastMsgs = messages
	return &LLMResult{Content: c.response, Model: "capture", Tokens: 11}, nil
}

func (c *captureProvider) ModelName() string { return "capture" }

func (s *sequenceProvider) Chat(systemPrompt string, messages []ChatMessage) (*LLMResult, error) {
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
