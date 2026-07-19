package types

// WorldFocusConfig represents configuration for promoting descendant nodes
// into world-tick or scope-tick context.
type WorldFocusConfig struct {
	Enabled           bool     `json:"enabled"`
	Tasks             []string `json:"tasks,omitempty"`
	Priority          int      `json:"priority,omitempty"`
	Reason            string   `json:"reason,omitempty"`
	MaxParentDistance int      `json:"max_parent_distance,omitempty"`
	SummaryOnly       bool     `json:"summary_only,omitempty"`
	IncludeChildren   int      `json:"include_children,omitempty"`
	IncludeRelations  []string `json:"include_relations,omitempty"`
}
