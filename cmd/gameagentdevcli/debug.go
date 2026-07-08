package main

import (
	"encoding/json"
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

func summarizeDebugTrace(trace sdk.DebugTrace) []string {
	pipeline := "-"
	if trace.ConfiguredPipelineMode != "" || trace.EffectivePipelineMode != "" {
		pipeline = strings.TrimSpace(trace.ConfiguredPipelineMode + " -> " + trace.EffectivePipelineMode)
		pipeline = strings.Trim(pipeline, " ->")
	}
	rounds := "-"
	if trace.MaxAnalysisRounds > 0 || trace.RoundsUsed > 0 {
		rounds = fmt.Sprintf("%d/%d", trace.RoundsUsed, trace.MaxAnalysisRounds)
	}
	lines := []string{
		fmt.Sprintf("[%s] %s %dms", shortID(trace.ID), trace.TaskType, trace.DurationMs),
		fmt.Sprintf("  world=%s node=%s request=%s", shortID(trace.WorldID), shortID(trace.NodeID), shortID(trace.RequestID)),
		fmt.Sprintf("  time=%s pipeline=%s rounds=%s", formatTraceTimestamp(trace.Timestamp), pipeline, rounds),
	}
	if trace.Error != "" {
		lines = append(lines, "  error="+trace.Error)
	}
	return lines
}

func compactJSON(value any) string {
	if value == nil {
		return "-"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "-"
	}
	text := string(data)
	if len(text) > 180 {
		return text[:177] + "..."
	}
	return text
}

func summarizeStateComponent(item sdk.StateComponentEnvelope) string {
	status := "missing"
	if item.Component != nil {
		status = "present"
	}
	preview := compactJSON(item.Data)
	return fmt.Sprintf("- %s [%s] %s", item.ComponentType, status, preview)
}

func printContinuityBundleSummary(bundle *sdk.ContinuityBundle) {
	if bundle == nil {
		fmt.Println("No continuity bundle found.")
		return
	}
	fmt.Printf("World: %s\n\n", shortID(bundle.WorldID))

	fmt.Println("Latest Timeline")
	if bundle.LatestTimeline == nil {
		fmt.Println("- none")
		fmt.Println()
	} else {
		item := bundle.LatestTimeline
		fmt.Printf("- #%d %s %s\n", item.TickNumber, item.TickType, item.GameTime)
		if item.Summary != "" {
			fmt.Printf("  summary=%s\n", item.Summary)
		}
		if item.FutureOutline != "" {
			fmt.Printf("  future=%s\n", item.FutureOutline)
		}
		if item.Data != nil {
			fmt.Printf("  payload=%s\n", compactJSON(item.Data))
		}
		fmt.Println()
	}

	fmt.Println("State Components")
	if len(bundle.StateComponents) == 0 {
		fmt.Println("- none")
	} else {
		for _, item := range bundle.StateComponents {
			fmt.Println(summarizeStateComponent(item))
		}
	}
	fmt.Println()

	fmt.Println("Recent Logs")
	if len(bundle.Logs) == 0 {
		fmt.Println("- none")
	} else {
		for _, log := range bundle.Logs {
			for _, line := range summarizeInferenceLog(log) {
				fmt.Println(line)
			}
			fmt.Println()
		}
	}

	fmt.Println("Recent Traces")
	if len(bundle.Traces) == 0 {
		fmt.Println("- none")
	} else {
		for _, trace := range bundle.Traces {
			for _, line := range summarizeDebugTrace(trace) {
				fmt.Println(line)
			}
			fmt.Println()
		}
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

var debugContinuityCmd = &cobra.Command{
	Use:   "continuity <world-id>",
	Short: "Diagnose world tick continuity state, logs, and traces",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logLimit, _ := cmd.Flags().GetInt("log-limit")
		traceLimit, _ := cmd.Flags().GetInt("trace-limit")
		taskType, _ := cmd.Flags().GetString("task-type")
		nodeID, _ := cmd.Flags().GetString("node")
		category, _ := cmd.Flags().GetString("category")
		eventName, _ := cmd.Flags().GetString("event")
		executionMode, _ := cmd.Flags().GetString("mode")
		requestID, _ := cmd.Flags().GetString("request-id")
		round, _ := cmd.Flags().GetInt("round")
		skipLogs, _ := cmd.Flags().GetBool("skip-logs")
		skipTraces, _ := cmd.Flags().GetBool("skip-traces")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		bundle, err := newClient().GetContinuityBundle(args[0], &sdk.ContinuityBundleOptions{
			LogLimit:   logLimit,
			TraceLimit: traceLimit,
			SkipLogs:   skipLogs,
			SkipTraces: skipTraces,
			LogQuery: &sdk.InferenceLogQuery{
				WorldID:       args[0],
				NodeID:        nodeID,
				TaskType:      taskType,
				Category:      category,
				EventName:     eventName,
				ExecutionMode: executionMode,
				RequestID:     requestID,
				Round:         round,
				Limit:         logLimit,
			},
		})
		if err != nil {
			fail(err)
		}
		if jsonOutput {
			printJSON(bundle)
			return
		}
		printContinuityBundleSummary(bundle)
	},
}

func init() {
	debugCmd.AddCommand(debugTracesCmd, debugContinuityCmd)
	debugTracesCmd.Flags().String("world", "", "Filter by world ID")
	debugTracesCmd.Flags().Int("limit", 20, "Maximum number of traces to return")
	debugTracesCmd.Flags().Bool("json", false, "Print raw JSON instead of the summary view")
	debugContinuityCmd.Flags().Int("log-limit", 20, "Maximum number of logs to return")
	debugContinuityCmd.Flags().Int("trace-limit", 10, "Maximum number of traces to return")
	debugContinuityCmd.Flags().String("task-type", "world_tick", "Filter logs by task type")
	debugContinuityCmd.Flags().String("node", "", "Filter logs by node ID")
	debugContinuityCmd.Flags().String("category", "", "Filter logs by log category")
	debugContinuityCmd.Flags().String("event", "", "Filter logs by event name")
	debugContinuityCmd.Flags().String("mode", "", "Filter logs by execution mode")
	debugContinuityCmd.Flags().String("request-id", "", "Filter logs by request ID")
	debugContinuityCmd.Flags().Int("round", 0, "Filter logs by round number")
	debugContinuityCmd.Flags().Bool("skip-logs", false, "Skip loading recent logs")
	debugContinuityCmd.Flags().Bool("skip-traces", false, "Skip loading recent traces")
	debugContinuityCmd.Flags().Bool("json", false, "Print raw JSON instead of the summary view")
}
