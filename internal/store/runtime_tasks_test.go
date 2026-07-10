package store

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func initRuntimeTaskTestDB(t *testing.T) {
	t.Helper()
	if err := Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func TestListPendingRuntimeTasksFiltersByConsumerAndAvailability(t *testing.T) {
	initRuntimeTaskTestDB(t)
	now := time.Now()
	later := now.Add(2 * time.Minute)
	rows := []RuntimeTaskModel{
		{TaskID: "task-1", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Status: RuntimeTaskStatusPending, PayloadJSON: `{"ok":1}`, AvailableAt: &now, Priority: 5},
		{TaskID: "task-2", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Status: RuntimeTaskStatusReleased, PayloadJSON: `{"ok":2}`, AvailableAt: &now, Priority: 1},
		{TaskID: "task-3", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: RuntimeTaskStatusPending, PayloadJSON: `{"ok":3}`, AvailableAt: &now, Priority: 9},
		{TaskID: "task-4", Category: "external_query", InterfaceName: "future_scene", DeliveryMode: "pull", Consumer: "game_client", Status: RuntimeTaskStatusPending, PayloadJSON: `{"ok":4}`, AvailableAt: &later, Priority: 99},
	}
	for i := range rows {
		if err := CreateRuntimeTask(&rows[i]); err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}

	tasks, err := ListPendingRuntimeTasks("game_client", 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	if tasks[0].TaskID != "task-1" || tasks[1].TaskID != "task-2" {
		t.Fatalf("unexpected task order: %+v", tasks)
	}
}

func TestListRuntimeTasksSupportsExtendedFilters(t *testing.T) {
	initRuntimeTaskTestDB(t)
	now := time.Now()
	rows := []RuntimeTaskModel{
		{TaskID: "filter-1", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Transport: "task_pull", WorldUUID: "world-a", Status: RuntimeTaskStatusPending, PayloadJSON: `{}`, AvailableAt: &now},
		{TaskID: "filter-2", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Transport: "task_pull", WorldUUID: "world-b", Status: RuntimeTaskStatusPending, PayloadJSON: `{}`, AvailableAt: &now},
		{TaskID: "filter-3", Category: "external_action", InterfaceName: "spawn_item", DeliveryMode: "push", Consumer: "bridge", Transport: "game_http", WorldUUID: "world-a", Status: RuntimeTaskStatusDispatched, PayloadJSON: `{}`, AvailableAt: &now},
	}
	for i := range rows {
		if err := CreateRuntimeTask(&rows[i]); err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}
	items, err := ListRuntimeTasks(RuntimeTaskListQuery{
		Category:      "external_query",
		InterfaceName: "fetch_scene",
		Transport:     "task_pull",
		WorldUUID:     "world-a",
		Statuses:      []string{RuntimeTaskStatusPending},
		Limit:         10,
	})
	if err != nil {
		t.Fatalf("list runtime tasks: %v", err)
	}
	if len(items) != 1 || items[0].TaskID != "filter-1" {
		t.Fatalf("unexpected filtered tasks: %+v", items)
	}
}

func TestGetRuntimeTaskStatsReturnsStructuredCounts(t *testing.T) {
	initRuntimeTaskTestDB(t)
	now := time.Now()
	old := now.Add(-5 * time.Minute)
	rows := []RuntimeTaskModel{
		{TaskID: "stats-1", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Transport: "task_pull", Status: RuntimeTaskStatusPending, PayloadJSON: `{}`, AvailableAt: &now, CreatedAt: old},
		{TaskID: "stats-2", Category: "external_action", InterfaceName: "spawn_item", DeliveryMode: "hybrid", Consumer: "bridge", Transport: "game_http", Status: RuntimeTaskStatusReleased, PayloadJSON: `{}`, AvailableAt: &now, LastDispatchError: "fallback", CreatedAt: now},
		{TaskID: "stats-3", Category: "external_action", InterfaceName: "spawn_item", DeliveryMode: "push", Consumer: "bridge", Transport: "game_http", Status: RuntimeTaskStatusDispatched, PayloadJSON: `{}`},
		{TaskID: "stats-4", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Transport: "task_pull", Status: RuntimeTaskStatusHeartbeatTimeout, PayloadJSON: `{}`},
		{TaskID: "stats-5", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "push", Consumer: "bridge", Transport: "game_http", Status: RuntimeTaskStatusSucceeded, PayloadJSON: `{}`},
	}
	for i := range rows {
		if err := CreateRuntimeTask(&rows[i]); err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}
	stats, err := GetRuntimeTaskStats()
	if err != nil {
		t.Fatalf("get runtime task stats: %v", err)
	}
	if stats.Total != 5 {
		t.Fatalf("expected total 5, got %d", stats.Total)
	}
	if stats.ReadyPull != 2 {
		t.Fatalf("expected ready pull 2, got %d", stats.ReadyPull)
	}
	if stats.InFlight != 1 {
		t.Fatalf("expected in flight 1, got %d", stats.InFlight)
	}
	if stats.Terminal != 1 {
		t.Fatalf("expected terminal 1, got %d", stats.Terminal)
	}
	if stats.HeartbeatTimeout != 1 {
		t.Fatalf("expected heartbeat timeout 1, got %d", stats.HeartbeatTimeout)
	}
	if stats.DispatchErrorTasks != 1 {
		t.Fatalf("expected dispatch error tasks 1, got %d", stats.DispatchErrorTasks)
	}
	if stats.ByStatus[RuntimeTaskStatusPending] != 1 || stats.ByStatus[RuntimeTaskStatusReleased] != 1 {
		t.Fatalf("unexpected by_status: %+v", stats.ByStatus)
	}
	if stats.ByInterface["spawn_item"] != 2 {
		t.Fatalf("unexpected by_interface: %+v", stats.ByInterface)
	}
	if stats.OldestReadyTaskAgeSecs <= 0 {
		t.Fatalf("expected oldest ready task age > 0, got %d", stats.OldestReadyTaskAgeSecs)
	}
}

func TestClaimRuntimeTaskMarksLeaseAndPreventsDoubleClaim(t *testing.T) {
	initRuntimeTaskTestDB(t)
	row := &RuntimeTaskModel{TaskID: "claim-me", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Status: RuntimeTaskStatusPending, PayloadJSON: `{"scene":"tavern"}`}
	if err := CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}

	claimed, err := ClaimRuntimeTask("claim-me", "game_client", "npc-client-1")
	if err != nil {
		t.Fatalf("claim task: %v", err)
	}
	if claimed.Status != RuntimeTaskStatusClaimed {
		t.Fatalf("expected claimed status, got %q", claimed.Status)
	}
	if claimed.LeaseOwner != "npc-client-1" || claimed.LeaseToken == "" {
		t.Fatalf("expected lease fields to be populated, got %+v", claimed)
	}
	if claimed.AttemptCount != 1 {
		t.Fatalf("expected attempt count 1, got %d", claimed.AttemptCount)
	}

	_, err = ClaimRuntimeTask("claim-me", "game_client", "npc-client-2")
	if !errors.Is(err, ErrRuntimeTaskNotClaimable) {
		t.Fatalf("expected not claimable error, got %v", err)
	}
}

func TestHeartbeatAndReleaseRuntimeTaskRequireMatchingLease(t *testing.T) {
	initRuntimeTaskTestDB(t)
	row := &RuntimeTaskModel{TaskID: "lease-task", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: RuntimeTaskStatusPending, PayloadJSON: `{"npc":"guard"}`}
	if err := CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	claimed, err := ClaimRuntimeTask("lease-task", "bridge", "bridge-1")
	if err != nil {
		t.Fatalf("claim task: %v", err)
	}

	heartbeat, err := HeartbeatRuntimeTask("lease-task", claimed.LeaseToken)
	if err != nil {
		t.Fatalf("heartbeat task: %v", err)
	}
	if heartbeat.LastHeartbeatAt == nil {
		t.Fatalf("expected heartbeat timestamp, got %+v", heartbeat)
	}

	_, err = HeartbeatRuntimeTask("lease-task", "wrong-token")
	if !errors.Is(err, ErrRuntimeTaskLeaseMismatch) {
		t.Fatalf("expected lease mismatch on heartbeat, got %v", err)
	}

	released, err := ReleaseRuntimeTask("lease-task", claimed.LeaseToken, 3*time.Second, "temporary failure")
	if err != nil {
		t.Fatalf("release task: %v", err)
	}
	if released.Status != RuntimeTaskStatusReleased {
		t.Fatalf("expected released status, got %q", released.Status)
	}
	if released.LeaseToken != "" || released.LeaseOwner != "" {
		t.Fatalf("expected lease cleared, got %+v", released)
	}
	if released.ErrorMessage != "temporary failure" {
		t.Fatalf("unexpected release error message: %q", released.ErrorMessage)
	}
	if released.AvailableAt == nil || !released.AvailableAt.After(time.Now().Add(2*time.Second)) {
		t.Fatalf("expected delayed requeue, got %+v", released.AvailableAt)
	}

	_, err = ReleaseRuntimeTask("lease-task", claimed.LeaseToken, 0, "again")
	if !errors.Is(err, ErrRuntimeTaskLeaseMismatch) {
		t.Fatalf("expected lease mismatch on second release, got %v", err)
	}
}

func TestStartRuntimeTaskTransitionsClaimedToRunningAndHeartbeatKeepsItAlive(t *testing.T) {
	initRuntimeTaskTestDB(t)
	row := &RuntimeTaskModel{TaskID: "run-task", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: RuntimeTaskStatusPending, PayloadJSON: `{"npc":"guard"}`}
	if err := CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	claimed, err := ClaimRuntimeTask("run-task", "bridge", "bridge-1")
	if err != nil {
		t.Fatalf("claim task: %v", err)
	}
	running, err := StartRuntimeTask("run-task", claimed.LeaseToken)
	if err != nil {
		t.Fatalf("start task: %v", err)
	}
	if running.Status != RuntimeTaskStatusRunning {
		t.Fatalf("expected running status, got %q", running.Status)
	}
	heartbeat, err := HeartbeatRuntimeTask("run-task", claimed.LeaseToken)
	if err != nil {
		t.Fatalf("heartbeat running task: %v", err)
	}
	if heartbeat.Status != RuntimeTaskStatusRunning {
		t.Fatalf("expected running status after heartbeat, got %q", heartbeat.Status)
	}
}

func TestMarkRuntimeTaskDispatchedTransitionsPendingTask(t *testing.T) {
	initRuntimeTaskTestDB(t)
	row := &RuntimeTaskModel{TaskID: "dispatch-task", Category: "external_query", InterfaceName: "game_client_request_data", DeliveryMode: "push", Consumer: "game_client", Transport: "game_http", Status: RuntimeTaskStatusPending, PayloadJSON: `{}`}
	if err := CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	dispatched, err := MarkRuntimeTaskDispatched("dispatch-task", "game_http", "task-key-1", 1, map[string]any{"status": "accepted"})
	if err != nil {
		t.Fatalf("mark dispatched: %v", err)
	}
	if dispatched.Status != RuntimeTaskStatusDispatched {
		t.Fatalf("expected dispatched status, got %q", dispatched.Status)
	}
	if dispatched.Transport != "game_http" {
		t.Fatalf("expected transport game_http, got %q", dispatched.Transport)
	}
	if dispatched.DispatchedAt == nil {
		t.Fatal("expected dispatched_at to be set")
	}
	if dispatched.LastDispatchAt == nil || dispatched.DispatchAttempts != 1 {
		t.Fatalf("expected dispatch attempt tracking, got %+v", dispatched)
	}
	if dispatched.IdempotencyKey != "task-key-1" {
		t.Fatalf("expected idempotency key to be recorded, got %q", dispatched.IdempotencyKey)
	}
	if dispatched.ResultJSON == "" {
		t.Fatal("expected dispatch result to be recorded")
	}
}

func TestCompleteRuntimeTaskByCallbackIDAllowsDispatchedTasks(t *testing.T) {
	initRuntimeTaskTestDB(t)
	row := &RuntimeTaskModel{TaskID: "dispatch-callback-task", Category: "external_action", InterfaceName: "spawn_item", DeliveryMode: "push", Consumer: "bridge", Transport: "game_http", CallbackID: "cb-1", Status: RuntimeTaskStatusDispatched, PayloadJSON: `{}`}
	if err := CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	if err := CompleteRuntimeTaskByCallbackID("cb-1", "success", map[string]any{"ok": true}); err != nil {
		t.Fatalf("complete task: %v", err)
	}
	completed, err := GetRuntimeTask("dispatch-callback-task")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if completed.Status != RuntimeTaskStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", completed.Status)
	}
}

