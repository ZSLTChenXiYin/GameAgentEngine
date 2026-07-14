package engine

import (
	"strings"
	"testing"
)

type summarizeCaptureProvider struct {
	responses   []string
	calls       int
	lastPrompts []string
}

func (s *summarizeCaptureProvider) Chat(req *LLMChatRequest) (*LLMResult, error) {
	if req != nil {
		s.lastPrompts = append(s.lastPrompts, req.SystemPrompt)
	}
	content := `{"reply":"done"}`
	if s.calls < len(s.responses) {
		content = s.responses[s.calls]
	}
	s.calls++
	return &LLMResult{Content: content, Model: "summarize-capture", Tokens: 3}, nil
}

func (s *summarizeCaptureProvider) ModelName() string { return "summarize-capture" }

func TestDAGSummarizeResultsCanRequestStoreDataAndContinue(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)

	provider := &summarizeCaptureProvider{responses: []string{
		`{"request_data":{"label":"fetch-node","queries":[{"type":"node_detail","node_id":"` + nodeID + `"}]}}`,
		`{"reply":"final merged reply"}`,
	}}

	dag := NewDAGInstance(NewTaskTree(TaskCustom, worldID, nodeID), provider, 2, 0)
	dag.results["a"] = &InvokeResponse{Reply: "first branch reply"}
	dag.results["b"] = &InvokeResponse{Reply: "second branch reply"}

	got := dag.summarizeResults()
	if got != "final merged reply" {
		t.Fatalf("expected final merged reply, got %q", got)
	}
	if provider.calls != 2 {
		t.Fatalf("expected 2 summarize rounds, got %d", provider.calls)
	}
	if len(provider.lastPrompts) < 2 {
		t.Fatalf("expected captured prompts, got %d", len(provider.lastPrompts))
	}
	if !strings.Contains(provider.lastPrompts[1], "[summarize request_data] fetch-node") {
		t.Fatalf("expected second prompt to include request_data context, got %s", provider.lastPrompts[1])
	}
	if !strings.Contains(provider.lastPrompts[1], nodeID) {
		t.Fatalf("expected second prompt to include fetched node detail, got %s", provider.lastPrompts[1])
	}
}

func TestDAGSummarizeResultsBlocksNonStoreRequestData(t *testing.T) {
	provider := &summarizeCaptureProvider{responses: []string{
		`{"request_data":{"label":"fetch-client","target":"game_client","queries":[{"type":"node_detail","node_id":"npc-1"}]}}`,
		`{"reply":"fallback final reply"}`,
	}}

	dag := NewDAGInstance(NewTaskTree(TaskCustom, "world-1", "node-1"), provider, 2, 0)
	dag.results["a"] = &InvokeResponse{Reply: "branch reply"}

	got := dag.summarizeResults()
	if got != "fallback final reply" {
		t.Fatalf("expected fallback final reply, got %q", got)
	}
	if provider.calls != 2 {
		t.Fatalf("expected 2 summarize rounds, got %d", provider.calls)
	}
	if len(provider.lastPrompts) < 2 {
		t.Fatalf("expected captured prompts, got %d", len(provider.lastPrompts))
	}
	if !strings.Contains(provider.lastPrompts[1], "only store target is supported during DAG summarization") {
		t.Fatalf("expected blocked target note in second prompt, got %s", provider.lastPrompts[1])
	}
}

func TestDAGSummarizeResultsRetriesAfterInvalidRequestData(t *testing.T) {
	provider := &summarizeCaptureProvider{responses: []string{
		`{"request_data":"bad"}`,
		`{"reply":"recovered final reply"}`,
	}}

	dag := NewDAGInstance(NewTaskTree(TaskCustom, "world-1", "node-1"), provider, 2, 0)
	dag.results["a"] = &InvokeResponse{Reply: "branch reply"}

	got := dag.summarizeResults()
	if got != "recovered final reply" {
		t.Fatalf("expected recovered final reply, got %q", got)
	}
	if provider.calls != 2 {
		t.Fatalf("expected 2 summarize rounds, got %d", provider.calls)
	}
	if len(provider.lastPrompts) < 2 {
		t.Fatalf("expected captured prompts, got %d", len(provider.lastPrompts))
	}
	if !strings.Contains(provider.lastPrompts[1], "[summarize request_data invalid]") {
		t.Fatalf("expected invalid request_data note in second prompt, got %s", provider.lastPrompts[1])
	}
}
