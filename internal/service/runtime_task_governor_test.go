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
		{TaskID: "gov-1", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusClaimed, LeaseOwner: "bridge-1", LeaseToken: "tok-1", PayloadJSON: `{"heartbeat_timeout_policy":{"auto_requeue":true,"requeue_delay_ms":1000,"reason":"interface auto requeue"}}`, LastHeartbeatAt: &old},
		{TaskID: "gov-2", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusRunning, LeaseOwner: "bridge-2", LeaseToken: "tok-2", PayloadJSON: `{"heartbeat_timeout_policy":{"auto_requeue":true,"requeue_delay_ms":1000,"reason":"interface auto requeue"}}`, LastHeartbeatAt: &old},
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
	if first.ErrorMessage != "interface auto requeue" {
		t.Fatalf("expected interface-specific reason, got %+v", first)
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

func TestRunRuntimeTaskGovernanceHonorsPerTaskAutoRequeuePolicy(t *testing.T) {
	initRuntimeTaskGovernorTestDB(t)
	old := time.Now().Add(-10 * time.Minute)
	rows := []store.RuntimeTaskModel{
		{TaskID: "gov-policy-1", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusClaimed, LeaseOwner: "bridge-1", LeaseToken: "tok-1", PayloadJSON: `{"heartbeat_timeout_policy":{"auto_requeue":false}}`, LastHeartbeatAt: &old},
		{TaskID: "gov-policy-2", Category: "external_action", InterfaceName: "spawn_npc", DeliveryMode: "pull", Consumer: "bridge", Status: store.RuntimeTaskStatusRunning, LeaseOwner: "bridge-2", LeaseToken: "tok-2", PayloadJSON: `{"heartbeat_timeout_policy":{"auto_requeue":true,"requeue_delay_ms":1500,"reason":"selective auto requeue"}}`, LastHeartbeatAt: &old},
	}
	for i := range rows {
		if err := store.CreateRuntimeTask(&rows[i]); err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
	}
	result, err := RunRuntimeTaskGovernance(RuntimeTaskGovernanceOptions{
		HeartbeatTimeout:  2 * time.Minute,
		AutoRequeue:       true,
		AutoRequeueLimit:  10,
		AutoRequeueDelay:  time.Second,
		AutoRequeueReason: "global auto requeue",
	})
	if err != nil {
		t.Fatalf("run runtime task governance: %v", err)
	}
	if result.TimedOut != 2 || result.Requeued != 1 || result.PolicySkipped != 1 {
		t.Fatalf("unexpected governance result: %+v", result)
	}
	first, err := store.GetRuntimeTask("gov-policy-1")
	if err != nil {
		t.Fatalf("get first task: %v", err)
	}
	second, err := store.GetRuntimeTask("gov-policy-2")
	if err != nil {
		t.Fatalf("get second task: %v", err)
	}
	if first.Status != store.RuntimeTaskStatusHeartbeatTimeout {
		t.Fatalf("expected first task to remain heartbeat_timeout, got %+v", first)
	}
	if second.Status != store.RuntimeTaskStatusReleased {
		t.Fatalf("expected second task to be released, got %+v", second)
	}
	if second.ErrorMessage != "selective auto requeue" {
		t.Fatalf("expected policy reason to override global reason, got %+v", second)
	}
	if second.AvailableAt == nil {
		t.Fatalf("expected second task to have delayed availability, got %+v", second)
	}
}
