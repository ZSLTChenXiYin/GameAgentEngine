package service

import (
	"errors"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

func getMemoryTx(tx *gorm.DB, id string) (*store.MemoryModel, error) {
	var memory store.MemoryModel
	if err := tx.Where("uuid = ?", id).First(&memory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorf(ErrorMemoryNotFound, "memory not found: %s", id)
		}
		return nil, err
	}
	return &memory, nil
}
