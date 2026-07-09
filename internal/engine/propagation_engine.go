package engine

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// ManualPropagateMemory wraps manual propagation requests with the same
// execution mode and world-setting context used by normal pipeline runs.
func (p *Pipeline) ManualPropagateMemory(req *InvokeRequest, mem MemoryUpdate, sourceNodeID string) {
	executionMode := p.getExecutionMode()
	_, maxRounds, retries, timeout, pipelineMode := p.loadWorldSettings(req.WorldID)
	configuredMode := PipelineMode(pipelineMode)
	if configuredMode == "" {
		configuredMode = PipelineFull
	}
	runtime := &executionConfig{
		maxRounds:              maxRounds,
		subTaskRetries:         retries,
		subTaskTimeout:         timeout,
		configuredPipelineMode: configuredMode,
		pipelineMode:           configuredMode,
		policyEngine:           p.loadWorldPolicy(req.WorldID),
	}
	p.PropagateMemoryByRule(req, runtime, executionMode, mem, sourceNodeID)
}

// PropagateMemoryByRule 根据传播规则选择传播路径。
//
// 当前约束：
// 1. 默认 upward 传播只沿主 parent 链工作，它反映的是稳定归属作用域，而不是当前环境作用域。
// 2. 在 parent 与 located_at 语义分离后，任何环境传播或组织传播都不应偷偷复用 upward 语义，而应通过新的显式
//    传播模式建模。
// 3. external_parent 是否参与默认传播必须由实现显式声明；不能因为它名字里有 parent 就自动混入 upward。
func (p *Pipeline) PropagateMemoryByRule(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, mem MemoryUpdate, sourceNodeID string) {
	rule := mem.Propagation
	mode := PropModeUpward
	maxDepth := 0
	publishUp := false
	p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_started", mem.Content, map[string]any{"source_node_id": sourceNodeID, "level": mem.Level, "tags": mem.Tags, "rule": rule})

	if rule != nil {
		mode = rule.Mode
		if rule.MaxDepth > 0 {
			maxDepth = rule.MaxDepth
		} else {
			if node, err := store.GetNode(sourceNodeID); err == nil {
				if settings, err := store.GetOrCreateWorldSettings(node.WorldUUID); err == nil && settings.PropagationMaxDepth > 0 {
					maxDepth = settings.PropagationMaxDepth
				}
			}
		}
		publishUp = rule.PublishUp
	} else {
		if node, err := store.GetNode(sourceNodeID); err == nil {
			if settings, err := store.GetOrCreateWorldSettings(node.WorldUUID); err == nil && settings.PropagationMaxDepth > 0 {
				maxDepth = settings.PropagationMaxDepth
			}
		}
	}

	switch mode {
	case PropModeUpward:
		p.PropagateUpward(req, runtime, executionMode, sourceNodeID, mem.Content, mem.Level, maxDepth, publishUp)
	case PropModeEnvironment:
		p.PropagateEnvironmentScope(req, runtime, executionMode, sourceNodeID, mem.Content, mem.Level, maxDepth)
	case PropModeOrganization:
		p.PropagateOrganizationScope(req, runtime, executionMode, sourceNodeID, mem.Content, mem.Level, maxDepth)
	case PropModeTagBroadcast:
		p.PropagateByTags(req, runtime, executionMode, sourceNodeID, mem, rule)
	case PropModeTargeted:
		p.PropagateToTargets(req, runtime, executionMode, mem, rule)
	case PropModeManual:
		p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_skipped", mem.Content, map[string]any{"source_node_id": sourceNodeID, "reason": "manual_mode"})
		return
	}
	p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_completed", mem.Content, map[string]any{"source_node_id": sourceNodeID, "mode": mode, "max_depth": maxDepth, "publish_up": publishUp})
}

