package planner

import (
	"log"
	"sync"
	"strings"
)

// PolicyEngine 根据入参判断动作是否允许执行。
//
// 优先级规则（E12）：
//   1. blocked_actions > allowed（默认）> safe_actions
//   2. 如果同一 action 同时出现在 blocked 和 safe 列表中，blocked 获胜。
//   3. 不在 blocked 中的动作默认允许执行。
//   4. safe_actions 仅用于额外标记，不阻止未标记的动作。
//
// 作用域优先级（E12）：
//   World-level policy > scope-level policy > node-level policy
//   当前实现仅支持 world-level policy；scope/node 级策略为未来扩展预留。
type PolicyEngine struct {
	mu             sync.RWMutex
	BlockedActions map[string]bool
	SafeActions    map[string]bool
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
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.BlockedActions = make(map[string]bool)
	pe.SafeActions = make(map[string]bool)
	for _, a := range blocked {
		pe.BlockedActions[a] = true
	}
	for _, a := range safe {
		pe.SafeActions[a] = true
	}
	// E12: detect conflicts - blocked always wins
	for _, a := range safe {
		if pe.BlockedActions[a] {
			log.Printf("[policy:conflict] action %q is BOTH blocked and safe; blocked wins", a)
		}
	}
}

// IsActionAllowed 判断某个动作是否被策略明确禁止。
func (pe *PolicyEngine) IsActionAllowed(actionID string) bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	if pe.BlockedActions[actionID] {
		return false
	}
	return true
}

// IsBlocked returns true if action is in the blocked list.
func (pe *PolicyEngine) IsBlocked(actionID string) bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.BlockedActions[actionID]
}

// Conflicts returns action IDs that appear in both blocked and safe lists.
func (pe *PolicyEngine) Conflicts() []string {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	var conflicts []string
	for a := range pe.SafeActions {
		if pe.BlockedActions[a] {
			conflicts = append(conflicts, a)
		}
	}
	return conflicts
}

// IsSafe returns true if action is in the safe list.
func (pe *PolicyEngine) IsSafe(actionID string) bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.SafeActions[actionID]
}

// Evaluate 对提议动作做策略过滤。
func (pe *PolicyEngine) Evaluate(plan *WorldChangePlan) *WorldChangePlan {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	var filtered []ProposedAction
	for _, action := range plan.ProposedActions {
		if pe.BlockedActions[action.APIName] {
			log.Printf("[policy:blocked] %s", action.APIName)
			continue
		}
		filtered = append(filtered, action)
	}
	plan.ProposedActions = filtered
	return plan
}

// ParseWorldChangePlanActions extracts deduplicated action API names from a plan.
func ParseWorldChangePlanActions(plan *WorldChangePlan) []string {
	if plan == nil {
		return nil
	}
	seen := map[string]bool{}
	var ids []string
	for _, a := range plan.ProposedActions {
		id := strings.TrimSpace(a.APIName)
		if id != "" && !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}
