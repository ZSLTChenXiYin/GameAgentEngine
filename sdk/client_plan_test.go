package sdk

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPendingPlansUsesPendingEndpoint(t *testing.T) {
	var gotMethod string
	var gotPath string
	var gotQuery string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"plan_id":   "plan-1",
			"world_id":  "world-1",
			"task_type": "world_tick",
			"status":    "pending",
		}})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	plans, err := client.ListPendingPlans("world-1")
	if err != nil {
		t.Fatalf("list pending plans: %v", err)
	}
	if len(plans) != 1 || plans[0].PlanID != "plan-1" {
		t.Fatalf("unexpected plans: %#v", plans)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/plans/pending" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotQuery != "world_id=world-1" {
		t.Fatalf("unexpected query: %s", gotQuery)
	}
}

func TestApprovePlanUsesApproveEndpoint(t *testing.T) {
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "approved",
			"plan_id": "plan-1",
			"plan": map[string]any{
				"plan_id":  "plan-1",
				"world_id": "world-1",
				"status":   "approved",
			},
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.ApprovePlan("world-1", "plan-1")
	if err != nil {
		t.Fatalf("approve plan: %v", err)
	}
	if resp.Status != "approved" || resp.PlanID != "plan-1" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if resp.Plan == nil || resp.Plan.Status != "approved" {
		t.Fatalf("unexpected plan payload: %#v", resp.Plan)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/worlds/world-1/plan/approve" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["plan_id"] != "plan-1" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
}

func TestRejectPlanUsesRejectEndpoint(t *testing.T) {
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "rejected",
			"plan_id": "plan-2",
		})
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.RejectPlan("world-1", "plan-2")
	if err != nil {
		t.Fatalf("reject plan: %v", err)
	}
	if resp.Status != "rejected" || resp.PlanID != "plan-2" {
		t.Fatalf("unexpected response: %#v", resp)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("unexpected method: %s", gotMethod)
	}
	if gotPath != "/api/v1/worlds/world-1/plan/reject" {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotBody["plan_id"] != "plan-2" {
		t.Fatalf("unexpected body: %#v", gotBody)
	}
}
