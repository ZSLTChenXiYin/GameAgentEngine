package workercli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workertest"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type callbackResumeResult struct {
	EnginePort              int                      `json:"engine_port"`
	ConfigPath              string                   `json:"config_path"`
	DBPath                  string                   `json:"db_path"`
	WorldID                 string                   `json:"world_id"`
	SceneCallbackID         string                   `json:"scene_callback_id"`
	SceneTaskID             string                   `json:"scene_task_id"`
	NoneCallbackID          string                   `json:"none_callback_id"`
	NoneTaskID              string                   `json:"none_task_id"`
	RecordOnlyCallbackID    string                   `json:"record_only_callback_id"`
	RecordOnlyTaskID        string                   `json:"record_only_task_id"`
	WriteMemoryCallbackID   string                   `json:"write_memory_callback_id"`
	WriteMemoryTaskID       string                   `json:"write_memory_task_id"`
	FailureCallbackID       string                   `json:"failure_callback_id"`
	FailureTaskID           string                   `json:"failure_task_id"`
	Checks                  []workertest.CheckResult `json:"checks"`
}

type callbackHTTPResponse struct {
	StatusCode int
	Header     http.Header
	Body       sdk.CallbackResponse
}

func (a *app) runCallbackResumeScenario() error {
	if strings.TrimSpace(a.cfg.TestEngineExePath) == "" {
		return fmt.Errorf("callback-resume requires --engine-exe")
	}
	if strings.TrimSpace(a.cfg.TestDevCLIExePath) == "" {
		return fmt.Errorf("callback-resume requires --devcli-exe")
	}
	workerExe := strings.TrimSpace(a.cfg.TestWorkerExePath)
	if workerExe == "" {
		if currentExe, err := os.Executable(); err == nil {
			workerExe = currentExe
		}
	}
	if strings.TrimSpace(workerExe) == "" {
		return fmt.Errorf("callback-resume requires --worker-exe")
	}
	testsDir := strings.TrimSpace(a.cfg.TestsDir)
	if testsDir == "" {
		return fmt.Errorf("callback-resume requires --tests-dir")
	}
	fixtureFile := filepath.Join(testsDir, "callback_resume_fixture.json")
	dataInterfacesFile := filepath.Join(testsDir, "runtime_task_dynamic_interfaces.json")
	actionInterfacesFile := filepath.Join(testsDir, "callback_resume_dynamic_actions.json")
	for _, path := range []string{fixtureFile, dataInterfacesFile, actionInterfacesFile} {
		if _, err := os.Stat(path); err != nil {
			return err
		}
	}
	enginePort := a.cfg.TestEnginePort
	apiKey := firstNonEmptyValue(a.cfg.EngineAPIKey, "dev-key")
	callbackToken := firstNonEmptyValue(a.cfg.CallbackToken, "dev-callback-token")
	runtimeTaskToken := firstNonEmptyValue(a.cfg.RuntimeTaskToken, "dev-task-token")

	tempRoot, err := workertest.MakeTempRoot("gae-s7-src")
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
  model: "fixture-s7"
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

  game_client_request_data_no_resume:
    category: "external_query"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "game_client"
    resume_policy: "none"

  spawn_item_record_only:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    resume_policy: "none"
    callback_post_process: "record_only"

  spawn_item_write_memory:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    resume_policy: "none"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"

  spawn_item_failure:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    resume_policy: "none"
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

	dataInterfaces, err := loadDynamicInterfacesFile(dataInterfacesFile)
	if err != nil {
		return err
	}
	actionInterfaces, err := loadDynamicInterfacesFile(actionInterfacesFile)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(dataInterfaces) >= 1, "callback-resume data interfaces missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(actionInterfaces) >= 3, "callback-resume action interfaces missing"); err != nil {
		return err
	}

	var world sdk.Node
	if err := devcli.RunJSON(&world, "node", "create", "--type", "world", "--name", "FullFunctionalCallbackResumeWorld"); err != nil {
		return err
	}
	worldID := world.ID
	var npc sdk.Node
	if err := devcli.RunJSON(&npc, "node", "create", "--world", worldID, "--type", "npc", "--name", "Callback Broker"); err != nil {
		return err
	}
	npcID := npc.ID

	sceneInvoke, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  worldID,
		NodeID:   npcID,
		TaskType: "custom",
		Messages: []sdk.ChatMessage{{Role: "user", Content: "scene pause"}},
		Context:  &sdk.InvokeContext{DynamicInterfaces: dataInterfaces},
	})
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(sceneInvoke.ActionCalls) >= 1, "scene invoke action_calls missing"); err != nil {
		return err
	}
	sceneCallbackID := sceneInvoke.ActionCalls[0].CallbackID
	if err := workertest.AssertTrue(strings.TrimSpace(sceneCallbackID) != "", "scene callback_id missing"); err != nil {
		return err
	}
	sceneTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, sceneCallbackID, sdk.RuntimeTaskStatusPending, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(sceneTask.InterfaceName, "game_client_request_data", "scene task interface mismatch"); err != nil {
		return err
	}

	sceneCallbackResp, err := invokeCallbackWithRequestID(engineBaseURL, callbackToken, sceneCallbackID, "success", map[string]any{"scene": "starter_inn", "npc": "merchant"}, "s7-scene-1")
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(sceneCallbackResp.StatusCode, http.StatusOK, "scene callback status code mismatch"); err != nil {
		return err
	}
	if sceneCallbackResp.Body.Resumed == nil {
		return fmt.Errorf("scene callback resumed payload missing")
	}
	if err := workertest.AssertEqual(sceneCallbackResp.Body.Resumed.Reply, "scene-resumed-final", "scene resumed reply mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.TrimSpace(sceneCallbackResp.Body.ResumeExecutionID) != "", "scene resume_execution_id missing"); err != nil {
		return err
	}
	sceneSucceededTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, sceneCallbackID, sdk.RuntimeTaskStatusSucceeded, 20*time.Second)
	if err != nil {
		return err
	}
	resumeLogs, err := waitLogCount(client, worldID, "resume_completed", 1, 15*time.Second)
	if err != nil {
		return err
	}
	reuseLogs, err := waitLogCount(client, worldID, "data_request_reused", 1, 15*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(resumeLogs), 1, "resume_completed count mismatch after first resume"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(reuseLogs), 1, "data_request_reused count mismatch"); err != nil {
		return err
	}

	sceneReplayResp, err := invokeCallbackWithRequestID(engineBaseURL, callbackToken, sceneCallbackID, "success", map[string]any{"scene": "starter_inn", "npc": "merchant"}, "s7-scene-1")
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(sceneReplayResp.StatusCode, http.StatusOK, "scene replay callback status code mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(sceneReplayResp.Header.Get("X-Callback-Replayed"), "true", "scene replay header mismatch"); err != nil {
		return err
	}
	resumeLogsAfterReplay, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, EventName: "resume_completed", Limit: 100})
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(resumeLogsAfterReplay), 1, "scene replay should not duplicate resume_completed log"); err != nil {
		return err
	}
	sceneTasks, err := workertest.GetWorldTasks(rtClient, worldID, 200)
	if err != nil {
		return err
	}
	sceneTaskCount := 0
	for _, task := range sceneTasks {
		if task.InterfaceName == "game_client_request_data" {
			sceneTaskCount++
		}
	}
	if err := workertest.AssertEqual(sceneTaskCount, 1, "scene resume should not create duplicate data_request runtime task"); err != nil {
		return err
	}
	collector.Add("callback", "success path and auto-resume", "http", "passed", fmt.Sprintf("callback_id=%s task_id=%s reply=%s", sceneCallbackID, sceneSucceededTask.TaskID, sceneCallbackResp.Body.Resumed.Reply))
	collector.Add("callback", "replay protection", "http", "passed", fmt.Sprintf("callback_id=%s replayed=%s", sceneCallbackID, sceneReplayResp.Header.Get("X-Callback-Replayed")))
	collector.Add("callback", "duplicate data_request suppression", "sdk", "passed", fmt.Sprintf("callback_id=%s reuse_logs=%d task_count=%d", sceneCallbackID, len(reuseLogs), sceneTaskCount))

	noneInvoke, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  worldID,
		NodeID:   npcID,
		TaskType: "custom",
		Messages: []sdk.ChatMessage{{Role: "user", Content: "scene none"}},
		Context: &sdk.InvokeContext{DynamicInterfaces: []sdk.DynamicInterface{{
			ID:                "scene_query_none",
			Kind:              sdk.DynamicInterfaceDataRequest,
			ExternalInterface: "game_client_request_data_no_resume",
			Description:       "Query scene without auto resume.",
			QueryTypes:        []string{"node_detail"},
			MaxQueries:        1,
		}}},
	})
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(noneInvoke.ActionCalls) >= 1, "none invoke action_calls missing"); err != nil {
		return err
	}
	noneCallbackID := noneInvoke.ActionCalls[0].CallbackID
	if _, err := workertest.WaitTaskStatus(rtClient, worldID, 200, noneCallbackID, sdk.RuntimeTaskStatusPending, 20*time.Second); err != nil {
		return err
	}
	noneCallbackResp, err := invokeCallbackWithRequestID(engineBaseURL, callbackToken, noneCallbackID, "success", map[string]any{"scene": "outer_gate"}, "s7-none-1")
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(noneCallbackResp.Body.Resumed == nil, "resume_policy=none should not return resumed payload"); err != nil {
		return err
	}
	noneTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, noneCallbackID, sdk.RuntimeTaskStatusSucceeded, 20*time.Second)
	if err != nil {
		return err
	}
	resumeLogsAfterNone, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, EventName: "resume_completed", Limit: 100})
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(resumeLogsAfterNone), 1, "resume_policy=none should not add resume_completed log"); err != nil {
		return err
	}
	collector.Add("callback", "resume_policy none", "http", "passed", fmt.Sprintf("callback_id=%s task_id=%s", noneCallbackID, noneTask.TaskID))

	recordInvoke, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  worldID,
		NodeID:   npcID,
		TaskType: "custom",
		Messages: []sdk.ChatMessage{{Role: "user", Content: "record only action"}},
		Context:  &sdk.InvokeContext{DynamicInterfaces: []sdk.DynamicInterface{actionInterfaces[0]}},
	})
	if err != nil {
		return err
	}
	recordCallbackID := recordInvoke.ActionCalls[0].CallbackID
	recordTaskPending, err := workertest.WaitTaskStatus(rtClient, worldID, 200, recordCallbackID, sdk.RuntimeTaskStatusPending, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(recordTaskPending.InterfaceName, "spawn_item_record_only", "record_only task interface mismatch"); err != nil {
		return err
	}
	if err := execWorkerCommand(workerExe, []string{"pull-once", "--engine-base-url", engineBaseURL, "--runtime-task-token", runtimeTaskToken, "--callback-token", callbackToken, "--consumer", "bridge", "--lease-owner", "s7-record"}); err != nil {
		return err
	}
	recordTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, recordCallbackID, sdk.RuntimeTaskStatusSucceeded, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.TrimSpace(recordTask.ResumeExecutionID) == "", "record_only task should not include resume_execution_id"); err != nil {
		return err
	}
	recordMemories, err := client.GetMemories(npcID)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(recordMemories), 0, "record_only should not write memory"); err != nil {
		return err
	}
	collector.Add("callback", "post-process record_only", "worker", "passed", fmt.Sprintf("callback_id=%s task_id=%s", recordCallbackID, recordTask.TaskID))

	writeInvoke, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  worldID,
		NodeID:   npcID,
		TaskType: "custom",
		Messages: []sdk.ChatMessage{{Role: "user", Content: "write memory action"}},
		Context:  &sdk.InvokeContext{DynamicInterfaces: []sdk.DynamicInterface{actionInterfaces[1]}},
	})
	if err != nil {
		return err
	}
	writeCallbackID := writeInvoke.ActionCalls[0].CallbackID
	writeTaskPending, err := workertest.WaitTaskStatus(rtClient, worldID, 200, writeCallbackID, sdk.RuntimeTaskStatusPending, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(writeTaskPending.InterfaceName, "spawn_item_write_memory", "write_memory task interface mismatch"); err != nil {
		return err
	}
	if err := execWorkerCommand(workerExe, []string{"pull-once", "--engine-base-url", engineBaseURL, "--runtime-task-token", runtimeTaskToken, "--callback-token", callbackToken, "--consumer", "bridge", "--lease-owner", "s7-write"}); err != nil {
		return err
	}
	writeTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, writeCallbackID, sdk.RuntimeTaskStatusSucceeded, 20*time.Second)
	if err != nil {
		return err
	}
	writeMemories, err := client.GetMemories(npcID)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(writeMemories), 1, "write_memory should create one memory"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(writeMemories[0].Level, "long_term", "write_memory level mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.Contains(writeMemories[0].Content, "spawn callback success:"), "write_memory content missing success template"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.Contains(writeMemories[0].Content, `"item_name":"potion"`), "write_memory content missing callback payload"); err != nil {
		return err
	}
	collector.Add("callback", "post-process write_memory", "worker", "passed", fmt.Sprintf("callback_id=%s memory_id=%s", writeCallbackID, writeMemories[0].ID))

	failureInvoke, err := client.Invoke(&sdk.InvokeRequest{
		WorldID:  worldID,
		NodeID:   npcID,
		TaskType: "custom",
		Messages: []sdk.ChatMessage{{Role: "user", Content: "failure action"}},
		Context:  &sdk.InvokeContext{DynamicInterfaces: []sdk.DynamicInterface{actionInterfaces[2]}},
	})
	if err != nil {
		return err
	}
	failureCallbackID := failureInvoke.ActionCalls[0].CallbackID
	failureTaskPending, err := workertest.WaitTaskStatus(rtClient, worldID, 200, failureCallbackID, sdk.RuntimeTaskStatusPending, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(failureTaskPending.InterfaceName, "spawn_item_failure", "failure task interface mismatch"); err != nil {
		return err
	}
	if err := execWorkerCommand(workerExe, []string{"pull-once", "--engine-base-url", engineBaseURL, "--runtime-task-token", runtimeTaskToken, "--callback-token", callbackToken, "--consumer", "bridge", "--lease-owner", "s7-fail", "--fail-interface", "spawn_item_failure"}); err != nil {
		return err
	}
	failureTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, failureCallbackID, sdk.RuntimeTaskStatusFailed, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.Contains(failureTask.ErrorMessage, `"status":"failed"`), "failure task error_message should include callback payload"); err != nil {
		return err
	}
	collector.Add("callback", "failure path", "worker", "passed", fmt.Sprintf("callback_id=%s task_id=%s", failureCallbackID, failureTask.TaskID))

	result := callbackResumeResult{
		EnginePort:            enginePort,
		ConfigPath:            files.ConfigPath,
		DBPath:                files.DBPath,
		WorldID:               worldID,
		SceneCallbackID:       sceneCallbackID,
		SceneTaskID:           sceneSucceededTask.TaskID,
		NoneCallbackID:        noneCallbackID,
		NoneTaskID:            noneTask.TaskID,
		RecordOnlyCallbackID:  recordCallbackID,
		RecordOnlyTaskID:      recordTask.TaskID,
		WriteMemoryCallbackID: writeCallbackID,
		WriteMemoryTaskID:     writeTask.TaskID,
		FailureCallbackID:     failureCallbackID,
		FailureTaskID:         failureTask.TaskID,
		Checks:                collector.Checks(),
	}
	return a.writeScenarioResult(result)
}

