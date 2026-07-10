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
	RuntimeTaskStatusPending          = "pending"
	RuntimeTaskStatusDispatched       = "dispatched"
	RuntimeTaskStatusClaimed          = "claimed"
	RuntimeTaskStatusRunning          = "running"
	RuntimeTaskStatusHeartbeatTimeout = "heartbeat_timeout"
	RuntimeTaskStatusReleased         = "released"
	RuntimeTaskStatusSucceeded        = "succeeded"
	RuntimeTaskStatusFailed           = "failed"
	RuntimeTaskStatusCancelled        = "cancelled"
)

var (
	ErrRuntimeTaskNotClaimable    = errors.New("runtime task not claimable")
	ErrRuntimeTaskLeaseMismatch   = errors.New("runtime task lease mismatch")
	ErrRuntimeTaskNotDispatchable = errors.New("runtime task not dispatchable")
)

type RuntimeTaskListQuery struct {
	Consumer                  string
	Category                  string
	InterfaceName             string
	DeliveryMode              string
	Transport                 string
	WorldUUID                 string
	Statuses                  []string
	DispatchFailureClass      string
	DispatchDecision          string
	TransitionReason          string
	RetryExhaustedOnly        bool
	RepeatedHeartbeatTimeouts int
	DispatchedBefore          *time.Time
	Limit                     int
	AvailableBefore           *time.Time
}

type RuntimeTaskStats struct {
	GeneratedAt               time.Time        `json:"generated_at"`
	Total                     int64            `json:"total"`
	ReadyPull                 int64            `json:"ready_pull"`
	InFlight                  int64            `json:"in_flight"`
	Terminal                  int64            `json:"terminal"`
	HeartbeatTimeout          int64            `json:"heartbeat_timeout"`
	DispatchErrorTasks        int64            `json:"dispatch_error_tasks"`
	RetryExhaustedTasks       int64            `json:"retry_exhausted_tasks"`
	DispatchedWithoutCallback int64            `json:"dispatched_without_callback"`
	RepeatedHeartbeatTimeouts int64            `json:"repeated_heartbeat_timeouts"`
	OldestDispatchedAgeSecs   int64            `json:"oldest_dispatched_age_secs"`
	ByStatus                  map[string]int64 `json:"by_status"`
	ByCategory                map[string]int64 `json:"by_category"`
	ByConsumer                map[string]int64 `json:"by_consumer"`
	ByDeliveryMode            map[string]int64 `json:"by_delivery_mode"`
	ByTransport               map[string]int64 `json:"by_transport"`
	ByInterface               map[string]int64 `json:"by_interface"`
	ByDispatchFailureClass    map[string]int64 `json:"by_dispatch_failure_class"`
	ByDispatchDecision        map[string]int64 `json:"by_dispatch_decision"`
	ByHeartbeatTimeoutCount   map[string]int64 `json:"by_heartbeat_timeout_count"`
	OldestReadyTaskAgeSecs    int64            `json:"oldest_ready_task_age_secs"`
}

