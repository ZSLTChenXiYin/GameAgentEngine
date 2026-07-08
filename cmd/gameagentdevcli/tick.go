package main

import (
	"github.com/spf13/cobra"
)

var tickCmd = &cobra.Command{
	Use:   "tick <world-id>",
	Short: "推进世界时间刻度（兼容入口）",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		tickType, _ := cmd.Flags().GetString("type")
		gameTime, _ := cmd.Flags().GetString("time")
		var requestedTicks *int
		if cmd.Flags().Changed("requested-ticks") {
			v, _ := cmd.Flags().GetInt("requested-ticks")
			requestedTicks = &v
		}
		var limit *int
		if cmd.Flags().Changed("autonomous-limit") {
			v, _ := cmd.Flags().GetInt("autonomous-limit")
			limit = &v
		}
		tr, err := client.AdvanceTickWithOptions(args[0], tickType, gameTime, requestedTicks, limit)
		if err != nil {
			fail(err)
		}
		printJSON(tr)
	},
}

func init() {
	tickCmd.Flags().String("type", "manual", "刻推进类型")
	tickCmd.Flags().String("time", "dev-cli", "游戏内时间标识")
	tickCmd.Flags().Int("requested-ticks", 1, "本次 world tick 请求推进的基础 tick 数量")
	tickCmd.Flags().Int("autonomous-limit", 10, "本次 tick 最多触发的 world_tick_sync 自主节点数；0 为不触发")
}
