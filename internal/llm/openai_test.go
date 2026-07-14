package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

func TestOpenAIProviderIncludesStructuredToolsInRequest(t *testing.T) {
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"reply\":\"ok\",\"action_calls\":[],\"memory_updates\":[]}"}}],"usage":{"total_tokens":12}}`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider("test-key", server.URL, "test-model")
	_, err := provider.Chat(&engine.LLMChatRequest{
		SystemPrompt: "system",
		Messages:     []engine.ChatMessage{{Role: "user", Content: "hello"}},
		Tools: []engine.LLMToolDefinition{{
			Name:        "merchant_ops",
			Description: "Perform trade actions",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"intent": map[string]any{"type": "string"},
				},
			},
		}},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if gotBody["tool_choice"] != "auto" {
		t.Fatalf("expected tool_choice auto, got %+v", gotBody)
	}
	tools, ok := gotBody["tools"].([]any)
	if !ok || len(tools) != 1 {
		t.Fatalf("expected one tool in request, got %+v", gotBody)
	}
	tool := tools[0].(map[string]any)
	fn := tool["function"].(map[string]any)
	if fn["name"] != "merchant_ops" {
		t.Fatalf("expected sanitized tool name merchant_ops, got %+v", fn)
	}
}

func TestOpenAIProviderNormalizesToolCallsIntoEngineJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"tool path","tool_calls":[{"id":"call_1","type":"function","function":{"name":"scene_facts","arguments":"{\"queries\":[{\"type\":\"node_detail\",\"node_id\":\"node-1\"}]}"}},{"id":"call_2","type":"function","function":{"name":"merchant_ops","arguments":"{\"intent\":\"quote\"}"}}]}}],"usage":{"total_tokens":21}}`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider("test-key", server.URL, "test-model")
	resp, err := provider.Chat(&engine.LLMChatRequest{
		SystemPrompt: "system",
		Tools: []engine.LLMToolDefinition{
			{
				Name:       "scene_facts",
				Invocation: engine.LLMToolInvocationDataRequest,
				DataRequest: &engine.LLMDataRequestTemplate{
					Label:             "fetch-scene",
					Target:            "game_client",
					ExternalInterface: "game_client_request_data",
				},
				Parameters: map[string]any{
					"type": "object",
				},
			},
			{
				Name:       "merchant_ops",
				Invocation: engine.LLMToolInvocationAction,
				ActionID:   "merchant_ops",
				Parameters: map[string]any{
					"type": "object",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	var payload struct {
		Reply       string `json:"reply"`
		RequestData struct {
			Label             string `json:"label"`
			Target            string `json:"target"`
			ExternalInterface string `json:"external_interface"`
			Queries           []struct {
				Type   string `json:"type"`
				NodeID string `json:"node_id"`
			} `json:"queries"`
		} `json:"request_data"`
		ActionCalls []struct {
			ActionID string         `json:"action_id"`
			Args     map[string]any `json:"args"`
		} `json:"action_calls"`
	}
	if err := json.Unmarshal([]byte(resp.Content), &payload); err != nil {
		t.Fatalf("unmarshal normalized content: %v", err)
	}
	if payload.Reply != "tool path" {
		t.Fatalf("expected assistant content to survive normalization, got %+v", payload)
	}
	if payload.RequestData.ExternalInterface != "game_client_request_data" || payload.RequestData.Target != "game_client" {
		t.Fatalf("expected normalized request_data, got %+v", payload.RequestData)
	}
	if len(payload.RequestData.Queries) != 1 || payload.RequestData.Queries[0].Type != "node_detail" {
		t.Fatalf("expected request_data queries, got %+v", payload.RequestData)
	}
	if len(payload.ActionCalls) != 1 || payload.ActionCalls[0].ActionID != "merchant_ops" {
		t.Fatalf("expected normalized action call, got %+v", payload.ActionCalls)
	}
	if payload.ActionCalls[0].Args["intent"] != "quote" {
		t.Fatalf("expected action args to survive normalization, got %+v", payload.ActionCalls[0].Args)
	}
	if resp.Metadata == nil {
		t.Fatal("expected provider metadata to be populated")
	}
	if normalized, _ := resp.Metadata["structured_output_normalized"].(bool); !normalized {
		t.Fatalf("expected structured_output_normalized metadata, got %+v", resp.Metadata)
	}
	rawCalls, ok := resp.Metadata["tool_calls"].([]map[string]any)
	if !ok || len(rawCalls) != 2 {
		t.Fatalf("expected normalized tool_calls metadata, got %+v", resp.Metadata)
	}
}