type RuntimeTaskDispatchMetadata struct {
	Transport             string
	FallbackTransport     string
	FallbackFromTransport string
	IdempotencyKey        string
	DispatchAttempts      int
	Result                any
	ErrorMessage          string
	StatusCode            int
	FailureClass          string
	Decision              string
	TransitionReason      string
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
	if query.InterfaceName != "" {
		qb = qb.Where("interface_name = ?", query.InterfaceName)
	}
	if query.DeliveryMode != "" {
		qb = qb.Where("delivery_mode = ?", query.DeliveryMode)
	}
	if query.Transport != "" {
		qb = qb.Where("transport = ?", query.Transport)
	}
	if query.WorldUUID != "" {
		qb = qb.Where("world_uuid = ?", query.WorldUUID)
	}
	if len(query.Statuses) > 0 {
		qb = qb.Where("status IN ?", query.Statuses)
	}
	if query.DispatchFailureClass != "" {
		qb = qb.Where("last_dispatch_failure_class = ?", query.DispatchFailureClass)
	}
	if query.DispatchDecision != "" {
		qb = qb.Where("last_dispatch_decision = ?", query.DispatchDecision)
	}
	if query.TransitionReason != "" {
		qb = qb.Where("last_transition_reason = ?", query.TransitionReason)
	}
	if query.RetryExhaustedOnly {
		qb = qb.Where("max_attempts > 0").Where("attempt_count >= max_attempts")
	}
	if query.RepeatedHeartbeatTimeouts > 0 {
		qb = qb.Where("heartbeat_timeout_count >= ?", query.RepeatedHeartbeatTimeouts)
	}
	if query.DispatchedBefore != nil {
		qb = qb.Where("status = ?", RuntimeTaskStatusDispatched).Where("dispatched_at IS NOT NULL AND dispatched_at <= ?", *query.DispatchedBefore)
	}
	if query.AvailableBefore != nil {
		qb = qb.Where("available_at IS NULL OR available_at <= ?", *query.AvailableBefore)
	}
	var list []RuntimeTaskModel
	err := qb.Order("priority DESC").Order("created_at ASC").Limit(limit).Find(&list).Error
	return list, err
}

func GetRuntimeTaskStats() (*RuntimeTaskStats, error) {
	stats := &RuntimeTaskStats{
		GeneratedAt:             time.Now(),
		ByStatus:                map[string]int64{},
		ByCategory:              map[string]int64{},
		ByConsumer:              map[string]int64{},
		ByDeliveryMode:          map[string]int64{},
		ByTransport:             map[string]int64{},
		ByInterface:             map[string]int64{},
		ByDispatchFailureClass:  map[string]int64{},
		ByDispatchDecision:      map[string]int64{},
		ByHeartbeatTimeoutCount: map[string]int64{},
	}
	if err := DB.Model(&RuntimeTaskModel{}).Count(&stats.Total).Error; err != nil {
		return nil, err
	}
	var err error
	if stats.ByStatus, err = aggregateRuntimeTaskCounts("status"); err != nil {
		return nil, err
	}
	if stats.ByCategory, err = aggregateRuntimeTaskCounts("category"); err != nil {
		return nil, err
	}
	if stats.ByConsumer, err = aggregateRuntimeTaskCounts("consumer"); err != nil {
		return nil, err
	}
	if stats.ByDeliveryMode, err = aggregateRuntimeTaskCounts("delivery_mode"); err != nil {
		return nil, err
	}
	if stats.ByTransport, err = aggregateRuntimeTaskCounts("transport"); err != nil {
		return nil, err
	}
	if stats.ByInterface, err = aggregateRuntimeTaskCounts("interface_name"); err != nil {
		return nil, err
	}
	if stats.ByDispatchFailureClass, err = aggregateRuntimeTaskCounts("last_dispatch_failure_class"); err != nil {
		return nil, err
	}
	if stats.ByDispatchDecision, err = aggregateRuntimeTaskCounts("last_dispatch_decision"); err != nil {
		return nil, err
	}
	if stats.ByHeartbeatTimeoutCount, err = aggregateRuntimeTaskCountsExpr("CAST(heartbeat_timeout_count AS TEXT)"); err != nil {
		return nil, err
	}
	now := time.Now()
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("status IN ?", []string{RuntimeTaskStatusPending, RuntimeTaskStatusReleased}).
		Where("available_at IS NULL OR available_at <= ?", now).
		Count(&stats.ReadyPull).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("status IN ?", []string{RuntimeTaskStatusDispatched, RuntimeTaskStatusClaimed, RuntimeTaskStatusRunning}).
		Count(&stats.InFlight).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("status IN ?", []string{RuntimeTaskStatusSucceeded, RuntimeTaskStatusFailed, RuntimeTaskStatusCancelled}).
		Count(&stats.Terminal).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("status = ?", RuntimeTaskStatusHeartbeatTimeout).
		Count(&stats.HeartbeatTimeout).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("last_dispatch_error IS NOT NULL").
		Where("TRIM(last_dispatch_error) <> ''").
		Count(&stats.DispatchErrorTasks).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("max_attempts > 0").
		Where("attempt_count >= max_attempts").
		Count(&stats.RetryExhaustedTasks).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("status = ?", RuntimeTaskStatusDispatched).
		Count(&stats.DispatchedWithoutCallback).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&RuntimeTaskModel{}).
		Where("heartbeat_timeout_count >= ?", 2).
		Count(&stats.RepeatedHeartbeatTimeouts).Error; err != nil {
		return nil, err
	}
	var oldest RuntimeTaskModel
	err = DB.Model(&RuntimeTaskModel{}).
		Where("status IN ?", []string{RuntimeTaskStatusPending, RuntimeTaskStatusReleased}).
		Where("available_at IS NULL OR available_at <= ?", now).
		Order("created_at ASC").
		First(&oldest).Error
	if err == nil {
		stats.OldestReadyTaskAgeSecs = int64(now.Sub(oldest.CreatedAt).Seconds())
	} else if !IsRecordNotFound(err) {
		return nil, err
	}
	var oldestDispatched RuntimeTaskModel
	err = DB.Model(&RuntimeTaskModel{}).
		Where("status = ?", RuntimeTaskStatusDispatched).
		Where("dispatched_at IS NOT NULL").
		Order("dispatched_at ASC").
		First(&oldestDispatched).Error
	if err == nil && oldestDispatched.DispatchedAt != nil {
		stats.OldestDispatchedAgeSecs = int64(now.Sub(*oldestDispatched.DispatchedAt).Seconds())
	} else if err != nil && !IsRecordNotFound(err) {
		return nil, err
	}
	return stats, nil
}

