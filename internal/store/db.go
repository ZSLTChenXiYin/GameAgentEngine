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
	return nil
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
