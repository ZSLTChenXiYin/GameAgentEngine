package workerstate

import "testing"

func TestStateViewQueriesAuthorityFields(t *testing.T) {
	view := NewStateView(&WorldState{
		WorldID: "demo_world",
		Actors: map[string]*ActorState{
			"player_001": {
				ID:         "player_001",
				Name:       "Hero",
				Kind:       "player",
				HP:         14,
				MaxHP:      20,
				Money:      77,
				LocationID: "scene_inn",
				Inventory: []InventoryEntry{
					{ItemID: "knife_bloody", Quantity: 1},
					{ItemID: "coin", Quantity: 77},
				},
			},
		},
		Scenes: map[string]*SceneState{
			"scene_inn": {ID: "scene_inn", Occupants: []string{"player_001", "npc_innkeeper"}},
		},
		Tasks: map[string]*QuestState{
			"task_murder": {ID: "task_murder", Status: "active", Stage: "investigate"},
		},
	})

	if got := view.WorldID(); got != "demo_world" {
		t.Fatalf("expected world id demo_world, got %q", got)
	}
	hp, maxHP, ok := view.ActorHP("player_001")
	if !ok || hp != 14 || maxHP != 20 {
		t.Fatalf("unexpected hp query result: hp=%d max=%d ok=%v", hp, maxHP, ok)
	}
	money, ok := view.ActorMoney("player_001")
	if !ok || money != 77 {
		t.Fatalf("unexpected money query result: %d ok=%v", money, ok)
	}
	locationID, ok := view.ActorLocation("player_001")
	if !ok || locationID != "scene_inn" {
		t.Fatalf("unexpected location query result: %q ok=%v", locationID, ok)
	}
	if !view.ItemPresentOnActor("player_001", "knife_bloody") {
		t.Fatal("expected knife_bloody to be present")
	}
	occupants := view.SceneOccupants("scene_inn")
	if len(occupants) != 2 || occupants[0] != "player_001" {
		t.Fatalf("unexpected occupants: %#v", occupants)
	}
	status, stage, ok := view.QuestStatus("task_murder")
	if !ok || status != "active" || stage != "investigate" {
		t.Fatalf("unexpected quest status: status=%q stage=%q ok=%v", status, stage, ok)
	}
}
