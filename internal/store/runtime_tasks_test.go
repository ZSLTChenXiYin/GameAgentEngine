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
