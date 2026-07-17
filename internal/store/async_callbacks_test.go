package store

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupPausedExecDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&PausedExecutionModel{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := db.AutoMigrate(&AsyncCallbackRecordModel{}); err != nil {
		t.Fatalf("migrate callback: %v", err)
	}
	DB = db
}

func makePausedExec(execID, worldID, nodeID string) *PausedExecutionModel {
	now := time.Now().UTC()
	return &PausedExecutionModel{
		ExecutionID:            execID,
		RequestID:              "req-" + execID,
		WorldUUID:              worldID,
		NodeUUID:               nodeID,
		TaskType:               "world_tick",
		ExecutionMode:          "production",
		ConfiguredPipelineMode: "full",
		EffectivePipelineMode:  "full",
		Status:                 "paused",
		PausedRound:            2,
		MaxRounds:              5,
		TargetNodeID:           nodeID,
		PauseReason:            "game_client_request_data",
		CallbackID:             "cb-" + execID,
		OriginalRequestJSON:    `{"world_id":"` + worldID + `"}`,
		BuiltContextJSON:       `{"node":{"uuid":"` + nodeID + `"}}`,
		RuntimeJSON:            `{"max_rounds":5}`,
		RoundStateJSON:         `{"round":2}`,
		PendingDataRequestJSON: `{"label":"test","target":"store","queries":[{"type":"node_components","node_id":"` + nodeID + `"}]}`,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

func TestCreateAndGetPausedExecution(t *testing.T) {
	setupPausedExecDB(t)

	m := makePausedExec("exec-1", "world-1", "node-1")
	if err := CreatePausedExecution(m); err != nil {
		t.Fatalf("create paused exec: %v", err)
	}

	got, err := GetPausedExecution("exec-1")
	if err != nil {
		t.Fatalf("get paused exec: %v", err)
	}
	if got.Status != "paused" {
		t.Fatalf("expected status paused, got %s", got.Status)
	}
	if got.PausedRound != 2 {
		t.Fatalf("expected round 2, got %d", got.PausedRound)
	}
	if got.WorldUUID != "world-1" {
		t.Fatalf("expected world-1, got %s", got.WorldUUID)
	}
}

func TestGetPausedExecutionByCallbackID(t *testing.T) {
	setupPausedExecDB(t)

	m := makePausedExec("exec-2", "world-1", "node-1")
	if err := CreatePausedExecution(m); err != nil {
		t.Fatalf("create paused exec: %v", err)
	}

	got, err := GetPausedExecutionByCallbackID("cb-exec-2")
	if err != nil {
		t.Fatalf("get by callback: %v", err)
	}
	if got.ExecutionID != "exec-2" {
		t.Fatalf("expected exec-2, got %s", got.ExecutionID)
	}

	// Non-existent callback should fail
	if _, err := GetPausedExecutionByCallbackID("nonexistent"); err == nil {
		t.Fatal("expected error for non-existent callback")
	}
}

func TestMarkPausedExecutionResumed(t *testing.T) {
	setupPausedExecDB(t)

	m := makePausedExec("exec-3", "world-1", "node-1")
	if err := CreatePausedExecution(m); err != nil {
		t.Fatalf("create paused exec: %v", err)
	}

	if err := MarkPausedExecutionResumed("exec-3", `{"callback_result":"ok"}`); err != nil {
		t.Fatalf("mark resumed: %v", err)
	}

	got, _ := GetPausedExecution("exec-3")
	if got.Status != "resuming" {
		t.Fatalf("expected status resuming, got %s", got.Status)
	}
	if got.ResumedAt == nil {
		t.Fatal("expected resumed_at to be set")
	}
	if got.ResumePayloadJSON != `{"callback_result":"ok"}` {
		t.Fatalf("unexpected resume payload: %s", got.ResumePayloadJSON)
	}
}

func TestMarkPausedExecutionCompleted(t *testing.T) {
	setupPausedExecDB(t)

	m := makePausedExec("exec-4", "world-1", "node-1")
	if err := CreatePausedExecution(m); err != nil {
		t.Fatalf("create paused exec: %v", err)
	}

	if err := MarkPausedExecutionCompleted("exec-4"); err != nil {
		t.Fatalf("mark completed: %v", err)
	}

	got, _ := GetPausedExecution("exec-4")
	if got.Status != "completed" {
		t.Fatalf("expected status completed, got %s", got.Status)
	}
	if got.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
	if got.LastError != "" {
		t.Fatalf("expected empty last_error, got %s", got.LastError)
	}
}

func TestMarkPausedExecutionFailed(t *testing.T) {
	setupPausedExecDB(t)

	m := makePausedExec("exec-5", "world-1", "node-1")
	if err := CreatePausedExecution(m); err != nil {
		t.Fatalf("create paused exec: %v", err)
	}

	if err := MarkPausedExecutionFailed("exec-5", "something went wrong"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	got, _ := GetPausedExecution("exec-5")
	if got.Status != "failed" {
		t.Fatalf("expected status failed, got %s", got.Status)
	}
	if got.LastError != "something went wrong" {
		t.Fatalf("expected error message, got %s", got.LastError)
	}
}

func TestCreateAndCompleteAsyncCallbackRecord(t *testing.T) {
	setupPausedExecDB(t)

	rec := &AsyncCallbackRecordModel{
		CallbackID: "cb-test-1",
		ActionID:   "adjust_relation",
		Status:     "pending",
		NodeUUID:   "node-1",
		WorldUUID:  "world-1",
		ArgsJSON:   `{"target":"node-2","delta":-10}`,
	}
	if err := CreateAsyncCallbackRecord(rec); err != nil {
		t.Fatalf("create callback: %v", err)
	}

	if err := CompleteAsyncCallbackRecord("cb-test-1", "completed", `{"ok":true}`, ""); err != nil {
		t.Fatalf("complete callback: %v", err)
	}

	got, err := GetAsyncCallbackRecord("cb-test-1")
	if err != nil {
		t.Fatalf("get callback: %v", err)
	}
	if got.Status != "completed" {
		t.Fatalf("expected status completed, got %s", got.Status)
	}
	if got.CompletedAt == nil {
		t.Fatal("expected completed_at to be set")
	}
}

func TestMarkAsyncCallbackPostProcessed(t *testing.T) {
	setupPausedExecDB(t)

	rec := &AsyncCallbackRecordModel{
		CallbackID: "cb-post-1",
		ActionID:   "spawn_item",
		Status:     "completed",
		NodeUUID:   "node-1",
		WorldUUID:  "world-1",
	}
	if err := CreateAsyncCallbackRecord(rec); err != nil {
		t.Fatalf("create callback: %v", err)
	}

	if err := MarkAsyncCallbackPostProcessed("cb-post-1", AsyncCallbackPostProcessSucceeded, `{"item_id":"item-1"}`, ""); err != nil {
		t.Fatalf("mark post-processed: %v", err)
	}

	got, _ := GetAsyncCallbackRecord("cb-post-1")
	if got.PostProcessStatus != AsyncCallbackPostProcessSucceeded {
		t.Fatalf("expected post-process status %s, got %s", AsyncCallbackPostProcessSucceeded, got.PostProcessStatus)
	}
	if got.PostProcessedAt == nil {
		t.Fatal("expected post_processed_at to be set")
	}
	if got.PostProcessResult != `{"item_id":"item-1"}` {
		t.Fatalf("unexpected post-process result: %s", got.PostProcessResult)
	}
}

func TestPausedExecutionFullLifecycle(t *testing.T) {
	setupPausedExecDB(t)

	// Create
	m := makePausedExec("exec-lifecycle", "world-1", "node-1")
	if err := CreatePausedExecution(m); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Resume
	if err := MarkPausedExecutionResumed("exec-lifecycle", `{"data":"returned"}`); err != nil {
		t.Fatalf("resume: %v", err)
	}
	got, _ := GetPausedExecution("exec-lifecycle")
	if got.Status != "resuming" {
		t.Fatalf("expected resuming mid-lifecycle, got %s", got.Status)
	}

	// Complete
	if err := MarkPausedExecutionCompleted("exec-lifecycle"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	got, _ = GetPausedExecution("exec-lifecycle")
	if got.Status != "completed" {
		t.Fatalf("expected completed at end of lifecycle, got %s", got.Status)
	}
}
