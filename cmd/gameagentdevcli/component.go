package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "管理组件实例",
}

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出节点组件",
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("node")
		if nodeID == "" {
			fail(fmt.Errorf("--node is required"))
		}
		items, err := newClient().GetComponents(nodeID)
		if err != nil {
			fail(err)
		}
		printJSON(items)
	},
}

var componentGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "获取组件详情",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		item, err := newClient().GetComponent(args[0])
		if err != nil {
			fail(err)
		}
		printJSON(item)
	},
}

var componentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "创建组件",
	Run: func(cmd *cobra.Command, args []string) {
		nodeID, _ := cmd.Flags().GetString("node")
		componentType, _ := cmd.Flags().GetString("type")
		data, _ := cmd.Flags().GetString("data")
		if nodeID == "" || componentType == "" {
			fail(fmt.Errorf("--node and --type are required"))
		}
		if err := validateCLIComponentData(componentType, data); err != nil {
			fail(err)
		}
		id, err := newClient().AddComponent(nodeID, componentType, data)
		if err != nil {
			fail(err)
		}
		item, err := newClient().GetComponent(id)
		if err != nil {
			fail(err)
		}
		printJSON(item)
	},
}

var componentUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "更新组件",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		componentType, typeChanged := getChangedString(cmd, "type")
		data, dataChanged := getChangedString(cmd, "data")
		if !typeChanged && !dataChanged {
			fail(fmt.Errorf("no update flags provided"))
		}
		if dataChanged {
			typeForValidation := componentType
			if !typeChanged {
				item, err := newClient().GetComponent(args[0])
				if err != nil {
					fail(err)
				}
				typeForValidation = item.ComponentType
			}
			if err := validateCLIComponentData(typeForValidation, data); err != nil {
				fail(err)
			}
		}
		item, err := newClient().UpdateComponent(args[0], ptrIfChanged(componentType, typeChanged), ptrIfChanged(data, dataChanged))
		if err != nil {
			fail(err)
		}
		printJSON(item)
	},
}

var componentDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "删除组件",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := newClient().DeleteComponent(args[0]); err != nil {
			fail(err)
		}
		printJSON(map[string]string{"status": "deleted", "id": args[0]})
	},
}

func init() {
	componentCmd.AddCommand(componentListCmd, componentGetCmd, componentCreateCmd, componentUpdateCmd, componentDeleteCmd)

	componentListCmd.Flags().String("node", "", "节点 ID")
	componentCreateCmd.Flags().String("node", "", "节点 ID")
	componentCreateCmd.Flags().String("type", "", "组件类型")
	componentCreateCmd.Flags().String("data", "", "组件数据，建议传 JSON 字符串")
	componentUpdateCmd.Flags().String("type", "", "新的组件类型")
	componentUpdateCmd.Flags().String("data", "", "新的组件数据")
}
