package store

import (
	"time"

	"gorm.io/gorm"
)

type WorldSettingsUpdateMask struct {
	MemoryLimit              bool
	MaxAnalysisRounds        bool
	MaxContextDepth          bool
	AutoApply                bool
	RequireReviewAbove       bool
	PipelineMode             bool
	PropagationMaxDepth      bool
	SubTaskMaxRetries        bool
	SubTaskTimeoutSecs       bool
	EnablePropagationMachine bool
	WorldTimeSettings        bool
}

// GetWorldSettings 获取世界的运行设置。
func GetWorldSettings(worldUUID string) (*WorldSettingsModel, error) {
	return GetWorldSettingsTx(DB, worldUUID)
}

func GetWorldSettingsTx(tx *gorm.DB, worldUUID string) (*WorldSettingsModel, error) {
	var s WorldSettingsModel
	if err := tx.Where("world_uuid = ?", worldUUID).First(&s).Error; err != nil {
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
	if err := Writer().Create(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

func ApplyWorldSettingsUpdate(s *WorldSettingsModel, updates *WorldSettingsModel, mask *WorldSettingsUpdateMask) {
	if mask == nil {
		mask = &WorldSettingsUpdateMask{}
	}
	if mask.MemoryLimit {
		s.MemoryLimit = updates.MemoryLimit
	}
	if mask.MaxAnalysisRounds {
		s.MaxAnalysisRounds = updates.MaxAnalysisRounds
	}
	if mask.MaxContextDepth {
		s.MaxContextDepth = updates.MaxContextDepth
	}
	if mask.AutoApply {
		s.AutoApply = updates.AutoApply
	}
	if mask.RequireReviewAbove {
		s.RequireReviewAbove = updates.RequireReviewAbove
	}
	if mask.PipelineMode {
		s.PipelineMode = updates.PipelineMode
	}
	if mask.PropagationMaxDepth {
		s.PropagationMaxDepth = updates.PropagationMaxDepth
	}
	if mask.SubTaskMaxRetries {
		s.SubTaskMaxRetries = updates.SubTaskMaxRetries
	}
	if mask.SubTaskTimeoutSecs {
		s.SubTaskTimeoutSecs = updates.SubTaskTimeoutSecs
	}
	if mask.EnablePropagationMachine {
		s.EnablePropagationMachine = updates.EnablePropagationMachine
	}
	if mask.WorldTimeSettings {
		s.WorldTimeSettingsJSON = updates.WorldTimeSettingsJSON
	}
	s.UpdatedAt = time.Now()
}

// UpsertWorldSettings 更新世界的运行设置。零值字段不会覆盖现有值。
func UpsertWorldSettings(worldUUID string, settings *WorldSettingsModel) (*WorldSettingsModel, error) {
	return UpsertWorldSettingsWithMask(worldUUID, settings, &WorldSettingsUpdateMask{
		MemoryLimit:         settings.MemoryLimit > 0,
		MaxAnalysisRounds:   settings.MaxAnalysisRounds > 0,
		MaxContextDepth:     settings.MaxContextDepth > 0,
		RequireReviewAbove:  settings.RequireReviewAbove != "",
		PipelineMode:        settings.PipelineMode != "",
		PropagationMaxDepth: settings.PropagationMaxDepth > 0,
		SubTaskMaxRetries:   settings.SubTaskMaxRetries > 0,
		SubTaskTimeoutSecs:  settings.SubTaskTimeoutSecs > 0,
		WorldTimeSettings:   settings.WorldTimeSettingsJSON != "",
	})
}

// UpsertWorldSettingsWithMask updates fields explicitly selected by the caller.
func UpsertWorldSettingsWithMask(worldUUID string, settings *WorldSettingsModel, mask *WorldSettingsUpdateMask) (*WorldSettingsModel, error) {
	s, err := GetWorldSettings(worldUUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s, err = GetOrCreateWorldSettings(worldUUID)
		}
		if err != nil {
			return nil, err
		}
	}

	ApplyWorldSettingsUpdate(s, settings, mask)

	if err := Writer().Save(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}
