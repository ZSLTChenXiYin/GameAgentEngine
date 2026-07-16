package workercli

import (
	"fmt"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workertest"
)

type testAllScenarioResult struct {
	StartedAt string                   `json:"started_at"`
	FinishedAt string                  `json:"finished_at"`
	Scenarios []testAllScenarioEntry   `json:"scenarios"`
	Checks    []workertest.CheckResult `json:"checks"`
}

type testAllScenarioEntry struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (a *app) runAllTestScenarios() error {
	scenarios := []string{
		"base-data",
		"continuity",
		"runtime-tasks",
		"callback-resume",
		"tooling-smoke",
		"machine-scenario",
	}
	startedAt := time.Now()
	entries := make([]testAllScenarioEntry, 0, len(scenarios))
	collector := &workertest.Collector{}
	baseEnginePort := a.cfg.TestEnginePort
	basePushPort := a.cfg.TestPushPort

	for index, scenario := range scenarios {
		a.cfg.TestEnginePort = baseEnginePort + index
		a.cfg.TestPushPort = basePushPort + index
		err := a.runNamedTestScenario(scenario)
		if err != nil {
			entries = append(entries, testAllScenarioEntry{
				Name:   scenario,
				Status: "failed",
				Error:  err.Error(),
			})
			collector.Add("test-all", scenario, "worker", "failed", err.Error())
			a.cfg.TestEnginePort = baseEnginePort
			a.cfg.TestPushPort = basePushPort
			return a.writeScenarioResult(testAllScenarioResult{
				StartedAt:  startedAt.Format(time.RFC3339Nano),
				FinishedAt: time.Now().Format(time.RFC3339Nano),
				Scenarios:  entries,
				Checks:     collector.Checks(),
			})
		}
		entries = append(entries, testAllScenarioEntry{Name: scenario, Status: "passed"})
		collector.Add("test-all", scenario, "worker", "passed", fmt.Sprintf("engine_port=%d push_port=%d", a.cfg.TestEnginePort, a.cfg.TestPushPort))
	}

	a.cfg.TestEnginePort = baseEnginePort
	a.cfg.TestPushPort = basePushPort
	return a.writeScenarioResult(testAllScenarioResult{
		StartedAt:  startedAt.Format(time.RFC3339Nano),
		FinishedAt: time.Now().Format(time.RFC3339Nano),
		Scenarios:  entries,
		Checks:     collector.Checks(),
	})
}

func isImplementedScenario(name string) bool {
	switch strings.TrimSpace(name) {
	case "base-data", "continuity", "runtime-tasks", "callback-resume", "tooling-smoke", "machine-scenario", "all":
		return true
	default:
		return false
	}
}

