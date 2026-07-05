# Go SDK 参考

**中文** | [**English**](./SDK_REFERENCE_EN.md)

GameAgentEngine v0.2.0 提供 Go SDK，用于在 Go 应用中与引擎服务通信。

---

## 基本用法

### 创建客户端

```go
import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

client := sdk.NewClient("http://127.0.0.1:8080", "dev-key")
```

### 获取世界设置

```go
settings, err := client.GetWorldSettings(worldID)
// settings 包含 pipeline_mode、propagation_max_depth 等动态配置
```

### 设置世界配置

```go
settings := &sdk.WorldSettings{
    PipelineMode:             "full",
    PropagationMaxDepth:      3,
    SubTaskMaxRetries:        3,
    SubTaskTimeoutSecs:       120,
    EnablePropagationMachine: true,
}
result, err := client.SetWorldSettings(worldID, settings)
```

### 推进世界时间

```go
result, err := client.AdvanceWorldTick(worldID, "scheduled", "第3天-傍晚", nil)
// result.tick, result.invoke, result.autonomous_runs
```

### 评估事件影响

```go
event := &sdk.WorldEvent{
    EventType:   "diplomatic_crisis",
    ScopeID:     scopeID,
    Description: "邻国在边境集结军队",
    Severity:    "critical",
}
resp, err := client.EvaluateWorldEvent(worldID, event)
```

---

## Client 结构体

```go
type Client struct {
    ServerURL string
    APIKey    string
}
```

### 节点操作

| 方法 | 说明 |
|---|---|
| `CreateNode(worldID, name, nodeType, parentID string) (string, error)` | 创建节点 |
| `GetNode(id string) (*NodeDetail, error)` | 获取节点详情 |
| `UpdateNode(id string, name, nodeType, parentID *string, parentIDSet bool) (*NodeDetail, error)` | 更新节点 |
| `DeleteNode(id string) error` | 删除节点 |
| `ListNodeByWorld(worldID string, limit, offset int) ([]NodeModel, error)` | 列出世界节点 |
| `ListNodeAll(limit, offset int) ([]NodeModel, error)` | 列出所有节点 |

### 组件操作

| 方法 | 说明 |
|---|---|
| `AddComponent(nodeID, componentType, data string) (string, error)` | 添加组件 |
| `GetComponent(id string) (*ComponentModel, error)` | 获取组件 |
| `GetComponents(nodeID string) ([]ComponentModel, error)` | 获取节点组件列表 |
| `UpdateComponent(id string, componentType, data *string) error` | 更新组件 |
| `DeleteComponent(id string) error` | 删除组件 |

### 记忆操作

| 方法 | 说明 |
|---|---|
| `AddMemory(nodeID, content, level, tags string) (*MemoryModel, error)` | 添加记忆 |
| `GetMemory(id string) (*MemoryModel, error)` | 获取记忆 |
| `GetMemories(nodeID string) ([]MemoryModel, error)` | 获取节点记忆 |
| `UpdateMemory(id string, content, level, tags *string) (*MemoryModel, error)` | 更新记忆 |
| `DeleteMemory(id string) error` | 删除记忆 |
| `PropagateMemory(memoryID, targetNode, mode string, tags []string, targetIDs []string, maxDepth int, publishUp bool) error` | 手动传播记忆 |

### 关系操作

| 方法 | 说明 |
|---|---|
| `CreateRelation(worldID, sourceID, targetID, relationType string) (string, error)` | 创建关系 |
| `CreateRelationWithProps(worldID, sourceID, targetID, relationType string, weight int, properties string) (string, error)` | 创建关系（含权重和属性） |
| `GetRelation(id string) (*RelationModel, error)` | 获取关系 |
| `GetRelations(worldID string) ([]RelationModel, error)` | 列出世界关系 |
| `UpdateRelation(id string, sourceID, targetID, relationType, properties *string, weight *int) error` | 更新关系 |
| `DeleteRelation(id string) error` | 删除关系 |

