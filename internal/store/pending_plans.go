package store

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// CreatePendingPlan persists a pending plan to the database.
func CreatePendingPlan(planID, worldUUID, taskType, status, dataJSON string, tickNumber int) error {
	return Write(func(db *gorm.DB) error {
		return db.Create(&PendingPlanModel{
			PlanID:     planID,
			WorldUUID:  worldUUID,
			TaskType:   taskType,
			Status:     status,
			DataJSON:   dataJSON,
			TickNumber: tickNumber,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}).Error
	})
}

// GetPendingPlan retrieves a single pending plan by its plan ID.
func GetPendingPlan(planID string) (*PendingPlanModel, error) {
	var m PendingPlanModel
	if err := DB.Where("plan_id = ?", planID).First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListPendingPlans returns all pending (not yet approved/rejected) plans.
func ListPendingPlans(worldUUID string) ([]PendingPlanModel, error) {
	var list []PendingPlanModel
	q := DB.Where("status = ?", "pending")
	if worldUUID != "" {
		q = q.Where("world_uuid = ?", worldUUID)
	}
	if err := q.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpdatePendingPlanStatus updates the status of a pending plan.
func UpdatePendingPlanStatus(planID, status string) error {
	return Write(func(db *gorm.DB) error {
		return db.Model(&PendingPlanModel{}).Where("plan_id = ?", planID).Updates(map[string]any{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
	})
}

// SerializePendingPlan serializes a PendingPlan into JSON for DB storage.
func SerializePendingPlan(plan interface{}) (string, error) {
	data, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
