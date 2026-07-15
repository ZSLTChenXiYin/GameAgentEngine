package workerstate

import "testing"

func TestLoadYAMLNormalizesState(t *testing.T) {
	state, err := LoadYAML([]byte(`
world_id: demo_world
actors:
  player_001:
    name: Hero
    kind: player
    hp: 12
    max_hp: 18
    money: 45
    location_id: scene_inn
    inventory:
      - item_id: knife_bloody
        quantity: 1
scenes:
  scene_inn:
    name: Old Inn
    occupants: [player_001, npc_innkeeper]
tasks:
  task_murder:
    status: active
    stage: investigate
`))
	if err != nil {
		t.Fatalf("load yaml: %v", err)
	}
	if state.WorldID != "demo_world" {
		t.Fatalf("expected world_id demo_world, got %q", state.WorldID)
	}
	if state.Actors["player_001"].ID != "player_001" {
		t.Fatalf("expected normalized actor id, got %#v", state.Actors["player_001"])
	}
	if state.Scenes["scene_inn"].ID != "scene_inn" {
		t.Fatalf("expected normalized scene id, got %#v", state.Scenes["scene_inn"])
	}
	if state.Tasks["task_murder"].ID != "task_murder" {
		t.Fatalf("expected normalized task id, got %#v", state.Tasks["task_murder"])
	}
}

func TestLoadJSONNormalizesEmptyMaps(t *testing.T) {
	state, err := LoadJSON([]byte(`{"world_id":"w1"}`))
	if err != nil {
		t.Fatalf("load json: %v", err)
	}
	if state.Actors == nil || state.Scenes == nil || state.Items == nil || state.Tasks == nil {
		t.Fatalf("expected initialized maps, got %#v", state)
	}
}
