package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

type playerInputInterpretRequest struct {
	WorldID            string                `json:"world_id"`
	PlayerNodeID       string                `json:"player_node_id"`
	SceneNodeID        string                `json:"scene_node_id,omitempty"`
	TargetNodeID       string                `json:"target_node_id,omitempty"`
	SessionID          string                `json:"session_id,omitempty"`
	Message            string                `json:"message"`
	ParticipantNodeIDs []string              `json:"participant_node_ids,omitempty"`
	Context            *engine.InvokeContext `json:"context,omitempty"`
}

// MakePlayerInputInterpretHandler returns a user-friendly player-input interpretation entrypoint.
// Internally it still reuses the unified invoke pipeline with task_type=custom.
func MakePlayerInputInterpretHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req playerInputInterpretRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "invalid json: "+err.Error())
			return
		}
		if strings.TrimSpace(req.WorldID) == "" || strings.TrimSpace(req.PlayerNodeID) == "" {
			errorJSON(w, http.StatusBadRequest, "world_id and player_node_id required")
			return
		}
		if strings.TrimSpace(req.Message) == "" {
			errorJSON(w, http.StatusBadRequest, "message required")
			return
		}

		ctx := req.Context
		if ctx == nil {
			ctx = &engine.InvokeContext{}
		}
		ctx.PlayerInputInterpret = true
		interaction, err := buildCanonicalInteractionContext(interactionContractInput{
			ActorNodeID:           req.PlayerNodeID,
			TargetNodeID:          req.TargetNodeID,
			SceneNodeID:           req.SceneNodeID,
			ParticipantNodeIDs:    req.ParticipantNodeIDs,
			InputSource:           "player_input_interpret",
			FallbackTargetToActor: true,
		}, ctx.Interaction)
		if err != nil {
			handleInvokeContractError(w, err)
			return
		}
		ctx.Interaction = interaction

		invokeReq := &engine.InvokeRequest{
			WorldID:   strings.TrimSpace(req.WorldID),
			NodeID:    strings.TrimSpace(req.PlayerNodeID),
			TaskType:  engine.TaskCustom,
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

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
