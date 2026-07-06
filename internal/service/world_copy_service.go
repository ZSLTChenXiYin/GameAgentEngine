package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/version"
	"gorm.io/gorm"
)

const worldCopyBatchSize = 200

type worldCopyMode struct {
	reason      string
	defaultName string
}

type worldCopyMappings struct {
	uuidByOldID map[int64]string
	idByOldID   map[int64]int64
}

// ForkWorld creates a working-copy fork of a world and all of its data.
func ForkWorld(worldID, newName string, lockWorld bool) (*store.NodeModel, error) {
	return duplicateWorld(worldID, newName, lockWorld, worldCopyMode{
		reason:      worldCopyReasonFork,
		defaultName: " (copy)",
	}, nil)
}

// CreateWorldSnapshot creates a save-oriented snapshot copy of a world.
func CreateWorldSnapshot(worldID, newName string, lockWorld bool) (*store.NodeModel, error) {
	return duplicateWorld(worldID, newName, lockWorld, worldCopyMode{
		reason:      worldCopyReasonSnapshot,
		defaultName: " snapshot",
	}, nil)
}

// RestoreWorld restores a saved snapshot into a new runnable world copy.
func RestoreWorld(snapshotWorldID, newName string, lockWorld bool) (*store.NodeModel, error) {
	return duplicateWorld(snapshotWorldID, newName, lockWorld, worldCopyMode{
		reason:      worldCopyReasonRestore,
		defaultName: " restored",
	}, func(tx *gorm.DB, sourceWorld *store.NodeModel) error {
		snapshotMeta, err := store.GetWorldSnapshotBySnapshotWorldTx(tx, sourceWorld.UUID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorf(ErrorNotFound, "snapshot metadata not found for world: %s", sourceWorld.UUID)
			}
			return err
		}
		validationResult, err := validateWorldSnapshotTx(tx, snapshotMeta)
		if err != nil {
			return err
		}
		if !validationResult.Valid {
			return snapshotValidationError(validationResult.Issues[0])
		}
		return nil
	})
}

func duplicateWorld(worldID, newName string, lockWorld bool, mode worldCopyMode, beforeCopy func(tx *gorm.DB, sourceWorld *store.NodeModel) error) (*store.NodeModel, error) {
	started := time.Now()
	if lockWorld {
		LockWorld(worldID)
		defer UnlockWorld(worldID)
	}

	var created *store.NodeModel
	err := store.DB.Transaction(func(tx *gorm.DB) error {
		sourceWorld, err := getNodeTx(tx, worldID)
		if err != nil {
			return err
		}
		if sourceWorld.NodeType != "world" {
			return errorf(ErrorInvalidNodeType, "can only copy a world node")
		}

		name := strings.TrimSpace(newName)
		if name == "" {
			name = sourceWorld.Name + mode.defaultName
		}

		if beforeCopy != nil {
			if err := beforeCopy(tx, sourceWorld); err != nil {
				return err
			}
		}

		newWorld, err := createNodeTx(tx, "", name, "world", nil)
		if err != nil {
			return err
		}
		created = newWorld

		oldWorldInt := sourceWorld.ID

		mappings, stats, err := copyWorldNodesTx(tx, newWorld, oldWorldInt)
		if err != nil {
			return err
		}
		if stats.ComponentCount, err = copyWorldComponentsTx(tx, oldWorldInt, mappings); err != nil {
			return err
		}
		if stats.MemoryCount, err = copyWorldMemoriesTx(tx, oldWorldInt, mappings); err != nil {
			return err
		}
		if stats.RelationCount, err = copyWorldRelationsTx(tx, created, oldWorldInt, mappings); err != nil {
			return err
		}

		if settings, err := store.GetWorldSettingsTx(tx, sourceWorld.UUID); err == nil && settings != nil {
			clonedSettings := &store.WorldSettingsModel{
				WorldID:   created.ID,
				WorldUUID: newWorld.UUID,
			}
			if err := tx.Create(clonedSettings).Error; err != nil {
				return err
			}
			if err := tx.Model(clonedSettings).Updates(map[string]any{
				"memory_limit":               settings.MemoryLimit,
				"max_analysis_rounds":        settings.MaxAnalysisRounds,
				"max_context_depth":          settings.MaxContextDepth,
				"auto_apply":                 settings.AutoApply,
				"require_review_above":       settings.RequireReviewAbove,
				"pipeline_mode":              settings.PipelineMode,
				"propagation_max_depth":      settings.PropagationMaxDepth,
				"sub_task_max_retries":       settings.SubTaskMaxRetries,
				"sub_task_timeout_secs":      settings.SubTaskTimeoutSecs,
				"enable_propagation_machine": settings.EnablePropagationMachine,
			}).Error; err != nil {
				return err
			}
		}

		if policy, err := store.GetWorldPolicyTx(tx, sourceWorld.UUID); err == nil && policy != nil {
			if err := tx.Create(&store.WorldPolicyModel{
				WorldID:        created.ID,
				WorldUUID:      newWorld.UUID,
				BlockedActions: policy.BlockedActions,
				SafeActions:    policy.SafeActions,
			}).Error; err != nil {
				return err
			}
		}

		compatibility, err := collectWorldSnapshotCompatibilityTx(tx, oldWorldInt, sourceWorld.UUID)
		if err != nil {
			return err
		}

		payloadHash := buildWorldSnapshotHash(sourceWorld.UUID, newWorld.UUID, stats)
		if err := tx.Create(&store.WorldSnapshotModel{
			UUID:                 store.NewUUID(),
			SourceWorldID:        oldWorldInt,
			SourceWorldUUID:      sourceWorld.UUID,
			SnapshotWorldID:      newWorld.ID,
			SnapshotWorldUUID:    newWorld.UUID,
			SnapshotName:         name,
			Reason:               mode.reason,
			EngineVersion:        version.Version,
			MinCompatibleVersion: version.MinCompatibleVersion,
			SchemaVersion:        worldSnapshotSchemaVersion,
			NodeCount:            stats.NodeCount,
			ComponentCount:       stats.ComponentCount,
			MemoryCount:          stats.MemoryCount,
			RelationCount:        stats.RelationCount,
			ComponentTypesJSON:   compatibility.ComponentTypesJSON,
			SettingsHash:         compatibility.SettingsHash,
			PolicyHash:           compatibility.PolicyHash,
			PayloadHash:          payloadHash,
		}).Error; err != nil {
			return err
		}

		return nil
	})
	if err == nil {
		log.Printf("[world-copy] reason=%s source=%s target=%s lock=%t duration_ms=%d", mode.reason, worldID, created.UUID, lockWorld, time.Since(started).Milliseconds())
	}
	return created, err
}

