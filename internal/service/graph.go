package service

import (
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

// ---- public functions ----
// CreateNode 创建节点并校验世界/父子关系是否合法。
func CreateNode(worldID, name, nodeType string, parentID *string) (*store.NodeModel, error) {
	var created *store.NodeModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		var err error
		created, err = createNodeTx(tx, worldID, name, nodeType, parentID)
		return err
	})
	if err == nil && created != nil {
		store.ResolveNodeParentUUID(created)
	}
	return created, err
}

// UpdateNode 更新节点，并确保不会制造跨世界父子链或循环父子链。
func UpdateNode(id string, name, nodeType, parentID *string, parentIDSet bool) (*store.NodeModel, error) {
	var updated *store.NodeModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		node, err := getNodeTx(tx, id)
		if err != nil {
			return err
		}
		updates := map[string]any{}
		if name != nil {
			updates["name"] = *name
		}
		if nodeType != nil {
			if node.NodeType == "world" || *nodeType == "world" {
				if node.NodeType != *nodeType {
					return errorf(ErrorWorldNodeConstraint, "world node type cannot be changed")
				}
			}
			updates["node_type"] = *nodeType
		}
		if parentIDSet {
			if node.NodeType == "world" && parentID != nil {
				return errorf(ErrorWorldNodeConstraint, "world node cannot have a parent")
			}
			if parentID == nil || *parentID == "" {
				updates["parent_id"] = nil
			} else {
				parent, err := getNodeTx(tx, *parentID)
				if err != nil {
					return err
				}
				if parent.WorldID != node.WorldID {
					return errorf(ErrorCrossWorldRelation, "parent node must be in the same world")
				}
				if err := ensureNoParentCycleTx(tx, node.ID, parent.ID); err != nil {
					return err
				}
				updates["parent_id"] = parent.ID
			}
		}
		if len(updates) == 0 {
			return errorf(ErrorNoUpdates, "no node updates provided")
		}
		if err := tx.Model(&store.NodeModel{}).Where("uuid = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		updated, err = getNodeTx(tx, id)
		return err
	})
	if err == nil && updated != nil {
		store.ResolveNodeParentUUID(updated)
	}
	return updated, err
}

// DeleteNode 删除一个叶子节点，并清理挂在其上的附属数据。
func DeleteNode(id string) error {
	return store.WriteTransaction(func(tx *gorm.DB) error {
		node, err := getNodeTx(tx, id)
		if err != nil {
			return err
		}
		var childCount int64
		if err := tx.Model(&store.NodeModel{}).Where("parent_id = ?", node.ID).Count(&childCount).Error; err != nil {
			return err
		}
		if childCount > 0 {
			return errorf(ErrorNodeHasChildren, "node still has child nodes")
		}
		if err := tx.Where("node_id = ?", node.ID).Delete(&store.ComponentModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("node_id = ?", node.ID).Delete(&store.MemoryModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("source_id = ? OR target_id = ?", node.ID, node.ID).Delete(&store.RelationModel{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&store.NodeModel{}, "id = ?", node.ID).Error; err != nil {
			return err
		}
		return nil
	})
}

// CreateComponent 创建组件，并确保目标节点存在。
func CreateComponent(nodeID, componentType, data string) (*store.ComponentModel, error) {
	var created *store.ComponentModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		if !engine.IsValidComponentType(componentType) {
			return errorf(ErrorInvalidComponentType, "invalid component_type: %s", componentType)
		}
		if err := ValidateComponentData(componentType, data); err != nil {
			return err
		}
		node, err := getNodeTx(tx, nodeID)
		if err != nil {
			return err
		}
		created = &store.ComponentModel{
			UUID:          store.NewUUID(),
			NodeID:        node.ID,
			ComponentType: componentType,
			Data:          data,
		}
		return tx.Create(created).Error
	})
	return created, err
}

// UpdateComponent 更新组件内容。
func UpdateComponent(id string, componentType, data *string) (*store.ComponentModel, error) {
	var updated *store.ComponentModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		component, err := getComponentTx(tx, id)
		if err != nil {
			return err
		}
		updates := map[string]any{}
		nextType := component.ComponentType
		if componentType != nil {
			nextType = *componentType
		}
		if componentType != nil {
			updates["component_type"] = *componentType
		}
		if data != nil {
			if err := ValidateComponentData(nextType, *data); err != nil {
				return err
			}
			updates["data"] = *data
		}
		if data == nil && componentType != nil {
			if err := ValidateComponentData(nextType, component.Data); err != nil {
				return err
			}
		}
		if len(updates) == 0 {
			return errorf(ErrorNoUpdates, "no component updates provided")
		}
		if err := tx.Model(&store.ComponentModel{}).Where("id = ?", component.ID).Updates(updates).Error; err != nil {
			return err
		}
		updated, err = getComponentTx(tx, id)
		return err
	})
	return updated, err
}

