package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/config"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type stubProvider struct {
	response string
	err      error
}

func (s *stubProvider) Chat(systemPrompt string, messages []engine.ChatMessage) (*engine.LLMResult, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &engine.LLMResult{Content: s.response, Model: "stub", Tokens: 9}, nil
}

func (s *stubProvider) ModelName() string { return "stub" }

func initWorldServiceTestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func createWorldRoot(t *testing.T) string {
	t.Helper()
	world := &store.NodeModel{UUID: store.NewUUID(), Name: "World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("set world id: %v", err)
	}
	return world.UUID
}

func putWorldTimeSettings(t *testing.T, worldID string, settings *engine.WorldTimeSettings) {
	t.Helper()
	raw, err := engine.EncodeWorldTimeSettings(settings)
	if err != nil {
		t.Fatalf("encode world time settings: %v", err)
	}
	if _, err := store.UpsertWorldSettingsWithMask(worldID, &store.WorldSettingsModel{WorldTimeSettingsJSON: raw}, &store.WorldSettingsUpdateMask{WorldTimeSettings: true}); err != nil {
		t.Fatalf("upsert world settings: %v", err)
	}
}

func decodeWorldTimeStateComponent(t *testing.T, worldID string) engine.WorldTimeStateComponent {
	t.Helper()
	component, err := GetStateComponent(worldID, engine.CompWorldTimeState)
	if err != nil {
		t.Fatalf("get world time state: %v", err)
	}
	if component == nil {
		t.Fatal("expected world_time_state component")
	}
	var state engine.WorldTimeStateComponent
	if err := json.Unmarshal([]byte(component.Data), &state); err != nil {
		t.Fatalf("decode world time state: %v", err)
	}
	return state
}

func TestAdvanceWorldTickWithAutonomousPersistsServiceLogs(t *testing.T) {
	initWorldServiceTestDB(t)
	worldID := createWorldRoot(t)
	previousMode := config.Global.Engine.ExecutionMode
	config.Global.Engine.ExecutionMode = "review"
	defer func() { config.Global.Engine.ExecutionMode = previousMode }()

	pipeline := engine.NewPipeline(&stubProvider{response: `{"reply":"地下52米处存在运行近3000年的非人类量子谐振腔。","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"世界推进","world_events":[],"proposed_actions":[]},"future_outline":"next"}`})
	tick, resp, worldTimeStateState, autonomousRuns, err := AdvanceWorldTickWithAutonomous(pipeline, worldID, "scheduled", "day-1", nil, nil)
	if err != nil {
		t.Fatalf("advance world tick: %v", err)
	}
	if tick == nil || resp == nil {
		t.Fatal("expected tick and response")
	}
	if tick.Data == "" {
		t.Fatal("expected timeline data payload")
	}
	if len(autonomousRuns) != 0 {
		t.Fatalf("expected no autonomous runs, got %#v", autonomousRuns)
	}
	worldState, err := GetStateComponent(worldID, engine.CompWorldState)
	if err != nil {
		t.Fatalf("get world state: %v", err)
	}
	if worldState == nil || worldState.Data == "" {
		t.Fatal("expected persisted world_state component")
	}
	if !strings.Contains(worldState.Data, "地下52米") {
		t.Fatalf("expected canonical fact retained in world_state, got %s", worldState.Data)
	}
	storyState, err := GetStateComponent(worldID, engine.CompStoryState)
	if err != nil {
		t.Fatalf("get story state: %v", err)
	}
	if storyState == nil || storyState.Data == "" {
		t.Fatal("expected persisted story_state component")
	}
	worldTimeState, err := GetStateComponent(worldID, engine.CompWorldTimeState)
	if err != nil {
		t.Fatalf("get world time state: %v", err)
	}
	if worldTimeState == nil || !strings.Contains(worldTimeState.Data, "day-1") {
		t.Fatalf("expected persisted world_time_state component, got %#v", worldTimeState)
	}
	if worldTimeStateState == nil || worldTimeStateState.CurrentTimeLabel == "" {
		t.Fatalf("expected returned world time state, got %#v", worldTimeStateState)
	}

	logs, err := store.GetInferenceLogs(worldID, 50, 0, string(engine.TaskWorldTick))
	if err != nil {
		t.Fatalf("get logs: %v", err)
	}
	var foundRequested, foundPersisted bool
	for _, item := range logs {
		if item.Category != "world_service" {
			continue
		}
		switch item.EventName {
		case "world_tick_requested":
			foundRequested = true
		case "world_tick_persisted":
			foundPersisted = true
		}
	}
	if !foundRequested || !foundPersisted {
		t.Fatalf("expected world service logs, got %#v", logs)
	}
}

