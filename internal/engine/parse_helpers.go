package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/action"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/external"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/planner"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func dynamicInterfacesByKind(req *InvokeRequest, kind DynamicInterfaceKind) []DynamicInterface {
	if req == nil || req.Context == nil || len(req.Context.DynamicInterfaces) == 0 {
		return nil
	}
	var result []DynamicInterface
	for _, item := range req.Context.DynamicInterfaces {
		if item.Kind == kind {
			result = append(result, item)
		}
	}
	return result
}

func resolveDynamicInterface(req *InvokeRequest, kind DynamicInterfaceKind, reference string) (*DynamicInterface, error) {
	items := dynamicInterfacesByKind(req, kind)
	if len(items) == 0 {
		if req == nil || req.Context == nil || len(req.Context.DynamicInterfaces) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("no %s interfaces allowed for this request", kind)
	}
	ref := strings.TrimSpace(reference)
	if ref != "" {
		for _, item := range items {
			if ref == item.ID || ref == item.ExternalInterface {
				matched := item
				return &matched, nil
			}
		}
		return nil, fmt.Errorf("%s interface %q not allowed for this request", kind, ref)
	}
	if len(items) == 1 {
		matched := items[0]
		return &matched, nil
	}
	return nil, fmt.Errorf("multiple %s interfaces allowed; external_interface is required", kind)
}

func normalizeDynamicDataRequest(req *InvokeRequest, dr *DataRequest) error {
	if dr == nil || dr.Target != "game_client" {
		return nil
	}
	matched, err := resolveDynamicInterface(req, DynamicInterfaceDataRequest, dr.ExternalInterface)
	if err != nil {
		return err
	}
	if matched == nil {
		return nil
	}
	if matched.MaxQueries > 0 && len(dr.Queries) > matched.MaxQueries {
		return fmt.Errorf("data_request exceeds max_queries for interface %q", matched.ID)
	}
	allowedQueryTypes := make(map[string]struct{}, len(matched.QueryTypes))
	for _, queryType := range matched.QueryTypes {
		allowedQueryTypes[strings.TrimSpace(queryType)] = struct{}{}
	}
	for _, query := range dr.Queries {
		if _, ok := allowedQueryTypes[strings.TrimSpace(query.Type)]; !ok {
			return fmt.Errorf("data_request query type %q not allowed for interface %q", query.Type, matched.ID)
		}
	}
	dr.ExternalInterface = matched.ExternalInterface
	return nil
}

func normalizeDynamicActionCall(req *InvokeRequest, call *ActionCall, actionRegistered bool) (*DynamicInterface, error) {
	if call == nil {
		return nil, nil
	}
	items := dynamicInterfacesByKind(req, DynamicInterfaceAction)
	if len(items) == 0 {
		return nil, nil
	}
	ref := ""
	if call.Args != nil {
		if raw, ok := call.Args["external_interface"].(string); ok && strings.TrimSpace(raw) != "" {
			ref = strings.TrimSpace(raw)
		}
	}
	if ref == "" {
		actionID := strings.TrimSpace(call.ActionID)
		for _, item := range items {
			if actionID == item.ID || actionID == item.ExternalInterface {
				ref = actionID
				break
			}
		}
	}
	if ref == "" {
		if actionRegistered {
			return nil, nil
		}
		return nil, fmt.Errorf("action %q not allowed for this request", strings.TrimSpace(call.ActionID))
	}
	matched, err := resolveDynamicInterface(req, DynamicInterfaceAction, ref)
	if err != nil {
		return nil, err
	}
	if matched == nil {
		return nil, nil
	}
	if call.Args == nil {
		call.Args = map[string]any{}
	}
	call.Args["external_interface"] = matched.ExternalInterface
	return matched, nil
}

func runtimeTaskConsumerFromArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	if consumer, ok := args["consumer"].(string); ok && strings.TrimSpace(consumer) != "" {
		return strings.TrimSpace(consumer)
	}
	return ""
}

func runtimeTaskDeliveryModeFromArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	for _, key := range []string{"delivery_mode", "mode"} {
		if mode, ok := args[key].(string); ok && strings.TrimSpace(mode) != "" {
			return strings.TrimSpace(mode)
		}
	}
	return ""
}

func runtimeTaskTransportFromArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	for _, key := range []string{"primary_transport", "integration", "transport"} {
		if value, ok := args[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func runtimeTaskTimeoutFromArgs(args map[string]any) int {
	if args == nil {
		return 0
	}
	if raw, ok := args["timeout_ms"]; ok {
		switch value := raw.(type) {
		case int:
			return value
		case float64:
			return int(value)
		}
	}
	return 0
}

func runtimeTaskMaxAttemptsFromArgs(args map[string]any) int {
	if args == nil {
		return 0
	}
	if raw, ok := args["max_attempts"]; ok {
		switch value := raw.(type) {
		case int:
			return value
		case float64:
			return int(value)
		}
	}
	return 0
}

func runtimeTaskCallbackPostProcessFromArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	if value, ok := args["callback_post_process"].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}

func runtimeTaskCallbackMemoryLevelFromArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	if value, ok := args["callback_memory_level"].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}

func runtimeTaskCallbackMemoryTemplateFromArgs(args map[string]any) string {
	if args == nil {
		return ""
	}
	if value, ok := args["callback_memory_template"].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}

func sanitizeExternalActionArgs(args map[string]any) map[string]any {
	if len(args) == 0 {
		return map[string]any{}
	}
	result := make(map[string]any, len(args))
	for k, v := range args {
		switch k {
		case "delivery_mode", "mode", "primary_transport", "integration", "transport", "timeout_ms", "consumer", "max_attempts", "callback_post_process", "callback_memory_level", "callback_memory_template":
			continue
		default:
			result[k] = v
		}
	}
	return result
}

func dispatchAttemptsFromResult(result *external.DispatchResult) int {
	if result == nil || result.Metadata == nil {
		return 1
	}
	if raw, ok := result.Metadata["dispatch_attempt"]; ok {
		switch value := raw.(type) {
		case int:
			if value > 0 {
				return value
			}
		case float64:
			if value > 0 {
				return int(value)
			}
		}
	}
	return 1
}

func dispatchStatusCodeFromResult(result *external.DispatchResult) int {
	if result == nil {
		return 0
	}
	return result.Status
}

func runtimeTaskInterfaceNameForAction(actionID string) string {
	if strings.TrimSpace(actionID) == "" {
		return "async_action"
	}
	return actionID
}

func resolveAsyncActionMaxAttempts(actionID string, args map[string]any) int {
	maxAttempts := runtimeTaskMaxAttemptsFromArgs(args)
	if maxAttempts > 0 {
		return maxAttempts
	}
	interfaceCfg, ok := externalInterfaceConfig(asyncActionInterfaceName(actionID, args))
	if ok && interfaceCfg.MaxAttempts > 0 {
		return interfaceCfg.MaxAttempts
	}
	return 0
}

func heartbeatTimeoutPolicySnapshot(cfg config.ExternalInterfaceConfig) map[string]any {
	autoRequeue := false
	if cfg.HeartbeatTimeoutAutoRequeue != nil {
		autoRequeue = *cfg.HeartbeatTimeoutAutoRequeue
	}
	return map[string]any{
		"auto_requeue":     autoRequeue,
		"requeue_delay_ms": cfg.HeartbeatTimeoutRequeueDelayMs,
		"reason":           strings.TrimSpace(cfg.HeartbeatTimeoutReason),
	}
}

