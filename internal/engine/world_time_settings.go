package engine

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
