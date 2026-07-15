package workerstate

import "strings"

func (v *StateView) WorldID() string {
	if v == nil || v.state == nil {
		return ""
	}
	return v.state.WorldID
}

func (v *StateView) Actor(id string) (*ActorState, bool) {
	if v == nil || v.state == nil {
		return nil, false
	}
	actor, ok := v.state.Actors[strings.TrimSpace(id)]
	return actor, ok
}

func (v *StateView) Actors() []*ActorState {
	if v == nil || v.state == nil {
		return nil
	}
	actors := make([]*ActorState, 0, len(v.state.Actors))
	for _, actor := range v.state.Actors {
		if actor != nil {
			actors = append(actors, actor)
		}
	}
	return actors
}

func (v *StateView) Scene(id string) (*SceneState, bool) {
	if v == nil || v.state == nil {
		return nil, false
	}
	scene, ok := v.state.Scenes[strings.TrimSpace(id)]
	return scene, ok
}

func (v *StateView) Scenes() []*SceneState {
	if v == nil || v.state == nil {
		return nil
	}
	scenes := make([]*SceneState, 0, len(v.state.Scenes))
	for _, scene := range v.state.Scenes {
		if scene != nil {
			scenes = append(scenes, scene)
		}
	}
	return scenes
}

func (v *StateView) Task(id string) (*QuestState, bool) {
	if v == nil || v.state == nil {
		return nil, false
	}
	task, ok := v.state.Tasks[strings.TrimSpace(id)]
	return task, ok
}

func (v *StateView) ActorInventory(actorID string) []InventoryEntry {
	actor, ok := v.Actor(actorID)
	if !ok || actor == nil {
		return nil
	}
	return append([]InventoryEntry(nil), actor.Inventory...)
}

func (v *StateView) ActorMoney(actorID string) (int, bool) {
	actor, ok := v.Actor(actorID)
	if !ok || actor == nil {
		return 0, false
	}
	return actor.Money, true
}

func (v *StateView) ActorHP(actorID string) (int, int, bool) {
	actor, ok := v.Actor(actorID)
	if !ok || actor == nil {
		return 0, 0, false
	}
	return actor.HP, actor.MaxHP, true
}

func (v *StateView) ActorLocation(actorID string) (string, bool) {
	actor, ok := v.Actor(actorID)
	if !ok || actor == nil || strings.TrimSpace(actor.LocationID) == "" {
		return "", false
	}
	return actor.LocationID, true
}

func (v *StateView) ActorsAtScene(sceneID string) []*ActorState {
	trimmed := strings.TrimSpace(sceneID)
	if trimmed == "" {
		return nil
	}
	actors := make([]*ActorState, 0)
	for _, actor := range v.Actors() {
		if actor != nil && strings.TrimSpace(actor.LocationID) == trimmed {
			actors = append(actors, actor)
		}
	}
	return actors
}

func (v *StateView) FindActorIDByName(name string) (string, bool) {
	needle := strings.ToLower(strings.TrimSpace(name))
	if needle == "" {
		return "", false
	}
	for id, actor := range v.state.Actors {
		if actor == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(id), needle) || strings.EqualFold(strings.TrimSpace(actor.ID), needle) || strings.EqualFold(strings.TrimSpace(actor.Name), needle) {
			return actor.ID, true
		}
	}
	return "", false
}

func (v *StateView) FindSceneIDByName(name string) (string, bool) {
	needle := strings.ToLower(strings.TrimSpace(name))
	if needle == "" {
		return "", false
	}
	for id, scene := range v.state.Scenes {
		if scene == nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(id), needle) || strings.EqualFold(strings.TrimSpace(scene.ID), needle) || strings.EqualFold(strings.TrimSpace(scene.Name), needle) {
			return scene.ID, true
		}
	}
	return "", false
}

func (v *StateView) SceneOccupants(sceneID string) []string {
	scene, ok := v.Scene(sceneID)
	if !ok || scene == nil {
		return nil
	}
	return append([]string(nil), scene.Occupants...)
}

func (v *StateView) ItemPresentOnActor(actorID, itemID string) bool {
	for _, entry := range v.ActorInventory(actorID) {
		if entry.ItemID == strings.TrimSpace(itemID) && entry.Quantity > 0 {
			return true
		}
	}
	return false
}

func (v *StateView) QuestStatus(taskID string) (string, string, bool) {
	task, ok := v.Task(taskID)
	if !ok || task == nil {
		return "", "", false
	}
	return task.Status, task.Stage, true
}
