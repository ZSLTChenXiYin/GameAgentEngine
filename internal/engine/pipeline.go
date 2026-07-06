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

// Pipeline is the shared execution shell. Request-specific state is created per Execute call.
type Pipeline struct {
	ctxBuilder  *ContextBuilder
	llmProvider LLMProvider
	actionReg   *action.Registry
}

type executionConfig struct {
	memoryLimit    int
	maxRounds      int
	subTaskRetries int
	subTaskTimeout int
	pipelineMode   PipelineMode
	policyEngine   *planner.PolicyEngine
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
	return s.MemoryLimit, s.MaxAnalysisRounds, s.SubTaskMaxRetries, s.SubTaskTimeoutSecs, s.PipelineMode
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
	start := time.Now()
	requestID := uuid.NewString()
	executionMode := p.getExecutionMode()

	depth := 3
	if req.Context != nil && req.Context.MaxDepth > 0 {
		depth = req.Context.MaxDepth
	}

	memoryLimit, maxRounds, retries, timeout, pipelineMode := p.loadWorldSettings(req.WorldID)
	mode := PipelineMode(pipelineMode)
	if mode == "" {
		mode = PipelineFull
	}

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
		memoryLimit:    memoryLimit,
		maxRounds:      maxRounds,
		subTaskRetries: retries,
		subTaskTimeout: timeout,
		pipelineMode:   mode,
		policyEngine:   p.loadWorldPolicy(req.WorldID),
	}

	ctx, err := p.ctxBuilder.Build(req.NodeID, depth, runtime.memoryLimit, includeRelated)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	switch runtime.pipelineMode {
	case PipelineVertical:
		return p.executeVertical(req, start, requestID, runtime, executionMode)
	case PipelinePolling:
		return p.executePolling(req, ctx, start, requestID, runtime, executionMode)
	default:
		return p.executeFull(req, ctx, start, requestID, runtime, executionMode)
	}
}

type RoundState struct {
	Context             *BuiltContext
	Tree                *TaskTree
	SystemPrompt        string
	Messages            []ChatMessage
	TargetNodeID        string
	MaxRounds           int
	SupplementalContext []string
}

