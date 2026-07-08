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
		UnitValueSequence: []WorldTimeUnitSequence{{Unit: "时辰", Values: []string{"子", "丑", "寅", "卯", "辰"}}},
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
