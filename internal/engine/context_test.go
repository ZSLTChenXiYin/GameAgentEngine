package engine

import (
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func TestContextBuilderBuildIncludesInteractionView(t *testing.T) {
	initTestDB(t)

	world := &store.NodeModel{UUID: store.NewUUID(), Name: "World", NodeType: "world"}
	if err := store.CreateNode(world); err != nil {
		t.Fatalf("create world: %v", err)
	}
	if err := store.DB.Model(world).Update("world_id", world.ID).Error; err != nil {
		t.Fatalf("update world id: %v", err)
	}

	scene := &store.NodeModel{UUID: store.NewUUID(), Name: "Tavern", NodeType: "location", WorldID: world.ID, WorldUUID: world.UUID, ParentID: &world.ID, ParentUUID: &world.UUID}
	if err := store.CreateNode(scene); err != nil {
		t.Fatalf("create scene: %v", err)
	}

	npc := &store.NodeModel{UUID: store.NewUUID(), Name: "Innkeeper", NodeType: "npc", WorldID: world.ID, WorldUUID: world.UUID, ParentID: &scene.ID, ParentUUID: &scene.UUID}
	if err := store.CreateNode(npc); err != nil {
		t.Fatalf("create npc: %v", err)
	}

	player := &store.NodeModel{UUID: store.NewUUID(), Name: "Player", NodeType: "npc", WorldID: world.ID, WorldUUID: world.UUID, ParentID: &scene.ID, ParentUUID: &scene.UUID}
	if err := store.CreateNode(player); err != nil {
		t.Fatalf("create player: %v", err)
	}

	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, SourceID: npc.ID, SourceUUID: npc.UUID, TargetID: scene.ID, TargetUUID: scene.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create npc located_at: %v", err)
	}
	if err := store.CreateRelation(&store.RelationModel{UUID: store.NewUUID(), WorldID: world.ID, WorldUUID: world.UUID, SourceID: player.ID, SourceUUID: player.UUID, TargetID: scene.ID, TargetUUID: scene.UUID, RelationType: string(RelLocatedAt), Weight: 1}); err != nil {
		t.Fatalf("create player located_at: %v", err)
	}

	builder := NewContextBuilder()
	ctx, err := builder.Build(TaskNPCDialogue, npc.UUID, 3, 20, true, &InteractionContext{
		Mode:               "direct_dialogue",
		SpeakerNodeID:      player.UUID,
		TargetNodeID:       npc.UUID,
		SceneNodeID:        scene.UUID,
		RoomID:             "room_tavern_main",
		ParticipantNodeIDs: []string{player.UUID, npc.UUID},
		AudienceScope:      "public",
		TurnIndex:          4,
		Event:              &InteractionEvent{Type: "speech"},
	})
	if err != nil {
		t.Fatalf("build context: %v", err)
	}
	if ctx.Interaction == nil {
		t.Fatal("expected interaction in built context")
	}
	if ctx.SpeakerNode == nil || ctx.SpeakerNode.UUID != player.UUID {
		t.Fatalf("expected speaker node %s, got %#v", player.UUID, ctx.SpeakerNode)
	}
	if ctx.TargetNode == nil || ctx.TargetNode.UUID != npc.UUID {
		t.Fatalf("expected target node %s, got %#v", npc.UUID, ctx.TargetNode)
	}
	if ctx.SceneNode == nil || ctx.SceneNode.UUID != scene.UUID {
		t.Fatalf("expected scene node %s, got %#v", scene.UUID, ctx.SceneNode)
	}
	if len(ctx.ParticipantNodes) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(ctx.ParticipantNodes))
	}
	if !strings.Contains(ctx.SystemPrompt, "交互语义：") {
		t.Fatalf("expected interaction prompt block, got %s", ctx.SystemPrompt)
	}
	if !strings.Contains(ctx.SystemPrompt, "[speaker] Player") {
		t.Fatalf("expected speaker in prompt, got %s", ctx.SystemPrompt)
	}
	if !strings.Contains(ctx.SystemPrompt, "[scene] Tavern") {
		t.Fatalf("expected scene in prompt, got %s", ctx.SystemPrompt)
	}
}

func TestContextBuilderBuildFallsBackTargetToFocusNode(t *testing.T) {
	initTestDB(t)
	worldID, nodeID := createWorldAndNode(t)
	_ = worldID

	builder := NewContextBuilder()
	ctx, err := builder.Build(TaskCustom, nodeID, 2, 10, false, &InteractionContext{
		Mode:          "trade_dialogue",
		SpeakerNodeID: nodeID,
		TargetNodeID:  "",
	})
	if err != nil {
		t.Fatalf("build context: %v", err)
	}
	if ctx.TargetNode == nil || ctx.TargetNode.UUID != nodeID {
		t.Fatalf("expected target fallback to focus node %s, got %#v", nodeID, ctx.TargetNode)
	}
}