func TestReleaseRuntimeTaskAllowsRunningTasks(t *testing.T) {
	initRuntimeTaskTestDB(t)
	row := &RuntimeTaskModel{TaskID: "running-release", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: RuntimeTaskStatusPending, PayloadJSON: `{}`}
	if err := CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	claimed, err := ClaimRuntimeTask("running-release", "bridge", "bridge-1")
	if err != nil {
		t.Fatalf("claim task: %v", err)
	}
	if _, err := StartRuntimeTask("running-release", claimed.LeaseToken); err != nil {
		t.Fatalf("start task: %v", err)
	}
	released, err := ReleaseRuntimeTask("running-release", claimed.LeaseToken, time.Second, "worker restart")
	if err != nil {
		t.Fatalf("release task: %v", err)
	}
	if released.Status != RuntimeTaskStatusReleased {
		t.Fatalf("expected released status, got %q", released.Status)
	}
}

func TestMarkRuntimeTasksHeartbeatTimeoutMarksStaleClaimedAndRunningTasks(t *testing.T) {
	initRuntimeTaskTestDB(t)
	old := time.Now().Add(-10 * time.Minute)
	rows := []RuntimeTaskModel{
		{TaskID: "stale-claimed", Category: "external_query", InterfaceName: "fetch_scene", DeliveryMode: "pull", Consumer: "game_client", Status: RuntimeTaskStatusClaimed, LeaseOwner: "client-1", LeaseToken: "tok-1", PayloadJSON: `{}`, LastHeartbeatAt: &old},
		{TaskID: "stale-running", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: RuntimeTaskStatusRunning, LeaseOwner: "bridge-1", LeaseToken: "tok-2", PayloadJSON: `{}`, LastHeartbeatAt: &old},
		{TaskID: "fresh-running", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: RuntimeTaskStatusRunning, LeaseOwner: "bridge-2", LeaseToken: "tok-3", PayloadJSON: `{}`, LastHeartbeatAt: ptrTime(time.Now())},
	}
	for i := range rows {
		if err := CreateRuntimeTask(&rows[i]); err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}
	affected, err := MarkRuntimeTasksHeartbeatTimeout(2 * time.Minute)
	if err != nil {
		t.Fatalf("mark timeout: %v", err)
	}
	if affected != 2 {
		t.Fatalf("expected 2 timed out tasks, got %d", affected)
	}
	staleClaimed, err := GetRuntimeTask("stale-claimed")
	if err != nil {
		t.Fatalf("get stale claimed: %v", err)
	}
	if staleClaimed.Status != RuntimeTaskStatusHeartbeatTimeout {
		t.Fatalf("expected heartbeat_timeout, got %q", staleClaimed.Status)
	}
	if staleClaimed.LeaseToken != "" || staleClaimed.LeaseOwner != "" {
		t.Fatalf("expected lease cleared after timeout, got %+v", staleClaimed)
	}
	freshRunning, err := GetRuntimeTask("fresh-running")
	if err != nil {
		t.Fatalf("get fresh running: %v", err)
	}
	if freshRunning.Status != RuntimeTaskStatusRunning {
		t.Fatalf("expected fresh running task unchanged, got %q", freshRunning.Status)
	}
}

