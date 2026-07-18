package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

// MakeColdStartHandler returns a handler that calls Pipeline.ColdStartWorld.
func MakeColdStartHandler(p *engine.Pipeline) http.HandlerFunc {
	type coldStartRequest struct {
		Mode string `json:"mode"` // "initial" or "rebuild"
	}
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		if worldID == "" {
			errorJSON(w, 400, "missing world_id")
			return
		}

		var req coldStartRequest
		if r.Body != nil && r.ContentLength > 0 {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				errorJSON(w, 400, "invalid request body: "+err.Error())
				return
			}
		}
		if req.Mode == "" {
			req.Mode = "initial"
		}
		if req.Mode != "initial" && req.Mode != "rebuild" {
			errorJSON(w, 400, "mode must be 'initial' or 'rebuild'")
			return
		}

		result, err := p.ColdStartWorld(worldID, req.Mode)
		if err != nil {
			errorJSON(w, 500, "cold-start failed: "+err.Error())
			return
		}
		writeJSON(w, 200, result)
	}
}