func copyWorldNodesTx(tx *gorm.DB, newWorld *store.NodeModel, oldWorldInt int64) (worldCopyMappings, cloneSnapshotStats, error) {
	var oldNodes []store.NodeModel
	if err := tx.Where("world_id = ?", oldWorldInt).Order("created_at ASC, id ASC").Find(&oldNodes).Error; err != nil {
		return worldCopyMappings{}, cloneSnapshotStats{}, err
	}

	mappings := worldCopyMappings{
		uuidByOldID: map[int64]string{oldWorldInt: newWorld.UUID},
		idByOldID:   map[int64]int64{oldWorldInt: newWorld.ID},
	}
	stats := cloneSnapshotStats{}
	childrenByParent := make(map[int64][]store.NodeModel)
	rootChildren := make([]store.NodeModel, 0)

	for _, oldNode := range oldNodes {
		if oldNode.ID == oldWorldInt {
			continue
		}
		if oldNode.ParentID == nil {
			rootChildren = append(rootChildren, oldNode)
			continue
		}
		childrenByParent[*oldNode.ParentID] = append(childrenByParent[*oldNode.ParentID], oldNode)
	}

	currentLevel := rootChildren
	for len(currentLevel) > 0 {
		nextLevelParents := make([]int64, 0, len(currentLevel))
		batchNodes := make([]store.NodeModel, 0, minInt(len(currentLevel), worldCopyBatchSize))
		batchOldIDs := make([]int64, 0, minInt(len(currentLevel), worldCopyBatchSize))
		flush := func() error {
			if len(batchNodes) == 0 {
				return nil
			}
			if err := tx.CreateInBatches(&batchNodes, worldCopyBatchSize).Error; err != nil {
				return err
			}
			for i := range batchNodes {
				mappings.uuidByOldID[batchOldIDs[i]] = batchNodes[i].UUID
				mappings.idByOldID[batchOldIDs[i]] = batchNodes[i].ID
				stats.NodeCount++
			}
			batchNodes = batchNodes[:0]
			batchOldIDs = batchOldIDs[:0]
			return nil
		}

		for _, oldNode := range currentLevel {
			var parentID *int64
			if oldNode.ParentID != nil {
				mappedParentID, ok := mappings.idByOldID[*oldNode.ParentID]
				if !ok {
					return worldCopyMappings{}, cloneSnapshotStats{}, fmt.Errorf("copy world: missing parent mapping for node %s", oldNode.UUID)
				}
				parentID = &mappedParentID
			}
			batchNodes = append(batchNodes, store.NodeModel{
				UUID:      store.NewUUID(),
				WorldID:   newWorld.ID,
				WorldUUID: newWorld.UUID,
				Name:      oldNode.Name,
				NodeType:  oldNode.NodeType,
				ParentID:  parentID,
			})
			batchOldIDs = append(batchOldIDs, oldNode.ID)
			nextLevelParents = append(nextLevelParents, oldNode.ID)
			if len(batchNodes) >= worldCopyBatchSize {
				if err := flush(); err != nil {
					return worldCopyMappings{}, cloneSnapshotStats{}, err
				}
			}
		}
		if err := flush(); err != nil {
			return worldCopyMappings{}, cloneSnapshotStats{}, err
		}

		nextLevel := make([]store.NodeModel, 0)
		for _, parentOldID := range nextLevelParents {
			nextLevel = append(nextLevel, childrenByParent[parentOldID]...)
		}
		currentLevel = nextLevel
	}

	return mappings, stats, nil
}

