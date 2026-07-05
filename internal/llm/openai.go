// Package llm 提供大模型接入实现。
// 当前包含 OpenAI 兼容接口和本地 Mock 两类 Provider。
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

// openAIProvider 是 OpenAI 兼容协议的 Provider 实现。
type openAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewOpenAIProvider 创建一个兼容 OpenAI Chat Completions 的 Provider。
func NewOpenAIProvider(apiKey, baseURL, model string) engine.LLMProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &openAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// openAIRequest 是发往聊天补全接口的请求结构。
type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

// openAIMessage 表示聊天消息项。
type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIResponse 是聊天补全接口的响应结构。
type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// Chat 调用远端聊天补全接口并返回统一的 LLM 结果。
func (p *openAIProvider) Chat(systemPrompt string, messages []engine.ChatMessage) (*engine.LLMResult, error) {
	req := openAIRequest{
		Model: p.model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
		},
	}
	for _, m := range messages {
		req.Messages = append(req.Messages, openAIMessage{Role: m.Role, Content: m.Content})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequest("POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llm api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices")
	}

	return &engine.LLMResult{
		Content: result.Choices[0].Message.Content,
		Model:   p.model,
		Tokens:  result.Usage.TotalTokens,
	}, nil
}

// ModelName 返回当前 Provider 使用的模型名称。
func (p *openAIProvider) ModelName() string { return p.model }
