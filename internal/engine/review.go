package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PendingPlan 表示一条待审批的世界变更计划。
type PendingPlan struct {
	PlanID         string           `json:"plan_id"`
	WorldID        string           `json:"world_id"`
	TickNumber     int              `json:"tick_number"`
	TaskType       TaskType         `json:"task_type"`
	WorldChangePlan *WorldChangePlan `json:"world_change_plan"`
	ActionCalls    []ActionCall     `json:"action_calls,omitempty"`
	MemoryUpdates  []MemoryUpdate   `json:"memory_updates,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	Status         string           `json:"status"` // pending / approved / rejected
}

// PlanReviewStore 管理待审批计划的存储。
type PlanReviewStore struct {
	mu      sync.Mutex
	plans   map[string]*PendingPlan
}

// NewPlanReviewStore 创建计划审批存储。
func NewPlanReviewStore() *PlanReviewStore {
	return &PlanReviewStore{
		plans: make(map[string]*PendingPlan),
	}
}

// Add 添加一条待审批计划。
func (s *PlanReviewStore) Add(plan *PendingPlan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans[plan.PlanID] = plan
}

// Get 获取指定计划。
func (s *PlanReviewStore) Get(planID string) (*PendingPlan, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.plans[planID]
	return p, ok
}

// Approve 批准计划，返回计划本身。
func (s *PlanReviewStore) Approve(planID string) (*PendingPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.plans[planID]
	if !ok {
		return nil, fmt.Errorf("plan %s not found", planID)
	}
	if p.Status != "pending" {
		return nil, fmt.Errorf("plan %s is not pending (status: %s)", planID, p.Status)
	}
	p.Status = "approved"
	return p, nil
}

// Reject 拒绝计划。
func (s *PlanReviewStore) Reject(planID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.plans[planID]
	if !ok {
		return fmt.Errorf("plan %s not found", planID)
	}
	if p.Status != "pending" {
		return fmt.Errorf("plan %s is not pending (status: %s)", planID, p.Status)
	}
	p.Status = "rejected"
	return nil
}

// ListPending 返回所有待审批计划。
func (s *PlanReviewStore) ListPending(worldID string) []*PendingPlan {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []*PendingPlan
	for _, p := range s.plans {
		if p.Status == "pending" && (worldID == "" || p.WorldID == worldID) {
			result = append(result, p)
		}
	}
	return result
}

// GlobalPlanReview 是全局计划审批存储。
var GlobalPlanReview = NewPlanReviewStore()

// NewPendingPlanID 生成一个新的计划 ID。
func NewPendingPlanID() string {
	return "plan_" + uuid.NewString()[:8]
}

// IsHighImpact 判断影响等级是否需要审批。
// 当 impact_level 为 "major" 或 "critical" 时返回 true。
func IsHighImpact(impactLevel string) bool {
	return impactLevel == "major" || impactLevel == "critical"
}
