package workertest

import (
	"fmt"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func GetWorldTasks(client *Client, worldID string, limit int) ([]sdk.RuntimeTask, error) {
	var resp struct {
		Tasks []sdk.RuntimeTask `json:"tasks"`
	}
	path := QueryWithValues("/api/v1/runtime/tasks", map[string]string{
		"world_id": worldID,
		"limit":    fmt.Sprintf("%d", limit),
	})
	if err := client.RuntimeTaskJSON("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

func FindTaskByCallbackID(client *Client, worldID string, limit int, callbackID string) (*sdk.RuntimeTask, error) {
	tasks, err := GetWorldTasks(client, worldID, limit)
	if err != nil {
		return nil, err
	}
	for i := range tasks {
		if tasks[i].CallbackID == callbackID {
			return &tasks[i], nil
		}
	}
	return nil, nil
}

func WaitTaskStatus(client *Client, worldID string, limit int, callbackID string, expectedStatus string, timeout time.Duration) (*sdk.RuntimeTask, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		task, err := FindTaskByCallbackID(client, worldID, limit, callbackID)
		if err != nil {
			return nil, err
		}
		if task != nil && task.Status == expectedStatus {
			return task, nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	last, err := FindTaskByCallbackID(client, worldID, limit, callbackID)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("task for callback_id=%s did not reach status=%s; last=%+v", callbackID, expectedStatus, last)
}
