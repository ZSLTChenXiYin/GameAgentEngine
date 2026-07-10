package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// MakeInvokeHandler 返回统一推理入口的处理函数。
// 处理函数会解析请求、调用引擎管线并返回推理结果。
func MakeInvokeHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req engine.InvokeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid json: "+err.Error())
			return
		}
		if req.WorldID == "" || req.NodeID == "" {
			errorJSON(w, 400, "world_id and node_id required")
			return
		}
		if req.Context != nil && req.Context.PipelineMode != "" && !engine.IsValidPipelineMode(string(req.Context.PipelineMode)) {
			errorJSONCode(w, http.StatusBadRequest, "invalid_pipeline_mode", "context.pipeline_mode must be one of: vertical, polling, full")
			return
		}
		resp, err := p.Execute(&req)
		if err != nil {
			errorJSON(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	}
}

// MakeActionCallbackHandler 返回异步动作回调接口处理函数。
// 游戏侧执行完动作后，可以通过该接口上报结果。
func MakeActionCallbackHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			CallbackID string `json:"callback_id"`
			Status     string `json:"status"`
			Result     any    `json:"result,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid json: "+err.Error())
			return
		}
		if req.CallbackID == "" || req.Status == "" {
			errorJSON(w, 400, "callback_id and status required")
			return
		}
		rec, err := p.ActionRegistry().HandleCallback(req.CallbackID, req.Status, req.Result)
		if err != nil {
			errorJSON(w, 404, err.Error())
			return
		}
		if err := store.CompleteRuntimeTaskByCallbackID(req.CallbackID, req.Status, req.Result); err != nil {
			errorJSON(w, 500, err.Error())
			return
		}
		resp := map[string]any{"status": "ok"}
		if rec != nil && rec.ResumeExecutionID != "" {
			resp["resume_execution_id"] = rec.ResumeExecutionID
			if req.Status == "success" || req.Status == "completed" || req.Status == "ok" {
				resumed, err := p.ResumePausedExecution(req.CallbackID, req.Result)
				if err != nil {
					errorJSON(w, 500, err.Error())
					return
				}
				resp["resumed"] = resumed
			}
		}
		writeJSON(w, 200, resp)
	}
}

// CreateNodeHandler 创建一个新节点。
// 对于 world 类型节点，世界 ID 应与节点自身保持一致。
func CreateNodeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorldID  string `json:"world_id"`
		Name     string `json:"name"`
		NodeType string `json:"node_type"`
		ParentID string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}
	if req.Name == "" || req.NodeType == "" {
		errorJSON(w, 400, "name and node_type required")
		return
	}
	if !engine.IsValidNodeType(req.NodeType) {
		errorJSON(w, 400, "invalid node_type")
		return
	}
	if req.WorldID == "" && req.NodeType != "world" {
		errorJSON(w, 400, "world_id required for non-world nodes")
		return
	}
	var parentID *string
	if req.ParentID != "" {
		parentID = &req.ParentID
	}
	m, err := service.CreateNode(req.WorldID, req.Name, req.NodeType, parentID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 201, m)
}

// AddComponentHandler 为节点挂载一个组件。
func AddComponentHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID        string `json:"node_id"`
		ComponentType string `json:"component_type"`
		Data          string `json:"data"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}
	if req.NodeID == "" || req.ComponentType == "" {
		errorJSON(w, 400, "node_id and component_type required")
		return
	}
	if !engine.IsValidComponentType(req.ComponentType) {
		errorJSON(w, 400, "invalid component_type")
		return
	}
	m, err := service.CreateComponent(req.NodeID, req.ComponentType, req.Data)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 201, m)
}

// GetComponentsHandler 返回指定节点的组件列表。
func GetComponentsHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		errorJSON(w, 400, "node_id required")
		return
	}
	comps, err := store.GetNodeComponents(nodeID)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, comps)
}

// GetComponentHandler 返回单个组件详情。
func GetComponentHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorJSON(w, 400, "id required")
		return
	}
	component, err := store.GetComponent(id)
	if err != nil {
		errorJSON(w, 404, "not found")
		return
	}
	writeJSON(w, 200, component)
}

