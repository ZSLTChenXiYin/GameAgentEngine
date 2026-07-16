package workercli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workertest"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type continuityResult struct {
	WorldID          string                   `json:"world_id"`
	RequestID        string                   `json:"request_id"`
	LatestTickNumber int                      `json:"latest_tick_number"`
	LatestTimeLabel  string                   `json:"latest_time_label"`
	Checks           []workertest.CheckResult `json:"checks"`
}

func (a *app) runContinuityScenario() error {
	baseData, err := a.executeBaseDataScenario()
	if err != nil {
		return err
	}
	if strings.TrimSpace(a.cfg.TestDevCLIExePath) == "" {
		return fmt.Errorf("continuity requires --devcli-exe")
	}
	testsDir := strings.TrimSpace(a.cfg.TestsDir)
	if testsDir == "" {
		return fmt.Errorf("continuity requires --tests-dir")
	}
	paths := map[string]string{
		"world_time_settings": filepath.Join(testsDir, "world_time_settings_flexible.json"),
		"world_state":         filepath.Join(testsDir, "state_world_state.json"),
		"story_state":         filepath.Join(testsDir, "state_story_state.json"),
		"story_history":       filepath.Join(testsDir, "state_story_history.json"),
		"tick_policy":         filepath.Join(testsDir, "state_tick_policy.json"),
	}
	for name, path := range paths {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("continuity fixture %s not found: %w", name, err)
		}
	}
	engineBaseURL := strings.TrimSpace(a.cfg.EngineBaseURL)
	if engineBaseURL == "" {
		engineBaseURL = fmt.Sprintf("http://127.0.0.1:%d", a.cfg.TestEnginePort)
	}
	apiKey := strings.TrimSpace(a.cfg.EngineAPIKey)
	if apiKey == "" {
		apiKey = "dev-key"
	}
	devcli := workertest.DevCLI{Executable: a.cfg.TestDevCLIExePath, Server: engineBaseURL, APIKey: apiKey}
	client := sdk.NewClient(engineBaseURL, apiKey)
	collector := &workertest.Collector{}
	worldID := baseData.WorldID

	var settings sdk.WorldSettings
	if err := devcli.RunJSON(&settings, "world", "settings", "set", worldID, "--world-time-settings-file", paths["world_time_settings"], "--pipeline-mode", "full"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(settings.PipelineMode, "full", "world settings pipeline mode mismatch"); err != nil {
		return err
	}
	if settings.WorldTimeSettings == nil {
		return fmt.Errorf("world time settings missing after update")
	}
	if err := workertest.AssertEqual(settings.WorldTimeSettings.TickScaleMode, "flexible", "world time settings tick_scale_mode mismatch"); err != nil {
		return err
	}
	collector.Add("world_time_settings", "set", "devcli", "passed", "tick_scale_mode="+settings.WorldTimeSettings.TickScaleMode)

	for componentType, filePath := range map[string]string{
		"world_state":   paths["world_state"],
		"story_state":   paths["story_state"],
		"story_history": paths["story_history"],
		"tick_policy":   paths["tick_policy"],
	} {
		payload, err := readJSONFile(filePath)
		if err != nil {
			return err
		}
		resp, err := client.PutStateComponent(worldID, componentType, payload)
		if err != nil {
			return err
		}
		if err := workertest.AssertEqual(resp.StateComponent.ComponentType, componentType, componentType+" set failed"); err != nil {
			return err
		}
	}
	collector.Add("state", "seed continuity components", "sdk", "passed", "world_state/story_state/story_history/tick_policy")

	requestedTicks := 2
	autonomousLimit := 0
	tick, err := client.AdvanceTickWithOptions(worldID, "manual", "day-12 hour-9", &requestedTicks, &autonomousLimit)
	if err != nil {
		return err
	}
	if err := workertest.AssertEqual(tick.AdvancedTicks, 2, "tick advanced_ticks mismatch"); err != nil {
		return err
	}
	if tick.WorldTimeState == nil {
		return fmt.Errorf("tick world_time_state missing")
	}
	if err := workertest.AssertTrue(strings.TrimSpace(tick.WorldTimeState.CurrentTimeLabel) != "", "tick world_time_state label missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(tick.WorldTimeState.CurrentTimeLabel != "day-12 hour-9", "tick world_time_state label did not advance"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(tick.WorldTimeState.LastAdvancedTicks, 2, "tick last_advanced_ticks mismatch"); err != nil {
		return err
	}
	if tick.Invoke == nil {
		return fmt.Errorf("tick invoke payload missing")
	}
	requestID := tick.Invoke.RequestID
	if err := workertest.AssertTrue(strings.TrimSpace(requestID) != "", "tick request_id missing"); err != nil {
		return err
	}
	collector.Add("tick", "advance world tick", "sdk", "passed", fmt.Sprintf("request_id=%s advanced_ticks=%d", requestID, tick.AdvancedTicks))

	latest, err := client.GetLatestTimeline(worldID)
	if err != nil {
		return err
	}
	timelines, err := client.GetTimelines(worldID, 5)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(latest.Timeline.TickNumber >= 1, "latest timeline tick_number should be at least 1"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(latest.Timeline.AdvancedTicks, 2, "latest timeline advanced_ticks mismatch"); err != nil {
		return err
	}
	currentTimeLabel, _ := timelineWorldTimeLabel(latest.Timeline.Data)
	if err := workertest.AssertEqual(currentTimeLabel, tick.WorldTimeState.CurrentTimeLabel, "latest timeline world time mismatch"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(timelines.Timelines) >= 1, "timeline list should contain at least one entry"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(timelines.Timelines[0].TickNumber, latest.Timeline.TickNumber, "timeline latest/list head mismatch"); err != nil {
		return err
	}
	collector.Add("timeline", "latest/list", "sdk", "passed", fmt.Sprintf("latest_tick=%d", latest.Timeline.TickNumber))

	stateList, err := client.GetStateComponents(worldID)
	if err != nil {
		return err
	}
	worldState, err := client.GetStateComponent(worldID, "world_state")
	if err != nil {
		return err
	}
	storyState, err := client.GetStateComponent(worldID, "story_state")
	if err != nil {
		return err
	}
	storyHistory, err := client.GetStateComponent(worldID, "story_history")
	if err != nil {
		return err
	}
	worldTimeState, err := client.GetStateComponent(worldID, "world_time_state")
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(stateList.Components) >= 5, "state list should contain at least five continuity components"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(nonEmptyMapField(worldState.StateComponent.Data, "summary"), "world_state summary missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(nonEmptyMapField(storyState.StateComponent.Data, "current_situation"), "story_state current_situation missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(arrayFieldLen(storyHistory.StateComponent.Data, "entries") >= 2, "story_history should include the new tick entry"); err != nil {
		return err
	}
	if current := stringMapField(worldTimeState.StateComponent.Data, "current_time_label"); current != tick.WorldTimeState.CurrentTimeLabel {
		return fmt.Errorf("world_time_state current_time_label mismatch. expected=[%s] actual=[%s]", tick.WorldTimeState.CurrentTimeLabel, current)
	}
	collector.Add("state", "list/get", "sdk", "passed", "world_time_label="+stringMapField(worldTimeState.StateComponent.Data, "current_time_label"))

	continuity, err := client.GetContinuityBundle(worldID, &sdk.ContinuityBundleOptions{
		LogLimit:   20,
		TraceLimit: 10,
		LogQuery: &sdk.InferenceLogQuery{
			WorldID:   worldID,
			TaskType:  "world_tick",
			RequestID: requestID,
			Limit:     20,
		},
	})
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(continuity.LatestTimeline != nil, "continuity latest_timeline missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(continuity.StateComponents) >= 5, "continuity state components missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(continuity.Logs) >= 1, "continuity logs missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(len(continuity.Traces) >= 1, "continuity traces missing"); err != nil {
		return err
	}
	collector.Add("continuity", "debug continuity", "sdk", "passed", fmt.Sprintf("logs=%d traces=%d", len(continuity.Logs), len(continuity.Traces)))

	logs, err := client.GetLogsByQuery(sdk.InferenceLogQuery{WorldID: worldID, TaskType: "world_tick", RequestID: requestID, Limit: 20})
	if err != nil {
		return err
	}
	traces, err := client.GetDebugTraces(worldID, 10)
	if err != nil {
		return err
	}
	requestLogs := 0
	for _, item := range logs {
		if item.RequestID == requestID {
			requestLogs++
		}
	}
	requestTraces := 0
	for _, item := range traces.Traces {
		if item.RequestID == requestID {
			requestTraces++
		}
	}
	if err := workertest.AssertTrue(requestLogs >= 1, "request-scoped logs missing"); err != nil {
		return err
	}
	if err := workertest.AssertTrue(requestTraces >= 1, "request-scoped traces missing"); err != nil {
		return err
	}
	collector.Add("observability", "logs/traces correlation", "sdk", "passed", fmt.Sprintf("request_id=%s logs=%d traces=%d", requestID, requestLogs, requestTraces))

	result := continuityResult{
		WorldID:          worldID,
		RequestID:        requestID,
		LatestTickNumber: latest.Timeline.TickNumber,
		LatestTimeLabel:  stringMapField(worldTimeState.StateComponent.Data, "current_time_label"),
		Checks:           collector.Checks(),
	}
	return a.writeScenarioResult(result)
}

func readJSONFile(path string) (any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return decodeScenarioJSONValue(string(raw))
}

func decodeScenarioJSONValue(raw string) (any, error) {
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func timelineWorldTimeLabel(data any) (string, bool) {
	item, ok := data.(map[string]any)
	if !ok {
		return "", false
	}
	worldTimeState, ok := item["world_time_state"].(map[string]any)
	if !ok {
		return "", false
	}
	label, ok := worldTimeState["current_time_label"].(string)
	return label, ok
}

func nonEmptyMapField(data any, key string) bool {
	value := stringMapField(data, key)
	return strings.TrimSpace(value) != ""
}

func stringMapField(data any, key string) string {
	item, ok := data.(map[string]any)
	if !ok {
		return ""
	}
	value, _ := item[key].(string)
	return value
}

func arrayFieldLen(data any, key string) int {
	item, ok := data.(map[string]any)
	if !ok {
		return 0
	}
	list, ok := item[key].([]any)
	if !ok {
		return 0
	}
	return len(list)
}