// PropagateUpward 沿主 parent 链向上写入传播记忆。
//
// 注意：这条路径只适用于稳定主归属链，不能替代 located_at 导航出的环境作用域，也不能自动表达组织控制链。
func (p *Pipeline) PropagateUpward(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, nodeID, content string, level MemoryLevel, maxDepth int, publishUp bool) {
	visited := map[string]bool{}
	var walk func(nid string, depth int)
	walk = func(nid string, depth int) {
		if visited[nid] {
			return
		}
		visited[nid] = true

		node, err := store.GetNode(nid)
		if err != nil || node.ParentUUID == nil {
			return
		}

		propLevel := level
		switch level {
		case MemShortTerm:
			if depth >= 1 {
				return
			}
		case MemLongTerm:
			propLevel = MemShared
		}

		parentUUID := *node.ParentUUID

		if hasPropagatedMemory(parentUUID, content) {
			return
		}

		parentNodeID := store.ResolveNodeUUID(parentUUID)
		if parentNodeID == 0 {
			return
		}
		m := &store.MemoryModel{
			NodeID:  parentNodeID,
			Content: content,
			Level:   string(propLevel),
			Tags:    "propagated",
		}
		if err := store.CreateMemory(m); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_write_failed", content, map[string]any{"target_node_id": parentUUID, "mode": "upward", "error": err.Error()})
			log.Printf("propagate upward: %v", err)
		} else {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_written", content, map[string]any{"target_node_id": parentUUID, "mode": "upward", "level": propLevel})
		}

		if publishUp {
			if node.NodeType == string(NodeTypeWorld) || node.NodeType == string(NodeTypeFaction) {
				m2 := &store.MemoryModel{
					NodeID:  parentNodeID,
					Content: content,
					Level:   string(MemWorld),
					Tags:    "propagated,published",
				}
				if err := store.CreateMemory(m2); err != nil {
					p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_write_failed", content, map[string]any{"target_node_id": parentUUID, "mode": "publish_up", "error": err.Error()})
					log.Printf("propagate publish: %v", err)
				} else {
					p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_written", content, map[string]any{"target_node_id": parentUUID, "mode": "publish_up", "level": MemWorld})
				}
				return
			}
		}

		if maxDepth > 0 && depth+1 >= maxDepth {
			return
		}

		walk(parentUUID, depth+1)
	}
	walk(nodeID, 0)
}

// PropagateEnvironmentScope 沿当前 located_at 环境节点及其主 parent 场景祖先传播。
// 这条路径只服务动态环境作用域，不读取 belongs_to/subordinate/external_parent。
func (p *Pipeline) PropagateEnvironmentScope(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, sourceNodeID, content string, level MemoryLevel, maxDepth int) {
	targets, err := resolveEnvironmentPropagationTargets(sourceNodeID, maxDepth)
	if err != nil {
		p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_failed", content, map[string]any{"source_node_id": sourceNodeID, "mode": "environment_scope", "error": err.Error()})
		log.Printf("propagate environment scope: %v", err)
		return
	}
	p.propagateToResolvedTargets(req, runtime, executionMode, targets, content, level, "environment_scope")
}

// PropagateOrganizationScope 沿 belongs_to/subordinate 指向的组织/控制节点及其主 parent 链传播。
// 这条路径只服务组织与控制作用域，不读取 located_at 或 external_parent。
func (p *Pipeline) PropagateOrganizationScope(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, sourceNodeID, content string, level MemoryLevel, maxDepth int) {
	targets, err := resolveOrganizationPropagationTargets(sourceNodeID, maxDepth)
	if err != nil {
		p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_failed", content, map[string]any{"source_node_id": sourceNodeID, "mode": "organization_scope", "error": err.Error()})
		log.Printf("propagate organization scope: %v", err)
		return
	}
	p.propagateToResolvedTargets(req, runtime, executionMode, targets, content, level, "organization_scope")
}

func (p *Pipeline) PropagateByTags(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, sourceNodeID string, mem MemoryUpdate, rule *PropagationRule) {
	sourceNode, err := store.GetNode(sourceNodeID)
	if err != nil {
		p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_failed", mem.Content, map[string]any{"source_node_id": sourceNodeID, "mode": "tag_broadcast", "error": err.Error()})
		log.Printf("propagate by tags: get source node %s: %v", sourceNodeID, err)
		return
	}

	settings, _ := store.GetOrCreateWorldSettings(sourceNode.WorldUUID)
	if settings != nil && settings.EnablePropagationMachine {
		p.runPropagationMachine(req, runtime, executionMode, sourceNode, mem)
		return
	}

	nodes, err := store.FindNodesByTags(sourceNode.WorldUUID, rule.TargetTags)
	if err != nil {
		p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_failed", mem.Content, map[string]any{"source_node_id": sourceNodeID, "mode": "tag_broadcast", "error": err.Error()})
		log.Printf("propagate by tags: %v", err)
		return
	}

	for _, n := range nodes {
		if n.UUID == sourceNodeID {
			continue
		}
		if hasPropagatedMemory(n.UUID, mem.Content) {
			continue
		}
		m := &store.MemoryModel{
			NodeID:  n.ID,
			Content: mem.Content,
			Level:   string(mem.Level),
			Tags:    "propagated,broadcast",
		}
		if err := store.CreateMemory(m); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_write_failed", mem.Content, map[string]any{"target_node_id": n.UUID, "mode": "tag_broadcast", "error": err.Error()})
			log.Printf("propagate broadcast to %s: %v", n.UUID, err)
		} else {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_written", mem.Content, map[string]any{"target_node_id": n.UUID, "mode": "tag_broadcast", "level": mem.Level})
		}
	}
}

