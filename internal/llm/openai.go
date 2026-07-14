// Package llm 提供大模型接入实现。
// 当前包含 OpenAI 兼容接口和本地 Mock 两类 Provider。
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

var openAIToolNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

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
		client: &http.Client{
			Timeout: 60 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

type openAIRequest struct {
	Model      string          `json:"model"`
	Messages   []openAIMessage `json:"messages"`
	Tools      []openAITool    `json:"tools,omitempty"`
	ToolChoice string          `json:"tool_choice,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type openAIToolCall struct {
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function openAIToolCallFunction `json:"function"`
}

type openAIToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func sanitizeOpenAIToolName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "tool"
	}
	cleaned := openAIToolNameSanitizer.ReplaceAllString(trimmed, "_")
	cleaned = strings.Trim(cleaned, "_")
	if cleaned == "" {
		return "tool"
	}
	if len(cleaned) > 64 {
		cleaned = cleaned[:64]
	}
	return cleaned
}

func buildOpenAIRequest(model string, chatReq *engine.LLMChatRequest) (openAIRequest, map[string]engine.LLMToolDefinition) {
	request := openAIRequest{Model: model}
	toolMap := map[string]engine.LLMToolDefinition{}

	if chatReq == nil {
		request.Messages = []openAIMessage{{Role: "system"}}
		return request, toolMap
	}

	request.Messages = []openAIMessage{{Role: "system", Content: chatReq.SystemPrompt}}
	for _, m := range chatReq.Messages {
		request.Messages = append(request.Messages, openAIMessage{Role: m.Role, Content: m.Content})
	}
	if len(chatReq.Tools) == 0 {
		return request, toolMap
	}

	request.ToolChoice = "auto"
	request.Tools = make([]openAITool, 0, len(chatReq.Tools))
	usedNames := map[string]int{}
	for _, tool := range chatReq.Tools {
		name := sanitizeOpenAIToolName(tool.Name)
		if count := usedNames[name]; count > 0 {
			name = fmt.Sprintf("%s_%d", name, count+1)
		}
		usedNames[sanitizeOpenAIToolName(tool.Name)]++
		toolMap[name] = tool
		request.Tools = append(request.Tools, openAITool{
			Type: "function",
			Function: openAIToolFunction{
				Name:        name,
				Description: strings.TrimSpace(tool.Description),
				Parameters:  tool.Parameters,
			},
		})
	}
	return request, toolMap
}

func normalizeOpenAIToolCalls(message openAIMessage, toolMap map[string]engine.LLMToolDefinition) (string, bool, map[string]any, error) {
	if len(message.ToolCalls) == 0 {
		return "", false, nil, nil
	}

	result := map[string]any{}
	if strings.TrimSpace(message.Content) != "" {
		result["reply"] = strings.TrimSpace(message.Content)
	}

	var actionCalls []map[string]any
	var dataRequest map[string]any
	var normalizedToolCalls []map[string]any

	for _, call := range message.ToolCalls {
		if call.Type != "" && call.Type != "function" {
			continue
		}
		tool, ok := toolMap[strings.TrimSpace(call.Function.Name)]
		if !ok {
			return "", false, nil, fmt.Errorf("unknown tool call returned by provider: %s", call.Function.Name)
		}
		args := map[string]any{}
		if strings.TrimSpace(call.Function.Arguments) != "" {
			if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
				return "", false, nil, fmt.Errorf("decode tool arguments for %s: %w", call.Function.Name, err)
			}
		}
		switch tool.Invocation {
		case engine.LLMToolInvocationDataRequest:
			if tool.DataRequest == nil {
				return "", false, nil, fmt.Errorf("tool %s missing data request template", tool.Name)
			}
			if dataRequest != nil {
				return "", false, nil, fmt.Errorf("multiple data request tool calls returned by provider are not supported")
			}
			merged := map[string]any{}
			if strings.TrimSpace(tool.DataRequest.Label) != "" {
				merged["label"] = tool.DataRequest.Label
			}
			if strings.TrimSpace(tool.DataRequest.Target) != "" {
				merged["target"] = tool.DataRequest.Target
			}
			if strings.TrimSpace(tool.DataRequest.ExternalInterface) != "" {
				merged["external_interface"] = tool.DataRequest.ExternalInterface
			}
			if strings.TrimSpace(tool.DataRequest.DeliveryMode) != "" {
				merged["delivery_mode"] = tool.DataRequest.DeliveryMode
			}
			if strings.TrimSpace(tool.DataRequest.PrimaryTransport) != "" {
				merged["primary_transport"] = tool.DataRequest.PrimaryTransport
			}
			if strings.TrimSpace(tool.DataRequest.Consumer) != "" {
				merged["consumer"] = tool.DataRequest.Consumer
			}
			if tool.DataRequest.TimeoutMs > 0 {
				merged["timeout_ms"] = tool.DataRequest.TimeoutMs
			}
			for key, value := range args {
				merged[key] = value
			}
			dataRequest = merged
			normalizedToolCalls = append(normalizedToolCalls, map[string]any{
				"name":       tool.Name,
				"invocation": tool.Invocation,
				"payload":    merged,
			})
		case engine.LLMToolInvocationAction:
			actionID := strings.TrimSpace(tool.ActionID)
			if actionID == "" {
				actionID = strings.TrimSpace(tool.Name)
			}
			actionCalls = append(actionCalls, map[string]any{"action_id": actionID, "args": args})
			normalizedToolCalls = append(normalizedToolCalls, map[string]any{
				"name":       tool.Name,
				"invocation": tool.Invocation,
				"payload": map[string]any{
					"action_id": actionID,
					"args":      args,
				},
			})
		default:
			return "", false, nil, fmt.Errorf("unsupported tool invocation kind for %s", tool.Name)
		}
	}

	if len(actionCalls) > 0 {
		result["action_calls"] = actionCalls
	}
	if dataRequest != nil {
		result["request_data"] = dataRequest
	}
	if _, ok := result["reply"]; !ok {
		result["reply"] = ""
	}
	if _, ok := result["action_calls"]; !ok {
		result["action_calls"] = []any{}
	}
	if _, ok := result["memory_updates"]; !ok {
		result["memory_updates"] = []any{}
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", false, nil, fmt.Errorf("marshal normalized tool call payload: %w", err)
	}
	return string(data), true, map[string]any{
		"structured_output_normalized": true,
		"tool_calls":                   normalizedToolCalls,
	}, nil
}

// Chat 调用远端聊天补全接口并返回统一的 LLM 结果。
func (p *openAIProvider) Chat(chatReq *engine.LLMChatRequest) (*engine.LLMResult, error) {
	request, toolMap := buildOpenAIRequest(p.model, chatReq)

	body, err := json.Marshal(request)
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm api error %d: %s", resp.StatusCode, string(respBody))
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices")
	}

	content := result.Choices[0].Message.Content
	var metadata map[string]any
	if normalized, ok, normalizedMetadata, err := normalizeOpenAIToolCalls(result.Choices[0].Message, toolMap); err != nil {
		return nil, err
	} else if ok {
		content = normalized
		metadata = normalizedMetadata
	}

	return &engine.LLMResult{
		Content:  content,
		Model:    p.model,
		Tokens:   result.Usage.TotalTokens,
		Metadata: metadata,
	}, nil
}

// ModelName 返回当前 Provider 使用的模型名称。
func (p *openAIProvider) ModelName() string { return p.model }

func (p *openAIProvider) SupportsStructuredTools() bool { return true }
