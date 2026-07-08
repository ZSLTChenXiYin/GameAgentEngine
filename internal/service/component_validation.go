package service

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/go-playground/validator/v10"
)

var componentValidator = validator.New()

func ValidateComponentData(componentType, data string) error {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return codedErrorf(ErrorInvalidComponentData, "component_data_empty", "component data cannot be empty")
	}
	meta, ok := engine.ComponentMetaFor(engine.ComponentType(componentType))
	if !ok {
		return nil
	}

	switch meta.ValidationMode {
	case engine.ComponentValidationWeak:
		if err := validateJSONObjectPayload(trimmed); err != nil {
			return err
		}
		return nil
	case engine.ComponentValidationStrong:
		switch engine.ComponentType(componentType) {
		case engine.CompAutonomous:
			return validateAutonomousComponentData(trimmed)
		case engine.CompWorldState:
			return validateWorldStateComponentData(trimmed)
		case engine.CompStoryState:
			return validateStoryStateComponentData(trimmed)
		case engine.CompStoryHistory:
			return validateStoryHistoryComponentData(trimmed)
		case engine.CompTickPolicy:
			return validateTickPolicyComponentData(trimmed)
		}
	}

	return nil
}

func validateJSONObjectPayload(data string) error {
	var payload map[string]any
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_json", "component data must be a valid JSON object: %v", err)
	}
	if payload == nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_shape", "component data must be a JSON object")
	}
	return nil
}

func validateAutonomousComponentData(data string) error {
	var cfg engine.AutonomousConfig
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_json", "autonomous component data must be valid JSON: %v", err)
	}
	if err := componentValidator.Struct(componentAutonomousPayload{AutonomousConfig: cfg}); err != nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_fields", "invalid autonomous component data: %s", humanizeValidationError(err))
	}
	for i, cap := range cfg.Capabilities {
		if strings.TrimSpace(cap.ID) == "" {
			return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_fields", "invalid autonomous component data: capabilities[%d].id is required", i)
		}
		for fieldName, rawSchema := range cap.Schema {
			schemaMap, ok := rawSchema.(map[string]any)
			if !ok {
				return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_fields", "invalid autonomous component data: capabilities[%d].schema.%s must be an object", i, fieldName)
			}
			fieldType, _ := schemaMap["type"].(string)
			if strings.TrimSpace(fieldType) == "" {
				return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_fields", "invalid autonomous component data: capabilities[%d].schema.%s.type is required", i, fieldName)
			}
		}
	}
	return nil
}

func validateWorldStateComponentData(data string) error {
	var payload engine.WorldStateComponent
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_json", "world_state component data must be valid structured JSON: %v", err)
	}
	return nil
}

func validateStoryStateComponentData(data string) error {
	var payload engine.StoryStateComponent
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_json", "story_state component data must be valid structured JSON: %v", err)
	}
	return nil
}

func validateStoryHistoryComponentData(data string) error {
	var payload engine.StoryHistoryComponent
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_json", "story_history component data must be valid structured JSON: %v", err)
	}
	for i, entry := range payload.Entries {
		if entry.TickNumber < 0 {
			return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_fields", "invalid story_history component data: entries[%d].tick_number must be greater than or equal to 0", i)
		}
	}
	return nil
}

func validateTickPolicyComponentData(data string) error {
	var payload engine.TickPolicyComponent
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return codedErrorf(ErrorInvalidComponentData, "component_data_invalid_json", "tick_policy component data must be valid structured JSON: %v", err)
	}
	return nil
}

type componentAutonomousPayload struct {
	engine.AutonomousConfig
}

func (p componentAutonomousPayload) Validate() error {
	return nil
}

func humanizeValidationError(err error) string {
	if err == nil {
		return ""
	}
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok || len(validationErrs) == 0 {
		return err.Error()
	}
	parts := make([]string, 0, len(validationErrs))
	for _, ve := range validationErrs {
		switch ve.Field() {
		case "Trigger":
			parts = append(parts, "trigger must be one of: manual, world_tick_sync, scheduled")
		case "IntervalSeconds":
			parts = append(parts, "interval_seconds must be greater than 0 when trigger is scheduled")
		case "TickNumber":
			parts = append(parts, "tick_number must be greater than or equal to 0")
		default:
			parts = append(parts, fmt.Sprintf("%s failed validation %s", ve.Field(), ve.Tag()))
		}
	}
	return strings.Join(parts, "; ")
}

func init() {
	componentValidator.RegisterStructValidation(func(sl validator.StructLevel) {
		payload, ok := sl.Current().Interface().(componentAutonomousPayload)
		if !ok {
			return
		}
		cfg := payload.AutonomousConfig
		switch cfg.Trigger {
		case engine.AutonomousTriggerManual, engine.AutonomousTriggerWorldTickSync, engine.AutonomousTriggerScheduled:
		default:
			sl.ReportError(cfg.Trigger, "Trigger", "trigger", "oneof", "manual world_tick_sync scheduled")
		}
		if cfg.Trigger == engine.AutonomousTriggerScheduled && cfg.IntervalSeconds <= 0 {
			sl.ReportError(cfg.IntervalSeconds, "IntervalSeconds", "interval_seconds", "gt", "0")
		}
	}, componentAutonomousPayload{})
	componentValidator.RegisterStructValidation(func(sl validator.StructLevel) {
		entry, ok := sl.Current().Interface().(engine.StoryHistoryEntry)
		if !ok {
			return
		}
		if entry.TickNumber < 0 || math.Signbit(float64(entry.TickNumber)) {
			sl.ReportError(entry.TickNumber, "TickNumber", "tick_number", "gte", "0")
		}
	}, engine.StoryHistoryEntry{})
}
