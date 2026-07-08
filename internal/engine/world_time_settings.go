package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// TickScaleMode describes how strictly one world tick is constrained by the engine.
type TickScaleMode string

const (
	TickScaleModeFixed    TickScaleMode = "fixed"
	TickScaleModeFlexible TickScaleMode = "flexible"
)

// WorldTimeSettings stores the configured world time system for one world.
type WorldTimeSettings struct {
	TickScaleMode     TickScaleMode           `json:"tick_scale_mode,omitempty"`
	TickMinUnit       string                  `json:"tick_min_unit,omitempty"`
	TickStep          int                     `json:"tick_step,omitempty"`
	TickUnits         []string                `json:"tick_units,omitempty"`
	TimeScaleCarry    []WorldTimeCarryRule    `json:"time_scale_carry,omitempty"`
	TimeCalendar      *WorldTimeCalendar      `json:"time_calendar,omitempty"`
	UnitValueSequence []WorldTimeUnitSequence `json:"unit_value_sequences,omitempty"`
}

// WorldTimeCarryRule declares one adjacent carry rule between two units.
type WorldTimeCarryRule struct {
	From string `json:"from"`
	To   string `json:"to"`
	Base int    `json:"base"`
}

// WorldTimeCalendar stores the optional named calendar template for one world.
type WorldTimeCalendar struct {
	Enabled      bool                    `json:"enabled"`
	CalendarName string                  `json:"calendar_name,omitempty"`
	Units        []WorldTimeCalendarUnit `json:"units,omitempty"`
}

// WorldTimeCalendarUnit stores one configured calendar unit and its current value.
type WorldTimeCalendarUnit struct {
	Unit  string `json:"unit"`
	Value string `json:"value,omitempty"`
}

// WorldTimeUnitSequence stores ordered symbolic values for a unit such as 时辰.
type WorldTimeUnitSequence struct {
	Unit   string   `json:"unit"`
	Values []string `json:"values,omitempty"`
}

// IsValidTickScaleMode reports whether the configured tick scale mode is supported.
func IsValidTickScaleMode(mode TickScaleMode) bool {
	switch mode {
	case TickScaleModeFixed, TickScaleModeFlexible:
		return true
	default:
		return false
	}
}

// ValidateWorldTimeSettings validates one world time system definition before persistence.
func ValidateWorldTimeSettings(settings *WorldTimeSettings) error {
	if settings == nil {
		return nil
	}
	if !IsValidTickScaleMode(settings.TickScaleMode) {
		return fmt.Errorf("tick_scale_mode must be one of: fixed, flexible")
	}
	if strings.TrimSpace(settings.TickMinUnit) == "" {
		return fmt.Errorf("tick_min_unit must not be empty")
	}
	if settings.TickStep <= 0 {
		return fmt.Errorf("tick_step must be greater than 0")
	}
	if len(settings.TickUnits) == 0 {
		return fmt.Errorf("tick_units must contain at least one unit")
	}

	seenUnits := map[string]bool{}
	for i, unit := range settings.TickUnits {
		unit = strings.TrimSpace(unit)
		if unit == "" {
			return fmt.Errorf("tick_units[%d] must not be empty", i)
		}
		if seenUnits[unit] {
			return fmt.Errorf("tick_units[%d] duplicates unit %q", i, unit)
		}
		seenUnits[unit] = true
	}
	if settings.TickUnits[len(settings.TickUnits)-1] != settings.TickMinUnit {
		return fmt.Errorf("tick_min_unit must match the smallest configured tick unit")
	}

	if len(settings.TickUnits) > 1 {
		if len(settings.TimeScaleCarry) != len(settings.TickUnits)-1 {
			return fmt.Errorf("time_scale_carry must define exactly %d adjacent carry rules", len(settings.TickUnits)-1)
		}
		for i := 0; i < len(settings.TickUnits)-1; i++ {
			rule := settings.TimeScaleCarry[i]
			expectedFrom := settings.TickUnits[len(settings.TickUnits)-1-i]
			expectedTo := settings.TickUnits[len(settings.TickUnits)-2-i]
			if strings.TrimSpace(rule.From) == "" || strings.TrimSpace(rule.To) == "" {
				return fmt.Errorf("time_scale_carry[%d] must define from and to", i)
			}
			if rule.Base <= 0 {
				return fmt.Errorf("time_scale_carry[%d].base must be greater than 0", i)
			}
			if rule.From != expectedFrom || rule.To != expectedTo {
				return fmt.Errorf("time_scale_carry[%d] must be %q -> %q", i, expectedFrom, expectedTo)
			}
		}
	}

	sequenceByUnit := map[string][]string{}
	for i, seq := range settings.UnitValueSequence {
		unit := strings.TrimSpace(seq.Unit)
		if unit == "" {
			return fmt.Errorf("unit_value_sequences[%d].unit must not be empty", i)
		}
		if !seenUnits[unit] {
			return fmt.Errorf("unit_value_sequences[%d].unit %q must exist in tick_units", i, unit)
		}
		if len(seq.Values) == 0 {
			return fmt.Errorf("unit_value_sequences[%d].values must not be empty", i)
		}
		sequenceByUnit[unit] = seq.Values
	}

	if settings.TimeCalendar == nil || !settings.TimeCalendar.Enabled {
		return nil
	}
	if strings.TrimSpace(settings.TimeCalendar.CalendarName) == "" {
		return fmt.Errorf("time_calendar.calendar_name must not be empty when time_calendar is enabled")
	}
	if len(settings.TimeCalendar.Units) != len(settings.TickUnits) {
		return fmt.Errorf("time_calendar.units must match tick_units exactly when time_calendar is enabled")
	}
	for i, unit := range settings.TimeCalendar.Units {
		if strings.TrimSpace(unit.Unit) == "" {
			return fmt.Errorf("time_calendar.units[%d].unit must not be empty", i)
		}
		if unit.Unit != settings.TickUnits[i] {
			return fmt.Errorf("time_calendar.units[%d].unit must be %q", i, settings.TickUnits[i])
		}
		if strings.TrimSpace(unit.Value) == "" {
			return fmt.Errorf("time_calendar.units[%d].value must not be empty", i)
		}
		if _, err := strconv.Atoi(unit.Value); err == nil {
			continue
		}
		seq := sequenceByUnit[unit.Unit]
		if len(seq) == 0 {
			return fmt.Errorf("time_calendar.units[%d].unit %q requires unit_value_sequences", i, unit.Unit)
		}
		found := false
		for _, item := range seq {
			if item == unit.Value {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("time_calendar.units[%d].value %q must exist in unit_value_sequences for %q", i, unit.Value, unit.Unit)
		}
	}
	if settings.TimeCalendar.Units[len(settings.TimeCalendar.Units)-1].Unit != settings.TickMinUnit {
		return fmt.Errorf("time_calendar smallest unit must match tick_min_unit")
	}
	return nil
}

// DecodeWorldTimeSettings parses stored world time settings JSON.
func DecodeWorldTimeSettings(raw string) (*WorldTimeSettings, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var settings WorldTimeSettings
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// EncodeWorldTimeSettings serializes one world time settings payload.
func EncodeWorldTimeSettings(settings *WorldTimeSettings) (string, error) {
	if settings == nil {
		return "", nil
	}
	data, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
