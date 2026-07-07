package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

func validateCLIComponentData(componentType, data string) error {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return fmt.Errorf("component data cannot be empty")
	}
	meta, ok := engine.ComponentMetaFor(engine.ComponentType(componentType))
	if !ok {
		return nil
	}
	switch meta.ValidationMode {
	case engine.ComponentValidationWeak:
		var payload map[string]any
		if err := json.Unmarshal([]byte(trimmed), &payload); err != nil || payload == nil {
			return fmt.Errorf("%s component data must be a valid JSON object", componentType)
		}
	case engine.ComponentValidationStrong:
		var payload map[string]any
		if err := json.Unmarshal([]byte(trimmed), &payload); err != nil || payload == nil {
			return fmt.Errorf("autonomous component data must be a valid JSON object")
		}
		trigger, _ := payload["trigger"].(string)
		if trigger == "" {
			trigger = "manual"
		}
		if trigger != "manual" && trigger != "world_tick_sync" && trigger != "scheduled" {
			return fmt.Errorf("autonomous trigger must be one of: manual, world_tick_sync, scheduled")
		}
		if trigger == "scheduled" {
			interval, ok := payload["interval_seconds"].(float64)
			if !ok || interval <= 0 {
				return fmt.Errorf("scheduled autonomous component requires interval_seconds > 0")
			}
		}
	}
	return nil
}
