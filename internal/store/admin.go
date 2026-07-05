package store

import "gorm.io/gorm"

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

// GetIdempotencyResult 查询幂等键的已缓存结果。
func GetIdempotencyResult(key string) (string, error) {
	var m IdempotencyKeyModel
	err := DB.First(&m, "id = ?", key).Error
	if err != nil {
		return "", err
	}
	return m.Result, nil
}

// SetIdempotencyResult 缓存幂等操作的结果。
func SetIdempotencyResult(key string, result string) error {
	return DB.Create(&IdempotencyKeyModel{ID: key, Result: result}).Error
}
