package service

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"gorm.io/gorm"
)

const defaultAutonomousTickLimit = 10

func emitWorldServiceLog(worldID, nodeID string, taskType engine.TaskType, eventName, message string, detail any) {
	mode := config.ExecutionMode()
	if mode == "" || mode == "full" {
		mode = string(engine.ModeProduction)
	}
	if err := store.CreateInferenceLog(&store.InferenceLogModel{
		WorldUUID:              worldID,
		TaskType:               string(taskType),
		NodeUUID:               nodeID,
		Category:               "world_service",
		EventName:              eventName,
		LogLevel:               "info",
		Message:                message,
		ExecutionMode:          mode,
		DetailData:             marshalWorldServiceDetail(detail, mode),
		ConfiguredPipelineMode: "",
		EffectivePipelineMode:  "",
	}); err != nil {
		log.Printf("[world-service-log] %s: %v", eventName, err)
	}
}

func marshalWorldServiceDetail(detail any, mode string) string {
	if detail == nil {
		return ""
	}
	if mode == string(engine.ModeProduction) {
		return ""
	}
	data, err := json.Marshal(detail)
	if err != nil {
		return ""
	}
	return string(data)
}

// AdvanceWorldTickWithAutonomous 推进世界刻，并按请求级限制触发 world_tick_sync 自主节点。
func AdvanceWorldTickWithAutonomous(p *engine.Pipeline, worldID, tickType, gameTime string, autonomousLimit *int) (*store.TimelineModel, *engine.InvokeResponse, []engine.AutonomousRunResult, error) {
	if tickType == "" {
		tickType = "scheduled"
	}
	emitWorldServiceLog(worldID, worldID, engine.TaskWorldTick, "world_tick_requested", tickType, map[string]any{"game_time": gameTime, "autonomous_limit": autonomousLimit})
	if _, err := ensureWorldNodeTx(store.DB, worldID); err != nil {
		return nil, nil, nil, err
	}

	resp, err := p.Execute(&engine.InvokeRequest{
		WorldID:  worldID,
		TaskType: engine.TaskWorldTick,
		NodeID:   worldID,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	emitWorldServiceLog(worldID, worldID, engine.TaskWorldTick, "world_tick_completed", worldPlanSummary(resp), resp)

	var tick *store.TimelineModel
	err = store.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		tick, err = persistWorldTickArtifactsTx(tx, worldID, tickType, gameTime, resp)
		return err
	})
	if err != nil {
		return nil, nil, nil, err
	}
	emitWorldServiceLog(worldID, worldID, engine.TaskWorldTick, "world_tick_persisted", tick.Summary, tick)

	autonomousRuns := RunWorldTickAutonomous(p, worldID, autonomousLimit)
	emitWorldServiceLog(worldID, worldID, engine.TaskWorldTick, "world_tick_autonomous_completed", "autonomous runs completed", autonomousRuns)
	return tick, resp, autonomousRuns, nil
}

func persistWorldTickArtifactsTx(tx *gorm.DB, worldID, tickType, gameTime string, resp *engine.InvokeResponse) (*store.TimelineModel, error) {
	latest, err := getLatestTickTx(tx, worldID)
	tickNum := 1
	if err == nil {
		tickNum = latest.TickNumber + 1
	} else if !IsKind(err, ErrorNotFound) {
		return nil, err
	}

	worldInt := txResolveWorldUUID(tx, worldID)
	timelinePayload, err := buildWorldTickTimelineData(resp)
	if err != nil {
		return nil, err
	}
	tick := &store.TimelineModel{
		UUID:          store.NewUUID(),
		WorldID:       worldInt,
		WorldUUID:     worldID,
		TickNumber:    tickNum,
		TickType:      tickType,
		GameTime:      gameTime,
		Summary:       worldPlanSummary(resp),
		Data:          timelinePayload,
		FutureOutline: resp.FutureOutline,
	}
	if err := tx.Create(tick).Error; err != nil {
		return nil, err
	}

	if err := persistWorldTickStateComponentsTx(tx, worldID, tick, resp); err != nil {
		return nil, err
	}
	return tick, nil
}

