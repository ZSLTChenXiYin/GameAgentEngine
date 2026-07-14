package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActionCallbackReturnsStructuredResponse(t *testing.T) {
	var gotPath string
	var gotPayload map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":              "ok",
			"resume_execution_id": "exec-1",
			"post_process": map[string]any{
				"status":  "succeeded",
				"applied": true,
				"details": map[string]any{"memory_level": "long_term"},
			},
			"resumed": map[string]any{
				"request_id":     "req-2",
				"task_type":      "npc_dialogue",
				"execution_mode": "production",
				"reply":          "resumed-final",
			},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.ActionCallback("cb-1", "success", map[string]any{"scene": "tavern"})
	if err != nil {
		t.Fatalf("action callback: %v", err)
	}
	if gotPath != "/api/v1/actions/callback" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotPayload["callback_id"] != "cb-1" || gotPayload["status"] != "success" {
		t.Fatalf("unexpected request payload: %#v", gotPayload)
	}
	if resp.Status != "ok" || resp.ResumeExecutionID != "exec-1" {
		t.Fatalf("unexpected callback response: %#v", resp)
	}
	if resp.PostProcess == nil || !resp.PostProcess.Applied || resp.PostProcess.Details["memory_level"] != "long_term" {
		t.Fatalf("unexpected post process: %#v", resp.PostProcess)
	}
	if resp.Resumed == nil || resp.Resumed.Reply != "resumed-final" {
		t.Fatalf("unexpected resumed response: %#v", resp.Resumed)
	}
}

func TestListPendingRuntimeTasksUsesPendingEndpoint(t *testing.T) {
	var gotPath string
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tasks": []map[string]any{{
				"task_id":         "task-1",
				"status":          "pending",
				"interface_name":  "game_client_request_data",
				"delivery_mode":   "pull",
				"consumer":        "game_client",
				"payload_json":    `{"scene":"tavern"}`,
				"attempt_count":   1,
				"max_attempts":    3,
				"available_at":    "2026-01-01T00:00:00Z",
				"created_at":      "2026-01-01T00:00:00Z",
				"updated_at":      "2026-01-01T00:00:01Z",
			}},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	tasks, err := client.ListPendingRuntimeTasks("game_client", 5)
	if err != nil {
		t.Fatalf("list pending runtime tasks: %v", err)
	}
	if gotPath != "/api/v1/runtime/tasks/pending" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery != "consumer=game_client&limit=5" && gotQuery != "limit=5&consumer=game_client" {
		t.Fatalf("unexpected query: %s", gotQuery)
	}
	if len(tasks) != 1 || tasks[0].InterfaceName != "game_client_request_data" || tasks[0].AvailableAt == "" {
		t.Fatalf("unexpected tasks: %#v", tasks)
	}
}

func TestGetRuntimeTaskStatsReturnsStructuredStats(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runtime/tasks/stats" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"stats": map[string]any{
				"total":                        9,
				"ready_pull":                   2,
				"dispatched_without_callback":  1,
				"by_status":                    map[string]any{"pending": 2, "dispatched": 1},
				"by_dispatch_decision":         map[string]any{"fallback_to_pull": 1},
				"by_heartbeat_timeout_count":   map[string]any{"2": 1},
			},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	stats, err := client.GetRuntimeTaskStats()
	if err != nil {
		t.Fatalf("get runtime task stats: %v", err)
	}
	if stats.Total != 9 || stats.ReadyPull != 2 || stats.ByStatus["pending"] != 2 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	if stats.ByDispatchDecision["fallback_to_pull"] != 1 || stats.ByHeartbeatTimeoutCount["2"] != 1 {
		t.Fatalf("unexpected stats detail: %#v", stats)
	}
}

func TestRequeueRuntimeTaskReturnsUpdatedTask(t *testing.T) {
	var gotPayload map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/runtime/tasks/requeue" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"task": map[string]any{
				"task_id":      "task-timeout-1",
				"status":       "released",
				"error_message": "manual requeue",
			},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	task, err := client.RequeueRuntimeTask("task-timeout-1", 1500, "manual requeue")
	if err != nil {
		t.Fatalf("requeue runtime task: %v", err)
	}
	if gotPayload["task_id"] != "task-timeout-1" || gotPayload["retry_delay_ms"] != float64(1500) {
		t.Fatalf("unexpected payload: %#v", gotPayload)
	}
	if task == nil || task.Status != RuntimeTaskStatusReleased || task.ErrorMessage != "manual requeue" {
		t.Fatalf("unexpected task: %#v", task)
	}
}