func buildAsyncActionRuntimeTaskPayload(req *InvokeRequest, actionID string, args map[string]any, callbackID string) string {
	route := resolveAsyncActionRoute(actionID, args)
	interfaceCfg, _ := externalInterfaceConfig(asyncActionInterfaceName(actionID, args))
	callbackPostProcess := firstNonEmpty(runtimeTaskCallbackPostProcessFromArgs(args), interfaceCfg.CallbackPostProcess)
	callbackMemoryLevel := firstNonEmpty(runtimeTaskCallbackMemoryLevelFromArgs(args), interfaceCfg.CallbackMemoryLevel)
	callbackMemoryTemplate := firstNonEmpty(runtimeTaskCallbackMemoryTemplateFromArgs(args), interfaceCfg.CallbackMemoryTemplate)
	maxAttempts := resolveAsyncActionMaxAttempts(actionID, args)
	payload := map[string]any{
		"task_type":                req.TaskType,
		"world_id":                 req.WorldID,
		"node_id":                  req.NodeID,
		"callback_id":              callbackID,
		"resume_policy":            firstNonEmpty(route.ResumePolicy, "none"),
		"external_interface":       asyncActionInterfaceName(actionID, args),
		"external_interaction":     "external_action",
		"action_id":                actionID,
		"delivery_mode":            route.DeliveryMode,
		"primary_transport":        route.PrimaryTransport,
		"fallback_transport":       route.FallbackTransport,
		"consumer":                 route.Consumer,
		"max_attempts":             maxAttempts,
		"heartbeat_timeout_policy": heartbeatTimeoutPolicySnapshot(interfaceCfg),
		"callback_post_process": map[string]any{
			"mode":            callbackPostProcess,
			"memory_level":    callbackMemoryLevel,
			"memory_template": callbackMemoryTemplate,
		},
		"args": sanitizeExternalActionArgs(args),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func enqueueAsyncActionRuntimeTask(req *InvokeRequest, actionID string, args map[string]any, callbackID string, route external.Route) (*store.RuntimeTaskModel, error) {
	maxAttempts := resolveAsyncActionMaxAttempts(actionID, args)
	item := &store.RuntimeTaskModel{
		Category:      "external_action",
		InterfaceName: asyncActionInterfaceName(actionID, args),
		DeliveryMode:  route.DeliveryMode,
		Consumer:      route.Consumer,
		Transport:     route.PrimaryTransport,
		WorldUUID:     req.WorldID,
		NodeUUID:      req.NodeID,
		CallbackID:    callbackID,
		MaxAttempts:   maxAttempts,
		Status:        store.RuntimeTaskStatusPending,
		Priority:      80,
		PayloadJSON:   buildAsyncActionRuntimeTaskPayload(req, actionID, args, callbackID),
	}
	if err := store.CreateRuntimeTask(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (p *Pipeline) dispatchAsyncActionRuntimeTask(task *store.RuntimeTaskModel, req *InvokeRequest, actionID string, args map[string]any, route external.Route) error {
	if task == nil || !route.ShouldPush() {
		return nil
	}
	idempotencyKey := task.TaskID
	dispatchReq := external.DispatchRequest{
		TaskID:           task.TaskID,
		IdempotencyKey:   idempotencyKey,
		Category:         task.Category,
		InterfaceName:    task.InterfaceName,
		DeliveryMode:     route.DeliveryMode,
		PrimaryTransport: route.PrimaryTransport,
		Consumer:         route.Consumer,
		WorldID:          req.WorldID,
		NodeID:           req.NodeID,
		CallbackID:       task.CallbackID,
		ResumePolicy:     "none",
		Payload: map[string]any{
			"action_id": actionID,
			"args":      sanitizeExternalActionArgs(args),
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
		if route.IsStrictPush() {
			_ = store.CompleteAsyncCallbackRecord(task.CallbackID, "failed", "", err.Error())
			_ = store.UpdateRuntimeTaskTerminalCallbackFailure(task.CallbackID, err.Error())
			return err
		}
		return nil
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

type llmParsedOutput struct {
	Reply                   string
	CleanedContent          string
	ParseError              string
	AdvancedTicks           int
	RawActionCalls          string
	RawMemoryUpdates        string
	RawPlan                 string
	RawRequestData          string
	RawInterimMemoryUpdates string
	RawFutureOutline        string
	RawSubTasks             string
}

// parseLLMJSON 解析大模型返回的 JSON 字符串。
func (p *Pipeline) parseLLMJSON(content string) *llmParsedOutput {
	out := &llmParsedOutput{Reply: content}

	cleaned := p.cleanJSON(content)
	out.CleanedContent = cleaned

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
		out.ParseError = err.Error()
		log.Printf("[warn] LLM output is not valid JSON, using raw: %v", err)
		return out
	}

	if reply, ok := raw["reply"]; ok {
		var s string
		json.Unmarshal(reply, &s)
		out.Reply = s
	}

	if advancedTicks, ok := raw["advanced_ticks"]; ok {
		var count int
		if err := json.Unmarshal(advancedTicks, &count); err == nil && count > 0 {
			out.AdvancedTicks = count
		}
	}

	if ac, ok := raw["action_calls"]; ok {
		out.RawActionCalls = string(ac)
	}

	if mu, ok := raw["memory_updates"]; ok {
		out.RawMemoryUpdates = string(mu)
	}

	if imu, ok := raw["interim_memory_updates"]; ok {
		out.RawInterimMemoryUpdates = string(imu)
	}
	if rd, ok := raw["request_data"]; ok {
		out.RawRequestData = string(rd)
	}
	if wcp, ok := raw["world_change_plan"]; ok {
		out.RawPlan = string(wcp)
	}
	if st, ok := raw["sub_tasks"]; ok {
		out.RawSubTasks = string(st)
	}

	if fo, ok := raw["future_outline"]; ok {
		var s string
		json.Unmarshal(fo, &s)
		out.RawFutureOutline = s
	}

	return out
}

// cleanJSON 清理模型输出中的 Markdown 代码块包装。
func (p *Pipeline) cleanJSON(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		lines := strings.SplitN(content, "\n", 2)
		if len(lines) == 2 {
			content = strings.TrimSpace(lines[1])
		}
	}
	if strings.HasSuffix(content, "```") {
		idx := strings.LastIndex(content, "```")
		if idx >= 0 {
			content = strings.TrimSpace(content[:idx])
		}
	}
	return content
}

func (p *Pipeline) parseActionCalls(rawJSON string, defaultNodeID string) []ActionCall {
	if rawJSON == "" || rawJSON == "null" || rawJSON == "[]" {
		return nil
	}

	var rawCalls []struct {
		ActionID string         `json:"action_id"`
		Args     map[string]any `json:"args"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &rawCalls); err != nil {
		log.Printf("[warn] parse action_calls: %v", err)
		return nil
	}

	var result []ActionCall
	for _, rc := range rawCalls {
		if rc.ActionID == "" {
			continue
		}
		ac := ActionCall{
			ActionID: rc.ActionID,
			Args:     rc.Args,
		}
		if ac.Args == nil {
			ac.Args = make(map[string]any)
		}
		if _, ok := ac.Args["node_id"]; !ok {
			ac.Args["node_id"] = defaultNodeID
		}
		result = append(result, ac)
	}
	return result
}

func (p *Pipeline) executeActions(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, policyEngine *planner.PolicyEngine, calls []ActionCall) []ActionCall {
	var result []ActionCall
	dynamicActionCounts := map[string]int{}

	for _, call := range calls {
		actionID := call.ActionID
		args := call.Args
		actionRegistered := p.actionReg.Exists(actionID)
		matchedDynamicInterface, dynamicErr := normalizeDynamicActionCall(req, &call, actionRegistered)
		if dynamicErr != nil {
			p.emitExecutionEvent(req, runtime, executionMode, "action_blocked", actionID, map[string]any{"action_id": actionID, "args": args, "reason": dynamicErr.Error()})
			log.Printf("[dynamic-interface] action %s blocked: %v", actionID, dynamicErr)
			continue
		}
		if matchedDynamicInterface != nil && matchedDynamicInterface.MaxCalls > 0 {
			dynamicActionCounts[matchedDynamicInterface.ID]++
			if dynamicActionCounts[matchedDynamicInterface.ID] > matchedDynamicInterface.MaxCalls {
				reason := fmt.Sprintf("action exceeds max_calls for interface %q", matchedDynamicInterface.ID)
				p.emitExecutionEvent(req, runtime, executionMode, "action_blocked", actionID, map[string]any{"action_id": actionID, "args": args, "reason": reason})
				log.Printf("[dynamic-interface] action %s blocked: %s", actionID, reason)
				continue
			}
		}
		actionID = call.ActionID
		args = call.Args

		if policyEngine != nil && !policyEngine.IsActionAllowed(actionID) {
			p.emitExecutionEvent(req, runtime, executionMode, "action_blocked", actionID, map[string]any{"action_id": actionID, "args": args, "reason": "policy_blocked"})
			log.Printf("[policy] action %s blocked", actionID)
			continue
		}

		if p.actionReg.IsSync(actionID) {
			out, err := p.actionReg.ExecuteSync(actionID, args)
			if err != nil {
				p.emitExecutionEvent(req, runtime, executionMode, "action_sync_failed", actionID, map[string]any{"action_id": actionID, "args": args, "error": err.Error()})
				log.Printf("[action:sync] %s failed: %v", actionID, err)
			} else {
				p.emitExecutionEvent(req, runtime, executionMode, "action_sync_succeeded", actionID, map[string]any{"action_id": actionID, "args": args, "result": out})
				log.Printf("[action:sync] %s success: %v", actionID, out)
			}
		} else if p.actionReg.IsAsync(actionID) {
			cbID := p.actionReg.CreateCallbackWithMetadata(actionID, args, action.CallbackMetadata{
				NodeID:    req.NodeID,
				WorldID:   req.WorldID,
				RequestID: "",
			})
			route := resolveAsyncActionRoute(actionID, args)
			task, err := enqueueAsyncActionRuntimeTask(req, actionID, args, cbID, route)
			if err != nil {
				p.emitExecutionEvent(req, runtime, executionMode, "action_async_enqueue_failed", actionID, map[string]any{"action_id": actionID, "args": args, "callback_id": cbID, "error": err.Error()})
				log.Printf("[action:async] %s enqueue failed callback=%s err=%v", actionID, cbID, err)
				continue
			}
			if err := p.dispatchAsyncActionRuntimeTask(task, req, actionID, args, route); err != nil {
				p.emitExecutionEvent(req, runtime, executionMode, "action_async_dispatch_failed", actionID, map[string]any{"action_id": actionID, "args": args, "callback_id": cbID, "error": err.Error(), "delivery_mode": route.DeliveryMode, "primary_transport": route.PrimaryTransport})
				log.Printf("[action:async] %s dispatch failed callback=%s err=%v", actionID, cbID, err)
				continue
			}
			call.Mode = "async"
			call.CallbackID = cbID
			result = append(result, call)
			p.emitExecutionEvent(req, runtime, executionMode, "action_async_enqueued", actionID, map[string]any{"action_id": actionID, "args": args, "callback_id": cbID, "delivery_mode": route.DeliveryMode, "primary_transport": route.PrimaryTransport})
			log.Printf("[action:async] %s callback=%s", actionID, cbID)
		} else if matchedDynamicInterface != nil {
			cbID := p.actionReg.CreateCallbackWithMetadata(actionID, args, action.CallbackMetadata{
				NodeID:    req.NodeID,
				WorldID:   req.WorldID,
				RequestID: "",
			})
			route := resolveAsyncActionRoute(actionID, args)
			task, err := enqueueAsyncActionRuntimeTask(req, actionID, args, cbID, route)
			if err != nil {
				p.emitExecutionEvent(req, runtime, executionMode, "action_async_enqueue_failed", actionID, map[string]any{"action_id": actionID, "args": args, "callback_id": cbID, "error": err.Error()})
				log.Printf("[action:dynamic] %s enqueue failed callback=%s err=%v", actionID, cbID, err)
				continue
			}
			if err := p.dispatchAsyncActionRuntimeTask(task, req, actionID, args, route); err != nil {
				p.emitExecutionEvent(req, runtime, executionMode, "action_async_dispatch_failed", actionID, map[string]any{"action_id": actionID, "args": args, "callback_id": cbID, "error": err.Error(), "delivery_mode": route.DeliveryMode, "primary_transport": route.PrimaryTransport})
				log.Printf("[action:dynamic] %s dispatch failed callback=%s err=%v", actionID, cbID, err)
				continue
			}
			call.Mode = "async"
			call.CallbackID = cbID
			result = append(result, call)
			p.emitExecutionEvent(req, runtime, executionMode, "action_async_enqueued", actionID, map[string]any{"action_id": actionID, "args": args, "callback_id": cbID, "delivery_mode": route.DeliveryMode, "primary_transport": route.PrimaryTransport, "dynamic_interface_id": matchedDynamicInterface.ID, "external_interface": matchedDynamicInterface.ExternalInterface})
			log.Printf("[action:dynamic] %s callback=%s interface=%s", actionID, cbID, matchedDynamicInterface.ExternalInterface)
		} else {
			call.Mode = "async"
			result = append(result, call)
			p.emitExecutionEvent(req, runtime, executionMode, "action_unknown_passthrough", actionID, map[string]any{"action_id": actionID, "args": args})
			log.Printf("[action:unknown] %s passed through", actionID)
		}
	}

	return result
}

func (p *Pipeline) parseMemoryUpdates(rawJSON string) []MemoryUpdate {
	if rawJSON == "" || rawJSON == "null" || rawJSON == "[]" {
		return nil
	}

	levelMap := map[string]MemoryLevel{
		"short_term": MemShortTerm,
		"long_term":  MemLongTerm,
		"shared":     MemShared,
		"world":      MemWorld,
	}

	var rawMems []struct {
		NodeID      string           `json:"node_id"`
		Content     string           `json:"content"`
		Level       string           `json:"level"`
		Tags        string           `json:"tags,omitempty"`
		Propagation *PropagationRule `json:"propagation,omitempty"`
	}
	if err := json.Unmarshal([]byte(rawJSON), &rawMems); err != nil {
		log.Printf("[warn] parse memory_updates: %v", err)
		return nil
	}

	var result []MemoryUpdate
	for _, rm := range rawMems {
		if rm.Content == "" {
			continue
		}
		level := levelMap[rm.Level]
		if level == "" {
			level = MemShortTerm
		}
		if rm.Propagation != nil {
			if rm.Propagation.Mode == "" {
				rm.Propagation.Mode = PropModeUpward
			} else if !IsValidPropagationMode(rm.Propagation.Mode) {
				log.Printf("[warn] parse memory_updates: unsupported propagation mode %q", rm.Propagation.Mode)
				rm.Propagation = nil
			}
		}
		result = append(result, MemoryUpdate{
			NodeID:      rm.NodeID,
			Content:     rm.Content,
			Level:       level,
			Tags:        rm.Tags,
			Propagation: rm.Propagation,
		})
	}
	return result
}

func (p *Pipeline) writeMemories(req *InvokeRequest, runtime *executionConfig, executionMode ExecutionMode, updates []MemoryUpdate) {
	batch := make([]memoryWriteRequest, 0, len(updates))
	validUpdates := make([]MemoryUpdate, 0, len(updates))
	for _, mu := range updates {
		nodeID := store.ResolveNodeUUID(mu.NodeID)
		if nodeID == 0 {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_write_skipped", mu.Content, map[string]any{"node_id": mu.NodeID, "reason": "unknown_node"})
			log.Printf("[warn] write memory: unknown node UUID %s", mu.NodeID)
			continue
		}
		batch = append(batch, memoryWriteRequest{
			NodeUUID: mu.NodeID,
			NodeID:   nodeID,
			Content:  mu.Content,
			Level:    mu.Level,
			Tags:     mu.Tags,
		})
		validUpdates = append(validUpdates, mu)
	}
	if err := persistMemoryBatch(batch); err != nil {
		for _, mu := range validUpdates {
			p.emitExecutionEvent(req, runtime, executionMode, "memory_write_failed", mu.Content, map[string]any{"node_id": mu.NodeID, "level": mu.Level, "tags": mu.Tags, "error": err.Error()})
		}
		logMemoryBatchFailure("write memory batch", err)
		return
	}
	for _, mu := range validUpdates {
		p.emitExecutionEvent(req, runtime, executionMode, "memory_written", mu.Content, map[string]any{"node_id": mu.NodeID, "level": mu.Level, "tags": mu.Tags})
	}
}

func sanitizeRoles(messages []ChatMessage) []ChatMessage {
	result := make([]ChatMessage, len(messages))
	for i, m := range messages {
		role := m.Role
		switch role {
		case "player", "narrator", "observer":
			role = "user"
		case "npc", "agent":
			role = "assistant"
		default:
			role = m.Role
		}
		result[i] = ChatMessage{Role: role, Content: m.Content}
	}
	return result
}

// handleDataRequest 解析模型发出的数据查询请求并把结果压平成补充文本。
//
// 约束：
//  1. 这里是按需扩图接口，不是默认上下文组装器；它应返回任务明确请求的数据，而不是替模型做无边界全量展开。
//  2. DataQuery.Filter 的含义必须与 types.go 中的注释一致：node_components=component_type,
//     node_relations=relation_type, node_memories=memory_level。
//  3. node_relations 返回的是关系视图，不应替代任务级关系图谱策略；尤其不能因为这里能查全量关系，就默认在主 prompt
//     中拼接全量关系。
func (p *Pipeline) handleDataRequest(policyEngine *planner.PolicyEngine, dr *DataRequest) string {
	var parts []string
	for _, q := range dr.Queries {
		switch q.Type {
		case "node_components":
			if comps, err := store.GetNodeComponents(q.NodeID); err == nil {
				for _, c := range comps {
					if q.Filter == "" || c.ComponentType == q.Filter {
						parts = append(parts, fmt.Sprintf("  【%s】%s", c.ComponentType, c.Data))
					}
				}
			}
		case "node_memories":
			if mems, err := store.GetNodeMemories(q.NodeID, limitOrDefault(q.Limit, 50)); err == nil {
				for _, m := range mems {
					if q.Filter != "" && m.Level != q.Filter {
						continue
					}
					parts = append(parts, fmt.Sprintf("  [记忆:%s] %s", m.Level, m.Content))
				}
			}
		case "node_relations":
			if rels, err := store.GetNodeRelations(q.NodeID); err == nil {
				for _, r := range rels {
					if q.Filter != "" && r.RelationType != q.Filter {
						continue
					}
					// SourceID/TargetID are int64 now; use UUID fields for display
					parts = append(parts, fmt.Sprintf("  [关系:%s] %s -> %s (权重:%d)", r.RelationType, r.SourceUUID, r.TargetUUID, r.Weight))
				}
			}
		case "memory_search":
			if mems, err := store.SearchMemories(q.NodeID, q.Filter, limitOrDefault(q.Limit, 10)); err == nil {
				for _, m := range mems {
					parts = append(parts, fmt.Sprintf("  [搜索记忆:%s] %s", m.Level, m.Content))
				}
			}
		case "policy_check":
			allowed := true
			if policyEngine != nil {
				allowed = policyEngine.IsActionAllowed(q.Filter)
			}
			status := "blocked"
			if allowed {
				status = "allowed"
			}
			parts = append(parts, fmt.Sprintf("  [策略检查] %s -> %s", q.Filter, status))
		case "node_detail":
			if nd, err := store.GetNode(q.NodeID); err == nil {
				parentInfo := "无"
				if nd.ParentUUID != nil {
					parentInfo = *nd.ParentUUID
				}
				parts = append(parts, fmt.Sprintf("  [节点] id=%s name=%s type=%s parent=%s world=%s 创建=%v", nd.UUID, nd.Name, nd.NodeType, parentInfo, nd.WorldUUID, nd.CreatedAt.Format("01-02 15:04")))
			}
		case "node_type_list":
			if nodes, err := store.GetAllNodes(q.NodeID, limitOrDefault(q.Limit, 20), 0, q.Filter); err == nil {
				for _, n := range nodes {
					parts = append(parts, fmt.Sprintf("  %s (%s) id=%s", n.Name, n.NodeType, n.UUID))
				}
			}
		case "world_timeline":
			if ticks, err := store.GetTimelineTicks(q.NodeID, limitOrDefault(q.Limit, 10)); err == nil {
				for _, t := range ticks {
					parts = append(parts, fmt.Sprintf("  [tick %d] %s type=%s", t.TickNumber, t.GameTime, t.TickType))
				}
			}
		}
	}
	return strings.Join(parts, "\n")
}

func limitOrDefault(v, def int) int {
	if v > 0 {
		return v
	}
	return def
}
