package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// impNode 描述导入文件中的节点定义。
type impNode struct {
	Name    string `yaml:"name" json:"name"`
	Type    string `yaml:"type" json:"type"`
	Parent  string `yaml:"parent,omitempty" json:"parent,omitempty"`
	Profile string `yaml:"profile,omitempty" json:"profile,omitempty"`
}

// impRelation 描述导入文件中的关系定义。
type impRelation struct {
	Source string `yaml:"source" json:"source"`
	Target string `yaml:"target" json:"target"`
	Type   string `yaml:"type" json:"type"`
	Weight int    `yaml:"weight" json:"weight"`
}

// impConfig 描述 CLI 导入命令使用的配置结构。
type impConfig struct {
	World     string        `yaml:"world" json:"world"`
	Nodes     []impNode     `yaml:"nodes" json:"nodes"`
	Relations []impRelation `yaml:"relations" json:"relations"`
}

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import world config into database",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config.Init(cfgFile)
		store.ConfigureLogSink(store.LogSinkOptions{
			Enabled:       config.Global.Database.LogBatchEnabled,
			BatchSize:     config.Global.Database.LogBatchSize,
			FlushInterval: time.Duration(config.Global.Database.LogBatchFlushMs) * time.Millisecond,
			QueueSize:     config.Global.Database.LogBatchQueueSize,
		})
		store.ConfigureWriteRetry(store.WriteRetryOptions{
			Enabled:     config.Global.Database.WriteRetryEnabled,
			MaxAttempts: config.Global.Database.WriteRetryMaxAttempts,
			BaseDelay:   time.Duration(config.Global.Database.WriteRetryBaseDelayMs) * time.Millisecond,
			MaxDelay:    time.Duration(config.Global.Database.WriteRetryMaxDelayMs) * time.Millisecond,
		})
		store.Init(config.Global.Database.Driver, config.Global.Database.DSN)
		defer store.CloseLogSink()
		data, _ := os.ReadFile(args[0])
		var cfg impConfig
		if len(data) > 0 && data[0] == '{' {
			json.Unmarshal(data, &cfg)
		} else {
			yaml.Unmarshal(data, &cfg)
		}
		worldUUID := store.NewUUID()
		if cfg.World == "" {
			cfg.World = "default"
		}
		store.CreateNode(&store.NodeModel{UUID: worldUUID, WorldUUID: worldUUID, Name: cfg.World, NodeType: "world"})
		uuids := map[string]string{cfg.World: worldUUID}
		for _, n := range cfg.Nodes {
			nid := store.NewUUID()
			pid := worldUUID
			if n.Parent != "" {
				if p, ok := uuids[n.Parent]; ok {
					pid = p
				}
			}
			store.CreateNode(&store.NodeModel{UUID: nid, WorldUUID: worldUUID, Name: n.Name, NodeType: n.Type, ParentUUID: &pid})
			uuids[n.Name] = nid
			if n.Profile != "" {
				store.CreateComponent(&store.ComponentModel{UUID: store.NewUUID(), NodeUUID: nid, ComponentType: "profile", Data: n.Profile})
			}
		}
		for _, r := range cfg.Relations {
			srcUUID, ok1 := uuids[r.Source]
			tgtUUID, ok2 := uuids[r.Target]
			if ok1 && ok2 {
				store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldUUID: worldUUID, SourceUUID: srcUUID, TargetUUID: tgtUUID, RelationType: r.Type, Weight: r.Weight})
			}
		}
		fmt.Printf("Imported %d nodes, %d relations\n", len(cfg.Nodes), len(cfg.Relations))
	},
}
