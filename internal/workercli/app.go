package workercli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/pflag"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workerstate"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type Options struct {
	CommandName       string
	DisplayName       string
	ShortDescription  string
	DefaultLeaseOwner string
	WorkerID          string
	DeprecatedAlias   string
}

type workerConfig struct {
	EngineBaseURL       string
	EngineAPIKey        string
	RuntimeTaskToken    string
	CallbackToken       string
	GameHTTPBearerToken string
	TestsDir            string
	StateFile           string
	Consumer            string
	LeaseOwner          string
	TestEngineExePath   string
	TestDevCLIExePath   string
	TestWorkerExePath   string
	TestOutFile         string
	TestScenario        string
	PlayWorldID         string
	PlayPlayerNodeID    string
	PlaySessionID       string
	PlayPipelineMode    string
	PlayIncludeRelated  bool
	PlayAutoWorker      bool
	PushPort            int
	TestEnginePort      int
	TestPushPort        int
	PollInterval        time.Duration
	HeartbeatInterval   time.Duration
	CallbackDelay       time.Duration
	LongTaskDuration    time.Duration
	TestKeepTemp        bool
	TestJSON            bool
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

type processedTaskResult struct {
	Task     *sdk.RuntimeTask
	Callback *sdk.CallbackResponse
}

type app struct {
	options     Options
	cfg         workerConfig
	authorityMu sync.RWMutex
	authority   *workerstate.WorldState
}

func newApp(options Options) *app {
	if strings.TrimSpace(options.CommandName) == "" {
		options.CommandName = "gameagentworker"
	}
	if strings.TrimSpace(options.DisplayName) == "" {
		options.DisplayName = "GameAgentWorker"
	}
	if strings.TrimSpace(options.ShortDescription) == "" {
		options.ShortDescription = "Deterministic external worker for GameAgentEngine"
	}
	if strings.TrimSpace(options.DefaultLeaseOwner) == "" {
		options.DefaultLeaseOwner = options.CommandName
	}
	if strings.TrimSpace(options.WorkerID) == "" {
		options.WorkerID = options.CommandName
	}

	a := &app{
		options: options,
		cfg: workerConfig{
			EngineBaseURL:       "http://127.0.0.1:8080",
			EngineAPIKey:        "dev-key",
			RuntimeTaskToken:    "dev-task-token",
			CallbackToken:       "dev-callback-token",
			GameHTTPBearerToken: "local-test-token",
			TestsDir:            "tools/source/tests",
			Consumer:            "game_client",
			LeaseOwner:          options.DefaultLeaseOwner,
			PlayIncludeRelated:  true,
			PlayAutoWorker:      true,
			PushPort:            9000,
			TestEnginePort:      18080,
			TestPushPort:        19000,
			PollInterval:        2 * time.Second,
			HeartbeatInterval:   2 * time.Second,
			CallbackDelay:       250 * time.Millisecond,
			LongTaskDuration:    6 * time.Second,
		},
	}
	return a
}

func newTestApp() *app {
	return newApp(Options{
		CommandName:       "gameagentworker",
		DisplayName:       "GameAgentWorker",
		ShortDescription:  "test",
		DefaultLeaseOwner: "gameagentworker",
		WorkerID:          "gameagentworker",
	})
}

func (a *app) setAuthorityState(state *workerstate.WorldState) {
	a.authorityMu.Lock()
	defer a.authorityMu.Unlock()
	a.authority = workerstate.NewStateView(state).State()
}

func (a *app) authorityView() *workerstate.StateView {
	a.authorityMu.RLock()
	defer a.authorityMu.RUnlock()
	if a.authority == nil {
		return nil
	}
	return workerstate.NewStateView(a.authority)
}

func (a *app) bindCommonFlags(flags *pflag.FlagSet) {
	flags.StringVar(&a.cfg.EngineBaseURL, "engine-base-url", a.cfg.EngineBaseURL, "Engine base URL")
	flags.StringVar(&a.cfg.EngineAPIKey, "engine-api-key", a.cfg.EngineAPIKey, "Engine API key for invoke requests")
	flags.StringVar(&a.cfg.RuntimeTaskToken, "runtime-task-token", a.cfg.RuntimeTaskToken, "Runtime task token for /api/v1/runtime/tasks/*")
	flags.StringVar(&a.cfg.CallbackToken, "callback-token", a.cfg.CallbackToken, "Callback token for /api/v1/actions/callback")
	flags.StringVar(&a.cfg.GameHTTPBearerToken, "game-http-bearer-token", a.cfg.GameHTTPBearerToken, "Expected bearer token for push dispatch receiver")
	flags.StringVar(&a.cfg.StateFile, "state-file", a.cfg.StateFile, "Optional YAML/JSON authority state file for game-side query responses")
	flags.StringVar(&a.cfg.Consumer, "consumer", a.cfg.Consumer, "Runtime task consumer")
	flags.StringVar(&a.cfg.LeaseOwner, "lease-owner", a.cfg.LeaseOwner, "Runtime task lease owner")
	flags.IntVar(&a.cfg.PushPort, "push-port", a.cfg.PushPort, "Push receiver port")
	flags.DurationVar(&a.cfg.PollInterval, "poll-interval", a.cfg.PollInterval, "Pending task poll interval")
	flags.DurationVar(&a.cfg.HeartbeatInterval, "heartbeat-interval", a.cfg.HeartbeatInterval, "Heartbeat interval for long-running task simulation")
	flags.DurationVar(&a.cfg.CallbackDelay, "callback-delay", a.cfg.CallbackDelay, "Delay before callback completion")
	flags.DurationVar(&a.cfg.LongTaskDuration, "long-task-duration", a.cfg.LongTaskDuration, "Duration used when simulating long-running tasks")
	flags.StringSliceVar(&a.cfg.FailInterfaces, "fail-interface", nil, "Interface names that should callback with failed status")
	flags.StringSliceVar(&a.cfg.LongTaskInterfaces, "long-task-interface", nil, "Interface names that should simulate long-running execution")
	flags.BoolVar(&a.cfg.Verbose, "verbose", false, "Enable verbose structured worker logs")
}

func (a *app) bindPlayFlags(flags *pflag.FlagSet) {
	flags.StringVar(&a.cfg.PlayWorldID, "world-id", a.cfg.PlayWorldID, "World ID used in play mode; defaults to state file world_id")
	flags.StringVar(&a.cfg.PlayPlayerNodeID, "player-node-id", a.cfg.PlayPlayerNodeID, "Player node ID controlled in play mode")
	flags.StringVar(&a.cfg.PlaySessionID, "session-id", a.cfg.PlaySessionID, "Optional fixed session ID for play mode")
	flags.StringVar(&a.cfg.PlayPipelineMode, "pipeline-mode", a.cfg.PlayPipelineMode, "Optional invoke pipeline mode override for play mode")
	flags.BoolVar(&a.cfg.PlayIncludeRelated, "include-related-nodes", a.cfg.PlayIncludeRelated, "Whether play mode dialogue requests include related nodes")
	flags.BoolVar(&a.cfg.PlayAutoWorker, "auto-worker", a.cfg.PlayAutoWorker, "Run an embedded pull worker during play mode for authority queries")
}

func (a *app) runServe(withPush bool, withPull bool) error {
	if !withPush && !withPull {
		return errors.New("nothing to run")
	}
	errCh := make(chan error, 2)
	if withPush {
		go func() { errCh <- a.runPushReceiver() }()
	}
	if withPull {
		go func() { errCh <- a.runPullLoop() }()
	}
	return <-errCh
}

func (a *app) runPushReceiver() error {
	server := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", a.cfg.PushPort),
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/api/v1/runtime/dispatch" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
				return
			}
			if want := "Bearer " + strings.TrimSpace(a.cfg.GameHTTPBearerToken); strings.TrimSpace(r.Header.Get("Authorization")) != want {
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
			decision := a.decideExecution(req.InterfaceName, req.Payload)
			a.logJSON("push_received", map[string]any{
				"task_id":        req.TaskID,
				"interface_name": req.InterfaceName,
				"callback_id":    req.CallbackID,
				"request_id":     req.RequestID,
				"decision":       decision.DecisionName,
			})
			if req.CallbackID != "" {
				go a.completeAfterDecision(req.TaskID, req.CallbackID, decision)
			}
			a.writeJSON(w, http.StatusOK, map[string]any{"status": 200, "accepted": true, "worker": a.options.WorkerID})
		}),
	}
	a.logJSON("push_listen", map[string]any{"addr": server.Addr})
	return server.ListenAndServe()
}

