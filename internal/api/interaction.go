package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

type interactionExecuteRequest struct {
	WorldID            string                   `json:"world_id"`
	ActorNodeID        string                   `json:"actor_node_id"`
	TargetNodeID       string                   `json:"target_node_id"`
	SceneNodeID        string                   `json:"scene_node_id,omitempty"`
	SessionID          string                   `json:"session_id,omitempty"`
	TaskType           string                   `json:"task_type,omitempty"`
	Message            string                   `json:"message"`
	ParticipantNodeIDs []string                 `json:"participant_node_ids,omitempty"`
	Mode               string                   `json:"mode,omitempty"`
	AudienceScope      string                   `json:"audience_scope,omitempty"`
	TurnIndex          int                      `json:"turn_index,omitempty"`
	Event              *engine.InteractionEvent `json:"event,omitempty"`
	Context            *engine.InvokeContext    `json:"context,omitempty"`
}

// MakeExecuteInteractionHandler exposes actor->target interaction as a first-class API.
// Internally it still uses invoke, but callers no longer need to hand-build context.interaction.
func MakeExecuteInteractionHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req interactionExecuteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if strings.TrimSpace(req.WorldID) == "" || strings.TrimSpace(req.ActorNodeID) == "" || strings.TrimSpace(req.TargetNodeID) == "" {
			errorJSON(w, http.StatusBadRequest, "world_id, actor_node_id and target_node_id required")
			return
		}
		if strings.TrimSpace(req.Message) == "" {
			errorJSON(w, http.StatusBadRequest, "message required")
			return
		}

		taskType := strings.TrimSpace(req.TaskType)
		if taskType == "" {
			taskType = string(engine.TaskNPCDialogue)
		}
		if taskType != string(engine.TaskNPCDialogue) {
			errorJSONCode(w, http.StatusBadRequest, "invalid_task_type", "interaction endpoint currently only supports task_type=npc_dialogue")
			return
		}

		ctx := req.Context
		if ctx == nil {
			ctx = &engine.InvokeContext{}
		}
		interaction := buildInteractionContext(req, ctx.Interaction)
		ctx.Interaction = interaction

		if ctx.PipelineMode != "" && !engine.IsValidPipelineMode(string(ctx.PipelineMode)) {
			errorJSONCode(w, http.StatusBadRequest, "invalid_pipeline_mode", "context.pipeline_mode must be one of: vertical, polling, full")
			return
		}
		if err := engine.ValidateDynamicInterfaces(ctx.DynamicInterfaces); err != nil {
			errorJSONCode(w, http.StatusBadRequest, "invalid_dynamic_interfaces", err.Error())
			return
		}
		if err := engine.ValidateInteractionContext(ctx.Interaction); err != nil {
			errorJSONCode(w, http.StatusBadRequest, "invalid_interaction", err.Error())
			return
		}

		invokeReq := &engine.InvokeRequest{
			WorldID:   strings.TrimSpace(req.WorldID),
			NodeID:    strings.TrimSpace(req.TargetNodeID),
			TaskType:  engine.TaskNPCDialogue,
			SessionID: strings.TrimSpace(req.SessionID),
			Context:   ctx,
			Messages: []engine.ChatMessage{{
				Role:    "user",
				Content: strings.TrimSpace(req.Message),
			}},
		}
		resp, err := p.Execute(invokeReq)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func buildInteractionContext(req interactionExecuteRequest, existing *engine.InteractionContext) *engine.InteractionContext {
	interaction := &engine.InteractionContext{}
	if existing != nil {
		copied := *existing
		if existing.ParticipantNodeIDs != nil {
			copied.ParticipantNodeIDs = append([]string(nil), existing.ParticipantNodeIDs...)
		}
		if existing.Event != nil {
			eventCopy := *existing.Event
			if existing.Event.Args != nil {
				eventCopy.Args = make(map[string]any, len(existing.Event.Args))
				for k, v := range existing.Event.Args {
					eventCopy.Args[k] = v
				}
			}
			copied.Event = &eventCopy
		}
		interaction = &copied
	}

	interaction.SpeakerNodeID = firstNonEmptyString(strings.TrimSpace(req.ActorNodeID), interaction.SpeakerNodeID)
	interaction.TargetNodeID = firstNonEmptyString(strings.TrimSpace(req.TargetNodeID), interaction.TargetNodeID)
	interaction.SceneNodeID = firstNonEmptyString(strings.TrimSpace(req.SceneNodeID), interaction.SceneNodeID)
	if req.TurnIndex > 0 {
		interaction.TurnIndex = req.TurnIndex
	}

	participants := req.ParticipantNodeIDs
	if len(participants) == 0 {
		participants = interaction.ParticipantNodeIDs
	}
	interaction.ParticipantNodeIDs = uniqueNonEmptyStrings(participants, []string{interaction.SpeakerNodeID, interaction.TargetNodeID})

	mode := strings.TrimSpace(req.Mode)
	if mode == "" {
		mode = strings.TrimSpace(interaction.Mode)
	}
	if mode == "" {
		mode = inferInteractionMode(interaction.ParticipantNodeIDs)
	}
	interaction.Mode = mode

	audience := strings.TrimSpace(req.AudienceScope)
	if audience == "" {
		audience = strings.TrimSpace(interaction.AudienceScope)
	}
	if audience == "" {
		audience = inferInteractionAudienceScope(mode)
	}
	interaction.AudienceScope = audience

	if req.Event != nil {
		interaction.Event = req.Event
	}
	if interaction.Event == nil {
		interaction.Event = &engine.InteractionEvent{Type: "speech"}
	}
	if strings.TrimSpace(interaction.Event.Type) == "" {
		interaction.Event.Type = "speech"
	}
	return interaction
}

func inferInteractionMode(participants []string) string {
	if len(participants) > 2 {
		return "group_chat"
	}
	return "direct_dialogue"
}

func inferInteractionAudienceScope(mode string) string {
	if strings.TrimSpace(mode) == "group_chat" {
		return "public"
	}
	return "private"
}

func uniqueNonEmptyStrings(values ...[]string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0)
	appendOne := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	for _, group := range values {
		for _, value := range group {
			appendOne(value)
		}
	}
	return result
}
