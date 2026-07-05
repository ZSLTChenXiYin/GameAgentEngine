package engine

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func (p *Pipeline) PropagateMemoryByRule(mem MemoryUpdate, sourceNodeID string) {
	rule := mem.Propagation
	mode := PropModeUpward
	maxDepth := 0
	publishUp := false

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
		p.PropagateUpward(sourceNodeID, mem.Content, mem.Level, maxDepth, publishUp)
	case PropModeTagBroadcast:
		p.PropagateByTags(sourceNodeID, mem, rule)
	case PropModeTargeted:
		p.PropagateToTargets(mem, rule)
	case PropModeManual:
		return
	}
}

func (p *Pipeline) PropagateUpward(nodeID, content string, level MemoryLevel, maxDepth int, publishUp bool) {
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
			log.Printf("propagate upward: %v", err)
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
					log.Printf("propagate publish: %v", err)
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

func (p *Pipeline) PropagateByTags(sourceNodeID string, mem MemoryUpdate, rule *PropagationRule) {
	sourceNode, err := store.GetNode(sourceNodeID)
	if err != nil {
		log.Printf("propagate by tags: get source node %s: %v", sourceNodeID, err)
		return
	}

	settings, _ := store.GetOrCreateWorldSettings(sourceNode.WorldUUID)
	if settings != nil && settings.EnablePropagationMachine {
		p.runPropagationMachine(sourceNode, mem)
		return
	}

	nodes, err := store.FindNodesByTags(sourceNode.WorldUUID, rule.TargetTags)
	if err != nil {
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
			log.Printf("propagate broadcast to %s: %v", n.UUID, err)
		}
	}
}

func (p *Pipeline) PropagateToTargets(mem MemoryUpdate, rule *PropagationRule) {
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
			log.Printf("propagate targeted to %s: %v", targetUUID, err)
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

func (p *Pipeline) runPropagationMachine(sourceNode *store.NodeModel, mem MemoryUpdate) {
	chains, err := store.GetPropagationChains(sourceNode.WorldUUID)
	if err != nil {
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
			log.Printf("propagation machine: parse actions for chain %s: %v", currentChain.UUID, err)
			continue
		}

		for _, action := range actions {
			actionMem := applyTransform(mem, &action)
			switch action.Mode {
			case PropModeTagBroadcast:
				p.executeMachineBroadcast(sourceNode.WorldUUID, sourceNode.UUID, actionMem, action)
			case PropModeTargeted:
				p.executeMachineTargeted(actionMem, action)
			case PropModeUpward:
				p.PropagateUpward(sourceNode.UUID, actionMem.Content, actionMem.Level, action.MaxDepth, action.PublishUp)
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

func (p *Pipeline) executeMachineBroadcast(worldUUID, sourceNodeUUID string, mem MemoryUpdate, action PropagateAction) {
	nodes, err := store.FindNodesByTags(worldUUID, action.TargetTags)
	if err != nil {
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
			log.Printf("propagation machine broadcast to %s: %v", n.UUID, err)
		}
	}
}

func (p *Pipeline) executeMachineTargeted(mem MemoryUpdate, action PropagateAction) {
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
			log.Printf("propagation machine targeted to %s: %v", targetUUID, err)
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