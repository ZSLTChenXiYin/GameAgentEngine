package workerstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadFile(path string) (*WorldState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return LoadYAML(data)
	case ".json":
		return LoadJSON(data)
	default:
		return nil, fmt.Errorf("unsupported state file extension: %s", ext)
	}
}

func LoadYAML(data []byte) (*WorldState, error) {
	var state WorldState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return normalizeWorldState(&state), nil
}

func LoadJSON(data []byte) (*WorldState, error) {
	var state WorldState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return normalizeWorldState(&state), nil
}

func normalizeWorldState(state *WorldState) *WorldState {
	if state == nil {
		state = &WorldState{}
	}
	if state.Meta == nil {
		state.Meta = map[string]any{}
	}
	if state.Actors == nil {
		state.Actors = map[string]*ActorState{}
	}
	if state.Scenes == nil {
		state.Scenes = map[string]*SceneState{}
	}
	if state.Items == nil {
		state.Items = map[string]*ItemState{}
	}
	if state.Tasks == nil {
		state.Tasks = map[string]*QuestState{}
	}
	for id, actor := range state.Actors {
		if actor == nil {
			actor = &ActorState{}
			state.Actors[id] = actor
		}
		if strings.TrimSpace(actor.ID) == "" {
			actor.ID = id
		}
		if actor.QuestStates == nil {
			actor.QuestStates = map[string]string{}
		}
		if actor.Flags == nil {
			actor.Flags = map[string]any{}
		}
	}
	for id, scene := range state.Scenes {
		if scene == nil {
			scene = &SceneState{}
			state.Scenes[id] = scene
		}
		if strings.TrimSpace(scene.ID) == "" {
			scene.ID = id
		}
		if scene.Flags == nil {
			scene.Flags = map[string]any{}
		}
	}
	for id, item := range state.Items {
		if item == nil {
			item = &ItemState{}
			state.Items[id] = item
		}
		if strings.TrimSpace(item.ID) == "" {
			item.ID = id
		}
		if item.Metadata == nil {
			item.Metadata = map[string]any{}
		}
	}
	for id, task := range state.Tasks {
		if task == nil {
			task = &QuestState{}
			state.Tasks[id] = task
		}
		if strings.TrimSpace(task.ID) == "" {
			task.ID = id
		}
		if task.Flags == nil {
			task.Flags = map[string]any{}
		}
	}
	return state
}
