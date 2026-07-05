package service

import (
	"errors"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

func getComponentTx(tx *gorm.DB, id string) (*store.ComponentModel, error) {
	var component store.ComponentModel
	if err := tx.Where("uuid = ?", id).First(&component).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorf(ErrorComponentNotFound, "component not found: %s", id)
		}
		return nil, err
	}
	return &component, nil
}