func buildWorldTickTimelineData(resp *engine.InvokeResponse) (string, error) {
	payload := map[string]any{
		"reply":             resp.Reply,
		"world_change_plan": resp.WorldChangePlan,
		"future_outline":    resp.FutureOutline,
		"memory_updates":    resp.MemoryUpdates,
		"action_calls":      resp.ActionCalls,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func persistWorldTickStateComponentsTx(tx *gorm.DB, worldID string, tick *store.TimelineModel, resp *engine.InvokeResponse) error {
	recentFacts := collectWorldTickFacts(resp)
	canonicalFacts := collectCanonicalWorldFacts(resp)
	if _, err := upsertStateComponentTx(tx, worldID, engine.CompWorldState, engine.WorldStateComponent{
		Summary:        worldPlanSummary(resp),
		KeyFacts:       recentFacts,
		CanonicalFacts: canonicalFacts,
		ActiveArcs:     collectPlanEventDescriptions(resp),
		Metadata: map[string]any{
			"tick_number":    tick.TickNumber,
			"tick_type":      tick.TickType,
			"game_time":      tick.GameTime,
			"future_outline": resp.FutureOutline,
		},
	}); err != nil {
		return err
	}
	if _, err := upsertStateComponentTx(tx, worldID, engine.CompStoryState, engine.StoryStateComponent{
		CurrentSituation: truncateWorldTickText(resp.Reply, 1200),
		RecentChanges:    append([]string{}, canonicalFacts...),
		PendingThreads:   collectPendingThreads(resp),
		Metadata: map[string]any{
			"tick_number": tick.TickNumber,
			"tick_type":   tick.TickType,
		},
	}); err != nil {
		return err
	}
	historyComp, historyErr := getStateComponentTx(tx, worldID, engine.CompStoryHistory)
	if historyErr != nil {
		return historyErr
	}
	history := engine.StoryHistoryComponent{}
	if historyComp != nil && strings.TrimSpace(historyComp.Data) != "" {
		_ = json.Unmarshal([]byte(historyComp.Data), &history)
	}
	history.Entries = append([]engine.StoryHistoryEntry{{
		TickNumber: tick.TickNumber,
		Summary:    worldPlanSummary(resp),
		Facts:      append([]string{}, canonicalFacts...),
		GameTime:   tick.GameTime,
	}}, history.Entries...)
	if len(history.Entries) > 12 {
		history.Entries = history.Entries[:12]
	}
	if _, err := upsertStateComponentTx(tx, worldID, engine.CompStoryHistory, history); err != nil {
		return err
	}
	if _, err := upsertStateComponentTx(tx, worldID, engine.CompStateSnapshot, engine.StateSnapshotComponent{
		SnapshotType: "world_tick",
		Version:      "v1",
		Payload: map[string]any{
			"tick_number":       tick.TickNumber,
			"summary":           tick.Summary,
			"future_outline":    resp.FutureOutline,
			"recent_facts":      recentFacts,
			"world_change_plan": resp.WorldChangePlan,
		},
	}); err != nil {
		return err
	}
	return nil
}

func collectWorldTickFacts(resp *engine.InvokeResponse) []string {
	set := map[string]bool{}
	var facts []string
	add := func(value string) {
		value = truncateWorldTickText(value, 240)
		value = strings.TrimSpace(value)
		if value == "" || set[value] {
			return
		}
		set[value] = true
		facts = append(facts, value)
	}
	for _, mem := range resp.MemoryUpdates {
		add(mem.Content)
	}
	if resp.WorldChangePlan != nil {
		add(resp.WorldChangePlan.Summary)
		for _, evt := range resp.WorldChangePlan.WorldEvents {
			add(evt.Description)
		}
	}
	for _, line := range strings.Split(resp.Reply, "\n") {
		add(line)
		if len(facts) >= 8 {
			break
		}
	}
	return facts
}

func collectCanonicalWorldFacts(resp *engine.InvokeResponse) []string {
	if resp == nil {
		return nil
	}
	set := map[string]bool{}
	var facts []string
	add := func(value string) {
		for _, candidate := range expandCanonicalFactCandidates(value) {
			candidate = normalizeCanonicalFact(candidate)
			if candidate == "" || set[candidate] {
				continue
			}
			set[candidate] = true
			facts = append(facts, candidate)
			if len(facts) >= 10 {
				return
			}
		}
	}
	for _, line := range strings.Split(resp.Reply, "\n") {
		add(line)
		if len(facts) >= 10 {
			return facts[:10]
		}
	}
	for _, mem := range resp.MemoryUpdates {
		add(mem.Content)
		if len(facts) >= 10 {
			return facts[:10]
		}
	}
	if resp.WorldChangePlan != nil {
		add(resp.WorldChangePlan.Summary)
		if len(facts) >= 10 {
			return facts[:10]
		}
		for _, evt := range resp.WorldChangePlan.WorldEvents {
			add(evt.Description)
			if len(facts) >= 10 {
				return facts[:10]
			}
		}
	}
	if len(facts) > 10 {
		facts = facts[:10]
	}
	return facts
}

func normalizeCanonicalFact(value string) string {
	value = normalizeWorldTickWhitespace(value)
	if value == "" {
		return ""
	}
	value = strings.Trim(value, `"'“”‘’「」『』[]()（）`)
	value = trimCanonicalListPrefix(value)
	value = strings.TrimSpace(value)
	if value == "" || !isConcreteCanonicalFact(value) {
		return ""
	}
	return truncateWorldTickText(value, 180)
}

func expandCanonicalFactCandidates(value string) []string {
	segments := splitCanonicalFactParts(value, func(r rune) bool {
		switch r {
		case '\n', '\r', '。', '！', '？', '!', '?', '；', ';':
			return true
		default:
			return false
		}
	})
	if len(segments) == 0 {
		return nil
	}
	set := map[string]bool{}
	var candidates []string
	add := func(item string) {
		item = normalizeWorldTickWhitespace(item)
		item = strings.TrimSpace(item)
		if item == "" || set[item] {
			return
		}
		set[item] = true
		candidates = append(candidates, item)
	}
	for _, segment := range segments {
		add(segment)
		if len([]rune(segment)) < 20 {
			continue
		}
		for _, clause := range splitCanonicalFactParts(segment, func(r rune) bool {
			switch r {
			case '，', ',', '、', ':', '：':
				return true
			default:
				return false
			}
		}) {
			add(clause)
		}
	}
	return candidates
}

func splitCanonicalFactParts(value string, separator func(rune) bool) []string {
	parts := strings.FieldsFunc(value, separator)
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func normalizeWorldTickWhitespace(value string) string {
	value = strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return ' '
		}
		return r
	}, value)
	return strings.Join(strings.Fields(value), " ")
}

