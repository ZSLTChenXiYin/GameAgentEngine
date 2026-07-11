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

	// Batch-load components and memories for identity ancestors
	if ids := b.collectNodeIDs(identityAncestors); len(ids) > 0 {
		if compMap, err := store.GetComponentsByNodeIDs(ids); err == nil {
			for _, cc := range compMap {
				comps = append(comps, cc...)
			}
		}
		if memMap, err := store.GetMemoriesByNodeIDs(ids, b.relatedMemoryLimit(memLimit)); err == nil {
			for _, mm := range memMap {
				mems = append(mems, mm...)
			}
		}
	}
	if environmentNode != nil {
		envIDs := b.collectNodeIDs([]store.NodeModel{*environmentNode})
		if len(envIDs) > 0 {
			if compMap, err := store.GetComponentsByNodeIDs(envIDs); err == nil {
				for _, cc := range compMap {
					comps = append(comps, cc...)
				}
			}
			if memMap, err := store.GetMemoriesByNodeIDs(envIDs, b.relatedMemoryLimit(memLimit)); err == nil {
				for _, mm := range memMap {
					mems = append(mems, mm...)
				}
			}
		}
	}
	// Batch-load for environment ancestors
	if envIDs := b.collectNodeIDs(environmentAncestors); len(envIDs) > 0 {
		if compMap, err := store.GetComponentsByNodeIDs(envIDs); err == nil {
			for _, cc := range compMap {
				comps = append(comps, cc...)
			}
		}
		if memMap, err := store.GetMemoriesByNodeIDs(envIDs, b.relatedMemoryLimit(memLimit)); err == nil {
			for _, mm := range memMap {
				mems = append(mems, mm...)
			}
		}
	}

	if includeRelated {
		// Batch-load components and memories for related nodes
		relatedIDs := b.collectRelatedNodeIDs(taskType, nodeID, rels)
		if len(relatedIDs) > 0 {
			if compMap, err := store.GetComponentsByNodeIDs(relatedIDs); err == nil {
				for _, cc := range compMap {
					comps = append(comps, cc...)
				}
			}
			if memMap, err := store.GetMemoriesByNodeIDs(relatedIDs, 5); err == nil {
				for _, mm := range memMap {
					mems = append(mems, mm...)
				}
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

func (b *ContextBuilder) collectNodeIDs(nodes []store.NodeModel) []int64 {
	ids := make([]int64, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.ID)
	}
	return ids
}

func (b *ContextBuilder) collectRelatedNodeIDs(taskType TaskType, nodeID string, rels []store.RelationModel) []int64 {
	ids := make([]int64, 0, len(rels))
	for _, r := range rels {
		if !shouldIncludeRelatedRelation(taskType, nodeID, r) {
			continue
		}
		relatedUUID := r.SourceUUID
		if relatedUUID == nodeID {
			relatedUUID = r.TargetUUID
		}
		if nid := store.ResolveNodeUUID(relatedUUID); nid != 0 {
			ids = append(ids, nid)
		}
	}
	return ids
}

func shouldIncludeRelatedRelation(taskType TaskType, nodeID string, rel store.RelationModel) bool {
	// external_parent 当前只保留结构化边语义，不进入默认 prompt 关系扩图。
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
	}
	// Batch-load environment node components and memories
	if envIDs := b.collectNodeIDs(dedupeNodes(nodes)); len(envIDs) > 0 {
		if compMap, err := store.GetComponentsByNodeIDs(envIDs); err == nil {
			for _, n := range dedupeNodes(nodes) {
				if cc, ok := compMap[n.ID]; ok {
					for _, comp := range cc {
						parts = append(parts, fmt.Sprintf("    【%s/%s】%s", n.Name, comp.ComponentType, comp.Data))
					}
				}
			}
		}
		if memMap, err := store.GetMemoriesByNodeIDs(envIDs, 5); err == nil {
			for _, n := range dedupeNodes(nodes) {
				if mm, ok := memMap[n.ID]; ok {
					for _, mem := range mm {
						parts = append(parts, fmt.Sprintf("    [环境记忆:%s] %s", mem.Level, mem.Content))
					}
				}
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
