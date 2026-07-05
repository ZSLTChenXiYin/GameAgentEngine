package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/action"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/planner"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// Pipeline 是一次 Agent 推理执行的核心协调器。
type Pipeline struct {
	ctxBuilder      *ContextBuilder
	llmProvider     LLMProvider
	policyEngine    *planner.PolicyEngine
	actionReg       *action.Registry
	subTaskRetries  int
	subTaskTimeout  int
	pipelineMode    PipelineMode
}

// NewPipeline 创建推理管线，并注册内置动作。
func NewPipeline(llmProvider LLMProvider) *Pipeline {
	p := &Pipeline{
		ctxBuilder:   NewContextBuilder(),
		llmProvider:  llmProvider,
		policyEngine: planner.NewPolicyEngine(),
		actionReg:    action.NewRegistry(),
		pipelineMode: PipelineFull,
	}
	p.registerBuiltinActions()
	return p
}

// registerBuiltinActions 注册引擎内置的同步和异步动作。
func (p *Pipeline) registerBuiltinActions() {
	p.actionReg.RegisterSync(&action.UpdateMood{})
	p.actionReg.RegisterSync(&action.AddMemory{})
	p.actionReg.RegisterSync(&action.SendDialogue{})
	p.actionReg.RegisterAsync(&action.AdjustRelation{})
	p.actionReg.RegisterAsync(&action.SpawnItem{})
}

// loadWorldSettings 从数据库加载世界运行设置，返回 memoryLimit 和 maxRounds。
func (p *Pipeline) loadWorldSettings(worldID string) (int, int, int, int, string) {
	s, err := store.GetOrCreateWorldSettings(worldID)
	if err != nil {
		return 50, 5, 2, 60, "full"
	}
	return s.MemoryLimit, s.MaxAnalysisRounds, s.SubTaskMaxRetries, s.SubTaskTimeoutSecs, s.PipelineMode
}

// loadWorldPolicy 从数据库加载世界策略，并设置到 policyEngine 中。
func (p *Pipeline) loadWorldPolicy(worldID string) {
	policy, err := store.GetWorldPolicy(worldID)
	if err != nil {
		p.policyEngine.SetActions(nil, nil)
		return
	}
	p.policyEngine.SetActions(policy.ParseBlockedActions(), policy.ParseSafeActions())
}

// ActionRegistry 返回当前管线使用的动作注册表。
func (p *Pipeline) ActionRegistry() *action.Registry {
	return p.actionReg
}

// Execute 执行一次完整的推理流程。
// 流程包括上下文构建、多轮 Prompt 生成、LLM 调用、动作解析和记忆持久化。
// 所有任务类型均支持 request_data 多轮数据和 game_client 异步回调。
// getExecutionMode 返回当前引擎的执行模式。
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

	// 获取执行模式
	executionMode := p.getExecutionMode()

	depth := 3
	if req.Context != nil && req.Context.MaxDepth > 0 {
		depth = req.Context.MaxDepth
	}

	// 从数据库加载世界运行设置和策略。
	memoryLimit, maxRounds, retries, timeout, pipelineMode := p.loadWorldSettings(req.WorldID)
	p.subTaskRetries = retries
	p.subTaskTimeout = timeout
	if pipelineMode != "" {
		p.pipelineMode = PipelineMode(pipelineMode)
	}
	includeRelated := false
	if req.Context != nil {
		if req.Context.MemoryLimit > 0 {
			memoryLimit = req.Context.MemoryLimit
		}
		if req.Context.MaxAnalysisRounds > 0 {
			maxRounds = req.Context.MaxAnalysisRounds
		}
		includeRelated = req.Context.IncludeRelatedNodes
	}

	p.loadWorldPolicy(req.WorldID)

	// 构建初始上下文。
	ctx, err := p.ctxBuilder.Build(req.NodeID, depth, memoryLimit, includeRelated)
	if err != nil {
		return nil, fmt.Errorf("build context: %w", err)
	}

	// 按任务类型执行。
	switch req.TaskType {
	case TaskNPCDialogue:
		return p.executeDialogue(req, ctx, start, requestID, maxRounds, executionMode)
	case TaskWorldTick:
		return p.executeWorldTick(req, ctx, start, requestID, maxRounds, executionMode)
	case TaskWorldEvent:
		return p.executeWorldEvent(req, ctx, start, requestID, maxRounds, executionMode)
	case TaskAutonomousAct:
		return p.executeAutonomousAct(req, ctx, start, requestID, maxRounds, executionMode)
	default:
		return p.executeCustom(req, ctx, start, requestID, maxRounds, executionMode)
	}
}
// RoundState 承载一次多轮推理的中间状态。
type RoundState struct {
	Context       *BuiltContext  // 当前累积的上下文
	Tree          *TaskTree      // 任务节点树，记录推理过程
	SystemPrompt  string         // 当轮构建的 system prompt
	Messages      []ChatMessage  // 用户消息
	TargetNodeID  string         // 当前推理的目标节点 ID
	MaxRounds     int            // 最大轮数
}