func (p *Pipeline) PropagateToTargets(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, mem MemoryUpdate, rule *PropagationRule) {
	for _, targetUUID := range rule.TargetNodeIDs {
		if hasPropagatedMemory(targetUUID, mem.Content) {
			continue
		}
		targetNodeID := store.ResolveNodeUUID(targetUUID)
		if targetNodeID == 0 {
			continue
		}
		m := &store.MemoryModel{
			NodeID:  targetNodeID,
			Content: mem.Content,
			Level:   string(mem.Level),
			Tags:    "propagated,targeted",
		}
		if err := store.CreateMemory(m); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_write_failed", mem.Content, map[string]any{"target_node_id": targetUUID, "mode": "targeted", "error": err.Error()})
			log.Printf("propagate targeted to %s: %v", targetUUID, err)
		} else {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_written", mem.Content, map[string]any{"target_node_id": targetUUID, "mode": "targeted", "level": mem.Level})
		}
	}
}

type propagationTarget struct {
	NodeUUID string
	NodeID   int64
	Level    MemoryLevel
}

func resolveEnvironmentPropagationTargets(sourceNodeID string, maxDepth int) ([]propagationTarget, error) {
	rels, err := store.GetNodeRelations(sourceNodeID)
	if err != nil {
		return nil, err
	}
	var targets []propagationTarget
	seen := map[string]bool{}
	for _, rel := range rels {
		if rel.SourceUUID != sourceNodeID || rel.RelationType != string(RelLocatedAt) {
			continue
		}
		appendPropagationTarget(&targets, seen, rel.TargetUUID, MemShared)
		ancestors, err := collectParentChainTargets(rel.TargetUUID, maxDepth, MemWorld)
		if err != nil {
			return nil, err
		}
		for _, item := range ancestors {
			appendResolvedTarget(&targets, seen, item)
		}
	}
	return targets, nil
}

func resolveOrganizationPropagationTargets(sourceNodeID string, maxDepth int) ([]propagationTarget, error) {
	rels, err := store.GetNodeRelations(sourceNodeID)
	if err != nil {
		return nil, err
	}
	var targets []propagationTarget
	seen := map[string]bool{}
	for _, rel := range rels {
		if rel.SourceUUID != sourceNodeID {
			continue
		}
		if rel.RelationType != string(RelBelongsTo) && rel.RelationType != string(RelSubordinate) {
			continue
		}
		appendPropagationTarget(&targets, seen, rel.TargetUUID, MemShared)
		ancestors, err := collectParentChainTargets(rel.TargetUUID, maxDepth, MemWorld)
		if err != nil {
			return nil, err
		}
		for _, item := range ancestors {
			appendResolvedTarget(&targets, seen, item)
		}
	}
	return targets, nil
}

func collectParentChainTargets(startUUID string, maxDepth int, level MemoryLevel) ([]propagationTarget, error) {
	var targets []propagationTarget
	seen := map[string]bool{}
	currentUUID := startUUID
	for depth := 0; ; depth++ {
		node, err := store.GetNode(currentUUID)
		if err != nil {
			return nil, err
		}
		if node.ParentUUID == nil {
			return targets, nil
		}
		if maxDepth > 0 && depth >= maxDepth {
			return targets, nil
		}
		parentUUID := *node.ParentUUID
		if seen[parentUUID] {
			return targets, nil
		}
		seen[parentUUID] = true
		targets = append(targets, propagationTarget{NodeUUID: parentUUID, NodeID: store.ResolveNodeUUID(parentUUID), Level: level})
		currentUUID = parentUUID
	}
}

func appendPropagationTarget(targets *[]propagationTarget, seen map[string]bool, nodeUUID string, level MemoryLevel) {
	appendResolvedTarget(targets, seen, propagationTarget{NodeUUID: nodeUUID, NodeID: store.ResolveNodeUUID(nodeUUID), Level: level})
}

