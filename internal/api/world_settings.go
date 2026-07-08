package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func parseWorldTimeSettings(raw string) (*engine.WorldTimeSettings, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var settings engine.WorldTimeSettings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func encodeWorldTimeSettings(settings *engine.WorldTimeSettings) (string, error) {
	if settings == nil {
		return "", nil
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func writeWorldSettingsResponse(w http.ResponseWriter, settings *store.WorldSettingsModel) {
	worldTimeSettings, err := parseWorldTimeSettings(settings.WorldTimeSettingsJSON)
	if err != nil {
		errorJSON(w, 500, "invalid world_time_settings: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"world_id":                   settings.WorldUUID,
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
		"world_time_settings":        worldTimeSettings,
	})
}

func validatePositiveSetting(name string, value *int) error {
	if value != nil && *value <= 0 {
		return fmt.Errorf("%s must be greater than 0", name)
	}
	return nil
}

func validateNonNegativeSetting(name string, value *int) error {
	if value != nil && *value < 0 {
		return fmt.Errorf("%s must be greater than or equal to 0", name)
	}
	return nil
}

// GetWorldSettingsHandler 获取世界的运行设置。
func GetWorldSettingsHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	settings, err := store.GetOrCreateWorldSettings(worldID)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeWorldSettingsResponse(w, settings)
}

// SetWorldSettingsHandler 更新世界的运行设置。
func SetWorldSettingsHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	var req struct {
		MemoryLimit              *int                      `json:"memory_limit,omitempty"`
		MaxAnalysisRounds        *int                      `json:"max_analysis_rounds,omitempty"`
		MaxContextDepth          *int                      `json:"max_context_depth,omitempty"`
		AutoApply                *bool                     `json:"auto_apply,omitempty"`
		RequireReviewAbove       *string                   `json:"require_review_above,omitempty"`
		PipelineMode             *string                   `json:"pipeline_mode,omitempty"`
		PropagationMaxDepth      *int                      `json:"propagation_max_depth,omitempty"`
		SubTaskMaxRetries        *int                      `json:"sub_task_max_retries,omitempty"`
		SubTaskTimeoutSecs       *int                      `json:"sub_task_timeout_secs,omitempty"`
		EnablePropagationMachine *bool                     `json:"enable_propagation_machine,omitempty"`
		WorldTimeSettings        *engine.WorldTimeSettings `json:"world_time_settings,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json: "+err.Error())
		return
	}
	if err := validatePositiveSetting("memory_limit", req.MemoryLimit); err != nil {
		errorJSONCode(w, http.StatusBadRequest, "invalid_world_setting", err.Error())
		return
	}
	if err := validatePositiveSetting("max_analysis_rounds", req.MaxAnalysisRounds); err != nil {
		errorJSONCode(w, http.StatusBadRequest, "invalid_world_setting", err.Error())
		return
	}
	if err := validatePositiveSetting("max_context_depth", req.MaxContextDepth); err != nil {
		errorJSONCode(w, http.StatusBadRequest, "invalid_world_setting", err.Error())
		return
	}
	if err := validateNonNegativeSetting("propagation_max_depth", req.PropagationMaxDepth); err != nil {
		errorJSONCode(w, http.StatusBadRequest, "invalid_world_setting", err.Error())
		return
	}
	if err := validateNonNegativeSetting("sub_task_max_retries", req.SubTaskMaxRetries); err != nil {
		errorJSONCode(w, http.StatusBadRequest, "invalid_world_setting", err.Error())
		return
	}
	if err := validateNonNegativeSetting("sub_task_timeout_secs", req.SubTaskTimeoutSecs); err != nil {
		errorJSONCode(w, http.StatusBadRequest, "invalid_world_setting", err.Error())
		return
	}

	if req.RequireReviewAbove != nil && *req.RequireReviewAbove == "" {
		errorJSONCode(w, http.StatusBadRequest, "invalid_world_setting", "require_review_above must not be empty")
		return
	}
	if req.PipelineMode != nil {
		if *req.PipelineMode == "" || !engine.IsValidPipelineMode(*req.PipelineMode) {
			errorJSONCode(w, http.StatusBadRequest, "invalid_pipeline_mode", "pipeline_mode must be one of: vertical, polling, full")
			return
		}
	}
	updates := &store.WorldSettingsModel{}
	mask := &store.WorldSettingsUpdateMask{}
	if req.MemoryLimit != nil {
		updates.MemoryLimit = *req.MemoryLimit
		mask.MemoryLimit = true
	}
	if req.MaxAnalysisRounds != nil {
		updates.MaxAnalysisRounds = *req.MaxAnalysisRounds
		mask.MaxAnalysisRounds = true
	}
	if req.MaxContextDepth != nil {
		updates.MaxContextDepth = *req.MaxContextDepth
		mask.MaxContextDepth = true
	}
	if req.AutoApply != nil {
		updates.AutoApply = *req.AutoApply
		mask.AutoApply = true
	}
	if req.RequireReviewAbove != nil {
		updates.RequireReviewAbove = *req.RequireReviewAbove
		mask.RequireReviewAbove = true
	}
	if req.PipelineMode != nil {
		updates.PipelineMode = *req.PipelineMode
		mask.PipelineMode = true
	}
	if req.PropagationMaxDepth != nil {
		updates.PropagationMaxDepth = *req.PropagationMaxDepth
		mask.PropagationMaxDepth = true
	}
	if req.SubTaskMaxRetries != nil {
		updates.SubTaskMaxRetries = *req.SubTaskMaxRetries
		mask.SubTaskMaxRetries = true
	}
	if req.SubTaskTimeoutSecs != nil {
		updates.SubTaskTimeoutSecs = *req.SubTaskTimeoutSecs
		mask.SubTaskTimeoutSecs = true
	}
	if req.EnablePropagationMachine != nil {
		updates.EnablePropagationMachine = *req.EnablePropagationMachine
		mask.EnablePropagationMachine = true
	}
	if req.WorldTimeSettings != nil {
		encoded, err := encodeWorldTimeSettings(req.WorldTimeSettings)
		if err != nil {
			errorJSONCode(w, http.StatusBadRequest, "invalid_world_time_settings", "world_time_settings must be valid structured JSON: "+err.Error())
			return
		}
		updates.WorldTimeSettingsJSON = encoded
		mask.WorldTimeSettings = true
	}
	settings, err := store.UpsertWorldSettingsWithMask(worldID, updates, mask)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeWorldSettingsResponse(w, settings)
}
