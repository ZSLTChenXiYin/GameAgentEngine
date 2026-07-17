package engine

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// DemoAuthorityData represents the authority state loaded from a demo YAML file.
// It mirrors the structure in workerstate but is kept independent to avoid
// creating a dependency from engine on workerstate.
type DemoAuthorityData struct {
	WorldID string                       `yaml:"world_id"`
	Actors  map[string]*DemoActorState   `yaml:"actors"`
	Scenes  map[string]*DemoSceneState   `yaml:"scenes"`
	Items   map[string]*DemoItemState    `yaml:"items"`
	Tasks   map[string]*DemoQuestState   `yaml:"tasks"`
}

type DemoActorState struct {
	ID          string            `yaml:"id"`
	Name        string            `yaml:"name"`
	Kind        string            `yaml:"kind"`
	HP          int               `yaml:"hp"`
	MaxHP       int               `yaml:"max_hp"`
	Money       int               `yaml:"money"`
	LocationID  string            `yaml:"location_id"`
	Inventory   []DemoInventoryEntry `yaml:"inventory"`
	QuestStates map[string]string `yaml:"quest_states"`
}

type DemoInventoryEntry struct {
	ItemID   string `yaml:"item_id"`
	Quantity int    `yaml:"quantity"`
}

type DemoSceneState struct {
	ID          string         `yaml:"id"`
	Name        string         `yaml:"name"`
	Kind        string         `yaml:"kind"`
	Occupants   []string       `yaml:"occupants"`
	Flags       map[string]any `yaml:"flags"`
	Description string         `yaml:"description"`
}

type DemoItemState struct {
	ID      string         `yaml:"id"`
	Name    string         `yaml:"name"`
	OwnerID string         `yaml:"owner_id"`
	SceneID string         `yaml:"scene_id"`
	Metadata map[string]any `yaml:"metadata"`
}

type DemoQuestState struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Status  string `yaml:"status"`
	OwnerID string `yaml:"owner_id"`
	Stage   string `yaml:"stage"`
}

// LoadDemoAuthorityFile reads a YAML file and returns structured authority data.
// Returns nil without error if the file does not exist (non-demo mode).
func LoadDemoAuthorityFile(path string) (*DemoAuthorityData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read demo authority: %w", err)
	}
	var auth DemoAuthorityData
	if err := yaml.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("parse demo authority: %w", err)
	}
	return normalizeDemoAuthority(&auth), nil
}

func normalizeDemoAuthority(auth *DemoAuthorityData) *DemoAuthorityData {
	if auth == nil {
		return nil
	}
	if auth.Actors == nil {
		auth.Actors = map[string]*DemoActorState{}
	}
	if auth.Scenes == nil {
		auth.Scenes = map[string]*DemoSceneState{}
	}
	if auth.Items == nil {
		auth.Items = map[string]*DemoItemState{}
	}
	if auth.Tasks == nil {
		auth.Tasks = map[string]*DemoQuestState{}
	}
	for id, actor := range auth.Actors {
		if actor == nil {
			actor = &DemoActorState{}
			auth.Actors[id] = actor
		}
		if actor.ID == "" {
			actor.ID = id
		}
	}
	for id, scene := range auth.Scenes {
		if scene == nil {
			scene = &DemoSceneState{}
			auth.Scenes[id] = scene
		}
		if scene.ID == "" {
			scene.ID = id
		}
	}
	for id, item := range auth.Items {
		if item == nil {
			item = &DemoItemState{}
			auth.Items[id] = item
		}
		if item.ID == "" {
			item.ID = id
		}
	}
	return auth
}

