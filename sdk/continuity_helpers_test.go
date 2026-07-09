package sdk

import "testing"

func TestContinuityBundleLatestWorldTimeState(t *testing.T) {
	bundle := &ContinuityBundle{
		StateComponents: []StateComponentEnvelope{{
			ComponentType: "world_time_state",
			Data: map[string]any{
				"current_time_label":  "太阴历 8年 7月 20日 卯时辰",
				"last_advanced_ticks": 3,
			},
		}},
	}
	state := bundle.LatestWorldTimeState()
	if state == nil {
		t.Fatal("expected world time state")
	}
	if state.CurrentTimeLabel != "太阴历 8年 7月 20日 卯时辰" || state.LastAdvancedTicks != 3 {
		t.Fatalf("unexpected world time state: %#v", state)
	}
}

func TestTimelineEnvelopeWorldTimeHelpers(t *testing.T) {
	timeline := &TimelineEnvelope{
		Data: map[string]any{
			"world_time_state": map[string]any{
				"current_time_label":  "Day 12",
				"last_advanced_ticks": 2,
			},
			"previous_world_time_state": map[string]any{
				"current_time_label":  "Day 10",
				"last_advanced_ticks": 1,
			},
		},
	}
	current := timeline.WorldTimeState()
	previous := timeline.PreviousWorldTimeState()
	if current == nil || previous == nil {
		t.Fatalf("expected both states, got current=%#v previous=%#v", current, previous)
	}
	if current.CurrentTimeLabel != "Day 12" || previous.CurrentTimeLabel != "Day 10" {
		t.Fatalf("unexpected states: current=%#v previous=%#v", current, previous)
	}
	if timeline.EffectiveAdvancedTicks() != 2 {
		t.Fatalf("expected effective advanced ticks 2, got %d", timeline.EffectiveAdvancedTicks())
	}
}

func TestTimelineEnvelopeEffectiveAdvancedTicksFallsBackToField(t *testing.T) {
	timeline := &TimelineEnvelope{AdvancedTicks: 4}
	if timeline.EffectiveAdvancedTicks() != 4 {
		t.Fatalf("expected effective advanced ticks 4, got %d", timeline.EffectiveAdvancedTicks())
	}
}
