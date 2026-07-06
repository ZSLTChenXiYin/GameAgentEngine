package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// GetWorldSettingsHandler 获取世界的运行设置。
func GetWorldSettingsHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	settings, err := store.GetOrCreateWorldSettings(worldID)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"world_id":                   settings.WorldID,
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
	})
}

// SetWorldSettingsHandler 更新世界的运行设置。
func SetWorldSettingsHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	var req struct {
		MemoryLimit              *int   `json:"memory_limit,omitempty"`
		MaxAnalysisRounds        *int   `json:"max_analysis_rounds,omitempty"`
		MaxContextDepth          *int   `json:"max_context_depth,omitempty"`
		AutoApply                *bool  `json:"auto_apply,omitempty"`
		RequireReviewAbove       string `json:"require_review_above,omitempty"`
		PipelineMode             string `json:"pipeline_mode,omitempty"`
		PropagationMaxDepth      *int   `json:"propagation_max_depth,omitempty"`
		SubTaskMaxRetries        *int   `json:"sub_task_max_retries,omitempty"`
		SubTaskTimeoutSecs       *int   `json:"sub_task_timeout_secs,omitempty"`
		EnablePropagationMachine *bool  `json:"enable_propagation_machine,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json: "+err.Error())
		return
	}

	partial := &store.WorldSettingsModel{WorldUUID: worldID}
	if req.MemoryLimit != nil {
		partial.MemoryLimit = *req.MemoryLimit
	}
	if req.MaxAnalysisRounds != nil {
		partial.MaxAnalysisRounds = *req.MaxAnalysisRounds
	}
	if req.MaxContextDepth != nil {
		partial.MaxContextDepth = *req.MaxContextDepth
	}
	if req.AutoApply != nil {
		partial.AutoApply = *req.AutoApply
	}
	if req.RequireReviewAbove != "" {
		partial.RequireReviewAbove = req.RequireReviewAbove
	}
	if req.PipelineMode != "" {
		if !engine.IsValidPipelineMode(req.PipelineMode) {
			errorJSONCode(w, http.StatusBadRequest, "invalid_pipeline_mode", "pipeline_mode must be one of: vertical, polling, full")
			return
		}
		partial.PipelineMode = req.PipelineMode
	}
	if req.PropagationMaxDepth != nil {
		partial.PropagationMaxDepth = *req.PropagationMaxDepth
	}
	if req.SubTaskMaxRetries != nil {
		partial.SubTaskMaxRetries = *req.SubTaskMaxRetries
	}
	if req.SubTaskTimeoutSecs != nil {
		partial.SubTaskTimeoutSecs = *req.SubTaskTimeoutSecs
	}
	if req.EnablePropagationMachine != nil {
		partial.EnablePropagationMachine = *req.EnablePropagationMachine
	}

	settings, err := store.UpsertWorldSettingsWithMask(worldID, partial, req.AutoApply != nil, req.EnablePropagationMachine != nil)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"world_id":                   settings.WorldID,
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
	})
}
