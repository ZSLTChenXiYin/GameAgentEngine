// Package api 实现 GameAgentEngine 的 HTTP 接口层。
// 这里负责请求解码、响应编码以及与引擎和存储层对接。
package api

import (
	"net/http"
	"strconv"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// Health 返回基础健康检查结果。
func Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// GetAllNodesHandler 返回节点列表。
// 当查询参数中包含 world_id 时，会按世界过滤。
func GetAllNodesHandler(w http.ResponseWriter, r *http.Request) {
	wid := r.URL.Query().Get("world_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	nodeType := r.URL.Query().Get("node_type")
	nodes, err := store.GetAllNodes(wid, limit, offset, nodeType)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, nodes)
}

// GetNodeHandler 返回单个节点的完整详情。
// 详情包含组件、记忆、子节点和关系等附属数据。
func GetNodeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorJSON(w, 400, "id required")
		return
	}
	node, err := store.GetNode(id)
	if err != nil {
		errorJSON(w, 404, "not found")
		return
	}
	comps, _ := store.GetNodeComponents(id)
	mems, _ := store.GetNodeMemories(id, 20)
	children, _ := store.GetChildNodes(id)
	rels, _ := store.GetNodeRelations(id)

	writeJSON(w, 200, map[string]any{
		"node":       node,
		"components": comps,
		"memories":   mems,
		"children":   children,
		"relations":  rels,
	})
}

// DeleteNodeHandler 软删除指定节点。
func DeleteNodeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := service.DeleteNode(id); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// GetWorldsHandler 返回所有 world 类型节点。
func GetWorldsHandler(w http.ResponseWriter, r *http.Request) {
	worlds, err := store.GetWorlds()
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, worlds)
}

// GetLogsHandler 返回最近的推理日志列表。
// 支持通过 world_id 和 limit 做筛选。
func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	wid := r.URL.Query().Get("world_id")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	taskType := r.URL.Query().Get("task_type")
	logs, err := store.GetInferenceLogs(wid, limit, offset, taskType)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, logs)
}
