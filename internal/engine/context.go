// Package engine 定义核心数据模型、上下文构建器与推理管线。
package engine

import (
	"fmt"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type ContextBuilder struct{}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{}
}

type BuiltContext struct {
	Node                 *store.NodeModel       `json:"node"`
	Components           []store.ComponentModel `json:"components"`
	Memories             []store.MemoryModel    `json:"memories"`
	Relations            []store.RelationModel  `json:"relations"`
	Children             []store.NodeModel      `json:"children"`
	Ancestors            []store.NodeModel      `json:"ancestors"`
	IdentityAncestors    []store.NodeModel      `json:"identity_ancestors,omitempty"`
	EnvironmentNode      *store.NodeModel       `json:"environment_node,omitempty"`
	EnvironmentAncestors []store.NodeModel      `json:"environment_ancestors,omitempty"`
	StateBlocks          []string               `json:"state_blocks,omitempty"`
	SystemPrompt         string                 `json:"system_prompt"`
}

// Build 根据任务焦点节点装配本轮推理的基础上下文。
//
// 设计约束：
//  1. 默认上下文应区分“稳定身份/归属链”和“动态环境链”。
//  2. 稳定身份/归属链默认由 parent 承担；组织/控制关系可作为补充，但不能替代主父链。
//  3. 动态环境链默认应由 located_at -> location -> location ancestors 装配；禁止再把当前位置硬塞回 parent。
//  4. includeRelatedNodes 只是受控的关系补充开关，不是“把所有关系边另一端数据全部塞进 prompt”的许可。
//  5. BuiltContext.Relations 保留结构化关系数据，供后续按任务类型进一步筛选和扩图；SystemPrompt 不应依赖无差别
//     关系拼接来模拟图谱。
func (b *ContextBuilder) Build(taskType TaskType, nodeID string, depth int, memoryLimit int, includeRelated bool) (*BuiltContext, error) {
	node, err := store.GetNode(nodeID)
	if err != nil {
		return nil, fmt.Errorf("get node %s: %w", nodeID, err)
	}
	comps, _ := store.GetNodeComponents(nodeID)
	memLimit := memoryLimit
	if memLimit <= 0 {
		memLimit = 50
	}
	mems, _ := store.GetNodeMemories(nodeID, memLimit)
	rels, _ := store.GetNodeRelations(nodeID)
	children, _ := store.GetChildNodes(nodeID)

	identityAncestors := b.collectAncestors(node, depth)
	environmentNode := b.resolveEnvironmentNode(taskType, node, rels)
	environmentAncestors := b.collectEnvironmentAncestors(environmentNode, depth)

	comps = append(comps, b.collectComponents(identityAncestors)...)
	mems = append(mems, b.collectMemories(identityAncestors, b.relatedMemoryLimit(memLimit))...)
	if environmentNode != nil {
		comps = append(comps, b.collectComponents([]store.NodeModel{*environmentNode})...)
		mems = append(mems, b.collectMemories([]store.NodeModel{*environmentNode}, b.relatedMemoryLimit(memLimit))...)
	}
	comps = append(comps, b.collectComponents(environmentAncestors)...)
	mems = append(mems, b.collectMemories(environmentAncestors, b.relatedMemoryLimit(memLimit))...)

	if includeRelated {
		for _, r := range rels {
			if !shouldIncludeRelatedRelation(taskType, nodeID, r) {
				continue
			}
			relatedUUID := r.SourceUUID
			if relatedUUID == nodeID {
				relatedUUID = r.TargetUUID
			}
			if rc, err := store.GetNodeComponents(relatedUUID); err == nil {
				for _, c := range rc {
					comps = append(comps, c)
				}
			}
			if rm, err := store.GetNodeMemories(relatedUUID, 5); err == nil {
				mems = append(mems, rm...)
			}
		}
	}

	comps = dedupeComponents(comps)
	mems = dedupeMemories(mems)

	stateBlocks := b.buildStateBlocks(node.UUID)
	sysPrompt := b.buildSystemPrompt(node, comps, mems, identityAncestors, environmentNode, environmentAncestors, stateBlocks)
	return &BuiltContext{
		Node:                 node,
		Components:           comps,
		Memories:             mems,
		Relations:            rels,
		Children:             children,
		Ancestors:            identityAncestors,
		IdentityAncestors:    identityAncestors,
		EnvironmentNode:      environmentNode,
		EnvironmentAncestors: environmentAncestors,
		StateBlocks:          stateBlocks,
		SystemPrompt:         sysPrompt,
	}, nil
}

