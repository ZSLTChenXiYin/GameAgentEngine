// Package planner 定义世界变更计划及其策略评估模型。
package planner

// WorldChangePlan 描述一次时间线推进后建议发生的世界变更。
type WorldChangePlan struct {
	TimelineID      string           `json:"timeline_id"`
	ImpactLevel     string           `json:"impact_level"`
	Summary         string           `json:"summary"`
	WorldEvents     []PlanEvent      `json:"world_events"`
	ProposedActions []ProposedAction `json:"proposed_actions"`
}

// PlanEvent 描述计划中的单个世界事件。
type PlanEvent struct {
	EventType   string  `json:"event_type"`
	Scope       string  `json:"scope"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// ProposedAction 描述计划建议触发的动作调用。
type ProposedAction struct {
	APIName string                 `json:"api_name"`
	Args    map[string]any `json:"args"`
}

// EvaluationResult 表示策略引擎对计划评估后的结果。
type EvaluationResult struct {
	Plan           *WorldChangePlan `json:"plan"`
	Approved       bool             `json:"approved"`
	RejectedAction []ProposedAction `json:"rejected_actions,omitempty"`
	Reason         string           `json:"reason,omitempty"`
}
