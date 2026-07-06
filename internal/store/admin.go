package store

import (
	"errors"

	"gorm.io/gorm"
)

// ResetAll 清空当前数据库中的全部业务数据。
// 该操作主要用于重建演示世界或测试环境初始化。
func ResetAll() error {
	return DB.Transaction(func(tx *gorm.DB) error {
		models := []any{
			&IdempotencyKeyModel{},
			&InferenceLogModel{},
			&TimelineModel{},
			&RelationModel{},
			&ComponentModel{},
			&MemoryModel{},
			&NodeModel{},
		}

		for _, model := range models {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetIdempotencyResult queries a cached idempotency record by key.
func GetIdempotencyResult(key string) (*IdempotencyKeyModel, error) {
	var m IdempotencyKeyModel
	err := DB.First(&m, "id = ?", key).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func IsRecordNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// SetIdempotencyResult caches the response for an idempotency key.
func SetIdempotencyResult(key, fingerprint string, statusCode int, result string) error {
	return DB.Save(&IdempotencyKeyModel{ID: key, Fingerprint: fingerprint, StatusCode: statusCode, Result: result}).Error
}