func TestAdvanceWorldTickWithAutonomousRejectsRequestedTicksForFixedScale(t *testing.T) {
	initWorldServiceTestDB(t)
	worldID := createWorldRoot(t)
	putWorldTimeSettings(t, worldID, &engine.WorldTimeSettings{
		TickScaleMode: engine.TickScaleModeFixed,
		TickMinUnit:   "时辰",
		TickStep:      1,
		TickUnits:     []string{"日", "时辰"},
		TimeScaleCarry: []engine.WorldTimeCarryRule{{
			From: "时辰",
			To:   "日",
			Base: 12,
		}},
	})
	pipeline := engine.NewPipeline(&stubProvider{response: `{"reply":"世界推进","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"世界推进","world_events":[],"proposed_actions":[]},"future_outline":"next"}`})
	requestedTicks := 2

	_, _, _, _, err := AdvanceWorldTickWithAutonomous(pipeline, worldID, "scheduled", "day-1", &requestedTicks, nil)
	if err == nil {
		t.Fatal("expected fixed scale mode to reject requested_ticks=2")
	}
	if !IsKind(err, ErrorInvalid) {
		t.Fatalf("expected invalid error, got %v", err)
	}
	if ErrorCode(err) != "invalid_world_tick_request" {
		t.Fatalf("expected invalid_world_tick_request code, got %q", ErrorCode(err))
	}
}

func TestAdvanceWorldTickWithAutonomousPersistsFlexibleRequestedTicks(t *testing.T) {
	initWorldServiceTestDB(t)
	worldID := createWorldRoot(t)
	putWorldTimeSettings(t, worldID, &engine.WorldTimeSettings{
		TickScaleMode: engine.TickScaleModeFlexible,
		TickMinUnit:   "时辰",
		TickStep:      1,
		TickUnits:     []string{"日", "时辰"},
		TimeScaleCarry: []engine.WorldTimeCarryRule{{
			From: "时辰",
			To:   "日",
			Base: 12,
		}},
	})
	pipeline := engine.NewPipeline(&stubProvider{response: `{"reply":"世界推进","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"世界推进","world_events":[],"proposed_actions":[]},"future_outline":"next"}`})
	requestedTicks := 3

	tick, _, _, _, err := AdvanceWorldTickWithAutonomous(pipeline, worldID, "scheduled", "day-1", &requestedTicks, nil)
	if err != nil {
		t.Fatalf("advance world tick: %v", err)
	}
	if tick == nil {
		t.Fatal("expected tick")
	}
	state := decodeWorldTimeStateComponent(t, worldID)
	if state.TickScaleMode != engine.TickScaleModeFlexible {
		t.Fatalf("expected flexible tick scale mode, got %q", state.TickScaleMode)
	}
	if state.LastAdvancedTicks != 3 {
		t.Fatalf("expected last_advanced_ticks=3, got %d", state.LastAdvancedTicks)
	}
	if got, _ := state.Metadata["advanced_ticks"].(float64); int(got) != 3 {
		t.Fatalf("expected metadata advanced_ticks=3, got %#v", state.Metadata)
	}
}

func TestAdvanceWorldTickWithAutonomousUsesModelAdvancedTicksForFlexibleScale(t *testing.T) {
	initWorldServiceTestDB(t)
	worldID := createWorldRoot(t)
	putWorldTimeSettings(t, worldID, &engine.WorldTimeSettings{
		TickScaleMode: engine.TickScaleModeFlexible,
		TickMinUnit:   "时辰",
		TickStep:      1,
		TickUnits:     []string{"日", "时辰"},
		TimeScaleCarry: []engine.WorldTimeCarryRule{{
			From: "时辰",
			To:   "日",
			Base: 12,
		}},
	})
	pipeline := engine.NewPipeline(&stubProvider{response: `{"reply":"世界推进","advanced_ticks":4,"action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"世界推进","world_events":[],"proposed_actions":[]},"future_outline":"next"}`})
	requestedTicks := 2

	_, resp, _, _, err := AdvanceWorldTickWithAutonomous(pipeline, worldID, "scheduled", "day-1", &requestedTicks, nil)
	if err != nil {
		t.Fatalf("advance world tick: %v", err)
	}
	if resp == nil || resp.AdvancedTicks != 4 {
		t.Fatalf("expected response advanced_ticks=4, got %#v", resp)
	}
	state := decodeWorldTimeStateComponent(t, worldID)
	if state.LastAdvancedTicks != 4 {
		t.Fatalf("expected persisted last_advanced_ticks=4, got %d", state.LastAdvancedTicks)
	}
}