func TestRequeueHeartbeatTimeoutTaskMovesTaskBackToReleased(t *testing.T) {
	initRuntimeTaskTestDB(t)
	now := time.Now()
	row := &RuntimeTaskModel{TaskID: "timeout-task", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: RuntimeTaskStatusHeartbeatTimeout, PayloadJSON: `{}`, HeartbeatTimeoutAt: &now, ErrorMessage: "heartbeat timeout"}
	if err := CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	requeued, err := RequeueHeartbeatTimeoutTask("timeout-task", 2*time.Second, "requeued after timeout")
	if err != nil {
		t.Fatalf("requeue task: %v", err)
	}
	if requeued.Status != RuntimeTaskStatusReleased {
		t.Fatalf("expected released status, got %q", requeued.Status)
	}
	if requeued.HeartbeatTimeoutAt != nil {
		t.Fatalf("expected heartbeat timeout cleared, got %+v", requeued.HeartbeatTimeoutAt)
	}
	if requeued.AvailableAt == nil || !requeued.AvailableAt.After(time.Now().Add(1*time.Second)) {
		t.Fatalf("expected delayed availability, got %+v", requeued.AvailableAt)
	}
	if requeued.ErrorMessage != "requeued after timeout" {
		t.Fatalf("unexpected error message: %q", requeued.ErrorMessage)
	}
	_, err = RequeueHeartbeatTimeoutTask("timeout-task", 0, "again")
	if !errors.Is(err, ErrRuntimeTaskNotClaimable) {
		t.Fatalf("expected not requeueable error, got %v", err)
	}
}

func ptrTime(v time.Time) *time.Time { return &v }
