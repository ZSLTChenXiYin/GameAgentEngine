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

func TestMakeListRuntimeTasksHandlerSupportsStructuredQuery(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	now := time.Now()
	seed := []store.RuntimeTaskModel{
		{TaskID: "all-1", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Transport: "task_pull", WorldUUID: "world-a", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{}`, AvailableAt: &now},
		{TaskID: "all-2", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Transport: "task_pull", WorldUUID: "world-b", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{}`, AvailableAt: &now},
		{TaskID: "all-3", Category: "external_action", InterfaceName: "spawn_item", DeliveryMode: "push", Consumer: "bridge", Transport: "game_http", WorldUUID: "world-a", Status: store.RuntimeTaskStatusDispatched, PayloadJSON: `{}`, AvailableAt: &now},
	}
	for i := range seed {
		if err := store.CreateRuntimeTask(&seed[i]); err != nil {
			t.Fatalf("seed task %d: %v", i, err)
		}
	}
	h := MakeListRuntimeTasksHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtime/tasks?category=external_query&interface_name=fetch_scene&transport=task_pull&world_id=world-a&status=pending&available_only=true&limit=5", nil)
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
	if body.Count != 1 || len(body.Tasks) != 1 || body.Tasks[0].TaskID != "all-1" {
		t.Fatalf("unexpected runtime task list body: %+v", body)
	}
}

