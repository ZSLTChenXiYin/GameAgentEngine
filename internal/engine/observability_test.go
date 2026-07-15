package engine

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func TestBuildContextLogDetailIncludesInteractionFields(t *testing.T) {
	started := time.Now().Add(-25 * time.Millisecond)
	detail := buildContextLogDetail(&BuiltContext{
		Node:             &store.NodeModel{UUID: "npc_focus"},
		Components:       []store.ComponentModel{{UUID: "c1"}},
		Memories:         []store.MemoryModel{{UUID: "m1"}},
		Relations:        []store.RelationModel{{UUID: "r1"}},
		Children:         []store.NodeModel{{UUID: "child1"}},
		Ancestors:        []store.NodeModel{{UUID: "ancestor1"}},
		SpeakerNode:      &store.NodeModel{UUID: "player_1"},
		TargetNode:       &store.NodeModel{UUID: "npc_target"},
		SceneNode:        &store.NodeModel{UUID: "scene_tavern"},
		ParticipantNodes: []store.NodeModel{{UUID: "player_1"}, {UUID: "npc_target"}, {UUID: "npc_guard"}},
		Interaction:      &InteractionContext{Mode: "group_chat"},
		SystemPrompt:     "prompt body",
	}, started)

	if detail == "" {
		t.Fatal("expected context log detail")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(detail), &payload); err != nil {
		t.Fatalf("unmarshal detail: %v", err)
	}
	if got := payload["node_id"]; got != "npc_focus" {
		t.Fatalf("expected node_id npc_focus, got %#v", got)
	}
	if got := payload["speaker_node_id"]; got != "player_1" {
		t.Fatalf("expected speaker_node_id player_1, got %#v", got)
	}
	if got := payload["target_node_id"]; got != "npc_target" {
		t.Fatalf("expected target_node_id npc_target, got %#v", got)
	}
	if got := payload["scene_node_id"]; got != "scene_tavern" {
		t.Fatalf("expected scene_node_id scene_tavern, got %#v", got)
	}
	if got := payload["interaction_mode"]; got != "group_chat" {
		t.Fatalf("expected interaction_mode group_chat, got %#v", got)
	}
	if got := payload["participant_count"]; got != float64(3) {
		t.Fatalf("expected participant_count 3, got %#v", got)
	}
	if _, ok := payload["built_at"]; !ok {
		t.Fatal("expected built_at in context detail")
	}
}
