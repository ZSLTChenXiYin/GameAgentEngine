package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/planner"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type llmParsedOutput struct {
	Reply                   string
	CleanedContent          string
	ParseError              string
	RawActionCalls          string
	RawMemoryUpdates        string
	RawPlan                 string
	RawRequestData          string
	RawInterimMemoryUpdates string
	RawFutureOutline        string
	RawSubTasks             string
}

// parseLLMJSON 解析大模型返回的 JSON 字符串。
func (p *Pipeline) parseLLMJSON(content string) *llmParsedOutput {
	out := &llmParsedOutput{Reply: content}

	cleaned := p.cleanJSON(content)
	out.CleanedContent = cleaned

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		out.ParseError = err.Error()
		log.Printf("[warn] LLM output is not valid JSON, using raw: %v", err)
		return out
	}

	if reply, ok := raw["reply"]; ok {
		var s string
		json.Unmarshal(reply, &s)
		out.Reply = s
	}

	if ac, ok := raw["action_calls"]; ok {
		out.RawActionCalls = string(ac)
	}

	if mu, ok := raw["memory_updates"]; ok {
		out.RawMemoryUpdates = string(mu)
	}

	if imu, ok := raw["interim_memory_updates"]; ok {
		out.RawInterimMemoryUpdates = string(imu)
	}
	if rd, ok := raw["request_data"]; ok {
		out.RawRequestData = string(rd)
	}
	if wcp, ok := raw["world_change_plan"]; ok {
		out.RawPlan = string(wcp)
	}
	if st, ok := raw["sub_tasks"]; ok {
		out.RawSubTasks = string(st)
	}

	if fo, ok := raw["future_outline"]; ok {
		var s string
		json.Unmarshal(fo, &s)
		out.RawFutureOutline = s
	}

	return out
}

// cleanJSON 清理模型输出中的 Markdown 代码块包装。
func (p *Pipeline) cleanJSON(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		lines := strings.SplitN(content, "\n", 2)
		if len(lines) == 2 {
			content = strings.TrimSpace(lines[1])
		}
	}
	if strings.HasSuffix(content, "```") {
		idx := strings.LastIndex(content, "```")
		if idx >= 0 {
			content = strings.TrimSpace(content[:idx])
		}
	}
	return content
}

func (p *Pipeline) parseActionCalls(rawJSON string, defaultNodeID string) []ActionCall {
	if rawJSON == "" || rawJSON == "null" || rawJSON == "[]" {
		return nil
	}

	var rawCalls []struct {
		ActionID string         `json:"action_id"`
		Args     map[string]any `json:"args"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &rawCalls); err != nil {
		log.Printf("[warn] parse action_calls: %v", err)
		return nil
	}

	var result []ActionCall
	for _, rc := range rawCalls {
		if rc.ActionID == "" {
			continue
		}
		ac := ActionCall{
			ActionID: rc.ActionID,
			Args:     rc.Args,
		}
		if ac.Args == nil {
			ac.Args = make(map[string]any)
		}
		if _, ok := ac.Args["node_id"]; !ok {
			ac.Args["node_id"] = defaultNodeID
		}
		result = append(result, ac)
	}
	return result
}

func (p *Pipeline) executeActions(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, policyEngine *planner.PolicyEngine, calls []ActionCall) []ActionCall {
	var result []ActionCall

	for _, call := range calls {
		actionID := call.ActionID
		args := call.Args

		if policyEngine != nil && !policyEngine.IsActionAllowed(actionID) {
			p.emitExecutionEvent(req, runtime, executionMode, "action_blocked", actionID, map[string]any{"action_id": actionID, "args": args, "reason": "policy_blocked"})
			log.Printf("[policy] action %s blocked", actionID)
			continue
		}

		if p.actionReg.IsSync(actionID) {
			out, err := p.actionReg.ExecuteSync(actionID, args)
			if err != nil {
				p.emitExecutionEvent(req, runtime, executionMode, "action_sync_failed", actionID, map[string]any{"action_id": actionID, "args": args, "error": err.Error()})
				log.Printf("[action:sync] %s failed: %v", actionID, err)
			} else {
				p.emitExecutionEvent(req, runtime, executionMode, "action_sync_succeeded", actionID, map[string]any{"action_id": actionID, "args": args, "result": out})
				log.Printf("[action:sync] %s success: %v", actionID, out)
			}
		} else if p.actionReg.IsAsync(actionID) {
			cbID := p.actionReg.CreateCallback(actionID, args)
			call.Mode = "async"
			call.CallbackID = cbID
			result = append(result, call)
			p.emitExecutionEvent(req, runtime, executionMode, "action_async_enqueued", actionID, map[string]any{"action_id": actionID, "args": args, "callback_id": cbID})
			log.Printf("[action:async] %s callback=%s", actionID, cbID)
		} else {
			call.Mode = "async"
			result = append(result, call)
			p.emitExecutionEvent(req, runtime, executionMode, "action_unknown_passthrough", actionID, map[string]any{"action_id": actionID, "args": args})
			log.Printf("[action:unknown] %s passed through", actionID)
		}
	}

	return result
}

func (p *Pipeline) parseMemoryUpdates(rawJSON string) []MemoryUpdate {
	if rawJSON == "" || rawJSON == "null" || rawJSON == "[]" {
		return nil
	}

	levelMap := map[string]MemoryLevel{
		"short_term": MemShortTerm,
		"long_term":  MemLongTerm,
		"shared":     MemShared,
		"world":      MemWorld,
	}

	var rawMems []struct {
		NodeID  string `json:"node_id"`
		Content string `json:"content"`
		Level   string `json:"level"`
		Tags    string `json:"tags,omitempty"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &rawMems); err != nil {
		log.Printf("[warn] parse memory_updates: %v", err)
		return nil
	}

	var result []MemoryUpdate
	for _, rm := range rawMems {
		if rm.Content == "" {
			continue
		}
		level := levelMap[rm.Level]
		if level == "" {
			level = MemShortTerm
		}
		result = append(result, MemoryUpdate{
			NodeID:  rm.NodeID,
			Content: rm.Content,
			Level:   level,
			Tags:    rm.Tags,
		})
	}
	return result
}

