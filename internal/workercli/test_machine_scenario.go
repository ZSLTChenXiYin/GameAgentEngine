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

type machineScenarioResult struct {
	EnginePort            int                      `json:"engine_port"`
	ConfigPath            string                   `json:"config_path"`
	DBPath                string                   `json:"db_path"`
	WorldID               string                   `json:"world_id"`
	NodeID                string                   `json:"node_id"`
	RequestID             string                   `json:"request_id"`
	CallbackID            string                   `json:"callback_id"`
	TaskID                string                   `json:"task_id"`
	LatestTimelinePresent bool                     `json:"latest_timeline_present"`
	Checks                []workertest.CheckResult `json:"checks"`
}

func (a *app) runMachineScenario() error {
	if strings.TrimSpace(a.cfg.TestEngineExePath) == "" {
		return fmt.Errorf("machine-scenario requires --engine-exe")
	}
	if strings.TrimSpace(a.cfg.TestDevCLIExePath) == "" {
		return fmt.Errorf("machine-scenario requires --devcli-exe")
	}
	workerExe := strings.TrimSpace(a.cfg.TestWorkerExePath)
	if workerExe == "" {
		if currentExe, err := os.Executable(); err == nil {
			workerExe = currentExe
		}
	}
	if strings.TrimSpace(workerExe) == "" {
		return fmt.Errorf("machine-scenario requires --worker-exe")
	}
	testsDir := strings.TrimSpace(a.cfg.TestsDir)
	if testsDir == "" {
		return fmt.Errorf("machine-scenario requires --tests-dir")
	}
	fixtureFile := filepath.Join(testsDir, "machine_scenario_fixture.json")
	dynamicInterfacesFile := filepath.Join(testsDir, "runtime_task_dynamic_interfaces.json")
	worldTimeSettingsPath := filepath.Join(testsDir, "world_time_settings_flexible.json")
	worldStatePath := filepath.Join(testsDir, "state_world_state.json")
	storyStatePath := filepath.Join(testsDir, "state_story_state.json")
	storyHistoryPath := filepath.Join(testsDir, "state_story_history.json")
	tickPolicyPath := filepath.Join(testsDir, "state_tick_policy.json")
	for _, path := range []string{fixtureFile, dynamicInterfacesFile, worldTimeSettingsPath, worldStatePath, storyStatePath, storyHistoryPath, tickPolicyPath} {
		if _, err := os.Stat(path); err != nil {
			return err
		}
	}

	enginePort := a.cfg.TestEnginePort
	apiKey := firstNonEmptyValue(a.cfg.EngineAPIKey, "dev-key")
	runtimeTaskToken := firstNonEmptyValue(a.cfg.RuntimeTaskToken, "dev-task-token")
	callbackToken := firstNonEmptyValue(a.cfg.CallbackToken, "dev-callback-token")

	tempRoot, err := workertest.MakeTempRoot("gae-s9-src")
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
  callback_token: "%s"
  runtime_task_token: "%s"
  callback_require_request_id: true

llm:
  provider: "fixture"
  model: "fixture-s9"
  api_key: ""
  base_url: ""
  fixture_file: "%s"

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  runtime_task_governance_interval_seconds: 0

external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "game_client"
    resume_policy: "resume_paused_execution"
`, enginePort, workertest.EscapeYAMLPath(files.DBPath), apiKey, callbackToken, runtimeTaskToken, workertest.EscapeYAMLPath(fixtureFile))
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
	rtClient := &workertest.Client{BaseURL: engineBaseURL, APIKey: apiKey, RuntimeTaskToken: runtimeTaskToken, CallbackToken: callbackToken}
	collector := &workertest.Collector{}

	dynamicInterfaces, err := loadDynamicInterfacesFile(dynamicInterfacesFile)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(dynamicInterfaces) >= 1, "machine scenario dynamic interfaces missing"); err != nil {
		return err
	}

	collector.Add("runtime", "isolated Engine runtime started", "process", "passed", fmt.Sprintf("port=%d config=%s", enginePort, files.ConfigPath))

	var world sdk.Node
	if err := devcli.RunJSON(&world, "node", "create", "--type", "world", "--name", "FullFunctionalMachineScenarioWorld"); err != nil {
		return err
	}
	worldID := world.ID
	var npc sdk.Node
	if err := devcli.RunJSON(&npc, "node", "create", "--world", worldID, "--type", "npc", "--name", "Scene Broker"); err != nil {
		return err
	}
	npcID := npc.ID

	var settings sdk.WorldSettings
	if err := devcli.RunJSON(&settings, "world", "settings", "set", worldID, "--world-time-settings-file", worldTimeSettingsPath, "--pipeline-mode", "full"); err != nil {
		return err
	}
	worldStatePayload, err := readJSONFile(worldStatePath)
	if err != nil {
		return err
	}
	storyStatePayload, err := readJSONFile(storyStatePath)
	if err != nil {
		return err
	}
	storyHistoryPayload, err := readJSONFile(storyHistoryPath)
	if err != nil {
		return err
	}
	tickPolicyPayload, err := readJSONFile(tickPolicyPath)
	if err != nil {
		return err
	}
	for componentType, payload := range map[string]any{
		"world_state":   worldStatePayload,
		"story_state":   storyStatePayload,
		"story_history": storyHistoryPayload,
		"tick_policy":   tickPolicyPayload,
	} {
		if _, err := client.PutStateComponent(worldID, componentType, payload); err != nil {
			return err
		}
	}
	if _, err := client.PutStateComponent(worldID, "world_time_state", map[string]any{
		"current_time_label":  "Cycle day 12 hour 8",
		"total_ticks":         2,
		"last_tick_number":    1,
		"last_tick_type":      "manual",
		"last_advanced_ticks": 2,
	}); err != nil {
		return err
	}

	invoke, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  worldID,
		NodeID:   npcID,
		TaskType: "npc_dialogue",
		Messages: []sdk.ChatMessage{{Role: "user", Content: "Before answering, query the nearby scene and then respond."}},
		Context: &sdk.InvokeContext{
			PipelineMode:      "full",
			DynamicInterfaces: dynamicInterfaces,
		},
	})
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(invoke.ActionCalls) >= 1, "machine scenario callback_id missing"); err != nil {
		return err
	}
	callbackID := invoke.ActionCalls[0].CallbackID
	requestID := invoke.RequestID
	if err := workertest.AssertTrue(strings.TrimSpace(callbackID) != "", "machine scenario callback_id missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.TrimSpace(requestID) != "", "machine scenario request_id missing"); err != nil {
		return err
	}
	collector.Add("invoke", "NPC dialogue invoke with request-scoped dynamic interfaces", "sdk", "passed", fmt.Sprintf("request_id=%s callback_id=%s", requestID, callbackID))

	task, err := workertest.WaitTaskStatus(rtClient, worldID, 50, callbackID, sdk.RuntimeTaskStatusPending, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(task.InterfaceName, "game_client_request_data", "machine scenario runtime task interface mismatch"); err != nil {
		return err
	}
	collector.Add("runtime", "runtime task created", "sdk", "passed", fmt.Sprintf("task_id=%s interface=%s", task.TaskID, task.InterfaceName))

	if err := execWorkerCommand(workerExe, []string{"pull-once", "--engine-base-url", engineBaseURL, "--runtime-task-token", runtimeTaskToken, "--callback-token", callbackToken, "--consumer", "game_client", "--lease-owner", "s9-worker"}); err != nil {
		return err
	}
	collector.Add("worker", "test worker started", "worker", "passed", "consumer=game_client lease_owner=s9-worker")

	completedTask, err := workertest.WaitTaskStatus(rtClient, worldID, 50, callbackID, sdk.RuntimeTaskStatusSucceeded, 20*time.Second)
	if err != nil {
		return err
	}
	collector.Add("worker", "callback completed by worker", "worker", "passed", fmt.Sprintf("task_id=%s callback_id=%s", completedTask.TaskID, callbackID))

	requestLogs, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, RequestID: requestID, Limit: 100})
	if err != nil {
		return err
	}
	resumeLogCount := 0
	pausedLogCount := 0
	for _, item := range requestLogs {
		if item.EventName == "resume_completed" {
			resumeLogCount++
		}
		if item.EventName == "data_request_paused_for_client" {
			pausedLogCount++
		}
	}
	if err := workertest.AssertEqual(resumeLogCount, 1, "machine scenario resume_completed log count mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(pausedLogCount, 1, "machine scenario paused log count mismatch"); err != nil {
		return err
	}
	collector.Add("resume", "paused execution resumed", "logs", "passed", fmt.Sprintf("resume_logs=%d paused_logs=%d", resumeLogCount, pausedLogCount))

	allLogs, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, Limit: 100})
	if err != nil {
		return err
	}
	traces, err := client.GetDebugTraces(worldID, 20)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(traces != nil && traces.Count >= 1, "machine scenario traces missing"); err != nil {
		return err
	}
	var continuity sdk.ContinuityBundle
	if err := devcli.RunJSON(&continuity, "debug", "continuity", worldID, "--request-id", requestID, "--log-limit", "20", "--trace-limit", "10", "--json"); err != nil {
		return err
	}
	latestTimelinePresent := continuity.LatestTimeline != nil
	if err := workertest.AssertTrue(len(continuity.Logs) >= 1, "machine scenario continuity logs missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(continuity.Traces) >= 1, "machine scenario continuity traces missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(continuity.StateComponents) >= 4, "machine scenario continuity state components missing"); err != nil {
		return err
	}
	collector.Add("observability", "logs / traces / continuity confirmed", "sdk/devcli", "passed", fmt.Sprintf("request_id=%s continuity_logs=%d continuity_traces=%d continuity_state_components=%d latest_timeline_present=%t total_logs=%d", requestID, len(continuity.Logs), len(continuity.Traces), len(continuity.StateComponents), latestTimelinePresent, len(allLogs)))

	result := machineScenarioResult{
		EnginePort:            enginePort,
		ConfigPath:            files.ConfigPath,
		DBPath:                files.DBPath,
		WorldID:               worldID,
		NodeID:                npcID,
		RequestID:             requestID,
		CallbackID:            callbackID,
		TaskID:                completedTask.TaskID,
		LatestTimelinePresent: latestTimelinePresent,
		Checks:                collector.Checks(),
	}
	return a.writeScenarioResult(result)
}