// executeMultiTurnLoop 是所有任务共享的多轮推理循环。
// 每轮创建一个 TaskNode，记录 prompt/response，上下文通过 BuildLLMContext() 构建。
// executeMultiTurnLoop 是所有任务共享的多轮推理循环。
// 每轮创建一个 TaskNode 记录完整推理轨迹，上下文通过树节点树继承。
func (p *Pipeline) executeMultiTurnLoop(
	req *InvokeRequest,
	ctx *BuiltContext,
	start time.Time,
	requestID string,
	maxRounds int,
	taskTree *TaskTree,
	taskPromptFn func(treeContext string, req *InvokeRequest, nodeID string, round int) string,
	finalizeFn func(*InvokeResponse, *llmParsedOutput, *BuiltContext, *InvokeRequest) *InvokeResponse,
	executionMode ExecutionMode,
) (*InvokeResponse, error) {

	targetNodeID := req.NodeID
	tree := taskTree
	if tree == nil {
		tree = NewTaskTree(req.TaskType, req.WorldID, req.NodeID)
	}

	state := &RoundState{
		Context:      ctx,
		Tree:         tree,
		Messages:     sanitizeRoles(req.Messages),
		TargetNodeID: targetNodeID,
		MaxRounds:    maxRounds,
	}

	for round := 0; round < maxRounds; round++ {
		// 创建本轮任务节点
		roundNode := tree.NewRound(fmt.Sprintf("round_%d", round+1))

		// 基础上下文来自任务节点树（包含所有历史轮次和查询结果）
		var treeContext string
		if tree != nil {
			treeContext = tree.BuildLLMContext()
			roundNode.Prompt = taskPromptFn(treeContext, req, targetNodeID, round)
			state.SystemPrompt = roundNode.Prompt
		} else {
			// 轮询模式：不创建 TaskTree，直接用上下文构建 prompt
			state.SystemPrompt = taskPromptFn(ctx.SystemPrompt, req, targetNodeID, round)
		}

		// 调用 LLM
		llmStart := time.Now()
		llmResp, err := p.llmProvider.Chat(state.SystemPrompt, state.Messages)
		if err != nil {
		}
		parsed := p.parseLLMJSON(llmResp.Content)
		roundNode.LLMResponse = llmResp.Content

		// 写入轮次中间记忆
		if parsed.RawInterimMemoryUpdates != "" {
			if imus := p.parseMemoryUpdates(parsed.RawInterimMemoryUpdates); len(imus) > 0 {
				writeMemories(imus)
			// interim 记忆也执行传播
			for _, imu := range imus {
				p.PropagateMemoryByRule(imu, imu.NodeID)
			}
			}
		}

		// 检查 LLM 是否请求额外数据
		if parsed.RawRequestData != "" {
			var dr DataRequest
			if err := json.Unmarshal([]byte(parsed.RawRequestData), &dr); err == nil && len(dr.Queries) > 0 {
				switch dr.Target {
				case "game_client":
					cbID := p.actionReg.CreateCallback("data_request", map[string]any{
						"label":   dr.Label,
						"queries": dr.Queries,
					})
					resp := &InvokeResponse{
						RequestID:   requestID,
						TaskType:    req.TaskType,
						DataRequest: &dr,
						ActionCalls: []ActionCall{{
							ActionID:   "data_request",
							Mode:       "async",
							CallbackID: cbID,
							Args:       map[string]any{"data_request": dr},
						}},
						Metadata: &ResponseMeta{
							LLMModel:         p.llmProvider.ModelName(),
							ProcessingTimeMs: time.Since(start).Milliseconds(),
						},
					}
					appendResponseLog(resp, req)
					return resp, nil

				default:
					result := p.handleDataRequest(&dr)
					roundNode.Analysis = result
					roundNode.Decision = "[数据查询] " + dr.Label
					continue
				}
			}
		}

		// 没有数据请求——记录分析结果并构建最终响应
		if tree != nil {
			roundNode.Analysis = parsed.Reply
		}
		if parsed.RawActionCalls != "" {
			if tree != nil {
				roundNode.Decision = fmt.Sprintf("动作: %s", truncateForContext(parsed.RawActionCalls, 100))
			}
		}

		resp := &InvokeResponse{
			RequestID:     requestID,
			TaskType:      req.TaskType,
			ExecutionMode: ExecutionMode(config.ExecutionMode()),
			Reply:         parsed.Reply,
			ActionCalls:   p.executeActions(p.parseActionCalls(parsed.RawActionCalls, targetNodeID)),
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
			mm := store.MemoryModel{
				NodeID:  memNodeID,
				Content: mem.Content,
				Level:   string(mem.Level),
				Tags:    mem.Tags,
			}
			if err := store.CreateMemory(&mm); err != nil {
				log.Printf("write memory: %v", err)
			}
		}
		for _, mem := range resp.MemoryUpdates {
			p.PropagateMemoryByRule(mem, mem.NodeID)
		}

		resp.Metadata = &ResponseMeta{
			LLMModel:         p.llmProvider.ModelName(),
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}
		// 解析并执行 LLM 声明的子任务（DAG 编排，仅全功能模式）
		if tree != nil && parsed.RawSubTasks != "" {
			var subTasks []SubTaskDeclaration
			if err := json.Unmarshal([]byte(parsed.RawSubTasks), &subTasks); err == nil && len(subTasks) > 0 {
				dag := NewDAGInstance(tree, p.llmProvider, p.subTaskRetries, time.Duration(p.subTaskTimeout)*time.Second)
				for _, st := range subTasks {
					if err := dag.Register(st); err != nil {
						log.Printf("[dag] register sub-task %s: %v", st.Label, err)
					}
				}
				// 循环执行就绪子任务，直到全部完成
				for dag.HasReady() {
					for _, st := range dag.ReadyTasks() {
						dag.MarkRunning(st.Label)
						log.Printf("[dag] executing sub-task %s (type=%s node=%s)", st.Label, st.TaskType, st.NodeID)

						subReq := &InvokeRequest{
							WorldID:  req.WorldID,
							TaskType: st.TaskType,
							NodeID:   st.NodeID,
							Context:  req.Context,
						}
						subResp, err := p.Execute(subReq)
						if err != nil {
							log.Printf("[dag] sub-task %s failed: %v", st.Label, err)
							dag.OnTaskFailed(st.Label, err)
						} else {
							log.Printf("[dag] sub-task %s completed", st.Label)
							dag.OnTaskComplete(st.Label, subResp)
						}
					}
				}
				// 汇聚子任务结果
				merged := dag.MergeResults()
				resp.SubTasks = subTasks
				// 将汇聚结果合并到当前响应
				if merged.Reply != "" {
					resp.Reply = merged.Reply
				}
				resp.ActionCalls = append(resp.ActionCalls, merged.ActionCalls...)
				resp.MemoryUpdates = append(resp.MemoryUpdates, merged.MemoryUpdates...)
			}
		}
		// ExecutionMode 分支行为
		switch executionMode {
		case ModeDebug:
			// Debug 模式：记录 LLM 调用 trace（含各步骤耗时）
			errStr := ""
			trace := buildDebugTrace(
				req.WorldID, requestID,
				req.TaskType, targetNodeID,
				start, llmStart,
				state.SystemPrompt, state.Messages,
				parsed.Reply,
				resp, round, errStr,
			)
			GlobalTraceRing.Push(trace)
		case ModeReview:
			// Review 模式：检查世界变更计划是否需审批
			if resp.WorldChangePlan != nil && IsHighImpact(resp.WorldChangePlan.ImpactLevel) {
				planID := NewPendingPlanID()
				pendingPlan := &PendingPlan{
					PlanID:          planID,
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
				// 清空动作和记忆，等待审批
				parsed.RawActionCalls = ""
				parsed.RawMemoryUpdates = ""
				resp.ActionCalls = nil
				resp.MemoryUpdates = nil
				resp.ExecutionMode = ModeReview
			}

		default:
			// Production 模式：不动（全自动执行）
		}

		if executionMode == ModeReview && resp.WorldChangePlan != nil && IsHighImpact(resp.WorldChangePlan.ImpactLevel) {
			resp.Reply = "[待审批] 变更计划已挂起，请调用审批 API 确认执行"
		}

		appendResponseLog(resp, req)
		return resp, nil
	}

	return nil, fmt.Errorf("%s exceeded max rounds (%d)", req.TaskType, maxRounds)
}

// appendResponseLog 记录推理日志到数据库。
func appendResponseLog(resp *InvokeResponse, req *InvokeRequest) {
	store.CreateInferenceLog(&store.InferenceLogModel{
		WorldUUID:  req.WorldID,
		TaskType:   string(req.TaskType),
		NodeUUID:   req.NodeID,
		LLMModel:   resp.Metadata.LLMModel,
		DurationMs: resp.Metadata.ProcessingTimeMs,
	})
}

// executeVertical 执行垂直管线模式：一次 LLM 调用，无轮询、无任务节点树、无 DAG 子任务。
func (p *Pipeline) executeVertical(req *InvokeRequest, start time.Time, requestID string) (*InvokeResponse, error) {
	// 构建简化 prompt（不经过 ContextBuilder，直接根据任务类型生成）
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
			eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s",
				req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
		}
		systemPrompt = buildEventImpactPrompt(ctxDesc, eventDesc, req.NodeID)
	case TaskAutonomousAct:
		if cfg, _, err := LoadAutonomousConfig(req.NodeID); err == nil && cfg != nil && cfg.Enabled {
			systemPrompt = buildAutonomousPrompt(ctxDesc, req.NodeID, cfg)
		} else {
			resp := &InvokeResponse{RequestID: requestID, TaskType: req.TaskType, Reply: "autonomous component not found or disabled", ExecutionMode: ModeDebug}
			resp.Metadata = &ResponseMeta{LLMModel: p.llmProvider.ModelName(), ProcessingTimeMs: time.Since(start).Milliseconds()}
			appendResponseLog(resp, req)
			return resp, nil
		}
	default:
		systemPrompt = ctxDesc
	}

	// 单次 LLM 调用
	messages := sanitizeRoles(req.Messages)
	llmResp, err := p.llmProvider.Chat(systemPrompt, messages)
	if err != nil {
		return nil, fmt.Errorf("vertical llm: %w", err)
	}

	// 解析输出
	parsed := p.parseLLMJSON(llmResp.Content)
	resp := &InvokeResponse{
		RequestID:     requestID,
		TaskType:      req.TaskType,
		ExecutionMode: ModeDebug,
		Reply:         parsed.Reply,
		ActionCalls:   p.executeActions(p.parseActionCalls(parsed.RawActionCalls, req.NodeID)),
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

	// 写记忆
	for _, mem := range resp.MemoryUpdates {
		nodeID := store.ResolveNodeUUID(mem.NodeID)
		if nodeID == 0 {
			log.Printf("[warn] write memory: unknown node UUID %s", mem.NodeID)
			continue
		}
		mm := store.MemoryModel{NodeID: nodeID, Content: mem.Content, Level: string(mem.Level), Tags: mem.Tags}
		store.CreateMemory(&mm)
	}
	for _, mem := range resp.MemoryUpdates {
		p.PropagateMemoryByRule(mem, mem.NodeID)
	}

	resp.Metadata = &ResponseMeta{LLMModel: p.llmProvider.ModelName(), ProcessingTimeMs: time.Since(start).Milliseconds()}
	appendResponseLog(resp, req)
	return resp, nil
}

