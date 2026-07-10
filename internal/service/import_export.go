package service

import (
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
	"gorm.io/gorm"
)

// ImportResult describes an import or validation summary.
type ImportResult struct {
	WorldID        string `json:"world_id,omitempty"`
	WorldName      string `json:"world_name"`
	DryRun         bool   `json:"dry_run"`
	NodeCount      int    `json:"node_count"`
	ComponentCount int    `json:"component_count"`
	MemoryCount    int    `json:"memory_count"`
	RelationCount  int    `json:"relation_count"`
}

func resetAllTx(tx *gorm.DB) error {
	if err := tx.Exec("DELETE FROM sqlite_sequence").Error; err != nil {
		// ignore for non-SQLite databases
	}
	models := []any{
		&store.InferenceLogModel{},
		&store.TimelineModel{},
		&store.RelationModel{},
		&store.ComponentModel{},
		&store.MemoryModel{},
		&store.NodeModel{},
	}
	for _, model := range models {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Unscoped().Delete(model).Error; err != nil {
			return err
		}
	}
	return nil
}

func validateImportConfig(cfg *sdk.ImportConfig) error {
	seenNames := map[string]struct{}{}
	for _, nodeCfg := range cfg.Nodes {
		if nodeCfg.Name == "" {
			return invalidf("import node name is required")
		}
		if _, exists := seenNames[nodeCfg.Name]; exists {
			return errorf(ErrorImportDuplicateName, "duplicate node name in import payload: %s", nodeCfg.Name)
		}
		seenNames[nodeCfg.Name] = struct{}{}
		if !engine.IsValidNodeType(nodeCfg.Type) {
			return errorf(ErrorInvalidNodeType, "invalid node type in import payload: %s", nodeCfg.Type)
		}
		for _, memory := range nodeCfg.Memories {
			level := memory.Level
			if level == "" {
				level = "long_term"
			}
			if !engine.IsValidMemoryLevel(level) {
				return errorf(ErrorInvalidMemoryLevel, "invalid memory level in import payload: %s", level)
			}
		}
	}
	for _, componentCfg := range cfg.Components {
		if componentCfg.NodeID == "" {
			return invalidf("component node_id is required")
		}
		if !engine.IsValidComponentType(componentCfg.Type) {
			return errorf(ErrorInvalidComponentType, "invalid component type in import payload: %s", componentCfg.Type)
		}
	}
	for _, relationCfg := range cfg.Relations {
		if relationCfg.Source == "" || relationCfg.Target == "" {
			return invalidf("relation source and target are required")
		}
		if !engine.IsValidRelationType(relationCfg.Type) {
			return errorf(ErrorInvalidRelationType, "invalid relation type in import payload: %s", relationCfg.Type)
		}
	}
	return nil
}

// ImportWorld imports a complete world graph from sdk.ImportConfig.
func ImportWorld(cfg *sdk.ImportConfig, reset, dryRun bool) (*ImportResult, error) {
	if err := validateImportConfig(cfg); err != nil {
		return nil, err
	}
	result := &ImportResult{
		WorldName:      cfg.World.Name,
		DryRun:         dryRun,
		NodeCount:      len(cfg.Nodes),
		ComponentCount: len(cfg.Components),
		RelationCount:  len(cfg.Relations),
	}
	for _, nodeCfg := range cfg.Nodes {
		result.MemoryCount += len(nodeCfg.Memories)
		if nodeCfg.Profile != "" {
			result.ComponentCount++
		}
		if nodeCfg.Lore != "" {
			result.ComponentCount++
		}
	}
	if dryRun {
		return result, nil
	}
	err := store.WriteTransaction(func(tx *gorm.DB) error {
		if reset {
			if err := resetAllTx(tx); err != nil {
				return err
			}
		}
		var world *store.NodeModel
		var err error
		if !reset {
			world, err = getWorldByNameTx(tx, cfg.World.Name)
			if err != nil && !IsKind(err, ErrorNotFound) && !IsKind(err, ErrorWorldNotFound) {
				return err
			}
		}
		if world == nil {
			world, err = createNodeTx(tx, "", cfg.World.Name, "world", nil)
			if err != nil {
				return err
			}
		}
		result.WorldID = world.UUID
		nodeMap := map[string]string{"world": world.UUID, cfg.World.Name: world.UUID}
		for _, nodeCfg := range cfg.Nodes {
			var parentID *string
			if nodeCfg.Parent != "" {
				resolved, ok := nodeMap[nodeCfg.Parent]
				if !ok {
					return errorf(ErrorParentNotFound, "parent node not found in import payload: %s", nodeCfg.Parent)
				}
				parentID = &resolved
			}
			node, err := createNodeTx(tx, world.UUID, nodeCfg.Name, nodeCfg.Type, parentID)
			if err != nil {
				return err
			}
			nodeMap[nodeCfg.Name] = node.UUID
			if nodeCfg.Profile != "" {
				if err := tx.Create(&store.ComponentModel{UUID: store.NewUUID(), NodeID: node.ID, NodeUUID: node.UUID, ComponentType: "profile", Data: nodeCfg.Profile}).Error; err != nil {
					return err
				}
			}
			if nodeCfg.Lore != "" {
				if err := tx.Create(&store.ComponentModel{UUID: store.NewUUID(), NodeID: node.ID, NodeUUID: node.UUID, ComponentType: "lore", Data: nodeCfg.Lore}).Error; err != nil {
					return err
				}
			}
			for _, memory := range nodeCfg.Memories {
				level := memory.Level
				if level == "" {
					level = "long_term"
				}
				if err := tx.Create(&store.MemoryModel{UUID: store.NewUUID(), NodeID: node.ID, NodeUUID: node.UUID, Content: memory.Content, Level: level, Tags: memory.Tags}).Error; err != nil {
					return err
				}
			}
		}
		for _, componentCfg := range cfg.Components {
			nodeUUID, ok := nodeMap[componentCfg.NodeID]
			if !ok {
				nodeUUID = componentCfg.NodeID
			}
			if _, err := getNodeTx(tx, nodeUUID); err != nil {
				return errorf(ErrorParentNotFound, "component target node not found: %s", componentCfg.NodeID)
			}
			var nodeIntID int64
			if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", nodeUUID).First(&nodeIntID).Error; err != nil {
				return err
			}
			if err := tx.Create(&store.ComponentModel{UUID: store.NewUUID(), NodeID: nodeIntID, NodeUUID: nodeUUID, ComponentType: componentCfg.Type, Data: componentCfg.Data}).Error; err != nil {
				return err
			}
		}
		for _, relationCfg := range cfg.Relations {
			sourceUUID, ok1 := nodeMap[relationCfg.Source]
			targetUUID, ok2 := nodeMap[relationCfg.Target]
			if !ok1 || !ok2 {
				return errorf(ErrorParentNotFound, "relation endpoint not found: %s -> %s", relationCfg.Source, relationCfg.Target)
			}
			if err := ensureNodesInWorldTx(tx, world.UUID, sourceUUID, targetUUID); err != nil {
				return err
			}
			var srcIntID, tgtIntID int64
			if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", sourceUUID).First(&srcIntID).Error; err != nil {
				return err
			}
			if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", targetUUID).First(&tgtIntID).Error; err != nil {
				return err
			}
			if err := tx.Create(&store.RelationModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, SourceID: srcIntID, SourceUUID: sourceUUID, TargetID: tgtIntID, TargetUUID: targetUUID, RelationType: relationCfg.Type, Weight: relationCfg.Weight, Properties: relationCfg.Props}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
