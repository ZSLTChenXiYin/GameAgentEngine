package store

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateInferenceLogsToLogsCopiesLegacyRows(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%d.db", t.Name(), time.Now().UnixNano()))
	db, err := gorm.Open(sqlite.New(sqlite.Config{DSN: dsn, DriverName: "sqlite"}), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	DB = db
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	defer sqlDB.Close()
	if err := db.AutoMigrate(&InferenceLogModel{}); err != nil {
		t.Fatalf("migrate logs table: %v", err)
	}
	if err := DB.Exec(`DROP TABLE IF EXISTS inference_logs`).Error; err != nil {
		t.Fatalf("drop legacy table: %v", err)
	}
	if err := DB.Exec(`CREATE TABLE IF NOT EXISTS inference_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid TEXT,
		world_id INTEGER,
		task_type TEXT,
		node_id INTEGER,
		request_data TEXT,
		response_data TEXT,
		llm_model TEXT,
		tokens_used INTEGER,
		duration_ms INTEGER,
		created_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	now := time.Now().UTC()
	if err := DB.Exec(`INSERT INTO inference_logs (uuid, world_id, task_type, node_id, request_data, response_data, llm_model, tokens_used, duration_ms, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"legacy-1", 11, "world_tick", 22, `{"hello":1}`, `{"ok":true}`, "stub", 7, 99, now).Error; err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}
	if err := migrateInferenceLogsToLogs(DB); err != nil {
		t.Fatalf("migrate logs: %v", err)
	}
	var rows []InferenceLogModel
	if err := DB.Find(&rows).Error; err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 copied row, got %d", len(rows))
	}
	if rows[0].EventName != "legacy_inference" {
		t.Fatalf("expected legacy_inference event, got %q", rows[0].EventName)
	}
	if rows[0].RequestData == "" || rows[0].ResponseData == "" {
		t.Fatalf("expected request/response data to be copied: %#v", rows[0])
	}
}

func TestInferenceLogSinkFlushPersistsQueuedRows(t *testing.T) {
	ConfigureLogSink(LogSinkOptions{
		Enabled:       true,
		BatchSize:     8,
		FlushInterval: time.Hour,
		QueueSize:     32,
	})
	t.Cleanup(func() {
		_ = CloseLogSink()
	})

	if err := Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}

	if err := CreateInferenceLog(&InferenceLogModel{WorldID: 1, TaskType: "world_tick", EventName: "queued_1"}); err != nil {
		t.Fatalf("enqueue log 1: %v", err)
	}
	if err := CreateInferenceLog(&InferenceLogModel{WorldID: 1, TaskType: "world_tick", EventName: "queued_2"}); err != nil {
		t.Fatalf("enqueue log 2: %v", err)
	}

	var count int64
	if err := DB.Model(&InferenceLogModel{}).Where("task_type = ?", "world_tick").Count(&count).Error; err != nil {
		t.Fatalf("count logs before flush: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected queued logs to stay buffered before flush, got %d", count)
	}

	if err := FlushLogSink(); err != nil {
		t.Fatalf("flush log sink: %v", err)
	}

	logs, err := GetInferenceLogsByQuery(InferenceLogQuery{TaskType: "world_tick", Limit: 10})
	if err != nil {
		t.Fatalf("query logs after flush: %v", err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 flushed logs, got %d", len(logs))
	}
}

func TestInferenceLogSinkDisabledFallsBackToDirectWrite(t *testing.T) {
	ConfigureLogSink(LogSinkOptions{
		Enabled:       false,
		BatchSize:     8,
		FlushInterval: time.Hour,
		QueueSize:     32,
	})
	t.Cleanup(func() {
		_ = CloseLogSink()
	})

	if err := Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}

	if err := CreateInferenceLog(&InferenceLogModel{WorldID: 2, TaskType: "custom", EventName: "direct_write"}); err != nil {
		t.Fatalf("create log: %v", err)
	}

	logs, err := GetInferenceLogsByQuery(InferenceLogQuery{TaskType: "custom", Limit: 10})
	if err != nil {
		t.Fatalf("query logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected direct write to persist immediately, got %d", len(logs))
	}
}