func (a *app) runPullLoop() error {
	a.logJSON("pull_loop_started", map[string]any{"consumer": a.cfg.Consumer, "poll_interval_ms": a.cfg.PollInterval.Milliseconds()})
	for {
		_, _, err := a.processOnePendingTask()
		if err != nil {
			a.logJSON("pull_loop_error", map[string]any{"error": err.Error()})
		}
		time.Sleep(a.cfg.PollInterval)
	}
}

func (a *app) processOnePendingTask() (*sdk.RuntimeTask, bool, error) {
	result, processed, err := a.processOnePendingTaskDetailed()
	if result == nil {
		return nil, processed, err
	}
	return result.Task, processed, err
}

func (a *app) processOnePendingTaskDetailed() (*processedTaskResult, bool, error) {
	pending, err := a.runtimeTaskRequest(http.MethodGet, fmt.Sprintf("/api/v1/runtime/tasks/pending?consumer=%s&limit=1", a.cfg.Consumer), nil)
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
	a.logJSON("pull_claiming", map[string]any{"task_id": task.TaskID, "interface_name": task.InterfaceName, "status": task.Status})
	claimed, err := a.claimTask(task.TaskID)
	if err != nil {
		return nil, false, err
	}
	started, err := a.startTask(claimed.TaskID, claimed.LeaseToken)
	if err != nil {
		return &processedTaskResult{Task: claimed}, true, err
	}
	decision := a.decideExecution(started.InterfaceName, parseRuntimeTaskPayload(started.PayloadJSON))
	if decision.LongRunning {
		if err := a.runLongTask(started, decision); err != nil {
			return &processedTaskResult{Task: started}, true, err
		}
	} else if decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}
	resp, err := a.postCallback(started.CallbackID, decision.Status, decision.Result, callbackRequestID(started.TaskID))
	if err != nil {
		return &processedTaskResult{Task: started}, true, err
	}
	a.logJSON("pull_callback_completed", map[string]any{
		"task_id":              started.TaskID,
		"callback_id":          started.CallbackID,
		"status":               decision.Status,
		"resume_execution_id":  resp.ResumeExecutionID,
		"resumed":              resp.Resumed != nil,
		"post_process_applied": resp.PostProcess != nil && resp.PostProcess.Applied,
	})
	return &processedTaskResult{Task: started, Callback: resp}, true, nil
}

