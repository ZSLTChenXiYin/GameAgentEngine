package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// Client 是 GameAgentEngine HTTP API 的轻量封装。
type Client struct {
	baseURL        string
	apiKey         string
	hc             *http.Client
	idempotencyKey string
}

// NewClient 创建一个 SDK 客户端实例。
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		hc:      &http.Client{},
	}
}

// WithIdempotency 返回一个携带指定幂等 key 的客户端副本。
// 调用时每个请求都会附带 Idempotency-Key 请求头。
func (c *Client) WithIdempotency(key string) *Client {
	clone := *c
	clone.idempotencyKey = key
	return &clone
}

// do 负责发送 HTTP 请求并统一处理错误响应。
func (c *Client) do(method, path string, body any) ([]byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, r)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", c.idempotencyKey)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("api %s %s: %d %s", method, path, resp.StatusCode, string(data))
	}
	return data, nil
}

// buildQuery 构造 URL 查询字符串，忽略空值和零值。
func buildQuery(params map[string]any) string {
	vals := url.Values{}
	for k, v := range params {
		switch val := v.(type) {
		case string:
			if val != "" {
				vals.Set(k, val)
			}
		case int:
			if val > 0 {
				vals.Set(k, strconv.Itoa(val))
			}
		}
	}
	return vals.Encode()
}

// Health 调用健康检查接口确认服务可达。
func (c *Client) Health() error {
	_, err := c.do("GET", "/health", nil)
	return err
}

// GetNodes 获取节点列表，支持分页和按类型过滤。
func (c *Client) GetNodes(worldID string, limit, offset int, nodeType string) ([]Node, error) {
	p := "/api/v1/nodes"
	query := buildQuery(map[string]any{
		"world_id":  worldID,
		"limit":     limit,
		"offset":    offset,
		"node_type": nodeType,
	})
	if query != "" {
		p += "?" + query
	}
	data, err := c.do("GET", p, nil)
	if err != nil {
		return nil, err
	}
	var nodes []Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}

// GetNode 获取单个节点详情。
func (c *Client) GetNode(id string) (*NodeDetail, error) {
	data, err := c.do("GET", "/api/v1/nodes/"+id, nil)
	if err != nil {
		return nil, err
	}
	var nd NodeDetail
	if err := json.Unmarshal(data, &nd); err != nil {
		return nil, err
	}
	return &nd, nil
}

// CreateNode 创建一个节点并返回其 ID。
func (c *Client) CreateNode(worldID, name, nodeType, parentID string) (string, error) {
	body := map[string]string{
		"world_id":  worldID,
		"name":      name,
		"node_type": nodeType,
	}
	if parentID != "" {
		body["parent_id"] = parentID
	}
	data, err := c.do("POST", "/api/v1/nodes", body)
	if err != nil {
		return "", err
	}
	var n Node
	if err := json.Unmarshal(data, &n); err != nil {
		return "", err
	}
	return n.ID, nil
}

