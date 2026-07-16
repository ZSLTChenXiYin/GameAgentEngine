package api

import (
	"fmt"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

type interactionContractInput struct {
	ActorNodeID           string
	TargetNodeID          string
	SceneNodeID           string
	ParticipantNodeIDs    []string
	Mode                  string
	AudienceScope         string
	TurnIndex             int
	Event                 *engine.InteractionEvent
	InputSource           string
	FallbackTargetToActor bool
}

func buildCanonicalInteractionContext(input interactionContractInput, existing *engine.InteractionContext) (*engine.InteractionContext, error) {
	interaction := &engine.InteractionContext{}
	if existing != nil {
		copied := *existing
		if existing.ParticipantNodeIDs != nil {
			copied.ParticipantNodeIDs = append([]string(nil), existing.ParticipantNodeIDs...)
		}
		if existing.Event != nil {
			eventCopy := *existing.Event
			if existing.Event.Args != nil {
				eventCopy.Args = cloneStringAnyMap(existing.Event.Args)
			}
			copied.Event = &eventCopy
		}
		interaction = &copied
	}

	if err := validateExplicitInteractionOverrides(input, interaction); err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.ActorNodeID) != "" {
		interaction.SpeakerNodeID = strings.TrimSpace(input.ActorNodeID)
	} else {
		interaction.SpeakerNodeID = firstNonEmptyString(interaction.SpeakerNodeID)
	}
	targetNodeID := strings.TrimSpace(input.TargetNodeID)
	if targetNodeID == "" {
		targetNodeID = firstNonEmptyString(interaction.TargetNodeID)
	}
	if targetNodeID == "" && input.FallbackTargetToActor {
		targetNodeID = interaction.SpeakerNodeID
	}
	interaction.TargetNodeID = targetNodeID
	if strings.TrimSpace(input.SceneNodeID) != "" {
		interaction.SceneNodeID = strings.TrimSpace(input.SceneNodeID)
	} else {
		interaction.SceneNodeID = firstNonEmptyString(interaction.SceneNodeID)
	}
	if input.TurnIndex > 0 {
		interaction.TurnIndex = input.TurnIndex
	}

	participants := input.ParticipantNodeIDs
	if len(participants) == 0 {
		participants = interaction.ParticipantNodeIDs
	}
	interaction.ParticipantNodeIDs = uniqueNonEmptyStrings(participants, []string{interaction.SpeakerNodeID, interaction.TargetNodeID})

	mode := strings.TrimSpace(input.Mode)
	if mode == "" {
		mode = strings.TrimSpace(interaction.Mode)
	}
	if mode == "" {
		mode = inferInteractionMode(interaction.ParticipantNodeIDs)
	}
	interaction.Mode = mode

	audience := strings.TrimSpace(input.AudienceScope)
	if audience == "" {
		audience = strings.TrimSpace(interaction.AudienceScope)
	}
	if audience == "" {
		audience = inferInteractionAudienceScope(mode)
	}
	interaction.AudienceScope = audience

	if input.Event != nil {
		interaction.Event = cloneInteractionEvent(input.Event)
	}
	if interaction.Event == nil {
		interaction.Event = &engine.InteractionEvent{Type: "speech"}
	}
	if strings.TrimSpace(interaction.Event.Type) == "" {
		interaction.Event.Type = "speech"
	}
	if strings.TrimSpace(input.InputSource) != "" {
		if interaction.Event.Args == nil {
			interaction.Event.Args = map[string]any{}
		}
		interaction.Event.Args["input_source"] = strings.TrimSpace(input.InputSource)
	}
	return interaction, nil
}