// executeDialogue 执行 NPC 对话推理。
// 第一轮做局势分析（数值→定性），结果写入任务节点树，后续走公共多轮循环。
func (p *Pipeline) executeDialogue(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, maxRounds int, executionMode ExecutionMode) (*InvokeResponse, error) {
	tree := NewTaskTree(req.TaskType, req.WorldID, req.NodeID)

	// 第 0 轮：内部局势分析（独立于公共循环）
	analysisNode := tree.NewRound("analysis")
	analysisPrompt := fmt.Sprintf(
		"请分析以下局势数据，并将其中的精确数值转化为模糊量词（如用「紧张」「充裕」「堪忧」「尚可」「严峻」「好转」替代具体数字），整理成后续对话可用的局势摘要。\n\n%s",
		ctx.SystemPrompt,
	)
	analysisNode.Prompt = analysisPrompt
	if analysisResp, err := p.llmProvider.Chat(analysisPrompt, nil); err == nil {
		analysisNode.LLMResponse = analysisResp.Content
		analysisNode.Analysis = analysisResp.Content
		analysisNode.Decision = "局势分析完成"
		// 将分析结果存入树，后续轮次通过 BuildLLMContext() 自动包含
	}

	dialogueFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildDialoguePrompt(treeContext, nodeID)
	}

	// 剩余轮次走公共循环
	return p.executeMultiTurnLoop(req, ctx, start, requestID, maxRounds-1, tree, dialogueFn, nil, executionMode)
}

