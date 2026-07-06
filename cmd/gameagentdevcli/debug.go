package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type debugTraceList struct {
	Traces []debugTrace `json:"traces"`
	Count  int          `json:"count"`
}

type debugTrace struct {
	ID                     string `json:"id"`
	WorldID                string `json:"world_id"`
	RequestID              string `json:"request_id"`
	TaskType               string `json:"task_type"`
	NodeID                 string `json:"node_id"`
	ConfiguredPipelineMode string `json:"configured_pipeline_mode"`
	EffectivePipelineMode  string `json:"effective_pipeline_mode"`
	MaxAnalysisRounds      int    `json:"max_analysis_rounds"`
	RoundsUsed             int    `json:"rounds_used"`
	Timestamp              string `json:"timestamp"`
	DurationMs             int64  `json:"duration_ms"`
	Error                  string `json:"error"`
}

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

func printDebugTraceSummary(payload *debugTraceList) {
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

		q := "/debug/traces"
		first := true
		if worldID != "" {
			if first {
				q += "?"
				first = false
			} else {
				q += "&"
			}
			q += "world_id=" + worldID
		}
		if limit > 0 {
			if first {
				q += "?"
				first = false
			} else {
				q += "&"
			}
			q += fmt.Sprintf("limit=%d", limit)
		}

		data, err := newClient().RawGet(q)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if jsonOutput {
			fmt.Println(string(data))
			return
		}

		var payload debugTraceList
		if err := json.Unmarshal(data, &payload); err != nil {
			fmt.Println(string(data))
			return
		}
		printDebugTraceSummary(&payload)
	},
}

func init() {
	debugCmd.AddCommand(debugTracesCmd)
	debugTracesCmd.Flags().String("world", "", "Filter by world ID")
	debugTracesCmd.Flags().Int("limit", 20, "Maximum number of traces to return")
	debugTracesCmd.Flags().Bool("json", false, "Print raw JSON instead of the summary view")
}
