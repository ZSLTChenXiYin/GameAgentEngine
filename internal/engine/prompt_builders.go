package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func builtinStoreRequestTool() LLMToolDefinition {
	return LLMToolDefinition{
		Name:        "request_store_data",
		Description: "Request additional engine-side store data for the next reasoning round",
		Invocation:  LLMToolInvocationDataRequest,
		DataRequest: &LLMDataRequestTemplate{Target: "store", Label: "request_store_data"},
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"label": map[string]any{"type": "string"},
				"queries": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"type": map[string]any{
								"type": "string",
								"enum": []any{"node_components", "node_memories", "node_relations", "memory_search", "policy_check", "node_detail", "node_type_list", "world_timeline"},
							},
							"node_id": map[string]any{"type": "string"},
							"filter":  map[string]any{"type": "string"},
							"limit":   map[string]any{"type": "integer"},
						},
						"required": []string{"type"},
					},
					"minItems": 1,
				},
			},
			"required": []string{"queries"},
		},
	}
}

func builtinActionTools(actionIDs []string) []LLMToolDefinition {
	if len(actionIDs) == 0 {
		return nil
	}
	builtins := map[string]LLMToolDefinition{
		"update_mood": {
			Name:        "update_mood",
			Description: "Update a character mood profile and write a short-term memory",
			Invocation:  LLMToolInvocationAction,
			ActionID:    "update_mood",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"node_id":   map[string]any{"type": "string"},
					"mood":      map[string]any{"type": "string"},
					"intensity": map[string]any{"type": "number"},
				},
				"required": []string{"node_id", "mood"},
			},
		},
		"add_memory": {
			Name:        "add_memory",
			Description: "Persist a memory record on a target node",
			Invocation:  LLMToolInvocationAction,
			ActionID:    "add_memory",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"node_id": map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
					"level":   map[string]any{"type": "string", "enum": []any{"short_term", "long_term", "shared", "world"}},
					"tags":    map[string]any{"type": "string"},
				},
				"required": []string{"node_id", "content"},
			},
		},
		"send_dialogue": {
			Name:        "send_dialogue",
			Description: "Emit a dialogue line and record it as short-term memory",
			Invocation:  LLMToolInvocationAction,
			ActionID:    "send_dialogue",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"node_id": map[string]any{"type": "string"},
					"content": map[string]any{"type": "string"},
				},
				"required": []string{"content"},
			},
		},
		"adjust_relation": {
			Name:        "adjust_relation",
			Description: "Request an async relation adjustment through the game bridge",
			Invocation:  LLMToolInvocationAction,
			ActionID:    "adjust_relation",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			},
		},
		"spawn_item": {
			Name:        "spawn_item",
			Description: "Request an async item spawn through the game bridge",
			Invocation:  LLMToolInvocationAction,
			ActionID:    "spawn_item",
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": true,
			},
		},
	}
	var tools []LLMToolDefinition
	for _, id := range actionIDs {
		if tool, ok := builtins[id]; ok {
			tools = append(tools, tool)
		}
	}
	return tools
}

func appendUniqueTools(base []LLMToolDefinition, extra ...LLMToolDefinition) []LLMToolDefinition {
	seen := map[string]struct{}{}
	result := make([]LLMToolDefinition, 0, len(base)+len(extra))
	for _, tool := range base {
		if strings.TrimSpace(tool.Name) == "" {
			continue
		}
		if _, ok := seen[tool.Name]; ok {
			continue
		}
		seen[tool.Name] = struct{}{}
		result = append(result, tool)
	}
	for _, tool := range extra {
		if strings.TrimSpace(tool.Name) == "" {
			continue
		}
		if _, ok := seen[tool.Name]; ok {
			continue
		}
		seen[tool.Name] = struct{}{}
		result = append(result, tool)
	}
	return result
}

