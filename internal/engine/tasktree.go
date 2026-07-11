// Package engine 提供任务节点树（DAG）：为管线内部多轮 LLM 交互提供数据装配与继承机制。
package engine

import (
	"fmt"
	"sync/atomic"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// TaskNode 是推理 DAG 上的一个节点。
// 支持多父节点（DAG），允许多个分支共享同一推理结果。
type TaskNode struct {
	ID       string      `json:"id"`
	Label    string      `json:"label"`
	Parents  []*TaskNode `json:"-"`          // 父节点列表（DAG 支持多父）
	Children []*TaskNode `json:"children,omitempty"`

	GameNodeID string                 `json:"game_node_id,omitempty"`
	Components []store.ComponentModel `json:"-"`
	Memories   []store.MemoryModel    `json:"-"`

	Round       int    `json:"round,omitempty"`
	Prompt      string `json:"prompt,omitempty"`
	LLMResponse string `json:"llm_response,omitempty"`
	Analysis    string `json:"analysis,omitempty"`
	Decision    string `json:"decision,omitempty"`
}

// Context 递归向上收集所有祖先路径的上下文（DAG 会合并多个父路径）。
func (n *TaskNode) Context(depth int) string {
	if depth <= 0 || len(n.Parents) == 0 {
		return n.Analysis
	}
	var parts []string
	if n.Analysis != "" {
		parts = append(parts, n.Analysis)
	}
	for _, p := range n.Parents {
		if c := p.Context(depth - 1); c != "" {
			parts = append(parts, c)
		}
	}
	return strings.Join(parts, "\n")
}

var taskNodeSeq atomic.Int64

func nextNodeID() string {
	n := taskNodeSeq.Add(1)
	return fmt.Sprintf("n%02d", n)
}

// NewTaskNode 创建一个节点，可选绑定游戏节点 ID 自动加载数据。
func NewTaskNode(label, gameNodeID string) *TaskNode {
	node := &TaskNode{
		ID:         nextNodeID(),
		Label:      label,
		GameNodeID: gameNodeID,
		Parents:    make([]*TaskNode, 0, 2),
	}
	if gameNodeID != "" {
		if comps, err := store.GetNodeComponents(gameNodeID); err == nil {
			node.Components = comps
		}
		if mems, err := store.GetNodeMemories(gameNodeID, 30); err == nil {
			node.Memories = mems
		}
	}
	return node
}

// AddChild 创建子节点并建立父子关系（单父，兼容旧用法）。
func (n *TaskNode) AddChild(label, gameNodeID string) *TaskNode {
	child := NewTaskNode(label, gameNodeID)
	child.Parents = append(child.Parents, n)
	n.Children = append(n.Children, child)
	return child
}

// LinkChild 将已存在的节点链接为子节点，建立 DAG 多父关系。
// 如果 child 已经是 n 的子节点则不做任何事。
func (n *TaskNode) LinkChild(child *TaskNode) {
	for _, existing := range n.Children {
		if existing == child {
			return
		}
	}
	for _, existingParent := range child.Parents {
		if existingParent == n {
			return
		}
	}
	child.Parents = append(child.Parents, n)
	n.Children = append(n.Children, child)
}

// WalkDAG 先根遍历 DAG，已访问的节点不会重复遍历。
func (n *TaskNode) WalkDAG(fn func(*TaskNode)) {
	visited := map[*TaskNode]bool{}
	var walk func(node *TaskNode)
	walk = func(node *TaskNode) {
		if visited[node] {
			return
		}
		visited[node] = true
		fn(node)
		for _, c := range node.Children {
			walk(c)
		}
	}
	walk(n)
}

// Walk 保持向后兼容的单父遍历（只走第一条父路径的子节点）。
func (n *TaskNode) Walk(fn func(*TaskNode)) {
	fn(n)
	for _, c := range n.Children {
		c.Walk(fn)
	}
}

// TaskTree 为一次推理请求构建的临时 DAG。
type TaskTree struct {
	Root     *TaskNode `json:"root"`
	TaskType TaskType  `json:"task_type"`
	WorldID  string    `json:"world_id"`
	NodeID   string    `json:"node_id"`

	allNodes   map[string]*TaskNode // id -> node，方便按 ID 引用
	roundNodes []*TaskNode          // 按顺序记录每轮的节点
}

// NewTaskTree 创建推理 DAG。
func NewTaskTree(taskType TaskType, worldID, nodeID string) *TaskTree {
	tree := &TaskTree{
		TaskType:   taskType,
		WorldID:    worldID,
		NodeID:     nodeID,
		allNodes:   make(map[string]*TaskNode),
		roundNodes: make([]*TaskNode, 0, 8),
	}
	tree.Root = NewTaskNode("task_root", nodeID)
	tree.allNodes[tree.Root.ID] = tree.Root
	return tree
}

// NewRound 创建新一轮的推理节点，从当前最新轮次或根节点派生。
func (t *TaskTree) NewRound(label string) *TaskNode {
	parent := t.CurrentRound()
	if parent == nil {
		parent = t.Root
	}
	node := parent.AddChild(label, "")
	node.Round = len(t.roundNodes)
	t.allNodes[node.ID] = node
	t.roundNodes = append(t.roundNodes, node)
	return node
}

// NewRoundFrom 从指定的父节点创建新轮次节点，实现 DAG 多父汇聚。
// 例如：两个不同分支的结果汇聚到同一轮分析中。
func (t *TaskTree) NewRoundFrom(label string, parents ...*TaskNode) *TaskNode {
	node := NewTaskNode(label, "")
	for _, p := range parents {
		p.LinkChild(node)
	}
	node.Round = len(t.roundNodes)
	t.allNodes[node.ID] = node
	t.roundNodes = append(t.roundNodes, node)
	return node
}

// CurrentRound 返回当前轮次的节点（最后一轮）。
func (t *TaskTree) CurrentRound() *TaskNode {
	if len(t.roundNodes) == 0 {
		return nil
	}
	return t.roundNodes[len(t.roundNodes)-1]
}

// FindNode 按标签查找最新轮次的节点。
func (t *TaskTree) FindNode(label string) *TaskNode {
	for i := len(t.roundNodes) - 1; i >= 0; i-- {
		if t.roundNodes[i].Label == label {
			return t.roundNodes[i]
		}
	}
	return nil
}

// BuildLLMContext 从根节点 DAG 遍历，构建供 LLM 消费的完整上下文。
func (t *TaskTree) BuildLLMContext() string {
	var parts []string
	t.Root.WalkDAG(func(n *TaskNode) {
		line := fmt.Sprintf("[%s] %s", n.Label, n.ID)
		parts = append(parts, line)

		for _, c := range n.Components {
			parts = append(parts, fmt.Sprintf("  【%s】%s", c.ComponentType, c.Data))
		}
		for _, m := range n.Memories {
			parts = append(parts, fmt.Sprintf("  [记忆:%s] %s", m.Level, m.Content))
		}
		if n.Analysis != "" {
			parts = append(parts, fmt.Sprintf("  [分析] %s", n.Analysis))
		}
		if n.Decision != "" {
			parts = append(parts, fmt.Sprintf("  [决策] %s", n.Decision))
		}
		if n.LLMResponse != "" && n.Label != "analysis" && n.Label != "task_root" {
			parts = append(parts, fmt.Sprintf("  [第%d轮回复] %s", n.Round+1, truncateForContext(n.LLMResponse, 200)))
		}
	})
	return strings.Join(parts, "\n")
}

// RoundCount 返回已完成的总轮数。
func (t *TaskTree) RoundCount() int {
	return len(t.roundNodes)
}

// AllNodes 返回 DAG 中所有节点的映射。
func (t *TaskTree) AllNodes() map[string]*TaskNode {
	return t.allNodes
}

func truncateForContext(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
