package sdk

import "testing"

func TestPlayerIntentInterpretationStructFields(t *testing.T) {
	payload := &PlayerIntentInterpretation{
		Intent: &PlayerIntent{
			Type:         "composite",
			ActorNodeID:  "player_001",
			SceneNodeID:  "scene_inn",
			TargetNodeID: "npc_innkeeper",
			Summary:      "show knife and ask innkeeper",
			RiskLevel:    "medium",
			Confidence:   0.9,
			Steps: []PlayerIntentStep{
				{Type: "show_item", TargetNodeID: "npc_innkeeper", ItemID: "knife_bloody"},
				{Type: "speech", Content: "今晚有没有见过这把刀的主人？"},
			},
		},
		MissingFacts:         []MissingFact{{Type: "scene_state", NodeID: "scene_inn"}},
		SuggestedInteraction: &SuggestedInteraction{Mode: "direct_dialogue", EventType: "show_item", AudienceScope: "private", TargetNodeID: "npc_innkeeper"},
	}
	if payload.Intent == nil || payload.Intent.Type != "composite" {
		t.Fatalf("unexpected intent payload: %#v", payload)
	}
	if len(payload.Intent.Steps) != 2 {
		t.Fatalf("expected two steps, got %#v", payload.Intent.Steps)
	}
	if payload.SuggestedInteraction == nil || payload.SuggestedInteraction.EventType != "show_item" {
		t.Fatalf("unexpected suggested interaction: %#v", payload.SuggestedInteraction)
	}
}
