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
		interaction, err := buildCanonicalInteractionContext(interactionContractInput{
			ActorNodeID:           req.ActorNodeID,
			TargetNodeID:          req.TargetNodeID,
			SceneNodeID:           req.SceneNodeID,
			ParticipantNodeIDs:    req.ParticipantNodeIDs,
			Mode:                  req.Mode,
			AudienceScope:         req.AudienceScope,
			TurnIndex:             req.TurnIndex,
			Event:                 req.Event,
			FallbackTargetToActor: false,
		}, ctx.Interaction)
		if err != nil {
			handleInvokeContractError(w, err)
			return
		}
		ctx.Interaction = interaction

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
		if err := validateInvokeRequestContract(invokeReq); err != nil {
			handleInvokeContractError(w, err)
			return
		}
		resp, err := p.Execute(invokeReq)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
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
