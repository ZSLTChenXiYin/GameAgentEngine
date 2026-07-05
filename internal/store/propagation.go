package store

// GetPropagationChains 查询指定世界的所有启用规则链。
func GetPropagationChains(worldUUID string) ([]PropagationChainModel, error) {
	worldID := ResolveWorldUUID(worldUUID)
	var list []PropagationChainModel
	err := DB.Where("world_id = ? AND enabled = ?", worldID, true).Find(&list).Error
	if err == nil && len(list) > 0 {
		resolvePropagationWorldUUIDs(list)
	}
	return list, err
}

// GetAllPropagationChains 查询指定世界的所有规则链（含禁用）。
func GetAllPropagationChains(worldUUID string) ([]PropagationChainModel, error) {
	worldID := ResolveWorldUUID(worldUUID)
	var list []PropagationChainModel
	err := DB.Where("world_id = ?", worldID).Find(&list).Error
	if err == nil && len(list) > 0 {
		resolvePropagationWorldUUIDs(list)
	}
	return list, err
}

// CreatePropagationChain 创建规则链。
func CreatePropagationChain(m *PropagationChainModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return DB.Create(m).Error
}

// UpdatePropagationChain 更新规则链。
func UpdatePropagationChain(uuid string, updates map[string]any) error {
	return DB.Model(&PropagationChainModel{}).Where("uuid = ?", uuid).Updates(updates).Error
}

// DeletePropagationChain 删除规则链。
func DeletePropagationChain(uuid string) error {
	return DB.Where("uuid = ?", uuid).Delete(&PropagationChainModel{}).Error
}
