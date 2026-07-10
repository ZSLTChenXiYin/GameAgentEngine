package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/action"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/planner"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
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
}

type executionConfig struct {
	memoryLimit            int
	maxRounds              int
	subTaskRetries         int
	subTaskTimeout         int
	configuredPipelineMode PipelineMode
	pipelineMode           PipelineMode
	policyEngine           *planner.PolicyEngine
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
	}
	p.registerBuiltinActions()
	return p
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
	executionMode := p.getExecutionMode()
	p.emitLog(req, nil, nil, executionMode, pipelineLogEvent{
		Category:  "pipeline",
		EventName: "request_started",
		Message:   fmt.Sprintf("start task %s", req.TaskType),
		DetailData: marshalLogDetail(map[string]any{
			"request_id": requestID,
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
	}

	ctx, err := p.ctxBuilder.Build(req.TaskType, req.NodeID, depth, runtime.memoryLimit, includeRelated)
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
func (p *Pipeline) ResumePausedExecution(callbackID string, result any) (*store.PausedExecutionModel, *InvokeRequest, *BuiltContext, *RoundState, *executionConfig, *DataRequest, ExecutionMode, int, error) {
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

type RoundState struct {
	Context             *BuiltContext
	Tree                *TaskTree
	TreeContext         string
	SystemPrompt        string
	Messages            []ChatMessage
	TargetNodeID        string
	MaxRounds           int
	SupplementalContext []string
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
	RoundState   *RoundState           `json:"round_state"`
	Runtime      pausedExecutionRuntime `json:"runtime"`
	DataRequest  *DataRequest          `json:"data_request,omitempty"`
	PausedRound  int                   `json:"paused_round"`
	CallbackID   string                `json:"callback_id"`
	ExecutionID  string                `json:"execution_id"`
}

func (s *RoundState) buildPrompt(base string) string {
	if len(s.SupplementalContext) == 0 {
		return base
	}
	return strings.TrimSpace(base + "\n\n补充上下文:\n" + strings.Join(s.SupplementalContext, "\n"))
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

func (p *Pipeline) executeMultiTurnLoop(
	req *InvokeRequest,
	ctx *BuiltContext,
	start time.Time,
	requestID string,
	runtime *executionConfig,
	taskTree *TaskTree,
	taskPromptFn func(treeContext string, req *InvokeRequest, nodeID string, round int) string,
	finalizeFn func(*InvokeResponse, *llmParsedOutput, *BuiltContext, *InvokeRequest) *InvokeResponse,
	executionMode ExecutionMode,
) (*InvokeResponse, error) {
	targetNodeID := req.NodeID
	state := &RoundState{
		Context:      ctx,
		Tree:         taskTree,
		TreeContext:  "",
		Messages:     sanitizeRoles(req.Messages),
		TargetNodeID: targetNodeID,
		MaxRounds:    runtime.maxRounds,
	}

	for round := 0; round < runtime.maxRounds; round++ {
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
		p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
			Category:   "pipeline_round",
			EventName:  "prompt_prepared",
			Message:    fmt.Sprintf("round %d prompt prepared", round+1),
			Round:      round + 1,
			DetailData: buildRoundLogDetail(state.SystemPrompt, state.Messages, round+1, targetNodeID, state.Tree),
		})

		llmStart := time.Now()
		llmResp, err := p.llmProvider.Chat(state.SystemPrompt, state.Messages)
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
			DetailData: buildLLMResponseDetail(llmResp.Content, parsed),
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
					resp := &InvokeResponse{
						RequestID:   requestID,
						ExecutionMode: executionMode,
						TaskType:    req.TaskType,
						DataRequest: &dr,
						ActionCalls: []ActionCall{{
							ActionID:   "data_request",
							Mode:       "async",
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
					appendResponseLog(resp, req)
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
					}
					appendRoundStateTreeEntry(state, round+1, parsed, result)
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

		resp := &InvokeResponse{
			RequestID:     requestID,
			TaskType:      req.TaskType,
			ExecutionMode: executionMode,
			Reply:         parsed.Reply,
			ActionCalls:   p.executeActions(req, runtime, executionMode, runtime.policyEngine, p.parseActionCalls(parsed.RawActionCalls, targetNodeID)),
			MemoryUpdates: p.parseMemoryUpdates(parsed.RawMemoryUpdates),
		}
		if parsed.RawPlan != "" {
			var wcp WorldChangePlan
			if err := json.Unmarshal([]byte(parsed.RawPlan), &wcp); err == nil {
				resp.WorldChangePlan = &wcp
			}
		}
		if parsed.RawFutureOutline != "" {
			resp.FutureOutline = parsed.RawFutureOutline
		}

		if finalizeFn != nil {
			resp = finalizeFn(resp, parsed, state.Context, req)
		}

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

		if state.Tree != nil && runtime.pipelineMode == PipelineFull && parsed.RawSubTasks != "" {
			var subTasks []SubTaskDeclaration
			if err := json.Unmarshal([]byte(parsed.RawSubTasks), &subTasks); err == nil && len(subTasks) > 0 {
				p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
					Category:   "pipeline_round",
					EventName:  "sub_tasks_declared",
					Message:    fmt.Sprintf("round %d declared %d sub tasks", round+1, len(subTasks)),
					Round:      round + 1,
					DetailData: marshalLogDetail(subTasks),
				})
				dag := NewDAGInstance(state.Tree, p.llmProvider, runtime.subTaskRetries, time.Duration(runtime.subTaskTimeout)*time.Second)
				for _, st := range subTasks {
					if err := dag.Register(st); err != nil {
						log.Printf("[dag] register sub-task %s: %v", st.Label, err)
					}
				}
				for dag.HasReady() {
					for _, st := range dag.ReadyTasks() {
						dag.MarkRunning(st.Label)
						subReq := &InvokeRequest{WorldID: req.WorldID, TaskType: st.TaskType, NodeID: st.NodeID, Context: req.Context}
						subResp, err := p.Execute(subReq)
						if err != nil {
							log.Printf("[dag] sub-task %s failed: %v", st.Label, err)
							dag.OnTaskFailed(st.Label, err)
						} else {
							dag.OnTaskComplete(st.Label, subResp)
						}
					}
				}
				merged := dag.MergeResults()
				resp.SubTasks = subTasks
				if merged.Reply != "" {
					resp.Reply = merged.Reply
				}
				resp.ActionCalls = append(resp.ActionCalls, merged.ActionCalls...)
				resp.MemoryUpdates = append(resp.MemoryUpdates, merged.MemoryUpdates...)
			}
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
		case ModeReview:
			if resp.WorldChangePlan != nil && IsHighImpact(resp.WorldChangePlan.ImpactLevel) {
				pendingPlan := &PendingPlan{
					PlanID:          NewPendingPlanID(),
					WorldID:         req.WorldID,
					TickNumber:      0,
					TaskType:        req.TaskType,
					WorldChangePlan: resp.WorldChangePlan,
					ActionCalls:     resp.ActionCalls,
					MemoryUpdates:   resp.MemoryUpdates,
					CreatedAt:       time.Now(),
					Status:          "pending",
				}
				GlobalPlanReview.Add(pendingPlan)
				p.emitLog(req, resp, runtime, executionMode, pipelineLogEvent{
					Category:   "pipeline_review",
					EventName:  "plan_pending_review",
					Message:    resp.WorldChangePlan.Summary,
					Round:      round + 1,
					DetailData: marshalLogDetail(pendingPlan),
				})
				resp.ActionCalls = nil
				resp.MemoryUpdates = nil
				resp.ExecutionMode = ModeReview
			}
		}

		if executionMode == ModeReview && resp.WorldChangePlan != nil && IsHighImpact(resp.WorldChangePlan.ImpactLevel) {
			resp.Reply = "[待审批] 变更计划已挂起，请调用审批 API 确认执行"
		}

		appendResponseLog(resp, req)
		return resp, nil
	}

	return nil, fmt.Errorf("%s exceeded max rounds (%d)", req.TaskType, runtime.maxRounds)
}

func appendResponseLog(resp *InvokeResponse, req *InvokeRequest) {
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
	logger := &Pipeline{}
	logger.emitLog(req, resp, runtime, mode, pipelineLogEvent{
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
		systemPrompt = buildDialoguePrompt(ctxDesc, req.NodeID)
	case TaskWorldTick:
		systemPrompt = buildWorldTickPrompt(ctxDesc, "", nil, nil, buildWorldTickTimeBlock(req.WorldID), "")
	case TaskWorldEvent:
		eventDesc := ""
		if req.Event != nil {
			eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s", req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
		}
		systemPrompt = buildEventImpactPrompt(ctxDesc, eventDesc, req.NodeID)
	case TaskAutonomousAct:
		if cfg, _, err := LoadAutonomousConfig(req.NodeID); err == nil && cfg != nil && cfg.Enabled {
			systemPrompt = buildAutonomousPrompt(ctxDesc, req.NodeID, cfg)
		} else {
			resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous component not found or disabled", ExecutionMode: executionMode}
			resp.Metadata = buildResponseMeta(runtime, p.llmProvider.ModelName(), 0, start, 0)
			appendResponseLog(resp, req)
			return resp, nil
		}
	default:
		systemPrompt = ctxDesc
	}

	llmResp, err := p.llmProvider.Chat(systemPrompt, sanitizeRoles(req.Messages))
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
	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline_round",
		EventName:  "prompt_prepared",
		Message:    "vertical prompt prepared",
		Round:      1,
		DetailData: buildRoundLogDetail(systemPrompt, sanitizeRoles(req.Messages), 1, req.NodeID, nil),
	})
	p.emitLog(req, nil, runtime, executionMode, pipelineLogEvent{
		Category:   "pipeline_round",
		EventName:  "llm_response_received",
		Message:    truncateForLog(parsed.Reply, 180),
		Round:      1,
		TokensUsed: llmResp.Tokens,
		DurationMs: time.Since(start).Milliseconds(),
		DetailData: buildLLMResponseDetail(llmResp.Content, parsed),
	})
	resp := &InvokeResponse{
		RequestID:     requestID,
		TaskType:      req.TaskType,
		ExecutionMode: executionMode,
		Reply:         parsed.Reply,
		ActionCalls:   p.executeActions(req, runtime, executionMode, runtime.policyEngine, p.parseActionCalls(parsed.RawActionCalls, req.NodeID)),
		MemoryUpdates: p.parseMemoryUpdates(parsed.RawMemoryUpdates),
	}
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
	appendResponseLog(resp, req)
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
		if analysisResp, err := p.llmProvider.Chat(analysisPrompt, nil); err == nil {
			analysisNode.LLMResponse = analysisResp.Content
			analysisNode.Analysis = analysisResp.Content
			analysisNode.Decision = "局势分析完成"
		}
	}

	dialogueFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildDialoguePrompt(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), nodeID)
	}

	loopRuntime := *runtime
	if loopRuntime.maxRounds > 1 && withTaskTree {
		loopRuntime.maxRounds--
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, &loopRuntime, tree, dialogueFn, nil, executionMode)
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

	tickFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		baseContext := mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext)
		return buildWorldTickPrompt(baseContext, currentOutline, ctx.StateBlocks, recentTimeline, worldTimeBlock, relationSummary)
	}

	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, tickFn, func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
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

func (p *Pipeline) executeWorldEvent(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode, withTaskTree bool) (*InvokeResponse, error) {
	eventDesc := ""
	if req.Event != nil {
		eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s", req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
	}

	eventFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildEventImpactPrompt(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), eventDesc, nodeID)
	}

	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, eventFn, nil, executionMode)
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
		return buildAutonomousPrompt(mergeBaseAndTreeContext(ctx.SystemPrompt, treeContext), nodeID, cfg)
	}

	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}

	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, autonomousFn, func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
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
			nodeID2 := store.ResolveNodeUUID(targetNodeID)
			if nodeID2 == 0 {
				log.Printf("[warn] pipeline write memory: unknown node UUID %s", targetNodeID)
				return resp
			}
			mm := store.MemoryModel{NodeID: nodeID2, Content: memUpdate.Content, Level: string(memUpdate.Level), Tags: memUpdate.Tags}
			if err := store.CreateMemory(&mm); err != nil {
				log.Printf("write memory: %v", err)
			}
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
		_ = nodeID
		_ = round
		return treeContext
	}
	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, customFn, nil, executionMode)
}
