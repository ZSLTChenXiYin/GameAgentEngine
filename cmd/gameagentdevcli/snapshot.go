package main

import (
	"fmt"
	"sort"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

// WorldSnapshotNode 表示世界快照中的单个节点及其附属数据。
type WorldSnapshotNode struct {
	Node       sdk.Node        `json:"node" yaml:"node"`
	Components []sdk.Component `json:"components,omitempty" yaml:"components,omitempty"`
	Memories   []sdk.Memory    `json:"memories,omitempty" yaml:"memories,omitempty"`
}

// WorldSnapshot 表示一个世界的完整运行时快照。
type WorldSnapshot struct {
	World     sdk.Node            `json:"world" yaml:"world"`
	Nodes     []WorldSnapshotNode `json:"nodes" yaml:"nodes"`
	Relations []sdk.Relation      `json:"relations" yaml:"relations"`
}

// buildWorldSnapshot 拉取一个世界的节点、组件、记忆和关系，形成可审阅的完整快照。
func buildWorldSnapshot(worldID string) (*WorldSnapshot, error) {
	client := newClient()
	worlds, err := client.GetWorlds()
	if err != nil {
		return nil, err
	}

	var world sdk.Node
	found := false
	for _, item := range worlds {
		if item.ID == worldID {
			world = item
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("world not found: %s", worldID)
	}

	nodes, err := client.GetNodes(worldID, 0, 0, "")
	if err != nil {
		return nil, err
	}
	relations, err := client.GetRelations(worldID, 0, 0, "")
	if err != nil {
		return nil, err
	}

	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].NodeType == nodes[j].NodeType {
			return nodes[i].Name < nodes[j].Name
		}
		return nodes[i].NodeType < nodes[j].NodeType
	})

	snapshotNodes := make([]WorldSnapshotNode, 0, len(nodes))
	for _, node := range nodes {
		components, err := client.GetComponents(node.ID)
		if err != nil {
			return nil, fmt.Errorf("get components for %s: %w", node.Name, err)
		}
		memories, err := client.GetMemories(node.ID)
		if err != nil {
			return nil, fmt.Errorf("get memories for %s: %w", node.Name, err)
		}
		snapshotNodes = append(snapshotNodes, WorldSnapshotNode{
			Node:       node,
			Components: components,
			Memories:   memories,
		})
	}

	return &WorldSnapshot{
		World:     world,
		Nodes:     snapshotNodes,
		Relations: relations,
	}, nil
}

// buildImportConfigFromSnapshot 将运行时快照降级为可再次导入的结构化配置。
func buildImportConfigFromSnapshot(snapshot *WorldSnapshot) *sdk.ImportConfig {
	cfg := &sdk.ImportConfig{
		World: sdk.WorldConfig{Name: snapshot.World.Name},
	}

	nameByID := map[string]string{}
	for _, item := range snapshot.Nodes {
		nameByID[item.Node.ID] = item.Node.Name
	}
	nameByID[snapshot.World.ID] = snapshot.World.Name

	for _, item := range snapshot.Nodes {
		if item.Node.ID == snapshot.World.ID || item.Node.NodeType == "world" {
			continue
		}

		nodeCfg := sdk.NodeConfig{
			Name: item.Node.Name,
			Type: item.Node.NodeType,
		}
		if item.Node.ParentID != nil {
			if parentName, ok := nameByID[*item.Node.ParentID]; ok {
				nodeCfg.Parent = parentName
			}
		}

		for _, component := range item.Components {
			switch component.ComponentType {
			case "profile":
				nodeCfg.Profile = component.Data
			case "lore":
				nodeCfg.Lore = component.Data
			default:
				cfg.Components = append(cfg.Components, sdk.ComponentConfig{
					NodeID: item.Node.Name,
					Type:   component.ComponentType,
					Data:   component.Data,
				})
			}
		}

		for _, memory := range item.Memories {
			nodeCfg.Memories = append(nodeCfg.Memories, struct {
				Content string `json:"content" yaml:"content"`
				Level   string `json:"level,omitempty" yaml:"level,omitempty"`
				Tags    string `json:"tags,omitempty" yaml:"tags,omitempty"`
			}{
				Content: memory.Content,
				Level:   memory.Level,
				Tags:    memory.Tags,
			})
		}

		cfg.Nodes = append(cfg.Nodes, nodeCfg)
	}

	for _, relation := range snapshot.Relations {
		sourceName, ok1 := nameByID[relation.SourceID]
		targetName, ok2 := nameByID[relation.TargetID]
		if !ok1 || !ok2 {
			continue
		}
		cfg.Relations = append(cfg.Relations, sdk.RelationConfig{
			Source: sourceName,
			Target: targetName,
			Type:   relation.RelationType,
			Weight: relation.Weight,
			Props:  relation.Properties,
		})
	}

	sort.Slice(cfg.Nodes, func(i, j int) bool {
		if cfg.Nodes[i].Type == cfg.Nodes[j].Type {
			return cfg.Nodes[i].Name < cfg.Nodes[j].Name
		}
		return cfg.Nodes[i].Type < cfg.Nodes[j].Type
	})
	sort.Slice(cfg.Components, func(i, j int) bool {
		if cfg.Components[i].NodeID == cfg.Components[j].NodeID {
			return cfg.Components[i].Type < cfg.Components[j].Type
		}
		return cfg.Components[i].NodeID < cfg.Components[j].NodeID
	})
	sort.Slice(cfg.Relations, func(i, j int) bool {
		if cfg.Relations[i].Source == cfg.Relations[j].Source {
			if cfg.Relations[i].Target == cfg.Relations[j].Target {
				return cfg.Relations[i].Type < cfg.Relations[j].Type
			}
			return cfg.Relations[i].Target < cfg.Relations[j].Target
		}
		return cfg.Relations[i].Source < cfg.Relations[j].Source
	})

	return cfg
}
