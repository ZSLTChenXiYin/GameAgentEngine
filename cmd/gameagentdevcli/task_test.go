package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func TestPrintRuntimeTaskSummaryIncludesCallbackAndResumeFields(t *testing.T) {
	task := &sdk.RuntimeTask{
		TaskID:            "task-12345678",
		Status:            sdk.RuntimeTaskStatusDispatched,
		Category:          "external_query",
		InterfaceName:     "game_client_request_data",
		DeliveryMode:      "hybrid",
		Consumer:          "game_client",
		Transport:         "game_http",
		CallbackID:        "cb-12345678",
		ResumeExecutionID: "exec-12345678",
		AttemptCount:      1,
		MaxAttempts:       3,
		DispatchAttempts:  2,
		ResultJSON:        `{"scene":"tavern"}`,
	}

	output := captureStdout(t, func() { printRuntimeTaskSummary(task) })
	for _, want := range []string{"task-12345678", "game_client_request_data", "callback=cb-12345678", "resume_execution=exec-12345678", `result={"scene":"tavern"}`} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output %q", want, output)
		}
	}
}

func TestPrintRuntimeTaskInspectionIncludesDispatchAndPayload(t *testing.T) {
	task := &sdk.RuntimeTask{
		TaskID:                   "task-12345678",
		Status:                   sdk.RuntimeTaskStatusReleased,
		PayloadJSON:              `{"resume_policy":"resume_paused_execution"}`,
		LastDispatchDecision:     "fallback_to_pull",
		LastDispatchFailureClass: "upstream_5xx",
		LastTransitionReason:     "push_dispatch_failed_then_fallback",
		DispatchedAt:             "2026-07-14T10:00:00Z",
		CompletedAt:              "2026-07-14T10:01:00Z",
	}

	output := captureStdout(t, func() { printRuntimeTaskInspection(task) })
	for _, want := range []string{"payload=", "fallback_to_pull", "upstream_5xx", "push_dispatch_failed_then_fallback", "dispatched_at=2026-07-14T10:00:00Z", "completed_at=2026-07-14T10:01:00Z"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output %q", want, output)
		}
	}
}

func TestPrintRuntimeTaskStatsIncludesCoreCounters(t *testing.T) {
	stats := &sdk.RuntimeTaskStats{
		Total:                     9,
		ReadyPull:                 2,
		InFlight:                  3,
		Terminal:                  4,
		HeartbeatTimeout:          1,
		DispatchErrorTasks:        1,
		RetryExhaustedTasks:       2,
		DispatchedWithoutCallback: 1,
		RepeatedHeartbeatTimeouts: 1,
		ByStatus:                  map[string]int64{"pending": 2, "released": 1},
		ByDispatchDecision:        map[string]int64{"fallback_to_pull": 1},
		ByHeartbeatTimeoutCount:   map[string]int64{"2": 1},
	}

	output := captureStdout(t, func() { printRuntimeTaskStats(stats) })
	for _, want := range []string{"total=9", "ready_pull=2", "dispatch_errors=1", "retry_exhausted=2", "fallback_to_pull", "by_status="} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in output %q", want, output)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = writer
	fn()
	_ = writer.Close()
	os.Stdout = originalStdout
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	return string(bytes.TrimSpace(data))
}
