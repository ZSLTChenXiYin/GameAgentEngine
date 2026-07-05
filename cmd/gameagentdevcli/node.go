package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "管理节点实例",
}

var nodeListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出节点",
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		nodeType, _ := cmd.Flags().GetString("type")
		nodes, err := newClient().GetNodes(worldID, limit, offset, nodeType)
		if err != nil {
			fail(err)
		}
		printJSON(nodes)
	},
}

var nodeLegacyListCmd = &cobra.Command{
	Use:   "nodes",
	Short: "列出所有节点",
	Run:   nodeListCmd.Run,
}

var nodeGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "获取节点详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		detail, err := newClient().GetNode(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(detail)
	},
}

var nodeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建节点",
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		name, _ := cmd.Flags().GetString("name")
		nodeType, _ := cmd.Flags().GetString("type")
		parentID, _ := cmd.Flags().GetString("parent")

		if name == "" || nodeType == "" {
			fail(fmt.Errorf("--name and --type are required"))
		}
		id, err := newClient().CreateNode(worldID, name, nodeType, parentID)
		if err != nil {
			fail(err)
		}
		detail, err := newClient().GetNode(id)
		if err != nil {
			fail(err)
		}
		printJSON(detail.Node)
	},
}

var nodeUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "更新节点",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name, nameChanged := getChangedString(cmd, "name")
		nodeType, typeChanged := getChangedString(cmd, "type")
		parentValue, parentChanged := getChangedString(cmd, "parent")
		clearParent, _ := cmd.Flags().GetBool("clear-parent")

		if !nameChanged && !typeChanged && !parentChanged && !clearParent {
			fail(fmt.Errorf("no update flags provided"))
		}

		var parentID *string
		if clearParent {
			empty := ""
			parentID = &empty
		} else if parentChanged {
			parentID = &parentValue
		}

		node, err := newClient().UpdateNode(args[0], valueIfChanged(name, nameChanged), valueIfChanged(nodeType, typeChanged), parentID)
		if err != nil {
			fail(err)
		}
		printJSON(node)
	},
}

var nodeDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "删除节点",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := newClient().DeleteNode(args[0]); err != nil {
			fail(err)
		}
		printJSON(map[string]string{"status": "deleted", "id": args[0]})
	},
}

var nodeAutonomousCmd = &cobra.Command{
	Use:   "autonomous",
	Short: "管理节点自主行为配置",
}

var nodeAutonomousGetCmd = &cobra.Command{
	Use:   "get <node-id>",
	Short: "查看节点自主行为配置",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := newClient().GetAutonomousConfig(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var nodeAutonomousSetCmd = &cobra.Command{
	Use:   "set <node-id>",
	Short: "创建或更新节点自主行为配置",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		enabled, _ := cmd.Flags().GetBool("enabled")
		trigger, _ := cmd.Flags().GetString("trigger")
		interval, _ := cmd.Flags().GetInt("interval")
		capabilitiesJSON, _ := cmd.Flags().GetString("capabilities")

		var capabilities []sdk.AgentCapability
		if capabilitiesJSON != "" {
			if err := json.Unmarshal([]byte(capabilitiesJSON), &capabilities); err != nil {
				fail(fmt.Errorf("invalid --capabilities json: %w", err))
			}
		}

		resp, err := newClient().SetAutonomousConfig(args[0], &sdk.AutonomousConfig{
			Enabled:         enabled,
			Trigger:         trigger,
			IntervalSeconds: interval,
			Capabilities:    capabilities,
		})
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var nodeAutonomousDisableCmd = &cobra.Command{
	Use:   "disable <node-id>",
	Short: "关闭节点自主行为",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		current, _ := newClient().GetAutonomousConfig(args[0])
		cfg := &sdk.AutonomousConfig{Enabled: false, Trigger: "manual"}
		if current != nil && current.Config != nil {
			cfg = current.Config
			cfg.Enabled = false
		}
		resp, err := newClient().SetAutonomousConfig(args[0], cfg)
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

var nodeAutonomousRunCmd = &cobra.Command{
	Use:   "run <node-id>",
	Short: "手动触发节点自主行为",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		if worldID == "" {
			detail, err := newClient().GetNode(args[0])
			if err != nil {
				fail(err)
			}
			worldID = detail.Node.WorldID
		}
		resp, err := newClient().RunAutonomousNode(worldID, args[0])
		if err != nil {
			fail(err)
		}
		printJSON(resp)
	},
}

func init() {
	nodeCmd.AddCommand(nodeListCmd, nodeGetCmd, nodeCreateCmd, nodeUpdateCmd, nodeDeleteCmd, nodeAutonomousCmd)
	nodeAutonomousCmd.AddCommand(nodeAutonomousGetCmd, nodeAutonomousSetCmd, nodeAutonomousDisableCmd, nodeAutonomousRunCmd)

	nodeListCmd.Flags().String("world", "", "世界 ID")
	nodeListCmd.Flags().Int("limit", 0, "返回条数上限（0 为不限制）")
	nodeListCmd.Flags().Int("offset", 0, "偏移量")
	nodeListCmd.Flags().String("type", "", "按节点类型过滤")

	nodeCreateCmd.Flags().String("world", "", "世界 ID；创建非 world 节点时必填")
	nodeCreateCmd.Flags().String("name", "", "节点名称")
	nodeCreateCmd.Flags().String("type", "", "节点类型")
	nodeCreateCmd.Flags().String("parent", "", "父节点 ID")

	nodeUpdateCmd.Flags().String("name", "", "新的节点名称")
	nodeUpdateCmd.Flags().String("type", "", "新的节点类型")
	nodeUpdateCmd.Flags().String("parent", "", "新的父节点 ID")
	nodeUpdateCmd.Flags().Bool("clear-parent", false, "清空父节点")

	nodeAutonomousSetCmd.Flags().Bool("enabled", false, "是否启用自主行为")
	nodeAutonomousSetCmd.Flags().String("trigger", "manual", "触发方式：manual / world_tick_sync / scheduled")
	nodeAutonomousSetCmd.Flags().Int("interval", 0, "scheduled 触发间隔秒数")
	nodeAutonomousSetCmd.Flags().String("capabilities", "[]", "能力白名单 JSON 数组")
	nodeAutonomousRunCmd.Flags().String("world", "", "世界 ID；为空时从节点详情推断")
}
