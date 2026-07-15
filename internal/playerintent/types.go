package playerintent

import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

type ValidationIssue struct {
	StepIndex   int              `json:"step_index,omitempty"`
	Code        string           `json:"code"`
	Message     string           `json:"message"`
	MissingFact *sdk.MissingFact `json:"missing_fact,omitempty"`
}

type ValidationResult struct {
	OK     bool              `json:"ok"`
	Issues []ValidationIssue `json:"issues,omitempty"`
}

type StepOutcome struct {
	StepIndex int    `json:"step_index,omitempty"`
	Type      string `json:"type"`
	Applied   bool   `json:"applied"`
	Summary   string `json:"summary,omitempty"`
}

type ExecutionResult struct {
	ActorNodeID string        `json:"actor_node_id,omitempty"`
	SceneNodeID string        `json:"scene_node_id,omitempty"`
	Outcomes    []StepOutcome `json:"outcomes,omitempty"`
}

type InteractionSpec struct {
	Mode          string   `json:"mode,omitempty"`
	AudienceScope string   `json:"audience_scope,omitempty"`
	EventType     string   `json:"event_type,omitempty"`
	ItemID        string   `json:"item_id,omitempty"`
	Input         string   `json:"input,omitempty"`
	TargetNodeID  string   `json:"target_node_id,omitempty"`
	Participants  []string `json:"participants,omitempty"`
}
