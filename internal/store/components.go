package store

// CreateComponent 创建一个组件记录。
func CreateComponent(m *ComponentModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return DB.Create(m).Error
}

// GetNodeComponents 获取某个节点挂载的全部组件。
func GetNodeComponents(nodeUUID string) ([]ComponentModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []ComponentModel
	err := DB.Where("node_id = ?", nodeID).Find(&list).Error
	return list, err
}

// GetComponentsByType 获取节点上指定类型的组件集合。
func GetComponentsByType(nodeUUID string, compType string) ([]ComponentModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []ComponentModel
	err := DB.Where("node_id = ? AND component_type = ?", nodeID, compType).Find(&list).Error
	return list, err
}

// GetComponentsByTypeForWorld 获取指定世界内所有某类型组件。
func GetComponentsByTypeForWorld(worldUUID string, compType string) ([]ComponentModel, error) {
	worldID := ResolveWorldUUID(worldUUID)
	var list []ComponentModel
	err := DB.Model(&ComponentModel{}).
		Joins("JOIN nodes ON nodes.id = components.node_id").
		Where("nodes.world_id = ? AND nodes.deleted_at IS NULL AND components.component_type = ?", worldID, compType).
		Find(&list).Error
	return list, err
}

// GetComponent 按组件 UUID 查询单个组件。
func GetComponent(uuid string) (*ComponentModel, error) {
	var m ComponentModel
	err := DB.Where("uuid = ?", uuid).First(&m).Error
	return &m, err
}

// UpdateComponent 更新组件类型或数据。
func UpdateComponent(uuid string, updates map[string]any) error {
	return DB.Model(&ComponentModel{}).Where("uuid = ?", uuid).Updates(updates).Error
}