func buildDynamicInterfaceTools(dynamicInterfaces []DynamicInterface) []LLMToolDefinition {
	if len(dynamicInterfaces) == 0 {
		return nil
	}
	var tools []LLMToolDefinition
	for _, item := range dynamicInterfaces {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		desc := strings.TrimSpace(item.Description)
		switch item.Kind {
		case DynamicInterfaceDataRequest:
			queryItems := make([]any, 0, len(item.QueryTypes))
			for _, queryType := range item.QueryTypes {
				queryItems = append(queryItems, map[string]any{"type": "string", "const": queryType})
			}
			tool := LLMToolDefinition{
				Name:        item.ID,
				Description: firstNonEmpty(desc, "Request game-client data for this turn"),
				Invocation:  LLMToolInvocationDataRequest,
				DataRequest: &LLMDataRequestTemplate{Target: "game_client", Label: item.ID, ExternalInterface: item.ExternalInterface},
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"queries": map[string]any{
							"type": "array",
							"items": map[string]any{
								"type": "object",
								"properties": map[string]any{
									"type":    map[string]any{"oneOf": queryItems},
									"node_id": map[string]any{"type": "string"},
									"filter":  map[string]any{"type": "string"},
									"limit":   map[string]any{"type": "integer"},
								},
								"required": []string{"type"},
							},
							"minItems": 1,
						},
					},
					"required": []string{"queries"},
				},
			}
			if item.MaxQueries > 0 {
				tool.Parameters["properties"].(map[string]any)["queries"].(map[string]any)["maxItems"] = item.MaxQueries
			}
			tools = append(tools, tool)
		case DynamicInterfaceAction:
			tool := LLMToolDefinition{
				Name:        item.ID,
				Description: firstNonEmpty(desc, "Invoke an external action for this turn"),
				Invocation:  LLMToolInvocationAction,
				ActionID:    item.ID,
				Parameters: map[string]any{
					"type":                 "object",
					"additionalProperties": true,
				},
			}
			if len(item.ArgsSchema) > 0 {
				tool.Parameters = item.ArgsSchema
			}
			tools = append(tools, tool)
		}
	}
	return tools
}

func buildDynamicInterfacePromptBlock(dynamicInterfaces []DynamicInterface) string {
	if len(dynamicInterfaces) == 0 {
		return ""
	}

	var dataLines []string
	var actionLines []string
	for _, item := range dynamicInterfaces {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		desc := strings.TrimSpace(item.Description)
		suffix := ""
		if desc != "" {
			suffix = ": " + desc
		}
		switch item.Kind {
		case DynamicInterfaceDataRequest:
			line := fmt.Sprintf("- %s%s", item.ID, suffix)
			if len(item.QueryTypes) > 0 {
				line += fmt.Sprintf(" (query_types: %s)", strings.Join(item.QueryTypes, ", "))
			}
			if item.MaxQueries > 0 {
				line += fmt.Sprintf(" (max_queries: %d)", item.MaxQueries)
			}
			dataLines = append(dataLines, line)
		case DynamicInterfaceAction:
			line := fmt.Sprintf("- %s%s", item.ID, suffix)
			if item.MaxCalls > 0 {
				line += fmt.Sprintf(" (max_calls: %d)", item.MaxCalls)
			}
			actionLines = append(actionLines, line)
		}
	}

	if len(dataLines) == 0 && len(actionLines) == 0 {
		return ""
	}

	sb := &strings.Builder{}
	sb.WriteString("\n\n========== Dynamic Interfaces ==========")
	sb.WriteString("\nOnly use interfaces listed in this block for the current request.")
	sb.WriteString("\nPrefer the smallest sufficient query or action. Do not invent interfaces that are not listed.")
	if len(dataLines) > 0 {
		sb.WriteString("\n\nData request interfaces:\n")
		sb.WriteString(strings.Join(dataLines, "\n"))
	}
	if len(actionLines) > 0 {
		sb.WriteString("\n\nAction interfaces:\n")
		sb.WriteString(strings.Join(actionLines, "\n"))
	}
	return sb.String()
}

func appendDynamicInterfaceContext(systemContext string, dynamicInterfaces []DynamicInterface) string {
	block := buildDynamicInterfacePromptBlock(dynamicInterfaces)
	if strings.TrimSpace(block) == "" {
		return systemContext
	}
	if strings.TrimSpace(systemContext) == "" {
		return strings.TrimSpace(block)
	}
	return strings.TrimSpace(systemContext + block)
}

