package workercli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workertest"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type runtimeTasksResult struct {
	EnginePort   int                      `json:"engine_port"`
	PushPort     int                      `json:"push_port"`
	ConfigPath   string                   `json:"config_path"`
	DBPath       string                   `json:"db_path"`
	WorldID      string                   `json:"world_id"`
	PushTaskID   string                   `json:"push_task_id"`
	PullTaskID   string                   `json:"pull_task_id"`
	HybridTaskID string                   `json:"hybrid_task_id"`
	TimeoutTaskID string                  `json:"timeout_task_id"`
	Checks       []workertest.CheckResult `json:"checks"`
}

func (a *app) runRuntimeTasksScenario() error {
	if strings.TrimSpace(a.cfg.TestEngineExePath) == "" {
		return fmt.Errorf("runtime-tasks requires --engine-exe")
	}
	if strings.TrimSpace(a.cfg.TestDevCLIExePath) == "" {
		return fmt.Errorf("runtime-tasks requires --devcli-exe")
	}
	workerExe := strings.TrimSpace(a.cfg.TestWorkerExePath)
	if workerExe == "" {
		if currentExe, err := os.Executable(); err == nil {
			workerExe = currentExe
		}
	}
	if strings.TrimSpace(workerExe) == "" {
		return fmt.Errorf("runtime-tasks requires --worker-exe")
	}
	testsDir := strings.TrimSpace(a.cfg.TestsDir)
	if testsDir == "" {
		return fmt.Errorf("runtime-tasks requires --tests-dir")
	}
	fixtureFile := filepath.Join(testsDir, "runtime_task_delivery_fixture.json")
	tradeFile := filepath.Join(testsDir, "runtime_task_dynamic_action_trade.json")
	enginePort := a.cfg.TestEnginePort
	pushPort := a.cfg.TestPushPort
	apiKey := firstNonEmptyValue(a.cfg.EngineAPIKey, "dev-key")
	callbackToken := firstNonEmptyValue(a.cfg.CallbackToken, "dev-callback-token")
	runtimeTaskToken := firstNonEmptyValue(a.cfg.RuntimeTaskToken, "dev-task-token")
	bearerToken := firstNonEmptyValue(a.cfg.GameHTTPBearerToken, "local-test-token")

	tempRoot, err := workertest.MakeTempRoot("gae-s6-src")
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
  write_retry_enabled: true
  write_retry_max_attempts: 3
  write_retry_base_delay_ms: 40
  write_retry_max_delay_ms: 250
  log_batch_enabled: true
  log_batch_size: 32
  log_batch_flush_ms: 750
  log_batch_queue_size: 1024

auth:
  api_key: "%s"
  callback_token: "%s"
  runtime_task_token: "%s"
  callback_require_request_id: true

llm:
  provider: "fixture"
  model: "fixture-s6"
  api_key: ""
  base_url: ""
  fixture_file: "%s"

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
  runtime_task_governance_interval_seconds: 0
  runtime_task_heartbeat_timeout_seconds: 2
  runtime_task_auto_requeue_enabled: false
  runtime_task_auto_requeue_limit: 100
  runtime_task_auto_requeue_delay_ms: 50

external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:%d"
    path: "/api/v1/runtime/dispatch"
    timeout_ms: 1000
    retry_max_attempts: 1
    retry_backoff_ms: 50
    idempotency_header: "Idempotency-Key"
    auth:
      mode: "bearer"
      token: "%s"

external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "push"
    primary_transport: "game_http"
    consumer: "game_client"
    resume_policy: "resume_paused_execution"

  spawn_item:
    category: "external_action"
    delivery_mode: "hybrid"
    primary_transport: "game_http"
    fallback_transport: "task_pull"
    consumer: "bridge"
    max_attempts: 3
    heartbeat_timeout_auto_requeue: true
    heartbeat_timeout_requeue_delay_ms: 500
    heartbeat_timeout_reason: "spawn_item timeout auto requeue"
    resume_policy: "none"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"

  npc_trade_action:
    category: "external_action"
    delivery_mode: "pull"
    primary_transport: "task_pull"
    consumer: "bridge"
    max_attempts: 3
    resume_policy: "none"
`, enginePort, workertest.EscapeYAMLPath(files.DBPath), apiKey, callbackToken, runtimeTaskToken, workertest.EscapeYAMLPath(fixtureFile), pushPort, bearerToken)
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

	var world sdk.Node
	if err := devcli.RunJSON(&world, "node", "create", "--type", "world", "--name", "FullFunctionalRuntimeTasksWorld"); err != nil {
		return err
	}
	worldID := world.ID
	var npc sdk.Node
	if err := devcli.RunJSON(&npc, "node", "create", "--world", worldID, "--type", "npc", "--name", "Broker Toma"); err != nil {
		return err
	}
	npcID := npc.ID

	pushWorkerProc, err := workertest.StartProcess(workerExe, []string{
		"push-receiver",
		"--engine-base-url", engineBaseURL,
		"--runtime-task-token", runtimeTaskToken,
		"--callback-token", callbackToken,
		"--game-http-bearer-token", bearerToken,
		"--push-port", fmt.Sprintf("%d", pushPort),
		"--callback-delay", "250ms",
	}, "", files.WorkerStdout, files.WorkerStderr)
	if err != nil {
		return err
	}
	defer workertest.StopProcess(pushWorkerProc)
	time.Sleep(time.Second)

	var pushInvoke sdk.InvokeResponse
	if err := devcli.RunJSON(&pushInvoke, "invoke", worldID, npcID, "--task-type", "custom", "--message", "push runtime task"); err != nil {
		return err
	}
	pushCallbackID := pushInvoke.ActionCalls[0].CallbackID
	if err := workertest.AssertTrue(strings.TrimSpace(pushCallbackID) != "", "push callback_id missing"); err != nil {
		return err
	}
	pushTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, pushCallbackID, sdk.RuntimeTaskStatusSucceeded, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(pushTask.DeliveryMode, "push", "push task delivery_mode mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(pushTask.Transport, "game_http", "push task transport mismatch"); err != nil {
		return err
	}
	collector.Add("push", "delivery success", "worker", "passed", fmt.Sprintf("task_id=%s dispatch_attempts=%d", pushTask.TaskID, pushTask.DispatchAttempts))

	var pullInvoke sdk.InvokeResponse
	if err := devcli.RunJSON(&pullInvoke, "invoke", worldID, npcID, "--task-type", "custom", "--dynamic-interfaces-file", tradeFile, "--message", "pull runtime task"); err != nil {
		return err
	}
	pullCallbackID := pullInvoke.ActionCalls[0].CallbackID
	pullPending, err := workertest.FindTaskByCallbackID(rtClient, worldID, 200, pullCallbackID)
	if err != nil {
		return err
	}
	if pullPending == nil {
		return fmt.Errorf("pull task not found")
	}
	if err := workertest.AssertEqual(pullPending.Status, sdk.RuntimeTaskStatusPending, "pull task should start pending"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(pullPending.InterfaceName, "npc_trade_action", "pull task interface mismatch"); err != nil {
		return err
	}
	if err := execWorkerCommand(workerExe, []string{"pull-once", "--engine-base-url", engineBaseURL, "--runtime-task-token", runtimeTaskToken, "--callback-token", callbackToken, "--consumer", "bridge", "--lease-owner", "s6-pull-once"}); err != nil {
		return err
	}
	pullTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, pullCallbackID, sdk.RuntimeTaskStatusSucceeded, 20*time.Second)
	if err != nil {
		return err
	}
	collector.Add("pull", "delivery success", "worker", "passed", fmt.Sprintf("task_id=%s completed_at=%s", pullTask.TaskID, pullTask.CompletedAt))

	_ = workertest.StopProcess(pushWorkerProc)
	time.Sleep(time.Second)

	var hybridInvoke sdk.InvokeResponse
	if err := devcli.RunJSON(&hybridInvoke, "invoke", worldID, npcID, "--task-type", "custom", "--message", "hybrid runtime task"); err != nil {
		return err
	}
	hybridCallbackID := hybridInvoke.ActionCalls[0].CallbackID
	hybridTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, hybridCallbackID, sdk.RuntimeTaskStatusReleased, 20*time.Second)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(hybridTask.DeliveryMode, "hybrid", "hybrid task delivery_mode mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(hybridTask.Transport, "task_pull", "hybrid fallback transport mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(hybridTask.LastDispatchDecision, "fallback_to_pull", "hybrid dispatch decision mismatch"); err != nil {
		return err
	}
	collector.Add("hybrid", "fallback transition", "worker", "passed", fmt.Sprintf("task_id=%s failure_class=%s", hybridTask.TaskID, hybridTask.LastDispatchFailureClass))

	var manualInvoke sdk.InvokeResponse
	if err := devcli.RunJSON(&manualInvoke, "invoke", worldID, npcID, "--task-type", "custom", "--dynamic-interfaces-file", tradeFile, "--message", "manual claim task"); err != nil {
		return err
	}
	manualTaskPending, err := workertest.FindTaskByCallbackID(rtClient, worldID, 200, manualInvoke.ActionCalls[0].CallbackID)
	if err != nil {
		return err
	}
	if manualTaskPending == nil {
		return fmt.Errorf("manual task not found")
	}
	claimed, err := client.ClaimRuntimeTask(manualTaskPending.TaskID, "bridge", "s6-manual")
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(claimed.Status, sdk.RuntimeTaskStatusClaimed, "manual task claim status mismatch"); err != nil {
		return err
	}
	started, err := client.StartRuntimeTask(manualTaskPending.TaskID, claimed.LeaseToken)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(started.Status, sdk.RuntimeTaskStatusRunning, "manual task start status mismatch"); err != nil {
		return err
	}
	if err := client.HeartbeatRuntimeTask(manualTaskPending.TaskID, claimed.LeaseToken); err != nil {
		return err
	}
	heartbeatTask, err := client.GetRuntimeTask(manualTaskPending.TaskID)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(heartbeatTask.Status, sdk.RuntimeTaskStatusRunning, "manual task heartbeat status mismatch"); err != nil {
		return err
	}
	if err := releaseRuntimeTask(rtClient, manualTaskPending.TaskID, claimed.LeaseToken, 0, "manual release"); err != nil {
		return err
	}
	releasedTask, err := client.GetRuntimeTask(manualTaskPending.TaskID)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(releasedTask.Status, sdk.RuntimeTaskStatusReleased, "manual task release status mismatch"); err != nil {
		return err
	}
	collector.Add("pull", "claim/start/heartbeat/release", "sdk", "passed", "task_id="+manualTaskPending.TaskID)

	var timeoutInvoke sdk.InvokeResponse
	if err := devcli.RunJSON(&timeoutInvoke, "invoke", worldID, npcID, "--task-type", "custom", "--dynamic-interfaces-file", tradeFile, "--message", "manual timeout task"); err != nil {
		return err
	}
	timeoutTaskPending, err := workertest.FindTaskByCallbackID(rtClient, worldID, 200, timeoutInvoke.ActionCalls[0].CallbackID)
	if err != nil {
		return err
	}
	if timeoutTaskPending == nil {
		return fmt.Errorf("timeout task not found")
	}
	timeoutClaimed, err := client.ClaimRuntimeTask(timeoutTaskPending.TaskID, "bridge", "s6-timeout")
	if err != nil {
		return err
	}
	timeoutStarted, err := client.StartRuntimeTask(timeoutTaskPending.TaskID, timeoutClaimed.LeaseToken)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(timeoutStarted.Status, sdk.RuntimeTaskStatusRunning, "timeout task start status mismatch"); err != nil {
		return err
	}
	time.Sleep(3 * time.Second)
	var sweepResp struct { Affected int `json:"affected"` }
	if err := rtClient.RuntimeTaskJSON("POST", "/api/v1/runtime/tasks/heartbeat-timeout/sweep", map[string]any{"timeout_seconds": 1}, &sweepResp); err != nil {
		return err
	}
	if err := workertest.AssertTrue(sweepResp.Affected >= 1, "heartbeat timeout sweep affected no tasks"); err != nil {
		return err
	}
	timeoutTask, err := workertest.WaitTaskStatus(rtClient, worldID, 200, timeoutInvoke.ActionCalls[0].CallbackID, sdk.RuntimeTaskStatusHeartbeatTimeout, 20*time.Second)
	if err != nil {
		return err
	}
	requeued, err := client.RequeueRuntimeTask(timeoutTask.TaskID, 0, "manual requeue")
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(requeued.Status, sdk.RuntimeTaskStatusReleased, "timeout task requeue status mismatch"); err != nil {
		return err
	}
	collector.Add("pull", "heartbeat-timeout and requeue", "sdk", "passed", fmt.Sprintf("task_id=%s timeout_count=%d", timeoutTask.TaskID, timeoutTask.HeartbeatTimeoutCount))

	worldTasks, err := workertest.GetWorldTasks(rtClient, worldID, 200)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(worldTasks) >= 5, "world runtime task count should be at least 5"); err != nil {
		return err
	}
	stats, err := client.GetRuntimeTaskStats()
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(stats != nil, "task stats missing"); err != nil {
		return err
	}
	if _, ok := stats.ByDispatchDecision["fallback_to_pull"]; !ok {
		return fmt.Errorf("task stats output missing fallback_to_pull")
	}
	inspectTask, err := client.GetRuntimeTask(hybridTask.TaskID)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(inspectTask.LastDispatchDecision, "fallback_to_pull", "task inspect missing dispatch decision"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.TrimSpace(inspectTask.PayloadJSON) != "", "task inspect missing payload"); err != nil {
		return err
	}
	collector.Add("diagnostics", "list/stats/inspect", "sdk", "passed", fmt.Sprintf("inspect_task_id=%s world_task_count=%d", inspectTask.TaskID, len(worldTasks)))

	result := runtimeTasksResult{
		EnginePort:    enginePort,
		PushPort:      pushPort,
		ConfigPath:    files.ConfigPath,
		DBPath:        files.DBPath,
		WorldID:       worldID,
		PushTaskID:    pushTask.TaskID,
		PullTaskID:    pullTask.TaskID,
		HybridTaskID:  hybridTask.TaskID,
		TimeoutTaskID: timeoutTask.TaskID,
		Checks:        collector.Checks(),
	}
	return a.writeScenarioResult(result)
}

func firstNonEmptyValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func execWorkerCommand(workerExe string, args []string) error {
	cmd := exec.Command(workerExe, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("worker command failed: %s\n%s", strings.Join(args, " "), strings.TrimSpace(string(out)))
	}
	return nil
}

func releaseRuntimeTask(client *workertest.Client, taskID string, leaseToken string, retryDelayMs int, errorMessage string) error {
	if client == nil {
		return fmt.Errorf("runtime task client is required")
	}
	return client.RuntimeTaskJSON("POST", "/api/v1/runtime/tasks/release", map[string]any{
		"task_id":        taskID,
		"lease_token":    leaseToken,
		"retry_delay_ms": retryDelayMs,
		"error_message":  errorMessage,
	}, nil)
}
