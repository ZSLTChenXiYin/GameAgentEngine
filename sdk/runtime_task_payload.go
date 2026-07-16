package sdk

import (
	"encoding/json"
	"strings"
)

// ParseRuntimeTaskPayloadJSON decodes a runtime task payload JSON string.
// When the payload is not valid JSON, it preserves the original body so
// callers can still report or inspect the raw content.
func ParseRuntimeTaskPayloadJSON(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return map[string]any{"raw_payload_json": raw}
	}
	return payload
}

// ParsePayloadJSON decodes this runtime task's payload JSON into a loose map.
func (t *RuntimeTask) ParsePayloadJSON() map[string]any {
	if t == nil {
		return nil
	}
	return ParseRuntimeTaskPayloadJSON(t.PayloadJSON)
}
