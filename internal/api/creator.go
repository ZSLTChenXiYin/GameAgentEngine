package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
	"gopkg.in/yaml.v3"
)

// CreatorImportHandler handles GameAgentCreator YAML/JSON world import requests.
func CreatorImportHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Format  string `json:"format"`
		Content string `json:"content"`
		Reset   bool   `json:"reset"`
		DryRun  bool   `json:"dry_run"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}
	if req.Content == "" {
		errorJSON(w, 400, "content required")
		return
	}

	var cfg sdk.ImportConfig
	switch req.Format {
	case "yaml", "yml":
		if err := yaml.Unmarshal([]byte(req.Content), &cfg); err != nil {
			errorJSON(w, 400, "invalid yaml: "+err.Error())
			return
		}
	case "json":
		if err := json.Unmarshal([]byte(req.Content), &cfg); err != nil {
			errorJSON(w, 400, "invalid json: "+err.Error())
			return
		}
	default:
		errorJSON(w, 400, "unsupported format: "+req.Format)
		return
	}

	result, err := service.ImportWorld(&cfg, req.Reset, req.DryRun)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, 200, result)
}
