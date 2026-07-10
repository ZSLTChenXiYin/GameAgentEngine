package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/version"
	"gorm.io/gorm"
)

const worldSnapshotSchemaVersion = "world_snapshot/v1"

const (
	worldCopyReasonFork     = "fork_world"
	worldCopyReasonSnapshot = "save_snapshot"
	worldCopyReasonRestore  = "restore_snapshot"
)

type cloneSnapshotStats struct {
	NodeCount      int
	ComponentCount int
	MemoryCount    int
	RelationCount  int
}

type worldSnapshotCompatibility struct {
	ComponentTypesJSON string
	SettingsHash       string
	PolicyHash         string
}

const (
	snapshotValidationCodeReasonInvalid         = "snapshot_reason_invalid"
	snapshotValidationCodeSchemaIncompatible    = "snapshot_schema_incompatible"
	snapshotValidationCodeEngineIncompatible    = "snapshot_engine_incompatible"
	snapshotValidationCodeRuntimeIncompatible   = "snapshot_runtime_incompatible"
	snapshotValidationCodeSnapshotWorldMissing  = "snapshot_world_missing"
	snapshotValidationCodeComponentTypesDrifted = "snapshot_component_types_drifted"
	snapshotValidationCodeSettingsDrifted       = "snapshot_settings_drifted"
	snapshotValidationCodePolicyDrifted         = "snapshot_policy_drifted"
)

type SnapshotValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SnapshotValidationResult struct {
	SnapshotWorldID             string                    `json:"snapshot_world_id"`
	SourceWorldID               string                    `json:"source_world_id"`
	SnapshotName                string                    `json:"snapshot_name"`
	Reason                      string                    `json:"reason"`
	Valid                       bool                      `json:"valid"`
	SchemaVersion               string                    `json:"schema_version"`
	EngineVersion               string                    `json:"engine_version"`
	MinCompatibleVersion        string                    `json:"min_compatible_version"`
	CurrentEngineVersion        string                    `json:"current_engine_version"`
	CurrentMinCompatibleVersion string                    `json:"current_min_compatible_version"`
	NodeCount                   int                       `json:"node_count"`
	ComponentCount              int                       `json:"component_count"`
	MemoryCount                 int                       `json:"memory_count"`
	RelationCount               int                       `json:"relation_count"`
	SavedComponentTypes         []string                  `json:"saved_component_types"`
	CurrentComponentTypes       []string                  `json:"current_component_types"`
	SavedSettingsHash           string                    `json:"saved_settings_hash"`
	CurrentSettingsHash         string                    `json:"current_settings_hash"`
	SavedPolicyHash             string                    `json:"saved_policy_hash"`
	CurrentPolicyHash           string                    `json:"current_policy_hash"`
	Issues                      []SnapshotValidationIssue `json:"issues,omitempty"`
}

