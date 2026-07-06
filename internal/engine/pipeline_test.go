package engine

import (
	"fmt"
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
