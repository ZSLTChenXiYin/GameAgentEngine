package sdk

import "encoding/json"

// FindStateComponent returns the first state component envelope for the given type.
func (b *ContinuityBundle) FindStateComponent(componentType string) *StateComponentEnvelope {
	if b == nil {
		return nil
	}
	for i := range b.StateComponents {
		if b.StateComponents[i].ComponentType == componentType {
			return &b.StateComponents[i]
		}
	}
	return nil
}

// LatestWorldTimeState returns the latest persisted world_time_state from state components.
func (b *ContinuityBundle) LatestWorldTimeState() *WorldTimeState {
	item := b.FindStateComponent("world_time_state")
	if item == nil {
		return nil
	}
	state, _ := DecodeWorldTimeState(item.Data)
	return state
}

// WorldTimeState extracts world_time_state from one timeline payload when present.
func (t *TimelineEnvelope) WorldTimeState() *WorldTimeState {
	return decodeTimelineWorldTimeState(t, "world_time_state")
}

// PreviousWorldTimeState extracts previous_world_time_state from one timeline payload when present.
func (t *TimelineEnvelope) PreviousWorldTimeState() *WorldTimeState {
	return decodeTimelineWorldTimeState(t, "previous_world_time_state")
}

// EffectiveAdvancedTicks returns the best known advanced tick count for one timeline entry.
func (t *TimelineEnvelope) EffectiveAdvancedTicks() int {
	if t == nil {
		return 0
	}
	if t.AdvancedTicks > 0 {
		return t.AdvancedTicks
	}
	if state := t.WorldTimeState(); state != nil && state.LastAdvancedTicks > 0 {
		return state.LastAdvancedTicks
	}
	return 0
}

// DecodeWorldTimeState converts a generic payload into a typed WorldTimeState.
func DecodeWorldTimeState(payload any) (*WorldTimeState, error) {
	if payload == nil {
		return nil, nil
	}
	if state, ok := payload.(*WorldTimeState); ok {
		return state, nil
	}
	if state, ok := payload.(WorldTimeState); ok {
		copyState := state
		return &copyState, nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	var state WorldTimeState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func decodeTimelineWorldTimeState(t *TimelineEnvelope, key string) *WorldTimeState {
	if t == nil || t.Data == nil {
		return nil
	}
	object, ok := t.Data.(map[string]any)
	if !ok {
		return nil
	}
	payload, ok := object[key]
	if !ok {
		return nil
	}
	state, err := DecodeWorldTimeState(payload)
	if err != nil {
		return nil
	}
	return state
}
