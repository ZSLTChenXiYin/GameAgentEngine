package service

import (
	"gorm.io/gorm"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// ---- transaction-safe UUID resolution helpers ----
func txResolveNodeUUID(tx *gorm.DB, uuid string) int64 {
	var id int64
	tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", uuid).First(&id)
	return id
}

func txResolveWorldUUID(tx *gorm.DB, uuid string) int64 {
	return txResolveNodeUUID(tx, uuid)
}

func deleteByID[T any](model *T, id string, fetch func(*gorm.DB, string) (*T, error)) error {
	return store.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := fetch(tx, id); err != nil {
			return err
		}
		return tx.Where("uuid = ?", id).Delete(model).Error
	})
}
