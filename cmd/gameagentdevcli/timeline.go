package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

var timelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "查看 world tick 时间线归档",
}

func printTimelineSummary(items []sdk.TimelineEnvelope) {
	if len(items) == 0 {
		fmt.Println("No timelines found.")
		return
	}
	for _, item := range items {
		fmt.Printf("#%d %s %s\n", item.TickNumber, item.TickType, item.GameTime)
		if item.Summary != "" {
			fmt.Printf("  summary=%s\n", item.Summary)
		}
		if item.FutureOutline != "" {
			fmt.Printf("  future=%s\n", item.FutureOutline)
		}
		fmt.Printf("  created=%s\n\n", item.Timeline.CreatedAt)
	}
	}

var timelineListCmd = &cobra.Command{
	Use:   "list <world-id>",
	Short: "列出世界最近的时间线刻",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		result, err := newClient().GetTimelines(args[0], limit)
		if err != nil {
			fail(err)
		}
		if jsonOutput {
			printJSON(result)
			return
		}
		printTimelineSummary(result.Timelines)
	},
}

var timelineLatestCmd = &cobra.Command{
	Use:   "latest <world-id>",
	Short: "查看世界最新的时间线刻",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := newClient().GetLatestTimeline(args[0])
		if err != nil {
			fail(err)
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			printJSON(result)
			return
		}
		printTimelineSummary([]sdk.TimelineEnvelope{result.Timeline})
	},
}

func init() {
	timelineCmd.AddCommand(timelineListCmd, timelineLatestCmd)
	timelineListCmd.Flags().Int("limit", 10, "返回最近多少条时间线刻")
	timelineListCmd.Flags().Bool("json", false, "输出原始 JSON")
	timelineLatestCmd.Flags().Bool("json", false, "输出原始 JSON")
}
