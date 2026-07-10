package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func MakeListPendingRuntimeTasksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 20
		if raw := r.URL.Query().Get("limit"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil || parsed <= 0 || parsed > 200 {
				errorJSONCode(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 200")
				return
			}
			limit = parsed
		}
		consumer := r.URL.Query().Get("consumer")
		items, err := store.ListPendingRuntimeTasks(consumer, limit)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"tasks":    items,
			"count":    len(items),
			"consumer": consumer,
		})
	}
}

func MakeClaimRuntimeTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TaskID     string `json:"task_id"`
			Consumer   string `json:"consumer,omitempty"`
			LeaseOwner string `json:"lease_owner"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if req.TaskID == "" || req.LeaseOwner == "" {
			errorJSONCode(w, http.StatusBadRequest, "invalid_runtime_task_claim", "task_id and lease_owner required")
			return
		}
		item, err := store.ClaimRuntimeTask(req.TaskID, req.Consumer, req.LeaseOwner)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrRuntimeTaskNotClaimable):
				errorJSONCode(w, http.StatusConflict, "runtime_task_not_claimable", err.Error())
			default:
				errorJSON(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": item})
	}
}

func MakeHeartbeatRuntimeTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TaskID     string `json:"task_id"`
			LeaseToken string `json:"lease_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if req.TaskID == "" || req.LeaseToken == "" {
			errorJSONCode(w, http.StatusBadRequest, "invalid_runtime_task_heartbeat", "task_id and lease_token required")
			return
		}
		item, err := store.HeartbeatRuntimeTask(req.TaskID, req.LeaseToken)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrRuntimeTaskLeaseMismatch):
				errorJSONCode(w, http.StatusConflict, "runtime_task_lease_mismatch", err.Error())
			default:
				errorJSON(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": item})
	}
}

func MakeReleaseRuntimeTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TaskID       string `json:"task_id"`
			LeaseToken   string `json:"lease_token"`
			RetryDelayMs int    `json:"retry_delay_ms,omitempty"`
			ErrorMessage string `json:"error_message,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if req.TaskID == "" || req.LeaseToken == "" {
			errorJSONCode(w, http.StatusBadRequest, "invalid_runtime_task_release", "task_id and lease_token required")
			return
		}
		if req.RetryDelayMs < 0 {
			errorJSONCode(w, http.StatusBadRequest, "invalid_retry_delay_ms", "retry_delay_ms must be >= 0")
			return
		}
		item, err := store.ReleaseRuntimeTask(req.TaskID, req.LeaseToken, time.Duration(req.RetryDelayMs)*time.Millisecond, req.ErrorMessage)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrRuntimeTaskLeaseMismatch):
				errorJSONCode(w, http.StatusConflict, "runtime_task_lease_mismatch", err.Error())
			default:
				errorJSON(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": item})
	}
}
