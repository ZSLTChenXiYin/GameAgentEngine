package store

import "gorm.io/gorm"

// ResolveNodeParentUUID 查询单个节点的父节点 UUID 并填充（供 service 层使用）。
func ResolveNodeParentUUID(m *NodeModel) {
	if m == nil || m.ParentID == nil {
		return
	}
	list := []NodeModel{*m}
	resolveNodeParentUUIDs(list)
	if list[0].ParentUUID != nil {
		m.ParentUUID = list[0].ParentUUID
	}
}

// CreateNode 创建一条节点记录。
func CreateNode(m *NodeModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return Write(func(db *gorm.DB) error {
		return db.Create(m).Error
	})
}

// GetNode 按节点 UUID 查询单个节点。
func GetNode(uuid string) (*NodeModel, error) {
	var m NodeModel
	err := DB.Where("uuid = ?", uuid).First(&m).Error
	if err == nil {
		list := []NodeModel{m}
		resolveNodeParentUUIDs(list)
		m.ParentUUID = list[0].ParentUUID
		m.WorldUUID = list[0].WorldUUID
	}
	return &m, err
}

// GetAllNodes 获取节点列表，支持分页和按类型过滤。
// worldUUID 非空时按世界过滤；nodeType 非空时按节点类型过滤。
// limit <= 0 表示不限制；offset < 0 时按 0 处理。
func GetAllNodes(worldUUID string, limit, offset int, nodeType string) ([]NodeModel, error) {
	var list []NodeModel
	q := DB.Order("created_at ASC")
	if worldUUID != "" {
		worldID := ResolveWorldUUID(worldUUID)
		q = q.Where("world_id = ?", worldID)
	}
	if nodeType != "" {
		q = q.Where("node_type = ?", nodeType)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	err := q.Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveNodeParentUUIDs(list)
	}
	return list, err
}

// GetChildNodes 获取指定父节点下的直接子节点。
func GetChildNodes(parentUUID string) ([]NodeModel, error) {
	parentID := ResolveNodeUUID(parentUUID)
	var list []NodeModel
	err := DB.Where("parent_id = ?", parentID).Order("created_at ASC").Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveNodeParentUUIDs(list)
	}
	return list, err
}

// GetWorlds 返回所有 world 类型节点。
func GetWorlds() ([]NodeModel, error) {
	var list []NodeModel
	err := DB.Where("node_type = ?", "world").Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveNodeParentUUIDs(list)
	}
	return list, err
}

// DeleteNode 软删除指定节点。
func DeleteNode(uuid string) error {
	return Write(func(db *gorm.DB) error {
		return db.Where("uuid = ?", uuid).Delete(&NodeModel{}).Error
	})
}

// UpdateNode 按字段映射对节点做局部更新。
func UpdateNode(uuid string, updates map[string]any) error {
	return Write(func(db *gorm.DB) error {
		return db.Model(&NodeModel{}).Where("uuid = ?", uuid).Updates(updates).Error
	})
}

// FindNodesByTags 在指定世界内通过节点名称模糊匹配 tag 查找节点（简化实现）。
func FindNodesByTags(worldUUID string, tags []string) ([]NodeModel, error) {
	if len(tags) == 0 {
		return nil, nil
	}
	worldID := ResolveWorldUUID(worldUUID)
	var list []NodeModel
	q := DB.Where("world_id = ?", worldID)
	for _, tag := range tags {
		q = q.Or("name LIKE ?", "%"+tag+"%")
	}
	err := q.Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveNodeParentUUIDs(list)
	}
	return list, err
}

// FindNodesByIDs 按 UUID 列表批量查询节点。
func FindNodesByIDs(uuids []string) ([]NodeModel, error) {
	if len(uuids) == 0 {
		return nil, nil
	}
	var list []NodeModel
	err := DB.Where("uuid IN ?", uuids).Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveNodeParentUUIDs(list)
	}
	return list, err
}
