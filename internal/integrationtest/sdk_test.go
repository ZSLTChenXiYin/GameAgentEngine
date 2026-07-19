//go:build integration

package integrationtest

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/api"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/llm"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func TestSDKIntegration(t *testing.T) {
	store.ConfigureLogSink(store.LogSinkOptions{Enabled: false, BatchSize: 1, FlushInterval: 0, QueueSize: 1})
	if err := store.Init("sqlite", "file:test-int-sdk?mode=memory&cache=shared"); err != nil {
		t.Fatalf("init store: %v", err)
	}
	pipeline := engine.NewPipeline(llm.NewMockProvider())
	mux := api.NewRouter(pipeline)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := sdk.NewClient(ts.URL, "dev-key")

	if err := client.Health(); err != nil {
		t.Fatalf("health: %v", err)
	}
	version, _, err := client.GetVersion()
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if version == "" {
		t.Error("expected non-empty version")
	}

	worldID, err := client.CreateNode("", "Test World", "world", "")
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	npcID, err := client.CreateNode(worldID, "Test NPC", "npc", worldID)
	if err != nil {
		t.Fatalf("create npc: %v", err)
	}

	nodes, err := client.GetNodes(worldID, 10, 0, "")
	if err != nil {
		t.Fatalf("get nodes: %v", err)
	}
	if len(nodes) < 2 {
		t.Errorf("expected >=2 nodes, got %d", len(nodes))
	}

	_, err = client.AddComponent(npcID, "autonomous", `{"trigger":"manual","enabled":true}`)
	if err != nil {
		t.Fatalf("add component: %v", err)
	}

	_, err = client.AddMemory(npcID, "Test memory", "short_term", "test")
	if err != nil {
		t.Fatalf("add memory: %v", err)
	}

	raw, err := client.RawGet("/api/v1/version")
	if err != nil {
		t.Fatalf("raw get: %v", err)
	}
	var versionResp map[string]string
	if err := json.Unmarshal(raw, &versionResp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if versionResp["version"] == "" {
		t.Error("expected version in response")
	}

	// World settings
	settings, err := client.GetWorldSettings(worldID)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings != nil && settings.MaxAnalysisRounds != 0 {
		t.Logf("settings: memory_limit=%d analysis_rounds=%d", settings.MemoryLimit, settings.MaxAnalysisRounds)
	}
	// Update world settings
	memLimit := 100
	maxRounds := 10
	updateSettings := &sdk.WorldSettingsUpdate{
		MemoryLimit:       &memLimit,
		MaxAnalysisRounds: &maxRounds,
	}
	if _, err := client.UpdateWorldSettings(worldID, updateSettings); err != nil {
		t.Fatalf("update settings: %v", err)
	}
	// Verify update took effect
	updatedSettings, err := client.GetWorldSettings(worldID)
	if err != nil {
		t.Fatalf("get settings after update: %v", err)
	}
	if updatedSettings == nil || updatedSettings.MaxAnalysisRounds != 10 {
		t.Errorf("expected MaxAnalysisRounds=10, got %+v", updatedSettings)
	}

	t.Log("SDK integration test passed")
}
