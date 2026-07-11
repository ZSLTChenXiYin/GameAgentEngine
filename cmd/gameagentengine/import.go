package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

// impNode 描述导入文件中的节点定义。
type impNode struct {
	Name    string      `yaml:"name" json:"name"`
	Type    string      `yaml:"type" json:"type"`
	Parent  string      `yaml:"parent,omitempty" json:"parent,omitempty"`
	Profile string      `yaml:"profile,omitempty" json:"profile,omitempty"`
	Lore    string      `yaml:"lore,omitempty" json:"lore,omitempty"`
	Memories []impMemory `yaml:"memories,omitempty" json:"memories,omitempty"`
}

// impMemory 描述导入文件中的记忆定义。
type impMemory struct {
	Content string `yaml:"content" json:"content"`
	Level   string `yaml:"level,omitempty" json:"level,omitempty"`
	Tags    string `yaml:"tags,omitempty" json:"tags,omitempty"`
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
	Short: "Import world config into database through the validated service layer",
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
		store.ConfigureMigrationsEnabled(config.Global.Database.MigrationsEnabled)
		service.ConfigureWorldLocks(config.Global.Engine.WorldLockEnabled)
		store.Init(config.Global.Database.Driver, config.Global.Database.DSN)
		defer store.CloseLogSink()

		data, err := os.ReadFile(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Read file: %v\n", err)
			os.Exit(1)
		}

		var legacy impConfig
		if len(data) > 0 && data[0] == '{' {
			if err := json.Unmarshal(data, &legacy); err != nil {
				fmt.Fprintf(os.Stderr, "Parse JSON: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := yaml.Unmarshal(data, &legacy); err != nil {
				fmt.Fprintf(os.Stderr, "Parse YAML: %v\n", err)
				os.Exit(1)
			}
		}
		if legacy.World == "" {
			legacy.World = "default"
		}

		// Convert legacy config to SDK ImportConfig and delegate to service layer
		sdkCfg := sdk.ImportConfig{
			World: sdk.WorldConfig{Name: legacy.World},
		}
		for _, n := range legacy.Nodes {
			nodeCfg := sdk.NodeConfig{
				Name:    n.Name,
				Type:    n.Type,
				Parent:  n.Parent,
				Profile: n.Profile,
				Lore:    n.Lore,
			}
			for _, m := range n.Memories {
				nodeCfg.Memories = append(nodeCfg.Memories, struct {
					Content string `json:"content" yaml:"content"`
					Level   string `json:"level,omitempty" yaml:"level,omitempty"`
					Tags    string `json:"tags,omitempty" yaml:"tags,omitempty"`
				}{Content: m.Content, Level: m.Level, Tags: m.Tags})
			}
			sdkCfg.Nodes = append(sdkCfg.Nodes, nodeCfg)
		}
		for _, r := range legacy.Relations {
			sdkCfg.Relations = append(sdkCfg.Relations, sdk.RelationConfig{
				Source: r.Source,
				Target: r.Target,
				Type:   r.Type,
				Weight: r.Weight,
			})
		}

		result, err := service.ImportWorld(&sdkCfg, false, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Imported world %q (%s): %d nodes, %d components, %d memories, %d relations\n",
			result.WorldName, result.WorldID,
			result.NodeCount, result.ComponentCount, result.MemoryCount, result.RelationCount)
	},
}
