package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	driverMySQL "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIsRetriableWriteErrorRecognizesSQLiteAndMySQLConflicts(t *testing.T) {
	if !isRetriableWriteError("sqlite", errors.New("database is locked")) {
		t.Fatal("expected sqlite database is locked to be retriable")
	}
	if !isRetriableWriteError("sqlite", errors.New("SQLITE_BUSY: database table is locked")) {
		t.Fatal("expected sqlite busy error to be retriable")
	}
	if !isRetriableWriteError("mysql", &driverMySQL.MySQLError{Number: 1213, Message: "Deadlock found"}) {
		t.Fatal("expected mysql deadlock to be retriable")
	}
	if !isRetriableWriteError("mysql", &driverMySQL.MySQLError{Number: 1205, Message: "Lock wait timeout exceeded"}) {
		t.Fatal("expected mysql lock wait timeout to be retriable")
	}
	if !isRetriableWriteError("postgres", &pgconn.PgError{Code: "40P01", Message: "deadlock detected"}) {
		t.Fatal("expected postgres deadlock to be retriable")
	}
	if isRetriableWriteError("sqlite", errors.New("syntax error")) {
		t.Fatal("expected non-lock sqlite error to be non-retriable")
	}
}

func TestWithWriteRetryRetriesRetriableErrors(t *testing.T) {
	ConfigureWriteRetry(WriteRetryOptions{Enabled: true, MaxAttempts: 3, BaseDelay: 0, MaxDelay: 0})
	t.Cleanup(func() {
		ConfigureWriteRetry(WriteRetryOptions{Enabled: true, MaxAttempts: 3, BaseDelay: 40 * time.Millisecond, MaxDelay: 250 * time.Millisecond})
	})
	setCurrentDriver("sqlite")

	attempts := 0
	err := withWriteRetry("test", func() error {
		attempts++
		if attempts < 3 {
			return errors.New("database is locked")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected retry to recover, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestWithWriteRetryStopsOnNonRetriableError(t *testing.T) {
	ConfigureWriteRetry(WriteRetryOptions{Enabled: true, MaxAttempts: 3, BaseDelay: 0, MaxDelay: 0})
	t.Cleanup(func() {
		ConfigureWriteRetry(WriteRetryOptions{Enabled: true, MaxAttempts: 3, BaseDelay: 40 * time.Millisecond, MaxDelay: 250 * time.Millisecond})
	})
	setCurrentDriver("sqlite")

	attempts := 0
	wantErr := errors.New("invalid query")
	err := withWriteRetry("test", func() error {
		attempts++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected original error, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected non-retriable error to stop after 1 attempt, got %d", attempts)
	}
}

func TestMigrationRunnerRunsStepsInOrder(t *testing.T) {
	var sequence []string
	runner := NewMigrationRunnerWithSteps([]MigrationStep{
		{Name: "one", Run: func(db *gorm.DB) error {
			sequence = append(sequence, "one")
			return nil
		}},
		{Name: "two", Run: func(db *gorm.DB) error {
			sequence = append(sequence, "two")
			return nil
		}},
	})
	if err := runner.Run(&gorm.DB{}); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	if len(sequence) != 2 || sequence[0] != "one" || sequence[1] != "two" {
		t.Fatalf("unexpected migration order: %#v", sequence)
	}
}

func TestMigrationRunnerReturnsStepNameOnFailure(t *testing.T) {
	runner := NewMigrationRunnerWithSteps([]MigrationStep{{
		Name: "failing_step",
		Run: func(db *gorm.DB) error {
			return errors.New("boom")
		},
	}})
	err := runner.Run(&gorm.DB{})
	if err == nil {
		t.Fatal("expected migration failure")
	}
	if !strings.Contains(err.Error(), "failing_step") {
		t.Fatalf("expected step name in error, got %v", err)
	}
}

func TestInitRejectsUnsupportedDriver(t *testing.T) {
	err := Init("oracle", "dsn")
	if err == nil {
		t.Fatal("expected unsupported driver error")
	}
	if !strings.Contains(err.Error(), "unsupported database driver") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunMigrationsCanBeDisabled(t *testing.T) {
	ConfigureMigrationsEnabled(false)
	t.Cleanup(func() {
		ConfigureMigrationsEnabled(true)
	})
	if err := RunMigrations(nil); err != nil {
		t.Fatalf("expected disabled migrations to short-circuit, got %v", err)
	}
}

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
