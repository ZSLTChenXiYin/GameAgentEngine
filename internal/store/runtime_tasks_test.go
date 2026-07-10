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

func ptrTime(v time.Time) *time.Time { return &v }
