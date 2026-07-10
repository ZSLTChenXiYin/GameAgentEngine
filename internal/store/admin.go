package store

import (
	"errors"

	"gorm.io/gorm"
)

// ResetAll 清空当前数据库中的全部业务数据。
// 该操作主要用于重建演示世界或测试环境初始化。
func ResetAll() error {
	return WriteTransaction(func(tx *gorm.DB) error {
		models := []any{
			&IdempotencyKeyModel{},
			&RuntimeTaskModel{},
			&PausedExecutionModel{},
			&AsyncCallbackRecordModel{},
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
	return Write(func(db *gorm.DB) error {
		return db.Save(&IdempotencyKeyModel{ID: key, Fingerprint: fingerprint, StatusCode: statusCode, Result: result}).Error
	})
}

// AcquireIdempotencyKey reserves an idempotency key for first use, or returns the existing record.
func AcquireIdempotencyKey(key, fingerprint string) (*IdempotencyKeyModel, bool, error) {
	var existing IdempotencyKeyModel
	created := false
	err := WriteTransaction(func(tx *gorm.DB) error {
		err := tx.First(&existing, "id = ?", key).Error
		if err == nil {
			return nil
		}
		if !IsRecordNotFound(err) {
			return err
		}
		model := IdempotencyKeyModel{ID: key, Fingerprint: fingerprint, StatusCode: 0, Result: ""}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
		existing = model
		created = true
		return nil
	})
	if err != nil {
		return nil, false, err
	}
	return &existing, created, nil
}