### 世界操作

| 方法 | 说明 |
|---|---|
| `AdvanceWorldTick(worldID, tickType, gameTime string, autonomousLimit *int) (*TickAdvanceResult, error)` | 推进世界时间 |
| `EvaluateWorldEvent(worldID string, event *WorldEvent) (*InvokeResponse, error)` | 评估事件影响 |
| `ReplanWorldTimeline(worldID string) (*InvokeResponse, error)` | 重新生成时间线 |
| `AdvanceWorldScope(worldID, scopeID string) (*InvokeResponse, error)` | 局部范围推进 |
| `CloneWorld(worldID, name string, lockWorld bool) (*Node, error)` | 复制世界（`lockWorld`: 复制期间锁定源世界） |
| `GetWorldSettings(worldID string) (*WorldSettings, error)` | 获取世界设置 |
| `SetWorldSettings(worldID string, settings *WorldSettings) (*WorldSettings, error)` | 设置世界设置 |
| `GetWorldPolicy(worldID string) (*WorldPolicy, error)` | 获取世界策略 |
| `SetWorldPolicy(worldID string, blocked, safe []string) (*WorldPolicy, error)` | 设置世界策略 |

### 自主行为操作

| 方法 | 说明 |
|---|---|
| `GetAutonomousConfig(nodeID string) (*AutonomousConfig, error)` | 获取自主行为配置 |
| `SetAutonomousConfig(nodeID string, cfg *AutonomousConfig) (*AutonomousConfig, error)` | 设置自主行为配置 |
| `RunAutonomousNode(worldID, nodeID string) (*InvokeResponse, error)` | 手动触发自主行为 |

### 日志与导入

| 方法 | 说明 |
|---|---|
| `GetInferenceLogs(worldID string, limit int) ([]InferenceLogModel, error)` | 读取推理日志 |
| `CreatorImport(format, content string, reset, dryRun bool) (*ImportResult, error)` | 导入世界配置 |
| `GetStatus() (*StatusResult, error)` | 获取服务状态 |

### 推理操作

| 方法 | 说明 |
|---|---|
| `Invoke(req *InvokeRequest) (*InvokeResponse, error)` | 统一推理入口 |
| `ActionCallback(callbackID, status string, result any) error` | 异步动作回调 |

---

## 常用类型定义

### WorldSettings

```go
type WorldSettings struct {
    PipelineMode              string `json:"pipeline_mode,omitempty"`
    MemoryLimit               int    `json:"memory_limit,omitempty"`
    MaxAnalysisRounds         int    `json:"max_analysis_rounds,omitempty"`
    MaxContextDepth           int    `json:"max_context_depth,omitempty"`
    AutoApply                 bool   `json:"auto_apply,omitempty"`
    RequireReviewAbove        string `json:"require_review_above,omitempty"`
    PropagationMaxDepth       int    `json:"propagation_max_depth"`
    EnablePropagationMachine  bool   `json:"enable_propagation_machine"`
    SubTaskMaxRetries         int    `json:"sub_task_max_retries"`
    SubTaskTimeoutSecs        int    `json:"sub_task_timeout_secs"`
}
```

### InvokeResponse

```go
type InvokeResponse struct {
    RequestID       string              `json:"request_id"`
    TaskType        string              `json:"task_type"`
    Reply           string              `json:"reply"`
    ActionCalls     []ActionCall        `json:"action_calls,omitempty"`
    MemoryUpdates   []MemoryUpdate      `json:"memory_updates,omitempty"`
    WorldChangePlan *WorldChangePlan    `json:"world_change_plan,omitempty"`
    FutureOutline   string              `json:"future_outline,omitempty"`
    Metadata        *ResponseMeta       `json:"metadata,omitempty"`
}
```

### WorldPolicy

```go
type WorldPolicy struct {
    WorldID        string   `json:"world_id"`
    BlockedActions []string `json:"blocked_actions"`
    SafeActions    []string `json:"safe_actions"`
}
```