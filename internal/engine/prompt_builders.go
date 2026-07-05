package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func buildDialoguePrompt(systemContext string, nodeID string) string {
	return fmt.Sprintf(`你是一个游戏Agent系统中的NPC角色扮演引擎。

重要：如果你需要查询更多数据来进行回复，可以在输出中包含 request_data 字段。
例如：需要查询某个节点的记忆或组件时，可以输出 request_data 让引擎自动加载。

你的输出必须是以下JSON格式，不能包含其他任何文字：

{"reply":"NPC的对话回复","action_calls":[{"action_id":"add_memory","args":{"node_id":"%s","content":"记忆内容","level":"short_term"}}],"memory_updates":[{"node_id":"%s","content":"短期记忆","level":"short_term"}]}

可用action_id: add_memory, update_mood, send_dialogue, adjust_relation, spawn_item
memory_updates的level可选: short_term, long_term, shared, world

request_data格式（可选，需要时添加）:
{"label":"查询标签","target":"store","queries":[{"type":"node_components","node_id":"节点ID","filter":"","limit":10}]}

========== NPC角色设定 ==========
%s
`, nodeID, nodeID, systemContext)
}

type resourceState struct {
	Food    int `json:"food,omitempty"`
	Order   int `json:"order,omitempty"`
	Defense int `json:"defense,omitempty"`
	Morale  int `json:"morale,omitempty"`
	Treasury int `json:"treasury,omitempty"`
}

// buildWorldTickPrompt 构建世界刻推进任务的系统提示词。
func buildWorldTickPrompt(systemContext string, outline string) string {
	sb := &strings.Builder{}
	sb.WriteString(systemContext)
	sb.WriteString("\n\n你正在推演世界时间线。")

	if outline != "" {
		sb.WriteString("\n\n现有世界线大纲：\n")
		sb.WriteString(outline)
		sb.WriteString("\n\n请将上述大纲细化到当前时间刻度，同时更新未来大纲。")
	} else {
		sb.WriteString("\n请先生成一个未来数个回合的世界线大纲，再细化当前刻度的事件。")
	}

	sb.WriteString("\n\n如果你需要查询更多数据来细化推演，可以在输出中包含 request_data 字段。")
	sb.WriteString("\n引擎会自动加载你请求的数据并让你继续。")
	sb.WriteString("\n\n输出JSON格式：\n")
	sb.WriteString(`{"reply":"","action_calls":[],"memory_updates":[],"world_change_plan":{},"future_outline":"未来数个回合的大纲"}`)
	return sb.String()
}

// buildEventImpactPrompt 构建事件影响评估任务的系统提示词。
func buildEventImpactPrompt(systemContext string, eventDesc string, nodeID string) string {
	sb := &strings.Builder{}
	sb.WriteString(systemContext)
	sb.WriteString("\n\n一个重要事件发生了:\n")
	sb.WriteString(eventDesc)
	sb.WriteString("\n\n评估此事件对世界的影响。")

	sb.WriteString("\n\n如果你需要查询更多数据来评估影响，可以在输出中包含 request_data 字段。")
	sb.WriteString("\n引擎会自动加载你请求的数据并让你继续。")

	sb.WriteString("\n\n输出JSON格式：\n")
	sb.WriteString(`{"reply":"事件影响评估","action_calls":[],"memory_updates":[{"node_id":"` + nodeID + `","content":"共享记忆","level":"shared"}],"world_change_plan":{"impact_level":"major","summary":"影响摘要","world_events":[{"event_type":"diplomatic_shift","scope":"region","description":"事件描述","confidence":0.85}],"proposed_actions":[{"api_name":"adjust_relation","args":{"delta":-20}}]}}`)
	return sb.String()
}

// buildAutonomousPrompt 构建自主行为任务的系统提示词。
func buildAutonomousPrompt(systemContext string, nodeID string, cfg *AutonomousConfig) string {
	capabilities, _ := json.MarshalIndent(cfg.Capabilities, "", "  ")
	return fmt.Sprintf(`你是游戏 Agent 引擎中的自主行为决策器。

你正在为节点 %s 执行一次自主行为周期。你只能使用下面 capabilities 中显式列出的能力；未列出的 action_id 绝对不能输出。
什么都不做是有效选择。如果没有明确必要行动，请返回空 action_calls。

重要：如果你需要查询更多数据来决策，可以在输出中包含 request_data 字段。
引擎会自动加载你请求的数据并让你继续。

输出必须是 JSON，不能包含其他文字：
{"reply":"本次自主行为摘要","action_calls":[],"memory_updates":[],"request_data":{"target":"store","queries":[]}}

========== 节点上下文 ==========
%s

========== 允许能力 capabilities ==========
%s
`, nodeID, systemContext, string(capabilities))
}

// obfuscateResourceData 将 resource_state 等精确数值替换为定性描述，
// 防止 NPC 在对话中透露具体数字破坏沉浸感。仅对对话任务生效。
func obfuscateResourceData(prompt string) string {
	re := regexp.MustCompile(`【resource_state】\{[^}]*\}`)
	return re.ReplaceAllStringFunc(prompt, func(match string) string {
		return "【资源状态】(内部参考，不要透露精确数字)"
	})
}

// ==================== Parse & Execute Helpers ====================