type WorldSnapshotInfo struct {
	ID                   string    `json:"id"`
	SourceWorldID        string    `json:"source_world_id"`
	SnapshotWorldID      string    `json:"snapshot_world_id"`
	SnapshotName         string    `json:"snapshot_name"`
	Reason               string    `json:"reason"`
	Status               string    `json:"status"`
	Restorable           bool      `json:"restorable"`
	EngineVersion        string    `json:"engine_version"`
	MinCompatibleVersion string    `json:"min_compatible_version"`
	SchemaVersion        string    `json:"schema_version"`
	NodeCount            int       `json:"node_count"`
	ComponentCount       int       `json:"component_count"`
	MemoryCount          int       `json:"memory_count"`
	RelationCount        int       `json:"relation_count"`
	ComponentTypes       []string  `json:"component_types"`
	SettingsHash         string    `json:"settings_hash"`
	PolicyHash           string    `json:"policy_hash"`
	PayloadHash          string    `json:"payload_hash"`
	ValidationIssues     []string  `json:"validation_issues,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func buildWorldSnapshotHash(sourceWorldID, snapshotWorldID string, stats cloneSnapshotStats) string {
	payload := []string{
		sourceWorldID,
		snapshotWorldID,
		fmt.Sprintf("nodes:%d", stats.NodeCount),
		fmt.Sprintf("components:%d", stats.ComponentCount),
		fmt.Sprintf("memories:%d", stats.MemoryCount),
		fmt.Sprintf("relations:%d", stats.RelationCount),
	}
	sort.Strings(payload)
	sum := sha256.Sum256([]byte(strings.Join(payload, "|")))
	return hex.EncodeToString(sum[:])
}

func buildStableHash(parts ...string) string {
	joined := make([]string, len(parts))
	copy(joined, parts)
	sort.Strings(joined)
	sum := sha256.Sum256([]byte(strings.Join(joined, "|")))
	return hex.EncodeToString(sum[:])
}

func normalizeJSONList(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "[]"
	}
	return trimmed
}

func decodeComponentTypes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	return values
}

func addSnapshotValidationIssue(result *SnapshotValidationResult, code, message string) {
	result.Issues = append(result.Issues, SnapshotValidationIssue{
		Code:    code,
		Message: message,
	})
}

func buildWorldSnapshotInfo(snapshotMeta *store.WorldSnapshotModel) *WorldSnapshotInfo {
	if snapshotMeta == nil {
		return nil
	}
	return &WorldSnapshotInfo{
		ID:                   snapshotMeta.UUID,
		SourceWorldID:        snapshotMeta.SourceWorldUUID,
		SnapshotWorldID:      snapshotMeta.SnapshotWorldUUID,
		SnapshotName:         snapshotMeta.SnapshotName,
		Reason:               snapshotMeta.Reason,
		Status:               "unknown",
		Restorable:           false,
		EngineVersion:        snapshotMeta.EngineVersion,
		MinCompatibleVersion: snapshotMeta.MinCompatibleVersion,
		SchemaVersion:        snapshotMeta.SchemaVersion,
		NodeCount:            snapshotMeta.NodeCount,
		ComponentCount:       snapshotMeta.ComponentCount,
		MemoryCount:          snapshotMeta.MemoryCount,
		RelationCount:        snapshotMeta.RelationCount,
		ComponentTypes:       decodeComponentTypes(snapshotMeta.ComponentTypesJSON),
		SettingsHash:         snapshotMeta.SettingsHash,
		PolicyHash:           snapshotMeta.PolicyHash,
		PayloadHash:          snapshotMeta.PayloadHash,
		CreatedAt:            snapshotMeta.CreatedAt,
		UpdatedAt:            snapshotMeta.UpdatedAt,
	}
}

func summarizeValidationIssues(issues []SnapshotValidationIssue) []string {
	if len(issues) == 0 {
		return nil
	}
	result := make([]string, 0, len(issues))
	for _, issue := range issues {
		result = append(result, issue.Code)
	}
	return result
}

func decorateWorldSnapshotInfo(info *WorldSnapshotInfo, validation *SnapshotValidationResult) *WorldSnapshotInfo {
	if info == nil {
		return nil
	}
	if validation == nil {
		switch info.Reason {
		case worldCopyReasonFork:
			info.Status = "working_copy"
		case worldCopyReasonRestore:
			info.Status = "restored_copy"
		case worldCopyReasonSnapshot:
			info.Status = "unknown"
		}
		return info
	}
	if validation.Valid {
		info.Status = "valid"
		info.Restorable = true
		return info
	}
	info.ValidationIssues = summarizeValidationIssues(validation.Issues)
	if validation.Reason != worldCopyReasonSnapshot {
		switch validation.Reason {
		case worldCopyReasonFork:
			info.Status = "working_copy"
		case worldCopyReasonRestore:
			info.Status = "restored_copy"
		default:
			info.Status = "invalid"
		}
		return info
	}
	info.Status = "invalid"
	return info
}

func collectWorldSnapshotCompatibilityTx(tx *gorm.DB, worldIntID int64, worldUUID string) (worldSnapshotCompatibility, error) {
	var componentTypes []string
	if err := tx.Model(&store.ComponentModel{}).
		Distinct("components.component_type").
		Joins("JOIN nodes ON nodes.id = components.node_id").
		Where("nodes.world_id = ? AND nodes.deleted_at IS NULL", worldIntID).
		Order("components.component_type ASC").
		Pluck("components.component_type", &componentTypes).Error; err != nil {
		return worldSnapshotCompatibility{}, err
	}
	componentTypesJSONBytes, err := json.Marshal(componentTypes)
	if err != nil {
		return worldSnapshotCompatibility{}, err
	}
	compatibility := worldSnapshotCompatibility{
		ComponentTypesJSON: string(componentTypesJSONBytes),
		SettingsHash:       buildStableHash(),
		PolicyHash:         buildStableHash(),
	}

	if settings, err := store.GetWorldSettingsTx(tx, worldUUID); err == nil && settings != nil {
		compatibility.SettingsHash = buildStableHash(
			fmt.Sprintf("memory_limit:%d", settings.MemoryLimit),
			fmt.Sprintf("max_analysis_rounds:%d", settings.MaxAnalysisRounds),
			fmt.Sprintf("max_context_depth:%d", settings.MaxContextDepth),
			fmt.Sprintf("auto_apply:%t", settings.AutoApply),
			fmt.Sprintf("require_review_above:%s", settings.RequireReviewAbove),
			fmt.Sprintf("pipeline_mode:%s", settings.PipelineMode),
			fmt.Sprintf("propagation_max_depth:%d", settings.PropagationMaxDepth),
			fmt.Sprintf("sub_task_max_retries:%d", settings.SubTaskMaxRetries),
			fmt.Sprintf("sub_task_timeout_secs:%d", settings.SubTaskTimeoutSecs),
			fmt.Sprintf("enable_propagation_machine:%t", settings.EnablePropagationMachine),
		)
	}
	if policy, err := store.GetWorldPolicyTx(tx, worldUUID); err == nil && policy != nil {
		compatibility.PolicyHash = buildStableHash(
			"blocked:"+normalizeJSONList(policy.BlockedActions),
			"safe:"+normalizeJSONList(policy.SafeActions),
		)
	}

	return compatibility, nil
}

func buildSnapshotValidationResultTx(tx *gorm.DB, snapshotMeta *store.WorldSnapshotModel) (*SnapshotValidationResult, error) {
	result := &SnapshotValidationResult{
		SnapshotWorldID:             snapshotMeta.SnapshotWorldUUID,
		SourceWorldID:               snapshotMeta.SourceWorldUUID,
		SnapshotName:                snapshotMeta.SnapshotName,
		Reason:                      snapshotMeta.Reason,
		Valid:                       true,
		SchemaVersion:               snapshotMeta.SchemaVersion,
		EngineVersion:               snapshotMeta.EngineVersion,
		MinCompatibleVersion:        snapshotMeta.MinCompatibleVersion,
		CurrentEngineVersion:        version.Version,
		CurrentMinCompatibleVersion: version.MinCompatibleVersion,
		NodeCount:                   snapshotMeta.NodeCount,
		ComponentCount:              snapshotMeta.ComponentCount,
		MemoryCount:                 snapshotMeta.MemoryCount,
		RelationCount:               snapshotMeta.RelationCount,
		SavedComponentTypes:         decodeComponentTypes(snapshotMeta.ComponentTypesJSON),
		SavedSettingsHash:           snapshotMeta.SettingsHash,
		SavedPolicyHash:             snapshotMeta.PolicyHash,
	}

	if snapshotMeta.Reason != worldCopyReasonSnapshot {
		addSnapshotValidationIssue(result, snapshotValidationCodeReasonInvalid, fmt.Sprintf("world %s is not a save snapshot", snapshotMeta.SnapshotWorldUUID))
	}
	if snapshotMeta.SchemaVersion != worldSnapshotSchemaVersion {
		addSnapshotValidationIssue(result, snapshotValidationCodeSchemaIncompatible, fmt.Sprintf("snapshot schema version incompatible: %s", snapshotMeta.SchemaVersion))
	}
	if compatible, message := version.Check(version.MinCompatibleVersion, snapshotMeta.EngineVersion); !compatible {
		addSnapshotValidationIssue(result, snapshotValidationCodeEngineIncompatible, fmt.Sprintf("snapshot engine version incompatible: %s", message))
	}
	if compatible, message := version.Check(snapshotMeta.MinCompatibleVersion, version.Version); !compatible {
		addSnapshotValidationIssue(result, snapshotValidationCodeRuntimeIncompatible, fmt.Sprintf("current engine version incompatible with snapshot requirements: %s", message))
	}

	worldIntID := txResolveWorldUUID(tx, snapshotMeta.SnapshotWorldUUID)
	if worldIntID == 0 {
		addSnapshotValidationIssue(result, snapshotValidationCodeSnapshotWorldMissing, fmt.Sprintf("snapshot world not found: %s", snapshotMeta.SnapshotWorldUUID))
	} else {
		compatibility, err := collectWorldSnapshotCompatibilityTx(tx, worldIntID, snapshotMeta.SnapshotWorldUUID)
		if err != nil {
			return nil, err
		}
		result.CurrentComponentTypes = decodeComponentTypes(compatibility.ComponentTypesJSON)
		result.CurrentSettingsHash = compatibility.SettingsHash
		result.CurrentPolicyHash = compatibility.PolicyHash

		if compatibility.ComponentTypesJSON != snapshotMeta.ComponentTypesJSON {
			addSnapshotValidationIssue(result, snapshotValidationCodeComponentTypesDrifted, "snapshot component types drifted after save")
		}
		if compatibility.SettingsHash != snapshotMeta.SettingsHash {
			addSnapshotValidationIssue(result, snapshotValidationCodeSettingsDrifted, "snapshot world settings drifted after save")
		}
		if compatibility.PolicyHash != snapshotMeta.PolicyHash {
			addSnapshotValidationIssue(result, snapshotValidationCodePolicyDrifted, "snapshot world policy drifted after save")
		}
	}

	result.Valid = len(result.Issues) == 0
	return result, nil
}

func validateSnapshotCompatibility(tx *gorm.DB, snapshotMeta *store.WorldSnapshotModel) error {
	result, err := buildSnapshotValidationResultTx(tx, snapshotMeta)
	if err != nil {
		return err
	}
	if result.Valid {
		return nil
	}
	firstIssue := result.Issues[0]
	if firstIssue.Code == snapshotValidationCodeReasonInvalid {
		return errorf(ErrorInvalid, "%s", firstIssue.Message)
	}
	return conflictf("%s", firstIssue.Message)
}

func snapshotValidationError(issue SnapshotValidationIssue) error {
	if issue.Code == snapshotValidationCodeReasonInvalid {
		return codedErrorf(ErrorInvalid, issue.Code, "%s", issue.Message)
	}
	return codedErrorf(ErrorConflict, issue.Code, "%s", issue.Message)
}

func validateWorldSnapshotTx(tx *gorm.DB, snapshotMeta *store.WorldSnapshotModel) (*SnapshotValidationResult, error) {
	if snapshotMeta == nil {
		return nil, errorf(ErrorNotFound, "snapshot metadata not found")
	}
	return buildSnapshotValidationResultTx(tx, snapshotMeta)
}

func getWorldSnapshotInfoTx(tx *gorm.DB, snapshotMeta *store.WorldSnapshotModel) (*WorldSnapshotInfo, error) {
	info := buildWorldSnapshotInfo(snapshotMeta)
	if info == nil {
		return nil, nil
	}
	validation, err := validateWorldSnapshotTx(tx, snapshotMeta)
	if err != nil {
		return decorateWorldSnapshotInfo(info, nil), err
	}
	return decorateWorldSnapshotInfo(info, validation), nil
}

func decorateWorldSnapshotListTx(tx *gorm.DB, list []store.WorldSnapshotModel) ([]WorldSnapshotInfo, error) {
	result := make([]WorldSnapshotInfo, 0, len(list))
	for i := range list {
		info, err := getWorldSnapshotInfoTx(tx, &list[i])
		if info != nil {
			result = append(result, *info)
		}
		if err != nil {
			log.Printf("[snapshot-list] snapshot=%s validation-decorate-fallback err=%v", list[i].SnapshotWorldUUID, err)
		}
	}
	return result, nil
}

func deleteWorldGraphTx(tx *gorm.DB, worldUUID string) error {
	worldIntID := txResolveWorldUUID(tx, worldUUID)
	if worldIntID == 0 {
		return errorf(ErrorWorldNotFound, "world not found: %s", worldUUID)
	}
	if err := tx.Where("world_id = ?", worldIntID).Delete(&store.InferenceLogModel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("world_id = ?", worldIntID).Delete(&store.TimelineModel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("world_uuid = ?", worldUUID).Delete(&store.WorldSettingsModel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("world_uuid = ?", worldUUID).Delete(&store.WorldPolicyModel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("world_id = ?", worldIntID).Delete(&store.RelationModel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("node_id IN (SELECT id FROM nodes WHERE world_id = ?)", worldIntID).Delete(&store.ComponentModel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("node_id IN (SELECT id FROM nodes WHERE world_id = ?)", worldIntID).Delete(&store.MemoryModel{}).Error; err != nil {
		return err
	}
	if err := tx.Unscoped().Where("snapshot_world_uuid = ?", worldUUID).Delete(&store.WorldSnapshotModel{}).Error; err != nil {
		return err
	}
	if err := tx.Where("world_id = ?", worldIntID).Delete(&store.NodeModel{}).Error; err != nil {
		return err
	}
	return nil
}

// ValidateWorldSnapshot validates whether a saved snapshot can be safely restored.
func ValidateWorldSnapshot(snapshotWorldID string) (*SnapshotValidationResult, error) {
	var result *SnapshotValidationResult
	if err := store.WriteTransaction(func(tx *gorm.DB) error {
		snapshotMeta, err := store.GetWorldSnapshotBySnapshotWorldTx(tx, snapshotWorldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorf(ErrorNotFound, "snapshot metadata not found for world: %s", snapshotWorldID)
			}
			return err
		}
		validationResult, err := validateWorldSnapshotTx(tx, snapshotMeta)
		if err != nil {
			return err
		}
		result = validationResult
		return nil
	}); err != nil {
		return nil, err
	}
	log.Printf("[snapshot-validation] snapshot=%s source=%s valid=%t issues=%v", snapshotWorldID, result.SourceWorldID, result.Valid, summarizeValidationIssues(result.Issues))
	return result, nil
}

// GetWorldSnapshotMetadata returns snapshot metadata for a copied world.
func GetWorldSnapshotMetadata(snapshotWorldID string) (*WorldSnapshotInfo, error) {
	var info *WorldSnapshotInfo
	if err := store.WriteTransaction(func(tx *gorm.DB) error {
		snapshotMeta, err := store.GetWorldSnapshotBySnapshotWorldTx(tx, snapshotWorldID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errorf(ErrorNotFound, "snapshot metadata not found for world: %s", snapshotWorldID)
			}
			return err
		}
		decoratedInfo, txErr := getWorldSnapshotInfoTx(tx, snapshotMeta)
		info = decoratedInfo
		return txErr
	}); err != nil {
		if info != nil {
			log.Printf("[snapshot-metadata] snapshot=%s validation-decorate-fallback err=%v", snapshotWorldID, err)
			return info, nil
		}
		return nil, err
	}
	return info, nil
}

// ListWorldSnapshots returns save snapshots belonging to a source world.
func ListWorldSnapshots(sourceWorldID string) ([]WorldSnapshotInfo, error) {
	var result []WorldSnapshotInfo
	if err := store.WriteTransaction(func(tx *gorm.DB) error {
		list, err := store.ListWorldSnapshotsBySourceWorldTx(tx, sourceWorldID, worldCopyReasonSnapshot)
		if err != nil {
			return err
		}
		decorated, txErr := decorateWorldSnapshotListTx(tx, list)
		result = decorated
		return txErr
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteWorldSnapshot deletes a saved snapshot world and its persisted data.
func DeleteWorldSnapshot(snapshotWorldID string) error {
	snapshotMeta, err := store.GetWorldSnapshotBySnapshotWorld(snapshotWorldID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errorf(ErrorNotFound, "snapshot metadata not found for world: %s", snapshotWorldID)
		}
		return err
	}
	if snapshotMeta.Reason != worldCopyReasonSnapshot {
		return codedErrorf(ErrorInvalid, snapshotValidationCodeReasonInvalid, "world %s is not a save snapshot", snapshotWorldID)
	}
	if err := store.WriteTransaction(func(tx *gorm.DB) error {
		return deleteWorldGraphTx(tx, snapshotWorldID)
	}); err != nil {
		return err
	}
	log.Printf("[snapshot-delete] snapshot=%s source=%s name=%q", snapshotWorldID, snapshotMeta.SourceWorldUUID, snapshotMeta.SnapshotName)
	return nil
}
