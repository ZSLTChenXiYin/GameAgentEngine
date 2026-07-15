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
		interaction := ctx.Interaction
		if interaction == nil {
			participants := append([]string(nil), req.ParticipantNodeIDs...)
			if len(participants) == 0 {
				participants = []string{strings.TrimSpace(req.PlayerNodeID)}
				if strings.TrimSpace(req.TargetNodeID) != "" {
					participants = append(participants, strings.TrimSpace(req.TargetNodeID))
				}
			}
			mode := "direct_dialogue"
			if len(participants) > 2 {
				mode = "group_chat"
			}
			interaction = &engine.InteractionContext{
				Mode:               mode,
				SpeakerNodeID:      strings.TrimSpace(req.PlayerNodeID),
				TargetNodeID:       firstNonEmptyString(strings.TrimSpace(req.TargetNodeID), strings.TrimSpace(req.PlayerNodeID)),
				SceneNodeID:        strings.TrimSpace(req.SceneNodeID),
				ParticipantNodeIDs: participants,
				AudienceScope:      inferPlayerInputAudienceScope(mode),
				Event: &engine.InteractionEvent{
					Type: "speech",
					Args: map[string]any{
						"input_source": "player_input_interpret",
					},
				},
			}
			ctx.Interaction = interaction
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
			NodeID:    strings.TrimSpace(req.PlayerNodeID),
			TaskType:  engine.TaskCustom,
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

func inferPlayerInputAudienceScope(mode string) string {
	if strings.TrimSpace(mode) == "group_chat" {
		return "public"
	}
	return "private"
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
