package store

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite"
)

var DB *gorm.DB
var writeDB *gorm.DB

func Writer() *gorm.DB {
	if writeDB != nil {
		return writeDB
	}
	return DB
}

func WriteTransaction(fn func(tx *gorm.DB) error) error {
	return withWriteRetry("transaction", func() error {
		return Writer().Transaction(fn)
	})
}

func Init(driver, dsn string) error {
	if err := CloseLogSink(); err != nil {
		return fmt.Errorf("close previous log sink: %w", err)
	}
	setCurrentDriver(driver)

	var dial gorm.Dialector
	var writeDial gorm.Dialector
	isSQLite := strings.EqualFold(driver, "sqlite")

	switch driver {
	case "sqlite":
		dial = sqlite.New(sqlite.Config{DSN: dsn, DriverName: "sqlite"})
		writeDial = sqlite.New(sqlite.Config{DSN: dsn, DriverName: "sqlite"})
	case "mysql":
		dial = mysql.Open(dsn)
		writeDial = mysql.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	gormLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
	})

	var err error
	DB, err = gorm.Open(dial, &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: isSQLite,
	})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	writeDB, err = gorm.Open(writeDial, &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: isSQLite,
	})
	if err != nil {
		return fmt.Errorf("open write db: %w", err)
	}

	readSQLDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	writerSQLDB, err := writeDB.DB()
	if err != nil {
		return fmt.Errorf("get write sql db: %w", err)
	}

	if isSQLite {
		readSQLDB.SetMaxOpenConns(4)
		readSQLDB.SetMaxIdleConns(4)
		writerSQLDB.SetMaxOpenConns(1)
		writerSQLDB.SetMaxIdleConns(1)
		if err := configureSQLiteHandle(DB); err != nil {
			return err
		}
		if err := configureSQLiteHandle(writeDB); err != nil {
			return err
		}
	} else {
		readSQLDB.SetMaxOpenConns(10)
		readSQLDB.SetMaxIdleConns(5)
		writerSQLDB.SetMaxOpenConns(10)
		writerSQLDB.SetMaxIdleConns(5)
	}

	readSQLDB.SetConnMaxLifetime(time.Hour)
	writerSQLDB.SetConnMaxLifetime(time.Hour)

	if err := Writer().AutoMigrate(
		&NodeModel{},
		&ComponentModel{},
		&RelationModel{},
		&MemoryModel{},
		&TimelineModel{},
		&InferenceLogModel{},
		&IdempotencyKeyModel{},
		&WorldSnapshotModel{},
		&WorldPolicyModel{},
		&WorldSettingsModel{},
		&PropagationChainModel{},
	); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := migrateInferenceLogsToLogs(Writer()); err != nil {
		return fmt.Errorf("migrate logs: %w", err)
	}
	initLogSink()
	return nil
}

func configureSQLiteHandle(db *gorm.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA synchronous=NORMAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA foreign_keys=ON;",
	}
	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			return fmt.Errorf("configure sqlite pragma %q: %w", pragma, err)
		}
	}
	return nil
}

func migrateInferenceLogsToLogs(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if !db.Migrator().HasTable("inference_logs") || !db.Migrator().HasTable("logs") {
		return nil
	}
	var count int64
	if err := db.Table("logs").Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	type legacyInferenceLogRow struct {
		ID           int64
		UUID         string
		WorldID      int64
		TaskType     string
		NodeID       *int64
		RequestData  string
		ResponseData string
		LLMModel     string
		TokensUsed   int
		DurationMs   int64
		CreatedAt    time.Time
	}
	var legacy []legacyInferenceLogRow
	if err := db.Table("inference_logs").Find(&legacy).Error; err != nil {
		return err
	}
	if len(legacy) == 0 {
		return nil
	}
	rows := make([]InferenceLogModel, 0, len(legacy))
	for _, item := range legacy {
		rows = append(rows, InferenceLogModel{
			ID:           item.ID,
			UUID:         item.UUID,
			WorldID:      item.WorldID,
			TaskType:     item.TaskType,
			NodeID:       item.NodeID,
			Category:     "pipeline",
			EventName:    "legacy_inference",
			LogLevel:     "info",
			RequestData:  item.RequestData,
			ResponseData: item.ResponseData,
			LLMModel:     item.LLMModel,
			TokensUsed:   item.TokensUsed,
			DurationMs:   item.DurationMs,
			CreatedAt:    item.CreatedAt,
		})
	}
	return db.Table("logs").Create(&rows).Error
}

func NewUUID() string {
	return uuid.NewString()
}

func ResolveNodeUUID(uuid string) int64 {
	var m NodeModel
	if err := DB.Select("id").Where("uuid = ?", uuid).First(&m).Error; err != nil {
		return 0
	}
	return m.ID
}

func ResolveWorldUUID(uuid string) int64 {
	return ResolveNodeUUID(uuid)
}
