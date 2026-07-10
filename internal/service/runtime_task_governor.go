package service

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type RuntimeTaskGovernanceOptions struct {
	HeartbeatTimeout  time.Duration
	AutoRequeue       bool
	AutoRequeueLimit  int
	AutoRequeueDelay  time.Duration
	AutoRequeueReason string
}

type RuntimeTaskGovernanceResult struct {
	TimedOut      int64 `json:"timed_out"`
	Requeued      int64 `json:"requeued"`
	PolicySkipped int64 `json:"policy_skipped"`
}

type runtimeTaskHeartbeatTimeoutPolicy struct {
	AutoRequeue    *bool  `json:"auto_requeue,omitempty"`
	RequeueDelayMs int    `json:"requeue_delay_ms,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type runtimeTaskPayloadPolicy struct {
	HeartbeatTimeoutPolicy runtimeTaskHeartbeatTimeoutPolicy `json:"heartbeat_timeout_policy"`
}

type RuntimeTaskGovernor struct {
	interval time.Duration
	options  RuntimeTaskGovernanceOptions
}

func NewRuntimeTaskGovernor(interval time.Duration, options RuntimeTaskGovernanceOptions) *RuntimeTaskGovernor {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	if options.HeartbeatTimeout <= 0 {
		options.HeartbeatTimeout = 5 * time.Minute
	}
	if options.AutoRequeueLimit <= 0 {
		options.AutoRequeueLimit = 100
	}
	if options.AutoRequeueReason == "" {
		options.AutoRequeueReason = "auto requeue after heartbeat timeout"
	}
	return &RuntimeTaskGovernor{interval: interval, options: options}
}

func (g *RuntimeTaskGovernor) Start(ctx context.Context) {
	g.runOnce()
	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.runOnce()
		}
	}
}

func (g *RuntimeTaskGovernor) runOnce() {
	result, err := RunRuntimeTaskGovernance(g.options)
	if err != nil {
		log.Printf("[runtime-task:governor] error: %v", err)
		return
	}
	if result.TimedOut > 0 || result.Requeued > 0 {
		log.Printf("[runtime-task:governor] timed_out=%d requeued=%d policy_skipped=%d", result.TimedOut, result.Requeued, result.PolicySkipped)
	}
}

func RunRuntimeTaskGovernance(options RuntimeTaskGovernanceOptions) (*RuntimeTaskGovernanceResult, error) {
	if options.HeartbeatTimeout <= 0 {
		options.HeartbeatTimeout = 5 * time.Minute
	}
	if options.AutoRequeueLimit <= 0 {
		options.AutoRequeueLimit = 100
	}
	if options.AutoRequeueReason == "" {
		options.AutoRequeueReason = "auto requeue after heartbeat timeout"
	}
	result := &RuntimeTaskGovernanceResult{}
	timedOut, err := store.MarkRuntimeTasksHeartbeatTimeout(options.HeartbeatTimeout)
	if err != nil {
		return nil, err
	}
	result.TimedOut = timedOut
	if options.AutoRequeue {
		items, err := store.ListRuntimeTasks(store.RuntimeTaskListQuery{Statuses: []string{store.RuntimeTaskStatusHeartbeatTimeout}, Limit: options.AutoRequeueLimit})
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			autoRequeue, delay, reason := resolveHeartbeatTimeoutPolicy(&item, options)
			if !autoRequeue {
				result.PolicySkipped++
				continue
			}
			if _, err := store.RequeueHeartbeatTimeoutTask(item.TaskID, delay, reason); err != nil {
				return nil, err
			}
			result.Requeued++
		}
	}
	return result, nil
}

func resolveHeartbeatTimeoutPolicy(task *store.RuntimeTaskModel, options RuntimeTaskGovernanceOptions) (bool, time.Duration, string) {
	auto := options.AutoRequeue
	delay := options.AutoRequeueDelay
	reason := options.AutoRequeueReason
	if task == nil || strings.TrimSpace(task.PayloadJSON) == "" {
		return auto, delay, reason
	}
	var payload runtimeTaskPayloadPolicy
	if err := json.Unmarshal([]byte(task.PayloadJSON), &payload); err != nil {
		return auto, delay, reason
	}
	if payload.HeartbeatTimeoutPolicy.AutoRequeue != nil {
		auto = *payload.HeartbeatTimeoutPolicy.AutoRequeue
	}
	if payload.HeartbeatTimeoutPolicy.RequeueDelayMs > 0 {
		delay = time.Duration(payload.HeartbeatTimeoutPolicy.RequeueDelayMs) * time.Millisecond
	}
	if strings.TrimSpace(payload.HeartbeatTimeoutPolicy.Reason) != "" {
		reason = strings.TrimSpace(payload.HeartbeatTimeoutPolicy.Reason)
	}
	return auto, delay, reason
}
