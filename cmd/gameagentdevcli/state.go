package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var stateCmd = &cobra.Command{
	Use:   "state",
	Short: "管理 world tick 连续性状态组件",
}

var stateListCmd = &cobra.Command{
	Use:   "list <world-id>",
	Short: "列出世界的连续性状态组件",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := newClient().GetStateComponents(args[0])
		if err != nil {
			fail(err)
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			printJSON(result)
			return
		}
		fmt.Printf("World: %s\n\n", shortID(result.WorldID))
		for _, item := range result.Components {
			status := "missing"
			if item.Component != nil {
				status = "present"
			}
			fmt.Printf("- %s [%s]\n", item.ComponentType, status)
		}
	},
}

var stateGetCmd = &cobra.Command{
	Use:   "get <world-id> <component-type>",
	Short: "读取单个连续性状态组件",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := newClient().GetStateComponent(args[0], args[1])
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

var stateSetCmd = &cobra.Command{
	Use:   "set <world-id> <component-type>",
	Short: "写入单个连续性状态组件",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		payloadText, _ := cmd.Flags().GetString("data")
		filePath, _ := cmd.Flags().GetString("file")
		if payloadText == "" && filePath == "" {
			fail(fmt.Errorf("--data or --file is required"))
		}
		if payloadText != "" && filePath != "" {
			fail(fmt.Errorf("use either --data or --file, not both"))
		}
		if filePath != "" {
			data, err := os.ReadFile(filePath)
			if err != nil {
				fail(err)
			}
			payloadText = string(data)
		}
		payload, err := decodeJSONValue(payloadText)
		if err != nil {
			fail(fmt.Errorf("invalid JSON payload: %w", err))
		}
		result, err := newClient().PutStateComponent(args[0], args[1], payload)
		if err != nil {
			fail(err)
		}
		printJSON(result)
	},
}

func init() {
	stateCmd.AddCommand(stateListCmd, stateGetCmd, stateSetCmd)
	stateListCmd.Flags().Bool("json", false, "输出原始 JSON")
	stateSetCmd.Flags().String("data", "", "JSON 字符串形式的组件数据")
	stateSetCmd.Flags().String("file", "", "从文件读取 JSON 组件数据")
}
