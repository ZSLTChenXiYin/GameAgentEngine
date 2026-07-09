package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func TestSummarizeStateComponentIncludesStatusAndPreview(t *testing.T) {
	line := summarizeStateComponent(sdk.StateComponentEnvelope{
		ComponentType: "world_state",
		Component:     &sdk.Component{ID: "comp-1"},
		Data:          map[string]any{"summary": "vault breach", "tick": 7},
	})
	for _, want := range []string{"world_state", "present", "vault breach"} {
		if !strings.Contains(line, want) {
			t.Fatalf("summary missing %q in %q", want, line)
		}
	}
}

func TestPrintContinuityBundleSummaryShowsCoreSections(t *testing.T) {
	bundle := &sdk.ContinuityBundle{
		WorldID: "world-12345678",
		LatestTimeline: &sdk.TimelineEnvelope{
			TickNumber:    7,
			TickType:      "daily",
			GameTime:      "Day 7",
			AdvancedTicks: 2,
			Summary:       "vault sealed",
			FutureOutline: "investigate chamber",
			Data: map[string]any{
				"focus": "reactor",
				"world_time_state": map[string]any{"current_time_label": "Day 7 dawn", "last_advanced_ticks": 2, "total_ticks": 14},
				"previous_world_time_state": map[string]any{"current_time_label": "Day 5 dusk", "last_advanced_ticks": 1},
			},
		},
		Timelines: []sdk.TimelineEnvelope{{
			TickNumber:    6,
			TickType:      "daily",
			GameTime:      "Day 6",
			AdvancedTicks: 1,
			Summary:       "outer gate secured",
			Timeline:      sdk.TimelineTick{CreatedAt: "2026-07-08T09:00:00Z"},
			Data:          map[string]any{"world_time_state": map[string]any{"current_time_label": "Day 6 dusk", "last_advanced_ticks": 1}},
		}},
		StateComponents: []sdk.StateComponentEnvelope{{
			ComponentType: "story_state",
			Component:     &sdk.Component{ID: "comp-2"},
			Data:          map[string]any{"open_threads": []string{"reactor hum"}},
		}},
		Logs: []sdk.InferenceLog{{
			ID:           "log-1",
			WorldID:      "world-12345678",
			NodeID:       "node-abcdef12",
			TaskType:     "world_tick",
			LLMModel:     "gpt-test",
			TokensUsed:   321,
			DurationMs:   88,
			CreatedAt:    "2026-07-08T10:00:00Z",
			RequestData:  `{"pipeline_mode":"full"}`,
			ResponseData: `{"execution_mode":"debug","rounds_used":2,"max_analysis_rounds":4}`,
		}},
		Traces: []sdk.DebugTrace{{
			ID:                     "trace-1",
			WorldID:                "world-12345678",
			NodeID:                 "node-abcdef12",
			RequestID:              "req-12345678",
			TaskType:               "world_tick",
			ConfiguredPipelineMode: "full",
			EffectivePipelineMode:  "polling",
			MaxAnalysisRounds:      4,
			RoundsUsed:             2,
			DurationMs:             77,
			Timestamp:              "2026-07-08T10:00:00Z",
		}},
	}

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = writer
	printContinuityBundleSummary(bundle)
	_ = writer.Close()
	os.Stdout = originalStdout
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	output := string(bytes.TrimSpace(data))
	for _, want := range []string{"Latest Timeline", "Recent Timelines", "vault sealed", "advanced_ticks=2", "world_time=Day 7 dawn", "previous_world_time=Day 5 dusk", "State Components", "story_state", "Recent Logs", "world_tick", "Recent Traces", "trace-1"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output %q", want, output)
		}
	}
}
