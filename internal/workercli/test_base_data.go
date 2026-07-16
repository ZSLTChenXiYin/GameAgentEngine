package workercli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/workertest"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
	"gopkg.in/yaml.v3"
)

type baseDataResult struct {
	RunSuffix        string                 `json:"run_suffix"`
	WorldName        string                 `json:"world_name"`
	WorldID          string                 `json:"world_id"`
	StressWorldName  string                 `json:"stress_world_name"`
	StressWorldID    string                 `json:"stress_world_id"`
	Checks           []workertest.CheckResult `json:"checks"`
}

func (a *app) runBaseDataScenario() error {
	if strings.TrimSpace(a.cfg.TestDevCLIExePath) == "" {
		return fmt.Errorf("base-data requires --devcli-exe")
	}
	testsDir := strings.TrimSpace(a.cfg.TestsDir)
	if testsDir == "" {
		return fmt.Errorf("base-data requires --tests-dir")
	}
	fixturePath := filepath.Join(testsDir, "full_functional_base_data_world.yaml")
	if _, err := os.Stat(fixturePath); err != nil {
		return fmt.Errorf("base-data fixture not found: %w", err)
	}
	engineBaseURL := strings.TrimSpace(a.cfg.EngineBaseURL)
	if engineBaseURL == "" {
		engineBaseURL = fmt.Sprintf("http://127.0.0.1:%d", a.cfg.TestEnginePort)
	}
	apiKey := strings.TrimSpace(a.cfg.EngineAPIKey)
	if apiKey == "" {
		apiKey = "dev-key"
	}
	client := &workertest.Client{BaseURL: engineBaseURL, APIKey: apiKey}
	devcli := workertest.DevCLI{Executable: a.cfg.TestDevCLIExePath, Server: engineBaseURL, APIKey: apiKey}
	collector := &workertest.Collector{}

	runSuffix := time.Now().Format("20060102150405")
	worldName := "FullFunctionalBaseDataWorld-" + runSuffix
	stressWorldName := "FullFunctionalBaseDataRaceWorld-" + runSuffix

	generatedFixturePath, err := a.prepareBaseDataFixture(fixturePath, worldName, runSuffix)
	if err != nil {
		return err
	}
	defer os.Remove(generatedFixturePath)

	var importResult sdk.ImportResult
	if err := devcli.RunJSON(&importResult, "import", generatedFixturePath); err != nil {
		return err
	}
	if err := workertest.AssertEqual(importResult.WorldName, worldName, "fixture import world name mismatch"); err != nil {
		return err
	}
	collector.Add("world", "import fixture", "devcli", "passed", "world_id="+importResult.WorldID)

	world, err := a.findWorldByName(client, worldName)
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(world != nil, "imported world not found via HTTP"); err != nil {
		return err
	}
	worldID := world.ID
	collector.Add("world", "locate imported world", "http", "passed", "world_id="+worldID)

	baseNPC, err := a.findNodeByName(client, worldID, "Quartermaster Lin")
	if err != nil {
		return err
	}
	if err := workertest.AssertTrue(baseNPC != nil, "base npc not found"); err != nil {
		return err
	}

	var createdNode sdk.Node
	if err := client.EngineJSON("POST", "/api/v1/nodes", map[string]any{
		"world_id":  worldID,
		"name":      "Watch Captain Rhea",
		"node_type": "npc",
		"parent_id": baseNPC.ID,
	}, nil, &createdNode); err != nil {
		return err
	}
	if err := workertest.AssertEqual(createdNode.Name, "Watch Captain Rhea", "http node create returned wrong name"); err != nil {
		return err
	}
	nodeID := createdNode.ID
	collector.Add("node", "create", "http", "passed", "node_id="+nodeID)

	var nodeDetail sdk.NodeDetail
	if err := devcli.RunJSON(&nodeDetail, "node", "get", nodeID); err != nil {
		return err
	}
	if err := workertest.AssertEqual(nodeDetail.Node.Name, "Watch Captain Rhea", "devcli node get mismatch after http create"); err != nil {
		return err
	}
	collector.Add("node", "get after create", "devcli", "passed", "node_id="+nodeID)

	var updatedNode sdk.Node
	if err := devcli.RunJSON(&updatedNode, "node", "update", nodeID, "--name", "Watch Captain Rhea II", "--type", "npc"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(updatedNode.Name, "Watch Captain Rhea II", "devcli node update returned wrong name"); err != nil {
		return err
	}
	var updatedNodeDetail sdk.NodeDetail
	if err := client.EngineJSON("GET", "/api/v1/nodes/"+nodeID, nil, nil, &updatedNodeDetail); err != nil {
		return err
	}
	if err := workertest.AssertEqual(updatedNodeDetail.Node.Name, "Watch Captain Rhea II", "http node get mismatch after devcli update"); err != nil {
		return err
	}
	collector.Add("node", "update", "devcli->http", "passed", "node_id="+nodeID)

	var legacyNodes []sdk.Node
	if err := devcli.RunJSON(&legacyNodes, "nodes", "--world", worldID); err != nil {
		return err
	}
	var httpNodes []sdk.Node
	if err := client.EngineJSON("GET", workertest.QueryWithValues("/api/v1/nodes", map[string]string{"world_id": worldID}), nil, nil, &httpNodes); err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(legacyNodes), len(httpNodes), "legacy nodes count mismatch"); err != nil {
		return err
	}
	collector.Add("node", "legacy list parity", "devcli/http", "passed", fmt.Sprintf("count=%d", len(httpNodes)))

	var component sdk.Component
	if err := client.EngineJSON("POST", "/api/v1/components", map[string]any{
		"node_id":        nodeID,
		"component_type": "rule",
		"data":           "watch=north",
	}, nil, &component); err != nil {
		return err
	}
	if err := workertest.AssertEqual(component.ComponentType, "rule", "http component create returned wrong type"); err != nil {
		return err
	}
	componentID := component.ID
	collector.Add("component", "create", "http", "passed", "component_id="+componentID)

	var componentDetail sdk.Component
	if err := devcli.RunJSON(&componentDetail, "component", "get", componentID); err != nil {
		return err
	}
	if err := workertest.AssertEqual(componentDetail.ID, componentID, "devcli component get mismatch after http create"); err != nil {
		return err
	}
	if _, err := devcli.Run("component", "update", componentID, "--data", "watch=west"); err != nil {
		return err
	}
	var componentAfterUpdate sdk.Component
	if err := client.EngineJSON("GET", "/api/v1/components/"+componentID, nil, nil, &componentAfterUpdate); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.Contains(componentAfterUpdate.Data, "west"), "http component get mismatch after devcli update"); err != nil {
		return err
	}
	collector.Add("component", "update", "devcli->http", "passed", "component_id="+componentID)

	var memory sdk.Memory
	if err := devcli.RunJSON(&memory, "memory", "create", "--node", nodeID, "--content", "Northern watch changed patrol route.", "--level", "short_term", "--tags", "patrol,watch"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(memory.NodeID, nodeID, "devcli memory create returned wrong node"); err != nil {
		return err
	}
	memoryID := memory.ID
	collector.Add("memory", "create", "devcli", "passed", "memory_id="+memoryID)

	var memoryHTTP sdk.Memory
	if err := client.EngineJSON("GET", "/api/v1/memories/"+memoryID, nil, nil, &memoryHTTP); err != nil {
		return err
	}
	if err := workertest.AssertEqual(memoryHTTP.Content, "Northern watch changed patrol route.", "http memory get mismatch after devcli create"); err != nil {
		return err
	}
	var memoryUpdated sdk.Memory
	if err := client.EngineJSON("PUT", "/api/v1/memories/"+memoryID, map[string]any{"content": "Northern watch rerouted through the western stairs.", "tags": "patrol,west"}, nil, &memoryUpdated); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.Contains(memoryUpdated.Content, "western stairs"), "http memory update returned wrong content"); err != nil {
		return err
	}
	var memoryAfterUpdate sdk.Memory
	if err := devcli.RunJSON(&memoryAfterUpdate, "memory", "get", memoryID); err != nil {
		return err
	}
	if err := workertest.AssertTrue(strings.Contains(memoryAfterUpdate.Content, "western stairs"), "devcli memory get mismatch after http update"); err != nil {
		return err
	}
	collector.Add("memory", "update", "http->devcli", "passed", "memory_id="+memoryID)

	var relation sdk.Relation
	if err := client.EngineJSON("POST", "/api/v1/relations", map[string]any{
		"world_id":      worldID,
		"source_id":     nodeID,
		"target_id":     baseNPC.ID,
		"relation_type": "subordinate",
		"weight":        5,
		"properties":    `{"duty":"watch_command"}`,
	}, nil, &relation); err != nil {
		return err
	}
	if err := workertest.AssertEqual(relation.RelationType, "subordinate", "http relation create returned wrong type"); err != nil {
		return err
	}
	relationID := relation.ID
	collector.Add("relation", "create", "http", "passed", "relation_id="+relationID)

	var relationDetail sdk.Relation
	if err := devcli.RunJSON(&relationDetail, "relation", "get", relationID); err != nil {
		return err
	}
	if err := workertest.AssertEqual(relationDetail.ID, relationID, "devcli relation get mismatch after http create"); err != nil {
		return err
	}
	if _, err := devcli.Run("relation", "update", relationID, "--weight", "7", "--props", `{"duty":"west_watch"}`); err != nil {
		return err
	}
	var relationAfterUpdate sdk.Relation
	if err := client.EngineJSON("GET", "/api/v1/relations/"+relationID, nil, nil, &relationAfterUpdate); err != nil {
		return err
	}
	if err := workertest.AssertEqual(relationAfterUpdate.Weight, 7, "http relation get mismatch after devcli update"); err != nil {
		return err
	}
	collector.Add("relation", "update", "devcli->http", "passed", "relation_id="+relationID)

	var worldSettings sdk.WorldSettings
	if err := devcli.RunJSON(&worldSettings,
		"world", "settings", "set", worldID,
		"--memory-limit", "24",
		"--analysis-rounds", "3",
		"--context-depth", "4",
		"--auto-apply=false",
		"--review-above", "high",
		"--propagation-max-depth", "2",
		"--enable-propagation-machine=true",
		"--sub-task-max-retries", "5",
		"--sub-task-timeout-secs", "90",
		"--pipeline-mode", "polling",
	); err != nil {
		return err
	}
	if err := workertest.AssertEqual(worldSettings.PipelineMode, "polling", "devcli world settings set returned wrong pipeline mode"); err != nil {
		return err
	}
	var worldSettingsHTTP sdk.WorldSettings
	if err := client.EngineJSON("GET", "/api/v1/worlds/"+worldID+"/settings", nil, nil, &worldSettingsHTTP); err != nil {
		return err
	}
	if err := workertest.AssertEqual(worldSettingsHTTP.PipelineMode, "polling", "http world settings get mismatch after devcli set"); err != nil {
		return err
	}
	if err := workertest.AssertEqual(worldSettingsHTTP.MemoryLimit, 24, "http world settings get mismatch for memory_limit"); err != nil {
		return err
	}
	collector.Add("world_settings", "set/get", "devcli->http", "passed", "pipeline_mode="+worldSettingsHTTP.PipelineMode)

	var worldPolicy sdk.WorldPolicy
	if err := client.EngineJSON("PUT", "/api/v1/worlds/"+worldID+"/policy", map[string]any{"blocked_actions": []string{"spawn_item"}, "safe_actions": []string{"inspect_map", "request_backup"}}, nil, &worldPolicy); err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(worldPolicy.BlockedActions), 1, "http world policy set mismatch"); err != nil {
		return err
	}
	var worldPolicyCLI sdk.WorldPolicy
	if err := devcli.RunJSON(&worldPolicyCLI, "world", "policy", "get", worldID); err != nil {
		return err
	}
	if err := workertest.AssertEqual(strings.Join(worldPolicyCLI.SafeActions, ","), "inspect_map,request_backup", "devcli world policy get mismatch after http set"); err != nil {
		return err
	}
	collector.Add("world_policy", "set/get", "http->devcli", "passed", "blocked="+strings.Join(worldPolicyCLI.BlockedActions, ","))

	var stressWorld sdk.Node
	if err := devcli.RunJSON(&stressWorld, "node", "create", "--type", "world", "--name", stressWorldName); err != nil {
		return err
	}
	stressWorldID := stressWorld.ID
	var stressNode sdk.Node
	if err := devcli.RunJSON(&stressNode, "node", "create", "--world", stressWorldID, "--type", "npc", "--name", "Stress Harness Node"); err != nil {
		return err
	}
	stressNodeID := stressNode.ID
	if err := a.runConcurrentComponentCreates(client, stressNodeID); err != nil {
		return err
	}
	var stressComponents []sdk.Component
	if err := client.EngineJSON("GET", workertest.QueryWithValues("/api/v1/components", map[string]string{"node_id": stressNodeID}), nil, nil, &stressComponents); err != nil {
		return err
	}
	if err := workertest.AssertEqual(len(stressComponents), 6, "concurrent component creation count mismatch"); err != nil {
		return err
	}
	collector.Add("component", "concurrent create on fresh world", "http", "passed", fmt.Sprintf("world_id=%s count=6", stressWorldID))

	if _, err := devcli.Run("component", "delete", componentID); err != nil {
		return err
	}
	if err := a.assertNotFound(func() error { return client.EngineJSON("GET", "/api/v1/components/"+componentID, nil, nil, &sdk.Component{}) }, "component still exists after delete"); err != nil {
		return err
	}
	collector.Add("component", "delete", "devcli->http", "passed", "component_id="+componentID)

	if _, err := devcli.Run("memory", "delete", memoryID); err != nil {
		return err
	}
	if err := a.assertNotFound(func() error { return client.EngineJSON("GET", "/api/v1/memories/"+memoryID, nil, nil, &sdk.Memory{}) }, "memory still exists after delete"); err != nil {
		return err
	}
	collector.Add("memory", "delete", "devcli->http", "passed", "memory_id="+memoryID)

	if err := client.EngineJSON("DELETE", "/api/v1/relations/"+relationID, nil, nil, nil); err != nil {
		return err
	}
	if err := a.assertNotFound(func() error { return client.EngineJSON("GET", "/api/v1/relations/"+relationID, nil, nil, &sdk.Relation{}) }, "relation still exists after delete"); err != nil {
		return err
	}
	collector.Add("relation", "delete", "http", "passed", "relation_id="+relationID)

	if _, err := devcli.Run("node", "delete", nodeID); err != nil {
		return err
	}
	if err := a.assertNotFound(func() error { return client.EngineJSON("GET", "/api/v1/nodes/"+nodeID, nil, nil, &sdk.NodeDetail{}) }, "node still exists after delete"); err != nil {
		return err
	}
	collector.Add("node", "delete", "devcli->http", "passed", "node_id="+nodeID)

	result := baseDataResult{
		RunSuffix:       runSuffix,
		WorldName:       worldName,
		WorldID:         worldID,
		StressWorldName: stressWorldName,
		StressWorldID:   stressWorldID,
		Checks:          collector.Checks(),
	}
	return a.writeScenarioResult(result)
}

