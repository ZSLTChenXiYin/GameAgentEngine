package service

import (
	"fmt"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/version"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

type testMemoryConfig struct {
	Content string
	Level   string
	Tags    string
}

func initImportExportTestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func makeNodeConfig(name, nodeType, parent, profile, lore string, memories ...testMemoryConfig) sdk.NodeConfig {
	node := sdk.NodeConfig{
		Name:    name,
		Type:    nodeType,
		Parent:  parent,
		Profile: profile,
		Lore:    lore,
	}
	for _, memory := range memories {
		node.Memories = append(node.Memories, struct {
			Content string `json:"content" yaml:"content"`
			Level   string `json:"level,omitempty" yaml:"level,omitempty"`
			Tags    string `json:"tags,omitempty" yaml:"tags,omitempty"`
		}{
			Content: memory.Content,
			Level:   memory.Level,
			Tags:    memory.Tags,
		})
	}
	return node
}

func TestForkWorldCopiesStructureAndSnapshotMetadata(t *testing.T) {
	initImportExportTestDB(t)

	cfg := &sdk.ImportConfig{
		World: sdk.WorldConfig{Name: "Origin World"},
		Nodes: []sdk.NodeConfig{
			makeNodeConfig("Hero", "npc", "", "brave hero", "", testMemoryConfig{Content: "arrived in town", Level: "short_term", Tags: "arrival"}),
			makeNodeConfig("Bag", "item", "Hero", "", "starter bag", testMemoryConfig{Content: "contains coins", Tags: "inventory"}),
			makeNodeConfig("Guard", "npc", "", "", ""),
		},
		Components: []sdk.ComponentConfig{
			{NodeID: "Guard", Type: "rule", Data: `{"patrol":true}`},
		},
		Relations: []sdk.RelationConfig{
			{Source: "Hero", Target: "Guard", Type: "ally", Weight: 3, Props: `{"trust":80}`},
			{Source: "Bag", Target: "Hero", Type: "belongs_to", Weight: 1, Props: `{"slot":"inventory"}`},
		},
	}

	result, err := ImportWorld(cfg, false, false)
	if err != nil {
		t.Fatalf("import world: %v", err)
	}
	originWorldID := result.WorldID

	if _, err := store.UpsertWorldSettingsWithMask(originWorldID, &store.WorldSettingsModel{
		MemoryLimit:              77,
		MaxAnalysisRounds:        4,
		MaxContextDepth:          5,
		AutoApply:                false,
		RequireReviewAbove:       "high",
		PipelineMode:             "polling",
		PropagationMaxDepth:      6,
		SubTaskMaxRetries:        7,
		SubTaskTimeoutSecs:       88,
		EnablePropagationMachine: true,
		WorldTimeSettingsJSON:    `{"tick_scale_mode":"fixed","tick_min_unit":"时辰","tick_step":2}`,
	}, &store.WorldSettingsUpdateMask{
		MemoryLimit:              true,
		MaxAnalysisRounds:        true,
		MaxContextDepth:          true,
		AutoApply:                true,
		RequireReviewAbove:       true,
		PipelineMode:             true,
		PropagationMaxDepth:      true,
		SubTaskMaxRetries:        true,
		SubTaskTimeoutSecs:       true,
		EnablePropagationMachine: true,
		WorldTimeSettings:        true,
	}); err != nil {
		t.Fatalf("upsert settings: %v", err)
	}
	if _, err := store.UpsertWorldPolicy(originWorldID, []string{"spawn_item"}, []string{"inspect_map"}); err != nil {
		t.Fatalf("upsert policy: %v", err)
	}

	clonedWorld, err := ForkWorld(originWorldID, "Origin World Save 1", false)
	if err != nil {
		t.Fatalf("fork world: %v", err)
	}
	if clonedWorld.UUID == originWorldID {
		t.Fatal("expected cloned world uuid to differ from source")
	}
	if clonedWorld.Name != "Origin World Save 1" {
		t.Fatalf("unexpected clone name: %s", clonedWorld.Name)
	}

	clonedNodes, err := store.GetAllNodes(clonedWorld.UUID, 0, 0, "")
	if err != nil {
		t.Fatalf("get cloned nodes: %v", err)
	}
	if len(clonedNodes) != 4 {
		t.Fatalf("expected 4 nodes in cloned world including world root, got %d", len(clonedNodes))
	}

	nodesByName := make(map[string]store.NodeModel, len(clonedNodes))
	for _, node := range clonedNodes {
		nodesByName[node.Name] = node
		if node.WorldUUID != clonedWorld.UUID {
			t.Fatalf("node %s has wrong world uuid: %s", node.Name, node.WorldUUID)
		}
	}
	heroNode, ok := nodesByName["Hero"]
	if !ok {
		t.Fatal("expected cloned Hero node")
	}
	bagNode, ok := nodesByName["Bag"]
	if !ok {
		t.Fatal("expected cloned Bag node")
	}
	guardNode, ok := nodesByName["Guard"]
	if !ok {
		t.Fatal("expected cloned Guard node")
	}
	if heroNode.ParentUUID != nil {
		t.Fatalf("expected Hero to remain a top-level node, got parent %v", *heroNode.ParentUUID)
	}
	if bagNode.ParentUUID == nil || *bagNode.ParentUUID != heroNode.UUID {
		t.Fatalf("expected Bag parent to map to cloned Hero, got %v", bagNode.ParentUUID)
	}
	if heroNode.UUID == "" || bagNode.UUID == "" || guardNode.UUID == "" {
		t.Fatal("expected cloned nodes to have uuids")
	}

	heroComponents, err := store.GetNodeComponents(heroNode.UUID)
	if err != nil {
		t.Fatalf("get hero components: %v", err)
	}
	if len(heroComponents) != 1 || heroComponents[0].ComponentType != "profile" || heroComponents[0].Data != "brave hero" {
		t.Fatalf("unexpected hero components: %#v", heroComponents)
	}
	if heroComponents[0].NodeUUID != heroNode.UUID {
		t.Fatalf("expected hero component to point to cloned hero, got %s", heroComponents[0].NodeUUID)
	}

	bagComponents, err := store.GetNodeComponents(bagNode.UUID)
	if err != nil {
		t.Fatalf("get bag components: %v", err)
	}
	if len(bagComponents) != 1 || bagComponents[0].ComponentType != "lore" || bagComponents[0].Data != "starter bag" {
		t.Fatalf("unexpected bag components: %#v", bagComponents)
	}

	guardComponents, err := store.GetNodeComponents(guardNode.UUID)
	if err != nil {
		t.Fatalf("get guard components: %v", err)
	}
	if len(guardComponents) != 1 || guardComponents[0].ComponentType != "rule" || guardComponents[0].Data != `{"patrol":true}` {
		t.Fatalf("unexpected guard components: %#v", guardComponents)
	}

	heroMemories, err := store.GetNodeMemories(heroNode.UUID, 0)
	if err != nil {
		t.Fatalf("get hero memories: %v", err)
	}
	if len(heroMemories) != 1 || heroMemories[0].Content != "arrived in town" || heroMemories[0].Level != "short_term" || heroMemories[0].Tags != "arrival" {
		t.Fatalf("unexpected hero memories: %#v", heroMemories)
	}
	if heroMemories[0].NodeUUID != heroNode.UUID {
		t.Fatalf("expected hero memory to point to cloned hero, got %s", heroMemories[0].NodeUUID)
	}

	bagMemories, err := store.GetNodeMemories(bagNode.UUID, 0)
	if err != nil {
		t.Fatalf("get bag memories: %v", err)
	}
	if len(bagMemories) != 1 || bagMemories[0].Content != "contains coins" || bagMemories[0].Level != "long_term" || bagMemories[0].Tags != "inventory" {
		t.Fatalf("unexpected bag memories: %#v", bagMemories)
	}

	clonedRelations, err := store.GetAllRelations(clonedWorld.UUID, 0, 0, "")
	if err != nil {
		t.Fatalf("get cloned relations: %v", err)
	}
	if len(clonedRelations) != 2 {
		t.Fatalf("expected 2 cloned relations, got %d", len(clonedRelations))
	}
	relationByType := make(map[string]store.RelationModel, len(clonedRelations))
	for _, relation := range clonedRelations {
		relationByType[relation.RelationType] = relation
		if relation.WorldUUID != clonedWorld.UUID {
			t.Fatalf("relation %s has wrong world uuid: %s", relation.RelationType, relation.WorldUUID)
		}
	}
	allyRelation, ok := relationByType["ally"]
	if !ok {
		t.Fatal("expected ally relation in clone")
	}
	if allyRelation.SourceUUID != heroNode.UUID || allyRelation.TargetUUID != guardNode.UUID || allyRelation.Weight != 3 || allyRelation.Properties != `{"trust":80}` {
		t.Fatalf("unexpected ally relation: %#v", allyRelation)
	}
	belongsRelation, ok := relationByType["belongs_to"]
	if !ok {
		t.Fatal("expected belongs_to relation in clone")
	}
	if belongsRelation.SourceUUID != bagNode.UUID || belongsRelation.TargetUUID != heroNode.UUID || belongsRelation.Weight != 1 || belongsRelation.Properties != `{"slot":"inventory"}` {
		t.Fatalf("unexpected belongs_to relation: %#v", belongsRelation)
	}

	clonedSettings, err := store.GetWorldSettings(clonedWorld.UUID)
	if err != nil {
		t.Fatalf("get cloned settings: %v", err)
	}
	if clonedSettings.MemoryLimit != 77 || clonedSettings.MaxAnalysisRounds != 4 || clonedSettings.MaxContextDepth != 5 || clonedSettings.AutoApply != false || clonedSettings.RequireReviewAbove != "high" || clonedSettings.PipelineMode != "polling" || clonedSettings.PropagationMaxDepth != 6 || clonedSettings.SubTaskMaxRetries != 7 || clonedSettings.SubTaskTimeoutSecs != 88 || clonedSettings.EnablePropagationMachine != true || clonedSettings.WorldTimeSettingsJSON != `{"tick_scale_mode":"fixed","tick_min_unit":"时辰","tick_step":2}` {
		t.Fatalf("unexpected cloned settings: %#v", clonedSettings)
	}

	clonedPolicy, err := store.GetWorldPolicy(clonedWorld.UUID)
	if err != nil {
		t.Fatalf("get cloned policy: %v", err)
	}
	if clonedPolicy.BlockedActions != `["spawn_item"]` || clonedPolicy.SafeActions != `["inspect_map"]` {
		t.Fatalf("unexpected cloned policy: %#v", clonedPolicy)
	}

	snapshot, err := store.GetWorldSnapshotBySnapshotWorld(clonedWorld.UUID)
	if err != nil {
		t.Fatalf("get snapshot metadata: %v", err)
	}
	if snapshot.SourceWorldUUID != originWorldID || snapshot.SnapshotWorldUUID != clonedWorld.UUID || snapshot.SnapshotName != "Origin World Save 1" {
		t.Fatalf("unexpected snapshot identity: %#v", snapshot)
	}
	if snapshot.Reason != worldCopyReasonFork || snapshot.EngineVersion != version.Version || snapshot.SchemaVersion != worldSnapshotSchemaVersion {
		t.Fatalf("unexpected snapshot metadata versions: %#v", snapshot)
	}
	if snapshot.MinCompatibleVersion != version.MinCompatibleVersion {
		t.Fatalf("unexpected min compatible version: %s", snapshot.MinCompatibleVersion)
	}
	if snapshot.NodeCount != 3 || snapshot.ComponentCount != 3 || snapshot.MemoryCount != 2 || snapshot.RelationCount != 2 {
		t.Fatalf("unexpected snapshot counts: %#v", snapshot)
	}
	if snapshot.ComponentTypesJSON != `["lore","profile","rule"]` {
		t.Fatalf("unexpected component types: %s", snapshot.ComponentTypesJSON)
	}
	if snapshot.SettingsHash == "" || snapshot.PolicyHash == "" {
		t.Fatalf("expected snapshot compatibility hashes to be populated: %#v", snapshot)
	}
	expectedHash := buildWorldSnapshotHash(originWorldID, clonedWorld.UUID, cloneSnapshotStats{
		NodeCount:      3,
		ComponentCount: 3,
		MemoryCount:    2,
		RelationCount:  2,
	})
	if snapshot.PayloadHash != expectedHash {
		t.Fatalf("unexpected snapshot hash: got %s want %s", snapshot.PayloadHash, expectedHash)
	}
}

func TestCreateWorldSnapshotUsesSnapshotReasonAndDefaultName(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Plain World", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	clonedWorld, err := CreateWorldSnapshot(world.UUID, "", false)
	if err != nil {
		t.Fatalf("create world snapshot: %v", err)
	}
	if clonedWorld.Name != "Plain World snapshot" {
		t.Fatalf("unexpected default clone name: %s", clonedWorld.Name)
	}

	snapshot, err := store.GetWorldSnapshotBySnapshotWorld(clonedWorld.UUID)
	if err != nil {
		t.Fatalf("get snapshot metadata: %v", err)
	}
	if snapshot.SnapshotName != "Plain World snapshot" {
		t.Fatalf("unexpected snapshot name: %s", snapshot.SnapshotName)
	}
	if snapshot.Reason != worldCopyReasonSnapshot {
		t.Fatalf("unexpected snapshot reason: %s", snapshot.Reason)
	}
	if snapshot.ComponentTypesJSON != `[]` {
		t.Fatalf("unexpected empty-world component types: %s", snapshot.ComponentTypesJSON)
	}
	if snapshot.SettingsHash == "" || snapshot.PolicyHash == "" {
		t.Fatalf("expected empty-world compatibility hashes to be populated: %#v", snapshot)
	}
	if snapshot.NodeCount != 0 || snapshot.ComponentCount != 0 || snapshot.MemoryCount != 0 || snapshot.RelationCount != 0 {
		t.Fatalf("unexpected empty clone counts: %#v", snapshot)
	}
}

func TestRestoreWorldCreatesRunnableCopyFromSnapshot(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Restore Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	hero, err := createNodeTx(store.DB, world.UUID, "Restored Hero", "npc", nil)
	if err != nil {
		t.Fatalf("create hero: %v", err)
	}
	if err := store.CreateMemory(&store.MemoryModel{NodeID: hero.ID, NodeUUID: hero.UUID, Content: "checkpoint", Level: "long_term", Tags: "save"}); err != nil {
		t.Fatalf("create hero memory: %v", err)
	}

	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Save Slot 1", false)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	restoredWorld, err := RestoreWorld(snapshotWorld.UUID, "Restored Session", false)
	if err != nil {
		t.Fatalf("restore world: %v", err)
	}
	if restoredWorld.Name != "Restored Session" {
		t.Fatalf("unexpected restored world name: %s", restoredWorld.Name)
	}

	restoredNodes, err := store.GetAllNodes(restoredWorld.UUID, 0, 0, "")
	if err != nil {
		t.Fatalf("get restored nodes: %v", err)
	}
	if len(restoredNodes) != 2 {
		t.Fatalf("expected restored world root + hero, got %d", len(restoredNodes))
	}

	var restoredHero store.NodeModel
	for _, node := range restoredNodes {
		if node.Name == "Restored Hero" {
			restoredHero = node
			break
		}
	}
	if restoredHero.UUID == "" {
		t.Fatal("expected restored hero node")
	}
	memories, err := store.GetNodeMemories(restoredHero.UUID, 0)
	if err != nil {
		t.Fatalf("get restored memories: %v", err)
	}
	if len(memories) != 1 || memories[0].Content != "checkpoint" {
		t.Fatalf("unexpected restored memories: %#v", memories)
	}

	restoredSnapshot, err := store.GetWorldSnapshotBySnapshotWorld(restoredWorld.UUID)
	if err != nil {
		t.Fatalf("get snapshot metadata: %v", err)
	}
	if restoredSnapshot.Reason != worldCopyReasonRestore {
		t.Fatalf("unexpected restore snapshot reason: %s", restoredSnapshot.Reason)
	}
	if restoredSnapshot.SourceWorldUUID != snapshotWorld.UUID {
		t.Fatalf("expected restore source to be snapshot world, got %s", restoredSnapshot.SourceWorldUUID)
	}
	if restoredSnapshot.SnapshotName != "Restored Session" {
		t.Fatalf("unexpected restore snapshot name: %s", restoredSnapshot.SnapshotName)
	}
}

func TestCreateWorldSnapshotAfterTickStateComponents(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Tick Snapshot World", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	location, err := createNodeTx(store.DB, world.UUID, "议事厅", "location", &world.UUID)
	if err != nil {
		t.Fatalf("create location: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: location.ID, NodeUUID: location.UUID, ComponentType: "profile", Data: "location profile"}); err != nil {
		t.Fatalf("create location profile: %v", err)
	}
	if _, err := upsertStateComponentTx(store.DB, world.UUID, engine.CompWorldState, engine.WorldStateComponent{Summary: "summary"}); err != nil {
		t.Fatalf("create world_state: %v", err)
	}
	if _, err := upsertStateComponentTx(store.DB, world.UUID, engine.CompStoryState, engine.StoryStateComponent{CurrentSituation: "situation"}); err != nil {
		t.Fatalf("create story_state: %v", err)
	}
	if _, err := upsertStateComponentTx(store.DB, world.UUID, engine.CompStoryHistory, engine.StoryHistoryComponent{}); err != nil {
		t.Fatalf("create story_history: %v", err)
	}
	if _, err := upsertStateComponentTx(store.DB, world.UUID, engine.CompWorldTimeState, engine.WorldTimeStateComponent{CurrentTimeLabel: "tick-1"}); err != nil {
		t.Fatalf("create world_time_state: %v", err)
	}
	if _, err := upsertStateComponentTx(store.DB, world.UUID, engine.CompStateSnapshot, engine.StateSnapshotComponent{SnapshotType: "world_tick", Version: "v1"}); err != nil {
		t.Fatalf("create state_snapshot: %v", err)
	}

	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Tick Snapshot Save", false)
	if err != nil {
		t.Fatalf("create snapshot after tick state components: %v", err)
	}
	if snapshotWorld == nil || snapshotWorld.UUID == "" {
		t.Fatal("expected snapshot world to be created")
	}
}

func TestRestoreWorldRejectsNonSnapshotWorld(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Fork Only World", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	forkedWorld, err := ForkWorld(world.UUID, "Forked Working Copy", false)
	if err != nil {
		t.Fatalf("fork world: %v", err)
	}

	_, err = RestoreWorld(forkedWorld.UUID, "Should Fail", false)
	if err == nil {
		t.Fatal("expected restore to reject non-snapshot world")
	}
	if !IsKind(err, ErrorInvalid) {
		t.Fatalf("expected invalid error, got %v", err)
	}
	if ErrorCode(err) != snapshotValidationCodeReasonInvalid {
		t.Fatalf("expected reason invalid code, got %s", ErrorCode(err))
	}
}

func TestRestoreWorldRejectsSnapshotDrift(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Drift Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	hero, err := createNodeTx(store.DB, world.UUID, "Hero Drift", "npc", nil)
	if err != nil {
		t.Fatalf("create hero: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: hero.ID, NodeUUID: hero.UUID, ComponentType: "profile", Data: "before save"}); err != nil {
		t.Fatalf("create component: %v", err)
	}

	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Drift Save", false)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	snapshotHeroNodes, err := store.GetAllNodes(snapshotWorld.UUID, 0, 0, "npc")
	if err != nil {
		t.Fatalf("get snapshot hero: %v", err)
	}
	if len(snapshotHeroNodes) != 1 {
		t.Fatalf("expected 1 snapshot hero node, got %d", len(snapshotHeroNodes))
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: snapshotHeroNodes[0].ID, NodeUUID: snapshotHeroNodes[0].UUID, ComponentType: "lore", Data: "mutated after save"}); err != nil {
		t.Fatalf("mutate snapshot: %v", err)
	}

	_, err = RestoreWorld(snapshotWorld.UUID, "Should Drift Fail", false)
	if err == nil {
		t.Fatal("expected restore to reject drifted snapshot")
	}
	if !IsKind(err, ErrorConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if ErrorCode(err) != snapshotValidationCodeComponentTypesDrifted {
		t.Fatalf("expected component drift code, got %s", ErrorCode(err))
	}
}

func TestValidateWorldSnapshotReportsValidity(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Validate Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	if _, err := CreateWorldSnapshot(world.UUID, "Validate Save", false); err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	snapshotWorlds, err := store.GetAllNodes("", 0, 0, "world")
	if err != nil {
		t.Fatalf("get worlds: %v", err)
	}
	var snapshotWorldID string
	for _, node := range snapshotWorlds {
		if node.Name == "Validate Save" {
			snapshotWorldID = node.UUID
			break
		}
	}
	if snapshotWorldID == "" {
		t.Fatal("expected snapshot world to exist")
	}

	result, err := ValidateWorldSnapshot(snapshotWorldID)
	if err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid snapshot result: %#v", result)
	}
	if result.Reason != worldCopyReasonSnapshot {
		t.Fatalf("unexpected snapshot reason: %s", result.Reason)
	}
	if result.CurrentEngineVersion != version.Version || result.CurrentMinCompatibleVersion != version.MinCompatibleVersion {
		t.Fatalf("unexpected runtime version info: %#v", result)
	}
	if len(result.Issues) != 0 {
		t.Fatalf("expected no issues: %#v", result.Issues)
	}
}

func TestValidateWorldSnapshotReportsDriftAndReasonIssues(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Validate Drift", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	forkedWorld, err := ForkWorld(world.UUID, "Forked Working Copy", false)
	if err != nil {
		t.Fatalf("fork world: %v", err)
	}

	forkResult, err := ValidateWorldSnapshot(forkedWorld.UUID)
	if err != nil {
		t.Fatalf("validate fork snapshot metadata: %v", err)
	}
	if forkResult.Valid {
		t.Fatalf("expected fork validation to be invalid: %#v", forkResult)
	}
	if len(forkResult.Issues) == 0 || forkResult.Issues[0].Code != snapshotValidationCodeReasonInvalid {
		t.Fatalf("expected reason invalid issue: %#v", forkResult.Issues)
	}

	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Drift Save", false)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	snapshotHero, err := createNodeTx(store.DB, snapshotWorld.UUID, "Snapshot Hero", "npc", nil)
	if err != nil {
		t.Fatalf("create snapshot hero: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: snapshotHero.ID, NodeUUID: snapshotHero.UUID, ComponentType: "profile", Data: "drifted"}); err != nil {
		t.Fatalf("mutate snapshot: %v", err)
	}

	driftResult, err := ValidateWorldSnapshot(snapshotWorld.UUID)
	if err != nil {
		t.Fatalf("validate drift snapshot: %v", err)
	}
	if driftResult.Valid {
		t.Fatalf("expected drift validation to be invalid: %#v", driftResult)
	}
	found := false
	for _, issue := range driftResult.Issues {
		if issue.Code == snapshotValidationCodeComponentTypesDrifted {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected component drift issue: %#v", driftResult.Issues)
	}
}

func TestListWorldSnapshotsReturnsOnlySaveSnapshots(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Snapshot Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	if _, err := ForkWorld(world.UUID, "Working Copy", false); err != nil {
		t.Fatalf("fork world: %v", err)
	}
	if _, err := CreateWorldSnapshot(world.UUID, "Save Slot A", false); err != nil {
		t.Fatalf("create snapshot A: %v", err)
	}
	if _, err := CreateWorldSnapshot(world.UUID, "Save Slot B", false); err != nil {
		t.Fatalf("create snapshot B: %v", err)
	}

	result, err := ListWorldSnapshots(world.UUID)
	if err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 save snapshots, got %d", len(result))
	}
	if result[0].Reason != worldCopyReasonSnapshot || result[1].Reason != worldCopyReasonSnapshot {
		t.Fatalf("expected only save snapshots: %#v", result)
	}
	if result[0].SnapshotName != "Save Slot B" || result[1].SnapshotName != "Save Slot A" {
		t.Fatalf("expected snapshots ordered newest-first, got %#v", result)
	}
	if result[0].Status != "valid" || !result[0].Restorable || result[1].Status != "valid" || !result[1].Restorable {
		t.Fatalf("expected listed snapshots to be valid and restorable, got %#v", result)
	}
}

func TestGetWorldSnapshotMetadataReturnsDecodedInfo(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Metadata Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	hero, err := createNodeTx(store.DB, world.UUID, "Metadata Hero", "npc", nil)
	if err != nil {
		t.Fatalf("create hero: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: hero.ID, NodeUUID: hero.UUID, ComponentType: "profile", Data: "meta"}); err != nil {
		t.Fatalf("create component: %v", err)
	}

	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Metadata Save", false)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	result, err := GetWorldSnapshotMetadata(snapshotWorld.UUID)
	if err != nil {
		t.Fatalf("get snapshot metadata: %v", err)
	}
	if result.SnapshotWorldID != snapshotWorld.UUID || result.SourceWorldID != world.UUID {
		t.Fatalf("unexpected snapshot identity: %#v", result)
	}
	if result.SnapshotName != "Metadata Save" || result.Reason != worldCopyReasonSnapshot {
		t.Fatalf("unexpected snapshot info: %#v", result)
	}
	if len(result.ComponentTypes) != 1 || result.ComponentTypes[0] != "profile" {
		t.Fatalf("expected decoded component types: %#v", result.ComponentTypes)
	}
	if result.Status != "valid" || !result.Restorable {
		t.Fatalf("expected valid restorable snapshot info: %#v", result)
	}
}

func TestGetWorldSnapshotMetadataDecoratesWorkingCopyAndRestoredCopy(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Metadata Copy Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	forkedWorld, err := ForkWorld(world.UUID, "Metadata Working Copy", false)
	if err != nil {
		t.Fatalf("fork world: %v", err)
	}
	forkInfo, err := GetWorldSnapshotMetadata(forkedWorld.UUID)
	if err != nil {
		t.Fatalf("get fork metadata: %v", err)
	}
	if forkInfo.Status != "working_copy" || forkInfo.Restorable {
		t.Fatalf("expected working copy metadata decoration, got %#v", forkInfo)
	}

	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Restore Metadata Save", false)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	restoredWorld, err := RestoreWorld(snapshotWorld.UUID, "Metadata Restored Copy", false)
	if err != nil {
		t.Fatalf("restore world: %v", err)
	}
	restoredInfo, err := GetWorldSnapshotMetadata(restoredWorld.UUID)
	if err != nil {
		t.Fatalf("get restored metadata: %v", err)
	}
	if restoredInfo.Status != "restored_copy" || restoredInfo.Restorable {
		t.Fatalf("expected restored copy metadata decoration, got %#v", restoredInfo)
	}
}

func TestListWorldSnapshotsDecoratesInvalidSnapshots(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Snapshot Drift Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	hero, err := createNodeTx(store.DB, world.UUID, "Snapshot Drift Hero", "npc", nil)
	if err != nil {
		t.Fatalf("create hero: %v", err)
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: hero.ID, NodeUUID: hero.UUID, ComponentType: "profile", Data: "before drift"}); err != nil {
		t.Fatalf("create component: %v", err)
	}

	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Drifted Slot", false)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	snapshotHeroNodes, err := store.GetAllNodes(snapshotWorld.UUID, 0, 0, "npc")
	if err != nil {
		t.Fatalf("get snapshot hero: %v", err)
	}
	if len(snapshotHeroNodes) != 1 {
		t.Fatalf("expected 1 snapshot hero node, got %d", len(snapshotHeroNodes))
	}
	if err := store.CreateComponent(&store.ComponentModel{NodeID: snapshotHeroNodes[0].ID, NodeUUID: snapshotHeroNodes[0].UUID, ComponentType: "lore", Data: "after drift"}); err != nil {
		t.Fatalf("mutate snapshot: %v", err)
	}

	result, err := ListWorldSnapshots(world.UUID)
	if err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(result))
	}
	if result[0].Status != "invalid" || result[0].Restorable {
		t.Fatalf("expected invalid non-restorable snapshot, got %#v", result[0])
	}
	if len(result[0].ValidationIssues) == 0 || result[0].ValidationIssues[0] != snapshotValidationCodeComponentTypesDrifted {
		t.Fatalf("expected component drift validation issue, got %#v", result[0].ValidationIssues)
	}
}

func TestDeleteWorldSnapshotRemovesSnapshotWorldAndMetadata(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Delete Snapshot Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}
	snapshotWorld, err := CreateWorldSnapshot(world.UUID, "Delete Me", false)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	if err := DeleteWorldSnapshot(snapshotWorld.UUID); err != nil {
		t.Fatalf("delete snapshot: %v", err)
	}
	if _, err := store.GetWorldSnapshotBySnapshotWorld(snapshotWorld.UUID); err == nil {
		t.Fatal("expected snapshot metadata to be deleted")
	}
	if _, err := store.GetNode(snapshotWorld.UUID); err == nil {
		t.Fatal("expected snapshot world node to be deleted")
	}
	listed, err := ListWorldSnapshots(world.UUID)
	if err != nil {
		t.Fatalf("list snapshots after delete: %v", err)
	}
	if len(listed) != 0 {
		t.Fatalf("expected no remaining snapshots after delete, got %#v", listed)
	}
	if _, err := ValidateWorldSnapshot(snapshotWorld.UUID); !IsKind(err, ErrorNotFound) {
		t.Fatalf("expected validate to report not found after delete, got %v", err)
	}
	if _, err := GetWorldSnapshotMetadata(snapshotWorld.UUID); !IsKind(err, ErrorNotFound) {
		t.Fatalf("expected metadata lookup to report not found after delete, got %v", err)
	}
}