// UpdateNode 更新节点名称、类型或父节点。
func (c *Client) UpdateNode(id string, name, nodeType string, parentID *string) (*Node, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if nodeType != "" {
		body["node_type"] = nodeType
	}
	if parentID != nil {
		body["parent_id"] = *parentID
	}
	data, err := c.do("PUT", "/api/v1/nodes/"+id, body)
	if err != nil {
		return nil, err
	}
	var node Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// DeleteNode 删除一个节点。
func (c *Client) DeleteNode(id string) error {
	_, err := c.do("DELETE", "/api/v1/nodes/"+id, nil)
	return err
}

// AddComponent 为节点创建一个组件并返回其 ID。
func (c *Client) AddComponent(nodeID, compType, data string) (string, error) {
	body := map[string]string{
		"node_id":        nodeID,
		"component_type": compType,
		"data":           data,
	}
	resp, err := c.do("POST", "/api/v1/components", body)
	if err != nil {
		return "", err
	}
	var comp Component
	if err := json.Unmarshal(resp, &comp); err != nil {
		return "", err
	}
	return comp.ID, nil
}

// GetComponents 获取节点的组件列表。
func (c *Client) GetComponents(nodeID string) ([]Component, error) {
	data, err := c.do("GET", "/api/v1/components?node_id="+nodeID, nil)
	if err != nil {
		return nil, err
	}
	var comps []Component
	if err := json.Unmarshal(data, &comps); err != nil {
		return nil, err
	}
	return comps, nil
}

// GetComponent 获取单个组件详情。
func (c *Client) GetComponent(id string) (*Component, error) {
	data, err := c.do("GET", "/api/v1/components/"+id, nil)
	if err != nil {
		return nil, err
	}
	var comp Component
	if err := json.Unmarshal(data, &comp); err != nil {
		return nil, err
	}
	return &comp, nil
}

// UpdateComponent 更新组件类型或数据。
func (c *Client) UpdateComponent(id string, componentType, data *string) (*Component, error) {
	body := map[string]any{}
	if componentType != nil {
		body["component_type"] = *componentType
	}
	if data != nil {
		body["data"] = *data
	}
	resp, err := c.do("PUT", "/api/v1/components/"+id, body)
	if err != nil {
		return nil, err
	}
	var comp Component
	if err := json.Unmarshal(resp, &comp); err != nil {
		return nil, err
	}
	return &comp, nil
}

// DeleteComponent 删除指定组件。
func (c *Client) DeleteComponent(id string) error {
	_, err := c.do("DELETE", "/api/v1/components/"+id, nil)
	return err
}

// AddMemory 为节点创建一条记忆并返回其 ID。
func (c *Client) AddMemory(nodeID, content, level, tags string) (string, error) {
	body := map[string]string{
		"node_id": nodeID,
		"content": content,
		"level":   level,
		"tags":    tags,
	}
	resp, err := c.do("POST", "/api/v1/memories", body)
	if err != nil {
		return "", err
	}
	var mem Memory
	if err := json.Unmarshal(resp, &mem); err != nil {
		return "", err
	}
	return mem.ID, nil
}

// GetMemories 获取节点的记忆列表。
func (c *Client) GetMemories(nodeID string) ([]Memory, error) {
	data, err := c.do("GET", "/api/v1/memories?node_id="+nodeID, nil)
	if err != nil {
		return nil, err
	}
	var memories []Memory
	if err := json.Unmarshal(data, &memories); err != nil {
		return nil, err
	}
	return memories, nil
}

// GetMemory 获取单条记忆详情。
func (c *Client) GetMemory(id string) (*Memory, error) {
	data, err := c.do("GET", "/api/v1/memories/"+id, nil)
	if err != nil {
		return nil, err
	}
	var memory Memory
	if err := json.Unmarshal(data, &memory); err != nil {
		return nil, err
	}
	return &memory, nil
}

// UpdateMemory 更新记忆内容、层级或标签。
func (c *Client) UpdateMemory(id string, content, level, tags *string) (*Memory, error) {
	body := map[string]any{}
	if content != nil {
		body["content"] = *content
	}
	if level != nil {
		body["level"] = *level
	}
	if tags != nil {
		body["tags"] = *tags
	}
	resp, err := c.do("PUT", "/api/v1/memories/"+id, body)
	if err != nil {
		return nil, err
	}
	var memory Memory
	if err := json.Unmarshal(resp, &memory); err != nil {
		return nil, err
	}
	return &memory, nil
}

// DeleteMemory 删除指定记忆。
func (c *Client) DeleteMemory(id string) error {
	_, err := c.do("DELETE", "/api/v1/memories/"+id, nil)
	return err
}

// GetRelations 获取关系列表，支持分页和按类型过滤。
func (c *Client) GetRelations(worldID string, limit, offset int, relationType string) ([]Relation, error) {
	p := "/api/v1/relations"
	query := buildQuery(map[string]any{
		"world_id":      worldID,
		"limit":         limit,
		"offset":        offset,
		"relation_type": relationType,
	})
	if query != "" {
		p += "?" + query
	}
	data, err := c.do("GET", p, nil)
	if err != nil {
		return nil, err
	}
	var rels []Relation
	if err := json.Unmarshal(data, &rels); err != nil {
		return nil, err
	}
	return rels, nil
}

// GetRelation 获取单条关系详情。
func (c *Client) GetRelation(id string) (*Relation, error) {
	data, err := c.do("GET", "/api/v1/relations/"+id, nil)
	if err != nil {
		return nil, err
	}
	var rel Relation
	if err := json.Unmarshal(data, &rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

// CreateRelation 创建一条关系并返回其 ID。
func (c *Client) CreateRelation(worldID, sourceID, targetID, relType string, weight int) (string, error) {
	return c.CreateRelationWithProps(worldID, sourceID, targetID, relType, weight, "")
}

// CreateRelationWithProps 创建一条带属性的关系并返回其 ID。
func (c *Client) CreateRelationWithProps(worldID, sourceID, targetID, relType string, weight int, props string) (string, error) {
	body := map[string]any{
		"world_id":      worldID,
		"source_id":     sourceID,
		"target_id":     targetID,
		"relation_type": relType,
		"weight":        weight,
		"properties":    props,
	}
	resp, err := c.do("POST", "/api/v1/relations", body)
	if err != nil {
		return "", err
	}
	var r Relation
	if err := json.Unmarshal(resp, &r); err != nil {
		return "", err
	}
	return r.ID, nil
}

// UpdateRelation 更新关系字段。
func (c *Client) UpdateRelation(id string, sourceID, targetID, relationType, properties *string, weight *int) (*Relation, error) {
	body := map[string]any{}
	if sourceID != nil {
		body["source_id"] = *sourceID
	}
	if targetID != nil {
		body["target_id"] = *targetID
	}
	if relationType != nil {
		body["relation_type"] = *relationType
	}
	if weight != nil {
		body["weight"] = *weight
	}
	if properties != nil {
		body["properties"] = *properties
	}
	resp, err := c.do("PUT", "/api/v1/relations/"+id, body)
	if err != nil {
		return nil, err
	}
	var relation Relation
	if err := json.Unmarshal(resp, &relation); err != nil {
		return nil, err
	}
	return &relation, nil
}

// DeleteRelation 删除指定关系。
func (c *Client) DeleteRelation(id string) error {
	_, err := c.do("DELETE", "/api/v1/relations/"+id, nil)
	return err
}

// ForkWorld creates a working-copy fork of a world and all its data.
func (c *Client) ForkWorld(worldID, name string, lockWorld bool) (*Node, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if lockWorld {
		body["lock_world"] = true
	}
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/fork", body)
	if err != nil {
		return nil, err
	}
	var result Node
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateWorldSnapshot creates a save-oriented snapshot copy of a world.
func (c *Client) CreateWorldSnapshot(worldID, name string, lockWorld bool) (*Node, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if lockWorld {
		body["lock_world"] = true
	}
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/snapshots", body)
	if err != nil {
		return nil, err
	}
	var result Node
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// RestoreWorld restores a saved snapshot into a new runnable world copy.
func (c *Client) RestoreWorld(worldID, name string, lockWorld bool) (*Node, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	if lockWorld {
		body["lock_world"] = true
	}
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/restore", body)
	if err != nil {
		return nil, err
	}
	var result Node
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWorlds 获取所有世界节点。
// ValidateWorldSnapshot validates whether a saved snapshot can be safely restored.
func (c *Client) ValidateWorldSnapshot(worldID string) (*SnapshotValidationResult, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/snapshot-validation", nil)
	if err != nil {
		return nil, err
	}
	var result SnapshotValidationResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWorldSnapshotMetadata returns snapshot metadata for a copied world.
func (c *Client) GetWorldSnapshotMetadata(worldID string) (*WorldSnapshotInfo, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/snapshot-metadata", nil)
	if err != nil {
		return nil, err
	}
	var result WorldSnapshotInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListWorldSnapshots returns save snapshots created from a source world.
func (c *Client) ListWorldSnapshots(worldID string) ([]WorldSnapshotInfo, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/snapshots", nil)
	if err != nil {
		return nil, err
	}
	var result []WorldSnapshotInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteWorldSnapshot deletes a saved snapshot world and its metadata.
func (c *Client) DeleteWorldSnapshot(worldID string) error {
	_, err := c.do("DELETE", "/api/v1/worlds/"+worldID+"/snapshot", nil)
	return err
}

func (c *Client) GetWorlds() ([]Node, error) {
	data, err := c.do("GET", "/api/v1/worlds", nil)
	if err != nil {
		return nil, err
	}
	var worlds []Node
	if err := json.Unmarshal(data, &worlds); err != nil {
		return nil, err
	}
	return worlds, nil
}

// UpdateWorld updates mutable world root fields such as the world name.
func (c *Client) UpdateWorld(worldID, name string) (*Node, error) {
	body := map[string]any{}
	if name != "" {
		body["name"] = name
	}
	data, err := c.do("PUT", "/api/v1/worlds/"+worldID, body)
	if err != nil {
		return nil, err
	}
	var world Node
	if err := json.Unmarshal(data, &world); err != nil {
		return nil, err
	}
	return &world, nil
}

// CopyNode duplicates a node in-place, optionally including descendants.
func (c *Client) CopyNode(nodeID, name string, parentID *string, includeDescendants bool) (*Node, error) {
	body := map[string]any{
		"include_descendants": includeDescendants,
	}
	if name != "" {
		body["name"] = name
	}
	if parentID != nil {
		body["parent_id"] = *parentID
	}
	data, err := c.do("POST", "/api/v1/nodes/"+nodeID+"/copy", body)
	if err != nil {
		return nil, err
	}
	var node Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// Invoke 调用统一推理入口。
func (c *Client) Invoke(req *InvokeRequest) (*InvokeResponse, error) {
	data, err := c.do("POST", "/api/v1/invoke", req)
	if err != nil {
		return nil, err
	}
	var resp InvokeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TickRequest 描述世界刻推进请求体。
type TickRequest struct {
	TickType        string `json:"tick_type"`
	GameTime        string `json:"game_time"`
	AutonomousLimit *int   `json:"autonomous_limit,omitempty"`
}

// TickResponse 描述世界刻推进响应体。
type TickResponse struct {
	Tick           *Timeline             `json:"tick"`
	Invoke         *InvokeResponse       `json:"invoke"`
	AutonomousRuns []AutonomousRunResult `json:"autonomous_runs,omitempty"`
}

// Timeline 表示时间线刻度的 SDK 结构。
type Timeline struct {
	ID         string `json:"id"`
	WorldID    string `json:"world_id"`
	TickNumber int    `json:"tick_number"`
	TickType   string `json:"tick_type"`
	GameTime   string `json:"game_time"`
	CreatedAt  string `json:"created_at"`
}

// AdvanceTick 推进一次世界时间线。
func (c *Client) AdvanceTick(worldID, tickType, gameTime string) (*TickResponse, error) {
	return c.AdvanceTickWithAutonomousLimit(worldID, tickType, gameTime, nil)
}

// AdvanceTickWithAutonomousLimit 推进世界时间线，并可限制本次 Tick 触发的自主节点数量。
func (c *Client) AdvanceTickWithAutonomousLimit(worldID, tickType, gameTime string, autonomousLimit *int) (*TickResponse, error) {
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/ticks/advance", TickRequest{
		TickType:        tickType,
		GameTime:        gameTime,
		AutonomousLimit: autonomousLimit,
	})
	if err != nil {
		return nil, err
	}
	var tr TickResponse
	if err := json.Unmarshal(data, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// GetAutonomousConfig 获取节点自主行为配置。
func (c *Client) GetAutonomousConfig(nodeID string) (*AutonomousConfigResponse, error) {
	data, err := c.do("GET", "/api/v1/nodes/"+nodeID+"/autonomous", nil)
	if err != nil {
		return nil, err
	}
	var resp AutonomousConfigResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// SetAutonomousConfig 创建或更新节点自主行为配置。
func (c *Client) SetAutonomousConfig(nodeID string, cfg *AutonomousConfig) (*AutonomousConfigResponse, error) {
	data, err := c.do("PUT", "/api/v1/nodes/"+nodeID+"/autonomous", cfg)
	if err != nil {
		return nil, err
	}
	var resp AutonomousConfigResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RunAutonomousNode 手动触发某个节点的自主行为周期。
func (c *Client) RunAutonomousNode(worldID, nodeID string) (*InvokeResponse, error) {
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/nodes/"+nodeID+"/autonomous/run", nil)
	if err != nil {
		return nil, err
	}
	var resp InvokeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// EventImpact 评估某个事件对世界的影响。
func (c *Client) EventImpact(worldID string, event *WorldEvent) (*InvokeResponse, error) {
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/events/impact", map[string]any{
		"event_type":  event.EventType,
		"scope_id":    event.ScopeID,
		"description": event.Description,
		"severity":    event.Severity,
	})
	if err != nil {
		return nil, err
	}
	var resp InvokeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ScopeAdvance 推进某个局部范围的世界演化。
func (c *Client) ScopeAdvance(worldID, scopeID string) (*InvokeResponse, error) {
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/scopes/"+scopeID+"/advance", nil)
	if err != nil {
		return nil, err
	}
	var resp InvokeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// TimelineReplan 重新生成一个世界的未来时间线大纲。
func (c *Client) TimelineReplan(worldID string) (*InvokeResponse, error) {
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/timeline/replan", nil)
	if err != nil {
		return nil, err
	}
	var resp InvokeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPendingPlans 列出等待人工审核的世界变更计划。
func (c *Client) ListPendingPlans(worldID string) ([]PendingPlan, error) {
	p := "/api/v1/plans/pending"
	query := buildQuery(map[string]any{"world_id": worldID})
	if query != "" {
		p += "?" + query
	}
	data, err := c.do("GET", p, nil)
	if err != nil {
		return nil, err
	}
	var plans []PendingPlan
	if err := json.Unmarshal(data, &plans); err != nil {
		return nil, err
	}
	return plans, nil
}

// ApprovePlan 批准一条待审批计划。
func (c *Client) ApprovePlan(worldID, planID string) (*PlanDecisionResponse, error) {
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/plan/approve", map[string]any{"plan_id": planID})
	if err != nil {
		return nil, err
	}
	var resp PlanDecisionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RejectPlan 拒绝一条待审批计划。
func (c *Client) RejectPlan(worldID, planID string) (*PlanDecisionResponse, error) {
	data, err := c.do("POST", "/api/v1/worlds/"+worldID+"/plan/reject", map[string]any{"plan_id": planID})
	if err != nil {
		return nil, err
	}
	var resp PlanDecisionResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetLogs 获取推理日志，支持分页和按类型过滤。
func (c *Client) GetLogs(worldID string, limit, offset int, taskType string) ([]InferenceLog, error) {
	return c.GetLogsByQuery(InferenceLogQuery{WorldID: worldID, Limit: limit, Offset: offset, TaskType: taskType})
}

// GetLogsByQuery 获取推理日志，支持服务端结构化筛选。
func (c *Client) GetLogsByQuery(query InferenceLogQuery) ([]InferenceLog, error) {
	p := "/api/v1/logs"
	queryString := buildQuery(map[string]any{
		"world_id":       query.WorldID,
		"node_id":        query.NodeID,
		"task_type":      query.TaskType,
		"category":       query.Category,
		"event_name":     query.EventName,
		"execution_mode": query.ExecutionMode,
		"request_id":     query.RequestID,
		"round":          query.Round,
		"limit":          query.Limit,
		"offset":         query.Offset,
	})
	if queryString != "" {
		p += "?" + queryString
	}
	data, err := c.do("GET", p, nil)
	if err != nil {
		return nil, err
	}
	var logs []InferenceLog
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetDebugTraces 读取最近的调试轨迹。
func (c *Client) GetDebugTraces(worldID string, limit int) (*DebugTraceList, error) {
	p := "/debug/traces"
	query := buildQuery(map[string]any{
		"world_id": worldID,
		"limit":    limit,
	})
	if query != "" {
		p += "?" + query
	}
	data, err := c.do("GET", p, nil)
	if err != nil {
		return nil, err
	}
	var payload DebugTraceList
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

// CreatorImport 调用 creator/import 接口执行导入或纯校验。
func (c *Client) CreatorImport(format, content string, reset, dryRun bool) (*ImportResult, error) {
	data, err := c.do("POST", "/api/v1/creator/import", map[string]any{
		"format":  format,
		"content": content,
		"reset":   reset,
		"dry_run": dryRun,
	})
	if err != nil {
		return nil, err
	}
	var result ImportResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWorldPolicy 获取世界的动作策略。
func (c *Client) GetWorldPolicy(worldID string) (*WorldPolicy, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/policy", nil)
	if err != nil {
		return nil, err
	}
	var policy WorldPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// SetWorldPolicy 更新世界的动作策略。
func (c *Client) SetWorldPolicy(worldID string, blocked, safe []string) (*WorldPolicy, error) {
	data, err := c.do("PUT", "/api/v1/worlds/"+worldID+"/policy", map[string]any{
		"blocked_actions": blocked,
		"safe_actions":    safe,
	})
	if err != nil {
		return nil, err
	}
	var policy WorldPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, err
	}
	return &policy, nil
}

// GetWorldSettings 获取世界的运行设置。
func (c *Client) GetWorldSettings(worldID string) (*WorldSettings, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/settings", nil)
	if err != nil {
		return nil, err
	}
	var settings WorldSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// GetStateComponents returns all engine-recognized continuity state components for a world.
func (c *Client) GetStateComponents(worldID string) (*StateComponentsResponse, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/state-components", nil)
	if err != nil {
		return nil, err
	}
	var result StateComponentsResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetStateComponent returns one continuity state component for a world.
func (c *Client) GetStateComponent(worldID, componentType string) (*StateComponentResponse, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/state-components/"+componentType, nil)
	if err != nil {
		return nil, err
	}
	var result StateComponentResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PutStateComponent creates or updates one continuity state component for a world.
func (c *Client) PutStateComponent(worldID, componentType string, payload any) (*StateComponentResponse, error) {
	data, err := c.do("PUT", "/api/v1/worlds/"+worldID+"/state-components/"+componentType, payload)
	if err != nil {
		return nil, err
	}
	var result StateComponentResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTimelines returns recent world tick archive entries.
func (c *Client) GetTimelines(worldID string, limit int) (*TimelinesResponse, error) {
	p := "/api/v1/worlds/" + worldID + "/timelines"
	query := buildQuery(map[string]any{"limit": limit})
	if query != "" {
		p += "?" + query
	}
	data, err := c.do("GET", p, nil)
	if err != nil {
		return nil, err
	}
	var result TimelinesResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLatestTimeline returns the latest world tick archive entry.
func (c *Client) GetLatestTimeline(worldID string) (*LatestTimelineResponse, error) {
	data, err := c.do("GET", "/api/v1/worlds/"+worldID+"/timelines/latest", nil)
	if err != nil {
		return nil, err
	}
	var result LatestTimelineResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetWorldSettings 更新世界的运行设置。
func buildWorldSettingsUpdateBody(settings *WorldSettingsUpdate) map[string]any {
	if settings == nil {
		return map[string]any{}
	}
	body := map[string]any{}
	if settings.MemoryLimit != nil {
		body["memory_limit"] = *settings.MemoryLimit
	}
	if settings.MaxAnalysisRounds != nil {
		body["max_analysis_rounds"] = *settings.MaxAnalysisRounds
	}
	if settings.MaxContextDepth != nil {
		body["max_context_depth"] = *settings.MaxContextDepth
	}
	if settings.AutoApply != nil {
		body["auto_apply"] = *settings.AutoApply
	}
	if settings.RequireReviewAbove != nil {
		body["require_review_above"] = *settings.RequireReviewAbove
	}
	if settings.PropagationMaxDepth != nil {
		body["propagation_max_depth"] = *settings.PropagationMaxDepth
	}
	if settings.EnablePropagationMachine != nil {
		body["enable_propagation_machine"] = *settings.EnablePropagationMachine
	}
	if settings.SubTaskMaxRetries != nil {
		body["sub_task_max_retries"] = *settings.SubTaskMaxRetries
	}
	if settings.SubTaskTimeoutSecs != nil {
		body["sub_task_timeout_secs"] = *settings.SubTaskTimeoutSecs
	}
	if settings.PipelineMode != nil {
		body["pipeline_mode"] = *settings.PipelineMode
	}
	return body
}

func (c *Client) UpdateWorldSettings(worldID string, settings *WorldSettingsUpdate) (*WorldSettings, error) {
	data, err := c.do("PUT", "/api/v1/worlds/"+worldID+"/settings", buildWorldSettingsUpdateBody(settings))
	if err != nil {
		return nil, err
	}
	var result WorldSettings
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) SetWorldSettings(worldID string, settings *WorldSettings) (*WorldSettings, error) {
	if settings == nil {
		return c.UpdateWorldSettings(worldID, nil)
	}
	return c.UpdateWorldSettings(worldID, &WorldSettingsUpdate{
		MemoryLimit:              &settings.MemoryLimit,
		MaxAnalysisRounds:        &settings.MaxAnalysisRounds,
		MaxContextDepth:          &settings.MaxContextDepth,
		AutoApply:                &settings.AutoApply,
		RequireReviewAbove:       &settings.RequireReviewAbove,
		PropagationMaxDepth:      &settings.PropagationMaxDepth,
		EnablePropagationMachine: &settings.EnablePropagationMachine,
		SubTaskMaxRetries:        &settings.SubTaskMaxRetries,
		SubTaskTimeoutSecs:       &settings.SubTaskTimeoutSecs,
		PipelineMode:             &settings.PipelineMode,
	})
}

// ActionCallback 上报异步动作执行结果。
func (c *Client) ActionCallback(callbackID, status string, result any) error {
	_, err := c.do("POST", "/api/v1/actions/callback", map[string]any{
		"callback_id": callbackID,
		"status":      status,
		"result":      result,
	})
	return err
}

// PropagateMemory 手动触发已有记忆的传播。
func (c *Client) PropagateMemory(memoryID, mode string, tags, targetIDs []string, maxDepth int, publishUp bool) error {
	body := map[string]any{
		"memory_id": memoryID,
		"mode":      mode,
	}
	if len(tags) > 0 {
		body["tags"] = tags
	}
	if len(targetIDs) > 0 {
		body["target_ids"] = targetIDs
	}
	if maxDepth > 0 {
		body["max_depth"] = maxDepth
	}
	if publishUp {
		body["publish_up"] = publishUp
	}
	_, err := c.do("POST", "/api/v1/memories/propagate", body)
	return err
}

// GetVersion 请求引擎返回版本信息。
func (c *Client) GetVersion() (string, string, error) {
	data, err := c.do("GET", "/api/v1/version", nil)
	if err != nil {
		return "", "", err
	}
	var resp struct {
		Version       string `json:"version"`
		MinCompatible string `json:"min_compatible"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", "", err
	}
	return resp.Version, resp.MinCompatible, nil
}

// RawGet 发送 GET 请求并以原始字节返回响应体（用于调试端点）。
func (c *Client) RawGet(path string) ([]byte, error) {
	return c.do("GET", path, nil)
}
