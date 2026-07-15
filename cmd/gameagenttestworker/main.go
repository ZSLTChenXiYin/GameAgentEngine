package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type workerConfig struct {
	EngineBaseURL       string
	RuntimeTaskToken    string
	CallbackToken       string
	GameHTTPBearerToken string
	Consumer            string
	LeaseOwner          string
	PushPort            int
	PollInterval        time.Duration
	HeartbeatInterval   time.Duration
	CallbackDelay       time.Duration
	LongTaskDuration    time.Duration
	Verbose             bool
	FailInterfaces      []string
	LongTaskInterfaces  []string
}

type runtimeTaskPayload struct {
	TaskType          string         `json:"task_type"`
	WorldID           string         `json:"world_id"`
	NodeID            string         `json:"node_id"`
	RequestID         string         `json:"request_id"`
	CallbackID        string         `json:"callback_id"`
	ResumeExecutionID string         `json:"resume_execution_id"`
	ResumePolicy      string         `json:"resume_policy"`
	ExternalInterface string         `json:"external_interface"`
	RequestData       map[string]any `json:"request_data"`
	ActionID          string         `json:"action_id"`
	DeliveryMode      string         `json:"delivery_mode"`
	PrimaryTransport  string         `json:"primary_transport"`
	Consumer          string         `json:"consumer"`
}

type taskExecutionDecision struct {
	Status       string
	Result       map[string]any
	LongRunning  bool
	Delay        time.Duration
	DecisionName string
}

var cfg = workerConfig{
	EngineBaseURL:       "http://127.0.0.1:8080",
	RuntimeTaskToken:    "dev-task-token",
	CallbackToken:       "dev-callback-token",
	GameHTTPBearerToken: "local-test-token",
	Consumer:            "game_client",
	LeaseOwner:          "gameagenttestworker",
	PushPort:            9000,
	PollInterval:        2 * time.Second,
	HeartbeatInterval:   2 * time.Second,
	CallbackDelay:       250 * time.Millisecond,
	LongTaskDuration:    6 * time.Second,
}

var rootCmd = &cobra.Command{
	Use:   "GameAgentTestWorker",
	Short: "Deterministic external test worker for GameAgentEngine functional scenarios",
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run both push receiver and pull worker loops",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe(true, true)
	},
}

var pushCmd = &cobra.Command{
	Use:   "push-receiver",
	Short: "Run only the push receiver",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe(true, false)
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull-worker",
	Short: "Run only the pull worker loop",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe(false, true)
	},
}

var pullOnceCmd = &cobra.Command{
	Use:   "pull-once",
	Short: "Claim, execute, and callback one pull task if present",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, processed, err := processOnePendingTask()
		if err != nil {
			return err
		}
		if !processed {
			logJSON("pull_noop", map[string]any{"consumer": cfg.Consumer})
		}
		return nil
	},
}

