package workertest

import "testing"

func TestQueryWithValues(t *testing.T) {
	got := QueryWithValues("/api/v1/runtime/tasks", map[string]string{"world_id": "w1", "limit": "1"})
	if got != "/api/v1/runtime/tasks?limit=1&world_id=w1" && got != "/api/v1/runtime/tasks?world_id=w1&limit=1" {
		t.Fatalf("unexpected query string: %s", got)
	}
}
