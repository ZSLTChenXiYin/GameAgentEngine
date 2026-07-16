package sdk

import "testing"

func TestParseRuntimeTaskPayloadJSONFallsBackToRawPayloadJSON(t *testing.T) {
	payload := ParseRuntimeTaskPayloadJSON("not-json")
	if payload["raw_payload_json"] != "not-json" {
		t.Fatalf("expected raw payload fallback, got %#v", payload)
	}
}

func TestRuntimeTaskParsePayloadJSONParsesStructuredBody(t *testing.T) {
	task := &RuntimeTask{PayloadJSON: `{"request_data":{"queries":[{"type":"scene_state","node_id":"scene_inn"}]}}`}
	payload := task.ParsePayloadJSON()
	requestData, ok := payload["request_data"].(map[string]any)
	if !ok {
		t.Fatalf("expected request_data map, got %#v", payload)
	}
	queries, ok := requestData["queries"].([]any)
	if !ok || len(queries) != 1 {
		t.Fatalf("expected one query, got %#v", requestData)
	}
}