func (b *ContextBuilder) relatedMemoryLimit(memoryLimit int) int {
	if memoryLimit <= 0 {
		return 5
	}
	if memoryLimit < 5 {
		return memoryLimit
	}
	if memoryLimit > 10 {
		return 10
	}
	return memoryLimit
}

func (b *ContextBuilder) collectComponents(nodes []store.NodeModel) []store.ComponentModel {
	var comps []store.ComponentModel
	for _, n := range nodes {
		if ac, err := store.GetNodeComponents(n.UUID); err == nil {
			comps = append(comps, ac...)
		}
	}
	return comps
}

func (b *ContextBuilder) collectMemories(nodes []store.NodeModel, limit int) []store.MemoryModel {
	if limit <= 0 {
		return nil
	}
	var mems []store.MemoryModel
	for _, n := range nodes {
		if rm, err := store.GetNodeMemories(n.UUID, limit); err == nil {
			mems = append(mems, rm...)
		}
	}
	return mems
}

func shouldIncludeRelatedRelation(taskType TaskType, nodeID string, rel store.RelationModel) bool {
	if rel.SourceUUID == nodeID && rel.RelationType == string(RelExternalParent) {
		return false
	}
	switch taskType {
	case TaskNPCDialogue, TaskAutonomousAct, TaskCustom:
		return rel.RelationType == string(RelLocatedAt) || rel.RelationType == string(RelBelongsTo) || rel.RelationType == string(RelSubordinate)
	case TaskWorldEvent:
		return rel.RelationType == string(RelLocatedAt) || rel.RelationType == string(RelBelongsTo) || rel.RelationType == string(RelSubordinate)
	case TaskWorldTick:
		return rel.RelationType == string(RelBelongsTo) || rel.RelationType == string(RelSubordinate)
	default:
		return false
	}
}

func (b *ContextBuilder) resolveEnvironmentNode(taskType TaskType, node *store.NodeModel, rels []store.RelationModel) *store.NodeModel {
	if node == nil {
		return nil
	}
	switch taskType {
	case TaskNPCDialogue, TaskAutonomousAct, TaskCustom:
		return b.resolveLocatedAtTarget(node.UUID, rels)
	case TaskWorldEvent:
		if node.NodeType == string(NodeTypeLocation) || node.NodeType == string(NodeTypeWorld) {
			return node
		}
		return b.resolveLocatedAtTarget(node.UUID, rels)
	default:
		if node.NodeType == string(NodeTypeLocation) {
			return node
		}
		return nil
	}
}

func (b *ContextBuilder) resolveLocatedAtTarget(nodeID string, rels []store.RelationModel) *store.NodeModel {
	for _, rel := range rels {
		if rel.SourceUUID != nodeID || rel.RelationType != string(RelLocatedAt) {
			continue
		}
		target, err := store.GetNode(rel.TargetUUID)
		if err != nil {
			continue
		}
		return target
	}
	return nil
}

func (b *ContextBuilder) collectEnvironmentAncestors(environmentNode *store.NodeModel, maxDepth int) []store.NodeModel {
	if environmentNode == nil {
		return nil
	}
	return b.collectAncestors(environmentNode, maxDepth)
}

func (b *ContextBuilder) collectAncestors(node *store.NodeModel, maxDepth int) []store.NodeModel {
	var ancestors []store.NodeModel
	current := node
	for i := 0; i < maxDepth && current.ParentID != nil; i++ {
		if current.ParentUUID == nil {
			break
		}
		parent, err := store.GetNode(*current.ParentUUID)
		if err != nil {
			break
		}
		ancestors = append(ancestors, *parent)
		current = parent
	}
	return ancestors
}