func (a *app) prepareBaseDataFixture(fixturePath, worldName, runSuffix string) (string, error) {
	raw, err := os.ReadFile(fixturePath)
	if err != nil {
		return "", err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return "", err
	}
	world, ok := doc["world"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("fixture missing world block")
	}
	world["name"] = worldName
	encoded, err := yaml.Marshal(doc)
	if err != nil {
		return "", err
	}
	path := filepath.Join(os.TempDir(), fmt.Sprintf("full-functional-base-data-%s.yaml", runSuffix))
	if err := os.WriteFile(path, encoded, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func (a *app) findWorldByName(client *workertest.Client, name string) (*struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}, error) {
	var worlds []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := client.EngineJSON("GET", "/api/v1/worlds", nil, nil, &worlds); err != nil {
		return nil, err
	}
	for i := range worlds {
		if worlds[i].Name == name {
			return &worlds[i], nil
		}
	}
	return nil, nil
}

func (a *app) findNodeByName(client *workertest.Client, worldID string, name string) (*sdk.Node, error) {
	var nodes []sdk.Node
	if err := client.EngineJSON("GET", workertest.QueryWithValues("/api/v1/nodes", map[string]string{"world_id": worldID}), nil, nil, &nodes); err != nil {
		return nil, err
	}
	for i := range nodes {
		if nodes[i].Name == name {
			return &nodes[i], nil
		}
	}
	return nil, nil
}

func (a *app) runConcurrentComponentCreates(client *workertest.Client, nodeID string) error {
	var wg sync.WaitGroup
	errCh := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			body := map[string]any{
				"node_id":        nodeID,
				"component_type": "rule",
				"data":           fmt.Sprintf(`{"slot":%d}`, index),
			}
			if err := client.EngineJSON("POST", "/api/v1/components", body, nil, &sdk.Component{}); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *app) assertNotFound(action func() error, message string) error {
	err := action()
	if err == nil {
		return fmt.Errorf("%s", message)
	}
	if strings.Contains(err.Error(), " 404 ") || strings.Contains(err.Error(), ": 404 ") {
		return nil
	}
	return err
}

func (a *app) writeScenarioResult(result any) error {
	data, err := sdkMarshalIndent(result)
	if err != nil {
		return err
	}
	if out := strings.TrimSpace(a.cfg.TestOutFile); out != "" {
		if err := workertest.WriteFile(out, data); err != nil {
			return err
		}
	}
	if a.cfg.TestJSON || strings.TrimSpace(a.cfg.TestOutFile) == "" {
		fmt.Println(string(data))
	}
	return nil
}

func sdkMarshalIndent(v any) ([]byte, error) {
	return jsonMarshalIndent(v)
}
