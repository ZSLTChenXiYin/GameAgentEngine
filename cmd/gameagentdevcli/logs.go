package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type inferenceLogRequestView struct {
	TaskType          string `json:"task_type,omitempty"`
	MessageCount      int    `json:"message_count,omitempty"`
	PipelineMode      string `json:"pipeline_mode,omitempty"`
	MaxAnalysisRounds int    `json:"max_analysis_rounds,omitempty"`
	MaxDepth          int    `json:"max_depth,omitempty"`
	MemoryLimit       int    `json:"memory_limit,omitempty"`
}

type inferenceLogResponseView struct {
	ExecutionMode          string   `json:"execution_mode,omitempty"`
	ReplyPreview           string   `json:"reply_preview,omitempty"`
	ActionCount            int      `json:"action_count,omitempty"`
	MemoryUpdateCount      int      `json:"memory_update_count,omitempty"`
	ConfiguredPipelineMode string   `json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode  string   `json:"effective_pipeline_mode,omitempty"`
	MaxAnalysisRounds      int      `json:"max_analysis_rounds,omitempty"`
	RoundsUsed             int      `json:"rounds_used,omitempty"`
	ActionPreview          []string `json:"action_preview,omitempty"`
	MemoryPreview          []string `json:"memory_preview,omitempty"`
	DataRequestLabel       string   `json:"data_request_label,omitempty"`
	WorldChangePlanImpact  string   `json:"world_change_plan_impact,omitempty"`
}

func parseInferenceLogRequestData(raw string) *inferenceLogRequestView {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out inferenceLogRequestView
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return &out
}

func parseInferenceLogResponseData(raw string) *inferenceLogResponseView {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out inferenceLogResponseView
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return &out
}

func formatInferenceLogTime(value string) string {
	if value == "" {
		return "-"
	}
	ts, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return value
	}
	return ts.Local().Format("2006-01-02 15:04:05")
}

func summarizeInferenceLog(log sdk.InferenceLog) []string {
	lines := []string{fmt.Sprintf("[%s] %s %dms %d tokens", formatInferenceLogTime(log.CreatedAt), log.TaskType, log.DurationMs, log.TokensUsed)}
	lines = append(lines, fmt.Sprintf("  world=%s node=%s model=%s", shortID(log.WorldID), shortID(log.NodeID), log.LLMModel))
	if log.Category != "" || log.EventName != "" || log.LogLevel != "" {
		parts := make([]string, 0, 4)
		if log.Category != "" {
			parts = append(parts, "category="+log.Category)
		}
		if log.EventName != "" {
			parts = append(parts, "event="+log.EventName)
		}
		if log.LogLevel != "" {
			parts = append(parts, "level="+log.LogLevel)
		}
		if log.ExecutionMode != "" {
			parts = append(parts, "mode="+log.ExecutionMode)
		}
		lines = append(lines, "  "+strings.Join(parts, " "))
	}
	if log.Message != "" {
		lines = append(lines, "  message="+log.Message)
	}

	request := parseInferenceLogRequestData(log.RequestData)
	if request != nil {
		parts := make([]string, 0, 4)
		if request.PipelineMode != "" {
			parts = append(parts, "request_pipeline="+request.PipelineMode)
		}
		if request.MessageCount > 0 {
			parts = append(parts, fmt.Sprintf("messages=%d", request.MessageCount))
		}
		if request.MaxAnalysisRounds > 0 {
			parts = append(parts, fmt.Sprintf("request_round_limit=%d", request.MaxAnalysisRounds))
		}
		if len(parts) > 0 {
			lines = append(lines, "  "+strings.Join(parts, " "))
		}
	}

	response := parseInferenceLogResponseData(log.ResponseData)
	if response != nil {
		parts := make([]string, 0, 6)
		if response.ExecutionMode != "" {
			parts = append(parts, "mode="+response.ExecutionMode)
		}
		if response.ConfiguredPipelineMode != "" || response.EffectivePipelineMode != "" {
			parts = append(parts, "pipeline="+strings.Trim(strings.TrimSpace(response.ConfiguredPipelineMode+" -> "+response.EffectivePipelineMode), " ->"))
		}
		if response.MaxAnalysisRounds > 0 || response.RoundsUsed > 0 {
			parts = append(parts, fmt.Sprintf("rounds=%d/%d", response.RoundsUsed, response.MaxAnalysisRounds))
		}
		if response.ActionCount > 0 {
			parts = append(parts, fmt.Sprintf("actions=%d", response.ActionCount))
		}
		if response.MemoryUpdateCount > 0 {
			parts = append(parts, fmt.Sprintf("memories=%d", response.MemoryUpdateCount))
		}
		if response.DataRequestLabel != "" {
			parts = append(parts, "data_request="+response.DataRequestLabel)
		}
		if len(parts) > 0 {
			lines = append(lines, "  "+strings.Join(parts, " "))
		}
		if response.WorldChangePlanImpact != "" {
			lines = append(lines, "  impact="+response.WorldChangePlanImpact)
		}
		if response.ReplyPreview != "" {
			lines = append(lines, "  reply="+response.ReplyPreview)
		}
		if len(response.ActionPreview) > 0 {
			lines = append(lines, "  action_preview="+strings.Join(response.ActionPreview, ", "))
		}
		if len(response.MemoryPreview) > 0 {
			lines = append(lines, "  memory_preview="+strings.Join(response.MemoryPreview, ", "))
		}
	}

	return lines
}

func printInferenceLogSummary(logs []sdk.InferenceLog) {
	if len(logs) == 0 {
		fmt.Println("No logs found.")
		return
	}
	for _, log := range logs {
		for _, line := range summarizeInferenceLog(log) {
			fmt.Println(line)
		}
		fmt.Println()
	}
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "读取最近的推理日志",
	Run: func(cmd *cobra.Command, args []string) {
		client := newClient()
		worldID, _ := cmd.Flags().GetString("world")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		taskType, _ := cmd.Flags().GetString("task-type")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		logs, err := client.GetLogs(worldID, limit, offset, taskType)
		if err != nil {
			fail(err)
		}
		if jsonOutput {
			printJSON(logs)
			return
		}
		printInferenceLogSummary(logs)
	},
}

func init() {
	logsCmd.Flags().String("world", "", "按世界 ID 过滤日志")
	logsCmd.Flags().Int("limit", 20, "返回日志条数")
	logsCmd.Flags().Int("offset", 0, "偏移量")
	logsCmd.Flags().String("task-type", "", "按任务类型过滤（如 npc_dialogue, world_tick）")
	logsCmd.Flags().Bool("json", false, "输出原始 JSON")
}