func (a *app) claimTask(taskID string) (*sdk.RuntimeTask, error) {
	body := map[string]any{"task_id": taskID, "consumer": a.cfg.Consumer, "lease_owner": a.cfg.LeaseOwner}
	data, err := a.runtimeTaskRequest(http.MethodPost, "/api/v1/runtime/tasks/claim", body)
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

func (a *app) startTask(taskID, leaseToken string) (*sdk.RuntimeTask, error) {
	body := map[string]any{"task_id": taskID, "lease_token": leaseToken}
	data, err := a.runtimeTaskRequest(http.MethodPost, "/api/v1/runtime/tasks/start", body)
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

func (a *app) heartbeatTask(taskID, leaseToken string) error {
	body := map[string]any{"task_id": taskID, "lease_token": leaseToken}
	_, err := a.runtimeTaskRequest(http.MethodPost, "/api/v1/runtime/tasks/heartbeat", body)
	return err
}

func (a *app) runLongTask(task *sdk.RuntimeTask, decision taskExecutionDecision) error {
	if task == nil {
		return errors.New("long task missing runtime task")
	}
	var beats atomic.Int64
	deadline := time.Now().Add(a.cfg.LongTaskDuration)
	for time.Now().Before(deadline) {
		if err := a.heartbeatTask(task.TaskID, task.LeaseToken); err != nil {
			return err
		}
		beats.Add(1)
		a.logJSON("pull_heartbeat", map[string]any{"task_id": task.TaskID, "count": beats.Load()})
		time.Sleep(a.cfg.HeartbeatInterval)
	}
	return nil
}

func (a *app) completeAfterDecision(taskID, callbackID string, decision taskExecutionDecision) {
	if decision.LongRunning {
		time.Sleep(a.cfg.LongTaskDuration)
	} else if decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}
	resp, err := a.postCallback(callbackID, decision.Status, decision.Result, callbackRequestID(taskID))
	if err != nil {
		a.logJSON("push_callback_error", map[string]any{"task_id": taskID, "callback_id": callbackID, "error": err.Error()})
		return
	}
	a.logJSON("push_callback_completed", map[string]any{
		"task_id":              taskID,
		"callback_id":          callbackID,
		"status":               decision.Status,
		"resume_execution_id":  resp.ResumeExecutionID,
		"resumed":              resp.Resumed != nil,
		"post_process_applied": resp.PostProcess != nil && resp.PostProcess.Applied,
	})
}

func (a *app) postCallback(callbackID, status string, result map[string]any, requestID string) (*sdk.CallbackResponse, error) {
	body := map[string]any{"callback_id": callbackID, "status": status, "result": result}
	data, err := a.callbackRequest(requestID, body)
	if err != nil {
		return nil, err
	}
	var resp sdk.CallbackResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (a *app) decideExecution(interfaceName string, payload map[string]any) taskExecutionDecision {
	name := strings.TrimSpace(interfaceName)
	status := "success"
	decisionName := "success"
	for _, item := range a.cfg.FailInterfaces {
		if strings.EqualFold(strings.TrimSpace(item), name) {
			status = "failed"
			decisionName = "forced_failure"
			break
		}
	}
	longRunning := false
	for _, item := range a.cfg.LongTaskInterfaces {
		if strings.EqualFold(strings.TrimSpace(item), name) {
			longRunning = true
			decisionName = "long_running"
			break
		}
	}
	result := a.buildFixtureResult(name, payload, status, longRunning)
	return taskExecutionDecision{
		Status:       status,
		Result:       result,
		LongRunning:  longRunning,
		Delay:        a.cfg.CallbackDelay,
		DecisionName: decisionName,
	}
}

func (a *app) buildFixtureResult(interfaceName string, payload map[string]any, status string, longRunning bool) map[string]any {
	result := map[string]any{
		"worker":         a.options.WorkerID,
		"interface_name": interfaceName,
		"status":         status,
		"long_running":   longRunning,
		"source":         "fixture",
	}
	switch interfaceName {
	case "game_client_request_data":
		if resolved := a.resolveAuthorityQueryResult(payload, status, longRunning); resolved != nil {
			for key, value := range resolved {
				result[key] = value
			}
		} else {
			result["scene"] = "starter_inn"
			result["world_state"] = map[string]any{"weather": "clear", "threat_level": "low"}
			result["echoed_payload"] = payload
		}
	case "spawn_item":
		result["spawned"] = true
		result["item_id"] = "fixture-item-1"
		result["inventory_target"] = firstString(payload, "node_id", "target_node_id")
	default:
		result["echoed_payload"] = payload
	}
	return result
}

func (a *app) resolveAuthorityQueryResult(payload map[string]any, status string, longRunning bool) map[string]any {
	view, err := a.loadAuthorityView()
	if err != nil {
		return map[string]any{
			"status":       status,
			"long_running": longRunning,
			"state_error":  err.Error(),
		}
	}
	if view == nil {
		return nil
	}
	queries := extractAuthorityQueries(payload)
	if len(queries) == 0 {
		return map[string]any{
			"status":       status,
			"long_running": longRunning,
			"world_id":     view.WorldID(),
		}
	}
	resolved := map[string]any{
		"status":       status,
		"long_running": longRunning,
		"world_id":     view.WorldID(),
		"queries":      resolveAuthorityQueries(view, queries),
	}
	return resolved
}

func (a *app) loadAuthorityView() (*workerstate.StateView, error) {
	if view := a.authorityView(); view != nil {
		return view, nil
	}
	if strings.TrimSpace(a.cfg.StateFile) == "" {
		return nil, nil
	}
	state, err := workerstate.LoadFile(a.cfg.StateFile)
	if err != nil {
		a.logJSON("state_file_load_error", map[string]any{"path": a.cfg.StateFile, "error": err.Error()})
		return nil, err
	}
	a.setAuthorityState(state)
	return a.authorityView(), nil
}

type authorityQuery struct {
	Type   string
	NodeID string
	Filter string
	Limit  int
}

func extractAuthorityQueries(payload map[string]any) []authorityQuery {
	if payload == nil {
		return nil
	}
	requestData, ok := payload["request_data"].(map[string]any)
	if !ok || requestData == nil {
		requestData = payload
	}
	rawQueries, ok := requestData["queries"].([]any)
	if !ok {
		return nil
	}
	queries := make([]authorityQuery, 0, len(rawQueries))
	for _, raw := range rawQueries {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		query := authorityQuery{
			Type:   firstString(item, "type"),
			NodeID: firstString(item, "node_id"),
			Filter: firstString(item, "filter"),
		}
		if limit, ok := item["limit"].(float64); ok {
			query.Limit = int(limit)
		}
		if strings.TrimSpace(query.Type) != "" {
			queries = append(queries, query)
		}
	}
	return queries
}

func resolveAuthorityQueries(view *workerstate.StateView, queries []authorityQuery) []map[string]any {
	results := make([]map[string]any, 0, len(queries))
	for _, query := range queries {
		result := map[string]any{"type": query.Type, "node_id": query.NodeID}
		switch query.Type {
		case "player_state":
			hp, maxHP, ok := view.ActorHP(query.NodeID)
			if ok {
				result["hp"] = hp
				result["max_hp"] = maxHP
			}
		case "player_inventory":
			result["inventory"] = view.ActorInventory(query.NodeID)
		case "player_wallet":
			if money, ok := view.ActorMoney(query.NodeID); ok {
				result["money"] = money
			}
		case "player_location", "npc_location":
			if locationID, ok := view.ActorLocation(query.NodeID); ok {
				result["location_id"] = locationID
			}
		case "scene_state", "room_state":
			if scene, ok := view.Scene(query.NodeID); ok {
				result["scene"] = scene
			}
		case "task_state":
			if status, stage, ok := view.QuestStatus(query.NodeID); ok {
				result["status"] = status
				result["stage"] = stage
			}
		case "item_presence":
			result["item_id"] = query.Filter
			result["present"] = view.ItemPresentOnActor(query.NodeID, query.Filter)
		default:
			result["unsupported"] = true
		}
		results = append(results, result)
	}
	return results
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

func (a *app) runtimeTaskRequest(method, path string, body any) ([]byte, error) {
	return a.doJSONRequest(method, a.cfg.EngineBaseURL+path, map[string]string{
		"X-Runtime-Task-Token": a.cfg.RuntimeTaskToken,
	}, body)
}

func (a *app) callbackRequest(requestID string, body any) ([]byte, error) {
	return a.doJSONRequest(http.MethodPost, a.cfg.EngineBaseURL+"/api/v1/actions/callback", map[string]string{
		"X-Callback-Token":      a.cfg.CallbackToken,
		"X-Callback-Request-Id": requestID,
	}, body)
}

func (a *app) doJSONRequest(method, rawURL string, extraHeaders map[string]string, body any) ([]byte, error) {
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

func (a *app) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (a *app) logJSON(event string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["event"] = event
	fields["ts"] = time.Now().Format(time.RFC3339Nano)
	if !a.cfg.Verbose {
		delete(fields, "echoed_payload")
	}
	data, err := json.Marshal(fields)
	if err != nil {
		log.Printf("{\"event\":%q,\"error\":%q}", event, err.Error())
		return
	}
	log.Printf("%s", data)
}
