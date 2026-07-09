package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseWorldTimeSettingsInputFromJSON(t *testing.T) {
	settings, err := parseWorldTimeSettingsInput(`{"tick_scale_mode":"fixed","tick_min_unit":"时辰","tick_step":2}`, "")
	if err != nil {
		t.Fatalf("parse json input: %v", err)
	}
	if settings == nil || settings.TickScaleMode != "fixed" || settings.TickMinUnit != "时辰" || settings.TickStep != 2 {
		t.Fatalf("unexpected settings: %#v", settings)
	}
}

func TestParseWorldTimeSettingsInputFromFile(t *testing.T) {
	content := `{"tick_scale_mode":"flexible","tick_min_unit":"时辰","tick_step":1}`
	dir := t.TempDir()
	file := filepath.Join(dir, "world-time-settings.json")
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	settings, err := parseWorldTimeSettingsInput("", file)
	if err != nil {
		t.Fatalf("parse file input: %v", err)
	}
	if settings == nil || settings.TickScaleMode != "flexible" || settings.TickStep != 1 {
		t.Fatalf("unexpected settings: %#v", settings)
	}
}

func TestParseWorldTimeSettingsInputRejectsBothSources(t *testing.T) {
	_, err := parseWorldTimeSettingsInput("{}", "test.json")
	if err == nil {
		t.Fatal("expected conflict error")
	}
}
