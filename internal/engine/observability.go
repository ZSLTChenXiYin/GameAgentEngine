package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type logDetail int

const (
	logDetailNone logDetail = iota
	logDetailSummary
	logDetailFull
)

type observabilityPolicy struct {
	consoleDetail logDetail
	dbDetail      logDetail
}

type pipelineLogEvent struct {
	Category     string
	EventName    string
	LogLevel     string
	Message      string
	Round        int
	RequestData  string
	ResponseData string
	DetailData   string
	DurationMs   int64
	TokensUsed   int
}

func observabilityPolicyForMode(mode ExecutionMode) observabilityPolicy {
	switch mode {
	case ModeDebug:
		return observabilityPolicy{consoleDetail: logDetailFull, dbDetail: logDetailFull}
	case ModeReview:
		return observabilityPolicy{consoleDetail: logDetailSummary, dbDetail: logDetailFull}
	default:
		return observabilityPolicy{consoleDetail: logDetailSummary, dbDetail: logDetailSummary}
	}
}

func marshalLogDetail(value any) string {
	if value == nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}

func buildFullRequestLogData(req *InvokeRequest) string {
	if req == nil {
		return ""
	}
	data, err := json.Marshal(req)
	if err != nil {
		return ""
	}
	return string(data)
}

func buildFullResponseLogData(resp *InvokeResponse) string {
	if resp == nil {
		return ""
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return ""
	}
	return string(data)
}

func summarizeLogMessage(message string) string {
	message = truncateForLog(message, 220)
	return strings.TrimSpace(message)
}

func (p *Pipeline) emitLog(req *InvokeRequest, resp *InvokeResponse, runtime *executionConfig, executionMode ExecutionMode, event pipelineLogEvent) {
	policy := observabilityPolicyForMode(executionMode)
	if event.LogLevel == "" {
		event.LogLevel = "info"
	}
	if event.Category == "" {
		event.Category = "engine"
	}
	if event.EventName == "" {
		event.EventName = "event"
	}

	if policy.consoleDetail != logDetailNone {
		message := summarizeLogMessage(event.Message)
		if message == "" {
			message = event.EventName
		}
		prefix := fmt.Sprintf("[%s][%s][%s]", executionMode, event.Category, event.EventName)
		if req != nil {
			prefix += fmt.Sprintf(" world=%s task=%s node=%s", req.WorldID, req.TaskType, req.NodeID)
		}
		if event.Round > 0 {
			prefix += fmt.Sprintf(" round=%d", event.Round)
		}
		if event.DurationMs > 0 {
			prefix += fmt.Sprintf(" duration=%dms", event.DurationMs)
		}
		if event.TokensUsed > 0 {
			prefix += fmt.Sprintf(" tokens=%d", event.TokensUsed)
		}
		log.Printf("%s %s", prefix, message)
		if policy.consoleDetail == logDetailFull && event.DetailData != "" {
			log.Printf("[%s][%s][detail] %s", executionMode, event.EventName, truncateForContext(event.DetailData, 4000))
		}
	}

	if policy.dbDetail == logDetailNone || req == nil {
		return
	}

	requestData := event.RequestData
	responseData := event.ResponseData
	detailData := event.DetailData
	if policy.dbDetail == logDetailSummary {
		if requestData == "" {
			requestData = buildInferenceLogRequestData(req)
		}
		if responseData == "" && resp != nil {
			responseData = buildInferenceLogResponseData(resp)
		}
		detailData = ""
	} else {
		if requestData == "" {
			requestData = buildFullRequestLogData(req)
		}
		if responseData == "" && resp != nil {
			responseData = buildFullResponseLogData(resp)
		}
	}

	model := &store.InferenceLogModel{
		WorldUUID:              req.WorldID,
		TaskType:               string(req.TaskType),
		NodeUUID:               req.NodeID,
		Category:               event.Category,
		EventName:              event.EventName,
		LogLevel:               event.LogLevel,
		Message:                summarizeLogMessage(event.Message),
		ExecutionMode:          string(executionMode),
		ConfiguredPipelineMode: configuredPipelineMode(runtime),
		EffectivePipelineMode:  effectivePipelineMode(runtime),
		Round:                  event.Round,
		RequestData:            requestData,
		ResponseData:           responseData,
		DetailData:             detailData,
		DurationMs:             event.DurationMs,
		TokensUsed:             event.TokensUsed,
	}
	if resp != nil {
		model.RequestID = resp.RequestID
	}
	if resp != nil && resp.Metadata != nil {
		model.LLMModel = resp.Metadata.LLMModel
		if event.DurationMs == 0 {
			model.DurationMs = resp.Metadata.ProcessingTimeMs
		}
		if event.TokensUsed == 0 {
			model.TokensUsed = resp.Metadata.TokensUsed
		}
	}
	if err := store.CreateInferenceLog(model); err != nil {
		log.Printf("[warn][logs] persist %s/%s failed: %v", event.Category, event.EventName, err)
	}
}

func buildContextLogDetail(ctx *BuiltContext, started time.Time) string {
	if ctx == nil {
		return ""
	}
	return marshalLogDetail(map[string]any{
		"node_id":         ctx.Node.UUID,
		"component_count": len(ctx.Components),
		"memory_count":    len(ctx.Memories),
		"relation_count":  len(ctx.Relations),
		"children_count":  len(ctx.Children),
		"ancestor_count":  len(ctx.Ancestors),
		"system_prompt":   truncateForContext(ctx.SystemPrompt, 4000),
		"built_at":        time.Since(started).Milliseconds(),
	})
}

func buildRoundLogDetail(systemPrompt string, messages []ChatMessage, round int, targetNodeID string, taskTree *TaskTree) string {
	data := map[string]any{
		"round":          round,
		"target_node_id": targetNodeID,
		"system_prompt":  systemPrompt,
		"messages":       messages,
	}
	if taskTree != nil {
		data["task_tree"] = taskTree
	}
	return marshalLogDetail(data)
}

func buildLLMResponseDetail(raw string, parsed *llmParsedOutput) string {
	data := map[string]any{
		"raw_response": raw,
	}
	if parsed != nil {
		data["cleaned_response"] = parsed.CleanedContent
		data["parse_error"] = parsed.ParseError
		data["reply"] = parsed.Reply
		data["action_calls"] = parsed.RawActionCalls
		data["memory_updates"] = parsed.RawMemoryUpdates
		data["world_change_plan"] = parsed.RawPlan
		data["request_data"] = parsed.RawRequestData
		data["interim_memory_updates"] = parsed.RawInterimMemoryUpdates
		data["future_outline"] = parsed.RawFutureOutline
		data["sub_tasks"] = parsed.RawSubTasks
	}
	return marshalLogDetail(data)
}

func buildResponseOutcomeDetail(resp *InvokeResponse) string {
	if resp == nil {
		return ""
	}
	return marshalLogDetail(resp)
}

func (p *Pipeline) emitExecutionEvent(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, eventName string, message string, detail any) {
	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "engine_execution",
		EventName:  eventName,
		Message:    message,
		DetailData: marshalLogDetail(detail),
	})
}
