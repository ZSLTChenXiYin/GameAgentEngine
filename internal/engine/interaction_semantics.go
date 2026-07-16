package engine

import "strings"

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
		return "group_chat"
	}
	return "direct_dialogue"
}

func InferInteractionAudienceScope(mode string) string {
	if strings.EqualFold(strings.TrimSpace(mode), "group_chat") {
		return "public"
	}
	return "private"
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
