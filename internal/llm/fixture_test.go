package llm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

func TestNewFixtureProviderLoadsSequenceAndRepeatsLast(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.json")
	content := `[
		{"content":"{\"reply\":\"first\",\"action_calls\":[],\"memory_updates\":[]}","tokens":3},
		{"content":"{\"reply\":\"second\",\"action_calls\":[],\"memory_updates\":[]}","tokens":5}
	]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	provider, err := NewFixtureProvider("fixture-model", path)
	if err != nil {
		t.Fatalf("new fixture provider: %v", err)
	}
	first, err := provider.Chat(&engine.LLMChatRequest{})
	if err != nil {
		t.Fatalf("first chat: %v", err)
	}
	second, err := provider.Chat(&engine.LLMChatRequest{})
	if err != nil {
		t.Fatalf("second chat: %v", err)
	}
	third, err := provider.Chat(&engine.LLMChatRequest{})
	if err != nil {
		t.Fatalf("third chat: %v", err)
	}
	if first.Content == second.Content {
		t.Fatalf("expected first and second response to differ, got %q", first.Content)
	}
	if second.Content != third.Content {
		t.Fatalf("expected final response to repeat, got second=%q third=%q", second.Content, third.Content)
	}
	if second.Tokens != 5 || third.Tokens != 5 {
		t.Fatalf("expected repeated final tokens=5, got second=%d third=%d", second.Tokens, third.Tokens)
	}
	if provider.ModelName() != "fixture-model" {
		t.Fatalf("unexpected model name: %s", provider.ModelName())
	}
	if !provider.(engine.LLMStructuredToolProvider).SupportsStructuredTools() {
		t.Fatal("fixture provider should support structured tools")
	}
}

func TestParseFixtureResponsesAcceptsSingleObject(t *testing.T) {
	items, err := parseFixtureResponses([]byte(`{"content":"{\"reply\":\"ok\"}"}`))
	if err != nil {
		t.Fatalf("parse single fixture: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one response, got %d", len(items))
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(items[0].Content), &payload); err != nil {
		t.Fatalf("unmarshal content: %v", err)
	}
	if payload["reply"] != "ok" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}