// executeWorldTick 执行世界刻推进推理。
func (p *Pipeline) executeWorldTick(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, maxRounds int, executionMode ExecutionMode) (*InvokeResponse, error) {
	var currentOutline string
	if latest, err := store.GetLatestTick(req.WorldID); err == nil {
		currentOutline = latest.FutureOutline
	}

	tickFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildWorldTickPrompt(treeContext, currentOutline)
	}

	return p.executeMultiTurnLoop(req, ctx, start, requestID, maxRounds, nil, tickFn, nil, executionMode)
}

// executeWorldEvent 执行世界事件影响评估推理。
func (p *Pipeline) executeWorldEvent(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, maxRounds int, executionMode ExecutionMode) (*InvokeResponse, error) {
	eventDesc := ""
	if req.Event != nil {
		eventDesc = fmt.Sprintf("事件类型:%s 范围:%s 描述:%s 严重度:%s",
			req.Event.EventType, req.Event.ScopeID, req.Event.Description, req.Event.Severity)
	}

	eventFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		return buildEventImpactPrompt(treeContext, eventDesc, nodeID)
	}

	return p.executeMultiTurnLoop(req, ctx, start, requestID, maxRounds, nil, eventFn, nil, executionMode)
}

// executeAutonomousAct 执行自主行为推理，带 capability 过滤。
func (p *Pipeline) executeAutonomousAct(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, maxRounds int, executionMode ExecutionMode) (*InvokeResponse, error) {
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

	return p.executeMultiTurnLoop(req, ctx, start, requestID, maxRounds, nil, autonomousFn, func(resp *InvokeResponse, parsed *llmParsedOutput, ctx *BuiltContext, req *InvokeRequest) *InvokeResponse {
		_ = parsed
		_ = ctx
		_ = req
		allowedCalls, rejected := filterActionCallsByCapabilities(resp.ActionCalls, cfg.Capabilities)
		allowedCalls, schemaRejected := validateActionCallsBySchema(allowedCalls, cfg.Capabilities)
		rejected = append(rejected, schemaRejected...)
		for _, call := range rejected {
			log.Printf("[autonomous:blocked] node=%s action=%s", targetNodeID, call.ActionID)
		}
		resp.ActionCalls = p.executeActions(allowedCalls)

		if len(resp.ActionCalls) == 0 && len(resp.MemoryUpdates) == 0 {
			memUpdate := MemoryUpdate{
				NodeID:  targetNodeID,
				Content: "自主行为周期未采取行动。",
				Level:   MemShortTerm,
				Tags:    "autonomous,no_action",
			}
			resp.MemoryUpdates = append(resp.MemoryUpdates, memUpdate)
			nodeID2 := store.ResolveNodeUUID(targetNodeID)
		if nodeID2 == 0 {
			log.Printf("[warn] pipeline write memory: unknown node UUID %s", targetNodeID)
			return resp
		}
		mm := store.MemoryModel{NodeID: nodeID2, Content: memUpdate.Content, Level: string(memUpdate.Level), Tags: memUpdate.Tags}
		store.CreateMemory(&mm)
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

// executeCustom 执行自定义推理。
func (p *Pipeline) executeCustom(req *InvokeRequest, ctx *BuiltContext, start time.Time, requestID string, maxRounds int, executionMode ExecutionMode) (*InvokeResponse, error) {
	customFn := func(treeContext string, req *InvokeRequest, nodeID string, round int) string {
		_ = nodeID
		_ = round
		return treeContext
	}
	return p.executeMultiTurnLoop(req, ctx, start, requestID, maxRounds, nil, customFn, nil, executionMode)
}








