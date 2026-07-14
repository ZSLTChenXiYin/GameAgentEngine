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
	initTestDB(t)
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
	if !strings.Contains(provider.lastPrompts[0], "standalone DAG summarize helper only supports store-backed request_data continuation") {
		t.Fatalf("expected standalone helper scope note in prompt, got %s", provider.lastPrompts[0])
	}
	if !strings.Contains(provider.lastPrompts[1], "use pipeline DAG summarize for game_client callbacks") {
		t.Fatalf("expected blocked target note in second prompt, got %s", provider.lastPrompts[1])
	}
}

func TestDAGSummarizeResultsRetriesAfterInvalidRequestData(t *testing.T) {
	initTestDB(t)
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

func TestDAGReadyTasksFollowRegistrationOrder(t *testing.T) {
	initTestDB(t)
	dag := NewDAGInstance(NewTaskTree(TaskCustom, "world-1", "node-1"), nil, 2, 0)
	for _, decl := range []SubTaskDeclaration{
		{Label: "first", TaskType: TaskCustom},
		{Label: "second", TaskType: TaskCustom, DependsOn: []string{"first"}},
		{Label: "third", TaskType: TaskCustom},
	} {
		if err := dag.Register(decl); err != nil {
			t.Fatalf("register %s: %v", decl.Label, err)
		}
	}

	ready := dag.ReadyTasks()
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready tasks, got %d", len(ready))
	}
	if ready[0].Label != "first" || ready[1].Label != "third" {
		t.Fatalf("expected ready order [first third], got [%s %s]", ready[0].Label, ready[1].Label)
	}

	dag.OnTaskComplete("first", &InvokeResponse{Reply: "done-first"})
	ready = dag.ReadyTasks()
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready tasks after unlocking dependency, got %d", len(ready))
	}
	if ready[0].Label != "second" || ready[1].Label != "third" {
		t.Fatalf("expected ready order [second third] after unlock, got [%s %s]", ready[0].Label, ready[1].Label)
	}
}

func TestDAGMergeResultsFollowsRegistrationOrder(t *testing.T) {
	initTestDB(t)
	dag := NewDAGInstance(NewTaskTree(TaskCustom, "world-1", "node-1"), nil, 2, 0)
	for _, decl := range []SubTaskDeclaration{
		{Label: "first", TaskType: TaskCustom, MergeMode: "append"},
		{Label: "second", TaskType: TaskCustom, MergeMode: "append"},
		{Label: "third", TaskType: TaskCustom, MergeMode: "append"},
	} {
		if err := dag.Register(decl); err != nil {
			t.Fatalf("register %s: %v", decl.Label, err)
		}
	}

	dag.results["third"] = &InvokeResponse{Reply: "reply-third"}
	dag.results["first"] = &InvokeResponse{Reply: "reply-first"}
	dag.results["second"] = &InvokeResponse{Reply: "reply-second"}

	merged := dag.MergeResults()
	if merged == nil {
		t.Fatal("expected merged response")
	}
	if merged.Reply != "reply-first\nreply-second\nreply-third" {
		t.Fatalf("expected registration-order merge, got %q", merged.Reply)
	}
}
