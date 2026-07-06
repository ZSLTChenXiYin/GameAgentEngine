package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func initMiddlewareTestDB(t *testing.T) {
	t.Helper()
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
}

func TestIdempotencyMiddlewareReplaysMatchingRequest(t *testing.T) {
	initMiddlewareTestDB(t)
	count := 0
	h := IdempotencyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		count++
		writeJSON(w, http.StatusCreated, map[string]any{"ok": true, "count": count})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":1}`))
	req1.Header.Set("Idempotency-Key", "same")
	w1 := httptest.NewRecorder()
	h(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":1}`))
	req2.Header.Set("Idempotency-Key", "same")
	w2 := httptest.NewRecorder()
	h(w2, req2)

	if count != 1 {
		t.Fatalf("expected handler to run once, got %d", count)
	}
	if w2.Code != http.StatusCreated {
		t.Fatalf("expected replayed status 201, got %d", w2.Code)
	}
	if replayed := w2.Header().Get("X-Idempotency-Replayed"); replayed != "true" {
		t.Fatalf("expected replay header, got %q", replayed)
	}
}

func TestIdempotencyMiddlewareRejectsConflictingPayload(t *testing.T) {
	initMiddlewareTestDB(t)
	h := IdempotencyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":1}`))
	req1.Header.Set("Idempotency-Key", "same")
	w1 := httptest.NewRecorder()
	h(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/demo", strings.NewReader(`{"x":2}`))
	req2.Header.Set("Idempotency-Key", "same")
	w2 := httptest.NewRecorder()
	h(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d", w2.Code)
	}
	if !strings.Contains(w2.Body.String(), "idempotency_key_conflict") {
		t.Fatalf("expected conflict code in body, got %s", w2.Body.String())
	}
}

func TestInvokeHandlerRejectsInvalidPipelineMode(t *testing.T) {
	initMiddlewareTestDB(t)
	h := MakeInvokeHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/invoke", strings.NewReader(`{
		"world_id":"w1",
		"node_id":"n1",
		"task_type":"custom",
		"context":{"pipeline_mode":"turbo"}
	}`))
	w := httptest.NewRecorder()
	h(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid_pipeline_mode") {
		t.Fatalf("expected invalid_pipeline_mode response, got %s", w.Body.String())
	}
}