func appendResolvedTarget(targets *[]propagationTarget, seen map[string]bool, target propagationTarget) {
	if target.NodeUUID == "" || target.NodeID == 0 || seen[target.NodeUUID] {
		return
	}
	seen[target.NodeUUID] = true
	*targets = append(*targets, target)
}

func (p *Pipeline) propagateToResolvedTargets(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, targets []propagationTarget, content string, level MemoryLevel, mode string) {
	for _, target := range targets {
		if hasPropagatedMemory(target.NodeUUID, content) {
			continue
		}
		m := &store.MemoryModel{
			NodeID:  target.NodeID,
			Content: content,
			Level:   string(target.Level),
			Tags:    "propagated," + mode,
		}
		if err := store.CreateMemory(m); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_write_failed", content, map[string]any{"target_node_id": target.NodeUUID, "mode": mode, "error": err.Error()})
			log.Printf("propagate %s to %s: %v", mode, target.NodeUUID, err)
		} else {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_written", content, map[string]any{"target_node_id": target.NodeUUID, "mode": mode, "level": target.Level})
		}
	}
}

func hasPropagatedMemory(nodeID, content string) bool {
	var count int64
	// nodeID is UUID string, query by UUID via store layer
	node := store.ResolveNodeUUID(nodeID)
	if node == 0 {
		return false
	}
	store.DB.Model(&store.MemoryModel{}).Where("node_id = ? AND content = ? AND tags LIKE ?", node, content, "%propagated%").Count(&count)
	return count > 0
}

func (p *Pipeline) runPropagationMachine(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, sourceNode *store.NodeModel, mem MemoryUpdate) {
	chains, err := store.GetPropagationChains(sourceNode.WorldUUID)
	if err != nil {
		p.emitExecutionEvent(req, runtime, executionMode, "propagation_machine_failed", mem.Content, map[string]any{"source_node_id": sourceNode.UUID, "error": err.Error()})
		log.Printf("propagation machine: %v", err)
		return
	}
	if len(chains) == 0 {
		return
	}

	processed := map[string]bool{}
	queue := make([]string, 0, len(chains))

	for _, chain := range chains {
		if !processed[chain.UUID] {
			processed[chain.UUID] = true
			queue = append(queue, chain.UUID)
		}
	}

	maxDepth := 10
	if len(chains) > 0 {
		if chains[0].MaxDepth > 0 {
			maxDepth = chains[0].MaxDepth
		}
	}

	for depth := 0; depth < maxDepth && len(queue) > 0; depth++ {
		currentUUID := queue[0]
		queue = queue[1:]

		var currentChain *store.PropagationChainModel
		for i := range chains {
			if chains[i].UUID == currentUUID {
				currentChain = &chains[i]
				break
			}
		}
		if currentChain == nil {
			continue
		}

		var actions []PropagateAction
		if err := json.Unmarshal([]byte(currentChain.Actions), &actions); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "propagation_machine_failed", mem.Content, map[string]any{"chain_id": currentChain.UUID, "error": err.Error()})
			log.Printf("propagation machine: parse actions for chain %s: %v", currentChain.UUID, err)
			continue
		}

		for _, action := range actions {
			actionMem := applyTransform(mem, &action)
			switch action.Mode {
			case PropModeTagBroadcast:
				p.executeMachineBroadcast(req, runtime, executionMode, sourceNode.WorldUUID, sourceNode.UUID, actionMem, action)
			case PropModeTargeted:
				p.executeMachineTargeted(req, runtime, executionMode, actionMem, action)
			case PropModeUpward:
				p.PropagateUpward(req, runtime, executionMode, sourceNode.UUID, actionMem.Content, actionMem.Level, action.MaxDepth, action.PublishUp)
			case PropModeEnvironment:
				p.PropagateEnvironmentScope(req, runtime, executionMode, sourceNode.UUID, actionMem.Content, actionMem.Level, action.MaxDepth)
			case PropModeOrganization:
				p.PropagateOrganizationScope(req, runtime, executionMode, sourceNode.UUID, actionMem.Content, actionMem.Level, action.MaxDepth)
			}

			for _, nextUUID := range action.NextChainIDs {
				if !processed[nextUUID] {
					processed[nextUUID] = true
					queue = append(queue, nextUUID)
				}
			}
		}
	}
}