func (p *Pipeline) writeMemories(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, updates []MemoryUpdate) {
	for _, mu := range updates {
		nodeID := store.ResolveNodeUUID(mu.NodeID)
		if nodeID == 0 {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_write_skipped", mu.Content, map[string]any{"node_id": mu.NodeID, "reason": "unknown_node"})
			log.Printf("[warn] write memory: unknown node UUID %s", mu.NodeID)
			continue
		}
		mm := store.MemoryModel{
			NodeID:  nodeID,
			Content: mu.Content,
			Level:   string(mu.Level),
			Tags:    mu.Tags,
		}
		if err := store.CreateMemory(&mm); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_write_failed", mu.Content, map[string]any{"node_id": mu.NodeID, "level": mu.Level, "tags": mu.Tags, "error": err.Error()})
			log.Printf("write memory: %v", err)
			continue
		}
		p.emitExecutionEvent(req, runtime, executionMode, "memory_written", mu.Content, map[string]any{"node_id": mu.NodeID, "level": mu.Level, "tags": mu.Tags})
	}
}

func sanitizeRoles(messages []ChatMessage) []ChatMessage {
	result := make([]ChatMessage, len(messages))
	for i, m := range messages {
		role := m.Role
		switch role {
		case "player", "narrator", "observer":
			role = "user"
		case "npc", "agent":
			role = "assistant"
		default:
			role = m.Role
		}
		result[i] = ChatMessage{Role: role, Content: m.Content}
	}
	return result
}

func (p *Pipeline) handleDataRequest(policyEngine *planner.PolicyEngine, dr *DataRequest) string {
	var parts []string
	for _, q := range dr.Queries {
		switch q.Type {
		case "node_components":
			if comps, err := store.GetNodeComponents(q.NodeID); err == nil {
				for _, c := range comps {
					if q.Filter == "" || c.ComponentType == q.Filter {
						parts = append(parts, fmt.Sprintf("  【%s】%s", c.ComponentType, c.Data))
					}
				}
			}
		case "node_memories":
			if mems, err := store.GetNodeMemories(q.NodeID, limitOrDefault(q.Limit, 50)); err == nil {
				for _, m := range mems {
					parts = append(parts, fmt.Sprintf("  [记忆:%s] %s", m.Level, m.Content))
				}
			}
		case "node_relations":
			if rels, err := store.GetNodeRelations(q.NodeID); err == nil {
				for _, r := range rels {
					// SourceID/TargetID are int64 now; use UUID fields for display
					parts = append(parts, fmt.Sprintf("  [关系:%s] %s -> %s (权重:%d)", r.RelationType, r.SourceUUID, r.TargetUUID, r.Weight))
				}
			}
		case "memory_search":
			if mems, err := store.SearchMemories(q.NodeID, q.Filter, limitOrDefault(q.Limit, 10)); err == nil {
				for _, m := range mems {
					parts = append(parts, fmt.Sprintf("  [搜索记忆:%s] %s", m.Level, m.Content))
				}
			}
		case "policy_check":
			allowed := true
			if policyEngine != nil {
				allowed = policyEngine.IsActionAllowed(q.Filter)
			}
			status := "blocked"
			if allowed {
				status = "allowed"
			}
			parts = append(parts, fmt.Sprintf("  [策略检查] %s -> %s", q.Filter, status))
		case "node_detail":
			if nd, err := store.GetNode(q.NodeID); err == nil {
				parentInfo := "无"
				if nd.ParentUUID != nil {
					parentInfo = *nd.ParentUUID
				}
				parts = append(parts, fmt.Sprintf("  [节点] id=%s name=%s type=%s parent=%s world=%s 创建=%v", nd.UUID, nd.Name, nd.NodeType, parentInfo, nd.WorldUUID, nd.CreatedAt.Format("01-02 15:04")))
			}
		case "node_type_list":
			if nodes, err := store.GetAllNodes(q.NodeID, limitOrDefault(q.Limit, 20), 0, q.Filter); err == nil {
				for _, n := range nodes {
					parts = append(parts, fmt.Sprintf("  %s (%s) id=%s", n.Name, n.NodeType, n.UUID))
				}
			}
		case "world_timeline":
			if ticks, err := store.GetTimelineTicks(q.NodeID, limitOrDefault(q.Limit, 10)); err == nil {
				for _, t := range ticks {
					parts = append(parts, fmt.Sprintf("  [tick %d] %s type=%s", t.TickNumber, t.GameTime, t.TickType))
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}

func limitOrDefault(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}
