package main

import (
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func TestSummarizeInferenceLogIncludesPipelineAndPreview(t *testing.T) {
	log := sdk.InferenceLog{
		WorldID:      "world-12345678",
		NodeID:       "node-abcdef12",
		TaskType:     "npc_dialogue",
		LLMModel:     "gpt-test",
		TokensUsed:   321,
		DurationMs:   987,
		CreatedAt:    "2026-07-06T10:00:00Z",
		RequestData:  `{"message_count":2,"pipeline_mode":"polling","max_analysis_rounds":4}`,
		ResponseData: `{"execution_mode":"debug","configured_pipeline_mode":"full","effective_pipeline_mode":"polling","max_analysis_rounds":4,"rounds_used":2,"action_count":1,"memory_update_count":1,"reply_preview":"hello","action_preview":["send_dialogue[sync]"],"memory_preview":["node-abcdef12:long_term"]}`,
	}

	lines := summarizeInferenceLog(log)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{
		"npc_dialogue",
		"321 tokens",
		"request_pipeline=polling",
		"pipeline=full -> polling",
		"rounds=2/4",
		"reply=hello",
		"action_preview=send_dialogue[sync]",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("summary missing %q in %q", want, joined)
		}
	}
}

func TestParseInferenceLogDataReturnsNilForInvalidJSON(t *testing.T) {
	if out := parseInferenceLogRequestData("{bad"); out != nil {
		t.Fatalf("expected nil request parse result, got %#v", out)
	}
	if out := parseInferenceLogResponseData("{bad"); out != nil {
		t.Fatalf("expected nil response parse result, got %#v", out)
	}
}
