package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForkWorldUsesForkEndpointAndLockFlag(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "forked-world", "world_id": "forked-world", "name": "Forked", "node_type": "world"})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.ForkWorld("world-1", "Forked", true)
	if err != nil {
		t.Fatalf("fork world: %v", err)
	}
	if result.ID != "forked-world" {
		t.Fatalf("unexpected result id: %s", result.ID)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/worlds/world-1/fork" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["name"] != "Forked" {
		t.Fatalf("unexpected name payload: %#v", gotBody)
	}
	if gotBody["lock_world"] != true {
		t.Fatalf("expected lock_world=true payload, got %#v", gotBody)
	}
}

func TestCreateWorldSnapshotUsesSnapshotEndpoint(t *testing.T) {
	var gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "snapshot-world", "world_id": "snapshot-world", "name": "Save 1", "node_type": "world"})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.CreateWorldSnapshot("world-2", "Save 1", false)
	if err != nil {
		t.Fatalf("create world snapshot: %v", err)
	}
	if result.ID != "snapshot-world" {
		t.Fatalf("unexpected result id: %s", result.ID)
	}
	if gotPath != "/api/v1/worlds/world-2/snapshots" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestRestoreWorldUsesRestoreEndpoint(t *testing.T) {
	var gotPath string
	var gotBody map[string]any

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "restored-world", "world_id": "restored-world", "name": "Restored", "node_type": "world"})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.RestoreWorld("world-3", "Restored", true)
	if err != nil {
		t.Fatalf("restore world: %v", err)
	}
	if result.ID != "restored-world" {
		t.Fatalf("unexpected result id: %s", result.ID)
	}
	if gotPath != "/api/v1/worlds/world-3/restore" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["lock_world"] != true {
		t.Fatalf("expected lock_world=true payload, got %#v", gotBody)
	}
}

func TestValidateWorldSnapshotUsesValidationEndpoint(t *testing.T) {
	var gotMethod string
	var gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"snapshot_world_id": "snapshot-world",
			"source_world_id":   "source-world",
			"snapshot_name":     "Save 1",
			"reason":            "save_snapshot",
			"valid":             true,
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.ValidateWorldSnapshot("snapshot-world")
	if err != nil {
		t.Fatalf("validate snapshot: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected snapshot to be valid: %#v", result)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/worlds/snapshot-world/snapshot-validation" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestGetWorldSnapshotMetadataUsesMetadataEndpoint(t *testing.T) {
	var gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                "meta-1",
			"snapshot_world_id": "snapshot-world",
			"source_world_id":   "source-world",
			"snapshot_name":     "Save 1",
			"reason":            "save_snapshot",
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.GetWorldSnapshotMetadata("snapshot-world")
	if err != nil {
		t.Fatalf("get snapshot metadata: %v", err)
	}
	if result.ID != "meta-1" {
		t.Fatalf("unexpected metadata result: %#v", result)
	}
	if gotPath != "/api/v1/worlds/snapshot-world/snapshot-metadata" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestListWorldSnapshotsUsesSnapshotsEndpoint(t *testing.T) {
	var gotPath string
	var gotMethod string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":                "snap-1",
			"snapshot_world_id": "snapshot-world",
			"source_world_id":   "source-world",
			"snapshot_name":     "Save 1",
			"reason":            "save_snapshot",
		}})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	result, err := client.ListWorldSnapshots("source-world")
	if err != nil {
		t.Fatalf("list snapshots: %v", err)
	}
	if len(result) != 1 || result[0].ID != "snap-1" {
		t.Fatalf("unexpected snapshot list: %#v", result)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/worlds/source-world/snapshots" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}

func TestDeleteWorldSnapshotUsesDeleteEndpoint(t *testing.T) {
	var gotMethod string
	var gotPath string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"deleted"}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	if err := client.DeleteWorldSnapshot("snapshot-world"); err != nil {
		t.Fatalf("delete snapshot: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/worlds/snapshot-world/snapshot" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
}