// DeleteComponent 删除一个组件。
func DeleteComponent(id string) error {
	return deleteByID(&store.ComponentModel{}, id, getComponentTx)
}

// CreateRelation 创建有向关系，写入两条记录（正向和逆向）。
func CreateRelation(worldID, sourceID, targetID, relationType string, weight float64, props string) (*store.RelationModel, error) {
	var created *store.RelationModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		world := txResolveWorldUUID(tx, worldID)
		if weight == 0 {
			weight = 1.0
		}
		source, err := getNodeTx(tx, sourceID)
		if err != nil {
			return err
		}
		target, err := getNodeTx(tx, targetID)
		if err != nil {
			return err
		}
		if source.ID == target.ID {
			return errorf(ErrorInvalidRelationType, "source node cannot point to itself")
		}
		if source.WorldID != world || target.WorldID != world {
			return errorf(ErrorCrossWorldRelation, "both nodes must be in the same world")
		}
		if !engine.IsValidRelationType(relationType) {
			return errorf(ErrorInvalidRelationType, "invalid relation_type: %s", relationType)
		}
		var existing int64
		if err := tx.Model(&store.RelationModel{}).Where("world_id = ? AND source_id = ? AND target_id = ? AND relation_type = ?", world, source.ID, target.ID, relationType).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return conflictf("relation already exists")
		}
		if relationType == string(engine.RelExternalParent) {
			if source.ParentID != nil && *source.ParentID == target.ID {
				return conflictf("target node is already the primary parent")
			}
			if err := ensureNoExternalParentCycleTx(tx, source.ID, target.ID); err != nil {
				return err
			}
		}
		created = &store.RelationModel{
			UUID:         store.NewUUID(),
			WorldID:      world,
			SourceID:     source.ID,
			TargetID:     target.ID,
			RelationType: relationType,
			Weight:       int(weight),
			Properties:   props,
		}
		return tx.Create(created).Error
	})
	return created, err
}

func ensureNoExternalParentCycleTx(tx *gorm.DB, sourceID, targetID int64) error {
	visited := map[int64]bool{}
	var walk func(int64) error
	walk = func(nodeID int64) error {
		if visited[nodeID] {
			return nil
		}
		visited[nodeID] = true
		if nodeID == sourceID {
			return errorf(ErrorParentCycle, "external parent link would create a cycle")
		}
		var node store.NodeModel
		if err := tx.Where("id = ?", nodeID).First(&node).Error; err != nil {
			return err
		}
		if node.ParentID != nil {
			if err := walk(*node.ParentID); err != nil {
				return err
			}
		}
		var extraParents []int64
		if err := tx.Model(&store.RelationModel{}).Where("source_id = ? AND relation_type = ?", nodeID, string(engine.RelExternalParent)).Pluck("target_id", &extraParents).Error; err != nil {
			return err
		}
		for _, parentID := range extraParents {
			if err := walk(parentID); err != nil {
				return err
			}
		}
		return nil
	}
	return walk(targetID)
}

