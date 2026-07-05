package api

import (
	"encoding/json"
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
)

// MakeTickAdvanceHandler 返回世界刻推进接口处理函数。
// 处理时会创建新的时间线刻度，并调用 world_tick 推理任务。
func MakeTickAdvanceHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		var req struct {
			TickType        string `json:"tick_type"`
			GameTime        string `json:"game_time"`
			AutonomousLimit *int   `json:"autonomous_limit,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid json")
			return
		}

		tick, resp, autonomousRuns, err := service.AdvanceWorldTickWithAutonomous(p, worldID, req.TickType, req.GameTime, req.AutonomousLimit)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{
			"tick":            tick,
			"invoke":          resp,
			"autonomous_runs": autonomousRuns,
		})
	}
}

// MakeAutonomousRunHandler 手动触发某个节点的自主行为。
func MakeAutonomousRunHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		nodeID := r.PathValue("node_id")
		resp, err := service.RunAutonomousNode(p, worldID, nodeID)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, 200, resp)
	}
}

// MakeAutonomousConfigGetHandler 返回节点自主行为配置。
func MakeAutonomousConfigGetHandler(_ *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.PathValue("node_id")
		cfg, component, err := service.GetAutonomousConfig(nodeID)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		if cfg == nil || component == nil {
			writeJSON(w, 200, map[string]any{"component": nil, "config": nil})
			return
		}
		writeJSON(w, 200, map[string]any{"component": component, "config": cfg})
	}
}

// MakeAutonomousConfigPutHandler 创建或更新节点自主行为配置。
func MakeAutonomousConfigPutHandler(_ *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.PathValue("node_id")
		var cfg engine.AutonomousConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			errorJSON(w, 400, "invalid json: "+err.Error())
			return
		}
		component, err := service.UpsertAutonomousConfig(nodeID, &cfg)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, 200, map[string]any{"component": component, "config": cfg})
	}
}

// MakeEventImpactHandler 返回事件影响评估接口处理函数。
func MakeEventImpactHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		var req struct {
			EventType   string `json:"event_type"`
			ScopeID     string `json:"scope_id"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid json")
			return
		}

		scopeID := req.ScopeID
		if scopeID == "" {
			scopeID = worldID
		}

		resp, err := service.EvaluateWorldEvent(p, worldID, &engine.WorldEvent{
			EventType:   req.EventType,
			ScopeID:     scopeID,
			Description: req.Description,
			Severity:    req.Severity,
		})
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, 200, resp)
	}
}

// MakeTimelineReplanHandler 重新生成世界未来时间线大纲。
func MakeTimelineReplanHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		resp, err := service.ReplanWorldTimeline(p, worldID)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, 200, resp)
	}
}

// MakeScopeAdvanceHandler 推进某个局部范围的世界演化。

// MakeCloneWorldHandler 返回世界复制接口处理函数。
func MakeCloneWorldHandler(_ *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		var req struct {
			Name      string `json:"name,omitempty"`
			LockWorld bool   `json:"lock_world,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid json")
			return
		}
		created, err := service.CloneWorld(worldID, req.Name, req.LockWorld)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, 201, created)
	}
}

func MakeScopeAdvanceHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		worldID := r.PathValue("world_id")
		scopeID := r.PathValue("scope_id")
		resp, err := service.AdvanceWorldScope(p, worldID, scopeID)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, 200, resp)
	}
}
