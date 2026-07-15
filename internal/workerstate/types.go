package workerstate

type WorldState struct {
	WorldID string                  `json:"world_id" yaml:"world_id"`
	Meta    map[string]any          `json:"meta,omitempty" yaml:"meta,omitempty"`
	Actors  map[string]*ActorState  `json:"actors,omitempty" yaml:"actors,omitempty"`
	Scenes  map[string]*SceneState  `json:"scenes,omitempty" yaml:"scenes,omitempty"`
	Items   map[string]*ItemState   `json:"items,omitempty" yaml:"items,omitempty"`
	Tasks   map[string]*QuestState  `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

type ActorState struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Kind        string            `json:"kind,omitempty" yaml:"kind,omitempty"`
	HP          int               `json:"hp,omitempty" yaml:"hp,omitempty"`
	MaxHP       int               `json:"max_hp,omitempty" yaml:"max_hp,omitempty"`
	Money       int               `json:"money,omitempty" yaml:"money,omitempty"`
	LocationID  string            `json:"location_id,omitempty" yaml:"location_id,omitempty"`
	Inventory   []InventoryEntry  `json:"inventory,omitempty" yaml:"inventory,omitempty"`
	QuestStates map[string]string `json:"quest_states,omitempty" yaml:"quest_states,omitempty"`
	Flags       map[string]any    `json:"flags,omitempty" yaml:"flags,omitempty"`
}

type InventoryEntry struct {
	ItemID    string         `json:"item_id" yaml:"item_id"`
	Quantity  int            `json:"quantity,omitempty" yaml:"quantity,omitempty"`
	Equipped  bool           `json:"equipped,omitempty" yaml:"equipped,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type SceneState struct {
	ID          string         `json:"id" yaml:"id"`
	Name        string         `json:"name,omitempty" yaml:"name,omitempty"`
	Kind        string         `json:"kind,omitempty" yaml:"kind,omitempty"`
	Occupants   []string       `json:"occupants,omitempty" yaml:"occupants,omitempty"`
	Flags       map[string]any `json:"flags,omitempty" yaml:"flags,omitempty"`
	Description string         `json:"description,omitempty" yaml:"description,omitempty"`
}

type ItemState struct {
	ID          string         `json:"id" yaml:"id"`
	Name        string         `json:"name,omitempty" yaml:"name,omitempty"`
	OwnerID     string         `json:"owner_id,omitempty" yaml:"owner_id,omitempty"`
	SceneID     string         `json:"scene_id,omitempty" yaml:"scene_id,omitempty"`
	Stackable   bool           `json:"stackable,omitempty" yaml:"stackable,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type QuestState struct {
	ID          string         `json:"id" yaml:"id"`
	Name        string         `json:"name,omitempty" yaml:"name,omitempty"`
	Status      string         `json:"status,omitempty" yaml:"status,omitempty"`
	OwnerID     string         `json:"owner_id,omitempty" yaml:"owner_id,omitempty"`
	Stage       string         `json:"stage,omitempty" yaml:"stage,omitempty"`
	Flags       map[string]any `json:"flags,omitempty" yaml:"flags,omitempty"`
}

type StateView struct {
	state *WorldState
}

func NewStateView(state *WorldState) *StateView {
	return &StateView{state: normalizeWorldState(state)}
}