func TestMakeGetRuntimeTaskHandlerReturnsTaskByID(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	row := &store.RuntimeTaskModel{TaskID: "detail-api", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{}`}
	if err := store.CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	h := MakeGetRuntimeTaskHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtime/tasks/detail-api", nil)
	req.SetPathValue("task_id", "detail-api")
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Task store.RuntimeTaskModel `json:"task"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Task.TaskID != "detail-api" {
		t.Fatalf("unexpected task detail: %+v", body.Task)
	}
}

func TestMakeRuntimeTaskStatsHandlerReturnsAggregates(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	now := time.Now()
	seed := []store.RuntimeTaskModel{
		{TaskID: "stats-api-1", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Transport: "task_pull", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{}`, AvailableAt: &now},
		{TaskID: "stats-api-2", Category: "external_action", InterfaceName: "spawn_item", DeliveryMode: "push", Consumer: "bridge", Transport: "game_http", Status: store.RuntimeTaskStatusDispatched, PayloadJSON: `{}`},
	}
	for i := range seed {
		if err := store.CreateRuntimeTask(&seed[i]); err != nil {
			t.Fatalf("seed task %d: %v", i, err)
		}
	}
	h := MakeRuntimeTaskStatsHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/runtime/tasks/stats", nil)
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Stats store.RuntimeTaskStats `json:"stats"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Stats.Total != 2 || body.Stats.ReadyPull != 1 || body.Stats.InFlight != 1 {
		t.Fatalf("unexpected stats body: %+v", body.Stats)
	}
}

func TestMakeSweepRuntimeTaskHeartbeatTimeoutHandlerMarksStaleTasks(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	old := time.Now().Add(-10 * time.Minute)
	row := &store.RuntimeTaskModel{TaskID: "sweep-api", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusClaimed, LeaseOwner: "bridge-1", LeaseToken: "tok-1", PayloadJSON: `{}`, LastHeartbeatAt: &old}
	if err := store.CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	h := MakeSweepRuntimeTaskHeartbeatTimeoutHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/heartbeat-timeout/sweep", strings.NewReader(`{"timeout_seconds":60}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Affected int64 `json:"affected"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Affected != 1 {
		t.Fatalf("expected affected 1, got %+v", body)
	}
	item, err := store.GetRuntimeTask("sweep-api")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if item.Status != store.RuntimeTaskStatusHeartbeatTimeout {
		t.Fatalf("expected heartbeat timeout status, got %+v", item)
	}
}

func TestMakeBatchRequeueHeartbeatTimeoutRuntimeTasksHandlerRequeuesFilteredTasks(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	now := time.Now()
	seed := []store.RuntimeTaskModel{
		{TaskID: "batch-api-1", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Transport: "task_pull", Status: store.RuntimeTaskStatusHeartbeatTimeout, PayloadJSON: `{}`, HeartbeatTimeoutAt: &now},
		{TaskID: "batch-api-2", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Transport: "task_pull", Status: store.RuntimeTaskStatusHeartbeatTimeout, PayloadJSON: `{}`, HeartbeatTimeoutAt: &now},
		{TaskID: "batch-api-3", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Transport: "task_pull", Status: store.RuntimeTaskStatusHeartbeatTimeout, PayloadJSON: `{}`, HeartbeatTimeoutAt: &now},
	}
	for i := range seed {
		if err := store.CreateRuntimeTask(&seed[i]); err != nil {
			t.Fatalf("seed task %d: %v", i, err)
		}
	}
	h := MakeBatchRequeueHeartbeatTimeoutRuntimeTasksHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/heartbeat-timeout/requeue", strings.NewReader(`{"consumer":"bridge","category":"external_action","transport":"task_pull","retry_delay_ms":500,"limit":1,"error_message":"auto requeue"}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Affected int64 `json:"affected"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Affected != 1 {
		t.Fatalf("expected affected 1, got %+v", body)
	}
	first, err := store.GetRuntimeTask("batch-api-1")
	if err != nil {
		t.Fatalf("get first task: %v", err)
	}
	second, err := store.GetRuntimeTask("batch-api-2")
	if err != nil {
		t.Fatalf("get second task: %v", err)
	}
	if first.Status != store.RuntimeTaskStatusReleased {
		t.Fatalf("expected first task released, got %+v", first)
	}
	if second.Status != store.RuntimeTaskStatusHeartbeatTimeout {
		t.Fatalf("expected second task unchanged due to limit, got %+v", second)
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

func TestMakeStartRuntimeTaskHandlerTransitionsClaimedTaskToRunning(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	row := &store.RuntimeTaskModel{TaskID: "start-api", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusPending, PayloadJSON: `{"npc":"merchant"}`}
	if err := store.CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	claimed, err := store.ClaimRuntimeTask("start-api", "bridge", "bridge-worker-1")
	if err != nil {
		t.Fatalf("claim task: %v", err)
	}
	h := MakeStartRuntimeTaskHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/start", strings.NewReader(`{"task_id":"start-api","lease_token":"`+claimed.LeaseToken+`"}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Task store.RuntimeTaskModel `json:"task"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal start response: %v", err)
	}
	if body.Task.Status != store.RuntimeTaskStatusRunning {
		t.Fatalf("expected running status, got %q", body.Task.Status)
	}
}

func TestMakeRequeueRuntimeTaskHandlerMovesHeartbeatTimeoutTaskToReleased(t *testing.T) {
	initRuntimeTaskAPITestDB(t)
	now := time.Now()
	row := &store.RuntimeTaskModel{TaskID: "requeue-api", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusHeartbeatTimeout, PayloadJSON: `{}`, HeartbeatTimeoutAt: &now, ErrorMessage: "heartbeat timeout"}
	if err := store.CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	h := MakeRequeueRuntimeTaskHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/requeue", strings.NewReader(`{"task_id":"requeue-api","retry_delay_ms":1500,"error_message":"manual requeue"}`))
	w := httptest.NewRecorder()
	h(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Task store.RuntimeTaskModel `json:"task"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal requeue response: %v", err)
	}
	if body.Task.Status != store.RuntimeTaskStatusReleased {
		t.Fatalf("expected released status, got %q", body.Task.Status)
	}
	if body.Task.HeartbeatTimeoutAt != nil {
		t.Fatalf("expected timeout cleared, got %+v", body.Task.HeartbeatTimeoutAt)
	}
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/runtime/tasks/requeue", strings.NewReader(`{"task_id":"requeue-api"}`))
	w2 := httptest.NewRecorder()
	h(w2, req2)
	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", w2.Code, w2.Body.String())
	}
	if !strings.Contains(w2.Body.String(), "runtime_task_not_requeueable") {
		t.Fatalf("expected runtime_task_not_requeueable, got %s", w2.Body.String())
	}
}