// UpdateRelation 更新关系的内容。
func UpdateRelation(id string, sourceID, targetID, relationType *string, weight *float64, props *string) (*store.RelationModel, error) {
	var updated *store.RelationModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		relation, err := getRelationTx(tx, id)
		if err != nil {
			return err
		}
		sourceNodeID := relation.SourceID
		targetNodeID := relation.TargetID
		updates := map[string]any{}
		if sourceID != nil {
			sourceNode, err := getNodeTx(tx, *sourceID)
			if err != nil {
				return err
			}
			if sourceNode.WorldID != relation.WorldID {
				return errorf(ErrorCrossWorldRelation, "source node must be in the same world")
			}
			sourceNodeID = sourceNode.ID
			updates["source_id"] = sourceNode.ID
		}
		if targetID != nil {
			targetNode, err := getNodeTx(tx, *targetID)
			if err != nil {
				return err
			}
			if targetNode.WorldID != relation.WorldID {
				return errorf(ErrorCrossWorldRelation, "target node must be in the same world")
			}
			targetNodeID = targetNode.ID
			updates["target_id"] = targetNode.ID
		}
		if sourceNodeID == targetNodeID {
			return errorf(ErrorInvalidRelationType, "source node cannot point to itself")
		}
		if relationType != nil {
			if !engine.IsValidRelationType(*relationType) {
				return errorf(ErrorInvalidRelationType, "invalid relation_type: %s", *relationType)
			}
			if *relationType == string(engine.RelExternalParent) {
				var sourceNode store.NodeModel
				if err := tx.Where("id = ?", sourceNodeID).First(&sourceNode).Error; err != nil {
					return err
				}
				if sourceNode.ParentID != nil && *sourceNode.ParentID == targetNodeID {
					return conflictf("target node is already the primary parent")
				}
				if err := ensureNoExternalParentCycleTx(tx, sourceNodeID, targetNodeID); err != nil {
					return err
				}
			}
			updates["relation_type"] = *relationType
		}
		nextRelationType := relation.RelationType
		if relationType != nil {
			nextRelationType = *relationType
		}
		var duplicateCount int64
		if err := tx.Model(&store.RelationModel{}).
			Where("world_id = ? AND source_id = ? AND target_id = ? AND relation_type = ? AND id <> ?", relation.WorldID, sourceNodeID, targetNodeID, nextRelationType, relation.ID).
			Count(&duplicateCount).Error; err != nil {
			return err
		}
		if duplicateCount > 0 {
			return conflictf("relation already exists")
		}
		if weight != nil {
			updates["weight"] = int(*weight)
		}
		if props != nil {
			updates["properties"] = *props
		}
		if len(updates) == 0 {
			return errorf(ErrorNoUpdates, "no relation updates provided")
		}
		if err := tx.Model(&store.RelationModel{}).Where("id = ?", relation.ID).Updates(updates).Error; err != nil {
			return err
		}
		updated, err = getRelationTx(tx, id)
		return err
	})
	return updated, err
}

// DeleteRelation 删除一条关系记录。
func DeleteRelation(id string) error {
	return deleteByID(&store.RelationModel{}, id, getRelationTx)
}

// CreateMemory 为节点创建一条记忆记录。
func CreateMemory(nodeID, content, level, tags string) (*store.MemoryModel, error) {
	var created *store.MemoryModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		node, err := getNodeTx(tx, nodeID)
		if err != nil {
			return err
		}
		if level == "" {
			level = "long_term"
		}
		if !engine.IsValidMemoryLevel(level) {
			return errorf(ErrorInvalidMemoryLevel, "invalid memory level: %s", level)
		}
		created = &store.MemoryModel{
			UUID:    store.NewUUID(),
			NodeID:  node.ID,
			Content: content,
			Level:   level,
			Tags:    tags,
		}
		return tx.Create(created).Error
	})
	return created, err
}

// UpdateMemory 更新记忆记录。
func UpdateMemory(id string, content, level, tags *string) (*store.MemoryModel, error) {
	var updated *store.MemoryModel
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		memory, err := getMemoryTx(tx, id)
		if err != nil {
			return err
		}
		updates := map[string]any{}
		if content != nil {
			updates["content"] = *content
		}
		if level != nil {
			if !engine.IsValidMemoryLevel(*level) {
				return errorf(ErrorInvalidMemoryLevel, "invalid memory level: %s", *level)
			}
			updates["level"] = *level
		}
		if tags != nil {
			updates["tags"] = *tags
		}
		if len(updates) == 0 {
			return errorf(ErrorNoUpdates, "no memory updates provided")
		}
		if err := tx.Model(&store.MemoryModel{}).Where("id = ?", memory.ID).Updates(updates).Error; err != nil {
			return err
		}
		updated, err = getMemoryTx(tx, id)
		return err
	})
	return updated, err
}

// DeleteMemory 删除一条记忆。
func DeleteMemory(id string) error {
	return deleteByID(&store.MemoryModel{}, id, getMemoryTx)
}
