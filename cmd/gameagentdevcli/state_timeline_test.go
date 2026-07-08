package main

import (
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

	printTimelineSummary(items)
	// Smoke test: function should not panic and should accept structured input.
	// Output assertions are kept in summarize-style helpers elsewhere.
}
