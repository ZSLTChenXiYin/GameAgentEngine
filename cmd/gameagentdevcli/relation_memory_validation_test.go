package main

import "testing"

func TestValidatePropagationModeAcceptsKnownModes(t *testing.T) {
	for _, mode := range []string{"upward", "environment_scope", "organization_scope", "tag_broadcast", "targeted", "manual"} {
		if err := validatePropagationMode(mode); err != nil {
			t.Fatalf("expected mode %q to be accepted, got %v", mode, err)
		}
	}
}

func TestValidatePropagationModeRejectsUnknownMode(t *testing.T) {
	if err := validatePropagationMode("bad_mode"); err == nil {
		t.Fatal("expected invalid propagation mode error")
	}
}
