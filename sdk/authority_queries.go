package sdk

import (
	"encoding/json"
	"strings"
)

// AuthorityDataRequest describes the structured payload sent through the
// standard game-client authority query interface.
type AuthorityDataRequest struct {
	Queries []AuthorityQuery `json:"queries,omitempty"`
}

// AuthorityQuery describes one authority-side fact lookup request.
type AuthorityQuery struct {
	Type   string `json:"type"`
	NodeID string `json:"node_id,omitempty"`
	Filter string `json:"filter,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// DecodeAuthorityDataRequest unwraps either a raw request payload or a
// runtime-task payload containing request_data.
func DecodeAuthorityDataRequest(payload map[string]any) (*AuthorityDataRequest, error) {
	requestData := authorityRequestDataPayload(payload)
	if len(requestData) == 0 {
		return &AuthorityDataRequest{}, nil
	}
	data, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	var req AuthorityDataRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	req.Queries = CompactAuthorityQueries(req.Queries)
	return &req, nil
}

// ExtractAuthorityQueries returns the normalized authority query list from a
// runtime-task payload or plain request_data payload.
func ExtractAuthorityQueries(payload map[string]any) ([]AuthorityQuery, error) {
	req, err := DecodeAuthorityDataRequest(payload)
	if err != nil || req == nil {
		return nil, err
	}
	return req.Queries, nil
}

// CompactAuthorityQueries drops empty query entries and trims string fields.
func CompactAuthorityQueries(queries []AuthorityQuery) []AuthorityQuery {
	if len(queries) == 0 {
		return nil
	}
	compact := make([]AuthorityQuery, 0, len(queries))
	for _, query := range queries {
		query.Type = strings.TrimSpace(query.Type)
		query.NodeID = strings.TrimSpace(query.NodeID)
		query.Filter = strings.TrimSpace(query.Filter)
		if query.Type == "" {
			continue
		}
		compact = append(compact, query)
	}
	if len(compact) == 0 {
		return nil
	}
	return compact
}

func authorityRequestDataPayload(payload map[string]any) map[string]any {
	if payload == nil {
		return nil
	}
	requestData, ok := payload["request_data"].(map[string]any)
	if ok && requestData != nil {
		return requestData
	}
	return payload
}