// buildSystemPrompt 将已装配的上下文压平成当前 prompt 文本。
//
// 当前实现仍以文本 prompt 为主，但后续扩展必须保持以下边界：
// 1. parent 祖先链表达稳定归属，而不是当前环境。
// 2. location 语义应来自 located_at 装配出的环境链，而不是靠上层调用方通过改 parent 模拟。
// 3. 社会语义关系不应在这里被无差别展开；它们应由任务特定的关系子图选择器控制是否进入 prompt。
func (b *ContextBuilder) buildSystemPrompt(node *store.NodeModel, comps []store.ComponentModel, mems []store.MemoryModel, identityAncestors []store.NodeModel, environmentNode *store.NodeModel, environmentAncestors []store.NodeModel, stateBlocks []string) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("你是 %s（%s）。", node.Name, node.NodeType))

	for _, c := range comps {
		parts = append(parts, fmt.Sprintf("【%s】%s", c.ComponentType, c.Data))
	}

	if len(identityAncestors) > 0 {
		var ap []string
		for _, a := range identityAncestors {
			ap = append(ap, fmt.Sprintf("%s(%s)", a.Name, a.NodeType))
		}
		parts = append(parts, fmt.Sprintf("稳定归属链：%s", strings.Join(ap, " > ")))
	}

	if environmentNode != nil {
		var ep []string
		ep = append(ep, fmt.Sprintf("%s(%s)", environmentNode.Name, environmentNode.NodeType))
		for _, a := range environmentAncestors {
			ep = append(ep, fmt.Sprintf("%s(%s)", a.Name, a.NodeType))
		}
		parts = append(parts, fmt.Sprintf("当前环境链：%s", strings.Join(ep, " > ")))
		parts = append(parts, b.buildEnvironmentPromptBlock(environmentNode, environmentAncestors)...)
	}

	if len(mems) > 0 {
		parts = append(parts, "记忆：")
		for _, m := range mems {
			parts = append(parts, fmt.Sprintf("  [%s] %s", m.Level, m.Content))
		}
	}
	if len(stateBlocks) > 0 {
		parts = append(parts, "连续性状态：")
		parts = append(parts, stateBlocks...)
	}
	return strings.Join(parts, "\n")
}

func (b *ContextBuilder) buildEnvironmentPromptBlock(environmentNode *store.NodeModel, environmentAncestors []store.NodeModel) []string {
	if environmentNode == nil {
		return nil
	}
	nodes := []store.NodeModel{*environmentNode}
	nodes = append(nodes, environmentAncestors...)
	var parts []string
	parts = append(parts, "环境信息：")
	for _, n := range dedupeNodes(nodes) {
		parts = append(parts, fmt.Sprintf("  [环境节点] %s(%s)", n.Name, n.NodeType))
		if comps, err := store.GetNodeComponents(n.UUID); err == nil {
			for _, comp := range comps {
				parts = append(parts, fmt.Sprintf("    【%s/%s】%s", n.Name, comp.ComponentType, comp.Data))
			}
		}
		if mems, err := store.GetNodeMemories(n.UUID, 5); err == nil {
			for _, mem := range mems {
				parts = append(parts, fmt.Sprintf("    [环境记忆:%s] %s", mem.Level, mem.Content))
			}
		}
	}
	return parts
}

func dedupeComponents(input []store.ComponentModel) []store.ComponentModel {
	seen := map[string]bool{}
	out := make([]store.ComponentModel, 0, len(input))
	for _, item := range input {
		key := item.UUID
		if key == "" {
			key = fmt.Sprintf("%s:%s:%s", item.NodeUUID, item.ComponentType, item.Data)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}

func dedupeMemories(input []store.MemoryModel) []store.MemoryModel {
	seen := map[string]bool{}
	out := make([]store.MemoryModel, 0, len(input))
	for _, item := range input {
		key := item.UUID
		if key == "" {
			key = fmt.Sprintf("%s:%s:%s:%s", item.NodeUUID, item.Level, item.Tags, item.Content)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}

func dedupeNodes(input []store.NodeModel) []store.NodeModel {
	seen := map[string]bool{}
	out := make([]store.NodeModel, 0, len(input))
	for _, item := range input {
		if item.UUID == "" || seen[item.UUID] {
			continue
		}
		seen[item.UUID] = true
		out = append(out, item)
	}
	return out
}

func (b *ContextBuilder) buildStateBlocks(nodeID string) []string {
	componentTypes := []string{string(CompWorldState), string(CompStoryState), string(CompStoryHistory), string(CompTickPolicy), string(CompWorldTimeState), string(CompStateSnapshot)}
	blocks := make([]string, 0, len(componentTypes))
	for _, componentType := range componentTypes {
		components, err := store.GetComponentsByType(nodeID, componentType)
		if err != nil || len(components) == 0 {
			continue
		}
		for _, comp := range components {
			blocks = append(blocks, fmt.Sprintf("【%s】%s", comp.ComponentType, comp.Data))
		}
	}
	return blocks
}
