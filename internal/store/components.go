package store

import (
	"fmt"

	"gorm.io/gorm"
)

// CreateComponent 创建一个组件记录。
func CreateComponent(m *ComponentModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return Write(func(db *gorm.DB) error {
		return db.Create(m).Error
	})
}

// GetNodeComponents 获取某个节点挂载的全部组件。
func GetNodeComponents(nodeUUID string) ([]ComponentModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []ComponentModel
	err := DB.Where("node_id = ?", nodeID).Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveComponentNodeUUIDs(list)
	}
	return list, err
}

// GetComponentsByType 获取节点上指定类型的组件集合。
func GetComponentsByType(nodeUUID string, compType string) ([]ComponentModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []ComponentModel
	err := DB.Where("node_id = ? AND component_type = ?", nodeID, compType).Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveComponentNodeUUIDs(list)
	}
	return list, err
}

// GetSingleComponentByType 返回节点上某个类型的首个组件。
func GetSingleComponentByType(nodeUUID string, compType string) (*ComponentModel, error) {
	list, err := GetComponentsByType(nodeUUID, compType)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return &list[0], nil
}

// GetComponentsByTypeForWorld 获取指定世界内所有某类型组件。
func GetComponentsByTypeForWorld(worldUUID string, compType string) ([]ComponentModel, error) {
	worldID := ResolveWorldUUID(worldUUID)
	var list []ComponentModel
	err := DB.Model(&ComponentModel{}).
		Joins("JOIN nodes ON nodes.id = components.node_id").
		Where("nodes.world_id = ? AND nodes.deleted_at IS NULL AND components.component_type = ?", worldID, compType).
		Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveComponentNodeUUIDs(list)
	}
	return list, err
}

// GetComponent 按组件 UUID 查询单个组件。
func GetComponent(uuid string) (*ComponentModel, error) {
	var m ComponentModel
	err := DB.Where("uuid = ?", uuid).First(&m).Error
	if err == nil {
		list := []ComponentModel{m}
		resolveComponentNodeUUIDs(list)
		m.NodeUUID = list[0].NodeUUID
	}
	return &m, err
}

// UpdateComponent 更新组件类型或数据。
func UpdateComponent(uuid string, updates map[string]any) error {
	return Write(func(db *gorm.DB) error {
		return db.Model(&ComponentModel{}).Where("uuid = ?", uuid).Updates(updates).Error
	})
}

// UpsertComponentByType creates or replaces a node-local component by type.
func UpsertComponentByType(nodeUUID string, compType string, data string) (*ComponentModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	if nodeID == 0 {
		return nil, fmt.Errorf("node %s not found", nodeUUID)
	}
	existing, err := GetSingleComponentByType(nodeUUID, compType)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		created := &ComponentModel{UUID: NewUUID(), NodeID: nodeID, NodeUUID: nodeUUID, ComponentType: compType, Data: data}
		if err := CreateComponent(created); err != nil {
			return nil, err
		}
		return GetComponent(created.UUID)
	}
	if err := UpdateComponent(existing.UUID, map[string]any{"data": data}); err != nil {
		return nil, err
	}
	return GetComponent(existing.UUID)
}
