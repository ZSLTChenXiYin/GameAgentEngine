package sdk

import "testing"

func TestAuthorityQueryResponseFields(t *testing.T) {
	hp := 15
	present := true
	resp := AuthorityQueryResponse{
		Status:      "success",
		LongRunning: false,
		WorldID:     "world_1",
		Queries: []AuthorityQueryResult{{
			Type:    AuthorityQueryPlayerState,
			NodeID:  "player_1",
			HP:      &hp,
			Present: &present,
		}},
	}
	fields := resp.Fields()
	if fields["status"] != "success" || fields["world_id"] != "world_1" {
		t.Fatalf("unexpected top-level fields: %#v", fields)
	}
	queries, ok := fields["queries"].([]AuthorityQueryResult)
	if !ok || len(queries) != 1 || queries[0].HP == nil || *queries[0].HP != 15 {
		t.Fatalf("unexpected queries payload: %#v", fields["queries"])
	}
	if _, ok := fields["request_error"]; ok {
		t.Fatalf("did not expect empty request_error in fields: %#v", fields)
	}
}
