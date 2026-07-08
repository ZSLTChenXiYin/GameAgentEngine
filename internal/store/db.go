// Package store 提供基于 GORM 的持久化能力。
// 该包负责数据库连接、模型迁移以及各实体的读写操作。
package store

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite"
)

var DB *gorm.DB

// Init 根据驱动类型初始化数据库连接并执行自动迁移。
func Init(driver, dsn string) error {
	var dial gorm.Dialector
	switch driver {
	case "sqlite":
		dial = sqlite.New(sqlite.Config{DSN: dsn, DriverName: "sqlite"})
	case "mysql":
		dial = mysql.Open(dsn)
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	var err error
	DB, err = gorm.Open(dial, &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := DB.AutoMigrate(
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
	if err := migrateInferenceLogsToLogs(DB); err != nil {
		return fmt.Errorf("migrate logs: %w", err)
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

// NewUUID 生成统一使用的 UUID 字符串标识。
func NewUUID() string {
	return uuid.NewString()
}

// ResolveNodeUUID 通过 UUID 查询节点的内部 int64 ID。
func ResolveNodeUUID(uuid string) int64 {
	var m NodeModel
	if err := DB.Select("id").Where("uuid = ?", uuid).First(&m).Error; err != nil {
		return 0
	}
	return m.ID
}

// ResolveWorldUUID 通过 UUID 查询世界的内部 int64 ID。
func ResolveWorldUUID(uuid string) int64 {
	return ResolveNodeUUID(uuid)
}
