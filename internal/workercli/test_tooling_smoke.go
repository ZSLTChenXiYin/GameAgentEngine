package workercli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workertest"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type toolingSmokeResult struct {
	EnginePort       int                      `json:"engine_port"`
	ConfigPath       string                   `json:"config_path"`
	DBPath           string                   `json:"db_path"`
	WorldID          string                   `json:"world_id"`
	NodeID           string                   `json:"node_id"`
	PendingTaskID    string                   `json:"pending_task_id"`
	LatestTickNumber int                      `json:"latest_tick_number"`
	Checks           []workertest.CheckResult `json:"checks"`
}

func (a *app) runToolingSmokeScenario() error {
	if strings.TrimSpace(a.cfg.TestEngineExePath) == "" {
		return fmt.Errorf("tooling-smoke requires --engine-exe")
	}
	if strings.TrimSpace(a.cfg.TestDevCLIExePath) == "" {
		return fmt.Errorf("tooling-smoke requires --devcli-exe")
	}
	testsDir := strings.TrimSpace(a.cfg.TestsDir)
	if testsDir == "" {
		return fmt.Errorf("tooling-smoke requires --tests-dir")
	}
	fixtureFile := filepath.Join(testsDir, "tooling_smoke_fixture.json")
	tradeFile := filepath.Join(testsDir, "runtime_task_dynamic_action_trade.json")
	worldTimeSettingsPath := filepath.Join(testsDir, "world_time_settings_flexible.json")
	worldStatePath := filepath.Join(testsDir, "state_world_state.json")
	tickPolicyPath := filepath.Join(testsDir, "state_tick_policy.json")
	for _, path := range []string{fixtureFile, tradeFile, worldTimeSettingsPath, worldStatePath, tickPolicyPath} {
		if _, err := os.Stat(path); err != nil {
			return err
		}
	}

	enginePort := a.cfg.TestEnginePort
	apiKey := firstNonEmptyValue(a.cfg.EngineAPIKey, "dev-key")
	runtimeTaskToken := firstNonEmptyValue(a.cfg.RuntimeTaskToken, "dev-task-token")

	tempRoot, err := workertest.MakeTempRoot("gae-s8-src")
	if err != nil {
		return err
	}
	defer workertest.RemoveTempRoot(tempRoot, a.cfg.TestKeepTemp)
	files := workertest.PrepareEngineFiles(tempRoot)
	configText := fmt.Sprintf(`server:
  host: "127.0.0.1"
  port: %d

database:
  driver: "sqlite"
  dsn: "%s"
  migrations_enabled: true

auth:
  api_key: "%s"
  callback_token: "dev-callback-token"
  runtime_task_token: "%s"

llm:
  provider: "fixture"
  model: "fixture-s8"
  api_key: ""
  base_url: ""
  fixture_file: "%s"

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  runtime_task_governance_interval_seconds: 0

external_interfaces:
  npc_trade_action:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    resume_policy: "none"
`, enginePort, workertest.EscapeYAMLPath(files.DBPath), apiKey, runtimeTaskToken, workertest.EscapeYAMLPath(fixtureFile))
	if err := workertest.WriteEngineConfig(files.ConfigPath, configText); err != nil {
		return err
	}

	engineProc, err := workertest.StartProcess(a.cfg.TestEngineExePath, []string{"serve", "--config", files.ConfigPath}, "", files.EngineStdout, files.EngineStderr)
	if err != nil {
		return err
	}
	defer workertest.StopProcess(engineProc)
	if err := workertest.WaitHealthy(fmt.Sprintf("http://127.0.0.1:%d/health", enginePort), 30*time.Second); err != nil {
		return err
	}

	engineBaseURL := fmt.Sprintf("http://127.0.0.1:%d", enginePort)
	client := sdk.NewClient(engineBaseURL, apiKey)
	devcli := workertest.DevCLI{Executable: a.cfg.TestDevCLIExePath, Server: engineBaseURL, APIKey: apiKey}
	rtClient := &workertest.Client{BaseURL: engineBaseURL, APIKey: apiKey, RuntimeTaskToken: runtimeTaskToken}
	collector := &workertest.Collector{}

	tradeInterfaces, err := loadDynamicInterfacesFile(tradeFile)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(tradeInterfaces) >= 1, "tooling-smoke dynamic interfaces missing"); err != nil {
		return err
	}

	var world sdk.Node
	if err := devcli.RunJSON(&world, "node", "create", "--type", "world", "--name", "FullFunctionalToolingSmokeWorld"); err != nil {
		return err
	}
	worldID := world.ID
	var npc sdk.Node
	if err := devcli.RunJSON(&npc, "node", "create", "--world", worldID, "--type", "npc", "--name", "Tooling Merchant"); err != nil {
		return err
	}
	npcID := npc.ID

	var settings sdk.WorldSettings
	if err := devcli.RunJSON(&settings, "world", "settings", "set", worldID, "--world-time-settings-file", worldTimeSettingsPath, "--pipeline-mode", "full"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(settings.PipelineMode, "full", "tooling smoke pipeline mode mismatch"); err != nil {
		return err
	}
	worldStatePayload, err := readJSONFile(worldStatePath)
	if err != nil {
		return err
	}
	tickPolicyPayload, err := readJSONFile(tickPolicyPath)
	if err != nil {
		return err
	}
	if _, err := client.PutStateComponent(worldID, "world_state", worldStatePayload); err != nil {
		return err
	}
	if _, err := client.PutStateComponent(worldID, "tick_policy", tickPolicyPayload); err != nil {
		return err
	}
	collector.Add("continuity", "seed settings and state", "devcli/sdk", "passed", "world_state/tick_policy/world_time_settings")

	invoke, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  worldID,
		NodeID:   npcID,
		TaskType: "custom",
		Messages: []sdk.ChatMessage{{Role: "user", Content: "tooling smoke task"}},
		Context:  &sdk.InvokeContext{DynamicInterfaces: tradeInterfaces},
	})
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(invoke.ActionCalls) >= 1, "tooling smoke callback_id missing"); err != nil {
		return err
	}
	callbackID := invoke.ActionCalls[0].CallbackID
	pendingTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, callbackID, sdk.RuntimeTaskStatusPending, 20*time.Second)
	if err != nil {
		return err
	}
	pendingTasks, err := client.ListPendingRuntimeTasks("bridge", 20)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(pendingTasks) >= 1, "tooling smoke expected pending tasks"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(pendingTask.InterfaceName, "npc_trade_action", "tooling smoke pending task interface mismatch"); err != nil {
		return err
	}
	collector.Add("sdk", "seed runtime task", "sdk", "passed", fmt.Sprintf("callback_id=%s pending_task_id=%s", callbackID, pendingTask.TaskID))

	requestedTicks := 2
	autonomousLimit := 0
	tick, err := client.AdvanceTickWithOptions(worldID, "manual", "day-12 hour-8", &requestedTicks, &autonomousLimit)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(tick.AdvancedTicks, 2, "tooling smoke tick advanced_ticks mismatch"); err != nil {
		return err
	}
	collector.Add("continuity", "seed world tick", "sdk", "passed", fmt.Sprintf("request_id=%s advanced_ticks=%d", tick.Invoke.RequestID, tick.AdvancedTicks))

	toolingSummary, err := a.runToolingSmokeSDKChecks(client, worldID)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(toolingSummary.PendingTaskIface, "npc_trade_action", "sdk pending task interface mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(toolingSummary.TraceCount >= 1, "sdk trace_count should be >= 1"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(toolingSummary.ContinuityTraceCount >= 1, "sdk continuity_trace_count should be >= 1"); err != nil {
		return err
	}
	collector.Add("sdk", "runtime task helper smoke", "sdk", "passed", fmt.Sprintf("pending_task_id=%s latest_tick=%d", toolingSummary.PendingTaskID, toolingSummary.LatestTickNumber))

	var nodeList []sdk.Node
	if err := devcli.RunJSON(&nodeList, "node", "list", "--world", worldID); err != nil {
		return err
	}
	var legacyNodes []sdk.Node
	if err := devcli.RunJSON(&legacyNodes, "nodes", "--world", worldID); err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(nodeList), len(legacyNodes), "DevCli node list and legacy nodes count mismatch"); err != nil {
		return err
	}
	var taskInspect sdk.RuntimeTask
	if err := devcli.RunJSON(&taskInspect, "task", "get", toolingSummary.PendingTaskID, "--json"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(taskInspect.TaskID, toolingSummary.PendingTaskID, "DevCli task get id mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(taskInspect.InterfaceName, "npc_trade_action", "DevCli task get interface mismatch"); err != nil {
		return err
	}
	var continuity sdk.ContinuityBundle
	if err := devcli.RunJSON(&continuity, "debug", "continuity", worldID, "--json"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(continuity.LatestTimeline != nil, "DevCli continuity latest_timeline missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(continuity.Traces) >= 1, "DevCli continuity traces missing"); err != nil {
		return err
	}
	var traces sdk.DebugTraceList
	if err := devcli.RunJSON(&traces, "debug", "traces", "--world", worldID, "--json"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(traces.Count >= 1, "DevCli traces count should be >= 1"); err != nil {
		return err
	}
	collector.Add("devcli", "node/task compatibility smoke", "devcli", "passed", fmt.Sprintf("node_count=%d task_id=%s", len(nodeList), taskInspect.TaskID))

	worldTasks, err := workertest.GetWorldTasks(rtClient, worldID, 100)
	if err != nil {
		return err
	}
	stats, err := client.GetRuntimeTaskStats()
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(worldTasks) >= 1, "Creator Tasks source task list should contain at least one task"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(containsRuntimeTaskID(worldTasks, toolingSummary.PendingTaskID), "Creator Tasks task list missing SDK pending task"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(stats != nil && stats.Total >= 1, "Creator Tasks stats total should be >= 1"); err != nil {
		return err
	}
	collector.Add("creator", "Tasks page smoke", "api", "passed", fmt.Sprintf("task_id=%s stats_total=%d", toolingSummary.PendingTaskID, stats.Total))

	latestTimeline, err := client.GetLatestTimeline(worldID)
	if err != nil {
		return err
	}
	timelines, err := client.GetTimelines(worldID, 6)
	if err != nil {
		return err
	}
	stateComponents, err := client.GetStateComponents(worldID)
	if err != nil {
		return err
	}
	logs, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, TaskType: "world_tick", Limit: 60})
	if err != nil {
		return err
	}
	continuityTraces, err := client.GetDebugTraces(worldID, 30)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(latestTimeline != nil && latestTimeline.Timeline.TickNumber >= 1, "Creator Continuity latest timeline missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(timelines.Timelines) >= 1, "Creator Continuity timelines missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(stateComponents.Components) >= 2, "Creator Continuity state components missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(logs) >= 1, "Creator Continuity logs missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(continuityTraces != nil && continuityTraces.Count >= 1, "Creator Continuity traces missing"); err != nil {
		return err
	}
	collector.Add("creator", "Continuity page smoke", "api", "passed", fmt.Sprintf("timeline_tick=%d state_components=%d", latestTimeline.Timeline.TickNumber, len(stateComponents.Components)))

	if err := workertest.AssertTrue(continuityTraces != nil && continuityTraces.Count >= 1, "Creator Traces page source missing traces"); err != nil {
		return err
	}
	collector.Add("creator", "Traces page smoke", "api", "passed", fmt.Sprintf("trace_count=%d", continuityTraces.Count))

	result := toolingSmokeResult{
		EnginePort:       enginePort,
		ConfigPath:       files.ConfigPath,
		DBPath:           files.DBPath,
		WorldID:          worldID,
		NodeID:           npcID,
		PendingTaskID:    toolingSummary.PendingTaskID,
		LatestTickNumber: toolingSummary.LatestTickNumber,
		Checks:           collector.Checks(),
	}
	return a.writeScenarioResult(result)
}

type toolingSmokeSDKSummary struct {
	NodeCount            int
	PendingTaskID        string
	PendingTaskStatus    string
	PendingTaskIface     string
	RuntimeTaskTotal     int64
	LatestTickNumber     int
	StateComponentCount  int
	ContinuityLogCount   int
	ContinuityTraceCount int
	TraceCount           int
	LogCount             int
}

func (a *app) runToolingSmokeSDKChecks(client *sdk.Client, worldID string) (*toolingSmokeSDKSummary, error) {
	nodes, err := client.GetNodes(worldID, 50, 0, "")
	if err != nil {
		return nil, err
	}
	if len(nodes) < 2 {
		return nil, fmt.Errorf("expected at least 2 nodes, got %d", len(nodes))
	}
	pendingTasks, err := client.ListPendingRuntimeTasks("bridge", 20)
	if err != nil {
		return nil, err
	}
	if len(pendingTasks) == 0 {
		return nil, fmt.Errorf("expected at least one pending runtime task")
	}
	stats, err := client.GetRuntimeTaskStats()
	if err != nil {
		return nil, err
	}
	if stats == nil || stats.Total < 1 {
		return nil, fmt.Errorf("expected runtime task stats total >= 1")
	}
	latest, err := client.GetLatestTimeline(worldID)
	if err != nil {
		return nil, err
	}
	if latest == nil || latest.Timeline.TickNumber < 1 {
		return nil, fmt.Errorf("expected latest timeline tick >= 1")
	}
	states, err := client.GetStateComponents(worldID)
	if err != nil {
		return nil, err
	}
	if states == nil || len(states.Components) < 3 {
		return nil, fmt.Errorf("expected at least 3 state components")
	}
	bundle, err := client.GetContinuityBundle(worldID, &sdk.ContinuityBundleOptions{LogLimit: 20, TraceLimit: 10})
	if err != nil {
		return nil, err
	}
	if bundle == nil || bundle.LatestTimeline == nil {
		return nil, fmt.Errorf("expected continuity bundle latest timeline")
	}
	if len(bundle.Logs) == 0 {
		return nil, fmt.Errorf("expected continuity bundle logs")
	}
	if len(bundle.Traces) == 0 {
		return nil, fmt.Errorf("expected continuity bundle traces")
	}
	traces, err := client.GetDebugTraces(worldID, 10)
	if err != nil {
		return nil, err
	}
	if traces == nil || traces.Count < 1 {
		return nil, fmt.Errorf("expected at least one trace")
	}
	logs, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, TaskType: "world_tick", Limit: 20})
	if err != nil {
		return nil, err
	}
	if len(logs) == 0 {
		return nil, fmt.Errorf("expected world_tick logs")
	}
	return &toolingSmokeSDKSummary{
		NodeCount:            len(nodes),
		PendingTaskID:        pendingTasks[0].TaskID,
		PendingTaskStatus:    pendingTasks[0].Status,
		PendingTaskIface:     pendingTasks[0].InterfaceName,
		RuntimeTaskTotal:     stats.Total,
		LatestTickNumber:     latest.Timeline.TickNumber,
		StateComponentCount:  len(states.Components),
		ContinuityLogCount:   len(bundle.Logs),
		ContinuityTraceCount: len(bundle.Traces),
		TraceCount:           traces.Count,
		LogCount:             len(logs),
	}, nil
}

func containsRuntimeTaskID(tasks []sdk.RuntimeTask, taskID string) bool {
	for _, task := range tasks {
		if task.TaskID == taskID {
			return true
		}
	}
	return false
}
