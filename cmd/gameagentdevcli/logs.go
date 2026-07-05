package main

import (
    "github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
    Use:   "logs",
    Short: "读取最近的推理日志",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		worldID, _ := cmd.Flags().GetString("world")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		taskType, _ := cmd.Flags().GetString("task-type")
		logs, err := client.GetLogs(worldID, limit, offset, taskType)
		if err != nil {
			fail(err)
		}
        printJSON(logs)
    },
}

func init() {
	logsCmd.Flags().String("world", "", "按世界 ID 过滤日志")
	logsCmd.Flags().Int("limit", 20, "返回日志条数")
	logsCmd.Flags().Int("offset", 0, "偏移量")
	logsCmd.Flags().String("task-type", "", "按任务类型过滤（如 npc_dialogue, world_tick）")
}