// CreateMemoryHandler 为节点写入一条显式记忆。
func CreateMemoryHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID  string `json:"node_id"`
		Content string `json:"content"`
		Level   string `json:"level,omitempty"`
		Tags    string `json:"tags,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}
	if req.NodeID == "" || req.Content == "" {
		errorJSON(w, 400, "node_id and content required")
		return
	}
	if req.Level == "" {
		req.Level = "long_term"
	}
	if !engine.IsValidMemoryLevel(req.Level) {
		errorJSON(w, 400, "invalid memory level")
		return
	}
	m, err := service.CreateMemory(req.NodeID, req.Content, req.Level, req.Tags)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 201, m)
}

// GetMemoriesHandler 返回节点的记忆列表。
func GetMemoriesHandler(w http.ResponseWriter, r *http.Request) {
	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		errorJSON(w, 400, "node_id required")
		return
	}
	mems, err := store.GetNodeMemories(nodeID, 100)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, mems)
}

// GetMemoryHandler 返回单条记忆详情。
func GetMemoryHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorJSON(w, 400, "id required")
		return
	}
	memory, err := store.GetMemory(id)
	if err != nil {
		errorJSON(w, 404, "not found")
		return
	}
	writeJSON(w, 200, memory)
}

// GetRelationsHandler 返回关系列表。
func GetRelationsHandler(w http.ResponseWriter, r *http.Request) {
	wid := r.URL.Query().Get("world_id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	relationType := r.URL.Query().Get("relation_type")
	rels, err := store.GetAllRelations(wid, limit, offset, relationType)
	if err != nil {
		errorJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, rels)
}

// GetRelationHandler 返回单条关系详情。
func GetRelationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errorJSON(w, 400, "id required")
		return
	}
	relation, err := store.GetRelation(id)
	if err != nil {
		errorJSON(w, 404, "not found")
		return
	}
	writeJSON(w, 200, relation)
}

// CreateRelationHandler 创建一条节点之间的有向关系。
func CreateRelationHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorldID      string `json:"world_id"`
		SourceID     string `json:"source_id"`
		TargetID     string `json:"target_id"`
		RelationType string `json:"relation_type"`
		Weight       int    `json:"weight"`
		Properties   string `json:"properties,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}
	if req.WorldID == "" || req.SourceID == "" || req.TargetID == "" || req.RelationType == "" {
		errorJSON(w, 400, "world_id, source_id, target_id and relation_type required")
		return
	}
	if !engine.IsValidRelationType(req.RelationType) {
		errorJSON(w, 400, "invalid relation_type")
		return
	}
	m, err := service.CreateRelation(req.WorldID, req.SourceID, req.TargetID, req.RelationType, float64(req.Weight), req.Properties)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 201, m)
}

// UpdateNodeHandler 更新节点名称、类型或父节点信息。
func UpdateNodeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name     *string `json:"name,omitempty"`
		NodeType *string `json:"node_type,omitempty"`
		ParentID *string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}
	if req.NodeType != nil {
		if *req.NodeType == "" || !engine.IsValidNodeType(*req.NodeType) {
			errorJSON(w, 400, "invalid node_type")
			return
		}
	}
	if req.Name == nil && req.NodeType == nil && req.ParentID == nil {
		errorJSON(w, 400, "no node updates provided")
		return
	}
	var parentID *string
	parentIDSet := req.ParentID != nil
	if req.ParentID != nil && *req.ParentID != "" {
		parentID = req.ParentID
	}
	node, err := service.UpdateNode(id, req.Name, req.NodeType, parentID, parentIDSet)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, node)
}

// CopyNodeHandler duplicates a node, optionally including its descendants.
func CopyNodeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Name               string  `json:"name,omitempty"`
		ParentID           *string `json:"parent_id,omitempty"`
		IncludeDescendants *bool   `json:"include_descendants,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}
	includeDescendants := true
	if req.IncludeDescendants != nil {
		includeDescendants = *req.IncludeDescendants
	}
	node, err := service.CopyNode(id, service.CopyNodeOptions{
		Name:               req.Name,
		ParentID:           req.ParentID,
		ParentIDSet:        req.ParentID != nil,
		IncludeDescendants: includeDescendants,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 201, node)
}

// UpdateComponentHandler 更新组件类型或数据。
func UpdateComponentHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		ComponentType *string `json:"component_type,omitempty"`
		Data          *string `json:"data,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}

	if req.ComponentType != nil {
		if *req.ComponentType == "" || !engine.IsValidComponentType(*req.ComponentType) {
			errorJSON(w, 400, "invalid component_type")
			return
		}
	}
	if req.ComponentType == nil && req.Data == nil {
		errorJSON(w, 400, "no component updates provided")
		return
	}
	component, err := service.UpdateComponent(id, req.ComponentType, req.Data)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, component)
}

