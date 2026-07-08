package engine

type ComponentValidationMode string

const (
	ComponentValidationStrong ComponentValidationMode = "strong"
	ComponentValidationWeak   ComponentValidationMode = "weak"
	ComponentValidationFree   ComponentValidationMode = "free"
)

type ComponentMeta struct {
	Type             ComponentType                `json:"type"`
	ValidationMode   ComponentValidationMode      `json:"validation_mode"`
	DataFormat       string                       `json:"data_format"`
	HelpText         string                       `json:"help_text,omitempty"`
	RequiredFields   []string                     `json:"required_fields,omitempty"`
	EnumFields       map[string][]string          `json:"enum_fields,omitempty"`
	PositiveIfEquals map[string]map[string]string `json:"positive_if_equals,omitempty"`
}

var componentMetaRegistry = map[ComponentType]ComponentMeta{
	CompProfile: {
		Type:           CompProfile,
		ValidationMode: ComponentValidationWeak,
		DataFormat:     "json_object",
		HelpText:       "JSON object required; fields are flexible.",
	},
	CompRule: {
		Type:           CompRule,
		ValidationMode: ComponentValidationFree,
		DataFormat:     "text",
		HelpText:       "Free text allowed.",
	},
	CompTimeline: {
		Type:           CompTimeline,
		ValidationMode: ComponentValidationFree,
		DataFormat:     "text",
		HelpText:       "Free text allowed.",
	},
	CompActionPolicy: {
		Type:           CompActionPolicy,
		ValidationMode: ComponentValidationFree,
		DataFormat:     "text",
		HelpText:       "Free text allowed.",
	},
	CompRelations: {
		Type:           CompRelations,
		ValidationMode: ComponentValidationFree,
		DataFormat:     "text",
		HelpText:       "Free text allowed.",
	},
	CompPromptProfile: {
		Type:           CompPromptProfile,
		ValidationMode: ComponentValidationFree,
		DataFormat:     "text",
		HelpText:       "Free text allowed.",
	},
	CompLore: {
		Type:           CompLore,
		ValidationMode: ComponentValidationFree,
		DataFormat:     "text",
		HelpText:       "Free text allowed.",
	},
	CompAutonomous: {
		Type:           CompAutonomous,
		ValidationMode: ComponentValidationStrong,
		DataFormat:     "json_object",
		HelpText:       "Structured autonomous config JSON.",
		RequiredFields: []string{"enabled", "trigger"},
		EnumFields: map[string][]string{
			"trigger": {AutonomousTriggerManual, AutonomousTriggerWorldTickSync, AutonomousTriggerScheduled},
		},
		PositiveIfEquals: map[string]map[string]string{
			"interval_seconds": {
				"trigger": AutonomousTriggerScheduled,
			},
		},
	},
	CompWorldState: {
		Type:           CompWorldState,
		ValidationMode: ComponentValidationWeak,
		DataFormat:     "json_object",
		HelpText:       "Structured current world state for tick continuity.",
	},
	CompStoryState: {
		Type:           CompStoryState,
		ValidationMode: ComponentValidationWeak,
		DataFormat:     "json_object",
		HelpText:       "Structured current narrative state and unresolved threads.",
	},
	CompStoryHistory: {
		Type:           CompStoryHistory,
		ValidationMode: ComponentValidationWeak,
		DataFormat:     "json_object",
		HelpText:       "Structured rolling history of recent story beats.",
	},
	CompTickPolicy: {
		Type:           CompTickPolicy,
		ValidationMode: ComponentValidationWeak,
		DataFormat:     "json_object",
		HelpText:       "Structured tick policy and continuity constraints.",
	},
	CompStateSnapshot: {
		Type:           CompStateSnapshot,
		ValidationMode: ComponentValidationWeak,
		DataFormat:     "json_object",
		HelpText:       "Structured snapshot payload for state rollups and checkpoints.",
	},
}

func ComponentMetaFor(componentType ComponentType) (ComponentMeta, bool) {
	meta, ok := componentMetaRegistry[componentType]
	return meta, ok
}

func ComponentMetaList() []ComponentMeta {
	ordered := []ComponentType{CompProfile, CompRule, CompTimeline, CompActionPolicy, CompRelations, CompPromptProfile, CompLore, CompAutonomous, CompWorldState, CompStoryState, CompStoryHistory, CompTickPolicy, CompStateSnapshot}
	items := make([]ComponentMeta, 0, len(ordered))
	for _, componentType := range ordered {
		if meta, ok := componentMetaRegistry[componentType]; ok {
			items = append(items, meta)
		}
	}
	return items
}