func buildPlayerIntentPrompt(systemContext string, nodeID string, interaction *InteractionContext) string {
	targetNodeID := ""
	sceneNodeID := ""
	mode := sdk.InteractionModeDirectDialogue
	if interaction != nil {
		targetNodeID = strings.TrimSpace(interaction.TargetNodeID)
		sceneNodeID = strings.TrimSpace(interaction.SceneNodeID)
		if strings.TrimSpace(interaction.Mode) != "" {
			mode = strings.TrimSpace(interaction.Mode)
		}
	}
	return fmt.Sprintf(`你现在不是在直接扮演 NPC 回话，而是在为玩家节点生成“行为意图提案”。
你必须把玩家输入理解为待验证的主张，而不是已成立事实。玩家自然语言里提到的地点、物品、金钱、任务进度、场景占用、即时状态，都只有在上下文明确给出或后续由权威接口验证后才能当真。

你的任务：
1. 将玩家输入整理成结构化 player_intent 提案。
2. 必要时用 request_data 请求权威数据来补事实。
3. 给出 missing_facts，指出哪些关键事实尚未验证。
4. 给出 suggested_interaction，说明后续更适合走哪类 interaction。
5. 不要把提案描述成“已经执行成功”的结果。

输出必须是单个 JSON 对象，不能包含额外文本。

允许输出字段：
- reply: 可选，给调用方的人类可读摘要，应该简短。
- request_data: 可选。当缺少权威事实时，用它请求游戏侧或引擎侧数据。
- player_intent: 必填。结构如下：
{
  "intent": {
    "type": "%s",
    "actor_node_id": "%s",
    "scene_node_id": "%s",
    "target_node_id": "%s",
    "summary": "一句话总结玩家意图",
    "risk_level": "%s",
    "confidence": 0.0,
    "steps": [
      {
        "type": "%s",
        "target_node_id": "可选",
        "scene_node_id": "可选",
        "item_id": "可选",
        "content": "speech 时填写",
        "args": {},
        "preconditions": [
          {
            "type": "%s",
            "actor_node_id": "%s",
            "target_node_id": "可选",
            "scene_node_id": "可选",
            "item_id": "可选",
            "task_id": "可选",
            "expected": "可选",
            "args": {}
          }
        ]
      }
    ]
  },
  "missing_facts": [
    {
      "type": "%s",
      "node_id": "可选",
      "item_id": "可选",
      "task_id": "可选",
      "reason": "为什么缺这个事实"
    }
  ],
  "suggested_interaction": {
    "mode": "%s",
    "event_type": "%s",
    "audience_scope": "%s",
    "target_node_id": "可选"
  }
}

request_data 格式：
{"label":"查询标签","target":"store","queries":[{"type":"node_components","node_id":"节点ID","filter":"","limit":10}]}

推理约束：
- 如果玩家说“我把沾血的短刀拍在柜台上”，但上下文没有确认玩家有短刀、也没有确认玩家在柜台附近，就必须把它拆成待验证步骤，并补 preconditions 或 missing_facts。
- 如果玩家输入本质上是“对 NPC 说一句话”，可以把语言内容放入 speech step 的 content 中。
- 如果一句输入包含多段动作，优先使用 composite，把动作序列拆开。
- 如果目标不明确，可以保留空 target_node_id，但要在 suggested_interaction 或 missing_facts 中说明。
- suggested_interaction 只描述后续适合触发哪类 NPC/群聊交互；像 move、inspect、use_item 这类纯玩家侧步骤不要伪装成 interaction event_type。
- 只有在确实不需要更多权威数据时，才省略 request_data。

当前交互参考：
- actor_node_id=%s
- interaction_mode=%s
- target_node_id=%s
- scene_node_id=%s

========== 上下文 ==========
%s
`,
		strings.Join(sdk.ValidPlayerIntentTypes(), "|"),
		nodeID,
		sceneNodeID,
		targetNodeID,
		strings.Join(sdk.ValidPlayerIntentRiskLevels(), "|"),
		strings.Join(sdk.ValidPlayerIntentStepTypes(), "|"),
		strings.Join(sdk.ValidPlayerIntentPreconditionTypes(), "|"),
		nodeID,
		strings.Join(sdk.ValidMissingFactTypes(), "|"),
		strings.Join([]string{sdk.InteractionModeDirectDialogue, sdk.InteractionModeGroupChat, sdk.InteractionModeGiftResponse, sdk.InteractionModeTradeDialogue}, "|"),
		strings.Join([]string{sdk.InteractionEventSpeech, sdk.InteractionEventShowItem, sdk.InteractionEventGift, sdk.InteractionEventTradeRequest, sdk.InteractionEventThreaten}, "|"),
		strings.Join([]string{sdk.InteractionAudiencePrivate, sdk.InteractionAudiencePublic, sdk.InteractionAudienceWhisper}, "|"),
		nodeID,
		mode,
		targetNodeID,
		sceneNodeID,
		systemContext,
	)
}