func TestCollectCanonicalWorldFactsPrefersConcreteFacts(t *testing.T) {
	resp := &engine.InvokeResponse{
		Reply: strings.Join([]string{
			"世界推进，局势持续变化。",
			"地下52米量子谐振腔仍在运行，并由Dar-shade检修站持续供能。",
			"设施状态稳定。",
		}, "\n"),
		MemoryUpdates: []engine.MemoryUpdate{{Content: "He-3精炼厂的三号冷却井出现间歇性结霜。"}},
		WorldChangePlan: &engine.WorldChangePlan{
			Summary:     "剧情继续推进",
			WorldEvents: []engine.PlanEvent{{Description: "轨道站A-17向地下城投送新的谐振腔备件。"}},
		},
	}

	facts := collectCanonicalWorldFacts(resp)
	joined := strings.Join(facts, "\n")
	if !strings.Contains(joined, "地下52米量子谐振腔仍在运行") {
		t.Fatalf("expected concrete underground fact, got %#v", facts)
	}
	if !strings.Contains(joined, "Dar-shade检修站持续供能") {
		t.Fatalf("expected facility detail retained, got %#v", facts)
	}
	if !strings.Contains(joined, "He-3精炼厂的三号冷却井出现间歇性结霜") {
		t.Fatalf("expected memory fact retained, got %#v", facts)
	}
	if !strings.Contains(joined, "轨道站A-17向地下城投送新的谐振腔备件") {
		t.Fatalf("expected event fact retained, got %#v", facts)
	}
	if strings.Contains(joined, "世界推进") || strings.Contains(joined, "局势持续变化") || strings.Contains(joined, "剧情继续推进") {
		t.Fatalf("expected generic continuity phrases filtered, got %#v", facts)
	}
}

func TestNormalizeCanonicalFactRejectsGenericPhrases(t *testing.T) {
	for _, value := range []string{
		"世界推进",
		"局势发生变化",
		"行动正在继续",
		"设施状态稳定",
	} {
		if fact := normalizeCanonicalFact(value); fact != "" {
			t.Fatalf("expected %q to be rejected, got %q", value, fact)
		}
	}

	for _, value := range []string{
		"地下52米量子谐振腔仍在运行",
		"Dar-shade检修站接管A-17轨道站的供能回路",
		"He-3精炼厂的三号冷却井出现结霜",
	} {
		if fact := normalizeCanonicalFact(value); fact == "" {
			t.Fatalf("expected %q to be retained", value)
		}
	}
}

func TestWithWorldLockSerializesSameWorld(t *testing.T) {
	enteredCh := make(chan string, 2)
	releaseCh := make(chan struct{}, 2)
	errCh := make(chan error, 2)
	run := func(label string) {
		errCh <- withWorldLock("same-world", func() error {
			enteredCh <- label
			<-releaseCh
			return nil
		})
	}

	go run("first")
	select {
	case got := <-enteredCh:
		if got != "first" {
			t.Fatalf("expected first lock holder, got %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("first lock holder did not enter")
	}
	go run("second")
	select {
	case got := <-enteredCh:
		t.Fatalf("second same-world operation entered too early: %q", got)
	case <-time.After(200 * time.Millisecond):
	}
	releaseCh <- struct{}{}
	select {
	case got := <-enteredCh:
		if got != "second" {
			t.Fatalf("expected second lock holder after release, got %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("second lock holder did not enter after release")
	}
	releaseCh <- struct{}{}
	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("withWorldLock failed: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("locked operation did not finish")
		}
	}
}

func TestWithWorldLockAllowsDifferentWorldConcurrency(t *testing.T) {
	enteredCh := make(chan string, 2)
	releaseCh := make(chan struct{}, 2)
	errCh := make(chan error, 2)
	activeMu := sync.Mutex{}
	active := 0
	maxActive := 0
	run := func(worldID string) {
		errCh <- withWorldLock(worldID, func() error {
			activeMu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			activeMu.Unlock()
			enteredCh <- worldID
			<-releaseCh
			activeMu.Lock()
			active--
			activeMu.Unlock()
			return nil
		})
	}

	go run("world-a")
	go run("world-b")
	seen := map[string]bool{}
	for len(seen) < 2 {
		select {
		case worldID := <-enteredCh:
			seen[worldID] = true
		case <-time.After(3 * time.Second):
			t.Fatal("expected both different-world operations to enter")
		}
	}
	releaseCh <- struct{}{}
	releaseCh <- struct{}{}
	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("withWorldLock failed: %v", err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("locked operation did not finish")
		}
	}
	if maxActive < 2 {
		t.Fatalf("expected different worlds to overlap, max active=%d", maxActive)
	}
}
