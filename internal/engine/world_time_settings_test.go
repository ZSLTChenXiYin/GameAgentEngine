package engine

import "testing"

func validWorldTimeSettings() *WorldTimeSettings {
	return &WorldTimeSettings{
		TickScaleMode: TickScaleModeFixed,
		TickMinUnit:   "时辰",
		TickStep:      2,
		TickUnits:     []string{"年", "月", "日", "时辰"},
		TimeScaleCarry: []WorldTimeCarryRule{
			{From: "时辰", To: "日", Base: 12},
			{From: "日", To: "月", Base: 30},
			{From: "月", To: "年", Base: 12},
		},
		TimeCalendar: &WorldTimeCalendar{
			Enabled:      true,
			CalendarName: "太阴",
			Units: []WorldTimeCalendarUnit{
				{Unit: "年", Value: "8"},
				{Unit: "月", Value: "7"},
				{Unit: "日", Value: "20"},
				{Unit: "时辰", Value: "卯"},
			},
		},
		UnitValueSequence: []WorldTimeUnitSequence{{
			Unit:   "时辰",
			Values: []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"},
		}},
	}
}

func TestValidateWorldTimeSettingsAcceptsValidConfig(t *testing.T) {
	if err := ValidateWorldTimeSettings(validWorldTimeSettings()); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateWorldTimeSettingsRejectsInvalidMode(t *testing.T) {
	settings := validWorldTimeSettings()
	settings.TickScaleMode = "broken"
	if err := ValidateWorldTimeSettings(settings); err == nil {
		t.Fatal("expected invalid mode error")
	}
}

func TestValidateWorldTimeSettingsRejectsMissingCalendarName(t *testing.T) {
	settings := validWorldTimeSettings()
	settings.TimeCalendar.CalendarName = ""
	if err := ValidateWorldTimeSettings(settings); err == nil {
		t.Fatal("expected missing calendar name error")
	}
}

func TestValidateWorldTimeSettingsRejectsMismatchedMinUnit(t *testing.T) {
	settings := validWorldTimeSettings()
	settings.TickMinUnit = "日"
	if err := ValidateWorldTimeSettings(settings); err == nil {
		t.Fatal("expected mismatched min unit error")
	}
}

func TestValidateWorldTimeSettingsRejectsMissingCarryRule(t *testing.T) {
	settings := validWorldTimeSettings()
	settings.TimeScaleCarry = settings.TimeScaleCarry[:2]
	if err := ValidateWorldTimeSettings(settings); err == nil {
		t.Fatal("expected missing carry rule error")
	}
}

func TestValidateWorldTimeSettingsRejectsUnknownSymbolicCalendarValue(t *testing.T) {
	settings := validWorldTimeSettings()
	settings.TimeCalendar.Units[len(settings.TimeCalendar.Units)-1].Value = "未知"
	if err := ValidateWorldTimeSettings(settings); err == nil {
		t.Fatal("expected invalid symbolic calendar value error")
	}
}

func TestValidateWorldTimeSettingsRejectsMismatchedSequenceLength(t *testing.T) {
	settings := validWorldTimeSettings()
	settings.UnitValueSequence[0].Values = []string{"子", "丑"}
	if err := ValidateWorldTimeSettings(settings); err == nil {
		t.Fatal("expected invalid sequence length error")
	}
}

func TestAdvanceWorldTimeStateAdvancesCalendarAndCarries(t *testing.T) {
	settings := validWorldTimeSettings()
	state, err := AdvanceWorldTimeState(settings, nil, 3, "manual-day-1")
	if err != nil {
		t.Fatalf("advance world time state: %v", err)
	}
	if state.TotalTicks != 6 {
		t.Fatalf("expected total_ticks=6, got %d", state.TotalTicks)
	}
	if state.CurrentTimeLabel != "太阴历 8年 7月 20日 酉时辰" {
		t.Fatalf("unexpected current_time_label: %q", state.CurrentTimeLabel)
	}
	if got := state.CurrentUnits[len(state.CurrentUnits)-1].Value; got != "酉" {
		t.Fatalf("expected 时辰=酉, got %q", got)
	}
	if got, _ := state.Metadata["advanced_min_units"].(int); got != 6 {
		t.Fatalf("expected advanced_min_units=6, got %#v", state.Metadata)
	}
}

func TestAdvanceWorldTimeStateUsesPreviousStateAndCarriesAcrossUnits(t *testing.T) {
	settings := validWorldTimeSettings()
	previous := &WorldTimeStateComponent{
		TickScaleMode: TickScaleModeFixed,
		TickMinUnit:   "时辰",
		TickStep:      2,
		TickUnits:     []string{"年", "月", "日", "时辰"},
		CurrentUnits: []WorldTimeCalendarUnit{
			{Unit: "年", Value: "8"},
			{Unit: "月", Value: "7"},
			{Unit: "日", Value: "20"},
			{Unit: "时辰", Value: "亥"},
		},
		TotalTicks: 24,
	}

	state, err := AdvanceWorldTimeState(settings, previous, 1, "manual-day-2")
	if err != nil {
		t.Fatalf("advance world time state: %v", err)
	}
	if state.TotalTicks != 26 {
		t.Fatalf("expected accumulated total_ticks=26, got %d", state.TotalTicks)
	}
	if state.CurrentTimeLabel != "太阴历 8年 7月 21日 丑时辰" {
		t.Fatalf("unexpected current_time_label: %q", state.CurrentTimeLabel)
	}
	if state.CurrentUnits[2].Value != "21" || state.CurrentUnits[3].Value != "丑" {
		t.Fatalf("unexpected current units: %#v", state.CurrentUnits)
	}
}

func TestAdvanceWorldTimeStateFallsBackToZeroInitializedUnits(t *testing.T) {
	settings := &WorldTimeSettings{
		TickScaleMode: TickScaleModeFlexible,
		TickMinUnit:   "秒",
		TickStep:      5,
		TickUnits:     []string{"分", "秒"},
		TimeScaleCarry: []WorldTimeCarryRule{{
			From: "秒",
			To:   "分",
			Base: 60,
		}},
	}

	state, err := AdvanceWorldTimeState(settings, nil, 3, "external")
	if err != nil {
		t.Fatalf("advance world time state: %v", err)
	}
	if state.CurrentTimeLabel != "0分 15秒" {
		t.Fatalf("unexpected current_time_label: %q", state.CurrentTimeLabel)
	}
	if state.CurrentUnits[0].Value != "0" || state.CurrentUnits[1].Value != "15" {
		t.Fatalf("unexpected current units: %#v", state.CurrentUnits)
	}
}
