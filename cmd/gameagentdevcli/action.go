package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var actionCmd = &cobra.Command{
	Use:   "action",
	Short: "管理异步动作回调与调试操作",
}

var actionCallbackCmd = &cobra.Command{
	Use:   "callback <callback-id>",
	Short: "上报异步动作执行结果",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		status, _ := cmd.Flags().GetString("status")
		resultText, _ := cmd.Flags().GetString("result")
		if status == "" {
			fail(fmt.Errorf("--status is required"))
		}

		var result any
		if resultText != "" {
			if err := json.Unmarshal([]byte(resultText), &result); err != nil {
				result = resultText
			}
		}

		if err := newClient().ActionCallback(args[0], status, result); err != nil {
			fail(err)
		}
		printJSON(map[string]any{
			"status":      "ok",
			"callback_id": args[0],
			"reported":    status,
			"result":      result,
		})
	},
}

func init() {
	actionCmd.AddCommand(actionCallbackCmd)
	actionCallbackCmd.Flags().String("status", "success", "回调状态，例如 success / failed")
	actionCallbackCmd.Flags().String("result", "", "结果 JSON；如果不是合法 JSON，则按纯文本上报")
}
