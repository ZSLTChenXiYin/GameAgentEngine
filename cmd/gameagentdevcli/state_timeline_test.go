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
		Summary:       "ridge secured",
		FutureOutline: "prepare the lower vault",
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
	for _, want := range []string{"#9", "ridge secured", "prepare the lower vault"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in %q", want, output)
		}
	}
}
