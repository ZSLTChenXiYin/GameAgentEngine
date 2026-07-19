package engine

import (
	"testing"
)

func TestDecodeWorldFocusConfigValid(t *testing.T) {
	json := `{"enabled":true,"tasks":["world_tick"],"priority":80,"reason":"quest_hub","max_parent_distance":3,"summary_only":true,"include_relations":["belongs_to","located_at"]}`
	
	cfg, err := DecodeWorldFocusConfig(json)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if !cfg.Enabled {
		t.Error("expected Enabled=true")
	}
	if cfg.Priority != 80 {
		t.Errorf("expected Priority=80, got %d", cfg.Priority)
	}
	if cfg.MaxParentDistance != 3 {
		t.Errorf("expected MaxParentDistance=3, got %d", cfg.MaxParentDistance)
	}
	if len(cfg.Tasks) != 1 || cfg.Tasks[0] != "world_tick" {
		t.Errorf("unexpected tasks: %v", cfg.Tasks)
	}
}

func TestDecodeWorldFocusConfigEmpty(t *testing.T) {
	cfg, err := DecodeWorldFocusConfig("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Error("expected nil for empty data")
	}
}

func TestDecodeWorldFocusConfigInvalidJSON(t *testing.T) {
	_, err := DecodeWorldFocusConfig("{invalid}")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDecodeWorldFocusConfigNegativeValues(t *testing.T) {
	json := `{"enabled":true,"max_parent_distance":-5,"priority":-10,"include_children":-3}`
	
	cfg, err := DecodeWorldFocusConfig(json)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxParentDistance < 0 {
		t.Errorf("negative MaxParentDistance should be clamped, got %d", cfg.MaxParentDistance)
	}
	if cfg.Priority < 0 {
		t.Errorf("negative Priority should be clamped, got %d", cfg.Priority)
	}
	if cfg.IncludeChildren < 0 {
		t.Errorf("negative IncludeChildren should be clamped, got %d", cfg.IncludeChildren)
	}
}
