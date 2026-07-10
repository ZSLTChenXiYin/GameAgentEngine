package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	RuntimeTaskStatusPending   = "pending"
	RuntimeTaskStatusClaimed   = "claimed"
	RuntimeTaskStatusRunning   = "running"
	RuntimeTaskStatusReleased  = "released"
	RuntimeTaskStatusSucceeded = "succeeded"
	RuntimeTaskStatusFailed    = "failed"
	RuntimeTaskStatusCancelled = "cancelled"
)

var (
	ErrRuntimeTaskNotClaimable = errors.New("runtime task not claimable")
	ErrRuntimeTaskLeaseMismatch = errors.New("runtime task lease mismatch")
)

type RuntimeTaskListQuery struct {
	Consumer        string
	Category        string
	Statuses        []string
	Limit           int
	AvailableBefore *time.Time
}

func CreateRuntimeTask(m *RuntimeTaskModel) error {
	if m.TaskID == "" {
		m.TaskID = NewUUID()
	}
	if m.Status == "" {
		m.Status = RuntimeTaskStatusPending
	}
	if m.AvailableAt == nil {
		now := time.Now()
		m.AvailableAt = &now
	}
	return Write(func(db *gorm.DB) error {
		return db.Create(m).Error
	})
}

func GetRuntimeTask(taskID string) (*RuntimeTaskModel, error) {
	var m RuntimeTaskModel
	err := DB.Where("task_id = ?", taskID).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetRuntimeTaskByCallbackID(callbackID string) (*RuntimeTaskModel, error) {
	var m RuntimeTaskModel
	err := DB.Where("callback_id = ?", callbackID).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func ListRuntimeTasks(query RuntimeTaskListQuery) ([]RuntimeTaskModel, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 20
	}
	qb := DB.Model(&RuntimeTaskModel{})
	if query.Consumer != "" {
		qb = qb.Where("consumer = ?", query.Consumer)
	}
	if query.Category != "" {
		qb = qb.Where("category = ?", query.Category)
	}
	if len(query.Statuses) > 0 {
		qb = qb.Where("status IN ?", query.Statuses)
	}
	if query.AvailableBefore != nil {
		qb = qb.Where("available_at IS NULL OR available_at <= ?", *query.AvailableBefore)
	}
	var list []RuntimeTaskModel
	err := qb.Order("priority DESC").Order("created_at ASC").Limit(limit).Find(&list).Error
	return list, err
}

func ListPendingRuntimeTasks(consumer string, limit int) ([]RuntimeTaskModel, error) {
	now := time.Now()
	return ListRuntimeTasks(RuntimeTaskListQuery{
		Consumer:        consumer,
		Statuses:        []string{RuntimeTaskStatusPending, RuntimeTaskStatusReleased},
		Limit:           limit,
		AvailableBefore: &now,
	})
}

func ClaimRuntimeTask(taskID string, consumer string, leaseOwner string) (*RuntimeTaskModel, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task id required")
	}
	leaseToken := NewUUID()
	now := time.Now()
	updates := map[string]any{
		"status":              RuntimeTaskStatusClaimed,
		"lease_owner":         leaseOwner,
		"lease_token":         leaseToken,
		"claimed_at":          &now,
		"last_heartbeat_at":   &now,
		"attempt_count":       gorm.Expr("attempt_count + 1"),
		"error_message":       "",
	}
	err := WriteTransaction(func(tx *gorm.DB) error {
		qb := tx.Model(&RuntimeTaskModel{}).
			Where("task_id = ?", taskID).
			Where("status IN ?", []string{RuntimeTaskStatusPending, RuntimeTaskStatusReleased}).
			Where("available_at IS NULL OR available_at <= ?", now)
		if consumer != "" {
			qb = qb.Where("consumer = ?", consumer)
		}
		result := qb.Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrRuntimeTaskNotClaimable
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func HeartbeatRuntimeTask(taskID string, leaseToken string) (*RuntimeTaskModel, error) {
	if taskID == "" || leaseToken == "" {
		return nil, fmt.Errorf("task id and lease token required")
	}
	now := time.Now()
	err := Write(func(db *gorm.DB) error {
		result := db.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status = ? AND lease_token = ?", taskID, RuntimeTaskStatusClaimed, leaseToken).
			Updates(map[string]any{"last_heartbeat_at": &now})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrRuntimeTaskLeaseMismatch
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func ReleaseRuntimeTask(taskID string, leaseToken string, retryDelay time.Duration, errMsg string) (*RuntimeTaskModel, error) {
	if taskID == "" || leaseToken == "" {
		return nil, fmt.Errorf("task id and lease token required")
	}
	availableAt := time.Now().Add(retryDelay)
	err := Write(func(db *gorm.DB) error {
		result := db.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status = ? AND lease_token = ?", taskID, RuntimeTaskStatusClaimed, leaseToken).
			Updates(map[string]any{
				"status":            RuntimeTaskStatusReleased,
				"lease_owner":       "",
				"lease_token":       "",
				"available_at":      &availableAt,
				"last_heartbeat_at": nil,
				"error_message":     errMsg,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrRuntimeTaskLeaseMismatch
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func CompleteRuntimeTaskByCallbackID(callbackID string, status string, result any) error {
	if callbackID == "" {
		return nil
	}
	mappedStatus, errMsg := normalizeRuntimeTaskCompletionStatus(status, result)
	resultJSON := marshalRuntimeTaskJSON(result)
	now := time.Now()
	return Write(func(db *gorm.DB) error {
		updates := map[string]any{
			"status":            mappedStatus,
			"result_json":       resultJSON,
			"error_message":     errMsg,
			"lease_owner":       "",
			"lease_token":       "",
			"completed_at":      &now,
			"last_heartbeat_at": &now,
		}
		return db.Model(&RuntimeTaskModel{}).
			Where("callback_id = ?", callbackID).
			Where("status IN ?", []string{RuntimeTaskStatusPending, RuntimeTaskStatusClaimed, RuntimeTaskStatusRunning, RuntimeTaskStatusReleased}).
			Updates(updates).Error
	})
}

func normalizeRuntimeTaskCompletionStatus(status string, result any) (string, string) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "success", "completed", "ok":
		return RuntimeTaskStatusSucceeded, ""
	case "cancelled", "canceled":
		return RuntimeTaskStatusCancelled, marshalRuntimeTaskJSON(result)
	default:
		errMsg := marshalRuntimeTaskJSON(result)
		if text, ok := result.(string); ok {
			errMsg = text
		}
		if errMsg == "" {
			errMsg = normalized
		}
		return RuntimeTaskStatusFailed, errMsg
	}
}

func marshalRuntimeTaskJSON(value any) string {
	if value == nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}