func buildDialoguePrompt(systemContext string, nodeID string) string {
	return fmt.Sprintf(`你是一个游戏 Agent 系统中的 NPC 角色扮演引擎。
重要：如果你需要查询更多数据来进行回复，可以在输出中包含 request_data 字段。例如：需要查询某个节点的记忆或组件时，可以输出 request_data 让引擎自动加载。
你的输出必须是以下 JSON 格式，不能包含其他任何文字：

{"reply":"NPC 的对话回复","action_calls":[{"action_id":"add_memory","args":{"node_id":"%s","content":"记忆内容","level":"short_term"}}],"memory_updates":[{"node_id":"%s","content":"短期记忆","level":"short_term"}]}

可用 action_id: add_memory, update_mood, send_dialogue, adjust_relation, spawn_item
memory_updates 的 level 可选: short_term, long_term, shared, world

request_data 格式（可选，需要时添加）：
{"label":"查询标签","target":"store","queries":[{"type":"node_components","node_id":"节点ID","filter":"","limit":10}]}

========== NPC 角色设定 ==========
%s
`, nodeID, nodeID, systemContext)
}

func buildInteractionDialoguePrompt(systemContext string, nodeID string, interaction *InteractionContext) string {
	mode := "direct_dialogue"
	if interaction != nil && strings.TrimSpace(interaction.Mode) != "" {
		mode = strings.TrimSpace(interaction.Mode)
	}
	modeInstructions := buildDialogueModeInstruction(mode, interaction)
	return fmt.Sprintf(`你是游戏 Agent 系统中的 NPC 角色扮演引擎。你必须只根据当前上下文中已经确认的事实回应，不能把玩家自然语言中的未证实主张直接当成事实。
如果你需要查询更多权威数据来进行回复，可以在输出中包含 request_data。你的输出必须是单个 JSON 对象，不能包含任何额外文字。

输出格式：
{"reply":"NPC 的对话回复","action_calls":[{"action_id":"add_memory","args":{"node_id":"%s","content":"记忆内容","level":"short_term"}}],"memory_updates":[{"node_id":"%s","content":"短期记忆","level":"short_term"}]}

可用 action_id: add_memory, update_mood, send_dialogue, adjust_relation, spawn_item
memory_updates 的 level 可选: short_term, long_term, shared, world

request_data 格式（可选，需要时添加）：
{"label":"查询标签","target":"store","queries":[{"type":"node_components","node_id":"节点ID","filter":"","limit":10}]}

交互模式约束：
%s

========== NPC 角色设定 ==========
%s
`, nodeID, nodeID, modeInstructions, systemContext)
}

func buildDialogueModeInstruction(mode string, interaction *InteractionContext) string {
	eventType := ""
	if interaction != nil && interaction.Event != nil {
		eventType = strings.TrimSpace(interaction.Event.Type)
	}
	switch mode {
	case "group_chat":
		return strings.Join([]string{
			"- 当前是群聊/房间对话。你只代表当前目标 NPC 发言，不要替其他 NPC 或玩家发言。",
			"- 默认把对话视为公开场景，可被在场参与者听见；回复时要考虑旁听者存在。",
			"- 只需给出当前目标 NPC 的单轮回应，不要模拟整段群体连锁对话。",
		}, "\n")
	case "gift_response":
		parts := []string{
			"- 当前交互是送礼反馈。重点回应礼物、送礼者意图、关系变化和即时态度。",
			"- 如果礼物属性、归属或可交付性没有权威事实支持，优先 request_data，不要编造。",
		}
		if interaction != nil && interaction.Event != nil && strings.TrimSpace(interaction.Event.ItemID) != "" {
			parts = append(parts, fmt.Sprintf("- 当前事件 item_id=%s，可围绕它表达态度和后续动作倾向。", interaction.Event.ItemID))
		}
		return strings.Join(parts, "\n")
	case "trade_dialogue":
		return strings.Join([]string{
			"- 当前交互是交易/议价对话。回复应围绕价格、条件、库存、金钱、信誉与是否愿意成交。",
			"- 如果库存、金钱、背包或物品归属需要确认，优先 request_data。",
			"- 不要把交易直接说成已经完成，除非权威状态或动作调用已经支持。",
		}, "\n")
	default:
		parts := []string{
			"- 当前交互是点对点对话。你只代表当前目标 NPC 回复当前说话者。",
			"- 对自然语言中的动作性主张保持事实约束：未被上下文或权威数据证实的地点、物品、状态、任务进度都不能视为既成事实。",
			"- 缺少关键事实时优先 request_data，而不是脑补。",
		}
		switch eventType {
		case "show_item":
			parts = append(parts,
				"- 当前事件是展示物品。你应优先围绕该物品的识别、态度、记忆、风险和线索价值来回应。",
				"- 如果并不确定该物品是否真的存在、是否属于玩家、是否与你当前情境相关，优先 request_data。",
			)
		case "threaten":
			parts = append(parts,
				"- 当前事件是威胁施压。你应结合关系、场景安全、身份立场与即时风险做出克制且符合角色的反应。",
				"- 不要无依据地假设玩家已经成功控制局面；缺少武器、人数、位置或支援事实时优先 request_data。",
			)
		}
		return strings.Join(parts, "\n")
	}
}

