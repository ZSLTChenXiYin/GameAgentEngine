package engine

import (
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

// CanonicalParticipantNodeIDs merges participant lists, drops blanks, and preserves first-seen order.
func CanonicalParticipantNodeIDs(groups ...[]string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	for _, group := range groups {
		for _, value := range group {
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			result = append(result, trimmed)
		}
	}
	return result
}

func InferInteractionMode(participantNodeIDs []string) string {
	if len(CanonicalParticipantNodeIDs(participantNodeIDs)) > 2 {
		return sdk.InteractionModeGroupChat
	}
	return sdk.InteractionModeDirectDialogue
}

func InferInteractionAudienceScope(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), sdk.InteractionModeGroupChat) {
		return sdk.InteractionAudiencePublic
	}
	return sdk.InteractionAudiencePrivate
}

func NormalizeInteractionSemantics(explicitMode string, explicitAudienceScope string, explicitParticipants []string, fallbackParticipants ...string) (string, string, []string) {
	participants := CanonicalParticipantNodeIDs(explicitParticipants, fallbackParticipants)
	mode := strings.TrimSpace(explicitMode)
	if mode == "" {
		mode = InferInteractionMode(participants)
	}
	audienceScope := strings.TrimSpace(explicitAudienceScope)
	if audienceScope == "" {
		audienceScope = InferInteractionAudienceScope(mode)
	}
	return mode, audienceScope, participants
}