func aggregateRuntimeTaskCounts(field string) (map[string]int64, error) {
	return aggregateRuntimeTaskCountsExpr(field)
}

func aggregateRuntimeTaskCountsExpr(expr string) (map[string]int64, error) {
	type row struct {
		Key   string `gorm:"column:key"`
		Count int64  `gorm:"column:count"`
	}
	var rows []row
	err := DB.Model(&RuntimeTaskModel{}).
		Select(expr + " AS key, COUNT(*) AS count").
		Group(expr).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := map[string]int64{}
	for _, item := range rows {
		result[strings.TrimSpace(item.Key)] = item.Count
	}
	return result, nil
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
		"status":            RuntimeTaskStatusClaimed,
		"lease_owner":       leaseOwner,
		"lease_token":       leaseToken,
		"claimed_at":        &now,
		"last_heartbeat_at": &now,
		"attempt_count":     gorm.Expr("attempt_count + 1"),
		"error_message":     "",
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
			Where("task_id = ? AND status IN ? AND lease_token = ?", taskID, []string{RuntimeTaskStatusClaimed, RuntimeTaskStatusRunning}, leaseToken).
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

func StartRuntimeTask(taskID string, leaseToken string) (*RuntimeTaskModel, error) {
	if taskID == "" || leaseToken == "" {
		return nil, fmt.Errorf("task id and lease token required")
	}
	now := time.Now()
	err := Write(func(db *gorm.DB) error {
		result := db.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status = ? AND lease_token = ?", taskID, RuntimeTaskStatusClaimed, leaseToken).
			Updates(map[string]any{
				"status":            RuntimeTaskStatusRunning,
				"last_heartbeat_at": &now,
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

func MarkRuntimeTaskDispatched(taskID string, meta RuntimeTaskDispatchMetadata) (*RuntimeTaskModel, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task id required")
	}
	if meta.DispatchAttempts <= 0 {
		meta.DispatchAttempts = 1
	}
	now := time.Now()
	err := Write(func(db *gorm.DB) error {
		resultDB := db.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status IN ?", taskID, []string{RuntimeTaskStatusPending, RuntimeTaskStatusReleased}).
			Updates(map[string]any{
				"status":                      RuntimeTaskStatusDispatched,
				"transport":                   meta.Transport,
				"idempotency_key":             meta.IdempotencyKey,
				"dispatched_at":               &now,
				"last_dispatch_at":            &now,
				"dispatch_attempts":           gorm.Expr("dispatch_attempts + ?", meta.DispatchAttempts),
				"last_dispatch_status_code":   meta.StatusCode,
				"last_dispatch_error":         "",
				"last_dispatch_failure_class": "",
				"last_dispatch_decision":      firstNonEmptyRuntimeTaskValue(meta.Decision, "dispatched"),
				"fallback_from_transport":     "",
				"last_transition_reason":      firstNonEmptyRuntimeTaskValue(meta.TransitionReason, "push_dispatch_succeeded"),
				"result_json":                 marshalRuntimeTaskJSON(meta.Result),
				"error_message":               "",
			})
		if resultDB.Error != nil {
			return resultDB.Error
		}
		if resultDB.RowsAffected == 0 {
			return ErrRuntimeTaskNotDispatchable
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func RecordRuntimeTaskDispatchFailure(taskID string, keepPending bool, meta RuntimeTaskDispatchMetadata) (*RuntimeTaskModel, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task id required")
	}
	if meta.DispatchAttempts <= 0 {
		meta.DispatchAttempts = 1
	}
	status := RuntimeTaskStatusFailed
	if keepPending {
		status = RuntimeTaskStatusPending
	}
	now := time.Now()
	err := Write(func(db *gorm.DB) error {
		resultDB := db.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status IN ?", taskID, []string{RuntimeTaskStatusPending, RuntimeTaskStatusReleased}).
			Updates(map[string]any{
				"status":                      status,
				"idempotency_key":             meta.IdempotencyKey,
				"last_dispatch_at":            &now,
				"dispatch_attempts":           gorm.Expr("dispatch_attempts + ?", meta.DispatchAttempts),
				"last_dispatch_status_code":   meta.StatusCode,
				"last_dispatch_error":         meta.ErrorMessage,
				"last_dispatch_failure_class": firstNonEmptyRuntimeTaskValue(meta.FailureClass, "unknown"),
				"last_dispatch_decision":      firstNonEmptyRuntimeTaskValue(meta.Decision, ternaryRuntimeTaskDecision(keepPending, "pending_retry", "failed_terminal")),
				"last_transition_reason":      firstNonEmptyRuntimeTaskValue(meta.TransitionReason, "push_dispatch_failed"),
				"error_message":               meta.ErrorMessage,
			})
		if resultDB.Error != nil {
			return resultDB.Error
		}
		if resultDB.RowsAffected == 0 {
			return ErrRuntimeTaskNotDispatchable
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func RecordRuntimeTaskDispatchFallback(taskID string, meta RuntimeTaskDispatchMetadata) (*RuntimeTaskModel, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task id required")
	}
	if meta.DispatchAttempts <= 0 {
		meta.DispatchAttempts = 1
	}
	now := time.Now()
	err := Write(func(db *gorm.DB) error {
		resultDB := db.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status IN ?", taskID, []string{RuntimeTaskStatusPending, RuntimeTaskStatusReleased}).
			Updates(map[string]any{
				"status":                      RuntimeTaskStatusReleased,
				"transport":                   meta.FallbackTransport,
				"available_at":                &now,
				"idempotency_key":             meta.IdempotencyKey,
				"last_dispatch_at":            &now,
				"dispatch_attempts":           gorm.Expr("dispatch_attempts + ?", meta.DispatchAttempts),
				"last_dispatch_status_code":   meta.StatusCode,
				"last_dispatch_error":         meta.ErrorMessage,
				"last_dispatch_failure_class": firstNonEmptyRuntimeTaskValue(meta.FailureClass, "unknown"),
				"last_dispatch_decision":      firstNonEmptyRuntimeTaskValue(meta.Decision, "fallback_to_pull"),
				"fallback_from_transport":     meta.FallbackFromTransport,
				"last_transition_reason":      firstNonEmptyRuntimeTaskValue(meta.TransitionReason, "push_dispatch_failed_then_fallback"),
				"error_message":               meta.ErrorMessage,
			})
		if resultDB.Error != nil {
			return resultDB.Error
		}
		if resultDB.RowsAffected == 0 {
			return ErrRuntimeTaskNotDispatchable
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func runtimeTaskRetryExhausted(task *RuntimeTaskModel) bool {
	return task != nil && task.MaxAttempts > 0 && task.AttemptCount >= task.MaxAttempts
}

func runtimeTaskRetryExhaustedMessage(task *RuntimeTaskModel, errMsg string) string {
	base := strings.TrimSpace(errMsg)
	if base == "" {
		base = "runtime task retry limit exhausted"
	}
	if task == nil || task.MaxAttempts <= 0 {
		return base
	}
	return fmt.Sprintf("%s (attempt_count=%d max_attempts=%d)", base, task.AttemptCount, task.MaxAttempts)
}

func ReleaseRuntimeTask(taskID string, leaseToken string, retryDelay time.Duration, errMsg string) (*RuntimeTaskModel, error) {
	if taskID == "" || leaseToken == "" {
		return nil, fmt.Errorf("task id and lease token required")
	}
	availableAt := time.Now().Add(retryDelay)
	err := WriteTransaction(func(tx *gorm.DB) error {
		var task RuntimeTaskModel
		if err := tx.Where("task_id = ?", taskID).First(&task).Error; err != nil {
			if IsRecordNotFound(err) {
				return ErrRuntimeTaskLeaseMismatch
			}
			return err
		}
		if task.LeaseToken != leaseToken || (task.Status != RuntimeTaskStatusClaimed && task.Status != RuntimeTaskStatusRunning) {
			return ErrRuntimeTaskLeaseMismatch
		}
		updates := map[string]any{
			"lease_owner":       "",
			"lease_token":       "",
			"last_heartbeat_at": nil,
			"error_message":     errMsg,
		}
		if runtimeTaskRetryExhausted(&task) {
			now := time.Now()
			updates["status"] = RuntimeTaskStatusFailed
			updates["completed_at"] = &now
			updates["available_at"] = nil
			updates["error_message"] = runtimeTaskRetryExhaustedMessage(&task, errMsg)
		} else {
			updates["status"] = RuntimeTaskStatusReleased
			updates["available_at"] = &availableAt
		}
		return tx.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status IN ? AND lease_token = ?", taskID, []string{RuntimeTaskStatusClaimed, RuntimeTaskStatusRunning}, leaseToken).
			Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func MarkRuntimeTasksHeartbeatTimeout(timeout time.Duration) (int64, error) {
	if timeout <= 0 {
		return 0, fmt.Errorf("timeout must be > 0")
	}
	deadline := time.Now().Add(-timeout)
	now := time.Now()
	var affected int64
	err := Write(func(db *gorm.DB) error {
		result := db.Model(&RuntimeTaskModel{}).
			Where("status IN ?", []string{RuntimeTaskStatusClaimed, RuntimeTaskStatusRunning}).
			Where("last_heartbeat_at IS NOT NULL AND last_heartbeat_at < ?", deadline).
			Updates(map[string]any{
				"status":                  RuntimeTaskStatusHeartbeatTimeout,
				"heartbeat_timeout_at":    &now,
				"heartbeat_timeout_count": gorm.Expr("heartbeat_timeout_count + 1"),
				"lease_owner":             "",
				"lease_token":             "",
				"error_message":           "heartbeat timeout",
			})
		if result.Error != nil {
			return result.Error
		}
		affected = result.RowsAffected
		return nil
	})
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func RequeueHeartbeatTimeoutTask(taskID string, retryDelay time.Duration, errMsg string) (*RuntimeTaskModel, error) {
	if taskID == "" {
		return nil, fmt.Errorf("task id required")
	}
	availableAt := time.Now().Add(retryDelay)
	err := WriteTransaction(func(tx *gorm.DB) error {
		var task RuntimeTaskModel
		if err := tx.Where("task_id = ?", taskID).First(&task).Error; err != nil {
			if IsRecordNotFound(err) {
				return ErrRuntimeTaskNotClaimable
			}
			return err
		}
		if task.Status != RuntimeTaskStatusHeartbeatTimeout {
			return ErrRuntimeTaskNotClaimable
		}
		updates := map[string]any{
			"lease_owner":          "",
			"lease_token":          "",
			"heartbeat_timeout_at": nil,
			"error_message":        errMsg,
		}
		if runtimeTaskRetryExhausted(&task) {
			now := time.Now()
			updates["status"] = RuntimeTaskStatusFailed
			updates["completed_at"] = &now
			updates["available_at"] = nil
			updates["error_message"] = runtimeTaskRetryExhaustedMessage(&task, errMsg)
		} else {
			updates["status"] = RuntimeTaskStatusReleased
			updates["available_at"] = &availableAt
		}
		return tx.Model(&RuntimeTaskModel{}).
			Where("task_id = ? AND status = ?", taskID, RuntimeTaskStatusHeartbeatTimeout).
			Updates(updates).Error
	})
	if err != nil {
		return nil, err
	}
	return GetRuntimeTask(taskID)
}

func RequeueHeartbeatTimeoutTasksBatch(consumer string, category string, transport string, retryDelay time.Duration, errMsg string, limit int) (int64, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	query := DB.Model(&RuntimeTaskModel{}).
		Select("task_id").
		Where("status = ?", RuntimeTaskStatusHeartbeatTimeout).
		Order("created_at ASC").
		Limit(limit)
	if strings.TrimSpace(consumer) != "" {
		query = query.Where("consumer = ?", strings.TrimSpace(consumer))
	}
	if strings.TrimSpace(category) != "" {
		query = query.Where("category = ?", strings.TrimSpace(category))
	}
	if strings.TrimSpace(transport) != "" {
		query = query.Where("transport = ?", strings.TrimSpace(transport))
	}
	type taskRef struct {
		TaskID string `gorm:"column:task_id"`
	}
	var refs []taskRef
	if err := query.Scan(&refs).Error; err != nil {
		return 0, err
	}
	if len(refs) == 0 {
		return 0, nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		if strings.TrimSpace(ref.TaskID) != "" {
			ids = append(ids, ref.TaskID)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	var affected int64
	for _, id := range ids {
		if _, err := RequeueHeartbeatTimeoutTask(id, retryDelay, errMsg); err != nil {
			if errors.Is(err, ErrRuntimeTaskNotClaimable) {
				continue
			}
			return 0, err
		}
		affected++
	}
	return affected, nil
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
			"status":               mappedStatus,
			"result_json":          resultJSON,
			"error_message":        errMsg,
			"lease_owner":          "",
			"lease_token":          "",
			"completed_at":         &now,
			"last_heartbeat_at":    &now,
			"heartbeat_timeout_at": nil,
		}
		return db.Model(&RuntimeTaskModel{}).
			Where("callback_id = ?", callbackID).
			Where("status IN ?", []string{RuntimeTaskStatusPending, RuntimeTaskStatusDispatched, RuntimeTaskStatusClaimed, RuntimeTaskStatusRunning, RuntimeTaskStatusReleased, RuntimeTaskStatusHeartbeatTimeout}).
			Updates(updates).Error
	})
}

func UpdateRuntimeTaskTerminalCallbackFailure(callbackID string, errMsg string) error {
	if callbackID == "" {
		return nil
	}
	now := time.Now()
	return Write(func(db *gorm.DB) error {
		return db.Model(&RuntimeTaskModel{}).
			Where("callback_id = ?", callbackID).
			Where("status IN ?", []string{RuntimeTaskStatusPending, RuntimeTaskStatusDispatched, RuntimeTaskStatusReleased}).
			Updates(map[string]any{
				"status":        RuntimeTaskStatusFailed,
				"error_message": errMsg,
				"completed_at":  &now,
			}).Error
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

func firstNonEmptyRuntimeTaskValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func ternaryRuntimeTaskDecision(cond bool, whenTrue string, whenFalse string) string {
	if cond {
		return whenTrue
	}
	return whenFalse
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