type resourceState struct {
	Food     int `json:"food,omitempty"`
	Order    int `json:"order,omitempty"`
	Defense  int `json:"defense,omitempty"`
	Morale   int `json:"morale,omitempty"`
	Treasury int `json:"treasury,omitempty"`
}

// buildWorldTickPrompt 构建世界刻推进任务的系统提示词。
// relationSummary 必须是高价值、低噪音的结构化摘要块，不能退回成全量关系转储。
func buildWorldTickPrompt(systemContext string, outline string, continuityBlocks []string, recentTimeline []string, worldTimeBlock string, relationSummary string) string {
	sb := &strings.Builder{}
	sb.WriteString(systemContext)
	sb.WriteString("\n\n你正在推进世界时间线。")
	if strings.TrimSpace(relationSummary) != "" {
		sb.WriteString("\n\n当前 scope 的高价值关系摘要：\n")
		sb.WriteString(relationSummary)
		sb.WriteString("\n以上摘要只用于把握当前 tick 最关键的归属、控制与位置结构，不代表允许你无边界展开整张关系图。")
	}
	if strings.TrimSpace(worldTimeBlock) != "" {
		sb.WriteString("\n\n世界时间约束与当前时间：\n")
		sb.WriteString(worldTimeBlock)
		sb.WriteString("\n以上时间规则是引擎硬约束，不能忽略、改写或跳过。")
	}
	if len(continuityBlocks) > 0 {
		sb.WriteString("\n\n必须延续的世界状态与剧情状态：\n")
		sb.WriteString(strings.Join(continuityBlocks, "\n"))
		sb.WriteString("\n这些事实默认仍然成立，除非当前 tick 明确给出新的因果变化；不要无故重置、遗忘或改写其中的关键地点、设施、人物关系和数值趋势。")
	}
	if len(recentTimeline) > 0 {
		sb.WriteString("\n\n最近时间线记录：\n")
		sb.WriteString(strings.Join(recentTimeline, "\n"))
	}

	if outline != "" {
		sb.WriteString("\n\n现有世界线大纲：\n")
		sb.WriteString(outline)
		sb.WriteString("\n\n请将上述大纲细化到当前时间刻度，同时更新未来大纲。")
	} else {
		sb.WriteString("\n请先生成一个未来数个回合的世界线大纲，再细化当前刻度的事件。")
	}

	sb.WriteString("\n\n如果你需要查询更多数据来细化推演，可以在输出中包含 request_data 字段。")
	sb.WriteString("\n引擎会自动加载你请求的数据并让你继续。")
	sb.WriteString("\n如果时间模式是 flexible，你需要在输出中明确 advanced_ticks，表示本次实际推进了多少个基础 tick。")
	sb.WriteString("\n如果时间模式是 fixed，advanced_ticks 只能等于 1。")
	sb.WriteString("\n\n输出 JSON 格式：\n")
	sb.WriteString(`{"reply":"","advanced_ticks":1,"action_calls":[],"memory_updates":[],"world_change_plan":{},"future_outline":"未来数个回合的大纲"}`)
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

	sb.WriteString("\n\n输出 JSON 格式：\n")
	sb.WriteString(`{"reply":"事件影响评估","action_calls":[],"memory_updates":[{"node_id":"` + nodeID + `","content":"共享记忆","level":"shared"}],"world_change_plan":{"impact_level":"major","summary":"影响摘要","world_events":[{"event_type":"diplomatic_shift","scope":"region","description":"事件描述","confidence":0.85}],"proposed_actions":[{"api_name":"adjust_relation","args":{"delta":-20}}]}}`)
	return sb.String()
}

// buildAutonomousPrompt 构建自主行为任务的系统提示词。
func buildAutonomousPrompt(systemContext string, nodeID string, cfg *AutonomousConfig) string {
	capabilities, _ := json.MarshalIndent(cfg.Capabilities, "", "  ")
	return fmt.Sprintf(`你是游戏 Agent 引擎中的自主行为决策器。
你正在为节点 %s 执行一次自主行为周期。你只能使用下面 capabilities 中显式列出的能力；未列出的 action_id 绝对不能输出。什么都不做是有效选择。如果没有明确必要行动，请返回空 action_calls。
重要：如果你需要查询更多数据来决策，可以在输出中包含 request_data 字段。引擎会自动加载你请求的数据并让你继续。
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
