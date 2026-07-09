package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetNodeIncludesDiagnostics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"node":                       map[string]any{"id": "node-1", "world_id": "world-1", "name": "Guard", "node_type": "npc"},
			"relation_validation_issues": []map[string]any{{"severity": "warning", "code": "multiple_located_at_edges", "message": "too many locations"}},
			"graph_context_preview": map[string]any{
				"primary_parent_chain": []string{"Guard(npc)", "World(world)"},
				"summary":              []string{"identity: Guard(npc) > World(world)"},
			},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	detail, err := client.GetNode("node-1")
	if err != nil {
		t.Fatalf("get node: %v", err)
	}
	if len(detail.RelationValidationIssues) != 1 || detail.RelationValidationIssues[0].Code != "multiple_located_at_edges" {
		t.Fatalf("unexpected relation validation issues: %#v", detail.RelationValidationIssues)
	}
	if detail.GraphContextPreview == nil || len(detail.GraphContextPreview.Summary) != 1 {
		t.Fatalf("unexpected graph context preview: %#v", detail.GraphContextPreview)
	}
}
