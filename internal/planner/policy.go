package planner

import "log"

// PolicyEngine 根据入参判断动作是否允许执行。
type PolicyEngine struct {
	BlockedActions   map[string]bool
	SafeActions      map[string]bool
}

// NewPolicyEngine 创建一个空策略引擎。策略数据通过 SetActions 设置。
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		BlockedActions: make(map[string]bool),
		SafeActions:    make(map[string]bool),
	}
}

// SetActions 设置阻塞动作和安全动作列表。
func (pe *PolicyEngine) SetActions(blocked, safe []string) {
	pe.BlockedActions = make(map[string]bool)
	pe.SafeActions = make(map[string]bool)
	for _, a := range blocked {
		pe.BlockedActions[a] = true
	}
	for _, a := range safe {
		pe.SafeActions[a] = true
	}
}

// IsActionAllowed 判断某个动作是否被策略明确禁止。
func (pe *PolicyEngine) IsActionAllowed(actionID string) bool {
	if pe.BlockedActions[actionID] {
		return false
	}
	return true
}

// Evaluate 对提议动作做策略过滤。
func (pe *PolicyEngine) Evaluate(plan *WorldChangePlan) *WorldChangePlan {
	var filtered []ProposedAction
	for _, action := range plan.ProposedActions {
		if pe.BlockedActions[action.APIName] {
			log.Printf("[policy:blocked] %s", action.APIName)
			continue
		}
		if pe.SafeActions[action.APIName] {
			filtered = append(filtered, action)
		} else {
			// 不在安全名单中的动作也放行，仅在日志中记录
			log.Printf("[policy:unlisted] %s not in safe_actions, passing", action.APIName)
			filtered = append(filtered, action)
		}
	}
	plan.ProposedActions = filtered
	return plan
}
