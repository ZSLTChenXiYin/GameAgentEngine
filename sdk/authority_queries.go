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

// AuthorityQueryResponse describes the typed worker-side response for one
// authority query batch while remaining JSON-compatible for callbacks.
type AuthorityQueryResponse struct {
	Status       string                 `json:"status,omitempty"`
	LongRunning  bool                   `json:"long_running"`
	WorldID      string                 `json:"world_id,omitempty"`
	Queries      []AuthorityQueryResult `json:"queries,omitempty"`
	RequestError string                 `json:"request_error,omitempty"`
	StateError   string                 `json:"state_error,omitempty"`
}

// AuthorityQueryResult captures one resolved authority-side fact lookup.
type AuthorityQueryResult struct {
	Type        string                    `json:"type"`
	NodeID      string                    `json:"node_id,omitempty"`
	HP          *int                      `json:"hp,omitempty"`
	MaxHP       *int                      `json:"max_hp,omitempty"`
	Inventory   []AuthorityInventoryEntry `json:"inventory,omitempty"`
	Money       *int                      `json:"money,omitempty"`
	LocationID  string                    `json:"location_id,omitempty"`
	Scene       *AuthoritySceneSnapshot   `json:"scene,omitempty"`
	Status      string                    `json:"status,omitempty"`
	Stage       string                    `json:"stage,omitempty"`
	ItemID      string                    `json:"item_id,omitempty"`
	Present     *bool                     `json:"present,omitempty"`
	Unsupported bool                      `json:"unsupported,omitempty"`
}

// AuthorityInventoryEntry mirrors the callback JSON shape used for inventory lookups.
type AuthorityInventoryEntry struct {
	ItemID   string         `json:"item_id"`
	Quantity int            `json:"quantity,omitempty"`
	Equipped bool           `json:"equipped,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// AuthoritySceneSnapshot mirrors the callback JSON shape used for scene/room lookups.
type AuthoritySceneSnapshot struct {
	ID          string         `json:"id"`
	Name        string         `json:"name,omitempty"`
	Kind        string         `json:"kind,omitempty"`
	Occupants   []string       `json:"occupants,omitempty"`
	Flags       map[string]any `json:"flags,omitempty"`
	Description string         `json:"description,omitempty"`
}

// Fields exposes the response as callback-ready fields while keeping internal generation typed.
func (r AuthorityQueryResponse) Fields() map[string]any {
	fields := map[string]any{
		"status":       r.Status,
		"long_running": r.LongRunning,
	}
	if strings.TrimSpace(r.WorldID) != "" {
		fields["world_id"] = r.WorldID
	}
	if len(r.Queries) > 0 {
		fields["queries"] = append([]AuthorityQueryResult(nil), r.Queries...)
	}
	if strings.TrimSpace(r.RequestError) != "" {
		fields["request_error"] = r.RequestError
	}
	if strings.TrimSpace(r.StateError) != "" {
		fields["state_error"] = r.StateError
	}
	return fields
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
