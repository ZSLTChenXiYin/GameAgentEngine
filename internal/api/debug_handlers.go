package api

import (
	"net/http"
	"strconv"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
)

// MakeDebugTracesHandler 返回读取调试轨迹的 HTTP handler。
func MakeDebugTracesHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.URL.Query().Get("world_id")
		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
				limit = n
			}
		}

		var traces []*engine.Trace
		if worldID != "" {
			traces = engine.GlobalTraceRing.FilterByWorld(worldID, limit)
		} else {
			traces = engine.GlobalTraceRing.List(limit)
		}

		writeJSON(w, 200, map[string]any{
			"traces": traces,
			"count":  len(traces),
		})
	}
}
