package service

import (
	"context"
	"log"
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
	TimedOut int64 `json:"timed_out"`
	Requeued int64 `json:"requeued"`
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
		log.Printf("[runtime-task:governor] timed_out=%d requeued=%d", result.TimedOut, result.Requeued)
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
		requeued, err := store.RequeueHeartbeatTimeoutTasksBatch("", "", "", options.AutoRequeueDelay, options.AutoRequeueReason, options.AutoRequeueLimit)
		if err != nil {
			return nil, err
		}
		result.Requeued = requeued
	}
	return result, nil
}
