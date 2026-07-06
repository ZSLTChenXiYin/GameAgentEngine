package store

import (
	"time"

	"gorm.io/gorm"
)

// GetWorldSettings 获取世界的运行设置。
func GetWorldSettings(worldUUID string) (*WorldSettingsModel, error) {
	var s WorldSettingsModel
	if err := DB.Where("world_uuid = ?", worldUUID).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

// GetOrCreateWorldSettings 获取或创建默认世界设置。
func GetOrCreateWorldSettings(worldUUID string) (*WorldSettingsModel, error) {
	s, err := GetWorldSettings(worldUUID)
	if err == nil {
		return s, nil
	}
	worldID := ResolveWorldUUID(worldUUID)
	s = &WorldSettingsModel{
		WorldID:                  worldID,
		WorldUUID:                worldUUID,
		MemoryLimit:              50,
		MaxAnalysisRounds:        5,
		MaxContextDepth:          3,
		AutoApply:                true,
		RequireReviewAbove:       "critical",
		EnablePropagationMachine: false,
		PropagationMaxDepth:      2,
		SubTaskMaxRetries:        2,
		SubTaskTimeoutSecs:       60,
		PipelineMode:             "full",
	}
	if err := DB.Create(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

func ApplyWorldSettingsUpdate(s *WorldSettingsModel, updates *WorldSettingsModel, autoApplySet, propagationMachineSet bool) {
	if updates.MemoryLimit > 0 {
		s.MemoryLimit = updates.MemoryLimit
	}
	if updates.MaxAnalysisRounds > 0 {
		s.MaxAnalysisRounds = updates.MaxAnalysisRounds
	}
	if updates.MaxContextDepth > 0 {
		s.MaxContextDepth = updates.MaxContextDepth
	}
	if updates.PipelineMode != "" {
		s.PipelineMode = updates.PipelineMode
	}
	if updates.PropagationMaxDepth > 0 {
		s.PropagationMaxDepth = updates.PropagationMaxDepth
	}
	if updates.SubTaskMaxRetries > 0 {
		s.SubTaskMaxRetries = updates.SubTaskMaxRetries
	}
	if updates.SubTaskTimeoutSecs > 0 {
		s.SubTaskTimeoutSecs = updates.SubTaskTimeoutSecs
	}
	if autoApplySet {
		s.AutoApply = updates.AutoApply
	}
	if updates.RequireReviewAbove != "" {
		s.RequireReviewAbove = updates.RequireReviewAbove
	}
	if propagationMachineSet {
		s.EnablePropagationMachine = updates.EnablePropagationMachine
	}
	s.UpdatedAt = time.Now()
}

// UpsertWorldSettings 更新世界的运行设置。零值字段不会覆盖现有值。
func UpsertWorldSettings(worldUUID string, settings *WorldSettingsModel) (*WorldSettingsModel, error) {
	return UpsertWorldSettingsWithMask(worldUUID, settings, false, false)
}

// UpsertWorldSettingsWithMask updates fields explicitly selected by the caller.
func UpsertWorldSettingsWithMask(worldUUID string, settings *WorldSettingsModel, autoApplySet, propagationMachineSet bool) (*WorldSettingsModel, error) {
	s, err := GetWorldSettings(worldUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s, err = GetOrCreateWorldSettings(worldUUID)
		}
		if err != nil {
			return nil, err
		}
	}

	ApplyWorldSettingsUpdate(s, settings, autoApplySet, propagationMachineSet)

	if err := DB.Save(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}
