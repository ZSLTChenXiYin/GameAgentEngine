package engine

import "github.com/ZSLTChenXiYin/GameAgentEngine/pkg/types"

// WorldStateComponent 保存当前世界的可继承状态快照。
type WorldStateComponent struct {
	Summary        string         `json:"summary,omitempty"`
	KeyFacts       []string       `json:"key_facts,omitempty"`
	CanonicalFacts []string       `json:"canonical_facts,omitempty"`
	OpenQuestions  []string       `json:"open_questions,omitempty"`
	ActiveArcs     []string       `json:"active_arcs,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// StoryStateComponent 保存当前剧情上下文与悬而未决的线索。
type StoryStateComponent struct {
	CurrentSituation string         `json:"current_situation,omitempty"`
	RecentChanges    []string       `json:"recent_changes,omitempty"`
	PendingThreads   []string       `json:"pending_threads,omitempty"`
	Tone             string         `json:"tone,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// StoryHistoryComponent 保存近期剧情历史，用于后续 tick 延续。
type StoryHistoryComponent struct {
	Entries  []StoryHistoryEntry `json:"entries,omitempty"`
	Metadata map[string]any      `json:"metadata,omitempty"`
}

type StoryHistoryEntry struct {
	TickNumber int      `json:"tick_number,omitempty"`
	Summary    string   `json:"summary,omitempty"`
	Facts      []string `json:"facts,omitempty"`
	GameTime   string   `json:"game_time,omitempty"`
}

// TickPolicyComponent 保存世界 tick 的约束和持续性要求。
type TickPolicyComponent struct {
	ContinuityRules []string       `json:"continuity_rules,omitempty"`
	FocusScopes     []string       `json:"focus_scopes,omitempty"`
	BannedResets    []string       `json:"banned_resets,omitempty"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// WorldTimeStateComponent stores the current world time state persisted by the engine.
type WorldTimeStateComponent struct {
	TickScaleMode     TickScaleMode           `json:"tick_scale_mode,omitempty"`
	TickMinUnit       string                  `json:"tick_min_unit,omitempty"`
	TickStep          int                     `json:"tick_step,omitempty"`
	TickUnits         []string                `json:"tick_units,omitempty"`
	CalendarName      string                  `json:"calendar_name,omitempty"`
	CurrentUnits      []WorldTimeCalendarUnit `json:"current_units,omitempty"`
	CurrentTimeLabel  string                  `json:"current_time_label,omitempty"`
	TotalTicks        int                     `json:"total_ticks,omitempty"`
	LastTickNumber    int                     `json:"last_tick_number,omitempty"`
	LastTickType      string                  `json:"last_tick_type,omitempty"`
	LastAdvancedTicks int                     `json:"last_advanced_ticks,omitempty"`
	Metadata          map[string]any          `json:"metadata,omitempty"`
}


// WorldFocusComponent marks a descendant node as structurally important
// for world tick context selection. It is consumed during context assembly.
type WorldFocusComponent = types.WorldFocusConfig

// StateSnapshotComponent 保存由引擎阶段性生成的结构化状态快照。
type StateSnapshotComponent struct {
	SnapshotType string         `json:"snapshot_type,omitempty"`
	Version      string         `json:"version,omitempty"`
	Payload      map[string]any `json:"payload,omitempty"`
}
