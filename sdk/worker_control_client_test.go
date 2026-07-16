package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWorkerControlClientListPendingRuntimeTasksUsesRuntimeTaskToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Runtime-Task-Token") != "rt-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"tasks": []map[string]any{{"task_id": "task-1", "interface_name": AuthorityInterfaceGameClientRequestData}}})
	}))
	defer server.Close()

	client := NewWorkerControlClient(server.URL, "rt-token", "cb-token")
	tasks, err := client.ListPendingRuntimeTasks("game_client", 1)
	if err != nil {
		t.Fatalf("ListPendingRuntimeTasks returned error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].InterfaceName != AuthorityInterfaceGameClientRequestData {
		t.Fatalf("unexpected tasks: %#v", tasks)
	}
}

func TestWorkerControlClientActionCallbackUsesCallbackHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Callback-Token") != "cb-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("X-Callback-Request-Id") != "cb-task-1" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "resume_execution_id": "exec-1"})
	}))
	defer server.Close()

	client := NewWorkerControlClient(server.URL, "rt-token", "cb-token")
	resp, err := client.ActionCallback("callback-1", "success", map[string]any{"scene": "inn"}, "cb-task-1")
	if err != nil {
		t.Fatalf("ActionCallback returned error: %v", err)
	}
	if resp == nil || resp.ResumeExecutionID != "exec-1" {
		t.Fatalf("unexpected callback response: %#v", resp)
	}
}
