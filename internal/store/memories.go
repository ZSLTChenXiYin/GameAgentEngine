package store

// CreateMemory 为节点写入一条记忆。
func CreateMemory(m *MemoryModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return Writer().Create(m).Error
}

// SearchMemories 按关键词搜索节点的记忆内容（LIKE 模糊匹配）。
func SearchMemories(nodeUUID, keyword string, limit int) ([]MemoryModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []MemoryModel
	q := DB.Where("node_id = ? AND content LIKE ?", nodeID, "%"+keyword+"%").Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveMemoryNodeUUIDs(list)
	}
	return list, err
}

// GetNodeMemories 获取节点最近的记忆列表。
func GetNodeMemories(nodeUUID string, limit int) ([]MemoryModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []MemoryModel
	q := DB.Where("node_id = ?", nodeID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveMemoryNodeUUIDs(list)
	}
	return list, err
}

// GetMemoriesByLevel 按记忆层级筛选节点记忆。
func GetMemoriesByLevel(nodeUUID string, level string) ([]MemoryModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []MemoryModel
	err := DB.Where("node_id = ? AND level = ?", nodeID, level).Order("created_at DESC").Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveMemoryNodeUUIDs(list)
	}
	return list, err
}

// GetMemory 按记忆 UUID 查询单条记忆。
func GetMemory(uuid string) (*MemoryModel, error) {
	var m MemoryModel
	err := DB.Where("uuid = ?", uuid).First(&m).Error
	if err == nil {
		list2 := []MemoryModel{m}; resolveMemoryNodeUUIDs(list2); m.NodeUUID = list2[0].NodeUUID
	}
	return &m, err
}

// UpdateMemory 更新记忆内容、层级或标签。
func UpdateMemory(uuid string, updates map[string]any) error {
	return Writer().Model(&MemoryModel{}).Where("uuid = ?", uuid).Updates(updates).Error
}

// DeleteMemory 删除指定记忆记录。
func DeleteMemory(uuid string) error {
	return Writer().Where("uuid = ?", uuid).Delete(&MemoryModel{}).Error
}

// CreateMemoriesBulk 批量创建多条记忆记录。
func CreateMemoriesBulk(mems []MemoryModel) error {
	for i := range mems {
		if mems[i].UUID == "" {
			mems[i].UUID = NewUUID()
		}
	}
	return Writer().Create(&mems).Error
}

