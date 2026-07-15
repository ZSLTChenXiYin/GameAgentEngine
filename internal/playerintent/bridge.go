package playerintent

import (
	"fmt"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
)

func BuildInteractionSpec(payload *sdk.PlayerIntentInterpretation, actorNodeID string, fallbackSceneID string) (*InteractionSpec, error) {
	if payload == nil || payload.Intent == nil {
		return nil, fmt.Errorf("player intent payload required")
	}
	intent := payload.Intent
	steps := intentExecutionSteps(intent)
	if len(steps) == 0 {
		return nil, fmt.Errorf("player intent has no executable steps")
	}
	primary := steps[0]
	spec := &InteractionSpec{
		Mode:          "direct_dialogue",
		AudienceScope: "private",
		EventType:     primary.Type,
		TargetNodeID:  firstNonEmpty(primary.TargetNodeID, intent.TargetNodeID),
	}
	if payload.SuggestedInteraction != nil {
		spec.Mode = firstNonEmpty(payload.SuggestedInteraction.Mode, spec.Mode)
		spec.AudienceScope = firstNonEmpty(payload.SuggestedInteraction.AudienceScope, spec.AudienceScope)
		spec.EventType = firstNonEmpty(payload.SuggestedInteraction.EventType, spec.EventType)
		spec.TargetNodeID = firstNonEmpty(payload.SuggestedInteraction.TargetNodeID, spec.TargetNodeID)
	}
	if strings.TrimSpace(spec.TargetNodeID) == "" {
		return nil, fmt.Errorf("interaction target required")
	}
	if strings.TrimSpace(primary.ItemID) != "" {
		spec.ItemID = primary.ItemID
	}
	spec.Participants = uniqueIDs(actorNodeID, spec.TargetNodeID)
	if strings.EqualFold(strings.TrimSpace(spec.Mode), "group_chat") {
		spec.AudienceScope = firstNonEmpty(spec.AudienceScope, "public")
	}
	spec.Input = buildInteractionInput(intent, steps)
	_ = fallbackSceneID
	return spec, nil
}

func buildInteractionInput(intent *sdk.PlayerIntent, steps []sdk.PlayerIntentStep) string {
	if intent == nil || len(steps) == 0 {
		return ""
	}
	if len(steps) == 1 {
		step := steps[0]
		if strings.TrimSpace(step.Content) != "" {
			return strings.TrimSpace(step.Content)
		}
		if strings.TrimSpace(intent.Summary) != "" {
			return strings.TrimSpace(intent.Summary)
		}
		return fmt.Sprintf("player intent: %s", step.Type)
	}
	parts := make([]string, 0, len(steps))
	for _, step := range steps {
		segment := strings.TrimSpace(step.Content)
		if segment == "" {
			segment = step.Type
			if strings.TrimSpace(step.ItemID) != "" {
				segment += "(" + strings.TrimSpace(step.ItemID) + ")"
			}
		}
		parts = append(parts, segment)
	}
	return strings.Join(parts, "；")
}

func uniqueIDs(values ...string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
