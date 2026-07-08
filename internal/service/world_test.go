package service

import (
	"fmt"
	"testing"

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

func TestAdvanceWorldTickWithAutonomousPersistsServiceLogs(t *testing.T) {
	initWorldServiceTestDB(t)
	worldID := createWorldRoot(t)
	previousMode := config.Global.Engine.ExecutionMode
	config.Global.Engine.ExecutionMode = "review"
	defer func() { config.Global.Engine.ExecutionMode = previousMode }()

	pipeline := engine.NewPipeline(&stubProvider{response: `{"reply":"tick","action_calls":[],"memory_updates":[],"world_change_plan":{"impact_level":"minor","summary":"世界推进","world_events":[],"proposed_actions":[]},"future_outline":"next"}`})
	tick, resp, autonomousRuns, err := AdvanceWorldTickWithAutonomous(pipeline, worldID, "scheduled", "day-1", nil)
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
	storyState, err := GetStateComponent(worldID, engine.CompStoryState)
	if err != nil {
		t.Fatalf("get story state: %v", err)
	}
	if storyState == nil || storyState.Data == "" {
		t.Fatal("expected persisted story_state component")
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
