package sdk

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ImportConfig 描述 SDK 导入文件的完整结构。
type ImportConfig struct {
	World      WorldConfig       `json:"world" yaml:"world"`
	Nodes      []NodeConfig      `json:"nodes" yaml:"nodes"`
	Components []ComponentConfig `json:"components" yaml:"components"`
	Relations  []RelationConfig  `json:"relations" yaml:"relations"`
}

// WorldConfig 描述导入文件中的世界信息。
type WorldConfig struct {
	Name string `json:"name" yaml:"name"`
}

// NodeConfig 描述导入文件中的节点定义。
type NodeConfig struct {
	Name     string `json:"name" yaml:"name"`
	Type     string `json:"type" yaml:"type"`
	Parent   string `json:"parent,omitempty" yaml:"parent,omitempty"`
	Profile  string `json:"profile,omitempty" yaml:"profile,omitempty"`
	Lore     string `json:"lore,omitempty" yaml:"lore,omitempty"`
	Memories []struct {
		Content string `json:"content" yaml:"content"`
		Level   string `json:"level,omitempty" yaml:"level,omitempty"`
		Tags    string `json:"tags,omitempty" yaml:"tags,omitempty"`
	} `json:"memories,omitempty" yaml:"memories,omitempty"`
}

// ComponentConfig 描述导入文件中的独立组件定义。
type ComponentConfig struct {
	NodeID string `json:"node_id" yaml:"node_id"`
	Type   string `json:"type" yaml:"type"`
	Data   string `json:"data" yaml:"data"`
}

// RelationConfig 描述导入文件中的关系定义。
type RelationConfig struct {
	Source string `json:"source" yaml:"source"`
	Target string `json:"target" yaml:"target"`
	Type   string `json:"type" yaml:"type"`
	Weight int    `json:"weight,omitempty" yaml:"weight,omitempty"`
	Props  string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

// ImportYAML 从 YAML 数据导入世界配置。
func (a *Agent) ImportYAML(data []byte) error {
	var cfg ImportConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}
	return a.importConfig(&cfg)
}

// ImportJSON 从 JSON 数据导入世界配置。
func (a *Agent) ImportJSON(data []byte) error {
	var cfg ImportConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse json: %w", err)
	}
	return a.importConfig(&cfg)
}

// ImportFile 根据文件扩展名选择导入 YAML 或 JSON。
func (a *Agent) ImportFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("empty file")
	}
	if path[len(path)-5:] == ".yaml" || path[len(path)-4:] == ".yml" {
		return a.ImportYAML(data)
	}
	return a.ImportJSON(data)
}

// importConfig 将解析后的导入配置真正写入远端服务。
func (a *Agent) importConfig(cfg *ImportConfig) error {
	if _, err := a.CreateWorld(cfg.World.Name); err != nil {
		return fmt.Errorf("world: %w", err)
	}
	nodeMap := map[string]string{
		"world":        a.worldID,
		cfg.World.Name: a.worldID,
	}

	// 第一轮先创建所有节点，建立名称到 ID 的映射。
	for _, n := range cfg.Nodes {
		parentID := ""
		if n.Parent != "" {
			if pid, ok := nodeMap[n.Parent]; ok {
				parentID = pid
			}
		}
		id, err := a.client.CreateNode(a.worldID, n.Name, n.Type, parentID)
		if err != nil {
			return fmt.Errorf("create node %s: %w", n.Name, err)
		}
		nodeMap[n.Name] = id

		if n.Profile != "" {
			a.client.AddComponent(id, "profile", n.Profile)
		}
		if n.Lore != "" {
			a.client.AddComponent(id, "lore", n.Lore)
		}
		for _, m := range n.Memories {
			level := m.Level
			if level == "" {
				level = "long_term"
			}
			if _, err := a.client.AddMemory(id, m.Content, level, m.Tags); err != nil {
				return fmt.Errorf("create memory for %s: %w", n.Name, err)
			}
		}
	}

	for _, c := range cfg.Components {
		nodeID, ok := nodeMap[c.NodeID]
		if !ok {
			nodeID = c.NodeID
		}
		if _, err := a.client.AddComponent(nodeID, c.Type, c.Data); err != nil {
			return fmt.Errorf("create component %s: %w", c.Type, err)
		}
	}

	// 最后再补建关系，避免出现前置节点未创建的问题。
	for _, r := range cfg.Relations {
		srcID, ok1 := nodeMap[r.Source]
		tgtID, ok2 := nodeMap[r.Target]
		if !ok1 || !ok2 {
			continue
		}
		if _, err := a.client.CreateRelationWithProps(a.worldID, srcID, tgtID, r.Type, r.Weight, r.Props); err != nil {
			return fmt.Errorf("create relation %s -> %s: %w", r.Source, r.Target, err)
		}
	}
	return nil
}
