package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

// MakePlanApproveHandler 返回批准计划的 HTTP handler。
func MakePlanApproveHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		if worldID == "" {
			errorJSON(w, 400, "world_id required")
			return
		}

		var req struct {
			PlanID string `json:"plan_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid request body: "+err.Error())
			return
		}
		if req.PlanID == "" {
			errorJSON(w, 400, "plan_id required")
			return
		}

		plan, err := engine.GlobalPlanReview.Approve(req.PlanID)
		if err != nil {
			errorJSON(w, 404, err.Error())
			return
		}

		writeJSON(w, 200, map[string]any{
			"status":  "approved",
			"plan_id": req.PlanID,
			"plan":    plan,
		})
	}
}

// MakePlanRejectHandler 返回拒绝计划的 HTTP handler。
func MakePlanRejectHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		if worldID == "" {
			errorJSON(w, 400, "world_id required")
			return
		}

		var req struct {
			PlanID string `json:"plan_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid request body: "+err.Error())
			return
		}
		if req.PlanID == "" {
			errorJSON(w, 400, "plan_id required")
			return
		}

		if err := engine.GlobalPlanReview.Reject(req.PlanID); err != nil {
			errorJSON(w, 404, err.Error())
			return
		}

		writeJSON(w, 200, map[string]any{
			"status":  "rejected",
			"plan_id": req.PlanID,
		})
	}
}

// MakeListPendingPlansHandler 返回列出待审批计划的 HTTP handler。
func MakeListPendingPlansHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.URL.Query().Get("world_id")
		plans := engine.GlobalPlanReview.ListPending(worldID)
		writeJSON(w, 200, plans)
	}
}