func TestForkWorldCopiesDeepHierarchyAcrossBatches(t *testing.T) {
	initImportExportTestDB(t)

	world, err := createNodeTx(store.DB, "", "Batch Hierarchy Source", "world", nil)
	if err != nil {
		t.Fatalf("create world: %v", err)
	}

	var lastRoot *store.NodeModel
	for i := 0; i < worldCopyBatchSize+5; i++ {
		root, err := createNodeTx(store.DB, world.UUID, fmt.Sprintf("District %03d", i), "location", nil)
		if err != nil {
			t.Fatalf("create root node %d: %v", i, err)
		}
		lastRoot = root
	}
	if lastRoot == nil {
		t.Fatal("expected last root node to exist")
	}

	child, err := createNodeTx(store.DB, world.UUID, "District Archive", "item", &lastRoot.UUID)
	if err != nil {
		t.Fatalf("create child node: %v", err)
	}
	grandchild, err := createNodeTx(store.DB, world.UUID, "Archive Key", "item", &child.UUID)
	if err != nil {
		t.Fatalf("create grandchild node: %v", err)
	}
	if err := store.CreateMemory(&store.MemoryModel{NodeID: grandchild.ID, NodeUUID: grandchild.UUID, Content: "deep branch memory", Level: "long_term", Tags: "batch"}); err != nil {
		t.Fatalf("create grandchild memory: %v", err)
	}

	clonedWorld, err := ForkWorld(world.UUID, "Batch Hierarchy Fork", false)
	if err != nil {
		t.Fatalf("fork world: %v", err)
	}

	clonedNodes, err := store.GetAllNodes(clonedWorld.UUID, 0, 0, "")
	if err != nil {
		t.Fatalf("get cloned nodes: %v", err)
	}
	expectedNodeCount := worldCopyBatchSize + 8
	if len(clonedNodes) != expectedNodeCount {
		t.Fatalf("expected %d cloned nodes including world root, got %d", expectedNodeCount, len(clonedNodes))
	}

	clonedByName := make(map[string]store.NodeModel, len(clonedNodes))
	for _, node := range clonedNodes {
		clonedByName[node.Name] = node
	}
	clonedLastRoot, ok := clonedByName[lastRoot.Name]
	if !ok {
		t.Fatalf("expected cloned root node %q", lastRoot.Name)
	}
	clonedChild, ok := clonedByName[child.Name]
	if !ok {
		t.Fatalf("expected cloned child node %q", child.Name)
	}
	clonedGrandchild, ok := clonedByName[grandchild.Name]
	if !ok {
		t.Fatalf("expected cloned grandchild node %q", grandchild.Name)
	}
	if clonedChild.ParentUUID == nil || *clonedChild.ParentUUID != clonedLastRoot.UUID {
		t.Fatalf("expected cloned child parent to map to cloned root, got %v", clonedChild.ParentUUID)
	}
	if clonedGrandchild.ParentUUID == nil || *clonedGrandchild.ParentUUID != clonedChild.UUID {
		t.Fatalf("expected cloned grandchild parent to map to cloned child, got %v", clonedGrandchild.ParentUUID)
	}

	grandchildMemories, err := store.GetNodeMemories(clonedGrandchild.UUID, 0)
	if err != nil {
		t.Fatalf("get cloned grandchild memories: %v", err)
	}
	if len(grandchildMemories) != 1 || grandchildMemories[0].Content != "deep branch memory" || grandchildMemories[0].NodeUUID != clonedGrandchild.UUID {
		t.Fatalf("unexpected cloned grandchild memories: %#v", grandchildMemories)
	}
}
