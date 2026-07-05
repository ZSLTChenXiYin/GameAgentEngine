package store

// CreateRelation 创建一条有向关系记录。
func CreateRelation(m *RelationModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return DB.Create(m).Error
}

// GetAllRelations 获取关系列表，支持分页和按类型过滤。
func GetAllRelations(worldUUID string, limit, offset int, relationType string) ([]RelationModel, error) {
	var list []RelationModel
	q := DB.Order("created_at ASC")
	if worldUUID != "" {
		worldID := ResolveWorldUUID(worldUUID)
		q = q.Where("world_id = ?", worldID)
	}
	if relationType != "" {
		q = q.Where("relation_type = ?", relationType)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	err := q.Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveRelationRefs(list)
	}
	return list, err
}

// GetNodeRelations 获取与指定节点相关的全部关系。
func GetNodeRelations(nodeUUID string) ([]RelationModel, error) {
	nodeID := ResolveNodeUUID(nodeUUID)
	var list []RelationModel
	err := DB.Where("source_id = ? OR target_id = ?", nodeID, nodeID).Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveRelationRefs(list)
	}
	return list, err
}

// GetRelation 按关系 UUID 查询单条关系。
func GetRelation(uuid string) (*RelationModel, error) {
	var m RelationModel
	err := DB.Where("uuid = ?", uuid).First(&m).Error
	if err == nil {
		l2 := []RelationModel{m}; resolveRelationRefs(l2); m.WorldUUID = l2[0].WorldUUID; m.SourceUUID = l2[0].SourceUUID; m.TargetUUID = l2[0].TargetUUID
	}
	return &m, err
}

// UpdateRelation 更新关系字段。
func UpdateRelation(uuid string, updates map[string]any) error {
	return DB.Model(&RelationModel{}).Where("uuid = ?", uuid).Updates(updates).Error
}

// DeleteRelation 删除指定关系记录。
func DeleteRelation(uuid string) error {
	return DB.Where("uuid = ?", uuid).Delete(&RelationModel{}).Error
}

