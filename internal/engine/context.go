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
	Interaction          *InteractionContext    `json:"interaction,omitempty"`
	SpeakerNode          *store.NodeModel       `json:"speaker_node,omitempty"`
	TargetNode           *store.NodeModel       `json:"target_node,omitempty"`
	SceneNode            *store.NodeModel       `json:"scene_node,omitempty"`
	ParticipantNodes     []store.NodeModel      `json:"participant_nodes,omitempty"`
	StateBlocks          []string               `json:"state_blocks,omitempty"`
	SystemPrompt         string                 `json:"system_prompt"`
}

type interactionView struct {
	interaction      *InteractionContext
	speakerNode      *store.NodeModel
	targetNode       *store.NodeModel
	sceneNode        *store.NodeModel
	participantNodes []store.NodeModel
}

// Build assembles the base reasoning context around one task focus node.
func (b *ContextBuilder) Build(taskType TaskType, nodeID string, depth int, memoryLimit int, includeRelated bool, interaction *InteractionContext) (*BuiltContext, error) {
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
	interactionData := b.resolveInteractionView(node, environmentNode, taskType, rels, interaction)

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

	interactionNodeIDs := b.collectInteractionNodeIDs(interactionData)
	if len(interactionNodeIDs) > 0 {
		if compMap, err := store.GetComponentsByNodeIDs(interactionNodeIDs); err == nil {
			for _, cc := range compMap {
				comps = append(comps, cc...)
			}
		}
		if memMap, err := store.GetMemoriesByNodeIDs(interactionNodeIDs, b.relatedMemoryLimit(memLimit)); err == nil {
			for _, mm := range memMap {
				mems = append(mems, mm...)
			}
		}
	}

	comps = dedupeComponents(comps)
	mems = dedupeMemories(mems)

	stateBlocks := b.buildStateBlocks(node.UUID)
	sysPrompt := b.buildSystemPrompt(node, comps, mems, identityAncestors, environmentNode, environmentAncestors, stateBlocks, interactionData)
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
		Interaction:          interactionData.interaction,
		SpeakerNode:          interactionData.speakerNode,
		TargetNode:           interactionData.targetNode,
		SceneNode:            interactionData.sceneNode,
		ParticipantNodes:     interactionData.participantNodes,
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

func (b *ContextBuilder) collectInteractionNodeIDs(view *interactionView) []int64 {
	if view == nil {
		return nil
	}
	var nodes []store.NodeModel
	if view.speakerNode != nil {
		nodes = append(nodes, *view.speakerNode)
	}
	if view.targetNode != nil {
		nodes = append(nodes, *view.targetNode)
	}
	if view.sceneNode != nil {
		nodes = append(nodes, *view.sceneNode)
	}
	nodes = append(nodes, view.participantNodes...)
	return b.collectNodeIDs(dedupeNodes(nodes))
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

func (b *ContextBuilder) resolveInteractionView(node *store.NodeModel, environmentNode *store.NodeModel, taskType TaskType, rels []store.RelationModel, interaction *InteractionContext) *interactionView {
	view := &interactionView{}
	if interaction == nil {
		return view
	}
	view.interaction = interaction
	view.targetNode = firstNonNilNode(b.loadNodeByUUID(interaction.TargetNodeID), node)
	view.speakerNode = b.loadNodeByUUID(interaction.SpeakerNodeID)
	view.sceneNode = b.loadNodeByUUID(interaction.SceneNodeID)
	if view.sceneNode == nil {
		if environmentNode != nil {
			view.sceneNode = environmentNode
		} else if taskType == TaskNPCDialogue || taskType == TaskAutonomousAct || taskType == TaskCustom {
			view.sceneNode = b.resolveEnvironmentNode(taskType, node, rels)
		}
	}
	view.participantNodes = b.loadParticipantNodes(interaction.ParticipantNodeIDs, view.speakerNode, view.targetNode)
	return view
}

func (b *ContextBuilder) loadNodeByUUID(nodeID string) *store.NodeModel {
	if strings.TrimSpace(nodeID) == "" {
		return nil
	}
	node, err := store.GetNode(nodeID)
	if err != nil {
		return nil
	}
	return node
}

func (b *ContextBuilder) loadParticipantNodes(participantIDs []string, extra ...*store.NodeModel) []store.NodeModel {
	seen := map[string]bool{}
	var participants []store.NodeModel
	appendNode := func(node *store.NodeModel) {
		if node == nil || strings.TrimSpace(node.UUID) == "" || seen[node.UUID] {
			return
		}
		seen[node.UUID] = true
		participants = append(participants, *node)
	}
	for _, participantID := range participantIDs {
		appendNode(b.loadNodeByUUID(participantID))
	}
	for _, node := range extra {
		appendNode(node)
	}
	return participants
}

func firstNonNilNode(nodes ...*store.NodeModel) *store.NodeModel {
	for _, node := range nodes {
		if node != nil {
			return node
		}
	}
	return nil
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
	for i := 0; i < maxDepth && current != nil && current.ParentID != nil; i++ {
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

func (b *ContextBuilder) buildSystemPrompt(node *store.NodeModel, comps []store.ComponentModel, mems []store.MemoryModel, identityAncestors []store.NodeModel, environmentNode *store.NodeModel, environmentAncestors []store.NodeModel, stateBlocks []string, interaction *interactionView) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("你是 %s（%s）。", node.Name, node.NodeType))

	for _, c := range comps {
		parts = append(parts, fmt.Sprintf("【%s】%s", c.ComponentType, c.Data))
	}

	if len(identityAncestors) > 0 {
		var chain []string
		for _, ancestor := range identityAncestors {
			chain = append(chain, fmt.Sprintf("%s(%s)", ancestor.Name, ancestor.NodeType))
		}
		parts = append(parts, fmt.Sprintf("稳定归属链：%s", strings.Join(chain, " > ")))
	}

	if environmentNode != nil {
		var chain []string
		chain = append(chain, fmt.Sprintf("%s(%s)", environmentNode.Name, environmentNode.NodeType))
		for _, ancestor := range environmentAncestors {
			chain = append(chain, fmt.Sprintf("%s(%s)", ancestor.Name, ancestor.NodeType))
		}
		parts = append(parts, fmt.Sprintf("当前环境链：%s", strings.Join(chain, " > ")))
		parts = append(parts, b.buildEnvironmentPromptBlock(environmentNode, environmentAncestors)...)
	}

	parts = append(parts, b.buildInteractionPromptBlock(interaction)...)

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

func (b *ContextBuilder) buildInteractionPromptBlock(view *interactionView) []string {
	if view == nil || view.interaction == nil {
		return nil
	}
	var parts []string
	parts = append(parts, "交互语义：")
	parts = append(parts, fmt.Sprintf("  [mode] %s", view.interaction.Mode))
	if view.speakerNode != nil {
		parts = append(parts, fmt.Sprintf("  [speaker] %s(%s)", view.speakerNode.Name, view.speakerNode.NodeType))
	}
	if view.targetNode != nil {
		parts = append(parts, fmt.Sprintf("  [target] %s(%s)", view.targetNode.Name, view.targetNode.NodeType))
	}
	if view.sceneNode != nil {
		parts = append(parts, fmt.Sprintf("  [scene] %s(%s)", view.sceneNode.Name, view.sceneNode.NodeType))
	}
	if strings.TrimSpace(view.interaction.RoomID) != "" {
		parts = append(parts, fmt.Sprintf("  [room_id] %s", view.interaction.RoomID))
	}
	if strings.TrimSpace(view.interaction.AudienceScope) != "" {
		parts = append(parts, fmt.Sprintf("  [audience_scope] %s", view.interaction.AudienceScope))
	}
	parts = append(parts, fmt.Sprintf("  [turn_index] %d", view.interaction.TurnIndex))
	if view.interaction.Event != nil {
		parts = append(parts, fmt.Sprintf("  [event] %s", view.interaction.Event.Type))
		if strings.TrimSpace(view.interaction.Event.ItemID) != "" {
			parts = append(parts, fmt.Sprintf("  [event_item_id] %s", view.interaction.Event.ItemID))
		}
	}
	if len(view.participantNodes) > 0 {
		parts = append(parts, "  [participants]")
		for _, participant := range view.participantNodes {
			parts = append(parts, fmt.Sprintf("    - %s(%s)", participant.Name, participant.NodeType))
		}
	}
	return parts
}

func (b *ContextBuilder) buildEnvironmentPromptBlock(environmentNode *store.NodeModel, environmentAncestors []store.NodeModel) []string {
	if environmentNode == nil {
		return nil
	}
	nodes := []store.NodeModel{*environmentNode}
	nodes = append(nodes, environmentAncestors...)
	dedupedNodes := dedupeNodes(nodes)
	var parts []string
	parts = append(parts, "环境信息：")
	for _, n := range dedupedNodes {
		parts = append(parts, fmt.Sprintf("  [环境节点] %s(%s)", n.Name, n.NodeType))
	}
	if envIDs := b.collectNodeIDs(dedupedNodes); len(envIDs) > 0 {
		if compMap, err := store.GetComponentsByNodeIDs(envIDs); err == nil {
			for _, n := range dedupedNodes {
				if cc, ok := compMap[n.ID]; ok {
					for _, comp := range cc {
						parts = append(parts, fmt.Sprintf("    【%s/%s】%s", n.Name, comp.ComponentType, comp.Data))
					}
				}
			}
		}
		if memMap, err := store.GetMemoriesByNodeIDs(envIDs, 5); err == nil {
			for _, n := range dedupedNodes {
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
