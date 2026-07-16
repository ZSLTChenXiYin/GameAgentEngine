package sdk

import "testing"

func TestInteractionConstantsRemainStable(t *testing.T) {
	if InteractionModeDirectDialogue != "direct_dialogue" || InteractionModeGroupChat != "group_chat" {
		t.Fatalf("unexpected interaction modes")
	}
	if InteractionAudiencePrivate != "private" || InteractionAudiencePublic != "public" || InteractionAudienceWhisper != "whisper" {
		t.Fatalf("unexpected interaction audiences")
	}
	if InteractionEventSpeech != "speech" || InteractionEventGift != "gift" || InteractionEventShowItem != "show_item" || InteractionEventTradeRequest != "trade_request" || InteractionEventThreaten != "threaten" {
		t.Fatalf("unexpected interaction event constants")
	}
}
