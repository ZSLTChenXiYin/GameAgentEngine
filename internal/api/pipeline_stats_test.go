package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

func TestGetPipelineStatsHandlerReturnsStructuredStats(t *testing.T) {
	if err := store.Init("sqlite", fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())); err != nil {
		t.Fatalf("init db: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/pipeline/stats", nil)
	w := httptest.NewRecorder()

	GetPipelineStatsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Store struct {
			Driver       string         `json:"driver"`
			WriteRetry   map[string]any `json:"write_retry"`
			Transactions map[string]any `json:"transactions"`
			LogSink      map[string]any `json:"log_sink"`
		} `json:"store"`
		WorldLocks map[string]any `json:"world_locks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Store.Driver != "sqlite" {
		t.Fatalf("expected sqlite driver, got %q", body.Store.Driver)
	}
	if len(body.Store.WriteRetry) == 0 || len(body.Store.Transactions) == 0 || len(body.Store.LogSink) == 0 {
		t.Fatalf("expected structured store stats, got %#v", body.Store)
	}
	if len(body.WorldLocks) == 0 {
		t.Fatalf("expected world lock stats, got %#v", body.WorldLocks)
	}
}
