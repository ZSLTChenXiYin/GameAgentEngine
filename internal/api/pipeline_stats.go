package api

import (
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// GetPipelineStatsHandler returns lightweight observability stats for the shared data pipeline.
func GetPipelineStatsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"store":       store.GetPipelineStats(),
		"world_locks": service.GetWorldLockStats(),
	})
}
