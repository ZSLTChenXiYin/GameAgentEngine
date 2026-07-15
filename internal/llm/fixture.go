package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

type fixtureProvider struct {
	model     string
	responses []fixtureResponse
	mu        sync.Mutex
	index     int
}

type fixtureResponse struct {
	Content  string         `json:"content"`
	Tokens   int            `json:"tokens,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

func NewFixtureProvider(model string, fixtureFile string) (engine.LLMProvider, error) {
	trimmedPath := strings.TrimSpace(fixtureFile)
	if trimmedPath == "" {
		return nil, fmt.Errorf("llm.fixture_file is required for fixture provider")
	}
	data, err := os.ReadFile(trimmedPath)
	if err != nil {
		return nil, fmt.Errorf("read llm fixture file: %w", err)
	}
	responses, err := parseFixtureResponses(data)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(model) == "" {
		model = "fixture"
	}
	return &fixtureProvider{model: model, responses: responses}, nil
}

func parseFixtureResponses(data []byte) ([]fixtureResponse, error) {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil, fmt.Errorf("llm fixture file is empty")
	}
	var list []fixtureResponse
	if err := json.Unmarshal([]byte(trimmed), &list); err == nil {
		if len(list) == 0 {
			return nil, fmt.Errorf("llm fixture file does not contain responses")
		}
		for i := range list {
			if strings.TrimSpace(list[i].Content) == "" {
				return nil, fmt.Errorf("llm fixture response[%d] content required", i)
			}
		}
		return list, nil
	}
	var single fixtureResponse
	if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
		return nil, fmt.Errorf("parse llm fixture file: %w", err)
	}
	if strings.TrimSpace(single.Content) == "" {
		return nil, fmt.Errorf("llm fixture response content required")
	}
	return []fixtureResponse{single}, nil
}

func (p *fixtureProvider) Chat(req *engine.LLMChatRequest) (*engine.LLMResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.responses) == 0 {
		return nil, fmt.Errorf("fixture provider has no responses")
	}
	idx := p.index
	if idx >= len(p.responses) {
		idx = len(p.responses) - 1
	}
	resp := p.responses[idx]
	if p.index < len(p.responses)-1 {
		p.index++
	}
	tokens := resp.Tokens
	if tokens <= 0 {
		tokens = 1
	}
	return &engine.LLMResult{
		Content:  resp.Content,
		Model:    p.model,
		Tokens:   tokens,
		Metadata: resp.Metadata,
	}, nil
}

func (p *fixtureProvider) ModelName() string { return p.model }

func (p *fixtureProvider) SupportsStructuredTools() bool { return true }