func (s *RoundState) buildPrompt(base string) string {
	if len(s.SupplementalContext) == 0 {
		return base
	}
	return strings.TrimSpace(base + "\n\n补充上下文:\n" + strings.Join(s.SupplementalContext, "\n"))
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
		Messages:     sanitizeRoles(req.Messages),
		TargetNodeID: targetNodeID,
		MaxRounds:    runtime.maxRounds,
	}

	for round := 0; round < runtime.maxRounds; round++ {
		var roundNode *TaskNode
		var promptSeed string
		if state.Tree != nil {
			roundNode = state.Tree.NewRound(fmt.Sprintf("round_%d", round+1))
			promptSeed = state.Tree.BuildLLMContext()
			state.SystemPrompt = taskPromptFn(promptSeed, req, targetNodeID, round)
			roundNode.Prompt = state.SystemPrompt
		} else {
			promptSeed = ctx.SystemPrompt
			state.SystemPrompt = state.buildPrompt(taskPromptFn(promptSeed, req, targetNodeID, round))
		}

		llmStart := time.Now()
		llmResp, err := p.llmProvider.Chat(state.SystemPrompt, state.Messages)
		if err != nil {
			return nil, fmt.Errorf("llm chat: %w", err)
		}

		parsed := p.parseLLMJSON(llmResp.Content)
		if roundNode != nil {
			roundNode.LLMResponse = llmResp.Content
		}

		if parsed.RawInterimMemoryUpdates != "" {
			if imus := p.parseMemoryUpdates(parsed.RawInterimMemoryUpdates); len(imus) > 0 {
				writeMemories(imus)
				for _, imu := range imus {
					p.PropagateMemoryByRule(imu, imu.NodeID)
				}
			}
		}

		if parsed.RawRequestData != "" {
			var dr DataRequest
			if err := json.Unmarshal([]byte(parsed.RawRequestData), &dr); err == nil && len(dr.Queries) > 0 {
				switch dr.Target {
				case "game_client":
					resp := &InvokeResponse{
						RequestID:   requestID,
						TaskType:    req.TaskType,
						DataRequest: &dr,
						ActionCalls: []ActionCall{{
							ActionID:   "data_request",
							Mode:       "async",
							CallbackID: p.actionReg.CreateCallback("data_request", map[string]any{"label": dr.Label, "queries": dr.Queries}),
							Args:       map[string]any{"data_request": dr},
						}},
						Metadata: &ResponseMeta{
							LLMModel:         p.llmProvider.ModelName(),
							TokensUsed:       llmResp.Tokens,
							ProcessingTimeMs: time.Since(start).Milliseconds(),
						},
					}
					appendResponseLog(resp, req)
					return resp, nil
				default:
					result := p.handleDataRequest(runtime.policyEngine, &dr)
					if roundNode != nil {
						roundNode.Analysis = result
						roundNode.Decision = "[数据查询] " + dr.Label
					} else {
						state.SupplementalContext = append(state.SupplementalContext, "[数据查询] "+dr.Label, result)
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

		resp := &InvokeResponse{
			RequestID:     requestID,
			TaskType:      req.TaskType,
			ExecutionMode: executionMode,
			Reply:         parsed.Reply,
			ActionCalls:   p.executeActions(runtime.policyEngine, p.parseActionCalls(parsed.RawActionCalls, targetNodeID)),
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

		for _, mem := range resp.MemoryUpdates {
			memNodeID := store.ResolveNodeUUID(mem.NodeID)
			if memNodeID == 0 {
				log.Printf("[warn] write memory: unknown node UUID %s", mem.NodeID)
				continue
			}
			mm := store.MemoryModel{NodeID: memNodeID, Content: mem.Content, Level: string(mem.Level), Tags: mem.Tags}
			if err := store.CreateMemory(&mm); err != nil {
				log.Printf("write memory: %v", err)
			}
		}
		for _, mem := range resp.MemoryUpdates {
			p.PropagateMemoryByRule(mem, mem.NodeID)
		}

		resp.Metadata = &ResponseMeta{
			LLMModel:         p.llmProvider.ModelName(),
			TokensUsed:       llmResp.Tokens,
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}

		if state.Tree != nil && runtime.pipelineMode == PipelineFull && parsed.RawSubTasks != "" {
			var subTasks []SubTaskDeclaration
			if err := json.Unmarshal([]byte(parsed.RawSubTasks), &subTasks); err == nil && len(subTasks) > 0 {
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
				resp, round, "",
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
	store.CreateInferenceLog(&store.InferenceLogModel{
		WorldUUID:  req.WorldID,
		TaskType:   string(req.TaskType),
		NodeUUID:   req.NodeID,
		LLMModel:   resp.Metadata.LLMModel,
		TokensUsed: resp.Metadata.TokensUsed,
		DurationMs: resp.Metadata.ProcessingTimeMs,
	})
}

func (p *Pipeline) executeVertical(req *InvokeRequest, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode) (*InvokeResponse, error) {
	var systemPrompt string
	ctxDesc := fmt.Sprintf("世界: %s, 节点: %s, 任务类型: %s", req.WorldID, req.NodeID, req.TaskType)
	switch req.TaskType {
	case TaskNPCDialogue:
		systemPrompt = buildDialoguePrompt(ctxDesc, req.NodeID)
	case TaskWorldTick:
		systemPrompt = buildWorldTickPrompt(ctxDesc, "")
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
			resp.Metadata = &ResponseMeta{LLMModel: p.llmProvider.ModelName(), ProcessingTimeMs: time.Since(start).Milliseconds()}
			appendResponseLog(resp, req)
			return resp, nil
		}
	default:
		systemPrompt = ctxDesc
	}

	llmResp, err := p.llmProvider.Chat(systemPrompt, sanitizeRoles(req.Messages))
	if err != nil {
		return nil, fmt.Errorf("vertical llm: %w", err)
	}

	parsed := p.parseLLMJSON(llmResp.Content)
	resp := &InvokeResponse{
		RequestID:     requestID,
		TaskType:      req.TaskType,
		ExecutionMode: executionMode,
		Reply:         parsed.Reply,
		ActionCalls:   p.executeActions(runtime.policyEngine, p.parseActionCalls(parsed.RawActionCalls, req.NodeID)),
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
	for _, mem := range resp.MemoryUpdates {
		nodeID := store.ResolveNodeUUID(mem.NodeID)
		if nodeID == 0 {
			log.Printf("[warn] write memory: unknown node UUID %s", mem.NodeID)
			continue
		}
		mm := store.MemoryModel{NodeID: nodeID, Content: mem.Content, Level: string(mem.Level), Tags: mem.Tags}
		if err := store.CreateMemory(&mm); err != nil {
			log.Printf("write memory: %v", err)
		}
	}
	for _, mem := range resp.MemoryUpdates {
		p.PropagateMemoryByRule(mem, mem.NodeID)
	}
	resp.Metadata = &ResponseMeta{LLMModel: p.llmProvider.ModelName(), TokensUsed: llmResp.Tokens, ProcessingTimeMs: time.Since(start).Milliseconds()}
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
		return buildDialoguePrompt(treeContext, nodeID)
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

	tickFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildWorldTickPrompt(treeContext, currentOutline)
	}

	var tree *TaskTree
	if withTaskTree {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, runtime, tree, tickFn, nil, executionMode)
}

func (p *Pipeline) executeWorldEvent(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, runtime *executionConfig, executionMode ExecutionMode, withTaskTree bool) (*InvokeResponse, error) {
	eventDesc := ""
	if req.Event != nil {
		eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s", req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
	}

	eventFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildEventImpactPrompt(treeContext, eventDesc, nodeID)
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
		resp.Metadata = &ResponseMeta{LLMModel: p.llmProvider.ModelName(), ProcessingTimeMs: time.Since(start).Milliseconds()}
		return resp, nil
	}
	if !cfg.Enabled {
		resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous behavior disabled"}
		resp.Metadata = &ResponseMeta{LLMModel: p.llmProvider.ModelName(), ProcessingTimeMs: time.Since(start).Milliseconds()}
		return resp, nil
	}
	if len(cfg.Capabilities) == 0 {
		resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous behavior has no capabilities"}
		resp.Metadata = &ResponseMeta{LLMModel: p.llmProvider.ModelName(), ProcessingTimeMs: time.Since(start).Milliseconds()}
		return resp, nil
	}

	autonomousFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildAutonomousPrompt(treeContext, nodeID, cfg)
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
		resp.ActionCalls = p.executeActions(runtime.policyEngine, allowedCalls)

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
