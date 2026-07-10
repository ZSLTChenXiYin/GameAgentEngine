package store

import (
	"time"

	"gorm.io/gorm"
)

func CreateAsyncCallbackRecord(m *AsyncCallbackRecordModel) error {
	return Write(func(db *gorm.DB) error {
		return db.Create(m).Error
	})
}

func GetAsyncCallbackRecord(callbackID string) (*AsyncCallbackRecordModel, error) {
	var m AsyncCallbackRecordModel
	err := DB.Where("callback_id = ?", callbackID).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func UpdateAsyncCallbackRecord(callbackID string, updates map[string]any) error {
	return Write(func(db *gorm.DB) error {
		return db.Model(&AsyncCallbackRecordModel{}).Where("callback_id = ?", callbackID).Updates(updates).Error
	})
}

func CompleteAsyncCallbackRecord(callbackID string, status string, resultJSON string, errMsg string) error {
	now := time.Now()
	updates := map[string]any{
		"status":       status,
		"result_json":  resultJSON,
		"error_message": errMsg,
		"completed_at": &now,
	}
	return UpdateAsyncCallbackRecord(callbackID, updates)
}

func CreatePausedExecution(m *PausedExecutionModel) error {
	return Write(func(db *gorm.DB) error {
		return db.Create(m).Error
	})
}

func GetPausedExecution(executionID string) (*PausedExecutionModel, error) {
	var m PausedExecutionModel
	err := DB.Where("execution_id = ?", executionID).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetPausedExecutionByCallbackID(callbackID string) (*PausedExecutionModel, error) {
	var m PausedExecutionModel
	err := DB.Where("callback_id = ?", callbackID).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func UpdatePausedExecution(executionID string, updates map[string]any) error {
	return Write(func(db *gorm.DB) error {
		return db.Model(&PausedExecutionModel{}).Where("execution_id = ?", executionID).Updates(updates).Error
	})
}

func MarkPausedExecutionResumed(executionID string, resumePayloadJSON string) error {
	now := time.Now()
	return UpdatePausedExecution(executionID, map[string]any{
		"status":              "resuming",
		"resume_payload_json": resumePayloadJSON,
		"resumed_at":          &now,
	})
}

func MarkPausedExecutionCompleted(executionID string) error {
	now := time.Now()
	return UpdatePausedExecution(executionID, map[string]any{
		"status":       "completed",
		"completed_at": &now,
		"last_error":   "",
	})
}

func MarkPausedExecutionFailed(executionID string, errMsg string) error {
	return UpdatePausedExecution(executionID, map[string]any{
		"status":     "failed",
		"last_error": errMsg,
	})
}

