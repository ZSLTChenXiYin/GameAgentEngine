package service

import (
	"encoding/json"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

func GetStateComponent(nodeID string, componentType engine.ComponentType) (*store.ComponentModel, error) {
	return getStateComponentTx(store.DB, nodeID, componentType)
}

func UpsertStateComponent(nodeID string, componentType engine.ComponentType, payload any) (*store.ComponentModel, error) {
	return upsertStateComponentTx(store.DB, nodeID, componentType, payload)
}

func getStateComponentTx(tx *gorm.DB, nodeID string, componentType engine.ComponentType) (*store.ComponentModel, error) {
	if _, err := getNodeTx(tx, nodeID); err != nil {
		return nil, err
	}
	comps, err := getComponentsByTypeTx(tx, nodeID, string(componentType))
	if err != nil {
		return nil, err
	}
	if len(comps) == 0 {
		return nil, nil
	}
	return &comps[0], nil
}

func upsertStateComponentTx(tx *gorm.DB, nodeID string, componentType engine.ComponentType, payload any) (*store.ComponentModel, error) {
	if _, err := getNodeTx(tx, nodeID); err != nil {
		return nil, err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, invalidf("invalid %s payload: %v", componentType, err)
	}
	if err := ValidateComponentData(string(componentType), string(data)); err != nil {
		return nil, err
	}
	existing, err := getStateComponentTx(tx, nodeID, componentType)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		node, err := getNodeTx(tx, nodeID)
		if err != nil {
			return nil, err
		}
		created := &store.ComponentModel{UUID: store.NewUUID(), NodeID: node.ID, NodeUUID: node.UUID, ComponentType: string(componentType), Data: string(data)}
		if err := tx.Create(created).Error; err != nil {
			return nil, err
		}
		return created, nil
	}
	if err := tx.Model(&store.ComponentModel{}).Where("id = ?", existing.ID).Updates(map[string]any{"data": string(data)}).Error; err != nil {
		return nil, err
	}
	return getStateComponentTx(tx, nodeID, componentType)
}

func getComponentsByTypeTx(tx *gorm.DB, nodeUUID string, compType string) ([]store.ComponentModel, error) {
	node, err := getNodeTx(tx, nodeUUID)
	if err != nil {
		return nil, err
	}
	var list []store.ComponentModel
	if err := tx.Where("node_id = ? AND component_type = ?", node.ID, compType).Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		list[i].NodeUUID = nodeUUID
	}
	return list, nil
}