func invokeCallbackWithRequestID(baseURL string, callbackToken string, callbackID string, status string, result any, callbackRequestID string) (*callbackHTTPResponse, error) {
	body, err := json.Marshal(map[string]any{
		"callback_id": callbackID,
		"status":      status,
		"result":      result,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(baseURL, "/")+"/api/v1/actions/callback", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Callback-Token", callbackToken)
	if strings.TrimSpace(callbackRequestID) != "" {
		req.Header.Set("X-Callback-Request-Id", callbackRequestID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out := &callbackHTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header.Clone()}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("callback request failed: %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(&out.Body); err != nil && err.Error() != "EOF" {
		return nil, err
	}
	return out, nil
}

func waitLogCount(client *sdk.Client, worldID string, eventName string, minimum int, timeout time.Duration) ([]sdk.InferenceLog, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		logs, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, EventName: eventName, Limit: 100})
		if err != nil {
			return nil, err
		}
		if len(logs) >= minimum {
			return logs, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, EventName: eventName, Limit: 100})
}

func loadDynamicInterfacesFile(path string) ([]sdk.DynamicInterface, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var items []sdk.DynamicInterface
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func execWorkerCommandWithOutput(workerExe string, args []string) ([]byte, error) {
	cmd := exec.Command(workerExe, args...)
	return cmd.CombinedOutput()
}
