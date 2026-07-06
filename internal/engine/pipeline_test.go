package engine

import (
	"fmt"
	"sync"
	"testing"

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
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{PipelineMode: "vertical", MaxAnalysisRounds: 3}, false, false); err != nil {
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
