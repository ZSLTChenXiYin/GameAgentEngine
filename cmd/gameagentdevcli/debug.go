package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "调试工具集",
}

var debugTracesCmd = &cobra.Command{
	Use:   "traces",
	Short: "查看最近的 LLM 推理轨迹",
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		limit, _ := cmd.Flags().GetInt("limit")

		q := "/debug/traces"
		first := true
		if worldID != "" {
			if first { q += "?"; first = false } else { q += "&" }
			q += "world_id=" + worldID
		}
		if limit > 0 {
			if first { q += "?"; first = false } else { q += "&" }
			q += fmt.Sprintf("limit=%d", limit)
		}

		data, err := newClient().RawGet(q)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(data))
	},
}

func init() {
	debugCmd.AddCommand(debugTracesCmd)
	debugTracesCmd.Flags().String("world", "", "按世界 ID 过滤")
	debugTracesCmd.Flags().Int("limit", 20, "最大返回条数")
}
