package service

import (
	"errors"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

// ---- tx helpers ----
// createNodeTx 在事务内创建节点。worldID / parentID 均为 UUID 字符串。
// 对 world 类型节点，自动将自身 ID 作为 WorldID。
func createNodeTx(tx *gorm.DB, worldID, name, nodeType string, parentID *string) (*store.NodeModel, error) {
	if name == "" || nodeType == "" {
		return nil, invalidf("name and node_type required")
	}
	if !engine.IsValidNodeType(nodeType) {
		return nil, errorf(ErrorInvalidNodeType, "invalid node_type: %s", nodeType)
	}
	if nodeType == "world" {
		if parentID != nil {
			return nil, errorf(ErrorWorldNodeConstraint, "world node cannot have a parent")
		}
		uuid := store.NewUUID()
		node := &store.NodeModel{UUID: uuid, Name: name, NodeType: nodeType}
		if err := tx.Create(node).Error; err != nil {
			return nil, err
		}
		// world 节点的 WorldID 与 id 相同
		if err := tx.Model(node).Update("world_id", node.ID).Error; err != nil {
			return nil, err
		}
		node.WorldUUID = uuid
		return node, nil
	}
	// worldInt is used below for WorldID
	worldInt := txResolveWorldUUID(tx, worldID)
	if _, err := ensureWorldNodeTx(tx, worldID); err != nil {
		return nil, err
	}
	var resolvedParentID *int64
	if parentID != nil && *parentID != "" {
		parent, err := getNodeTx(tx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent.WorldID != txResolveWorldUUID(tx, worldID) {
			return nil, errorf(ErrorCrossWorldRelation, "parent node must be in the same world")
		}
		resolvedParentID = &parent.ID
	}
	uuid := store.NewUUID()
	node := &store.NodeModel{
		UUID:      uuid,
		WorldID:   worldInt,
		WorldUUID: worldID,
		Name:      name,
		NodeType:  nodeType,
		ParentID:  resolvedParentID,
	}
	if err := tx.Create(node).Error; err != nil {
		return nil, err
	}
	return node, nil
}

func ensureWorldNodeTx(tx *gorm.DB, worldID string) (*store.NodeModel, error) {
	if worldID == "" {
		return nil, invalidf("world_id required")
	}
	world, err := getNodeTx(tx, worldID)
	if err != nil {
		return nil, err
	}
	if world.NodeType != "world" {
		return nil, invalidf("world_id must reference a world node")
	}
	return world, nil
}

func getWorldByNameTx(tx *gorm.DB, name string) (*store.NodeModel, error) {
	var world store.NodeModel
	if err := tx.Where("node_type = ? AND name = ?", "world", name).First(&world).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorf(ErrorWorldNotFound, "world not found: %s", name)
		}
		return nil, err
	}
	return &world, nil
}

func ensureNoParentCycleTx(tx *gorm.DB, nodeID, parentID int64) error {
	if nodeID == parentID {
		return errorf(ErrorParentCycle, "node cannot be parent of itself")
	}
	currentID := parentID
	for currentID != 0 {
		if currentID == nodeID {
			return errorf(ErrorParentCycle, "parent update would create a cycle")
		}
		var current store.NodeModel
		if err := tx.First(&current, "id = ?", currentID).Error; err != nil {
			return err
		}
		if current.ParentID == nil {
			break
		}
		currentID = *current.ParentID
	}
	return nil
}

func ensureNodesInWorldTx(tx *gorm.DB, worldID string, nodeIDs ...string) error {
	worldInt := txResolveWorldUUID(tx, worldID)
	for _, nodeID := range nodeIDs {
		node, err := getNodeTx(tx, nodeID)
		if err != nil {
			return err
		}
		if node.WorldID != int64(worldInt) {
			return errorf(ErrorCrossWorldRelation, "node %s is not in world %s", nodeID, worldID)
		}
	}
	return nil
}

func getNodeTx(tx *gorm.DB, id string) (*store.NodeModel, error) {
	var node store.NodeModel
	if err := tx.Where("uuid = ?", id).First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errorf(ErrorNodeNotFound, "node not found: %s", id)
		}
		return nil, err
	}
	return &node, nil
}
