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
		settings.TickUnits[i] = unit
	}
	if settings.TickUnits[len(settings.TickUnits)-1] != settings.TickMinUnit {
		return fmt.Errorf("tick_min_unit must match the smallest configured tick unit")
	}

	carryByFrom := map[string]WorldTimeCarryRule{}
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
			carryByFrom[rule.From] = rule
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
		if unit == settings.TickUnits[0] {
			return fmt.Errorf("unit_value_sequences[%d].unit %q cannot be the largest tick unit", i, unit)
		}
		rule, ok := carryByFrom[unit]
		if !ok {
			return fmt.Errorf("unit_value_sequences[%d].unit %q requires a matching time_scale_carry rule", i, unit)
		}
		if len(seq.Values) != rule.Base {
			return fmt.Errorf("unit_value_sequences[%d].values for %q must contain exactly %d entries", i, unit, rule.Base)
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
		if numericValue, err := strconv.Atoi(unit.Value); err == nil {
			if rule, ok := carryByFrom[unit.Unit]; ok {
				if numericValue < 0 || numericValue >= rule.Base {
					return fmt.Errorf("time_calendar.units[%d].value for %q must be between 0 and %d", i, unit.Unit, rule.Base-1)
				}
			}
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

// AdvanceWorldTimeState advances one persisted world time state by the configured base tick count.
func AdvanceWorldTimeState(settings *WorldTimeSettings, previous *WorldTimeStateComponent, advancedTicks int, fallbackLabel string) (WorldTimeStateComponent, error) {
	state := WorldTimeStateComponent{}
	if settings == nil {
		state.CurrentTimeLabel = strings.TrimSpace(fallbackLabel)
		state.LastAdvancedTicks = advancedTicks
		state.TotalTicks = advancedTicks
		return state, nil
	}
	if err := ValidateWorldTimeSettings(settings); err != nil {
		return state, err
	}
	if advancedTicks <= 0 {
		return state, fmt.Errorf("advanced_ticks must be greater than 0")
	}

	units, compatible, err := worldTimeUnitsForAdvance(settings, previous)
	if err != nil {
		return state, err
	}
	advancedMinUnits := settings.TickStep * advancedTicks
	units, err = advanceWorldTimeUnits(settings, units, advancedMinUnits)
	if err != nil {
		return state, err
	}

	state.TickScaleMode = settings.TickScaleMode
	state.TickMinUnit = settings.TickMinUnit
	state.TickStep = settings.TickStep
	state.TickUnits = append([]string{}, settings.TickUnits...)
	if settings.TimeCalendar != nil && settings.TimeCalendar.Enabled {
		state.CalendarName = settings.TimeCalendar.CalendarName
	}
	state.CurrentUnits = units
	state.CurrentTimeLabel = buildWorldTimeLabel(settings, units)
	if state.CurrentTimeLabel == "" {
		state.CurrentTimeLabel = strings.TrimSpace(fallbackLabel)
	}
	state.LastAdvancedTicks = advancedTicks
	state.TotalTicks = advancedMinUnits
	if compatible && previous != nil && previous.TotalTicks > 0 {
		state.TotalTicks = previous.TotalTicks + advancedMinUnits
	}
	state.Metadata = map[string]any{
		"advanced_min_units": advancedMinUnits,
	}
	if strings.TrimSpace(fallbackLabel) != "" {
		state.Metadata["external_time_label"] = strings.TrimSpace(fallbackLabel)
	}
	return state, nil
}

func worldTimeUnitsForAdvance(settings *WorldTimeSettings, previous *WorldTimeStateComponent) ([]WorldTimeCalendarUnit, bool, error) {
	if previous != nil && worldTimeStateCompatible(settings, previous) {
		units := cloneWorldTimeCalendarUnits(previous.CurrentUnits)
		if len(units) == len(settings.TickUnits) {
			return units, true, nil
		}
	}
	return initialWorldTimeUnits(settings), false, nil
}

func worldTimeStateCompatible(settings *WorldTimeSettings, previous *WorldTimeStateComponent) bool {
	if previous == nil {
		return false
	}
	if previous.TickScaleMode != settings.TickScaleMode || previous.TickMinUnit != settings.TickMinUnit || previous.TickStep != settings.TickStep {
		return false
	}
	if len(previous.TickUnits) != len(settings.TickUnits) || len(previous.CurrentUnits) != len(settings.TickUnits) {
		return false
	}
	for i, unit := range settings.TickUnits {
		if previous.TickUnits[i] != unit || previous.CurrentUnits[i].Unit != unit {
			return false
		}
	}
	return true
}

func initialWorldTimeUnits(settings *WorldTimeSettings) []WorldTimeCalendarUnit {
	if settings.TimeCalendar != nil && settings.TimeCalendar.Enabled && len(settings.TimeCalendar.Units) == len(settings.TickUnits) {
		return cloneWorldTimeCalendarUnits(settings.TimeCalendar.Units)
	}
	sequences := unitSequenceMap(settings)
	units := make([]WorldTimeCalendarUnit, 0, len(settings.TickUnits))
	for _, unit := range settings.TickUnits {
		value := "0"
		if seq := sequences[unit]; len(seq) > 0 {
			value = seq[0]
		}
		units = append(units, WorldTimeCalendarUnit{Unit: unit, Value: value})
	}
	return units
}

func cloneWorldTimeCalendarUnits(units []WorldTimeCalendarUnit) []WorldTimeCalendarUnit {
	cloned := make([]WorldTimeCalendarUnit, len(units))
	copy(cloned, units)
	return cloned
}

func advanceWorldTimeUnits(settings *WorldTimeSettings, units []WorldTimeCalendarUnit, advancedMinUnits int) ([]WorldTimeCalendarUnit, error) {
	if len(units) != len(settings.TickUnits) {
		return nil, fmt.Errorf("current units do not match tick_units")
	}
	carryByFrom := carryRuleMap(settings)
	sequences := unitSequenceMap(settings)
	result := cloneWorldTimeCalendarUnits(units)
	carry := advancedMinUnits

	for idx := len(result) - 1; idx >= 0; idx-- {
		unitName := settings.TickUnits[idx]
		if result[idx].Unit == "" {
			result[idx].Unit = unitName
		}
		if result[idx].Unit != unitName {
			return nil, fmt.Errorf("current_units[%d].unit must be %q", idx, unitName)
		}
		base := 0
		if idx > 0 {
			rule := carryByFrom[unitName]
			base = rule.Base
		}

		seq := sequences[unitName]
		if len(seq) > 0 {
			position := indexOfString(seq, result[idx].Value)
			if position < 0 {
				return nil, fmt.Errorf("current_units[%d].value %q is not valid for %q", idx, result[idx].Value, unitName)
			}
			if base == 0 {
				position += carry
				result[idx].Value = seq[position%len(seq)]
				carry = position / len(seq)
				continue
			}
			position += carry
			carry = position / base
			position = position % base
			result[idx].Value = seq[position]
			continue
		}

		value, err := strconv.Atoi(strings.TrimSpace(result[idx].Value))
		if err != nil {
			return nil, fmt.Errorf("current_units[%d].value %q must be numeric for %q", idx, result[idx].Value, unitName)
		}
		if base == 0 {
			value += carry
			carry = 0
			result[idx].Value = strconv.Itoa(value)
			continue
		}
		value += carry
		carry = value / base
		value = value % base
		result[idx].Value = strconv.Itoa(value)
	}

	if carry > 0 {
		largestValue, err := strconv.Atoi(strings.TrimSpace(result[0].Value))
		if err == nil {
			result[0].Value = strconv.Itoa(largestValue + carry)
			return result, nil
		}
		return nil, fmt.Errorf("largest tick unit %q overflowed but is not numeric", result[0].Unit)
	}
	return result, nil
}

func buildWorldTimeLabel(settings *WorldTimeSettings, units []WorldTimeCalendarUnit) string {
	parts := make([]string, 0, len(units))
	for _, unit := range units {
		value := strings.TrimSpace(unit.Value)
		if value == "" || strings.TrimSpace(unit.Unit) == "" {
			continue
		}
		parts = append(parts, value+unit.Unit)
	}
	if len(parts) == 0 {
		return ""
	}
	if settings != nil && settings.TimeCalendar != nil && settings.TimeCalendar.Enabled && strings.TrimSpace(settings.TimeCalendar.CalendarName) != "" {
		return strings.TrimSpace(settings.TimeCalendar.CalendarName) + "历 " + strings.Join(parts, " ")
	}
	return strings.Join(parts, " ")
}

func carryRuleMap(settings *WorldTimeSettings) map[string]WorldTimeCarryRule {
	result := make(map[string]WorldTimeCarryRule, len(settings.TimeScaleCarry))
	for _, rule := range settings.TimeScaleCarry {
		result[rule.From] = rule
	}
	return result
}

func unitSequenceMap(settings *WorldTimeSettings) map[string][]string {
	result := make(map[string][]string, len(settings.UnitValueSequence))
	for _, seq := range settings.UnitValueSequence {
		result[seq.Unit] = append([]string{}, seq.Values...)
	}
	return result
}

func indexOfString(items []string, target string) int {
	for i, item := range items {
		if item == target {
			return i
		}
	}
	return -1
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
