package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRelationRejectsInvalidRelationType(t *testing.T) {
	client := NewClient("http://example.com", "test-key")
	if _, err := client.CreateRelation("world-1", "source-1", "target-1", "bad_relation", 1); err == nil {
		t.Fatal("expected invalid relation_type error")
	}
}

func TestUpdateRelationRejectsInvalidRelationType(t *testing.T) {
	client := NewClient("http://example.com", "test-key")
	relType := "bad_relation"
	if _, err := client.UpdateRelation("rel-1", nil, nil, &relType, nil, nil); err == nil {
		t.Fatal("expected invalid relation_type error")
	}
}

func TestPropagateMemoryDefaultsToUpwardMode(t *testing.T) {
	var payload map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	if err := client.PropagateMemory("memory-1", "", nil, nil, 0, false); err != nil {
		t.Fatalf("propagate memory: %v", err)
	}
	if payload["mode"] != PropagationModeUpward {
		t.Fatalf("expected default upward mode, got %#v", payload)
	}
}

func TestPropagateMemoryRejectsInvalidMode(t *testing.T) {
	client := NewClient("http://example.com", "test-key")
	if err := client.PropagateMemory("memory-1", "bad_mode", nil, nil, 0, false); err == nil {
		t.Fatal("expected invalid propagation mode error")
	}
}