func (p *Pipeline) executeMachineBroadcast(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, worldUUID, sourceNodeUUID string, mem MemoryUpdate, action PropagateAction) {
	nodes, err := store.FindNodesByTags(worldUUID, action.TargetTags)
	if err != nil {
		p.emitExecutionEvent(req, runtime, executionMode, "propagation_machine_failed", mem.Content, map[string]any{"mode": "broadcast", "error": err.Error()})
		log.Printf("propagation machine broadcast: %v", err)
		return
	}
	for _, n := range nodes {
		if n.UUID == sourceNodeUUID {
			continue
		}
		if hasPropagatedMemory(n.UUID, mem.Content) {
			continue
		}
		m := &store.MemoryModel{
			NodeID:  n.ID,
			Content: mem.Content,
			Level:   string(mem.Level),
			Tags:    "propagated,machine",
		}
		if err := store.CreateMemory(m); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_write_failed", mem.Content, map[string]any{"target_node_id": n.UUID, "mode": "machine_broadcast", "error": err.Error()})
			log.Printf("propagation machine broadcast to %s: %v", n.UUID, err)
		} else {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_written", mem.Content, map[string]any{"target_node_id": n.UUID, "mode": "machine_broadcast", "level": mem.Level})
		}
	}
}

func (p *Pipeline) executeMachineTargeted(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, mem MemoryUpdate, action PropagateAction) {
	for _, targetUUID := range action.TargetNodeIDs {
		if hasPropagatedMemory(targetUUID, mem.Content) {
			continue
		}
		targetNodeID := store.ResolveNodeUUID(targetUUID)
		if targetNodeID == 0 {
			continue
		}
		m := &store.MemoryModel{
			NodeID:  targetNodeID,
			Content: mem.Content,
			Level:   string(mem.Level),
			Tags:    "propagated,machine",
		}
		if err := store.CreateMemory(m); err != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_write_failed", mem.Content, map[string]any{"target_node_id": targetUUID, "mode": "machine_targeted", "error": err.Error()})
			log.Printf("propagation machine targeted to %s: %v", targetUUID, err)
		} else {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_propagation_written", mem.Content, map[string]any{"target_node_id": targetUUID, "mode": "machine_targeted", "level": mem.Level})
		}
	}
}

func applyTransform(mem MemoryUpdate, action *PropagateAction) MemoryUpdate {
	if action.Transform == nil {
		return mem
	}
	result := mem
	t := action.Transform

	if t.ContentPrefix != "" {
		result.Content = t.ContentPrefix + result.Content
	}
	if t.LevelUp {
		switch result.Level {
		case MemShortTerm:
			result.Level = MemLongTerm
		case MemLongTerm:
			result.Level = MemShared
		case MemShared:
			result.Level = MemWorld
		}
	}
	if len(t.AppendTags) > 0 {
		existingTags := parseTags(result.Tags)
		tagSet := map[string]bool{}
		for _, t := range existingTags {
			tagSet[t] = true
		}
		for _, t := range t.AppendTags {
			if !tagSet[t] {
				tagSet[t] = true
			}
		}
		var newTags []string
		for t := range tagSet {
			newTags = append(newTags, t)
		}
		result.Tags = strings.Join(newTags, ",")
	}
	return result
}

func matchesTrigger(triggerTagsJSON string, sourceTags []string) bool {
	var triggerTags []string
	if err := json.Unmarshal([]byte(triggerTagsJSON), &triggerTags); err != nil {
		return false
	}
	sourceSet := map[string]bool{}
	for _, t := range sourceTags {
		sourceSet[strings.TrimSpace(strings.ToLower(t))] = true
	}
	for _, t := range triggerTags {
		if sourceSet[strings.TrimSpace(strings.ToLower(t))] {
			return true
		}
	}
	return false
}

func matchesNodeType(triggerNodeTypesJSON, nodeType string) bool {
	if triggerNodeTypesJSON == "" || triggerNodeTypesJSON == "null" || triggerNodeTypesJSON == "[]" {
		return true
	}
	var types []string
	if err := json.Unmarshal([]byte(triggerNodeTypesJSON), &types); err != nil {
		return true
	}
	for _, t := range types {
		if t == nodeType {
			return true
		}
	}
	return false
}

func parseTags(tags string) []string {
	if tags == "" {
		return nil
	}
	var result []string
	for _, t := range strings.Split(tags, ",") {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
