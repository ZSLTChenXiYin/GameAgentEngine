package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// GetWorldPolicyHandler 获取世界的动作策略。
func GetWorldPolicyHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	policy, err := store.GetWorldPolicy(worldID)
	if err != nil {
		writeJSON(w, 200, map[string]any{
			"world_id":        worldID,
			"blocked_actions": []string{},
			"safe_actions":    []string{},
		})
		return
	}
	writeJSON(w, 200, map[string]any{
		"world_id":        policy.WorldID,
		"blocked_actions": policy.ParseBlockedActions(),
		"safe_actions":    policy.ParseSafeActions(),
	})
}

// SetWorldPolicyHandler 创建或更新世界的动作策略。
func SetWorldPolicyHandler(w http.ResponseWriter, r *http.Request) {
	worldID := r.PathValue("world_id")
	var req struct {
		BlockedActions []string `json:"blocked_actions"`
		SafeActions    []string `json:"safe_actions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json: "+err.Error())
		return
	}
	policy, err := store.UpsertWorldPolicy(worldID, req.BlockedActions, req.SafeActions)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{
		"world_id":        policy.WorldID,
		"blocked_actions": policy.ParseBlockedActions(),
		"safe_actions":    policy.ParseSafeActions(),
	})
}
