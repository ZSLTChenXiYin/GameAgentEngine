package store

import "time"

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
		WorldID:            worldID,
		WorldUUID:          worldUUID,
		MemoryLimit:        50,
		MaxAnalysisRounds:  5,
		MaxContextDepth:    3,
		AutoApply:          true,
		RequireReviewAbove: "critical",
	}
	if err := DB.Create(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

// UpsertWorldSettings 更新世界的运行设置。零值字段不会覆盖现有值。
func UpsertWorldSettings(worldUUID string, settings *WorldSettingsModel) (*WorldSettingsModel, error) {
	s, err := GetWorldSettings(worldUUID)
	if err != nil {
		// 不存在则创建
		worldID := ResolveWorldUUID(worldUUID)
		s = &WorldSettingsModel{WorldID: worldID, WorldUUID: worldUUID}
	}

	if settings.MemoryLimit > 0 {
		s.MemoryLimit = settings.MemoryLimit
	}
	if settings.MaxAnalysisRounds > 0 {
		s.MaxAnalysisRounds = settings.MaxAnalysisRounds
	}
	if settings.MaxContextDepth > 0 {
		s.MaxContextDepth = settings.MaxContextDepth
	}
	s.AutoApply = settings.AutoApply
	if settings.RequireReviewAbove != "" {
		s.RequireReviewAbove = settings.RequireReviewAbove
	}
	s.UpdatedAt = time.Now()

	if err := DB.Save(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}