func copyWorldComponentsTx(tx *gorm.DB, oldWorldInt int64, mappings worldCopyMappings) (int, error) {
	var allComps []store.ComponentModel
	if err := tx.Where("node_id IN (SELECT id FROM nodes WHERE world_id = ?)", oldWorldInt).Find(&allComps).Error; err != nil {
		return 0, err
	}
	newComps := make([]store.ComponentModel, 0, len(allComps))
	for _, comp := range allComps {
		newNodeUUID, ok := mappings.uuidByOldID[comp.NodeID]
		if !ok {
			return 0, fmt.Errorf("copy world: missing component node uuid mapping for node_id=%d", comp.NodeID)
		}
		newNodeID, ok := mappings.idByOldID[comp.NodeID]
		if !ok {
			return 0, fmt.Errorf("copy world: missing component node id mapping for node_id=%d", comp.NodeID)
		}
		newComps = append(newComps, store.ComponentModel{
			UUID:          store.NewUUID(),
			NodeID:        newNodeID,
			NodeUUID:      newNodeUUID,
			ComponentType: comp.ComponentType,
			Data:          comp.Data,
		})
	}
	if len(newComps) == 0 {
		return 0, nil
	}
	if err := tx.CreateInBatches(&newComps, worldCopyBatchSize).Error; err != nil {
		return 0, err
	}
	return len(newComps), nil
}

func copyWorldMemoriesTx(tx *gorm.DB, oldWorldInt int64, mappings worldCopyMappings) (int, error) {
	var allMems []store.MemoryModel
	if err := tx.Where("node_id IN (SELECT id FROM nodes WHERE world_id = ?)", oldWorldInt).Find(&allMems).Error; err != nil {
		return 0, err
	}
	newMems := make([]store.MemoryModel, 0, len(allMems))
	for _, mem := range allMems {
		newNodeUUID, ok := mappings.uuidByOldID[mem.NodeID]
		if !ok {
			return 0, fmt.Errorf("copy world: missing memory node uuid mapping for node_id=%d", mem.NodeID)
		}
		newNodeID, ok := mappings.idByOldID[mem.NodeID]
		if !ok {
			return 0, fmt.Errorf("copy world: missing memory node id mapping for node_id=%d", mem.NodeID)
		}
		newMems = append(newMems, store.MemoryModel{
			UUID:     store.NewUUID(),
			NodeID:   newNodeID,
			NodeUUID: newNodeUUID,
			Content:  mem.Content,
			Level:    mem.Level,
			Tags:     mem.Tags,
		})
	}
	if len(newMems) == 0 {
		return 0, nil
	}
	if err := tx.CreateInBatches(&newMems, worldCopyBatchSize).Error; err != nil {
		return 0, err
	}
	return len(newMems), nil
}

func copyWorldRelationsTx(tx *gorm.DB, newWorld *store.NodeModel, oldWorldInt int64, mappings worldCopyMappings) (int, error) {
	var allRels []store.RelationModel
	if err := tx.Where("world_id = ?", oldWorldInt).Find(&allRels).Error; err != nil {
		return 0, err
	}
	newRels := make([]store.RelationModel, 0, len(allRels))
	for _, rel := range allRels {
		newSrcUUID, ok1 := mappings.uuidByOldID[rel.SourceID]
		newTgtUUID, ok2 := mappings.uuidByOldID[rel.TargetID]
		newSrcID, ok3 := mappings.idByOldID[rel.SourceID]
		newTgtID, ok4 := mappings.idByOldID[rel.TargetID]
		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}
		newRels = append(newRels, store.RelationModel{
			UUID:         store.NewUUID(),
			WorldID:      newWorld.ID,
			WorldUUID:    newWorld.UUID,
			SourceID:     newSrcID,
			SourceUUID:   newSrcUUID,
			TargetID:     newTgtID,
			TargetUUID:   newTgtUUID,
			RelationType: rel.RelationType,
			Weight:       rel.Weight,
			Properties:   rel.Properties,
		})
	}
	if len(newRels) == 0 {
		return 0, nil
	}
	if err := tx.CreateInBatches(&newRels, worldCopyBatchSize).Error; err != nil {
		return 0, err
	}
	return len(newRels), nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