func main() {
	bindCommonFlags(rootCmd.PersistentFlags())
	rootCmd.AddCommand(serveCmd, pushCmd, pullCmd, pullOnceCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func bindCommonFlags(flags *pflag.FlagSet) {
	flags.StringVar(&cfg.EngineBaseURL, "engine-base-url", cfg.EngineBaseURL, "Engine base URL")
	flags.StringVar(&cfg.RuntimeTaskToken, "runtime-task-token", cfg.RuntimeTaskToken, "Runtime task token for /api/v1/runtime/tasks/*")
	flags.StringVar(&cfg.CallbackToken, "callback-token", cfg.CallbackToken, "Callback token for /api/v1/actions/callback")
	flags.StringVar(&cfg.GameHTTPBearerToken, "game-http-bearer-token", cfg.GameHTTPBearerToken, "Expected bearer token for push dispatch receiver")
	flags.StringVar(&cfg.Consumer, "consumer", cfg.Consumer, "Runtime task consumer")
	flags.StringVar(&cfg.LeaseOwner, "lease-owner", cfg.LeaseOwner, "Runtime task lease owner")
	flags.IntVar(&cfg.PushPort, "push-port", cfg.PushPort, "Push receiver port")
	flags.DurationVar(&cfg.PollInterval, "poll-interval", cfg.PollInterval, "Pending task poll interval")
	flags.DurationVar(&cfg.HeartbeatInterval, "heartbeat-interval", cfg.HeartbeatInterval, "Heartbeat interval for long-running task simulation")
	flags.DurationVar(&cfg.CallbackDelay, "callback-delay", cfg.CallbackDelay, "Delay before callback completion")
	flags.DurationVar(&cfg.LongTaskDuration, "long-task-duration", cfg.LongTaskDuration, "Duration used when simulating long-running tasks")
	flags.StringSliceVar(&cfg.FailInterfaces, "fail-interface", nil, "Interface names that should callback with failed status")
	flags.StringSliceVar(&cfg.LongTaskInterfaces, "long-task-interface", nil, "Interface names that should simulate long-running execution")
	flags.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose structured worker logs")
}

func runServe(withPush bool, withPull bool) error {
	if !withPush && !withPull {
		return errors.New("nothing to run")
	}
	errCh := make(chan error, 2)
	if withPush {
		go func() { errCh <- runPushReceiver() }()
	}
	if withPull {
		go func() { errCh <- runPullLoop() }()
	}
	return <-errCh
}

func runPushReceiver() error {
	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", cfg.PushPort),
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/api/v1/runtime/dispatch" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
				return
			}
			if want := "Bearer " + strings.TrimSpace(cfg.GameHTTPBearerToken); strings.TrimSpace(r.Header.Get("Authorization")) != want {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("unauthorized"))
				return
			}
			var req struct {
				TaskID        string         `json:"task_id"`
				InterfaceName string         `json:"interface_name"`
				CallbackID    string         `json:"callback_id"`
				RequestID     string         `json:"request_id"`
				Payload       map[string]any `json:"payload"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			decision := decideExecution(req.InterfaceName, req.Payload)
			logJSON("push_received", map[string]any{
				"task_id":        req.TaskID,
				"interface_name": req.InterfaceName,
				"callback_id":    req.CallbackID,
				"request_id":     req.RequestID,
				"decision":       decision.DecisionName,
			})
			if req.CallbackID != "" {
				go completeAfterDecision(req.TaskID, req.CallbackID, decision)
			}
			writeJSON(w, http.StatusOK, map[string]any{"status": 200, "accepted": true, "worker": "gameagenttestworker"})
		}),
	}
	logJSON("push_listen", map[string]any{"addr": server.Addr})
	return server.ListenAndServe()
}

func runPullLoop() error {
	logJSON("pull_loop_started", map[string]any{"consumer": cfg.Consumer, "poll_interval_ms": cfg.PollInterval.Milliseconds()})
	for {
		_, _, err := processOnePendingTask()
		if err != nil {
			logJSON("pull_loop_error", map[string]any{"error": err.Error()})
		}
		time.Sleep(cfg.PollInterval)
	}
}

func processOnePendingTask() (*sdk.RuntimeTask, bool, error) {
	pending, err := runtimeTaskRequest(http.MethodGet, fmt.Sprintf("/api/v1/runtime/tasks/pending?consumer=%s&limit=1", cfg.Consumer), nil)
	if err != nil {
		return nil, false, err
	}
	var list struct {
		Tasks []sdk.RuntimeTask `json:"tasks"`
	}
	if err := json.Unmarshal(pending, &list); err != nil {
		return nil, false, err
	}
	if len(list.Tasks) == 0 {
		return nil, false, nil
	}
	task := list.Tasks[0]
	logJSON("pull_claiming", map[string]any{"task_id": task.TaskID, "interface_name": task.InterfaceName, "status": task.Status})
	claimed, err := claimTask(task.TaskID)
	if err != nil {
		return nil, false, err
	}
	started, err := startTask(claimed.TaskID, claimed.LeaseToken)
	if err != nil {
		return claimed, true, err
	}
	decision := decideExecution(started.InterfaceName, parseRuntimeTaskPayload(started.PayloadJSON))
	if decision.LongRunning {
		if err := runLongTask(started, decision); err != nil {
			return started, true, err
		}
	} else if decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}
	resp, err := postCallback(started.CallbackID, decision.Status, decision.Result, callbackRequestID(started.TaskID))
	if err != nil {
		return started, true, err
	}
	logJSON("pull_callback_completed", map[string]any{
		"task_id":              started.TaskID,
		"callback_id":          started.CallbackID,
		"status":               decision.Status,
		"resume_execution_id":  resp.ResumeExecutionID,
		"resumed":              resp.Resumed != nil,
		"post_process_applied": resp.PostProcess != nil && resp.PostProcess.Applied,
	})
	return started, true, nil
}

func claimTask(taskID string) (*sdk.RuntimeTask, error) {
	body := map[string]any{"task_id": taskID, "consumer": cfg.Consumer, "lease_owner": cfg.LeaseOwner}
	data, err := runtimeTaskRequest(http.MethodPost, "/api/v1/runtime/tasks/claim", body)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Task *sdk.RuntimeTask `json:"task"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Task, nil
}

func startTask(taskID, leaseToken string) (*sdk.RuntimeTask, error) {
	body := map[string]any{"task_id": taskID, "lease_token": leaseToken}
	data, err := runtimeTaskRequest(http.MethodPost, "/api/v1/runtime/tasks/start", body)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Task *sdk.RuntimeTask `json:"task"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Task, nil
}

func heartbeatTask(taskID, leaseToken string) error {
	body := map[string]any{"task_id": taskID, "lease_token": leaseToken}
	_, err := runtimeTaskRequest(http.MethodPost, "/api/v1/runtime/tasks/heartbeat", body)
	return err
}

func runLongTask(task *sdk.RuntimeTask, decision taskExecutionDecision) error {
	if task == nil {
		return errors.New("long task missing runtime task")
	}
	var beats atomic.Int64
	deadline := time.Now().Add(cfg.LongTaskDuration)
	for time.Now().Before(deadline) {
		if err := heartbeatTask(task.TaskID, task.LeaseToken); err != nil {
			return err
		}
		beats.Add(1)
		logJSON("pull_heartbeat", map[string]any{"task_id": task.TaskID, "count": beats.Load()})
		time.Sleep(cfg.HeartbeatInterval)
	}
	return nil
}

func completeAfterDecision(taskID, callbackID string, decision taskExecutionDecision) {
	if decision.LongRunning {
		time.Sleep(cfg.LongTaskDuration)
	} else if decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}
	resp, err := postCallback(callbackID, decision.Status, decision.Result, callbackRequestID(taskID))
	if err != nil {
		logJSON("push_callback_error", map[string]any{"task_id": taskID, "callback_id": callbackID, "error": err.Error()})
		return
	}
	logJSON("push_callback_completed", map[string]any{
		"task_id":              taskID,
		"callback_id":          callbackID,
		"status":               decision.Status,
		"resume_execution_id":  resp.ResumeExecutionID,
		"resumed":              resp.Resumed != nil,
		"post_process_applied": resp.PostProcess != nil && resp.PostProcess.Applied,
	})
}

func postCallback(callbackID, status string, result map[string]any, requestID string) (*sdk.CallbackResponse, error) {
	body := map[string]any{"callback_id": callbackID, "status": status, "result": result}
	data, err := callbackRequest(requestID, body)
	if err != nil {
		return nil, err
	}
	var resp sdk.CallbackResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func decideExecution(interfaceName string, payload map[string]any) taskExecutionDecision {
	name := strings.TrimSpace(interfaceName)
	status := "success"
	decisionName := "success"
	for _, item := range cfg.FailInterfaces {
		if strings.EqualFold(strings.TrimSpace(item), name) {
			status = "failed"
			decisionName = "forced_failure"
			break
		}
	}
	longRunning := false
	for _, item := range cfg.LongTaskInterfaces {
		if strings.EqualFold(strings.TrimSpace(item), name) {
			longRunning = true
			decisionName = "long_running"
			break
		}
	}
	result := buildFixtureResult(name, payload, status, longRunning)
	return taskExecutionDecision{
		Status:       status,
		Result:       result,
		LongRunning:  longRunning,
		Delay:        cfg.CallbackDelay,
		DecisionName: decisionName,
	}
}

func buildFixtureResult(interfaceName string, payload map[string]any, status string, longRunning bool) map[string]any {
	result := map[string]any{
		"worker":         "gameagenttestworker",
		"interface_name": interfaceName,
		"status":         status,
		"long_running":   longRunning,
		"source":         "fixture",
	}
	switch interfaceName {
	case "game_client_request_data":
		result["scene"] = "starter_inn"
		result["world_state"] = map[string]any{"weather": "clear", "threat_level": "low"}
		result["echoed_payload"] = payload
	case "spawn_item":
		result["spawned"] = true
		result["item_id"] = "fixture-item-1"
		result["inventory_target"] = firstString(payload, "node_id", "target_node_id")
	default:
		result["echoed_payload"] = payload
	}
	return result
}

func parseRuntimeTaskPayload(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return map[string]any{"raw_payload_json": raw}
	}
	return payload
}

func firstString(payload map[string]any, keys ...string) string {
	if payload == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func callbackRequestID(taskID string) string {
	return fmt.Sprintf("cb-%s", strings.TrimSpace(taskID))
}

func runtimeTaskRequest(method, path string, body any) ([]byte, error) {
	return doJSONRequest(method, cfg.EngineBaseURL+path, map[string]string{
		"X-Runtime-Task-Token": cfg.RuntimeTaskToken,
	}, body)
}

func callbackRequest(requestID string, body any) ([]byte, error) {
	return doJSONRequest(http.MethodPost, cfg.EngineBaseURL+"/api/v1/actions/callback", map[string]string{
		"X-Callback-Token":      cfg.CallbackToken,
		"X-Callback-Request-Id": requestID,
	}, body)
}

func doJSONRequest(method, rawURL string, extraHeaders map[string]string, body any) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range extraHeaders {
		if strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("%s %s failed: %d %s", method, rawURL, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	return data, nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func logJSON(event string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["event"] = event
	fields["ts"] = time.Now().Format(time.RFC3339Nano)
	if !cfg.Verbose {
		delete(fields, "echoed_payload")
	}
	data, err := json.Marshal(fields)
	if err != nil {
		log.Printf("{\"event\":%q,\"error\":%q}", event, err.Error())
		return
	}
	log.Printf("%s", data)
}
