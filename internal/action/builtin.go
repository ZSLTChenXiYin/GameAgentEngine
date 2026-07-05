// Package action 提供动作接口定义和内置动作实现。
package action

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// UpdateMood 是一个同步动作，用于更新角色的情绪状态。
type UpdateMood struct{}

// ID 返回动作标识。
func (a *UpdateMood) ID() string { return "update_mood" }

// Validate 校验更新情绪所需的关键参数是否存在。
func (a *UpdateMood) Validate(args map[string]any) error {
	if _, ok := args["mood"]; !ok {
		return fmt.Errorf("mood required")
	}
	return nil
}

// Execute 执行情绪更新，并同步写入角色记忆。
func (a *UpdateMood) Execute(args map[string]any) (any, error) {
	nodeID, _ := args["node_id"].(string)
	mood, _ := args["mood"].(string)
	intensity, _ := args["intensity"].(float64)

	if nodeID == "" || mood == "" {
		return nil, fmt.Errorf("node_id and mood required")
	}

	comps, err := store.GetComponentsByType(nodeID, "profile")
	if err != nil || len(comps) == 0 {
		return nil, fmt.Errorf("profile not found for node %s", nodeID)
	}

	profile := make(map[string]any)
	json.Unmarshal([]byte(comps[0].Data), &profile)
	profile["mood"] = mood
	profile["mood_intensity"] = intensity

	data, _ := json.Marshal(profile)
	store.DB.Model(&comps[0]).Update("data", string(data))

	nodeIntID := store.ResolveNodeUUID(nodeID)
	store.CreateMemory(&store.MemoryModel{
		NodeID:  nodeIntID,
		Content: fmt.Sprintf("情绪变为: %s (强度: %.0f)", mood, intensity),
		Level:   "short_term",
	})

	return map[string]any{"mood": mood, "intensity": intensity}, nil
}

// AddMemory 是一个同步动作，用于直接写入记忆。
type AddMemory struct{}

// ID 返回动作标识。
func (a *AddMemory) ID() string { return "add_memory" }

// Validate 校验记忆内容是否存在。
func (a *AddMemory) Validate(args map[string]any) error {
	if _, ok := args["content"]; !ok {
		return fmt.Errorf("content required")
	}
	return nil
}

// Execute 创建一条新的记忆记录。
func (a *AddMemory) Execute(args map[string]any) (any, error) {
	nodeID, _ := args["node_id"].(string)
	content, _ := args["content"].(string)
	level, _ := args["level"].(string)

	if content == "" {
		return nil, fmt.Errorf("content required")
	}
	if nodeID == "" {
		return nil, fmt.Errorf("node_id required")
	}
	if level == "" {
		level = "short_term"
	}

	nodeIntID := store.ResolveNodeUUID(nodeID)
	m := &store.MemoryModel{
		NodeID:  nodeIntID,
		Content: content,
		Level:   level,
	}
	if tags, ok := args["tags"].(string); ok {
		m.Tags = tags
	}
	if err := store.CreateMemory(m); err != nil {
		return nil, err
	}

	log.Printf("[memory] %s -> %s: %s", nodeID, level, content)
	return map[string]any{"memory_id": m.ID}, nil
}

// AdjustRelation 是一个异步动作，用于请求调整关系权重。
type AdjustRelation struct{}

// ID 返回动作标识。
func (a *AdjustRelation) ID() string { return "adjust_relation" }

// Validate 预留参数校验入口。
func (a *AdjustRelation) Validate(args map[string]any) error {
	return nil
}

// OnResult 处理游戏侧上报的异步执行结果。
func (a *AdjustRelation) OnResult(callbackID string, status string, result any) error {
	log.Printf("[callback] adjust_relation %s: %s -> %v", callbackID, status, result)
	return nil
}

// SendDialogue 是一个同步动作，用于记录角色说出的文本。
type SendDialogue struct{}

// ID 返回动作标识。
func (a *SendDialogue) ID() string { return "send_dialogue" }

// Validate 预留参数校验入口。
func (a *SendDialogue) Validate(args map[string]any) error {
	return nil
}

// Execute 记录台词内容，并给角色写入一条短期记忆。
func (a *SendDialogue) Execute(args map[string]any) (any, error) {
	nodeID, _ := args["node_id"].(string)
	content, _ := args["content"].(string)
	if content == "" {
		return nil, fmt.Errorf("content required")
	}

	if nodeID != "" {
		nodeIntID := store.ResolveNodeUUID(nodeID)
		store.CreateMemory(&store.MemoryModel{
			NodeID:  nodeIntID,
			Content: "说出: " + content,
			Level:   "short_term",
		})
	}
	return map[string]any{"dialogue": content}, nil
}

// SpawnItem 是一个异步动作，用于请求游戏侧生成物品。
type SpawnItem struct{}

// ID 返回动作标识。
func (a *SpawnItem) ID() string { return "spawn_item" }

// Validate 预留参数校验入口。
func (a *SpawnItem) Validate(args map[string]any) error {
	return nil
}

// OnResult 处理物品生成动作的回调结果。
func (a *SpawnItem) OnResult(callbackID string, status string, result any) error {
	log.Printf("[callback] spawn_item %s: %s -> %v", callbackID, status, result)
	return nil
}
