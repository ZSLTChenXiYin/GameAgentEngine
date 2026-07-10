package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initRuntimeTaskGovernorTestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func TestRunRuntimeTaskGovernanceMarksTimeoutAndAutoRequeues(t *testing.T) {
	initRuntimeTaskGovernorTestDB(t)
	old := time.Now().Add(-10 * time.Minute)
	rows := []store.RuntimeTaskModel{
		{TaskID: "gov-1", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusClaimed, LeaseOwner: "bridge-1", LeaseToken: "tok-1", PayloadJSON: `{}`, LastHeartbeatAt: &old},
		{TaskID: "gov-2", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusRunning, LeaseOwner: "bridge-2", LeaseToken: "tok-2", PayloadJSON: `{}`, LastHeartbeatAt: &old},
	}
	for i := range rows {
		if err := store.CreateRuntimeTask(&rows[i]); err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}
	result, err := RunRuntimeTaskGovernance(RuntimeTaskGovernanceOptions{
		HeartbeatTimeout: 2 * time.Minute,
		AutoRequeue:      true,
		AutoRequeueLimit: 10,
		AutoRequeueDelay: time.Second,
	})
	if err != nil {
		t.Fatalf("run runtime task governance: %v", err)
	}
	if result.TimedOut != 2 {
		t.Fatalf("expected timed out 2, got %+v", result)
	}
	if result.Requeued != 2 {
		t.Fatalf("expected requeued 2, got %+v", result)
	}
	first, err := store.GetRuntimeTask("gov-1")
	if err != nil {
		t.Fatalf("get first task: %v", err)
	}
	if first.Status != store.RuntimeTaskStatusReleased {
		t.Fatalf("expected released task, got %+v", first)
	}
	if first.HeartbeatTimeoutAt != nil {
		t.Fatalf("expected cleared heartbeat timeout, got %+v", first)
	}
}

func TestRunRuntimeTaskGovernanceCanOnlyMarkTimeoutWithoutRequeue(t *testing.T) {
	initRuntimeTaskGovernorTestDB(t)
	old := time.Now().Add(-10 * time.Minute)
	row := &store.RuntimeTaskModel{TaskID: "gov-timeout-only", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusClaimed, LeaseOwner: "bridge-1", LeaseToken: "tok-1", PayloadJSON: `{}`, LastHeartbeatAt: &old}
	if err := store.CreateRuntimeTask(row); err != nil {
		t.Fatalf("create task: %v", err)
	}
	result, err := RunRuntimeTaskGovernance(RuntimeTaskGovernanceOptions{HeartbeatTimeout: 2 * time.Minute})
	if err != nil {
		t.Fatalf("run runtime task governance: %v", err)
	}
	if result.TimedOut != 1 || result.Requeued != 0 {
		t.Fatalf("unexpected governance result: %+v", result)
	}
	item, err := store.GetRuntimeTask("gov-timeout-only")
	if err != nil {
		t.Fatalf("get task: %v", err)
	}
	if item.Status != store.RuntimeTaskStatusHeartbeatTimeout {
		t.Fatalf("expected heartbeat timeout task, got %+v", item)
	}
}
