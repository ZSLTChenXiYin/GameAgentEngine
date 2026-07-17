package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/action"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/external"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/planner"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type inferenceLogRequestData struct {
	WorldID             string        `json:"world_id,omitempty"`
	NodeID              string        `json:"node_id,omitempty"`
	TaskType            TaskType      `json:"task_type,omitempty"`
	MessageCount        int           `json:"message_count,omitempty"`
	IncludeRelatedNodes bool          `json:"include_related_nodes,omitempty"`
	MemoryLimit         int           `json:"memory_limit,omitempty"`
	MaxDepth            int           `json:"max_depth,omitempty"`
	MaxAnalysisRounds   int           `json:"max_analysis_rounds,omitempty"`
	PipelineMode        PipelineMode  `json:"pipeline_mode,omitempty"`
	Event               *WorldEvent   `json:"event,omitempty"`
	MessagePreview      []ChatMessage `json:"message_preview,omitempty"`
}

type inferenceLogResponseData struct {
	RequestID               string        `json:"request_id,omitempty"`
	ExecutionMode           ExecutionMode `json:"execution_mode,omitempty"`
	ReplyPreview            string        `json:"reply_preview,omitempty"`
	ActionCount             int           `json:"action_count,omitempty"`
	MemoryUpdateCount       int           `json:"memory_update_count,omitempty"`
	SubTaskCount            int           `json:"sub_task_count,omitempty"`
	HasWorldChangePlan      bool          `json:"has_world_change_plan,omitempty"`
	HasFutureOutline        bool          `json:"has_future_outline,omitempty"`
	HasDataRequest          bool          `json:"has_data_request,omitempty"`
	DataRequestLabel        string        `json:"data_request_label,omitempty"`
	ConfiguredPipelineMode  string        `json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode   string        `json:"effective_pipeline_mode,omitempty"`
	MaxAnalysisRounds       int           `json:"max_analysis_rounds,omitempty"`
	RoundsUsed              int           `json:"rounds_used,omitempty"`
	ActionPreview           []string      `json:"action_preview,omitempty"`
	MemoryPreview           []string      `json:"memory_preview,omitempty"`
	WorldChangePlanSummary  string        `json:"world_change_plan_summary,omitempty"`
	WorldChangePlanImpact   string        `json:"world_change_plan_impact,omitempty"`
	DataRequestQueryPreview []string      `json:"data_request_query_preview,omitempty"`
}

func truncateForLog(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit]) + "..."
}

func previewMessages(messages []ChatMessage, limit int) []ChatMessage {
	if len(messages) == 0 || limit <= 0 {
		return nil
	}
	count := len(messages)
	if count > limit {
		count = limit
	}
	result := make([]ChatMessage, 0, count)
	for i := 0; i < count; i++ {
		msg := messages[i]
		result = append(result, ChatMessage{Role: msg.Role, Content: truncateForLog(msg.Content, 280)})
	}
	return result
}

func buildInferenceLogRequestData(req *InvokeRequest) string {
	if req == nil {
		return ""
	}
	payload := inferenceLogRequestData{
		WorldID:        req.WorldID,
		NodeID:         req.NodeID,
		TaskType:       req.TaskType,
		MessageCount:   len(req.Messages),
		MessagePreview: previewMessages(req.Messages, 3),
		Event:          req.Event,
	}
	if req.Context != nil {
		payload.IncludeRelatedNodes = req.Context.IncludeRelatedNodes
		payload.MemoryLimit = req.Context.MemoryLimit
		payload.MaxDepth = req.Context.MaxDepth
		payload.MaxAnalysisRounds = req.Context.MaxAnalysisRounds
		payload.PipelineMode = req.Context.PipelineMode
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

func buildInferenceLogResponseData(resp *InvokeResponse) string {
	if resp == nil {
		return ""
	}
	payload := inferenceLogResponseData{
		RequestID:          resp.RequestID,
		ExecutionMode:      resp.ExecutionMode,
		ReplyPreview:       truncateForLog(resp.Reply, 400),
		ActionCount:        len(resp.ActionCalls),
		MemoryUpdateCount:  len(resp.MemoryUpdates),
		SubTaskCount:       len(resp.SubTasks),
		HasWorldChangePlan: resp.WorldChangePlan != nil,
		HasFutureOutline:   strings.TrimSpace(resp.FutureOutline) != "",
		HasDataRequest:     resp.DataRequest != nil,
	}
	if resp.Metadata != nil {
		payload.ConfiguredPipelineMode = resp.Metadata.ConfiguredPipelineMode
		payload.EffectivePipelineMode = resp.Metadata.EffectivePipelineMode
		payload.MaxAnalysisRounds = resp.Metadata.MaxAnalysisRounds
		payload.RoundsUsed = resp.Metadata.RoundsUsed
	}
	if resp.DataRequest != nil {
		payload.DataRequestLabel = resp.DataRequest.Label
		for _, q := range resp.DataRequest.Queries {
			label := q.Type
			if q.NodeID != "" {
				label += ":" + q.NodeID
			}
			if q.Filter != "" {
				label += "#" + q.Filter
			}
			payload.DataRequestQueryPreview = append(payload.DataRequestQueryPreview, label)
			if len(payload.DataRequestQueryPreview) >= 5 {
				break
			}
		}
	}
	for _, call := range resp.ActionCalls {
		preview := call.ActionID
		if call.Mode != "" {
			preview += "[" + call.Mode + "]"
		}
		payload.ActionPreview = append(payload.ActionPreview, preview)
		if len(payload.ActionPreview) >= 5 {
			break
		}
	}
	for _, mem := range resp.MemoryUpdates {
		payload.MemoryPreview = append(payload.MemoryPreview, mem.NodeID+":"+string(mem.Level))
		if len(payload.MemoryPreview) >= 5 {
			break
		}
	}
	if resp.WorldChangePlan != nil {
		payload.WorldChangePlanSummary = truncateForLog(resp.WorldChangePlan.Summary, 200)
		payload.WorldChangePlanImpact = resp.WorldChangePlan.ImpactLevel
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(data)
}

// Pipeline is the shared execution shell. Request-specific state is created per Execute call.
//
// 任务级关系装配策略约束：
// 1. npc_dialogue: 默认优先装配环境关系（located_at 及其场景祖先），再装配稳定身份/归属链，最后按需补社会关系。
// 2. autonomous_act: 默认优先装配环境关系和行动约束关系（belongs_to/subordinate 等），再按需补目标相关社会关系。
// 3. world_tick: 默认应围绕 world 或 scope 节点装配摘要性高价值关系图谱，禁止无差别展开所有 NPC 社交边。
// 4. world_event_impact: 默认应围绕事件 scope 装配局部关系子图，而不是直接使用全世界关系集合。
// 5. vertical/polling/full 的差异不仅是轮次，还包括关系子图装配强度；任何后续实现都不能忽略这一点。
type Pipeline struct {
	ctxBuilder  *ContextBuilder
	llmProvider LLMProvider
	actionReg   *action.Registry
	dispatcher  *external.Dispatcher
}

type executionConfig struct {
	memoryLimit            int
	maxRounds              int
	subTaskRetries         int
	subTaskTimeout         int
	configuredPipelineMode PipelineMode
	pipelineMode           PipelineMode
	policyEngine           *planner.PolicyEngine
	queryBudget            int   // E2: remaining allowed query rounds in current tick
	queryRoundLimit        int   // E2: max consecutive query rounds before forced convergence
}

func configuredPipelineMode(runtime *executionConfig) string {
	if runtime == nil {
		return ""
	}
	return string(runtime.configuredPipelineMode)
}

func effectivePipelineMode(runtime *executionConfig) string {
	if runtime == nil {
		return ""
	}
	return string(runtime.pipelineMode)
}

func reqInteractionContext(req *InvokeRequest) *InteractionContext {
	if req == nil || req.Context == nil {
		return nil
	}
	return req.Context.Interaction
}

func isPlayerInputInterpretRequest(req *InvokeRequest) bool {
	if req == nil {
		return false
	}
	if req.Context != nil && req.Context.PlayerInputInterpret {
		return true
	}
	for _, msg := range req.Messages {
		if strings.HasPrefix(strings.TrimSpace(msg.Content), "[player_input_interpret]") {
			return true
		}
	}
	return false
}

func stripPlayerInputInterpretPrefix(content string) string {
	trimmed := strings.TrimSpace(content)
	const marker = "[player_input_interpret]"
	if !strings.HasPrefix(trimmed, marker) {
		return trimmed
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, marker))
}

func normalizedPlayerInputMessages(messages []ChatMessage) []ChatMessage {
	if len(messages) == 0 {
		return nil
	}
	result := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		result[i] = msg
		result[i].Content = stripPlayerInputInterpretPrefix(msg.Content)
	}
	return result
}

func applyParsedPlayerIntent(resp *InvokeResponse, parsed *llmParsedOutput) *InvokeResponse {
	if resp == nil || parsed == nil || strings.TrimSpace(parsed.RawPlayerIntent) == "" {
		return resp
	}
	var payload PlayerIntentInterpretation
	if err := json.Unmarshal([]byte(parsed.RawPlayerIntent), &payload); err != nil {
		log.Printf("[player-intent] parse failed: %v", err)
		return resp
	}
	if err := ValidatePlayerIntentInterpretation(&payload); err != nil {
		log.Printf("[player-intent] validation failed: %v", err)
		return resp
	}
	resp.PlayerIntent = &payload
	return resp
}

func targetDialogueNodeID(req *InvokeRequest, ctx *BuiltContext) string {
	if ctx != nil && ctx.TargetNode != nil && strings.TrimSpace(ctx.TargetNode.UUID) != "" {
		return ctx.TargetNode.UUID
	}
	if interaction := reqInteractionContext(req); interaction != nil && strings.TrimSpace(interaction.TargetNodeID) != "" {
		return strings.TrimSpace(interaction.TargetNodeID)
	}
	if req == nil {
		return ""
	}
	return req.NodeID
}

func maxAnalysisRounds(runtime *executionConfig) int {
	if runtime == nil {
		return 0
	}
	return runtime.maxRounds
}

func buildResponseMeta(runtime *executionConfig, model string, tokens int, started time.Time, roundsUsed int) *ResponseMeta {
	return &ResponseMeta{
		LLMModel:               model,
		TokensUsed:             tokens,
		ProcessingTimeMs:       time.Since(started).Milliseconds(),
		ConfiguredPipelineMode: configuredPipelineMode(runtime),
		EffectivePipelineMode:  effectivePipelineMode(runtime),
		MaxAnalysisRounds:      maxAnalysisRounds(runtime),
		RoundsUsed:             roundsUsed,
	}
}

// NewPipeline creates a pipeline and registers built-in actions.
func NewPipeline(llmProvider LLMProvider) *Pipeline {
	p := &Pipeline{
		ctxBuilder:  NewContextBuilder(),
		llmProvider: llmProvider,
		actionReg:   action.NewRegistry(),
		dispatcher:  external.NewDispatcher(),
	}
	p.registerBuiltinActions()
	return p
}

func (p *Pipeline) externalDispatcher() *external.Dispatcher {
	if p.dispatcher == nil {
		p.dispatcher = external.NewDispatcher()
	}
	return p.dispatcher
}

func (p *Pipeline) registerBuiltinActions() {
	p.actionReg.RegisterSync(&action.UpdateMood{})
	p.actionReg.RegisterSync(&action.AddMemory{})
	p.actionReg.RegisterSync(&action.SendDialogue{})
	p.actionReg.RegisterAsync(&action.AdjustRelation{})
	p.actionReg.RegisterAsync(&action.SpawnItem{})
}

func (p *Pipeline) loadWorldSettings(worldID string) (int, int, int, int, string) {
	s, err := store.GetOrCreateWorldSettings(worldID)
	if err != nil {
		return 50, 5, 2, 60, string(PipelineFull)
	}
	memoryLimit := s.MemoryLimit
	if memoryLimit <= 0 {
		memoryLimit = 50
	}
	maxRounds := s.MaxAnalysisRounds
	if maxRounds <= 0 {
		maxRounds = 5
	}
	retries := s.SubTaskMaxRetries
	if retries < 0 {
		retries = 2
	}
	timeout := s.SubTaskTimeoutSecs
	if timeout <= 0 {
		timeout = 60
	}
	mode := s.PipelineMode
	if !IsValidPipelineMode(mode) {
		mode = string(PipelineFull)
	}
	return memoryLimit, maxRounds, retries, timeout, mode
}

func (p *Pipeline) loadWorldPolicy(worldID string) *planner.PolicyEngine {
	policyEngine := planner.NewPolicyEngine()
	policy, err := store.GetWorldPolicy(worldID)
	if err != nil {
		policyEngine.SetActions(nil, nil)
		return policyEngine
	}
	policyEngine.SetActions(policy.ParseBlockedActions(), policy.ParseSafeActions())
	return policyEngine
}

// ActionRegistry returns the shared action registry used for callbacks.
func (p *Pipeline) ActionRegistry() *action.Registry {
	return p.actionReg
}

func (p *Pipeline) getExecutionMode() ExecutionMode {
	switch em := config.ExecutionMode(); em {
	case "debug":
		return ModeDebug
	case "review":
		return ModeReview
	case "production":
		return ModeProduction
	default:
		return ModeProduction
	}
}

func (p *Pipeline) Execute(req *InvokeRequest) (*InvokeResponse, error) {
	// Execute 统一负责：读取世界设置、构建基础上下文、再按 PipelineMode 分发到不同执行路径。
	// 这里构建出的 BuiltContext 只是任务起点；更细的关系图谱扩展必须遵守上面的任务级策略，而不是默认拿全量关系。
	start := time.Now()
	requestID := uuid.NewString()
	traceID := "trace-" + requestID[:8]
	executionMode := p.getExecutionMode()
	p.emitLog(req, nil, nil, executionMode, pipelineLogEvent{
		Category:  "pipeline",
		EventName: "request_started",
		Message:   fmt.Sprintf("start task %s", req.TaskType),
		DetailData: marshalLogDetail(map[string]any{
			"request_id": requestID,
			"trace_id":    traceID,
			"request":    req,
		}),
	})

	depth := 3
	if req.Context != nil && req.Context.MaxDepth > 0 {
		depth = req.Context.MaxDepth
	}

	memoryLimit, maxRounds, retries, timeout, pipelineMode := p.loadWorldSettings(req.WorldID)
	configuredMode := PipelineMode(pipelineMode)
	if configuredMode == "" {
		configuredMode = PipelineFull
	}
	mode := configuredMode

	includeRelated := false
	if req.Context != nil {
		if req.Context.MemoryLimit > 0 {
			memoryLimit = req.Context.MemoryLimit
		}
		if req.Context.MaxAnalysisRounds > 0 {
			maxRounds = req.Context.MaxAnalysisRounds
		}
		if req.Context.PipelineMode != "" {
			mode = req.Context.PipelineMode
		}
		includeRelated = req.Context.IncludeRelatedNodes
	}

	runtime := &executionConfig{
		memoryLimit:            memoryLimit,
		maxRounds:              maxRounds,
		subTaskRetries:         retries,
		subTaskTimeout:         timeout,
		configuredPipelineMode: configuredMode,
		pipelineMode:           mode,
		policyEngine:           p.loadWorldPolicy(req.WorldID),
		queryBudget:            maxRounds,      // E2: budget = max rounds
		queryRoundLimit:        3,               // E2: force convergence after 3 consecutive query rounds
	}

	var interaction *InteractionContext
	if req.Context != nil {
		interaction = req.Context.Interaction
	}
	ctx, err := p.ctxBuilder.buildWithMode(req.TaskType, req.NodeID, depth, runtime.memoryLimit, includeRelated, interaction, runtime.pipelineMode)
	if err != nil {
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline",
			EventName:  "context_build_failed",
			LogLevel:   "error",
			Message:    err.Error(),
			DurationMs: time.Since(start).Milliseconds(),
		})
		return nil, fmt.Errorf("build context: %w", err)
	}
	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline",
		EventName:  "context_built",
		Message:    "context ready",
		DurationMs: time.Since(start).Milliseconds(),
		DetailData: buildContextLogDetail(ctx, start),
	})

	switch runtime.pipelineMode {
	case PipelineVertical:
		return p.executeVertical(req, start, requestID, runtime, executionMode)
	case PipelinePolling:
		return p.executePolling(req, ctx, start, requestID, runtime, executionMode)
	default:
		return p.executeFull(req, ctx, start, requestID, runtime, executionMode)
	}
}

// ResumePausedExecution restores a persisted paused execution snapshot and returns the decoded state.
// The actual round continuation is implemented separately so the callback path can reuse the same snapshot.
func (p *Pipeline) loadPausedExecutionForResume(callbackID string, result any) (*store.PausedExecutionModel, *InvokeRequest, *BuiltContext, *RoundState, *executionConfig, *DataRequest, ExecutionMode, int, error) {
	paused, err := store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, ModeProduction, 0, err
	}
	resumeJSON, err := json.Marshal(result)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, ModeProduction, 0, err
	}
	if err := store.MarkPausedExecutionResumed(paused.ExecutionID, string(resumeJSON)); err != nil {
		return nil, nil, nil, nil, nil, nil, ModeProduction, 0, err
	}
	req, ctx, state, runtime, dr, executionMode, pausedRound, err := decodePausedExecutionSnapshot(paused)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, ModeProduction, 0, err
	}
	return paused, req, ctx, state, runtime, dr, executionMode, pausedRound, nil
}

// ResumePausedExecution restores a paused execution, injects callback data into context,
// and continues the remaining multi-turn reasoning loop automatically.
func (p *Pipeline) ResumePausedExecution(callbackID string, result any) (*InvokeResponse, error) {
	paused, req, ctx, state, runtime, dr, executionMode, pausedRound, err := p.loadPausedExecutionForResume(callbackID, result)
	if err != nil {
		return nil, err
	}
	if state != nil && (state.PendingResumePhase == "sub_tasks" || state.PendingResumePhase == "sub_task_summary") {
		start := time.Now()
		resp, err := p.resumePendingSubTaskDAG(paused.ExecutionID, req, ctx, state, runtime, executionMode, paused.RequestID, pausedRound, result)
		if err != nil {
			_ = store.MarkPausedExecutionFailed(paused.ExecutionID, err.Error())
			p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
				Category:   "pipeline_resume",
				EventName:  "resume_failed",
				LogLevel:   "error",
				Message:    err.Error(),
				Round:      pausedRound,
				DetailData: marshalLogDetail(map[string]any{"callback_id": callbackID, "execution_id": paused.ExecutionID}),
			})
			return nil, err
		}
		if pausedResponse(resp) {
			return resp, nil
		}
		if err := store.MarkPausedExecutionCompleted(paused.ExecutionID); err != nil {
			return nil, err
		}
		p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
			Category:     "pipeline_resume",
			EventName:    "resume_completed",
			Message:      truncateForLog(resp.Reply, 180),
			Round:        resp.Metadata.RoundsUsed,
			ResponseData: buildFullResponseLogData(resp),
			DetailData:   marshalLogDetail(map[string]any{"callback_id": callbackID, "execution_id": paused.ExecutionID}),
			DurationMs:   time.Since(start).Milliseconds(),
		})
		return resp, nil
	}
	resolved := marshalLogDetail(map[string]any{
		"callback_id": callbackID,
		"result":      result,
	})
	if dr != nil {
		resolvedCache := ensureResolvedDataRequests(state)
		if resolvedCache != nil {
			if signature := dataRequestSignature(dr); signature != "" {
				resolvedCache[signature] = resolved
			}
		}
		label := dr.Label
		if label == "" {
			label = "game_client"
		}
		state.SupplementalContext = append(state.SupplementalContext, "[数据查询回填] "+label, resolved)
				if len(state.SupplementalContext) > 16 { state.SupplementalContext = state.SupplementalContext[len(state.SupplementalContext)-16:] }
		appendRoundStateTreeEntry(state, pausedRound, nil, resolved)
	}
	start := time.Now()
	resp, err := p.executeMultiTurnLoopFromState(req, ctx, start, paused.RequestID, runtime, state, pausedRound, executionMode)
	if err != nil {
		_ = store.MarkPausedExecutionFailed(paused.ExecutionID, err.Error())
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline_resume",
			EventName:  "resume_failed",
			LogLevel:   "error",
			Message:    err.Error(),
			Round:      pausedRound,
			DetailData: marshalLogDetail(map[string]any{"callback_id": callbackID, "execution_id": paused.ExecutionID}),
		})
		return nil, err
	}
	if err := store.MarkPausedExecutionCompleted(paused.ExecutionID); err != nil {
		return nil, err
	}
	p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
		Category:     "pipeline_resume",
		EventName:    "resume_completed",
		Message:      truncateForLog(resp.Reply, 180),
		Round:        resp.Metadata.RoundsUsed,
		ResponseData: buildFullResponseLogData(resp),
		DetailData:   marshalLogDetail(map[string]any{"callback_id": callbackID, "execution_id": paused.ExecutionID}),
		DurationMs:   time.Since(start).Milliseconds(),
	})
	return resp, nil
}

type RoundState struct {
	Context                      *BuiltContext
	Tree                         *TaskTree
	TreeContext                  string
	SystemPrompt                 string
	Messages                     []ChatMessage
	TargetNodeID                 string
	MaxRounds                    int
	SupplementalContext          []string
	ResolvedDataRequests         map[string]string
	PendingResumePhase           string
	PendingResponse              *InvokeResponse
	PendingMergedSubTaskResponse *InvokeResponse
	PendingSubTasks              []SubTaskDeclaration
	PendingSubTaskResults        map[string]*InvokeResponse
	PendingSubTaskFailed         map[string]string
	PendingSubTaskResume         *pausedSubTaskResume
	PendingSummaries             []pausedSummaryMerge
}

type pausedExecutionRuntime struct {
	ConfiguredPipelineMode string `json:"configured_pipeline_mode"`
	EffectivePipelineMode  string `json:"effective_pipeline_mode"`
	MaxRounds              int    `json:"max_rounds"`
	SubTaskRetries         int    `json:"sub_task_retries"`
	SubTaskTimeout         int    `json:"sub_task_timeout"`
	MemoryLimit            int    `json:"memory_limit"`
	ExecutionMode          string `json:"execution_mode"`
}

type pausedExecutionSnapshot struct {
	RoundState  *RoundState            `json:"round_state"`
	Runtime     pausedExecutionRuntime `json:"runtime"`
	DataRequest *DataRequest           `json:"data_request,omitempty"`
	PausedRound int                    `json:"paused_round"`
	CallbackID  string                 `json:"callback_id"`
	ExecutionID string                 `json:"execution_id"`
}

type pausedSubTaskResume struct {
	ExecutionID  string                 `json:"execution_id,omitempty"`
	CallbackID   string                 `json:"callback_id,omitempty"`
	Label        string                 `json:"label"`
	RequestID    string                 `json:"request_id,omitempty"`
	Request      *InvokeRequest         `json:"request,omitempty"`
	BuiltContext *BuiltContext          `json:"built_context,omitempty"`
	RoundState   *RoundState            `json:"round_state,omitempty"`
	Runtime      pausedExecutionRuntime `json:"runtime"`
	DataRequest  *DataRequest           `json:"data_request,omitempty"`
	PausedRound  int                    `json:"paused_round"`
}

type pausedSummaryMerge struct {
	Label      string   `json:"label"`
	StateLines []string `json:"state_lines,omitempty"`
}

func (s *RoundState) buildPrompt(base string) string {
	if len(s.SupplementalContext) == 0 {
		return base
	}
	return strings.TrimSpace(base + "\n\n补充上下文:\n" + strings.Join(s.SupplementalContext, "\n"))
}

func dataRequestSignature(dr *DataRequest) string {
	if dr == nil {
		return ""
	}
	type requestSignature struct {
		Label             string      `json:"label,omitempty"`
		Target            string      `json:"target,omitempty"`
		ExternalInterface string      `json:"external_interface,omitempty"`
		Queries           []DataQuery `json:"queries,omitempty"`
	}
	data, err := json.Marshal(requestSignature{
		Label:             dr.Label,
		Target:            dr.Target,
		ExternalInterface: dr.ExternalInterface,
		Queries:           dr.Queries,
	})
	if err != nil {
		return ""
	}
	return string(data)
}

func ensureResolvedDataRequests(state *RoundState) map[string]string {
	if state == nil {
		return nil
	}
	if state.ResolvedDataRequests == nil {
		state.ResolvedDataRequests = map[string]string{}
	}
	return state.ResolvedDataRequests
}

func mergeBaseAndTreeContext(baseContext, treeContext string) string {
	baseContext = strings.TrimSpace(baseContext)
	treeContext = strings.TrimSpace(treeContext)
	if treeContext == "" {
		return baseContext
	}
	if baseContext == "" {
		return treeContext
	}
	return strings.TrimSpace(baseContext + "\n\n任务树分析：\n" + treeContext)
}

func buildRoundStateTreeContext(state *RoundState) string {
	if state == nil {
		return ""
	}
	if strings.TrimSpace(state.TreeContext) != "" {
		return state.TreeContext
	}
	if state.Tree != nil {
		return state.Tree.BuildLLMContext()
	}
	return ""
}

func cloneInvokeResponse(resp *InvokeResponse) *InvokeResponse {
	if resp == nil {
		return nil
	}
	clone := *resp
	if resp.ActionCalls != nil {
		clone.ActionCalls = append([]ActionCall(nil), resp.ActionCalls...)
	}
	if resp.MemoryUpdates != nil {
		clone.MemoryUpdates = append([]MemoryUpdate(nil), resp.MemoryUpdates...)
	}
	if resp.SubTasks != nil {
		clone.SubTasks = append([]SubTaskDeclaration(nil), resp.SubTasks...)
	}
	return &clone
}

func cloneSubTaskResultsMap(src map[string]*InvokeResponse) map[string]*InvokeResponse {
	if len(src) == 0 {
		return nil
	}
	result := make(map[string]*InvokeResponse, len(src))
	for key, value := range src {
		result[key] = cloneInvokeResponse(value)
	}
	return result
}

func cloneSubTaskFailedMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	result := make(map[string]string, len(src))
	for key, value := range src {
		result[key] = value
	}
	return result
}

func clonePausedSummaries(src []pausedSummaryMerge) []pausedSummaryMerge {
	if len(src) == 0 {
		return nil
	}
	result := make([]pausedSummaryMerge, 0, len(src))
	for _, item := range src {
		clone := pausedSummaryMerge{Label: item.Label}
		if len(item.StateLines) > 0 {
			clone.StateLines = append([]string(nil), item.StateLines...)
		}
		result = append(result, clone)
	}
	return result
}

func pausedResponse(resp *InvokeResponse) bool {
	return resp != nil && resp.DataRequest != nil && len(resp.ActionCalls) > 0 && strings.TrimSpace(resp.ActionCalls[0].CallbackID) != ""
}

func clearPendingSubTaskState(state *RoundState) {
	if state == nil {
		return
	}
	state.PendingResumePhase = ""
	state.PendingResponse = nil
	state.PendingMergedSubTaskResponse = nil
	state.PendingSubTasks = nil
	state.PendingSubTaskResults = nil
	state.PendingSubTaskFailed = nil
	state.PendingSubTaskResume = nil
	state.PendingSummaries = nil
}

func buildParentPausedResponse(parentReq *InvokeRequest, parentRequestID string, executionMode ExecutionMode, base *InvokeResponse) *InvokeResponse {
	resp := cloneInvokeResponse(base)
	if resp == nil {
		resp = &InvokeResponse{}
	}
	resp.RequestID = parentRequestID
	resp.TaskType = parentReq.TaskType
	resp.ExecutionMode = executionMode
	return resp
}

func mergeSubTaskResultIntoParent(resp *InvokeResponse, merged *InvokeResponse, subTasks []SubTaskDeclaration) {
	if resp == nil || merged == nil {
		return
	}
	resp.SubTasks = append([]SubTaskDeclaration(nil), subTasks...)
	if len(merged.SubTasks) > 0 {
		resp.SubTasks = append([]SubTaskDeclaration(nil), merged.SubTasks...)
	}
	if merged.Reply != "" {
		resp.Reply = merged.Reply
	}
	resp.ActionCalls = append(resp.ActionCalls, merged.ActionCalls...)
	resp.MemoryUpdates = append(resp.MemoryUpdates, merged.MemoryUpdates...)
}

func reconstructDAGInstance(tree *TaskTree, llmProvider LLMProvider, runtime *executionConfig, subTasks []SubTaskDeclaration, results map[string]*InvokeResponse, failed map[string]string) *DAGInstance {
	retries := 0
	timeout := time.Duration(0)
	if runtime != nil {
		retries = runtime.subTaskRetries
		timeout = time.Duration(runtime.subTaskTimeout) * time.Second
	}
	dag := NewDAGInstance(tree, llmProvider, retries, timeout)
	for _, st := range subTasks {
		if err := dag.Register(st); err != nil {
			continue
		}
	}
	for label, resp := range results {
		dag.OnTaskComplete(label, cloneInvokeResponse(resp))
	}
	for label, errMsg := range failed {
		dag.OnTaskFailed(label, fmt.Errorf("%s", errMsg))
	}
	return dag
}

func executionModeFromSnapshot(snapshot pausedExecutionRuntime) ExecutionMode {
	switch snapshot.ExecutionMode {
	case string(ModeDebug):
		return ModeDebug
	case string(ModeReview):
		return ModeReview
	default:
		return ModeProduction
	}
}

func buildPausedSubTaskResume(callbackID string, label string) (*pausedSubTaskResume, error) {
	paused, err := store.GetPausedExecutionByCallbackID(callbackID)
	if err != nil {
		return nil, err
	}
	req, ctx, state, runtime, dr, executionMode, pausedRound, err := decodePausedExecutionSnapshot(paused)
	if err != nil {
		return nil, err
	}
	return &pausedSubTaskResume{
		ExecutionID:  paused.ExecutionID,
		CallbackID:   callbackID,
		Label:        label,
		RequestID:    paused.RequestID,
		Request:      req,
		BuiltContext: ctx,
		RoundState:   state,
		Runtime:      buildPausedExecutionRuntime(runtime, executionMode),
		DataRequest:  dr,
		PausedRound:  pausedRound,
	}, nil
}

func (p *Pipeline) overwritePausedExecutionSnapshot(executionID string, req *InvokeRequest, ctx *BuiltContext, state *RoundState, runtime *executionConfig, executionMode ExecutionMode, requestID string, pausedRound int, dr *DataRequest, callbackID string) error {
	if strings.TrimSpace(executionID) == "" {
		return fmt.Errorf("execution id required")
	}
	if state != nil {
		state.Tree = nil
		state.Context = nil
		state.TreeContext = buildRoundStateTreeContext(state)
	}
	originalReqJSON, err := json.Marshal(req)
	if err != nil {
		return err
	}
	builtContextJSON, err := json.Marshal(ctx)
	if err != nil {
		return err
	}
	roundStateJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}
	runtimeJSON, err := json.Marshal(buildPausedExecutionRuntime(runtime, executionMode))
	if err != nil {
		return err
	}
	dataRequestJSON := ""
	if dr != nil {
		if data, err := json.Marshal(dr); err == nil {
			dataRequestJSON = string(data)
		}
	}
	return store.UpdatePausedExecution(executionID, map[string]any{
		"request_id":                requestID,
		"world_uuid":                req.WorldID,
		"node_uuid":                 req.NodeID,
		"task_type":                 string(req.TaskType),
		"execution_mode":            string(executionMode),
		"configured_pipeline_mode":  configuredPipelineMode(runtime),
		"effective_pipeline_mode":   effectivePipelineMode(runtime),
		"status":                    "paused",
		"paused_round":              pausedRound,
		"max_rounds":                runtime.maxRounds,
		"target_node_id":            state.TargetNodeID,
		"pause_reason":              "game_client_request_data",
		"callback_id":               callbackID,
		"original_request_json":     string(originalReqJSON),
		"built_context_json":        string(builtContextJSON),
		"runtime_json":              string(runtimeJSON),
		"round_state_json":          string(roundStateJSON),
		"pending_data_request_json": dataRequestJSON,
		"resume_payload_json":       "",
		"last_error":                "",
		"resumed_at":                nil,
		"completed_at":              nil,
	})
}

func (p *Pipeline) applySubTaskMergeMode(merged *InvokeResponse, st SubTaskDeclaration, resp *InvokeResponse, summaryReply string) *InvokeResponse {
	if merged == nil {
		merged = &InvokeResponse{}
	}
	switch st.MergeMode {
	case "override":
		return cloneInvokeResponse(resp)
	case "summarize":
		merged.Reply = mergeStrings(merged.Reply, summaryReply, "\n")
		if resp != nil {
			merged.ActionCalls = append(merged.ActionCalls, resp.ActionCalls...)
			merged.MemoryUpdates = append(merged.MemoryUpdates, resp.MemoryUpdates...)
		}
		return merged
	default:
		if resp != nil {
			merged.ActionCalls = append(merged.ActionCalls, resp.ActionCalls...)
			merged.MemoryUpdates = append(merged.MemoryUpdates, resp.MemoryUpdates...)
			if merged.Reply == "" {
				merged.Reply = resp.Reply
			} else {
				merged.Reply = merged.Reply + "\n" + resp.Reply
			}
		}
		return merged
	}
}

func buildDAGFailureSummary(dag *DAGInstance) string {
	if dag == nil || len(dag.failed) == 0 {
		return ""
	}
	var errParts []string
	for _, label := range dag.orderedTaskLabels() {
		if errMsg := strings.TrimSpace(dag.failed[label]); errMsg != "" {
			errParts = append(errParts, fmt.Sprintf("[子任务 %s 失败] %s", label, errMsg))
		}
	}
	return strings.Join(errParts, "\n")
}

func (p *Pipeline) resumePendingSubTaskSummary(req *InvokeRequest, merged *InvokeResponse, dag *DAGInstance, pending pausedSummaryMerge, callbackResult any, runtime *executionConfig, executionMode ExecutionMode, round int) (*InvokeResponse, *DataRequest, *string, []string, error) {
	if dag == nil {
		return merged, nil, nil, nil, fmt.Errorf("dag required for pending summary resume")
	}
	stateLines := append([]string(nil), pending.StateLines...)
	resolved := marshalLogDetail(map[string]any{"result": callbackResult})
	label := pending.Label
	if label == "" {
		label = "game_client"
	}
	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:  "pipeline_round",
		EventName: "summarize_data_request_resolved_from_client",
		Message:   label,
		Round:     round,
		DetailData: marshalLogDetail(map[string]any{
			"label":           label,
			"source":          "game_client_callback",
			"callback_result": callbackResult,
		}),
	})
	stateLines = append(stateLines, "[summarize request_data resolved] "+label, resolved)
	return p.continueSubTaskSummaryMerge(req, merged, dag, pending.Label, stateLines, runtime, executionMode, round)
}

func (p *Pipeline) continueSubTaskSummaryMerge(req *InvokeRequest, merged *InvokeResponse, dag *DAGInstance, summaryLabel string, stateLines []string, runtime *executionConfig, executionMode ExecutionMode, round int) (*InvokeResponse, *DataRequest, *string, []string, error) {
	if dag == nil {
		return merged, nil, nil, nil, fmt.Errorf("dag required for summary merge")
	}
	mergedResp := cloneInvokeResponse(merged)
	if mergedResp == nil {
		mergedResp = &InvokeResponse{}
	}
	baseLines := dag.summarizeBaseLines()
	var normalizedState []string
	if len(stateLines) > 0 {
		normalizedState = append([]string(nil), stateLines...)
	}
	for _, label := range dag.orderedTaskLabels() {
		st := dag.tasks[label]
		resp := dag.results[label]
		if st == nil || resp == nil {
			continue
		}
		if summaryLabel != "" && label != summaryLabel {
			mergedResp = p.applySubTaskMergeMode(mergedResp, st.Decl, resp, "")
			continue
		}
		if st.Decl.MergeMode != "summarize" {
			mergedResp = p.applySubTaskMergeMode(mergedResp, st.Decl, resp, "")
			continue
		}
		currentState := normalizedState
		for round := 0; round < 4; round++ {
			roundLines := append([]string{}, baseLines...)
			if len(currentState) > 0 {
				roundLines = append(roundLines, "", "========== Summarize Context ==========")
				roundLines = append(roundLines, currentState...)
			}
			prompt := strings.Join(roundLines, "\n")
			llmResp, err := p.llmProvider.Chat(&LLMChatRequest{SystemPrompt: prompt})
			if err != nil {
				return nil, nil, nil, nil, err
			}
			parsed := p.parseLLMJSON(llmResp.Content)
			rawRequestData := strings.TrimSpace(parsed.RawRequestData)
			if rawRequestData != "" && rawRequestData != "null" {
				var dr DataRequest
				if err := json.Unmarshal([]byte(rawRequestData), &dr); err != nil {
					p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
						Category:  "pipeline_round",
						EventName: "summarize_data_request_invalid",
						LogLevel:  "warn",
						Message:   err.Error(),
						Round:     round,
						DetailData: marshalLogDetail(map[string]any{
							"raw_request_data": rawRequestData,
							"summary_label":    summaryLabel,
							"error":            err.Error(),
						}),
					})
					currentState = append(currentState, "[summarize request_data invalid] "+err.Error())
					continue
				}
				if strings.TrimSpace(dr.Target) == "" {
					dr.Target = "store"
				}
				if dr.Target == "game_client" {
					if err := normalizeDynamicDataRequest(req, &dr); err != nil {
						p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
							Category:  "pipeline_round",
							EventName: "summarize_data_request_blocked",
							LogLevel:  "warn",
							Message:   err.Error(),
							Round:     round,
							DetailData: marshalLogDetail(map[string]any{
								"request":       dr,
								"summary_label": summaryLabel,
								"error":         err.Error(),
							}),
						})
						currentState = append(currentState, "[summarize request_data blocked] "+err.Error())
						continue
					}
					if strings.TrimSpace(dr.Label) == "" {
						dr.Label = firstNonEmpty(summaryLabel, "game_client")
					}
					p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
						Category:   "pipeline_round",
						EventName:  "summarize_data_request_emitted",
						Message:    dr.Label,
						Round:      round,
						DetailData: marshalLogDetail(map[string]any{"request": dr, "summary_label": summaryLabel}),
					})
					return mergedResp, &dr, &label, append([]string(nil), currentState...), nil
				}
				if dr.Target != "store" {
					p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
						Category:  "pipeline_round",
						EventName: "summarize_data_request_blocked",
						LogLevel:  "warn",
						Message:   "only store target is supported during DAG summarization",
						Round:     round,
						DetailData: marshalLogDetail(map[string]any{
							"request":       dr,
							"summary_label": summaryLabel,
						}),
					})
					currentState = append(currentState, "[summarize request_data blocked] only store target is supported during DAG summarization")
					continue
				}
				if len(dr.Queries) == 0 {
					p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
						Category:  "pipeline_round",
						EventName: "summarize_data_request_blocked",
						LogLevel:  "warn",
						Message:   "no queries requested",
						Round:     round,
						DetailData: marshalLogDetail(map[string]any{
							"request":       dr,
							"summary_label": summaryLabel,
						}),
					})
					currentState = append(currentState, "[summarize request_data blocked] no queries requested")
					continue
				}
				p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
					Category:   "pipeline_round",
					EventName:  "summarize_data_request_emitted",
					Message:    firstNonEmpty(dr.Label, "store_query"),
					Round:      round,
					DetailData: marshalLogDetail(map[string]any{"request": dr, "summary_label": summaryLabel}),
				})
				currentState = append(currentState, "[summarize request_data] "+firstNonEmpty(dr.Label, "store_query"))
				result := p.handleDataRequest(nil, &dr)
				if strings.TrimSpace(result) == "" {
					result = "[no data returned]"
				}
				p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
					Category:  "pipeline_round",
					EventName: "summarize_data_request_resolved",
					Message:   firstNonEmpty(dr.Label, "store_query"),
					Round:     round,
					DetailData: marshalLogDetail(map[string]any{
						"request":       dr,
						"summary_label": summaryLabel,
						"result":        result,
						"source":        "store",
					}),
				})
				currentState = append(currentState, result)
				continue
			}
			if reply := strings.TrimSpace(parsed.Reply); reply != "" {
				mergedResp = p.applySubTaskMergeMode(mergedResp, st.Decl, resp, reply)
				normalizedState = nil
				summaryLabel = ""
				goto nextLabel
			}
			if raw := strings.TrimSpace(llmResp.Content); raw != "" {
				mergedResp = p.applySubTaskMergeMode(mergedResp, st.Decl, resp, raw)
				normalizedState = nil
				summaryLabel = ""
				goto nextLabel
			}
		}
		mergedResp = p.applySubTaskMergeMode(mergedResp, st.Decl, resp, "[summarize] LLM 摘要失败")
		normalizedState = nil
		summaryLabel = ""
	nextLabel:
		continue
	}
	if errSummary := buildDAGFailureSummary(dag); errSummary != "" {
		mergedResp.Reply = mergeStrings(mergedResp.Reply, errSummary, "\n")
	}
	return mergedResp, nil, nil, nil, nil
}

func (p *Pipeline) persistPendingSubTaskSummaryPause(executionID string, req *InvokeRequest, ctx *BuiltContext, state *RoundState, runtime *executionConfig, executionMode ExecutionMode, requestID string, pausedRound int, dr *DataRequest, callbackID string, baseResp *InvokeResponse, merged *InvokeResponse, label string, stateLines []string) error {
	if state == nil {
		return fmt.Errorf("round state required for pending summary pause")
	}
	state.PendingResumePhase = "sub_task_summary"
	state.PendingResponse = cloneInvokeResponse(baseResp)
	state.PendingMergedSubTaskResponse = cloneInvokeResponse(merged)
	state.PendingSummaries = []pausedSummaryMerge{{Label: label, StateLines: append([]string(nil), stateLines...)}}
	return p.overwritePausedExecutionSnapshot(executionID, req, ctx, state, runtime, executionMode, requestID, pausedRound, dr, callbackID)
}

func (p *Pipeline) pauseForPendingSubTaskSummary(executionID string, req *InvokeRequest, ctx *BuiltContext, state *RoundState, runtime *executionConfig, executionMode ExecutionMode, requestID string, pausedRound int, dr *DataRequest, baseResp *InvokeResponse, merged *InvokeResponse, subTasks []SubTaskDeclaration, results map[string]*InvokeResponse, failed map[string]string, label string, stateLines []string) (*InvokeResponse, error) {
	callbackID := p.actionReg.CreateCallbackWithMetadata("data_request", map[string]any{"label": dr.Label, "queries": dr.Queries}, action.CallbackMetadata{
		NodeID:    req.NodeID,
		WorldID:   req.WorldID,
		RequestID: requestID,
	})
	if strings.TrimSpace(executionID) == "" {
		var err error
		executionID, err = p.persistPausedExecution(req, ctx, state, runtime, executionMode, requestID, pausedRound, dr, callbackID)
		if err != nil {
			return nil, fmt.Errorf("persist summarize paused execution: %w", err)
		}
	}
	if err := store.UpdateAsyncCallbackRecord(callbackID, map[string]any{"resume_execution_id": executionID}); err != nil {
		return nil, fmt.Errorf("link summarize callback to paused execution: %w", err)
	}
	state.PendingSubTasks = append([]SubTaskDeclaration(nil), subTasks...)
	state.PendingSubTaskResults = cloneSubTaskResultsMap(results)
	state.PendingSubTaskFailed = cloneSubTaskFailedMap(failed)
	if err := p.persistPendingSubTaskSummaryPause(executionID, req, ctx, state, runtime, executionMode, requestID, pausedRound, dr, callbackID, baseResp, merged, label, stateLines); err != nil {
		return nil, fmt.Errorf("persist pending summarize pause: %w", err)
	}
	route := resolveGameClientRoute(dr)
	task, err := enqueueGameClientRuntimeTask(req, dr, callbackID, executionID, requestID, route)
	if err != nil {
		return nil, fmt.Errorf("enqueue summarize runtime task: %w", err)
	}
	if err := p.dispatchGameClientRuntimeTask(task, req, dr, route); err != nil {
		if route.IsStrictPush() {
			_ = store.CompleteAsyncCallbackRecord(callbackID, "failed", "", err.Error())
			_ = store.MarkPausedExecutionFailed(executionID, err.Error())
			return nil, fmt.Errorf("dispatch summarize game client request: %w", err)
		}
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:  "pipeline_round",
			EventName: "summarize_data_request_dispatch_failed",
			LogLevel:  "error",
			Message:   err.Error(),
			Round:     pausedRound,
			DetailData: marshalLogDetail(map[string]any{
				"callback_id":        callbackID,
				"delivery_mode":      route.DeliveryMode,
				"primary_transport":  route.PrimaryTransport,
				"data_request_label": dr.Label,
			}),
		})
	}
	resp := &InvokeResponse{
		RequestID:     requestID,
		ExecutionMode: executionMode,
		TaskType:      req.TaskType,
		Reply:         baseResp.Reply,
		SubTasks:      append([]SubTaskDeclaration(nil), subTasks...),
		DataRequest:   dr,
		ActionCalls: []ActionCall{{
			ActionID:   sdk.ActionIDDataRequest,
			Mode:       sdk.ActionModeAsync,
			CallbackID: callbackID,
			Args:       map[string]any{"data_request": *dr},
		}},
		Metadata: buildResponseMeta(runtime, p.llmProvider.ModelName(), 0, time.Now(), pausedRound),
	}
	p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
		Category:     "pipeline_round",
		EventName:    "summarize_data_request_paused_for_client",
		Message:      dr.Label,
		Round:        pausedRound,
		ResponseData: buildFullResponseLogData(resp),
		DetailData: marshalLogDetail(map[string]any{
			"request":       dr,
			"summary_label": label,
			"callback_id":   callbackID,
			"execution_id":  executionID,
		}),
	})
	return resp, nil
}

func (p *Pipeline) resumePendingSubTaskDAG(parentExecutionID string, parentReq *InvokeRequest, parentCtx *BuiltContext, state *RoundState, runtime *executionConfig, executionMode ExecutionMode, parentRequestID string, parentPausedRound int, callbackResult any) (*InvokeResponse, error) {
	if state == nil || state.PendingResponse == nil {
		return nil, fmt.Errorf("pending sub-task resume state missing")
	}
	baseResp := cloneInvokeResponse(state.PendingResponse)
	mergedResp := cloneInvokeResponse(state.PendingMergedSubTaskResponse)
	dag := reconstructDAGInstance(nil, p.llmProvider, runtime, state.PendingSubTasks, state.PendingSubTaskResults, state.PendingSubTaskFailed)
	if state.PendingResumePhase == "sub_tasks" {
		if state.PendingSubTaskResume == nil {
			return nil, fmt.Errorf("pending sub-task resume state missing")
		}
		pendingSub := state.PendingSubTaskResume
		resumedSubResp, err := p.executeOrResumeSubTask(state.PendingSubTaskResume, callbackResult)
		if err != nil {
			dag.OnTaskFailed(pendingSub.Label, err)
		} else if pausedResponse(resumedSubResp) {
			callbackID := resumedSubResp.ActionCalls[0].CallbackID
			resumeState, loadErr := buildPausedSubTaskResume(callbackID, pendingSub.Label)
			if loadErr != nil {
				return nil, fmt.Errorf("load resumed sub-task paused snapshot: %w", loadErr)
			}
			state.PendingResumePhase = "sub_tasks"
			state.PendingResponse = baseResp
			state.PendingSubTaskResults = cloneSubTaskResultsMap(dag.results)
			state.PendingSubTaskFailed = cloneSubTaskFailedMap(dag.failed)
			state.PendingSubTaskResume = resumeState
			if err := p.overwritePausedExecutionSnapshot(parentExecutionID, parentReq, parentCtx, state, runtime, executionMode, parentRequestID, parentPausedRound, resumedSubResp.DataRequest, callbackID); err != nil {
				return nil, fmt.Errorf("overwrite parent paused execution from resumed sub-task pause: %w", err)
			}
			return buildParentPausedResponse(parentReq, parentRequestID, executionMode, resumedSubResp), nil
		} else {
			dag.OnTaskComplete(pendingSub.Label, resumedSubResp)
		}
		state.PendingSubTaskResume = nil
		for dag.HasReady() {
			for _, st := range dag.ReadyTasks() {
				dag.MarkRunning(st.Label)
				subReq := &InvokeRequest{WorldID: parentReq.WorldID, TaskType: st.TaskType, NodeID: st.NodeID, Context: parentReq.Context}
				subResp, err := p.Execute(subReq)
				if err != nil {
					log.Printf("[dag] sub-task %s failed: %v", st.Label, err)
					dag.OnTaskFailed(st.Label, err)
					continue
				}
				if pausedResponse(subResp) {
					callbackID := subResp.ActionCalls[0].CallbackID
					resumeState, err := buildPausedSubTaskResume(callbackID, st.Label)
					if err != nil {
						return nil, fmt.Errorf("load resumed sub-task paused snapshot: %w", err)
					}
					state.PendingResumePhase = "sub_tasks"
					state.PendingResponse = baseResp
					state.PendingSubTaskResults = cloneSubTaskResultsMap(dag.results)
					state.PendingSubTaskFailed = cloneSubTaskFailedMap(dag.failed)
					state.PendingSubTaskResume = resumeState
					if err := p.overwritePausedExecutionSnapshot(resumeState.ExecutionID, parentReq, parentCtx, state, runtime, executionMode, parentRequestID, parentPausedRound, subResp.DataRequest, callbackID); err != nil {
						return nil, fmt.Errorf("overwrite parent paused execution from resumed DAG pause: %w", err)
					}
					return buildParentPausedResponse(parentReq, parentRequestID, executionMode, subResp), nil
				}
				dag.OnTaskComplete(st.Label, subResp)
			}
		}
	}
	if len(state.PendingSummaries) > 0 {
		var err error
		var pausedDR *DataRequest
		var pausedLabel *string
		var pausedStateLines []string
		mergedResp, pausedDR, pausedLabel, pausedStateLines, err = p.resumePendingSubTaskSummary(parentReq, mergedResp, dag, state.PendingSummaries[0], callbackResult, runtime, executionMode, parentPausedRound)
		if err != nil {
			return nil, err
		}
		if pausedDR != nil && pausedLabel != nil {
			pauseResp, err := p.pauseForPendingSubTaskSummary(parentExecutionID, parentReq, parentCtx, state, runtime, executionMode, parentRequestID, parentPausedRound, pausedDR, baseResp, mergedResp, state.PendingSubTasks, dag.results, dag.failed, *pausedLabel, pausedStateLines)
			if err != nil {
				return nil, err
			}
			return pauseResp, nil
		}
	} else {
		var err error
		var pausedDR *DataRequest
		var pausedLabel *string
		var pausedStateLines []string
		mergedResp, pausedDR, pausedLabel, pausedStateLines, err = p.continueSubTaskSummaryMerge(parentReq, mergedResp, dag, "", nil, runtime, executionMode, parentPausedRound)
		if err != nil {
			return nil, err
		}
		if pausedDR != nil && pausedLabel != nil {
			pauseResp, err := p.pauseForPendingSubTaskSummary(parentExecutionID, parentReq, parentCtx, state, runtime, executionMode, parentRequestID, parentPausedRound, pausedDR, baseResp, mergedResp, state.PendingSubTasks, dag.results, dag.failed, *pausedLabel, pausedStateLines)
			if err != nil {
				return nil, err
			}
			return pauseResp, nil
		}
	}
	mergeSubTaskResultIntoParent(baseResp, mergedResp, state.PendingSubTasks)
	clearPendingSubTaskState(state)
	return baseResp, nil
}

func (p *Pipeline) executeOrResumeSubTask(sub *pausedSubTaskResume, callbackResult any) (*InvokeResponse, error) {
	if sub == nil || sub.Request == nil || sub.BuiltContext == nil || sub.RoundState == nil {
		return nil, fmt.Errorf("paused sub-task resume snapshot incomplete")
	}
	runtime := restoreExecutionRuntime(sub.Runtime)
	runtime.policyEngine = p.loadWorldPolicy(sub.Request.WorldID)
	mode := executionModeFromSnapshot(sub.Runtime)
	if sub.DataRequest != nil {
		label := sub.DataRequest.Label
		if label == "" {
			label = "game_client"
		}
		resolved := marshalLogDetail(map[string]any{
			"callback_id": sub.CallbackID,
			"result":      callbackResult,
		})
		sub.RoundState.SupplementalContext = append(sub.RoundState.SupplementalContext, "[鏁版嵁鏌ヨ鍥炲～] "+label, resolved)
		appendRoundStateTreeEntry(sub.RoundState, sub.PausedRound, nil, resolved)
	}
	return p.executeMultiTurnLoopFromState(sub.Request, sub.BuiltContext, time.Now(), sub.RequestID, runtime, sub.RoundState, sub.PausedRound, mode)
}

func (p *Pipeline) runSubTaskDAG(req *InvokeRequest, resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, state *RoundState, runtime *executionConfig, executionMode ExecutionMode, requestID string, round int) (*InvokeResponse, bool, error) {
	if state == nil || state.Tree == nil || runtime == nil || runtime.pipelineMode != PipelineFull || parsed == nil || strings.TrimSpace(parsed.RawSubTasks) == "" {
		return resp, false, nil
	}
	var subTasks []SubTaskDeclaration
	if err := json.Unmarshal([]byte(parsed.RawSubTasks), &subTasks); err != nil || len(subTasks) == 0 {
		return resp, false, nil
	}
	p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline_round",
		EventName:  "sub_tasks_declared",
		Message:    fmt.Sprintf("round %d declared %d sub tasks", round, len(subTasks)),
		Round:      round,
		DetailData: marshalLogDetail(subTasks),
	})

	dag := reconstructDAGInstance(state.Tree, p.llmProvider, runtime, subTasks, nil, nil)

	for dag.HasReady() {
		for _, st := range dag.ReadyTasks() {
			dag.MarkRunning(st.Label)
			subReq := &InvokeRequest{WorldID: req.WorldID, TaskType: st.TaskType, NodeID: st.NodeID, Context: req.Context}
			subResp, err := p.Execute(subReq)
			if err != nil {
				log.Printf("[dag] sub-task %s failed: %v", st.Label, err)
				dag.OnTaskFailed(st.Label, err)
				continue
			}
			if pausedResponse(subResp) {
				callbackID := subResp.ActionCalls[0].CallbackID
				resumeState, err := buildPausedSubTaskResume(callbackID, st.Label)
				if err != nil {
					return nil, false, fmt.Errorf("load sub-task paused snapshot: %w", err)
				}
				state.PendingResumePhase = "sub_tasks"
				state.PendingResponse = cloneInvokeResponse(resp)
				state.PendingSubTasks = append([]SubTaskDeclaration(nil), subTasks...)
				state.PendingSubTaskResults = cloneSubTaskResultsMap(dag.results)
				state.PendingSubTaskFailed = cloneSubTaskFailedMap(dag.failed)
				state.PendingSubTaskResume = resumeState
				if err := p.overwritePausedExecutionSnapshot(resumeState.ExecutionID, req, ctx, state, runtime, executionMode, requestID, round, subResp.DataRequest, callbackID); err != nil {
					return nil, false, fmt.Errorf("overwrite parent paused execution from sub-task pause: %w", err)
				}
				return buildParentPausedResponse(req, requestID, executionMode, subResp), true, nil
			}
			dag.OnTaskComplete(st.Label, subResp)
		}
	}

	merged, pausedDR, pausedLabel, pausedStateLines, err := p.continueSubTaskSummaryMerge(req, nil, dag, "", nil, runtime, executionMode, round)
	if err != nil {
		return nil, false, err
	}
	if pausedDR != nil && pausedLabel != nil {
		pauseResp, err := p.pauseForPendingSubTaskSummary("", req, ctx, state, runtime, executionMode, requestID, round, pausedDR, resp, merged, subTasks, dag.results, dag.failed, *pausedLabel, pausedStateLines)
		if err != nil {
			return nil, false, err
		}
		return pauseResp, true, nil
	}
	mergeSubTaskResultIntoParent(resp, merged, subTasks)
	clearPendingSubTaskState(state)
	return resp, false, nil
}

func requestDynamicInterfaces(req *InvokeRequest) []DynamicInterface {
	if req == nil || req.Context == nil {
		return nil
	}
	return req.Context.DynamicInterfaces
}

func requestDynamicTools(req *InvokeRequest) []LLMToolDefinition {
	return buildDynamicInterfaceTools(requestDynamicInterfaces(req))
}

func requestLLMTools(req *InvokeRequest, builtinTools []LLMToolDefinition) []LLMToolDefinition {
	return appendUniqueTools(builtinTools, requestDynamicTools(req)...)
}

func llmProviderSupportsStructuredTools(provider LLMProvider) bool {
	if capable, ok := provider.(LLMStructuredToolProvider); ok {
		return capable.SupportsStructuredTools()
	}
	return true
}

func (p *Pipeline) negotiatedLLMTools(tools []LLMToolDefinition) []LLMToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	if !llmProviderSupportsStructuredTools(p.llmProvider) {
		return nil
	}
	return tools
}

func appendRoundStateTreeEntry(state *RoundState, round int, parsed *llmParsedOutput, resolvedData string) {
	if state == nil {
		return
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("[round_%d]", round))
	if parsed != nil && strings.TrimSpace(parsed.Reply) != "" {
		parts = append(parts, "  [分析] "+truncateForContext(parsed.Reply, 500))
	}
	if parsed != nil && strings.TrimSpace(parsed.RawActionCalls) != "" {
		parts = append(parts, "  [动作] "+truncateForContext(parsed.RawActionCalls, 500))
	}
	if strings.TrimSpace(resolvedData) != "" {
		parts = append(parts, "  [数据查询] "+truncateForContext(resolvedData, 1500))
	}
	entry := strings.Join(parts, "\n")
	if strings.TrimSpace(entry) == "" {
		return
	}
	if strings.TrimSpace(state.TreeContext) == "" {
		state.TreeContext = entry
		return
	}
	state.TreeContext = strings.TrimSpace(state.TreeContext + "\n" + entry)
}

func buildPausedExecutionRuntime(runtime *executionConfig, executionMode ExecutionMode) pausedExecutionRuntime {
	if runtime == nil {
		return pausedExecutionRuntime{ExecutionMode: string(executionMode)}
	}
	return pausedExecutionRuntime{
		ConfiguredPipelineMode: string(runtime.configuredPipelineMode),
		EffectivePipelineMode:  string(runtime.pipelineMode),
		MaxRounds:              runtime.maxRounds,
		SubTaskRetries:         runtime.subTaskRetries,
		SubTaskTimeout:         runtime.subTaskTimeout,
		MemoryLimit:            runtime.memoryLimit,
		ExecutionMode:          string(executionMode),
	}
}

func restoreExecutionRuntime(snapshot pausedExecutionRuntime) *executionConfig {
	configured := PipelineMode(snapshot.ConfiguredPipelineMode)
	if configured == "" {
		configured = PipelineFull
	}
	effective := PipelineMode(snapshot.EffectivePipelineMode)
	if effective == "" {
		effective = configured
	}
	return &executionConfig{
		memoryLimit:            snapshot.MemoryLimit,
		maxRounds:              snapshot.MaxRounds,
		subTaskRetries:         snapshot.SubTaskRetries,
		subTaskTimeout:         snapshot.SubTaskTimeout,
		configuredPipelineMode: configured,
		pipelineMode:           effective,
	}
}

func decodePausedExecutionSnapshot(model *store.PausedExecutionModel) (*InvokeRequest, *BuiltContext, *RoundState, *executionConfig, *DataRequest, ExecutionMode, int, error) {
	if model == nil {
		return nil, nil, nil, nil, nil, ModeProduction, 0, fmt.Errorf("paused execution is nil")
	}
	var req InvokeRequest
	if err := json.Unmarshal([]byte(model.OriginalRequestJSON), &req); err != nil {
		return nil, nil, nil, nil, nil, ModeProduction, 0, fmt.Errorf("decode original request: %w", err)
	}
	var ctx BuiltContext
	if err := json.Unmarshal([]byte(model.BuiltContextJSON), &ctx); err != nil {
		return nil, nil, nil, nil, nil, ModeProduction, 0, fmt.Errorf("decode built context: %w", err)
	}
	var state RoundState
	if err := json.Unmarshal([]byte(model.RoundStateJSON), &state); err != nil {
		return nil, nil, nil, nil, nil, ModeProduction, 0, fmt.Errorf("decode round state: %w", err)
	}
	state.Context = &ctx
	state.Tree = nil
	var runtimeSnapshot pausedExecutionRuntime
	if err := json.Unmarshal([]byte(model.RuntimeJSON), &runtimeSnapshot); err != nil {
		return nil, nil, nil, nil, nil, ModeProduction, 0, fmt.Errorf("decode runtime: %w", err)
	}
	runtime := restoreExecutionRuntime(runtimeSnapshot)
	runtime.policyEngine = planner.NewPolicyEngine()
	runtime.policyEngine = (&Pipeline{}).loadWorldPolicy(req.WorldID)
	var dataRequest *DataRequest
	if strings.TrimSpace(model.PendingDataRequestJSON) != "" {
		var dr DataRequest
		if err := json.Unmarshal([]byte(model.PendingDataRequestJSON), &dr); err != nil {
			return nil, nil, nil, nil, nil, ModeProduction, 0, fmt.Errorf("decode pending data request: %w", err)
		}
		dataRequest = &dr
	}
	executionMode := ModeProduction
	switch runtimeSnapshot.ExecutionMode {
	case string(ModeDebug):
		executionMode = ModeDebug
	case string(ModeReview):
		executionMode = ModeReview
	}
	return &req, &ctx, &state, runtime, dataRequest, executionMode, model.PausedRound, nil
}

func (p *Pipeline) persistPausedExecution(req *InvokeRequest, ctx *BuiltContext, state *RoundState, runtime *executionConfig, executionMode ExecutionMode, requestID string, pausedRound int, dr *DataRequest, callbackID string) (string, error) {
	executionID := uuid.NewString()
	if state != nil {
		state.Tree = nil
		state.Context = nil
		state.TreeContext = buildRoundStateTreeContext(state)
	}
	originalReqJSON, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	builtContextJSON, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}
	roundStateJSON, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	runtimeJSON, err := json.Marshal(buildPausedExecutionRuntime(runtime, executionMode))
	if err != nil {
		return "", err
	}
	dataRequestJSON := ""
	if dr != nil {
		if data, err := json.Marshal(dr); err == nil {
			dataRequestJSON = string(data)
		}
	}
	model := &store.PausedExecutionModel{
		ExecutionID:            executionID,
		RequestID:              requestID,
		WorldUUID:              req.WorldID,
		NodeUUID:               req.NodeID,
		TaskType:               string(req.TaskType),
		ExecutionMode:          string(executionMode),
		ConfiguredPipelineMode: configuredPipelineMode(runtime),
		EffectivePipelineMode:  effectivePipelineMode(runtime),
		Status:                 "paused",
		PausedRound:            pausedRound,
		MaxRounds:              runtime.maxRounds,
		TargetNodeID:           state.TargetNodeID,
		PauseReason:            "game_client_request_data",
		CallbackID:             callbackID,
		OriginalRequestJSON:    string(originalReqJSON),
		BuiltContextJSON:       string(builtContextJSON),
		RuntimeJSON:            string(runtimeJSON),
		RoundStateJSON:         string(roundStateJSON),
		PendingDataRequestJSON: dataRequestJSON,
	}
	if err := store.CreatePausedExecution(model); err != nil {
		return "", err
	}
	return executionID, nil
}

func buildRuntimeTaskPayload(req *InvokeRequest, dr *DataRequest, callbackID string, executionID string, requestID string, route external.Route) string {
	interfaceCfg, _ := externalInterfaceConfig(gameClientInterfaceName(dr))
	maxAttempts := 0
	if dr != nil && dr.MaxAttempts > 0 {
		maxAttempts = dr.MaxAttempts
	} else if interfaceCfg.MaxAttempts > 0 {
		maxAttempts = interfaceCfg.MaxAttempts
	}
	payload := map[string]any{
		"task_type":                req.TaskType,
		"world_id":                 req.WorldID,
		"node_id":                  req.NodeID,
		"request_id":               requestID,
		"callback_id":              callbackID,
		"resume_execution_id":      executionID,
		"resume_policy":            firstNonEmpty(route.ResumePolicy, "resume_paused_execution"),
		"external_interface":       gameClientInterfaceName(dr),
		"external_interaction":     "external_query",
		"delivery_mode":            route.DeliveryMode,
		"primary_transport":        route.PrimaryTransport,
		"fallback_transport":       route.FallbackTransport,
		"consumer":                 route.Consumer,
		"max_attempts":             maxAttempts,
		"heartbeat_timeout_policy": heartbeatTimeoutPolicySnapshot(interfaceCfg),
		"request_data":             dr,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func enqueueGameClientRuntimeTask(req *InvokeRequest, dr *DataRequest, callbackID string, executionID string, requestID string, route external.Route) (*store.RuntimeTaskModel, error) {
	interfaceCfg, _ := externalInterfaceConfig(gameClientInterfaceName(dr))
	maxAttempts := 0
	if dr != nil && dr.MaxAttempts > 0 {
		maxAttempts = dr.MaxAttempts
	} else if interfaceCfg.MaxAttempts > 0 {
		maxAttempts = interfaceCfg.MaxAttempts
	}
	item := &store.RuntimeTaskModel{
		Category:          "external_query",
		InterfaceName:     gameClientInterfaceName(dr),
		DeliveryMode:      route.DeliveryMode,
		Consumer:          route.Consumer,
		Transport:         route.PrimaryTransport,
		WorldUUID:         req.WorldID,
		NodeUUID:          req.NodeID,
		RequestID:         requestID,
		CallbackID:        callbackID,
		ResumeExecutionID: executionID,
		MaxAttempts:       maxAttempts,
		Status:            store.RuntimeTaskStatusPending,
		Priority:          100,
		PayloadJSON:       buildRuntimeTaskPayload(req, dr, callbackID, executionID, requestID, route),
	}
	if err := store.CreateRuntimeTask(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (p *Pipeline) dispatchGameClientRuntimeTask(task *store.RuntimeTaskModel, req *InvokeRequest, dr *DataRequest, route external.Route) error {
	if task == nil || !route.ShouldPush() {
		return nil
	}
	idempotencyKey := task.TaskID
	dispatchReq := external.DispatchRequest{
		TaskID:            task.TaskID,
		IdempotencyKey:    idempotencyKey,
		Category:          task.Category,
		InterfaceName:     task.InterfaceName,
		DeliveryMode:      route.DeliveryMode,
		PrimaryTransport:  route.PrimaryTransport,
		Consumer:          route.Consumer,
		WorldID:           req.WorldID,
		NodeID:            req.NodeID,
		RequestID:         task.RequestID,
		CallbackID:        task.CallbackID,
		ResumeExecutionID: task.ResumeExecutionID,
		ResumePolicy:      firstNonEmpty(route.ResumePolicy, "resume_paused_execution"),
		Payload: map[string]any{
			"request_data": dr,
		},
		RawPayloadJSON: task.PayloadJSON,
	}
	result, err := p.externalDispatcher().Dispatch(context.Background(), route, dispatchReq)
	if err != nil {
		attempts := dispatchAttemptsFromResult(result)
		failureClass := external.ClassifyDispatchFailure(result, err)
		meta := store.RuntimeTaskDispatchMetadata{
			Transport:             route.PrimaryTransport,
			FallbackTransport:     route.FallbackTransport,
			FallbackFromTransport: route.PrimaryTransport,
			IdempotencyKey:        idempotencyKey,
			DispatchAttempts:      attempts,
			ErrorMessage:          err.Error(),
			StatusCode:            dispatchStatusCodeFromResult(result),
			FailureClass:          failureClass,
		}
		if route.ShouldQueuePullTask() && route.FallbackTransport != "" {
			meta.Decision = "fallback_to_pull"
			meta.TransitionReason = "push_dispatch_failed_then_fallback"
			_, _ = store.RecordRuntimeTaskDispatchFallback(task.TaskID, meta)
		} else {
			meta.Decision = "failed_terminal"
			meta.TransitionReason = "push_dispatch_failed"
			if route.ShouldQueuePullTask() {
				meta.Decision = "pending_retry"
			}
			_, _ = store.RecordRuntimeTaskDispatchFailure(task.TaskID, route.ShouldQueuePullTask(), meta)
		}
		return err
	}
	_, err = store.MarkRuntimeTaskDispatched(task.TaskID, store.RuntimeTaskDispatchMetadata{
		Transport:        route.PrimaryTransport,
		IdempotencyKey:   idempotencyKey,
		DispatchAttempts: dispatchAttemptsFromResult(result),
		Result:           result,
		StatusCode:       dispatchStatusCodeFromResult(result),
		Decision:         "dispatched",
		TransitionReason: "push_dispatch_succeeded",
	})
	return err
}

func (p *Pipeline) executeMultiTurnLoop(
	req *InvokeRequest,
	ctx *BuiltContext,
	start time.Time,
	requestID string,
	runtime *executionConfig,
	taskTree *TaskTree,
	taskPromptFn func(treeContext string, req *InvokeRequest, nodeID string, round int) string,
	toolFn func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition,
	finalizeFn func(*InvokeResponse, *llmParsedOutput, *BuiltContext, *InvokeRequest) *InvokeResponse,
	executionMode ExecutionMode,
) (*InvokeResponse, error) {
	targetNodeID := req.NodeID
	if ctx != nil && ctx.TargetNode != nil && strings.TrimSpace(ctx.TargetNode.UUID) != "" {
		targetNodeID = ctx.TargetNode.UUID
	}
	state := &RoundState{
		Context:      ctx,
		Tree:         taskTree,
		TreeContext:  "",
		Messages:     sanitizeRoles(req.Messages),
		TargetNodeID: targetNodeID,
		MaxRounds:    runtime.maxRounds,
	}
	if isPlayerInputInterpretRequest(req) {
		state.Messages = sanitizeRoles(normalizedPlayerInputMessages(req.Messages))
	}
	return p.executeMultiTurnLoopInternal(req, ctx, start, requestID, runtime, state, 0, taskPromptFn, toolFn, finalizeFn, executionMode)
}

func (p *Pipeline) executeMultiTurnLoopFromState(
	req *InvokeRequest,
	ctx *BuiltContext,
	start time.Time,
	requestID string,
	runtime *executionConfig,
	state *RoundState,
	startRound int,
	executionMode ExecutionMode,
) (*InvokeResponse, error) {
	var taskPromptFn func(treeContext string, req *InvokeRequest, nodeID string, round int) string
	var toolFn func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition
	var finalizeFn func(*InvokeResponse, *llmParsedOutput, *BuiltContext, *InvokeRequest) *InvokeResponse
	switch req.TaskType {
	case TaskNPCDialogue:
		taskPromptFn = func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
			return buildInteractionDialoguePrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), nodeID, reqInteractionContext(req))
		}
		toolFn = func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
			_ = nodeID
			_ = round
			return requestLLMTools(req, append([]LLMToolDefinition{builtinStoreRequestTool()}, builtinActionTools([]string{"add_memory", "update_mood", "send_dialogue", "adjust_relation", "spawn_item"})...))
		}
	case TaskWorldTick:
		var currentOutline string
		if latest, err := store.GetLatestTick(req.WorldID); err == nil {
			currentOutline = latest.FutureOutline
		}
		worldTimeBlock := buildWorldTickTimeBlock(req.WorldID)
		relationSummary := buildWorldTickRelationSummary(ctx)
		var recentTimeline []string
		if ticks, err := store.GetTimelineTicks(req.WorldID, 3); err == nil {
			for _, tick := range ticks {
				summary := strings.TrimSpace(tick.Summary)
				if summary == "" {
					continue
				}
				recentTimeline = append(recentTimeline, fmt.Sprintf("[tick %d] %s", tick.TickNumber, summary))
			}
		}
		taskPromptFn = func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
			baseContext := mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext)
			return buildWorldTickPrompt(appendDynamicInterfaceContext(baseContext, requestDynamicInterfaces(req)), currentOutline, ctx.StateBlocks, recentTimeline, worldTimeBlock, relationSummary, ctx.BootstrapBlock)
		}
		toolFn = func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
			_ = nodeID
			_ = round
			return requestLLMTools(req, []LLMToolDefinition{builtinStoreRequestTool()})
		}
		finalizeFn = func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
			_ = ctx
			_ = req
			if parsed != nil && parsed.AdvancedTicks > 0 {
				resp.AdvancedTicks = parsed.AdvancedTicks
			}
			return resp
		}
	case TaskWorldEvent:
		eventDesc := ""
		if req.Event != nil {
			eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s", req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
		}
		taskPromptFn = func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
			return buildEventImpactPrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), eventDesc, nodeID)
		}
		toolFn = func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
			_ = nodeID
			_ = round
			return requestLLMTools(req, append([]LLMToolDefinition{builtinStoreRequestTool()}, builtinActionTools([]string{"adjust_relation"})...))
		}
	case TaskAutonomousAct:
		cfg, _, err := LoadAutonomousConfig(req.NodeID)
		if err != nil {
			return nil, err
		}
		taskPromptFn = func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
			return buildAutonomousPrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), nodeID, cfg)
		}
		toolFn = func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
			_ = nodeID
			_ = round
			var ids []string
			for _, cap := range cfg.Capabilities {
				if strings.TrimSpace(cap.ID) != "" {
					ids = append(ids, cap.ID)
				}
			}
			return requestLLMTools(req, append([]LLMToolDefinition{builtinStoreRequestTool()}, builtinActionTools(ids)...))
		}
		finalizeFn = func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
			_ = parsed
			_ = ctx
			_ = req
			allowedCalls, rejected := filterActionCallsByCapabilities(resp.ActionCalls, cfg.Capabilities)
			allowedCalls, schemaRejected := validateActionCallsBySchema(allowedCalls, cfg.Capabilities)
			rejected = append(rejected, schemaRejected...)
			for _, call := range rejected {
				log.Printf("[autonomous:blocked] node=%s action=%s", req.NodeID, call.ActionID)
			}
			resp.ActionCalls = p.executeActions(req, runtime, executionMode, runtime.policyEngine, allowedCalls)
			if len(resp.ActionCalls) == 0 && len(resp.MemoryUpdates) == 0 {
				memUpdate := MemoryUpdate{NodeID: req.NodeID, Content: "自主行为周期未采取行动。", Level: MemShortTerm, Tags: "autonomous,no_action"}
				resp.MemoryUpdates = append(resp.MemoryUpdates, memUpdate)
			}
			return resp
		}
	default:
		if isPlayerInputInterpretRequest(req) {
			taskPromptFn = func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
				return buildPlayerIntentPrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), nodeID, reqInteractionContext(req))
			}
			toolFn = func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
				_ = nodeID
				_ = round
				return requestLLMTools(req, []LLMToolDefinition{builtinStoreRequestTool()})
			}
			finalizeFn = func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
				_ = ctx
				_ = req
				return applyParsedPlayerIntent(resp, parsed)
			}
		} else {
			taskPromptFn = func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
				return treeContext
			}
			toolFn = func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
				_ = nodeID
				_ = round
				return requestLLMTools(req, []LLMToolDefinition{builtinStoreRequestTool()})
			}
		}
	}
	return p.executeMultiTurnLoopInternal(req, ctx, start, requestID, runtime, state, startRound, taskPromptFn, toolFn, finalizeFn, executionMode)
}

func (p *Pipeline) executeMultiTurnLoopInternal(
	req *InvokeRequest,
	ctx *BuiltContext,
	start time.Time,
	requestID string,
	runtime *executionConfig,
	state *RoundState,
	startRound int,
	taskPromptFn func(treeContext string, req *InvokeRequest, nodeID string, round int) string,
	toolFn func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition,
	finalizeFn func(*InvokeResponse, *llmParsedOutput, *BuiltContext, *InvokeRequest) *InvokeResponse,
	executionMode ExecutionMode,
) (*InvokeResponse, error) {
	targetNodeID := req.NodeID
	if ctx != nil && ctx.TargetNode != nil && strings.TrimSpace(ctx.TargetNode.UUID) != "" {
		targetNodeID = ctx.TargetNode.UUID
	}
	if state == nil {
		state = &RoundState{
			Context:      ctx,
			Messages:     sanitizeRoles(req.Messages),
			TargetNodeID: targetNodeID,
			MaxRounds:    runtime.maxRounds,
		}
	}
	state.Context = ctx

	for round := startRound; round < runtime.maxRounds; round++ {
		var roundNode *TaskNode
		var promptSeed string
		if state.Tree != nil {
			roundNode = state.Tree.NewRound(fmt.Sprintf("round_%d", round+1))
			promptSeed = buildRoundStateTreeContext(state)
			state.SystemPrompt = taskPromptFn(promptSeed, req, targetNodeID, round)
			roundNode.Prompt = state.SystemPrompt
		} else {
			promptSeed = ctx.SystemPrompt
			state.SystemPrompt = state.buildPrompt(taskPromptFn(promptSeed, req, targetNodeID, round))
		}
		var tools []LLMToolDefinition
		if toolFn != nil {
			tools = toolFn(req, targetNodeID, round)
		} else {
			tools = requestDynamicTools(req)
		}
		providerSupportsTools := llmProviderSupportsStructuredTools(p.llmProvider)
		exposedTools := p.negotiatedLLMTools(tools)
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline_round",
			EventName:  "prompt_prepared",
			Message:    fmt.Sprintf("round %d prompt prepared", round+1),
			Round:      round + 1,
			DetailData: buildRoundLogDetailWithTools(state.SystemPrompt, state.Messages, round+1, targetNodeID, state.Tree, tools, exposedTools, providerSupportsTools),
		})

		llmStart := time.Now()
		llmResp, err := p.llmProvider.Chat(&LLMChatRequest{SystemPrompt: state.SystemPrompt, Messages: state.Messages, Tools: exposedTools})
		if err != nil {
			p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
				Category:   "pipeline_round",
				EventName:  "llm_call_failed",
				LogLevel:   "error",
				Message:    err.Error(),
				Round:      round + 1,
				DurationMs: time.Since(llmStart).Milliseconds(),
			})
			return nil, fmt.Errorf("llm chat: %w", err)
		}

		parsed := p.parseLLMJSON(llmResp.Content)
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline_round",
			EventName:  "llm_response_received",
			Message:    truncateForLog(parsed.Reply, 180),
			Round:      round + 1,
			TokensUsed: llmResp.Tokens,
			DurationMs: time.Since(llmStart).Milliseconds(),
			DetailData: buildLLMResponseDetailWithMetadata(llmResp.Content, parsed, llmResp.Metadata),
		})
		if roundNode != nil {
			roundNode.LLMResponse = llmResp.Content
		}

		if parsed.RawInterimMemoryUpdates != "" {
			if imus := p.parseMemoryUpdates(parsed.RawInterimMemoryUpdates); len(imus) > 0 {
				p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
					Category:   "pipeline_round",
					EventName:  "interim_memory_updates",
					Message:    fmt.Sprintf("round %d produced %d interim memories", round+1, len(imus)),
					Round:      round + 1,
					DetailData: marshalLogDetail(imus),
				})
				p.writeMemories(req, runtime, executionMode, imus)
				for _, imu := range imus {
					p.PropagateMemoryByRule(req, runtime, executionMode, imu, imu.NodeID)
				}
			}
		}

		if parsed.RawRequestData != "" {
			var dr DataRequest
			if err := json.Unmarshal([]byte(parsed.RawRequestData), &dr); err == nil && len(dr.Queries) > 0 {
				if err := normalizeDynamicDataRequest(req, &dr); err != nil {
					p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
						Category:   "pipeline_round",
						EventName:  "data_request_blocked",
						LogLevel:   "warn",
						Message:    err.Error(),
						Round:      round + 1,
						DetailData: marshalLogDetail(map[string]any{"request": dr, "error": err.Error()}),
					})
					appendRoundStateTreeEntry(state, round+1, parsed, "[dynamic interface blocked] "+err.Error())
					continue
				}
				if cached := ensureResolvedDataRequests(state); cached != nil {
					if signature := dataRequestSignature(&dr); signature != "" {
						if resolved, ok := cached[signature]; ok && strings.TrimSpace(resolved) != "" {
							p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
								Category:   "pipeline_round",
								EventName:  "data_request_reused",
								Message:    dr.Label,
								Round:      round + 1,
								DetailData: marshalLogDetail(map[string]any{"request": dr, "resolved": resolved}),
							})
							if roundNode != nil {
								roundNode.Analysis = resolved
								roundNode.Decision = "[reused data_request] " + dr.Label
							} else {
								state.SupplementalContext = append(state.SupplementalContext, "[reused data_request] "+dr.Label, resolved)
							}
							appendRoundStateTreeEntry(state, round+1, parsed, resolved)
							continue
						}
					}
				}
				p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
					Category:   "pipeline_round",
					EventName:  "data_request_emitted",
					Message:    dr.Label,
					Round:      round + 1,
					DetailData: marshalLogDetail(dr),
				})
				switch dr.Target {
				case "game_client":
					callbackID := p.actionReg.CreateCallbackWithMetadata("data_request", map[string]any{"label": dr.Label, "queries": dr.Queries}, action.CallbackMetadata{
						NodeID:    req.NodeID,
						WorldID:   req.WorldID,
						RequestID: requestID,
					})
					executionID, err := p.persistPausedExecution(req, ctx, state, runtime, executionMode, requestID, round+1, &dr, callbackID)
					if err != nil {
						return nil, fmt.Errorf("persist paused execution: %w", err)
					}
					if err := store.UpdateAsyncCallbackRecord(callbackID, map[string]any{"resume_execution_id": executionID}); err != nil {
						return nil, fmt.Errorf("link callback to paused execution: %w", err)
					}
					route := resolveGameClientRoute(&dr)
					task, err := enqueueGameClientRuntimeTask(req, &dr, callbackID, executionID, requestID, route)
					if err != nil {
						return nil, fmt.Errorf("enqueue runtime task: %w", err)
					}
					if err := p.dispatchGameClientRuntimeTask(task, req, &dr, route); err != nil {
						if route.IsStrictPush() {
							_ = store.CompleteAsyncCallbackRecord(callbackID, "failed", "", err.Error())
							_ = store.MarkPausedExecutionFailed(executionID, err.Error())
							return nil, fmt.Errorf("dispatch game client request: %w", err)
						}
						p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
							Category:   "pipeline_round",
							EventName:  "data_request_dispatch_failed",
							LogLevel:   "error",
							Message:    err.Error(),
							Round:      round + 1,
							DetailData: marshalLogDetail(map[string]any{"callback_id": callbackID, "delivery_mode": route.DeliveryMode, "primary_transport": route.PrimaryTransport}),
						})
					}
					resp := &InvokeResponse{
						RequestID:     requestID,
						ExecutionMode: executionMode,
						TaskType:      req.TaskType,
						DataRequest:   &dr,
						ActionCalls: []ActionCall{{
							ActionID:   sdk.ActionIDDataRequest,
							Mode:       sdk.ActionModeAsync,
							CallbackID: callbackID,
							Args:       map[string]any{"data_request": dr},
						}},
						Metadata: buildResponseMeta(runtime, p.llmProvider.ModelName(), llmResp.Tokens, start, round+1),
					}
					p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
						Category:     "pipeline_round",
						EventName:    "data_request_paused_for_client",
						Message:      dr.Label,
						Round:        round + 1,
						TokensUsed:   llmResp.Tokens,
						DurationMs:   time.Since(start).Milliseconds(),
						ResponseData: buildFullResponseLogData(resp),
						DetailData:   buildResponseOutcomeDetail(resp),
					})
					appendResponseLog(p, resp, req)
					return resp, nil
				default:
					result := p.handleDataRequest(runtime.policyEngine, &dr)
					p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
						Category:   "pipeline_round",
						EventName:  "data_request_resolved",
						Message:    dr.Label,
						Round:      round + 1,
						DetailData: marshalLogDetail(map[string]any{"request": dr, "result": result}),
					})
					if roundNode != nil {
						roundNode.Analysis = result
						roundNode.Decision = "[数据查询] " + dr.Label
					} else {
						state.SupplementalContext = append(state.SupplementalContext, "[数据查询] "+dr.Label, result)
				if len(state.SupplementalContext) > 16 { state.SupplementalContext = state.SupplementalContext[len(state.SupplementalContext)-16:] }
					}
					appendRoundStateTreeEntry(state, round+1, parsed, result)
					// E2: inject convergence instruction before the next round
					if conv := convergenceCheck(runtime, round+1, &dr); conv != "" {
						state.SupplementalContext = append(state.SupplementalContext, "[收敛指令]", conv)
						if len(state.SupplementalContext) > 16 { state.SupplementalContext = state.SupplementalContext[len(state.SupplementalContext)-16:] }
					}
					continue
				}
			}
		}

		if roundNode != nil {
			roundNode.Analysis = parsed.Reply
			if parsed.RawActionCalls != "" {
				roundNode.Decision = fmt.Sprintf("动作: %s", truncateForContext(parsed.RawActionCalls, 100))
			}
		}
		appendRoundStateTreeEntry(state, round+1, parsed, "")

		// Parse action calls and memory updates without executing side effects,
		// so we can check review status before anything is persisted.
		parsedActionCalls := p.parseActionCalls(parsed.RawActionCalls, targetNodeID)
		parsedMemoryUpdates := p.parseMemoryUpdates(parsed.RawMemoryUpdates)

		// Parse world change plan to check if review is needed
		var parsedWCP *WorldChangePlan
		if parsed.RawPlan != "" {
			var wcp WorldChangePlan
			if err := json.Unmarshal([]byte(parsed.RawPlan), &wcp); err == nil {
				parsedWCP = &wcp
			}
		}

		// When in review mode and the plan has high impact, defer execution
		// to after human approval instead of applying side effects immediately.
		if executionMode == ModeReview && parsedWCP != nil && IsHighImpact(parsedWCP.ImpactLevel) {
			pendingPlan := &PendingPlan{
				PlanID:          NewPendingPlanID(),
				WorldID:         req.WorldID,
				TickNumber:      0,
				TaskType:        req.TaskType,
				WorldChangePlan: parsedWCP,
				ActionCalls:     parsedActionCalls,
				MemoryUpdates:   parsedMemoryUpdates,
				CreatedAt:       time.Now(),
				Status:          "pending",
			}
			GlobalPlanReview.Add(pendingPlan)

			// Persist to database for crash recovery
			planData, _ := store.SerializePendingPlan(pendingPlan)
			if err := store.CreatePendingPlan(pendingPlan.PlanID, pendingPlan.WorldID, string(pendingPlan.TaskType), pendingPlan.Status, planData, pendingPlan.TickNumber); err != nil {
				log.Printf("[warn] persist pending plan to DB: %v", err)
			}

			resp := &InvokeResponse{
				RequestID:       requestID,
				TaskType:        req.TaskType,
				ExecutionMode:   ModeReview,
				Reply:           "[待审批] 变更计划已挂起，请调用审批 API 确认执行",
				WorldChangePlan: parsedWCP,
				ActionCalls:     nil,
				MemoryUpdates:   nil,
				Metadata:        buildResponseMeta(runtime, p.llmProvider.ModelName(), llmResp.Tokens, start, round+1),
			}

			p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
				Category:   "pipeline_review",
				EventName:  "plan_pending_review",
				Message:    parsedWCP.Summary,
				Round:      round + 1,
				DetailData: marshalLogDetail(pendingPlan),
			})

			appendResponseLog(p, resp, req)
			return resp, nil
		}

		// Normal execution path: parse and execute side effects
		resp := &InvokeResponse{
			RequestID:     requestID,
			TaskType:      req.TaskType,
			ExecutionMode: executionMode,
			Reply:         parsed.Reply,
			ActionCalls:   p.executeActions(req, runtime, executionMode, runtime.policyEngine, parsedActionCalls),
			MemoryUpdates: parsedMemoryUpdates,
		}
		if parsedWCP != nil {
			resp.WorldChangePlan = parsedWCP
		}
		if parsed.RawFutureOutline != "" {
			resp.FutureOutline = parsed.RawFutureOutline
		}

		if finalizeFn != nil {
			resp = finalizeFn(resp, parsed, state.Context, req)
		}
		resp = applyParsedPlayerIntent(resp, parsed)

		p.writeMemories(req, runtime, executionMode, resp.MemoryUpdates)
		for _, mem := range resp.MemoryUpdates {
			p.PropagateMemoryByRule(req, runtime, executionMode, mem, mem.NodeID)
		}

		resp.Metadata = buildResponseMeta(runtime, p.llmProvider.ModelName(), llmResp.Tokens, start, round+1)
		p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
			Category:     "pipeline_round",
			EventName:    "round_completed",
			Message:      truncateForLog(resp.Reply, 180),
			Round:        round + 1,
			TokensUsed:   llmResp.Tokens,
			DurationMs:   time.Since(start).Milliseconds(),
			ResponseData: buildFullResponseLogData(resp),
			DetailData:   buildResponseOutcomeDetail(resp),
		})

		if nextResp, paused, err := p.runSubTaskDAG(req, resp, parsed, ctx, state, runtime, executionMode, requestID, round+1); err != nil {
			return nil, err
		} else if paused {
			appendResponseLog(p, nextResp, req)
			return nextResp, nil
		} else if nextResp != nil {
			resp = nextResp
		}

		switch executionMode {
		case ModeDebug:
			trace := buildDebugTrace(
				req.WorldID, requestID,
				req.TaskType, targetNodeID,
				start, llmStart,
				state.SystemPrompt, state.Messages,
				parsed.Reply,
				resp, runtime, round, "",
			)
			GlobalTraceRing.Push(trace)
		}

		appendResponseLog(p, resp, req)
		return resp, nil
	}

	return nil, fmt.Errorf("%s exceeded max rounds (%d)", req.TaskType, runtime.maxRounds)
}

func appendResponseLog(p *Pipeline, resp *InvokeResponse, req *InvokeRequest) {
	if resp == nil || resp.Metadata == nil {
		return
	}
	runtime := &executionConfig{
		configuredPipelineMode: PipelineMode(resp.Metadata.ConfiguredPipelineMode),
		pipelineMode:           PipelineMode(resp.Metadata.EffectivePipelineMode),
		maxRounds:              resp.Metadata.MaxAnalysisRounds,
	}
	mode := resp.ExecutionMode
	if mode == "" {
		mode = ModeProduction
	}
	p.emitLog(req, resp, runtime, mode, pipelineLogEvent{
		Category:     "pipeline",
		EventName:    "response_completed",
		Message:      resp.Reply,
		Round:        resp.Metadata.RoundsUsed,
		RequestData:  buildInferenceLogRequestData(req),
		ResponseData: buildInferenceLogResponseData(resp),
		DetailData: marshalLogDetail(map[string]any{
			"request":  req,
			"response": resp,
		}),
		DurationMs: resp.Metadata.ProcessingTimeMs,
		TokensUsed: resp.Metadata.TokensUsed,
	})
}

func (p *Pipeline) executeVertical(req *InvokeRequest, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode) (*InvokeResponse, error) {
	var systemPrompt string
	ctxDesc := fmt.Sprintf("世界: %s, 节点: %s, 任务类型: %s", req.WorldID, req.NodeID, req.TaskType)
	switch req.TaskType {
	case TaskNPCDialogue:
		systemPrompt = buildInteractionDialoguePrompt(appendDynamicInterfaceContext(ctxDesc, requestDynamicInterfaces(req)), targetDialogueNodeID(req, nil), reqInteractionContext(req))
	case TaskWorldTick:
		systemPrompt = buildWorldTickPrompt(appendDynamicInterfaceContext(ctxDesc, requestDynamicInterfaces(req)), "", nil, nil, buildWorldTickTimeBlock(req.WorldID), "", "")
	case TaskWorldEvent:
		eventDesc := ""
		if req.Event != nil {
			eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s", req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
		}
		systemPrompt = buildEventImpactPrompt(appendDynamicInterfaceContext(ctxDesc, requestDynamicInterfaces(req)), eventDesc, req.NodeID)
	case TaskAutonomousAct:
		if cfg, _, err := LoadAutonomousConfig(req.NodeID); err == nil && cfg != nil && cfg.Enabled {
			systemPrompt = buildAutonomousPrompt(appendDynamicInterfaceContext(ctxDesc, requestDynamicInterfaces(req)), req.NodeID, cfg)
		} else {
			resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous component not found or disabled", ExecutionMode: executionMode}
			resp.Metadata = buildResponseMeta(runtime, p.llmProvider.ModelName(), 0, start, 0)
			appendResponseLog(p, resp, req)
			return resp, nil
		}
	default:
		if isPlayerInputInterpretRequest(req) {
			systemPrompt = buildPlayerIntentPrompt(appendDynamicInterfaceContext(ctxDesc, requestDynamicInterfaces(req)), req.NodeID, reqInteractionContext(req))
		} else {
			systemPrompt = ctxDesc
		}
	}

	messages := sanitizeRoles(req.Messages)
	if isPlayerInputInterpretRequest(req) {
		messages = sanitizeRoles(normalizedPlayerInputMessages(req.Messages))
	}

	llmResp, err := p.llmProvider.Chat(&LLMChatRequest{SystemPrompt: systemPrompt, Messages: messages, Tools: p.negotiatedLLMTools(requestDynamicTools(req))})
	if err != nil {
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline_round",
			EventName:  "llm_call_failed",
			LogLevel:   "error",
			Message:    err.Error(),
			Round:      1,
			DurationMs: time.Since(start).Milliseconds(),
		})
		return nil, fmt.Errorf("vertical llm: %w", err)
	}

	parsed := p.parseLLMJSON(llmResp.Content)
	providerSupportsTools := llmProviderSupportsStructuredTools(p.llmProvider)
	plannedTools := requestDynamicTools(req)
	exposedTools := p.negotiatedLLMTools(plannedTools)
	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline_round",
		EventName:  "prompt_prepared",
		Message:    "vertical prompt prepared",
		Round:      1,
		DetailData: buildRoundLogDetailWithTools(systemPrompt, sanitizeRoles(req.Messages), 1, req.NodeID, nil, plannedTools, exposedTools, providerSupportsTools),
	})
	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline_round",
		EventName:  "llm_response_received",
		Message:    truncateForLog(parsed.Reply, 180),
		Round:      1,
		TokensUsed: llmResp.Tokens,
		DurationMs: time.Since(start).Milliseconds(),
		DetailData: buildLLMResponseDetailWithMetadata(llmResp.Content, parsed, llmResp.Metadata),
	})
	resp := &InvokeResponse{
		RequestID:     requestID,
		TaskType:      req.TaskType,
		ExecutionMode: executionMode,
		Reply:         parsed.Reply,
		ActionCalls:   p.executeActions(req, runtime, executionMode, runtime.policyEngine, p.parseActionCalls(parsed.RawActionCalls, req.NodeID)),
		MemoryUpdates: p.parseMemoryUpdates(parsed.RawMemoryUpdates),
	}
	resp = applyParsedPlayerIntent(resp, parsed)
	if parsed.RawPlan != "" {
		var wcp WorldChangePlan
		if err := json.Unmarshal([]byte(parsed.RawPlan), &wcp); err == nil {
			resp.WorldChangePlan = &wcp
		}
	}
	if parsed.RawFutureOutline != "" {
		resp.FutureOutline = parsed.RawFutureOutline
	}
	p.writeMemories(req, runtime, executionMode, resp.MemoryUpdates)
	for _, mem := range resp.MemoryUpdates {
		p.PropagateMemoryByRule(req, runtime, executionMode, mem, mem.NodeID)
	}
	resp.Metadata = buildResponseMeta(runtime, p.llmProvider.ModelName(), llmResp.Tokens, start, 1)
	appendResponseLog(p, resp, req)
	return resp, nil
}

func (p *Pipeline) executePolling(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode) (*InvokeResponse, error) {
	switch req.TaskType {
	case TaskNPCDialogue:
		return p.executeDialogue(req, ctx, start, requestID, runtime, executionMode, false)
	case TaskWorldTick:
		return p.executeWorldTick(req, ctx, start, requestID, runtime, executionMode, false)
	case TaskWorldEvent:
		return p.executeWorldEvent(req, ctx, start, requestID, runtime, executionMode, false)
	case TaskAutonomousAct:
		return p.executeAutonomousAct(req, ctx, start, requestID, runtime, executionMode, false)
	default:
		return p.executeCustom(req, ctx, start, requestID, runtime, executionMode, false)
	}
}

func (p *Pipeline) executeFull(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode) (*InvokeResponse, error) {
	switch req.TaskType {
	case TaskNPCDialogue:
		return p.executeDialogue(req, ctx, start, requestID, runtime, executionMode, true)
	case TaskWorldTick:
		return p.executeWorldTick(req, ctx, start, requestID, runtime, executionMode, true)
	case TaskWorldEvent:
		return p.executeWorldEvent(req, ctx, start, requestID, runtime, executionMode, true)
	case TaskAutonomousAct:
		return p.executeAutonomousAct(req, ctx, start, requestID, runtime, executionMode, true)
	default:
		return p.executeCustom(req, ctx, start, requestID, runtime, executionMode, true)
	}
}

func (p *Pipeline) executeDialogue(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode, withTaskTree bool) (*InvokeResponse, error) {
	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
		analysisNode := tree.NewRound("analysis")
		analysisPrompt := fmt.Sprintf("请分析以下局势数据，并将其中的精确数值转化为模糊量词，整理成后续对话可用的局势摘要。\n\n%s", ctx.SystemPrompt)
		analysisNode.Prompt = analysisPrompt
		if analysisResp, err := p.llmProvider.Chat(&LLMChatRequest{SystemPrompt: analysisPrompt}); err == nil {
			analysisNode.LLMResponse = analysisResp.Content
			analysisNode.Analysis = analysisResp.Content
			analysisNode.Decision = "局势分析完成"
		}
	}

	dialogueFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildInteractionDialoguePrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), nodeID, reqInteractionContext(req))
	}

	loopRuntime := *runtime
	if loopRuntime.maxRounds > 1 && withTaskTree {
		loopRuntime.maxRounds--
	}
	dialogueToolFn := func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
		_ = nodeID
		_ = round
		return requestLLMTools(req, append([]LLMToolDefinition{builtinStoreRequestTool()}, builtinActionTools([]string{"add_memory", "update_mood", "send_dialogue", "adjust_relation", "spawn_item"})...))
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, &loopRuntime, tree, dialogueFn, dialogueToolFn, nil, executionMode)
}

func (p *Pipeline) executeWorldTick(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode, withTaskTree bool) (*InvokeResponse, error) {
	var currentOutline string
	if latest, err := store.GetLatestTick(req.WorldID); err == nil {
		currentOutline = latest.FutureOutline
	}
	worldTimeBlock := buildWorldTickTimeBlock(req.WorldID)
	relationSummary := buildWorldTickRelationSummary(ctx)
	var recentTimeline []string
	if ticks, err := store.GetTimelineTicks(req.WorldID, 3); err == nil {
		for _, tick := range ticks {
			summary := strings.TrimSpace(tick.Summary)
			if summary == "" {
				continue
			}
			recentTimeline = append(recentTimeline, fmt.Sprintf("[tick %d] %s", tick.TickNumber, summary))
		}
	}

	// E1: Bootstrap phase -- pre-fetch key authority data before the first LLM round.
	// Reuses existing store query semantics through a lightweight bootstrap helper.
	// The result is injected into ctx.BootstrapBlock as an optional prompt supplement.
	ctx.BootstrapBlock = p.runWorldTickBootstrap(req.WorldID, ctx)

	tickFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		baseContext := mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext)
		return buildWorldTickPrompt(appendDynamicInterfaceContext(baseContext, requestDynamicInterfaces(req)), currentOutline, ctx.StateBlocks, recentTimeline, worldTimeBlock, relationSummary, ctx.BootstrapBlock)
	}

	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}
	tickToolFn := func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
		_ = nodeID
		_ = round
		return requestLLMTools(req, []LLMToolDefinition{builtinStoreRequestTool()})
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, tickFn, tickToolFn, func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
		_ = ctx
		_ = req
		if parsed != nil && parsed.AdvancedTicks > 0 {
			resp.AdvancedTicks = parsed.AdvancedTicks
		}
		return resp
	}, executionMode)
}

func buildWorldTickRelationSummary(ctx *BuiltContext) string {
	if ctx == nil || ctx.Node == nil {
		return ""
	}
	scopeNodes := buildWorldTickScopeNodes(ctx)
	if len(scopeNodes) == 0 {
		return ""
	}
	childrenByParent := map[string][]store.NodeModel{}
	for _, node := range scopeNodes {
		if children, err := store.GetChildNodes(node.UUID); err == nil {
			childrenByParent[node.UUID] = children
		}
	}

	var parts []string
	for _, scope := range scopeNodes {
		parts = append(parts, fmt.Sprintf("- scope: %s(%s)", scope.Name, scope.NodeType))
		rels, err := store.GetNodeRelations(scope.UUID)
		if err == nil {
			for _, line := range summarizeWorldTickRelations(scope, rels) {
				parts = append(parts, "  "+line)
			}
		}
		for _, line := range summarizeWorldTickChildren(scope, childrenByParent[scope.UUID]) {
			parts = append(parts, "  "+line)
		}
		for _, line := range summarizeWorldTickChildRelations(scope, childrenByParent[scope.UUID]) {
			parts = append(parts, "  "+line)
		}
	}
	return strings.TrimSpace(strings.Join(dedupeStrings(parts), "\n"))
}

func buildWorldTickScopeNodes(ctx *BuiltContext) []store.NodeModel {
	var nodes []store.NodeModel
	if ctx.Node != nil {
		nodes = append(nodes, *ctx.Node)
	}
	if ctx.EnvironmentNode != nil {
		nodes = append(nodes, *ctx.EnvironmentNode)
	}
	nodes = append(nodes, ctx.EnvironmentAncestors...)
	nodes = append(nodes, ctx.IdentityAncestors...)
	return dedupeNodes(nodes)
}

func summarizeWorldTickRelations(scope store.NodeModel, rels []store.RelationModel) []string {
	var lines []string
	seen := map[string]bool{}
	for _, rel := range rels {
		if rel.SourceUUID != scope.UUID {
			continue
		}
		summary := summarizeWorldTickRelation(rel)
		if summary == "" || seen[summary] {
			continue
		}
		seen[summary] = true
		lines = append(lines, summary)
	}
	return lines
}

func summarizeWorldTickRelation(rel store.RelationModel) string {
	switch rel.RelationType {
	case string(RelLocatedAt):
		return fmt.Sprintf("位置锚点: %s 位于 %s", relationEndpointName(rel.SourceUUID, rel.SourceUUID), relationEndpointName(rel.TargetUUID, rel.TargetUUID))
	case string(RelBelongsTo):
		return fmt.Sprintf("归属结构: %s 属于 %s", relationEndpointName(rel.SourceUUID, rel.SourceUUID), relationEndpointName(rel.TargetUUID, rel.TargetUUID))
	case string(RelSubordinate):
		return fmt.Sprintf("控制结构: %s 受 %s 指挥", relationEndpointName(rel.SourceUUID, rel.SourceUUID), relationEndpointName(rel.TargetUUID, rel.TargetUUID))
	default:
		return ""
	}
}

func summarizeWorldTickChildren(scope store.NodeModel, children []store.NodeModel) []string {
	if len(children) == 0 {
		return nil
	}
	counts := map[string]int{}
	samples := map[string][]string{}
	for _, child := range children {
		counts[child.NodeType]++
		if len(samples[child.NodeType]) < 3 {
			samples[child.NodeType] = append(samples[child.NodeType], child.Name)
		}
	}
	orderedTypes := []string{string(NodeTypeLocation), string(NodeTypeFaction), string(NodeTypeNPC), string(NodeTypeEvent), string(NodeTypeItem), string(NodeTypeQuestLine)}
	var lines []string
	for _, nodeType := range orderedTypes {
		count := counts[nodeType]
		if count == 0 {
			continue
		}
		line := fmt.Sprintf("子节点分布: %s 下有 %d 个 %s", scope.Name, count, nodeType)
		if names := samples[nodeType]; len(names) > 0 {
			line += fmt.Sprintf("（样本: %s）", strings.Join(names, ", "))
		}
		lines = append(lines, line)
	}
	return lines
}

func summarizeWorldTickChildRelations(scope store.NodeModel, children []store.NodeModel) []string {
	if len(children) == 0 {
		return nil
	}
	allowedChildren := map[string]bool{}
	for _, child := range children {
		allowedChildren[child.UUID] = true
	}
	var lines []string
	seen := map[string]bool{}
	for _, child := range children {
		rels, err := store.GetNodeRelations(child.UUID)
		if err != nil {
			continue
		}
		for _, rel := range rels {
			if rel.SourceUUID != child.UUID {
				continue
			}
			if rel.RelationType != string(RelLocatedAt) && rel.RelationType != string(RelBelongsTo) && rel.RelationType != string(RelSubordinate) {
				continue
			}
			line := summarizeWorldTickRelation(rel)
			if line == "" || seen[line] {
				continue
			}
			if rel.TargetUUID != scope.UUID && !allowedChildren[rel.TargetUUID] {
				targetNode, err := store.GetNode(rel.TargetUUID)
				if err != nil || targetNode == nil || targetNode.ParentUUID == nil || *targetNode.ParentUUID != scope.UUID {
					continue
				}
			}
			seen[line] = true
			lines = append(lines, line)
			if len(lines) >= 6 {
				return lines
			}
		}
	}
	return lines
}

func relationEndpointName(nodeUUID string, fallback string) string {
	node, err := store.GetNode(nodeUUID)
	if err != nil || node == nil || strings.TrimSpace(node.Name) == "" {
		return fallback
	}
	return fmt.Sprintf("%s(%s)", node.Name, node.NodeType)
}

func dedupeStrings(input []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, item := range input {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	return out
}

func buildWorldTickTimeBlock(worldID string) string {
	settingsModel, err := store.GetWorldSettings(worldID)
	if err != nil || settingsModel == nil || strings.TrimSpace(settingsModel.WorldTimeSettingsJSON) == "" {
		return ""
	}
	settings, err := DecodeWorldTimeSettings(settingsModel.WorldTimeSettingsJSON)
	if err != nil || settings == nil {
		return ""
	}
	parts := []string{
		fmt.Sprintf("- tick_scale_mode: %s", settings.TickScaleMode),
		fmt.Sprintf("- tick_min_unit: %s", settings.TickMinUnit),
		fmt.Sprintf("- tick_step: %d", settings.TickStep),
		fmt.Sprintf("- tick_units(big_to_small): %s", strings.Join(settings.TickUnits, " > ")),
	}
	if len(settings.TimeScaleCarry) > 0 {
		carryRules := make([]string, 0, len(settings.TimeScaleCarry))
		for _, rule := range settings.TimeScaleCarry {
			carryRules = append(carryRules, fmt.Sprintf("%s->%s(base=%d)", rule.From, rule.To, rule.Base))
		}
		parts = append(parts, fmt.Sprintf("- time_scale_carry(small_to_big): %s", strings.Join(carryRules, ", ")))
	}
	if settings.TimeCalendar != nil && settings.TimeCalendar.Enabled {
		parts = append(parts, fmt.Sprintf("- calendar_name: %s历", settings.TimeCalendar.CalendarName))
	}
	if components, err := store.GetComponentsByType(worldID, string(CompWorldTimeState)); err == nil && len(components) > 0 {
		state := WorldTimeStateComponent{}
		if json.Unmarshal([]byte(components[0].Data), &state) == nil {
			if strings.TrimSpace(state.CurrentTimeLabel) != "" {
				parts = append(parts, fmt.Sprintf("- current_time_label: %s", state.CurrentTimeLabel))
			}
			if len(state.CurrentUnits) > 0 {
				unitParts := make([]string, 0, len(state.CurrentUnits))
				for _, unit := range state.CurrentUnits {
					unitParts = append(unitParts, fmt.Sprintf("%s=%s", unit.Unit, unit.Value))
				}
				parts = append(parts, fmt.Sprintf("- current_units: %s", strings.Join(unitParts, ", ")))
			}
		}
	}
	if settings.TickScaleMode == TickScaleModeFixed {
		parts = append(parts, "- constraint: each world tick must advance exactly 1 configured standard tick")
	} else {
		parts = append(parts, "- constraint: each world tick may advance multiple standard ticks, but you must return advanced_ticks")
	}
	return strings.Join(parts, "\n")
}

// runWorldTickBootstrap pre-fetches a compact set of authority-relevant store data
// before the world tick LLM loop begins. It reuses existing store query semantics
// and injects results into a request-scoped BootstrapBlock string.
// This reduces the number of low-value request_data rounds in the first LLM pass.
func (p *Pipeline) runWorldTickBootstrap(worldID string, ctx *BuiltContext) string {
	if authPath := authorityDemoStatePath(worldID); authPath != "" {
		if auth, err := LoadDemoAuthorityFile(authPath); err == nil && auth != nil {
			if block := BuildDemoAuthorityBlock(auth); block != "" {
				return block
			}
		}
	}
	if ctx == nil || ctx.Node == nil {
		return ""
	}

	var parts []string
	parts = append(parts, "========== Bootstrap: Pre-fetched Authority Context ==========")
	parts = append(parts, "The following data was pre-loaded before this tick began. Use it to")
	parts = append(parts, "reduce redundant queries. If critical facts are still missing, use request_data.")

	// 1. World-level state components (world_state, story_state, tick_policy, world_time_state)
	stateTypes := []string{"world_state", "story_state", "tick_policy", "world_time_state"}
	for _, ct := range stateTypes {
		comps, err := store.GetComponentsByType(worldID, ct)
		if err != nil || len(comps) == 0 {
			continue
		}
		for _, comp := range comps {
			parts = append(parts, fmt.Sprintf("[bootstrap:component] %s: %s", comp.ComponentType, comp.Data))
		}
	}

	// 2. Recent timeline (up to 5 entries for context)
	if ticks, err := store.GetTimelineTicks(worldID, 5); err == nil && len(ticks) > 0 {
		for _, tick := range ticks {
			summary := strings.TrimSpace(tick.Summary)
			if summary == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("[bootstrap:timeline] tick %d: %s", tick.TickNumber, summary))
		}
	}

	// 3. Scope node's high-level children (for situation awareness, up to 20)
	if children, err := store.GetChildNodes(ctx.Node.UUID); err == nil && len(children) > 0 {
		var childLines []string
		counts := map[string]int{}
		for _, child := range children {
			counts[child.NodeType]++
		}
		for _, nt := range []string{"location", "faction", "npc", "event", "item", "quest_line"} {
			if c := counts[nt]; c > 0 {
				childLines = append(childLines, fmt.Sprintf("%s:%d", nt, c))
			}
		}
		if len(childLines) > 0 {
			parts = append(parts, "[bootstrap:children] scope children: "+strings.Join(childLines, ", "))
		}
	}

	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts, "\n")
}

// convergenceCheck determines whether the current round should force-converge.
// Returns a convergence prompt instruction string, or empty string if normal flow continues.
func convergenceCheck(runtime *executionConfig, round int, dr *DataRequest) string {
	if runtime == nil || runtime.maxRounds <= 0 {
		return ""
	}
	// E2.3: Hard converge at 80% of max rounds
	if round >= runtime.maxRounds {
		return "[收敛指令] 已达到最大分析轮次。你必须立即基于已有信息完成当前 tick，不能再发起新的 request_data。如果关键事实确实缺失，请在 reply 中注明缺口。"
	}
	if round >= int(float64(runtime.maxRounds)*0.8) {
		r := ""
		if dr != nil {
			r = "你的上一轮查询已返回结果。"
		}
		return "[收敛指令] 已接近最大分析轮次上限。" + r + "请优先基于已有事实完成当前 tick 的闭环输出。只有在缺少对当前 tick 输出有决定性影响的关键事实时才继续 request_data。禁止为锦上添花的细节发起新查询。"
	}
	// E2.2: consecutive query round limit
	if dr != nil {
		runtime.queryBudget--
	}
	if runtime.queryBudget <= runtime.queryRoundLimit {
		return "[收敛预算] 剩余查询预算已偏低。请优先完成当前 tick 的核心输出（reply、world_change_plan、future_outline），仅在核心闭环所需时才继续查询。"
	}
	return ""
}

// ColdStartResult describes the outcome of a world cold-start operation.
type ColdStartResult struct {
	WorldID     string   `json:"world_id"`
	Components  []string `json:"components"`
	Warnings    []string `json:"warnings,omitempty"`
	Initialized bool     `json:"initialized"`
}

// ColdStartWorld initializes the runtime baseline for a world after import.
// It does not trigger LLM inference and is independent of world tick.
// The operation generates default runtime components (world_state, story_state,
// tick_policy) if they do not already exist.
// It supports two modes:
//   - initial: generate only missing runtime components
//   - rebuild: regenerate all runtime components (overwrite existing)
func (p *Pipeline) ColdStartWorld(worldID string, mode string) (*ColdStartResult, error) {
	if strings.TrimSpace(worldID) == "" {
		return nil, fmt.Errorf("world_id required")
	}
	rebuild := strings.EqualFold(mode, "rebuild")

	stateTypes := []ComponentType{CompWorldState, CompStoryState, CompTickPolicy, CompWorldTimeState}
	var created []string
	var warnings []string

	for _, ct := range stateTypes {
		existing, err := store.GetComponentsByType(worldID, string(ct))
		if err != nil || len(existing) == 0 || rebuild {
			if !rebuild && err == nil && len(existing) > 0 {
				continue
			}
			var payload string
			switch ct {
			case CompWorldState:
				payload = "{\"summary\":\"\",\"key_facts\":[],\"canonical_facts\":[],\"open_questions\":[],\"active_arcs\":[],\"metadata\":{}}"
			case CompStoryState:
				payload = "{\"current_situation\":\"\",\"recent_changes\":[],\"pending_threads\":[],\"tone\":\"\",\"metadata\":{}}"
			case CompTickPolicy:
				payload = "{\"continuity_rules\":[],\"focus_scopes\":[],\"banned_resets\":[],\"metadata\":{}}"
			case CompWorldTimeState:
				payload = "{\"tick_scale_mode\":\"fixed\",\"tick_min_unit\":\"\",\"tick_step\":1,\"tick_units\":[],\"current_units\":[],\"current_time_label\":\"\",\"total_ticks\":0,\"last_tick_number\":0,\"last_tick_type\":\"\",\"last_advanced_ticks\":0,\"metadata\":{}}"
			default:
				payload = "{}"
			}
			if err := store.AddComponent(&store.ComponentModel{
				UUID:          store.NewUUID(),
				NodeUUID:      worldID,
				ComponentType: string(ct),
				Data:          payload,
			}); err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to create %s: %v", ct, err))
				continue
			}
			created = append(created, string(ct))
		}
	}

	if len(created) == 0 {
		warnings = append(warnings, "all runtime components already exist (use rebuild to regenerate)")
	}

	return &ColdStartResult{
		WorldID:     worldID,
		Components:  created,
		Warnings:    warnings,
		Initialized: len(created) > 0 || rebuild,
	}, nil
}

func (p *Pipeline) executeWorldEvent(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode, withTaskTree bool) (*InvokeResponse, error) {
	eventDesc := ""
	if req.Event != nil {
		eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s", req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
	}

	eventFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildEventImpactPrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), eventDesc, nodeID)
	}

	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}
	eventToolFn := func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
		_ = nodeID
		_ = round
		return requestLLMTools(req, append([]LLMToolDefinition{builtinStoreRequestTool()}, builtinActionTools([]string{"adjust_relation"})...))
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, eventFn, eventToolFn, nil, executionMode)
}

func (p *Pipeline) executeAutonomousAct(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode, withTaskTree bool) (*InvokeResponse, error) {
	targetNodeID := req.NodeID
	cfg, comp, err := LoadAutonomousConfig(targetNodeID)
	if err != nil {
		return nil, err
	}
	if cfg == nil || comp == nil {
		resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous component not found"}
		resp.Metadata = buildResponseMeta(runtime, p.llmProvider.ModelName(), 0, start, 0)
		return resp, nil
	}
	if !cfg.Enabled {
		resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous behavior disabled"}
		resp.Metadata = buildResponseMeta(runtime, p.llmProvider.ModelName(), 0, start, 0)
		return resp, nil
	}
	if len(cfg.Capabilities) == 0 {
		resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous behavior has no capabilities"}
		resp.Metadata = buildResponseMeta(runtime, p.llmProvider.ModelName(), 0, start, 0)
		return resp, nil
	}

	autonomousFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildAutonomousPrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), nodeID, cfg)
	}

	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}

	autonomousToolFn := func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
		_ = nodeID
		_ = round
		var ids []string
		for _, cap := range cfg.Capabilities {
			if strings.TrimSpace(cap.ID) != "" {
				ids = append(ids, cap.ID)
			}
		}
		return requestLLMTools(req, append([]LLMToolDefinition{builtinStoreRequestTool()}, builtinActionTools(ids)...))
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, autonomousFn, autonomousToolFn, func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
		_ = parsed
		_ = ctx
		_ = req
		allowedCalls, rejected := filterActionCallsByCapabilities(resp.ActionCalls, cfg.Capabilities)
		allowedCalls, schemaRejected := validateActionCallsBySchema(allowedCalls, cfg.Capabilities)
		rejected = append(rejected, schemaRejected...)
		for _, call := range rejected {
			log.Printf("[autonomous:blocked] node=%s action=%s", targetNodeID, call.ActionID)
		}
		resp.ActionCalls = p.executeActions(req, runtime, executionMode, runtime.policyEngine, allowedCalls)

		if len(resp.ActionCalls) == 0 && len(resp.MemoryUpdates) == 0 {
			memUpdate := MemoryUpdate{NodeID: targetNodeID, Content: "自主行为周期未采取行动。", Level: MemShortTerm, Tags: "autonomous,no_action"}
			resp.MemoryUpdates = append(resp.MemoryUpdates, memUpdate)
		}

		now := time.Now()
		cfg.LastRunAt = &now
		cfg.LastError = ""
		if err := SaveAutonomousConfig(comp.UUID, cfg); err != nil {
			log.Printf("save autonomous runtime state: %v", err)
		}
		return resp
	}, executionMode)
}

func (p *Pipeline) executeCustom(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode, withTaskTree bool) (*InvokeResponse, error) {
	customFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		_ = round
		if isPlayerInputInterpretRequest(req) {
			return buildPlayerIntentPrompt(appendDynamicInterfaceContext(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), requestDynamicInterfaces(req)), nodeID, reqInteractionContext(req))
		}
		return treeContext
	}
	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}
	customToolFn := func(req *InvokeRequest, nodeID string, round int) []LLMToolDefinition {
		_ = nodeID
		_ = round
		return requestLLMTools(req, []LLMToolDefinition{builtinStoreRequestTool()})
	}
	finalizeFn := func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
		_ = ctx
		_ = req
		return applyParsedPlayerIntent(resp, parsed)
	}
	if !isPlayerInputInterpretRequest(req) {
		finalizeFn = nil
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, customFn, customToolFn, finalizeFn, executionMode)
}
func authorityDemoStatePath(worldID string) string {
	// Check common demo state file locations relative to the working directory.
	// In dev mode: tools/source/workerhome/demo/demo-state.yaml
	// In packaged mode: demo/demo-state.yaml (relative to working directory)
	candidates := []string{
		"tools/source/workerhome/demo/demo-state.yaml",
		"demo/demo-state.yaml",
		"./demo-state.yaml",
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if data, err := os.ReadFile(p); err == nil {
				if containsWorldID(data, worldID) {
					return p
				}
			}
		}
	}
	return ""
}

func containsWorldID(data []byte, worldID string) bool {
	prefix := []byte("world_id: " + worldID)
	return len(data) >= len(prefix) && string(data[:len(prefix)]) == string(prefix)
}
