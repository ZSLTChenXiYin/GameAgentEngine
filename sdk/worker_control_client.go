package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// WorkerControlClient wraps runtime-task and callback endpoints that are
// authenticated with worker-side control tokens rather than the public API key.
type WorkerControlClient struct {
	baseURL          string
	runtimeTaskToken string
	callbackToken    string
	hc               *http.Client
}

// NewWorkerControlClient creates a worker-side control client.
func NewWorkerControlClient(baseURL, runtimeTaskToken, callbackToken string) *WorkerControlClient {
	return &WorkerControlClient{
		baseURL:          baseURL,
		runtimeTaskToken: runtimeTaskToken,
		callbackToken:    callbackToken,
		hc:               &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *WorkerControlClient) doJSON(method, path string, body any, headers map[string]string, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		if strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s %s failed: %d %s", method, c.baseURL+path, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// ListPendingRuntimeTasks returns pull-ready runtime tasks for one consumer.
func (c *WorkerControlClient) ListPendingRuntimeTasks(consumer string, limit int) ([]RuntimeTask, error) {
	q := url.Values{}
	if consumer != "" {
		q.Set("consumer", consumer)
	}
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	path := "/api/v1/runtime/tasks/pending"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var resp struct {
		Tasks []RuntimeTask `json:"tasks"`
	}
	err := c.doJSON(http.MethodGet, path, nil, map[string]string{
		"X-Runtime-Task-Token": c.runtimeTaskToken,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

// ClaimRuntimeTask claims a pending runtime task for this consumer.
func (c *WorkerControlClient) ClaimRuntimeTask(taskID, consumer, leaseOwner string) (*RuntimeTask, error) {
	var resp struct {
		Task *RuntimeTask `json:"task"`
	}
	err := c.doJSON(http.MethodPost, "/api/v1/runtime/tasks/claim", map[string]any{
		"task_id":     taskID,
		"consumer":    consumer,
		"lease_owner": leaseOwner,
	}, map[string]string{
		"X-Runtime-Task-Token": c.runtimeTaskToken,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Task, nil
}

// StartRuntimeTask marks a claimed runtime task as running.
func (c *WorkerControlClient) StartRuntimeTask(taskID, leaseToken string) (*RuntimeTask, error) {
	var resp struct {
		Task *RuntimeTask `json:"task"`
	}
	err := c.doJSON(http.MethodPost, "/api/v1/runtime/tasks/start", map[string]any{
		"task_id":     taskID,
		"lease_token": leaseToken,
	}, map[string]string{
		"X-Runtime-Task-Token": c.runtimeTaskToken,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Task, nil
}

// HeartbeatRuntimeTask sends a heartbeat for a running runtime task.
func (c *WorkerControlClient) HeartbeatRuntimeTask(taskID, leaseToken string) error {
	return c.doJSON(http.MethodPost, "/api/v1/runtime/tasks/heartbeat", map[string]any{
		"task_id":     taskID,
		"lease_token": leaseToken,
	}, map[string]string{
		"X-Runtime-Task-Token": c.runtimeTaskToken,
	}, nil)
}

// ActionCallback reports a callback result and returns structured resume data.
func (c *WorkerControlClient) ActionCallback(callbackID, status string, result any, requestID string) (*CallbackResponse, error) {
	var resp CallbackResponse
	headers := map[string]string{
		"X-Callback-Token": c.callbackToken,
	}
	if strings.TrimSpace(requestID) != "" {
		headers["X-Callback-Request-Id"] = requestID
	}
	err := c.doJSON(http.MethodPost, "/api/v1/actions/callback", map[string]any{
		"callback_id": callbackID,
		"status":      status,
		"result":      result,
	}, headers, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
