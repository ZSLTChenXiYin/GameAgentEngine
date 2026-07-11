package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
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

		// Execute the approved plan's side effects (memory writes, action calls)
		if err := p.ApplyPendingPlan(plan); err != nil {
			errorJSON(w, 500, fmt.Sprintf("apply plan: %v", err))
			return
		}

		// Persist approval status to database
		if err := store.UpdatePendingPlanStatus(req.PlanID, "approved"); err != nil {
			errorJSON(w, 500, fmt.Sprintf("persist plan approval: %v", err))
			return
		}

		writeJSON(w, 200, map[string]any{
			"status":  "approved",
			"plan_id": req.PlanID,
			"plan":    plan,
			"applied": true,
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

		// Persist rejection status to database
		if err := store.UpdatePendingPlanStatus(req.PlanID, "rejected"); err != nil {
			errorJSON(w, 500, fmt.Sprintf("persist plan rejection: %v", err))
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
