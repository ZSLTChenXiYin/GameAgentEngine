package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// DecodeAutonomousConfig parses an autonomous component payload and applies safe defaults.
func DecodeAutonomousConfig(data string) (*AutonomousConfig, error) {
	var cfg AutonomousConfig
	if strings.TrimSpace(data) == "" {
		cfg.Trigger = AutonomousTriggerManual
		return &cfg, nil
	}
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, fmt.Errorf("parse autonomous config: %w", err)
	}
	if cfg.Trigger == "" {
		cfg.Trigger = AutonomousTriggerManual
	}
	return &cfg, nil
}

// LoadAutonomousConfig loads the node-local autonomous component.
func LoadAutonomousConfig(nodeID string) (*AutonomousConfig, *store.ComponentModel, error) {
	if _, err := store.GetNode(nodeID); err != nil {
		return nil, nil, err
	}
	comps, err := store.GetComponentsByType(nodeID, string(CompAutonomous))
	if err != nil {
		return nil, nil, err
	}
	if len(comps) == 0 {
		return nil, nil, nil
	}
	cfg, err := DecodeAutonomousConfig(comps[0].Data)
	if err != nil {
		return nil, &comps[0], err
	}
	return cfg, &comps[0], nil
}

// SaveAutonomousConfig writes the runtime status back to the autonomous component.
func SaveAutonomousConfig(componentID string, cfg *AutonomousConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return store.UpdateComponent(componentID, map[string]any{"data": string(data)})
}

// CapabilitySchema describes one field's type and required flag from a capability schema.
type CapabilitySchema struct {
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
}

func filterActionCallsByCapabilities(calls []ActionCall, capabilities []AgentCapability) ([]ActionCall, []ActionCall) {
	allowedIDs := map[string]bool{}
	for _, cap := range capabilities {
		if cap.ID != "" {
			allowedIDs[cap.ID] = true
		}
	}
	var allowed []ActionCall
	var rejected []ActionCall
	for _, call := range calls {
		if allowedIDs[call.ActionID] {
			allowed = append(allowed, call)
		} else {
			rejected = append(rejected, call)
		}
	}
	return allowed, rejected
}

// validateActionCallsBySchema checks capability-level schema for each action call.
func validateActionCallsBySchema(calls []ActionCall, capabilities []AgentCapability) (accepted, rejected []ActionCall) {
	schemaMap := map[string]map[string]CapabilitySchema{}
	for _, cap := range capabilities {
		if len(cap.Schema) == 0 || cap.ID == "" {
			continue
		}
		fields := map[string]CapabilitySchema{}
		for k, v := range cap.Schema {
			if m, ok := v.(map[string]any); ok {
				s := CapabilitySchema{}
				if t, ok := m["type"].(string); ok {
					s.Type = t
				}
				s.Required, _ = m["required"].(bool)
				fields[k] = s
			}
		}
		schemaMap[cap.ID] = fields
	}

	for _, call := range calls {
		fields, hasSchema := schemaMap[call.ActionID]
		if !hasSchema {
			accepted = append(accepted, call)
			continue
		}
		valid := true
		if call.Args == nil {
			call.Args = map[string]any{}
		}
		for key, sf := range fields {
			argVal, argExists := call.Args[key]
			if sf.Required && !argExists {
				log.Printf("[autonomous:schema] action=%s missing required arg %q", call.ActionID, key)
				valid = false
				break
			}
			if !argExists {
				continue
			}
			if sf.Type != "" && !typeMatchesSchema(argVal, sf.Type) {
				log.Printf("[autonomous:schema] action=%s arg %q expected %q got %T", call.ActionID, key, sf.Type, argVal)
				valid = false
				break
			}
		}
		if valid {
			accepted = append(accepted, call)
		} else {
			rejected = append(rejected, call)
		}
	}
	return
}

func typeMatchesSchema(v any, schemaType string) bool {
	switch schemaType {
	case "string":
		_, ok := v.(string)
		return ok
	case "integer", "int":
		switch v.(type) {
		case float64, int, int64:
			return true
		}
		return false
	case "number":
		switch v.(type) {
		case float64, int, int64:
			return true
		}
		return false
	case "boolean", "bool":
		_, ok := v.(bool)
		return ok
	case "object":
		_, ok := v.(map[string]any)
		return ok
	case "array":
		_, ok := v.([]any)
		return ok
	default:
		return true
	}
}

// AutonomousPriority returns the configured priority, defaulting to 0.
func AutonomousPriority(cfg *AutonomousConfig) int {
	if cfg == nil {
		return 0
	}
	return cfg.Priority
}

// ReadyForCooldown checks whether cooldown_seconds has elapsed since last run.
func ReadyForCooldown(cfg *AutonomousConfig) bool {
	if cfg == nil || cfg.CooldownSeconds <= 0 || cfg.LastRunAt == nil {
		return true
	}
	return time.Since(*cfg.LastRunAt) >= time.Duration(cfg.CooldownSeconds)*time.Second
}
