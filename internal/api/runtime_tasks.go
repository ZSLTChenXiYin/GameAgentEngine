package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func parseRuntimeTaskStatuses(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func MakeListRuntimeTasksHandler() http.HandlerFunc {
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
		query := store.RuntimeTaskListQuery{
			Consumer:      strings.TrimSpace(r.URL.Query().Get("consumer")),
			Category:      strings.TrimSpace(r.URL.Query().Get("category")),
			InterfaceName: strings.TrimSpace(r.URL.Query().Get("interface_name")),
			Transport:     strings.TrimSpace(r.URL.Query().Get("transport")),
			WorldUUID:     strings.TrimSpace(r.URL.Query().Get("world_id")),
			Statuses:      parseRuntimeTaskStatuses(r.URL.Query().Get("status")),
			Limit:         limit,
		}
		if raw := strings.TrimSpace(r.URL.Query().Get("available_only")); raw != "" && raw != "false" && raw != "0" {
			now := time.Now()
			query.AvailableBefore = &now
		}
		items, err := store.ListRuntimeTasks(query)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"tasks": items,
			"count": len(items),
			"query": query,
		})
	}
}

func MakeGetRuntimeTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := strings.TrimSpace(r.PathValue("task_id"))
		if taskID == "" {
			errorJSONCode(w, http.StatusBadRequest, "invalid_runtime_task_id", "task_id required")
			return
		}
		item, err := store.GetRuntimeTask(taskID)
		if err != nil {
			if store.IsRecordNotFound(err) {
				errorJSONCode(w, http.StatusNotFound, "runtime_task_not_found", "runtime task not found")
				return
			}
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": item})
	}
}

func MakeRuntimeTaskStatsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats, err := store.GetRuntimeTaskStats()
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"stats": stats})
	}
}

func MakeSweepRuntimeTaskHeartbeatTimeoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TimeoutSeconds int `json:"timeout_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if req.TimeoutSeconds <= 0 {
			errorJSONCode(w, http.StatusBadRequest, "invalid_timeout_seconds", "timeout_seconds must be > 0")
			return
		}
		affected, err := store.MarkRuntimeTasksHeartbeatTimeout(time.Duration(req.TimeoutSeconds) * time.Second)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":          "ok",
			"affected":        affected,
			"timeout_seconds": req.TimeoutSeconds,
		})
	}
}

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

func MakeStartRuntimeTaskHandler() http.HandlerFunc {
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
			errorJSONCode(w, http.StatusBadRequest, "invalid_runtime_task_start", "task_id and lease_token required")
			return
		}
		item, err := store.StartRuntimeTask(req.TaskID, req.LeaseToken)
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

func MakeRequeueRuntimeTaskHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			TaskID       string `json:"task_id"`
			RetryDelayMs int    `json:"retry_delay_ms,omitempty"`
			ErrorMessage string `json:"error_message,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if req.TaskID == "" {
			errorJSONCode(w, http.StatusBadRequest, "invalid_runtime_task_requeue", "task_id required")
			return
		}
		if req.RetryDelayMs < 0 {
			errorJSONCode(w, http.StatusBadRequest, "invalid_retry_delay_ms", "retry_delay_ms must be >= 0")
			return
		}
		item, err := store.RequeueHeartbeatTimeoutTask(req.TaskID, time.Duration(req.RetryDelayMs)*time.Millisecond, req.ErrorMessage)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrRuntimeTaskNotClaimable):
				errorJSONCode(w, http.StatusConflict, "runtime_task_not_requeueable", err.Error())
			default:
				errorJSON(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"task": item})
	}
}