func validateInvokeRequestContract(req *engine.InvokeRequest) error {
	if req == nil {
		return fmt.Errorf("request required")
	}
	if strings.TrimSpace(req.WorldID) == "" || strings.TrimSpace(req.NodeID) == "" {
		return fmt.Errorf("world_id and node_id required")
	}
	if req.Context == nil {
		return nil
	}
	if req.Context.PipelineMode != "" && !engine.IsValidPipelineMode(string(req.Context.PipelineMode)) {
		return fmt.Errorf("context.pipeline_mode must be one of: vertical, polling, full")
	}
	if err := engine.ValidateDynamicInterfaces(req.Context.DynamicInterfaces); err != nil {
		return err
	}
	if err := engine.ValidateInteractionContext(req.Context.Interaction); err != nil {
		return err
	}
	interaction := req.Context.Interaction
	if interaction == nil {
		return nil
	}
	nodeID := strings.TrimSpace(req.NodeID)
	if req.TaskType == engine.TaskNPCDialogue {
		targetID := strings.TrimSpace(interaction.TargetNodeID)
		if targetID != "" && nodeID != targetID {
			return fmt.Errorf("npc_dialogue node_id must match interaction.target_node_id")
		}
	}
	if req.Context.PlayerInputInterpret {
		speakerID := strings.TrimSpace(interaction.SpeakerNodeID)
		if speakerID != "" && nodeID != speakerID {
			return fmt.Errorf("player_input_interpret node_id must match interaction.speaker_node_id")
		}
	}
	return nil
}

func cloneInteractionEvent(event *engine.InteractionEvent) *engine.InteractionEvent {
	if event == nil {
		return nil
	}
	cloned := *event
	if event.Args != nil {
		cloned.Args = cloneStringAnyMap(event.Args)
	}
	return &cloned
}

func cloneStringAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func validateExplicitInteractionOverrides(input interactionContractInput, existing *engine.InteractionContext) error {
	if existing == nil {
		return nil
	}
	if err := validateInteractionFieldOverride("speaker_node_id", input.ActorNodeID, existing.SpeakerNodeID); err != nil {
		return err
	}
	targetNodeID := strings.TrimSpace(input.TargetNodeID)
	if targetNodeID == "" && input.FallbackTargetToActor {
		targetNodeID = strings.TrimSpace(input.ActorNodeID)
	}
	if err := validateInteractionFieldOverride("target_node_id", targetNodeID, existing.TargetNodeID); err != nil {
		return err
	}
	if err := validateInteractionFieldOverride("scene_node_id", input.SceneNodeID, existing.SceneNodeID); err != nil {
		return err
	}
	if err := validateInteractionFieldOverride("mode", input.Mode, existing.Mode); err != nil {
		return err
	}
	if err := validateInteractionFieldOverride("audience_scope", input.AudienceScope, existing.AudienceScope); err != nil {
		return err
	}
	if input.TurnIndex > 0 && existing.TurnIndex > 0 && input.TurnIndex != existing.TurnIndex {
		return fmt.Errorf("interaction.turn_index conflicts with request turn_index")
	}
	if len(input.ParticipantNodeIDs) > 0 && len(existing.ParticipantNodeIDs) > 0 {
		canonicalInput := uniqueNonEmptyStrings(input.ParticipantNodeIDs)
		canonicalExisting := uniqueNonEmptyStrings(existing.ParticipantNodeIDs)
		if strings.Join(canonicalInput, "\x00") != strings.Join(canonicalExisting, "\x00") {
			return fmt.Errorf("interaction.participant_node_ids conflicts with request participant_node_ids")
		}
	}
	if input.Event != nil && existing.Event != nil {
		if err := validateInteractionFieldOverride("event.type", input.Event.Type, existing.Event.Type); err != nil {
			return err
		}
	}
	return nil
}

func validateInteractionFieldOverride(field string, requested string, existing string) error {
	requested = strings.TrimSpace(requested)
	existing = strings.TrimSpace(existing)
	if requested == "" || existing == "" || requested == existing {
		return nil
	}
	return fmt.Errorf("interaction.%s conflicts with request %s", field, field)
}
