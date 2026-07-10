package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initRuntimeTaskAPITestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func TestMakeListPendingRuntimeTasksHandlerFiltersTasks(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	now := time.Now()
	later := now.Add(5 * time.Minute)
	seed := []store.RuntimeTaskModel{
		{TaskID: "task-1", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{"scene":"inn"}`, AvailableAt: &now},
		{TaskID: "task-2", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{"scene":"gate"}`, AvailableAt: &now},
		{TaskID: "task-3", Category: "external_query", InterfaceName: "future_scene", DeliveryMode: "pull", Consumer: "game_client", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{"scene":"future"}`, AvailableAt: &later},
	}
	for i := range seed {
		if err := store.CreateRuntimeTask(&seed[i]); err != nil {
			t.Fatalf("seed task %d: %v", i, err)
		}
	}

	h := MakeListPendingRuntimeTasksHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtime/tasks/pending?consumer=game_client&limit=5", nil)
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Tasks []store.RuntimeTaskModel `json:"tasks"`
		Count int                      `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Count != 1 || len(body.Tasks) != 1 || body.Tasks[0].TaskID != "task-1" {
		t.Fatalf("unexpected pending tasks body: %+v", body)
	}
}

func TestMakeClaimRuntimeTaskHandlerClaimsAndRejectsDoubleClaim(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	row := &store.RuntimeTaskModel{TaskID: "claim-api", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{"npc":"merchant"}`}
	if err := store.CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}

	h := MakeClaimRuntimeTaskHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/claim", strings.NewReader(`{"task_id":"claim-api","consumer":"bridge","lease_owner":"bridge-worker-1"}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var body struct {
		Task store.RuntimeTaskModel `json:"task"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal claim response: %v", err)
	}
	if body.Task.Status != store.RuntimeTaskStatusClaimed || body.Task.LeaseToken == "" {
		t.Fatalf("unexpected claimed task: %+v", body.Task)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/claim", strings.NewReader(`{"task_id":"claim-api","consumer":"bridge","lease_owner":"bridge-worker-2"}`))
	w2 := httptest.NewRecorder()
	h(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", w2.Code, w2.Body.String())
	}
	if !strings.Contains(w2.Body.String(), "runtime_task_not_claimable") {
		t.Fatalf("expected runtime_task_not_claimable, got %s", w2.Body.String())
	}
}

func TestMakeHeartbeatAndReleaseRuntimeTaskHandlersValidateLease(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	row := &store.RuntimeTaskModel{TaskID: "lease-api", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{"scene":"market"}`}
	if err := store.CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	claimed, err := store.ClaimRuntimeTask("lease-api", "game_client", "client-1")
	if err != nil {
		t.Fatalf("claim task: %v", err)
	}

	heartbeat := MakeHeartbeatRuntimeTaskHandler()
	hbReq := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/heartbeat", strings.NewReader(`{"task_id":"lease-api","lease_token":"`+claimed.LeaseToken+`"}`))
	hbResp := httptest.NewRecorder()
	heartbeat(hbResp, hbReq)
	if hbResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", hbResp.Code, hbResp.Body.String())
	}

	hbConflictReq := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/heartbeat", strings.NewReader(`{"task_id":"lease-api","lease_token":"wrong"}`))
	hbConflictResp := httptest.NewRecorder()
	heartbeat(hbConflictResp, hbConflictReq)
	if hbConflictResp.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", hbConflictResp.Code, hbConflictResp.Body.String())
	}

	release := MakeReleaseRuntimeTaskHandler()
	releaseReq := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/release", strings.NewReader(`{"task_id":"lease-api","lease_token":"`+claimed.LeaseToken+`","retry_delay_ms":2500,"error_message":"retry later"}`))
	releaseResp := httptest.NewRecorder()
	release(releaseResp, releaseReq)
	if releaseResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", releaseResp.Code, releaseResp.Body.String())
	}

	var body struct {
		Task store.RuntimeTaskModel `json:"task"`
	}
	if err := json.Unmarshal(releaseResp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal release response: %v", err)
	}
	if body.Task.Status != store.RuntimeTaskStatusReleased || body.Task.LeaseToken != "" {
		t.Fatalf("unexpected released task: %+v", body.Task)
	}
	if body.Task.AvailableAt == nil {
		t.Fatalf("expected available_at after release, got %+v", body.Task)
	}

	releaseConflictReq := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/release", strings.NewReader(`{"task_id":"lease-api","lease_token":"`+claimed.LeaseToken+`"}`))
	releaseConflictResp := httptest.NewRecorder()
	release(releaseConflictResp, releaseConflictReq)
	if releaseConflictResp.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", releaseConflictResp.Code, releaseConflictResp.Body.String())
	}
	if !strings.Contains(releaseConflictResp.Body.String(), "runtime_task_lease_mismatch") {
		t.Fatalf("expected runtime_task_lease_mismatch, got %s", releaseConflictResp.Body.String())
	}
}
