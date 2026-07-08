package service

import (
	"encoding/json"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func GetStateComponent(nodeID string, componentType engine.ComponentType) (*store.ComponentModel, error) {
	if _, err := getNodeTx(store.DB, nodeID); err != nil {
		return nil, err
	}
	return store.GetSingleComponentByType(nodeID, string(componentType))
}

func UpsertStateComponent(nodeID string, componentType engine.ComponentType, payload any) (*store.ComponentModel, error) {
	if _, err := getNodeTx(store.DB, nodeID); err != nil {
		return nil, err
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, invalidf("invalid %s payload: %v", componentType, err)
	}
	if err := ValidateComponentData(string(componentType), string(data)); err != nil {
		return nil, err
	}
	return store.UpsertComponentByType(nodeID, string(componentType), string(data))
}
