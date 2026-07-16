package sdk

import "testing"

func TestRequiresFollowupInteractionUsesCanonicalRule(t *testing.T) {
	for _, stepType := range []string{
		PlayerIntentTypeSpeech,
		PlayerIntentTypeShowItem,
		PlayerIntentTypeGift,
		PlayerIntentTypeTradeRequest,
		PlayerIntentTypeThreaten,
	} {
		if !RequiresFollowupInteraction(stepType, nil) {
			t.Fatalf("expected %s to require follow-up interaction", stepType)
		}
	}
	for _, stepType := range []string{PlayerIntentTypeMove, PlayerIntentTypeInspect, PlayerIntentTypeUseItem} {
		if RequiresFollowupInteraction(stepType, nil) {
			t.Fatalf("did not expect %s to require follow-up interaction", stepType)
		}
	}
	if !RequiresFollowupInteraction(PlayerIntentTypeMove, &SuggestedInteraction{EventType: InteractionEventSpeech}) {
		t.Fatal("expected suggested interaction event type to force follow-up interaction")
	}
}

func TestIsFollowupInteractionEventType(t *testing.T) {
	if !IsFollowupInteractionEventType(InteractionEventSpeech) {
		t.Fatal("expected speech to be recognized as follow-up interaction event type")
	}
	if IsFollowupInteractionEventType(PlayerIntentTypeMove) {
		t.Fatal("did not expect move to be recognized as follow-up interaction event type")
	}
}