// BuildDemoAuthorityBlock generates a formatted authority-fact text block
// from the demo authority data, suitable for inclusion in world tick prompts.
func BuildDemoAuthorityBlock(auth *DemoAuthorityData) string {
	if auth == nil {
		return ""
	}
	var parts []string
	parts = append(parts, "========== Demo Authority State ==========")
	parts = append(parts, "The following pre-loaded authority facts are available for this world tick.")
	parts = append(parts, "These facts are authoritative for the current demo session.")
	parts = append(parts, "")

	// Actors
	if len(auth.Actors) > 0 {
		parts = append(parts, "--- Characters ---")
		actorKeys := make([]string, 0, len(auth.Actors))
		for k := range auth.Actors {
			actorKeys = append(actorKeys, k)
		}
		sort.Strings(actorKeys)
		for _, id := range actorKeys {
			actor := auth.Actors[id]
			line := fmt.Sprintf("- %s (%s)", actorDisplayName(actor), id)
			if actor.MaxHP > 0 {
				line += fmt.Sprintf(" [HP: %d/%d]", actor.HP, actor.MaxHP)
			}
			if actor.Money > 0 {
				line += fmt.Sprintf(" [Money: %d]", actor.Money)
			}
			if actor.LocationID != "" {
				sceneName := actor.LocationID
				if s, ok := auth.Scenes[actor.LocationID]; ok && s != nil && s.Name != "" {
					sceneName = s.Name
				}
				line += fmt.Sprintf(" [Location: %s]", sceneName)
			}
			if len(actor.Inventory) > 0 {
				invParts := make([]string, 0, len(actor.Inventory))
				for _, inv := range actor.Inventory {
					itemName := inv.ItemID
					if item, ok := auth.Items[inv.ItemID]; ok && item != nil && item.Name != "" {
						itemName = item.Name
					}
					invParts = append(invParts, fmt.Sprintf("%s x%d", itemName, inv.Quantity))
				}
				line += fmt.Sprintf(" [Inventory: %s]", strings.Join(invParts, ", "))
			}
			if len(actor.QuestStates) > 0 {
				qs := make([]string, 0, len(actor.QuestStates))
				for qID, status := range actor.QuestStates {
					taskName := qID
					if task, ok := auth.Tasks[qID]; ok && task != nil && task.Name != "" {
						taskName = task.Name
					}
					qs = append(qs, fmt.Sprintf("%s=%s", taskName, status))
				}
				sort.Strings(qs)
				line += fmt.Sprintf(" [Quests: %s]", strings.Join(qs, ", "))
			}
			parts = append(parts, line)
		}
		parts = append(parts, "")
	}

	// Scenes
	if len(auth.Scenes) > 0 {
		parts = append(parts, "--- Scenes ---")
		sceneKeys := make([]string, 0, len(auth.Scenes))
		for k := range auth.Scenes {
			sceneKeys = append(sceneKeys, k)
		}
		sort.Strings(sceneKeys)
		for _, id := range sceneKeys {
			scene := auth.Scenes[id]
			line := fmt.Sprintf("- %s (%s) [%s]", scene.Name, id, scene.Kind)
			if scene.Description != "" {
				line += fmt.Sprintf(": %s", truncateString(scene.Description, 100))
			}
			if len(scene.Occupants) > 0 {
				occ := make([]string, 0, len(scene.Occupants))
				for _, oid := range scene.Occupants {
					occ = append(occ, actorDisplayName(auth.Actors[oid]))
				}
				line += fmt.Sprintf(" [Occupants: %s]", strings.Join(occ, ", "))
			}
			if len(scene.Flags) > 0 {
				flagParts := make([]string, 0, len(scene.Flags))
				for k, v := range scene.Flags {
					flagParts = append(flagParts, fmt.Sprintf("%s=%v", k, v))
				}
				sort.Strings(flagParts)
				line += fmt.Sprintf(" [Flags: %s]", strings.Join(flagParts, ", "))
			}
			parts = append(parts, line)
		}
		parts = append(parts, "")
	}

	// Items
	if len(auth.Items) > 0 {
		parts = append(parts, "--- Items ---")
		itemKeys := make([]string, 0, len(auth.Items))
		for k := range auth.Items {
			itemKeys = append(itemKeys, k)
		}
		sort.Strings(itemKeys)
		for _, id := range itemKeys {
			item := auth.Items[id]
			ownerName := item.OwnerID
			if actor, ok := auth.Actors[item.OwnerID]; ok && actor != nil && actor.Name != "" {
				ownerName = actor.Name
			}
			line := fmt.Sprintf("- %s (%s) [Owner: %s]", itemDisplayName(item), id, ownerName)
			if len(item.Metadata) > 0 {
				metaParts := make([]string, 0, len(item.Metadata))
				for k, v := range item.Metadata {
					metaParts = append(metaParts, fmt.Sprintf("%s=%v", k, v))
				}
				sort.Strings(metaParts)
				line += fmt.Sprintf(" [%s]", strings.Join(metaParts, ", "))
			}
			parts = append(parts, line)
		}
		parts = append(parts, "")
	}

	// Tasks
	if len(auth.Tasks) > 0 {
		parts = append(parts, "--- Tasks ---")
		taskKeys := make([]string, 0, len(auth.Tasks))
		for k := range auth.Tasks {
			taskKeys = append(taskKeys, k)
		}
		sort.Strings(taskKeys)
		for _, id := range taskKeys {
			task := auth.Tasks[id]
			ownerName := task.OwnerID
			if actor, ok := auth.Actors[task.OwnerID]; ok && actor != nil && actor.Name != "" {
				ownerName = actor.Name
			}
			line := fmt.Sprintf("- %s (%s) [Status: %s] [Owner: %s]", taskDisplayName(task), id, task.Status, ownerName)
			if task.Stage != "" {
				line += fmt.Sprintf(" [Stage: %s]", task.Stage)
			}
			parts = append(parts, line)
		}
		parts = append(parts, "")
	}

	if len(parts) <= 1 {
		return ""
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}

func actorDisplayName(a *DemoActorState) string {
	if a == nil {
		return "unknown"
	}
	if a.Name != "" {
		return a.Name
	}
	return a.ID
}

func itemDisplayName(i *DemoItemState) string {
	if i == nil {
		return "unknown"
	}
	if i.Name != "" {
		return i.Name
	}
	return i.ID
}

func taskDisplayName(t *DemoQuestState) string {
	if t == nil {
		return "unknown"
	}
	if t.Name != "" {
		return t.Name
	}
	return t.ID
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}

