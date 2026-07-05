package service

import (
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
	"gorm.io/gorm"
)

// ImportResult 描述一次导入或纯校验的结果摘要。
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

// CloneWorld 复制一个世界及其所有节点、组件、记忆、关系、设置和策略。
// 新世界使用新的 UUID，原世界不受影响。
// lockWorld 为 true 时将在复制期间锁定源世界，阻止并发写入。
func CloneWorld(worldID, newName string, lockWorld bool) (*store.NodeModel, error) {
	if lockWorld {
		LockWorld(worldID)
		defer UnlockWorld(worldID)
	}
	world, err := getNodeTx(store.DB, worldID)
	if err != nil {
		return nil, err
	}
	if world.NodeType != "world" {
		return nil, errorf(ErrorInvalidNodeType, "can only clone a world node")
	}
	name := newName
	if name == "" {
		name = world.Name + " (副本)"
	}
	var created *store.NodeModel
	err = store.DB.Transaction(func(tx *gorm.DB) error {
		newWorld, err := createNodeTx(tx, "", name, "world", nil)
		if err != nil {
			return err
		}
		created = newWorld
		// 解析原世界的 int64 ID 用于查询
		oldWorldInt := txResolveWorldUUID(tx, worldID)
		var oldNodes []store.NodeModel
		if err := tx.Where("world_id = ? AND id != ?", oldWorldInt, oldWorldInt).Order("created_at ASC").Find(&oldNodes).Error; err != nil {
			return err
		}
		idMap := map[string]string{worldID: newWorld.UUID}
		for _, oldNode := range oldNodes {
			newUUID := store.NewUUID()
			idMap[oldNode.UUID] = newUUID
			var parentID *int64
			if oldNode.ParentID != nil {
				// 在 oldNodes 中按 int64 查找父节点的 UUID 以映射
				var parentUUID string
				if err := tx.Model(&store.NodeModel{}).Select("uuid").Where("id = ?", *oldNode.ParentID).First(&parentUUID).Error; err != nil {
					return err
				}
				if pid, ok := idMap[parentUUID]; ok {
					// 需要找到新节点的 int64 ID
					var newParentID int64
					if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", pid).First(&newParentID).Error; err != nil {
						return err
					}
					parentID = &newParentID
				}
			}
			newNode := store.NodeModel{
				UUID:      newUUID,
				WorldID:   created.ID,
				WorldUUID: newWorld.UUID,
				Name:      oldNode.Name,
				NodeType:  oldNode.NodeType,
				ParentID:  parentID,
			}
			if err := tx.Create(&newNode).Error; err != nil {
				return err
			}
		}
		var allComps []store.ComponentModel
		if err := tx.Where("node_id IN (SELECT id FROM nodes WHERE world_id = ?)", oldWorldInt).Find(&allComps).Error; err != nil {
			return err
		}
		for _, comp := range allComps {
			newNodeUUID, ok := idMap[comp.NodeUUID]
			if !ok {
				newNodeUUID = newWorld.UUID
			}
			var newNodeID int64
			if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", newNodeUUID).First(&newNodeID).Error; err != nil {
				return err
			}
			if err := tx.Create(&store.ComponentModel{
				UUID:          store.NewUUID(),
				NodeID:        newNodeID,
				NodeUUID:      newNodeUUID,
				ComponentType: comp.ComponentType,
				Data:          comp.Data,
			}).Error; err != nil {
				return err
			}
		}
		var allMems []store.MemoryModel
		if err := tx.Where("node_id IN (SELECT id FROM nodes WHERE world_id = ?)", oldWorldInt).Find(&allMems).Error; err != nil {
			return err
		}
		for _, mem := range allMems {
			newNodeUUID, ok := idMap[mem.NodeUUID]
			if !ok {
				newNodeUUID = newWorld.UUID
			}
			var newNodeID int64
			if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", newNodeUUID).First(&newNodeID).Error; err != nil {
				return err
			}
			if err := tx.Create(&store.MemoryModel{
				UUID:     store.NewUUID(),
				NodeID:   newNodeID,
				NodeUUID: newNodeUUID,
				Content:  mem.Content,
				Level:    mem.Level,
				Tags:     mem.Tags,
			}).Error; err != nil {
				return err
			}
		}
		var allRels []store.RelationModel
		if err := tx.Where("world_id = ?", oldWorldInt).Find(&allRels).Error; err != nil {
			return err
		}
		for _, rel := range allRels {
			newSrcUUID, ok1 := idMap[rel.SourceUUID]
			newTgtUUID, ok2 := idMap[rel.TargetUUID]
			if !ok1 || !ok2 {
				continue
			}
			var newSrcID, newTgtID int64
			if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", newSrcUUID).First(&newSrcID).Error; err != nil {
				return err
			}
			if err := tx.Model(&store.NodeModel{}).Select("id").Where("uuid = ?", newTgtUUID).First(&newTgtID).Error; err != nil {
				return err
			}
			if err := tx.Create(&store.RelationModel{
				UUID:         store.NewUUID(),
				WorldID:      created.ID,
				WorldUUID:    newWorld.UUID,
				SourceID:     newSrcID,
				SourceUUID:   newSrcUUID,
				TargetID:     newTgtID,
				TargetUUID:   newTgtUUID,
				RelationType: rel.RelationType,
				Weight:       rel.Weight,
				Properties:   rel.Properties,
			}).Error; err != nil {
				return err
			}
		}
		if settings, err := store.GetWorldSettings(worldID); err == nil && settings != nil {
			if err := tx.Create(&store.WorldSettingsModel{
				WorldUUID:                newWorld.UUID,
				MemoryLimit:              settings.MemoryLimit,
				MaxAnalysisRounds:        settings.MaxAnalysisRounds,
				MaxContextDepth:          settings.MaxContextDepth,
				AutoApply:                settings.AutoApply,
				RequireReviewAbove:       settings.RequireReviewAbove,
				PipelineMode:             settings.PipelineMode,
				PropagationMaxDepth:      settings.PropagationMaxDepth,
				SubTaskMaxRetries:        settings.SubTaskMaxRetries,
				SubTaskTimeoutSecs:       settings.SubTaskTimeoutSecs,
				EnablePropagationMachine: settings.EnablePropagationMachine,
			}).Error; err != nil {
				return err
			}
		}
		if policy, err := store.GetWorldPolicy(worldID); err == nil && policy != nil {
			if err := tx.Create(&store.WorldPolicyModel{
				WorldUUID:      newWorld.UUID,
				BlockedActions: policy.BlockedActions,
				SafeActions:    policy.SafeActions,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return created, err
}

// ImportWorld 从 sdk.ImportConfig 导入一个世界及其完整的节点/组件/记忆/关系结构。
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
	err := store.DB.Transaction(func(tx *gorm.DB) error {
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
				if err := tx.Create(&store.ComponentModel{
					UUID:          store.NewUUID(),
					NodeID:        node.ID,
					NodeUUID:      node.UUID,
					ComponentType: "profile",
					Data:          nodeCfg.Profile,
				}).Error; err != nil {
					return err
				}
			}
			if nodeCfg.Lore != "" {
				if err := tx.Create(&store.ComponentModel{
					UUID:          store.NewUUID(),
					NodeID:        node.ID,
					NodeUUID:      node.UUID,
					ComponentType: "lore",
					Data:          nodeCfg.Lore,
				}).Error; err != nil {
					return err
				}
			}
			for _, memory := range nodeCfg.Memories {
				level := memory.Level
				if level == "" {
					level = "long_term"
				}
				if err := tx.Create(&store.MemoryModel{
					UUID:     store.NewUUID(),
					NodeID:   node.ID,
					NodeUUID: node.UUID,
					Content:  memory.Content,
					Level:    level,
					Tags:     memory.Tags,
				}).Error; err != nil {
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
			if err := tx.Create(&store.ComponentModel{
				UUID:          store.NewUUID(),
				NodeID:        nodeIntID,
				NodeUUID:      nodeUUID,
				ComponentType: componentCfg.Type,
				Data:          componentCfg.Data,
			}).Error; err != nil {
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
			if err := tx.Create(&store.RelationModel{
				UUID:         store.NewUUID(),
				WorldID:      world.ID,
				WorldUUID:    world.UUID,
				SourceID:     srcIntID,
				SourceUUID:   sourceUUID,
				TargetID:     tgtIntID,
				TargetUUID:   targetUUID,
				RelationType: relationCfg.Type,
				Weight:       relationCfg.Weight,
				Properties:   relationCfg.Props,
			}).Error; err != nil {
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
