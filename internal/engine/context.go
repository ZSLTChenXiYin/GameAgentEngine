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
	Node         *store.NodeModel       `json:"node"`
	Components   []store.ComponentModel `json:"components"`
	Memories     []store.MemoryModel    `json:"memories"`
	Relations    []store.RelationModel  `json:"relations"`
	Children     []store.NodeModel      `json:"children"`
	Ancestors    []store.NodeModel      `json:"ancestors"`
	StateBlocks  []string               `json:"state_blocks,omitempty"`
	SystemPrompt string                 `json:"system_prompt"`
}

func (b *ContextBuilder) Build(nodeID string, depth int, memoryLimit int, includeRelated bool) (*BuiltContext, error) {
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

	ancestors := b.collectAncestors(node, depth)

	for _, a := range ancestors {
		if ac, err := store.GetNodeComponents(a.UUID); err == nil {
			comps = append(comps, ac...)
		}
	}

	if includeRelated {
		for _, r := range rels {
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

	stateBlocks := b.buildStateBlocks(node.UUID)
	sysPrompt := b.buildSystemPrompt(node, comps, mems, ancestors, stateBlocks)
	return &BuiltContext{
		Node:         node,
		Components:   comps,
		Memories:     mems,
		Relations:    rels,
		Children:     children,
		Ancestors:    ancestors,
		StateBlocks:  stateBlocks,
		SystemPrompt: sysPrompt,
	}, nil
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

func (b *ContextBuilder) buildSystemPrompt(node *store.NodeModel, comps []store.ComponentModel, mems []store.MemoryModel, ancestors []store.NodeModel, stateBlocks []string) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("你是 %s（%s）。", node.Name, node.NodeType))

	for _, c := range comps {
		parts = append(parts, fmt.Sprintf("【%s】%s", c.ComponentType, c.Data))
	}

	if len(ancestors) > 0 {
		var ap []string
		for _, a := range ancestors {
			ap = append(ap, fmt.Sprintf("%s(%s)", a.Name, a.NodeType))
		}
		parts = append(parts, fmt.Sprintf("所属层级：%s", strings.Join(ap, " > ")))
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

func (b *ContextBuilder) buildStateBlocks(nodeID string) []string {
	componentTypes := []string{string(CompWorldState), string(CompStoryState), string(CompStoryHistory), string(CompTickPolicy), string(CompStateSnapshot)}
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
