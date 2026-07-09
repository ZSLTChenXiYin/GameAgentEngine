package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func TestAppendInferenceLogDetailsIncludesPayloads(t *testing.T) {
	lines := appendInferenceLogDetails(nil, sdk.InferenceLog{
		RequestData:  `{"task_type":"world_tick"}`,
		ResponseData: `{"execution_mode":"review"}`,
		DetailData:   `{"raw_response":"..."}`,
	})
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"request_data=", "response_data=", "detail_data="} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %q", want, joined)
		}
	}
}

func TestPrintTimelineSummaryShowsFutureOutline(t *testing.T) {
	items := []sdk.TimelineEnvelope{{
		TickNumber:    9,
		TickType:      "daily",
		GameTime:      "Day 9",
		AdvancedTicks: 3,
		Summary:       "ridge secured",
		FutureOutline: "prepare the lower vault",
		Data: map[string]any{
			"world_time_state": map[string]any{"current_time_label": "Day 9 dusk", "last_advanced_ticks": 3},
		},
		Timeline:      sdk.TimelineTick{CreatedAt: "2026-01-01T00:00:00Z"},
	}}
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = writer
	printTimelineSummary(items)
	_ = writer.Close()
	os.Stdout = originalStdout
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	output := string(bytes.TrimSpace(data))
	for _, want := range []string{"#9", "ridge secured", "prepare the lower vault", "advanced_ticks=3", "world_time=Day 9 dusk"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in %q", want, output)
		}
	}
}
