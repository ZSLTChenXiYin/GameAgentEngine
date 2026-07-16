package sdk

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientInterpretPlayerInput(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/player/input/interpret" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"request_id":"req-1","task_type":"custom","execution_mode":"debug","reply":"ok","player_intent":{"intent":{"type":"speech","actor_node_id":"player_001"}}}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dev-key")
	resp, err := client.InterpretPlayerInput(&PlayerInputInterpretRequest{
		WorldID:      "world-1",
		PlayerNodeID: "player_001",
		Message:      "我问老板今晚见过谁。",
	})
	if err != nil {
		t.Fatalf("InterpretPlayerInput returned error: %v", err)
	}
	if resp.PlayerIntent == nil || resp.PlayerIntent.Intent == nil || resp.PlayerIntent.Intent.ActorNodeID != "player_001" {
		t.Fatalf("unexpected player_intent response: %#v", resp)
	}
}

func TestClientExecuteInteraction(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/interactions/execute" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"request_id":"req-2","task_type":"npc_dialogue","execution_mode":"debug","reply":"Innkeeper responds."}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "dev-key")
	resp, err := client.ExecuteInteraction(&InteractionExecuteRequest{
		WorldID:     "world-1",
		ActorNodeID: "player_001",
		TargetNodeID:"npc_innkeeper",
		Message:     "今晚见过这把刀的主人吗？",
	})
	if err != nil {
		t.Fatalf("ExecuteInteraction returned error: %v", err)
	}
	if resp == nil || resp.TaskType != "npc_dialogue" || resp.Reply != "Innkeeper responds." {
		t.Fatalf("unexpected interaction response: %#v", resp)
	}
}
