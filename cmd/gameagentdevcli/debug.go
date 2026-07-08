package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func shortID(value string) string {
	if len(value) <= 8 {
		return value
	}
	return value[:8]
}

func formatTraceTimestamp(value string) string {
	if value == "" {
		return "-"
	}
	ts, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}
	return ts.Local().Format("2006-01-02 15:04:05")
}

func printDebugTraceSummary(payload *sdk.DebugTraceList) {
	if payload == nil || len(payload.Traces) == 0 {
		fmt.Println("No traces found.")
		return
	}
	fmt.Printf("Traces: %d\n\n", payload.Count)
	for _, trace := range payload.Traces {
		pipeline := "-"
		if trace.ConfiguredPipelineMode != "" || trace.EffectivePipelineMode != "" {
			pipeline = strings.TrimSpace(trace.ConfiguredPipelineMode + " -> " + trace.EffectivePipelineMode)
			pipeline = strings.Trim(pipeline, " ->")
		}
		rounds := "-"
		if trace.MaxAnalysisRounds > 0 || trace.RoundsUsed > 0 {
			rounds = fmt.Sprintf("%d/%d", trace.RoundsUsed, trace.MaxAnalysisRounds)
		}
		fmt.Printf("[%s] %s %dms\n", shortID(trace.ID), trace.TaskType, trace.DurationMs)
		fmt.Printf("  world=%s node=%s request=%s\n", shortID(trace.WorldID), shortID(trace.NodeID), shortID(trace.RequestID))
		fmt.Printf("  time=%s pipeline=%s rounds=%s\n", formatTraceTimestamp(trace.Timestamp), pipeline, rounds)
		if trace.Error != "" {
			fmt.Printf("  error=%s\n", trace.Error)
		}
		fmt.Println()
	}
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug utilities",
}

var debugTracesCmd = &cobra.Command{
	Use:   "traces",
	Short: "Show recent LLM debug traces",
	Run: func(cmd *cobra.Command, args []string) {
		worldID, _ := cmd.Flags().GetString("world")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		payload, err := newClient().GetDebugTraces(worldID, limit)
		if err != nil {
			fail(err)
		}
		if jsonOutput {
			printJSON(payload)
			return
		}
		printDebugTraceSummary(payload)
	},
}

func init() {
	debugCmd.AddCommand(debugTracesCmd)
	debugTracesCmd.Flags().String("world", "", "Filter by world ID")
	debugTracesCmd.Flags().Int("limit", 20, "Maximum number of traces to return")
	debugTracesCmd.Flags().Bool("json", false, "Print raw JSON instead of the summary view")
}
