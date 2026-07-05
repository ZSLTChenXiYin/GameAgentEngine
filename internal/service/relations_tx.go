package service

import (
	"errors"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

func getRelationTx(tx *gorm.DB, id string) (*store.RelationModel, error) {
	var relation store.RelationModel
	if err := tx.Where("uuid = ?", id).First(&relation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorf(ErrorRelationNotFound, "relation not found: %s", id)
		}
		return nil, err
	}
	return &relation, nil
}