func trimCanonicalListPrefix(value string) string {
	value = strings.TrimLeftFunc(value, unicode.IsSpace)
	for _, prefix := range []string{"- ", "* ", "• ", "1. ", "1) ", "1、", "2. ", "2) ", "2、", "3. ", "3) ", "3、"} {
		if strings.HasPrefix(value, prefix) {
			return strings.TrimSpace(value[len(prefix):])
		}
	}
	return value
}

func isConcreteCanonicalFact(value string) bool {
	runeCount := len([]rune(value))
	if runeCount < 6 {
		return false
	}
	score := 0
	if runeCount >= 10 {
		score++
	}
	if containsDigit(value) {
		score += 2
	}
	if containsMeasurement(value) {
		score += 2
	}
	if containsStructuredIdentifier(value) {
		score += 2
	}
	if containsSpecificEntitySuffix(value) {
		score += 2
	}
	if strings.ContainsAny(value, "“”‘’\"'()（）[]") {
		score++
	}
	if looksGenericCanonicalFact(value) {
		score -= 2
	}
	return score >= 2
}

func containsDigit(value string) bool {
	for _, r := range value {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func containsMeasurement(value string) bool {
	if !containsDigit(value) {
		return false
	}
	units := []string{"米", "公里", "公尺", "层", "级", "号", "年", "月", "日", "小时", "分钟", "秒", "%", "％", "吨", "人", "座", "次", "度", "m", "km"}
	for _, unit := range units {
		if strings.Contains(value, unit) {
			return true
		}
	}
	return false
}

func containsStructuredIdentifier(value string) bool {
	hasLetter := false
	hasDigit := false
	hasSymbol := false
	for _, r := range value {
		switch {
		case unicode.IsLetter(r) && r <= unicode.MaxASCII:
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		case r == '-' || r == '_' || r == '/':
			hasSymbol = true
		}
	}
	return (hasLetter && hasDigit) || (hasLetter && hasSymbol)
}

func containsSpecificEntitySuffix(value string) bool {
	suffixes := []string{"谐振腔", "实验室", "观测井", "检修站", "中继站", "反应堆", "发射井", "轨道站", "地下城", "基地", "要塞", "站", "塔", "港", "城", "区", "层", "室", "井", "门", "桥", "线", "轨道", "走廊", "矿场", "舰", "号", "团", "军", "会", "所", "院"}
	for _, token := range strings.FieldsFunc(value, func(r rune) bool {
		switch r {
		case ' ', '，', ',', '。', '；', ';', '：', ':', '、', '(', ')', '（', '）', '[', ']':
			return true
		default:
			return false
		}
	}) {
		token = strings.TrimSpace(token)
		for _, suffix := range suffixes {
			if strings.HasSuffix(token, suffix) && len([]rune(token)) >= len([]rune(suffix))+1 {
				return true
			}
		}
	}
	return false
}

func looksGenericCanonicalFact(value string) bool {
	genericPhrases := []string{"局势", "情况", "事件", "计划", "推进", "变化", "发展", "影响", "问题", "消息", "线索", "风险", "危机", "秘密", "行动", "设施", "装置"}
	concreteHints := []string{"地下", "量子", "谐振腔", "反应堆", "实验室", "观测井", "检修站", "轨道", "Dar-shade", "He-3"}
	for _, hint := range concreteHints {
		if strings.Contains(value, hint) {
			return false
		}
	}
	for _, phrase := range genericPhrases {
		if strings.Contains(value, phrase) {
			return true
		}
	}
	return false
}

func collectPlanEventDescriptions(resp *engine.InvokeResponse) []string {
	if resp == nil || resp.WorldChangePlan == nil {
		return nil
	}
	result := make([]string, 0, len(resp.WorldChangePlan.WorldEvents))
	for _, evt := range resp.WorldChangePlan.WorldEvents {
		if strings.TrimSpace(evt.Description) != "" {
			result = append(result, evt.Description)
		}
	}
	return result
}

func collectPendingThreads(resp *engine.InvokeResponse) []string {
	if strings.TrimSpace(resp.FutureOutline) == "" {
		return nil
	}
	parts := strings.Split(resp.FutureOutline, "\n")
	threads := make([]string, 0, len(parts))
	for _, part := range parts {
		part = truncateWorldTickText(part, 200)
		part = strings.TrimSpace(part)
		if part != "" {
			threads = append(threads, part)
		}
		if len(threads) >= 6 {
			break
		}
	}
	return threads
}

func truncateWorldTickText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

// RunAutonomousNode 手动触发一个节点的自主行为周期。
func RunAutonomousNode(p *engine.Pipeline, worldID, nodeID string) (*engine.InvokeResponse, error) {
	if _, err := ensureWorldNodeTx(store.DB, worldID); err != nil {
		return nil, err
	}
	if err := ensureNodesInWorldTx(store.DB, worldID, nodeID); err != nil {
		return nil, err
	}
	return p.Execute(&engine.InvokeRequest{WorldID: worldID, TaskType: engine.TaskAutonomousAct, NodeID: nodeID})
}

// RunWorldTickAutonomous 触发同世界中声明为 world_tick_sync 的自主节点。
func RunWorldTickAutonomous(p *engine.Pipeline, worldID string, limit *int) []engine.AutonomousRunResult {
	return runAutonomousByTrigger(p, worldID, engine.AutonomousTriggerWorldTickSync, limit)
}

// RunScheduledAutonomous 触发同世界中到期的 scheduled 自主节点。
func RunScheduledAutonomous(p *engine.Pipeline, worldID string, limit *int, now time.Time) []engine.AutonomousRunResult {
	return runAutonomousByTrigger(p, worldID, engine.AutonomousTriggerScheduled, limit, now)
}

func runAutonomousByTrigger(p *engine.Pipeline, worldID string, trigger string, limit *int, nowOpt ...time.Time) []engine.AutonomousRunResult {
	maxRuns := defaultAutonomousTickLimit
	if limit != nil {
		maxRuns = *limit
	}
	if maxRuns <= 0 {
		return nil
	}

	// 通过世界 UUID 解析 int64 ID 后查询组件
	worldInt := store.ResolveWorldUUID(worldID)
	components, err := store.GetComponentsByTypeForWorld(worldID, string(engine.CompAutonomous))
	if err != nil {
		log.Printf("load autonomous components: %v", err)
		emitWorldServiceLog(worldID, worldID, engine.TaskAutonomousAct, "autonomous_load_failed", trigger, map[string]any{"error": err.Error()})
		return []engine.AutonomousRunResult{{Error: err.Error()}}
	}
	emitWorldServiceLog(worldID, worldID, engine.TaskAutonomousAct, "autonomous_scan_started", trigger, map[string]any{"component_count": len(components), "limit": maxRuns})
	_ = worldInt

	results := make([]engine.AutonomousRunResult, 0, maxRuns)
	for _, comp := range components {
		if len(results) >= maxRuns {
			break
		}
		cfg, err := engine.DecodeAutonomousConfig(comp.Data)
		if err != nil {
			emitWorldServiceLog(worldID, comp.NodeUUID, engine.TaskAutonomousAct, "autonomous_decode_failed", trigger, map[string]any{"error": err.Error()})
			results = append(results, engine.AutonomousRunResult{NodeID: comp.NodeUUID, Error: err.Error()})
			continue
		}
		if !cfg.Enabled || cfg.Trigger != trigger {
			continue
		}
		if trigger == engine.AutonomousTriggerScheduled && !isScheduledAutonomousDue(cfg, nowOpt) {
			continue
		}

		// 通过 NodeID 查询节点 UUID
		var nodeUUID string
		if err := store.DB.Model(&store.NodeModel{}).Select("uuid").Where("id = ?", comp.NodeID).First(&nodeUUID).Error; err != nil {
			emitWorldServiceLog(worldID, "", engine.TaskAutonomousAct, "autonomous_node_lookup_failed", trigger, map[string]any{"error": err.Error()})
			results = append(results, engine.AutonomousRunResult{NodeID: "", Error: err.Error()})
			continue
		}
		emitWorldServiceLog(worldID, nodeUUID, engine.TaskAutonomousAct, "autonomous_node_started", trigger, map[string]any{"node_id": nodeUUID})

		resp, err := RunAutonomousNode(p, worldID, nodeUUID)
		if err != nil {
			emitWorldServiceLog(worldID, nodeUUID, engine.TaskAutonomousAct, "autonomous_node_failed", trigger, map[string]any{"node_id": nodeUUID, "error": err.Error()})
			results = append(results, engine.AutonomousRunResult{NodeID: nodeUUID, Error: err.Error()})
			continue
		}
		emitWorldServiceLog(worldID, nodeUUID, engine.TaskAutonomousAct, "autonomous_node_completed", trigger, resp)
		results = append(results, engine.AutonomousRunResult{NodeID: nodeUUID, Response: resp})
	}
	return results
}

func isScheduledAutonomousDue(cfg *engine.AutonomousConfig, nowOpt []time.Time) bool {
	if cfg.IntervalSeconds <= 0 {
		return true
	}
	if cfg.LastRunAt == nil {
		return true
	}
	now := time.Now()
	if len(nowOpt) > 0 {
		now = nowOpt[0]
	}
	return now.Sub(*cfg.LastRunAt) >= time.Duration(cfg.IntervalSeconds)*time.Second
}

// UpsertAutonomousConfig creates or replaces a node's autonomous component data.
func UpsertAutonomousConfig(nodeID string, cfg *engine.AutonomousConfig) (*store.ComponentModel, error) {
	if _, err := getNodeTx(store.DB, nodeID); err != nil {
		return nil, err
	}
	if cfg.Trigger == "" {
		cfg.Trigger = engine.AutonomousTriggerManual
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, invalidf("invalid autonomous config: %v", err)
	}
	comps, err := store.GetComponentsByType(nodeID, string(engine.CompAutonomous))
	if err != nil {
		return nil, err
	}
	if len(comps) == 0 {
		return CreateComponent(nodeID, string(engine.CompAutonomous), string(data))
	}
	if err := store.UpdateComponent(comps[0].UUID, map[string]any{"data": string(data)}); err != nil {
		return nil, err
	}
	return store.GetComponent(comps[0].UUID)
}

// GetAutonomousConfig returns the node-local autonomous component config, if any.
func GetAutonomousConfig(nodeID string) (*engine.AutonomousConfig, *store.ComponentModel, error) {
	if _, err := getNodeTx(store.DB, nodeID); err != nil {
		return nil, nil, err
	}
	comps, err := store.GetComponentsByType(nodeID, string(engine.CompAutonomous))
	if err != nil {
		return nil, nil, err
	}
	if len(comps) == 0 {
		return nil, nil, nil
	}
	cfg, err := engine.DecodeAutonomousConfig(comps[0].Data)
	if err != nil {
		return nil, &comps[0], err
	}
	return cfg, &comps[0], nil
}

// EvaluateWorldEvent 校验世界和 scope 后，执行一次事件影响评估。
func EvaluateWorldEvent(p *engine.Pipeline, worldID string, event *engine.WorldEvent) (*engine.InvokeResponse, error) {
	if _, err := ensureWorldNodeTx(store.DB, worldID); err != nil {
		return nil, err
	}
	scopeID := event.ScopeID
	if scopeID == "" {
		scopeID = worldID
	} else if err := ensureNodesInWorldTx(store.DB, worldID, scopeID); err != nil {
		return nil, err
	}

	return p.Execute(&engine.InvokeRequest{
		WorldID:  worldID,
		TaskType: engine.TaskWorldEvent,
		NodeID:   scopeID,
		Event:    event,
	})
}

// ReplanWorldTimeline 清空现有 future outline 并重新生成世界大纲。
func ReplanWorldTimeline(p *engine.Pipeline, worldID string) (*engine.InvokeResponse, error) {
	if _, err := ensureWorldNodeTx(store.DB, worldID); err != nil {
		return nil, err
	}
	worldInt := store.ResolveWorldUUID(worldID)
	if err := store.DB.Model(&store.TimelineModel{}).Where("world_id = ?", worldInt).Update("future_outline", "").Error; err != nil {
		return nil, err
	}
	return p.Execute(&engine.InvokeRequest{WorldID: worldID, TaskType: engine.TaskWorldTick, NodeID: worldID})
}

// AdvanceWorldScope 推进某个范围节点的局部演化。
func AdvanceWorldScope(p *engine.Pipeline, worldID, scopeID string) (*engine.InvokeResponse, error) {
	if _, err := ensureWorldNodeTx(store.DB, worldID); err != nil {
		return nil, err
	}
	if err := ensureNodesInWorldTx(store.DB, worldID, scopeID); err != nil {
		return nil, err
	}
	return p.Execute(&engine.InvokeRequest{WorldID: worldID, TaskType: engine.TaskWorldTick, NodeID: scopeID})
}

func getLatestTickTx(tx *gorm.DB, worldID string) (*store.TimelineModel, error) {
	worldInt := txResolveWorldUUID(tx, worldID)
	var tick store.TimelineModel
	if err := tx.Where("world_id = ?", worldInt).Order("tick_number DESC").First(&tick).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, notFoundf("tick not found for world: %s", worldID)
		}
		return nil, err
	}
	return &tick, nil
}

func worldPlanSummary(resp *engine.InvokeResponse) string {
	if resp.WorldChangePlan == nil {
		return ""
	}
	return resp.WorldChangePlan.Summary
}
