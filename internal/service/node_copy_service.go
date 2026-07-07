package service

import (
	"fmt"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

type CopyNodeOptions struct {
	Name               string
	ParentID           *string
	ParentIDSet        bool
	IncludeDescendants bool
}

// CopyNode duplicates a node inside the same world.
// By default it keeps the same parent and copies the full subtree.
func CopyNode(nodeID string, opts CopyNodeOptions) (*store.NodeModel, error) {
	if !opts.IncludeDescendants {
		// keep explicit false when requested; defaulting is handled by callers
	} else {
		opts.IncludeDescendants = true
	}

	var copied *store.NodeModel
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		source, err := getNodeTx(tx, nodeID)
		if err != nil {
			return err
		}
		if source.NodeType == "world" {
			return errorf(ErrorWorldNodeConstraint, "world node should be copied via world fork/snapshot flows")
		}

		name := opts.Name
		if name == "" {
			name = source.Name + " (copy)"
		}

		var targetParentID *int64
		switch {
		case opts.ParentIDSet && opts.ParentID != nil && *opts.ParentID != "":
			parent, err := getNodeTx(tx, *opts.ParentID)
			if err != nil {
				return err
			}
			if parent.WorldID != source.WorldID {
				return errorf(ErrorCrossWorldRelation, "parent node must be in the same world")
			}
			targetParentID = &parent.ID
		case opts.ParentIDSet:
			targetParentID = nil
		case source.ParentID != nil:
			parentID := *source.ParentID
			targetParentID = &parentID
		}

		copied = &store.NodeModel{
			UUID:      store.NewUUID(),
			WorldID:   source.WorldID,
			WorldUUID: source.WorldUUID,
			Name:      name,
			NodeType:  source.NodeType,
			ParentID:  targetParentID,
		}
		if err := tx.Create(copied).Error; err != nil {
			return err
		}

		if err := copyNodeComponentsTx(tx, source.ID, copied.ID); err != nil {
			return err
		}
		if err := copyNodeMemoriesTx(tx, source.ID, copied.ID); err != nil {
			return err
		}

		if !opts.IncludeDescendants {
			return copyNodeInternalRelationsTx(tx, source.WorldID, map[int64]int64{source.ID: copied.ID})
		}

		allNodes, err := listNodesByWorldIntTx(tx, source.WorldID)
		if err != nil {
			return err
		}
		childrenByParent := make(map[int64][]store.NodeModel)
		for _, node := range allNodes {
			if node.ParentID == nil {
				continue
			}
			childrenByParent[*node.ParentID] = append(childrenByParent[*node.ParentID], node)
		}

		idMap := map[int64]int64{source.ID: copied.ID}
		uuidMap := map[int64]string{source.ID: copied.UUID}
		if err := copyNodeDescendantsTx(tx, source.WorldID, source.ID, copied.ID, childrenByParent, idMap, uuidMap); err != nil {
			return err
		}

		return copyNodeInternalRelationsTx(tx, source.WorldID, idMap)
	})
	if err != nil {
		return nil, err
	}
	store.ResolveNodeParentUUID(copied)
	return copied, nil
}

func copyNodeDescendantsTx(
	tx *gorm.DB,
	worldID int64,
	sourceParentID int64,
	targetParentID int64,
	childrenByParent map[int64][]store.NodeModel,
	idMap map[int64]int64,
	uuidMap map[int64]string,
) error {
	children := childrenByParent[sourceParentID]
	for _, child := range children {
		newParentID := targetParentID
		copied := &store.NodeModel{
			UUID:     store.NewUUID(),
			WorldID:  worldID,
			Name:     child.Name,
			NodeType: child.NodeType,
			ParentID: &newParentID,
		}
		if child.WorldUUID != "" {
			copied.WorldUUID = child.WorldUUID
		}
		if err := tx.Create(copied).Error; err != nil {
			return err
		}
		idMap[child.ID] = copied.ID
		uuidMap[child.ID] = copied.UUID
		if err := copyNodeComponentsTx(tx, child.ID, copied.ID); err != nil {
			return err
		}
		if err := copyNodeMemoriesTx(tx, child.ID, copied.ID); err != nil {
			return err
		}
		if err := copyNodeDescendantsTx(tx, worldID, child.ID, copied.ID, childrenByParent, idMap, uuidMap); err != nil {
			return err
		}
	}
	return nil
}