// DeleteComponentHandler 删除指定组件。
func DeleteComponentHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := service.DeleteComponent(id); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// UpdateMemoryHandler 更新记忆内容、层级或标签。
func UpdateMemoryHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Content *string `json:"content,omitempty"`
		Level   *string `json:"level,omitempty"`
		Tags    *string `json:"tags,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}

	if req.Level != nil {
		if *req.Level == "" || !engine.IsValidMemoryLevel(*req.Level) {
			errorJSON(w, 400, "invalid memory level")
			return
		}
	}
	if req.Content == nil && req.Level == nil && req.Tags == nil {
		errorJSON(w, 400, "no memory updates provided")
		return
	}
	memory, err := service.UpdateMemory(id, req.Content, req.Level, req.Tags)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, memory)
}

// DeleteMemoryHandler 删除指定记忆。
func DeleteMemoryHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := service.DeleteMemory(id); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// UpdateRelationHandler 更新一条关系记录。
func UpdateRelationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		SourceID     *string `json:"source_id,omitempty"`
		TargetID     *string `json:"target_id,omitempty"`
		RelationType *string `json:"relation_type,omitempty"`
		Weight       *int    `json:"weight,omitempty"`
		Properties   *string `json:"properties,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorJSON(w, 400, "invalid json")
		return
	}

	if req.SourceID != nil {
		if *req.SourceID == "" {
			errorJSON(w, 400, "source_id cannot be empty")
			return
		}
	}
	if req.TargetID != nil {
		if *req.TargetID == "" {
			errorJSON(w, 400, "target_id cannot be empty")
			return
		}
	}
	if req.RelationType != nil {
		if *req.RelationType == "" || !engine.IsValidRelationType(*req.RelationType) {
			errorJSON(w, 400, "invalid relation_type")
			return
		}
	}
	if req.SourceID == nil && req.TargetID == nil && req.RelationType == nil && req.Weight == nil && req.Properties == nil {
		errorJSON(w, 400, "no relation updates provided")
		return
	}
	var weightF64 *float64
	if req.Weight != nil {
		v := float64(*req.Weight)
		weightF64 = &v
	}
	relation, err := service.UpdateRelation(id, req.SourceID, req.TargetID, req.RelationType, weightF64, req.Properties)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, relation)
}

// DeleteRelationHandler 删除指定关系。
func DeleteRelationHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := service.DeleteRelation(id); err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

// MakePropagateMemoryHandler 返回手动触发记忆传播的处理函数。
// 开发者可通过此 API 手动传播已有记忆到目标节点。
// MakePropagateMemoryHandler 返回手动触发记忆传播的处理函数。
func MakePropagateMemoryHandler(p *engine.Pipeline) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			MemoryID   string   `json:"memory_id"`
			TargetNode string   `json:"target_node"`
			Mode       string   `json:"mode"`
			Tags       []string `json:"tags,omitempty"`
			TargetIDs  []string `json:"target_ids,omitempty"`
			MaxDepth   int      `json:"max_depth,omitempty"`
			PublishUp  bool     `json:"publish_up,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, 400, "invalid json: "+err.Error())
			return
		}
		if req.MemoryID == "" {
			errorJSON(w, 400, "memory_id required")
			return
		}
		memory, err := store.GetMemory(req.MemoryID)
		if err != nil {
			errorJSON(w, 404, "memory not found: "+err.Error())
			return
		}
		mode := engine.PropagationMode(req.Mode)
		if mode == "" {
			mode = engine.PropModeUpward
		} else if !engine.IsValidPropagationMode(mode) {
			errorJSON(w, 400, "unsupported propagation mode")
			return
		}
		level := engine.MemoryLevel(memory.Level)
		if level == "" {
			level = engine.MemLongTerm
		}
		memUpdate := engine.MemoryUpdate{
			NodeID:  memory.NodeUUID,
			Content: memory.Content,
			Level:   level,
			Tags:    memory.Tags,
			Propagation: &engine.PropagationRule{
				Mode:          mode,
				TargetTags:    req.Tags,
				TargetNodeIDs: req.TargetIDs,
				MaxDepth:      req.MaxDepth,
				PublishUp:     req.PublishUp,
			},
		}
		targetNode := req.TargetNode
		if targetNode == "" {
			targetNode = memory.NodeUUID
		}
		memoryNode, err := store.GetNode(memory.NodeUUID)
		if err != nil {
			errorJSON(w, 404, "memory node not found: "+err.Error())
			return
		}
		invokeReq := &engine.InvokeRequest{
			WorldID:  memoryNode.WorldUUID,
			TaskType: engine.TaskCustom,
			NodeID:   memory.NodeUUID,
		}
		p.ManualPropagateMemory(invokeReq, memUpdate, targetNode)
		writeJSON(w, 200, map[string]string{"status": "propagated"})
	}
}
