package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// PromotedFocusNode represents a descendant node promoted into tick context
// via the world_focus component.
type PromotedFocusNode struct {
	NodeID   string `json:"node_id"`
	Name     string `json:"name,omitempty"`
	NodeType string `json:"node_type,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

const (
	maxFocusScanDepth = 5
	maxFocusNodeCount = 10
)

// ScanWorldFocusDescendants scans descendant nodes under the given focus node
// that carry world_focus components matching the specified task type.
// Returns promoted nodes sorted by priority (descending), capped at maxFocusNodeCount.
func ScanWorldFocusDescendants(focusNodeID string, taskType string) ([]PromotedFocusNode, error) {
	nodes, err := store.GetChildNodes(focusNodeID)
	if err != nil {
		return nil, fmt.Errorf("get child nodes for %s: %w", focusNodeID, err)
	}
	_ = nodes // We walk recursively below

	var promoted []PromotedFocusNode
	scanned := map[string]bool{}

	var walk func(nodeID string, depth int)
	walk = func(nodeID string, depth int) {
		if depth > maxFocusScanDepth {
			return
		}
		if scanned[nodeID] {
			return
		}
		scanned[nodeID] = true

		// Look for world_focus component on this node
		comps, err := store.GetComponentsByType(nodeID, string(CompWorldFocus))
		if err == nil && len(comps) > 0 {
			for _, comp := range comps {
				cfg, err := DecodeWorldFocusConfig(comp.Data)
				if err != nil || cfg == nil || !cfg.Enabled {
					continue
				}
				// Check task type match
				if !matchesTaskType(cfg.Tasks, taskType) {
					continue
				}
				// Get node details
				nodeModel, err := store.GetNode(nodeID)
				if err != nil {
					log.Printf("[world_focus] get node %s: %v", nodeID, err)
					continue
				}
				promoted = append(promoted, PromotedFocusNode{
					NodeID:   nodeID,
					Name:     nodeModel.Name,
					NodeType: nodeModel.NodeType,
					Reason:   cfg.Reason,
					Priority: cfg.Priority,
				})
			}
		}

		// Recurse into children
		children, err := store.GetChildNodes(nodeID)
		if err != nil {
			log.Printf("[world_focus] get children for %s: %v", nodeID, err)
			return
		}
		for _, child := range children {
			walk(child.UUID, depth+1)
		}
	}

	walk(focusNodeID, 0)

	// Sort by priority descending, then by name
	sort.Slice(promoted, func(i, j int) bool {
		if promoted[i].Priority != promoted[j].Priority {
			return promoted[i].Priority > promoted[j].Priority
		}
		return promoted[i].Name < promoted[j].Name
	})

	// Cap at maxFocusNodeCount
	if len(promoted) > maxFocusNodeCount {
		promoted = promoted[:maxFocusNodeCount]
	}

	return promoted, nil
}

// DecodeWorldFocusConfig parses a world_focus component payload.
func DecodeWorldFocusConfig(data string) (*WorldFocusConfig, error) {
	if strings.TrimSpace(data) == "" {
		return nil, nil
	}
	var cfg WorldFocusConfig
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, fmt.Errorf("parse world_focus config: %w", err)
	}
	return &cfg, nil
}

func matchesTaskType(tasks []string, taskType string) bool {
	if len(tasks) == 0 {
		return true // empty = match all
	}
	for _, t := range tasks {
		if t == taskType {
			return true
		}
	}
	return false
}

// BuildWorldFocusBlock generates a formatted text block for promoted focus nodes.
func BuildWorldFocusBlock(promoted []PromotedFocusNode) string {
	if len(promoted) == 0 {
		return ""
	}
	var parts []string
	parts = append(parts, "\u3010world_focus \u89c2\u5bdf\u8282\u70b9\u3011\u4ee5\u4e0b\u5b50\u8282\u70b9\u5df2\u88ab\u6807\u8bb0\u4e3a\u672c tick \u7684\u91cd\u70b9\u89c2\u5bdf\u5bf9\u8c61\uff1a")
	for _, p := range promoted {
		line := fmt.Sprintf("- %s (%s)", p.Name, p.NodeType)
		if p.Reason != "" {
			line += fmt.Sprintf(" [\u539f\u56e0: %s]", p.Reason)
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, "\n")
}
