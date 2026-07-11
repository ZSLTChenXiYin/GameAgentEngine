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

// ApplyPendingPlan executes the action calls and memory updates stored in an approved
// world change plan. It is called when a human approves a previously deferred plan.
//
// The function uses the pipeline's action registry and memory writing infrastructure
// to apply the plan's effects, and records the execution in the inference log.
func (p *Pipeline) ApplyPendingPlan(plan *PendingPlan) error {
	req := &InvokeRequest{
		WorldID:  plan.WorldID,
		TaskType: plan.TaskType,
		NodeID:   plan.WorldID,
	}
	executionMode := ModeProduction
	_, maxRounds, retries, timeout, pipelineMode := p.loadWorldSettings(plan.WorldID)
	configuredMode := PipelineMode(pipelineMode)
	if configuredMode == "" {
		configuredMode = PipelineFull
	}
	runtime := &executionConfig{
		maxRounds:              maxRounds,
		subTaskRetries:         retries,
		subTaskTimeout:         timeout,
		configuredPipelineMode: configuredMode,
		pipelineMode:           configuredMode,
		policyEngine:           p.loadWorldPolicy(plan.WorldID),
	}

	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline_review",
		EventName:  "plan_apply_started",
		Message:    fmt.Sprintf("applying approved plan %s for world %s", plan.PlanID, plan.WorldID),
		DetailData: marshalLogDetail(plan),
	})

	// Write memory updates
	if len(plan.MemoryUpdates) > 0 {
		p.writeMemories(req, runtime, executionMode, plan.MemoryUpdates)
		for _, mem := range plan.MemoryUpdates {
			p.PropagateMemoryByRule(req, runtime, executionMode, mem, mem.NodeID)
		}
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline_review",
			EventName:  "plan_memories_applied",
			Message:    fmt.Sprintf("applied %d memory updates from plan %s", len(plan.MemoryUpdates), plan.PlanID),
			DetailData: marshalLogDetail(plan.MemoryUpdates),
		})
	}

	// Execute action calls
	if len(plan.ActionCalls) > 0 {
		p.executeActions(req, runtime, executionMode, runtime.policyEngine, plan.ActionCalls)
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline_review",
			EventName:  "plan_actions_applied",
			Message:    fmt.Sprintf("applied %d action calls from plan %s", len(plan.ActionCalls), plan.PlanID),
			DetailData: marshalLogDetail(plan.ActionCalls),
		})
	}

	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline_review",
		EventName:  "plan_apply_completed",
		Message:    fmt.Sprintf("plan %s applied successfully", plan.PlanID),
	})

	return nil
}
