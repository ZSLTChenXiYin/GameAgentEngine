package sdk

import "testing"

func TestDecodeAuthorityDataRequestUnwrapsRequestDataEnvelope(t *testing.T) {
	req, err := DecodeAuthorityDataRequest(map[string]any{
		"request_data": map[string]any{
			"queries": []any{
				map[string]any{"type": AuthorityQuerySceneState, "node_id": "scene_inn"},
				map[string]any{"type": "   ", "node_id": "ignored"},
			},
		},
	})
	if err != nil {
		t.Fatalf("DecodeAuthorityDataRequest returned error: %v", err)
	}
	if req == nil || len(req.Queries) != 1 {
		t.Fatalf("unexpected request: %#v", req)
	}
	if req.Queries[0].Type != AuthorityQuerySceneState || req.Queries[0].NodeID != "scene_inn" {
		t.Fatalf("unexpected query: %#v", req.Queries[0])
	}
}

func TestExtractAuthorityQueriesAcceptsBarePayload(t *testing.T) {
	queries, err := ExtractAuthorityQueries(map[string]any{
		"queries": []any{
			map[string]any{"type": AuthorityQueryPlayerState, "node_id": "player_1", "limit": float64(2)},
		},
	})
	if err != nil {
		t.Fatalf("ExtractAuthorityQueries returned error: %v", err)
	}
	if len(queries) != 1 {
		t.Fatalf("unexpected queries: %#v", queries)
	}
	if queries[0].Type != AuthorityQueryPlayerState || queries[0].NodeID != "player_1" || queries[0].Limit != 2 {
		t.Fatalf("unexpected query payload: %#v", queries[0])
	}
}

func TestCompactAuthorityQueriesTrimsAndDropsEmptyItems(t *testing.T) {
	queries := CompactAuthorityQueries([]AuthorityQuery{
		{Type: "  "},
		{Type: " scene_state ", NodeID: " scene_inn ", Filter: " knife "},
	})
	if len(queries) != 1 {
		t.Fatalf("unexpected compact queries: %#v", queries)
	}
	if queries[0].Type != "scene_state" || queries[0].NodeID != "scene_inn" || queries[0].Filter != "knife" {
		t.Fatalf("unexpected compacted query: %#v", queries[0])
	}
}