func copyNodeComponentsTx(tx *gorm.DB, sourceNodeID int64, targetNodeID int64) error {
	var components []store.ComponentModel
	if err := tx.Where("node_id = ?", sourceNodeID).Order("id ASC").Find(&components).Error; err != nil {
		return err
	}
	for _, component := range components {
		copyComponent := &store.ComponentModel{
			UUID:          store.NewUUID(),
			NodeID:        targetNodeID,
			ComponentType: component.ComponentType,
			Data:          component.Data,
		}
		if err := tx.Create(copyComponent).Error; err != nil {
			return err
		}
	}
	return nil
}

func copyNodeMemoriesTx(tx *gorm.DB, sourceNodeID int64, targetNodeID int64) error {
	var memories []store.MemoryModel
	if err := tx.Where("node_id = ?", sourceNodeID).Order("id ASC").Find(&memories).Error; err != nil {
		return err
	}
	for _, memory := range memories {
		copyMemory := &store.MemoryModel{
			UUID:    store.NewUUID(),
			NodeID:  targetNodeID,
			Content: memory.Content,
			Level:   memory.Level,
			Tags:    memory.Tags,
		}
		if err := tx.Create(copyMemory).Error; err != nil {
			return err
		}
	}
	return nil
}

func copyNodeInternalRelationsTx(tx *gorm.DB, worldID int64, idMap map[int64]int64) error {
	oldIDs := make([]int64, 0, len(idMap))
	for oldID := range idMap {
		oldIDs = append(oldIDs, oldID)
	}
	if len(oldIDs) == 0 {
		return nil
	}

	var relations []store.RelationModel
	if err := tx.Where("world_id = ? AND source_id IN ? AND target_id IN ?", worldID, oldIDs, oldIDs).
		Order("id ASC").Find(&relations).Error; err != nil {
		return err
	}

	for _, relation := range relations {
		newSourceID, ok := idMap[relation.SourceID]
		if !ok {
			return fmt.Errorf("copy node: missing source mapping for relation %d", relation.ID)
		}
		newTargetID, ok := idMap[relation.TargetID]
		if !ok {
			return fmt.Errorf("copy node: missing target mapping for relation %d", relation.ID)
		}
		copyRelation := &store.RelationModel{
			UUID:         store.NewUUID(),
			WorldID:      worldID,
			SourceID:     newSourceID,
			TargetID:     newTargetID,
			RelationType: relation.RelationType,
			Weight:       relation.Weight,
			Properties:   relation.Properties,
		}
		if err := tx.Create(copyRelation).Error; err != nil {
			return err
		}
	}
	return nil
}

func listNodesByWorldIntTx(tx *gorm.DB, worldID int64) ([]store.NodeModel, error) {
	var nodes []store.NodeModel
	if err := tx.Where("world_id = ?", worldID).Order("created_at ASC, id ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

// UpdateWorld updates mutable fields on a world root node.
func UpdateWorld(worldID string, name *string) (*store.NodeModel, error) {
	var updated *store.NodeModel
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		world, err := ensureWorldNodeTx(tx, worldID)
		if err != nil {
			return err
		}
		if name == nil {
			return errorf(ErrorNoUpdates, "no world updates provided")
		}
		updates := map[string]any{"name": *name}
		if err := tx.Model(&store.NodeModel{}).Where("id = ?", world.ID).Updates(updates).Error; err != nil {
			return err
		}
		updated, err = getNodeTx(tx, worldID)
		return err
	})
	if err != nil {
		return nil, err
	}
	store.ResolveNodeParentUUID(updated)
	return updated, nil
}
