# Go SDK 参考

**中文** | [**English**](./SDK_REFERENCE_EN.md)

GameAgentEngine 提供 Go SDK，用于在 Go 应用中调用引擎服务。

---

## 基本用法

### 创建客户端

```go
import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

client := sdk.NewClient("http://127.0.0.1:8080", "dev-key")
```

### 健康检查与版本

```go
if err := client.Health(); err != nil {
    panic(err)
}

engineVersion, minCompatibleVersion, err := client.GetVersion()
```

### 获取世界设置

```go
settings, err := client.GetWorldSettings(worldID)
// settings.PipelineMode, settings.PropagationMaxDepth, ...
```

### 更新世界设置

使用 `UpdateWorldSettings` 做部分更新，只会修改你显式传入的字段：

```go
pipelineMode := "polling"
propagationDepth := 0
enableMachine := false

settings, err := client.UpdateWorldSettings(worldID, &sdk.WorldSettingsUpdate{
    PipelineMode:             &pipelineMode,
    PropagationMaxDepth:      &propagationDepth,
    EnablePropagationMachine: &enableMachine,
})
```

如果你希望一次提交完整 `WorldSettings` 载荷，则使用 `SetWorldSettings`。

### 推进世界时间

```go
tick, err := client.AdvanceTick(worldID, "scheduled", "第 3 天 - 傍晚")
// tick.Tick, tick.Invoke, tick.AutonomousRuns
```

如果你需要限制本次 Tick 可触发的自主节点数量，可使用 `AdvanceTickWithAutonomousLimit`。

### 评估事件影响

```go
event := &sdk.WorldEvent{
    EventType:   "diplomatic_crisis",
    ScopeID:     scopeID,
    Description: "邻国正在边境集结军队",
    Severity:    "critical",
}

resp, err := client.EventImpact(worldID, event)
```

---

## Client 方法

### 服务与原始访问

| 方法 | 说明 |
|---|---|
| `Health() error` | 调用健康检查接口 |
| `GetVersion() (string, string, error)` | 读取当前引擎版本和最低兼容版本 |
| `RawGet(path string) ([]byte, error)` | 发起原始的已鉴权 GET 请求 |
| `WithIdempotency(key string) *Client` | 复制一个附带 `Idempotency-Key` 的客户端 |

### 节点操作

| 方法 | 说明 |
|---|---|
| `GetNodes(worldID string, limit, offset int, nodeType string) ([]Node, error)` | 按条件列出节点 |
| `GetNode(id string) (*NodeDetail, error)` | 获取节点详情 |
| `CreateNode(worldID, name, nodeType, parentID string) (string, error)` | 创建节点 |
| `UpdateNode(id string, name, nodeType string, parentID *string) (*Node, error)` | 更新节点 |
| `DeleteNode(id string) error` | 删除节点 |
| `CopyNode(nodeID, name string, parentID *string, includeDescendants bool) (*Node, error)` | 复制节点，可选连同整棵子树 |

### 组件操作

| 方法 | 说明 |
|---|---|
| `AddComponent(nodeID, compType, data string) (string, error)` | 添加组件 |
| `GetComponents(nodeID string) ([]Component, error)` | 列出节点组件 |
| `GetComponent(id string) (*Component, error)` | 获取单个组件 |
| `UpdateComponent(id string, componentType, data *string) (*Component, error)` | 更新组件 |
| `DeleteComponent(id string) error` | 删除组件 |

### 记忆操作

| 方法 | 说明 |
|---|---|
| `AddMemory(nodeID, content, level, tags string) (string, error)` | 添加记忆 |
| `GetMemories(nodeID string) ([]Memory, error)` | 列出节点记忆 |
| `GetMemory(id string) (*Memory, error)` | 获取单条记忆 |
| `UpdateMemory(id string, content, level, tags *string) (*Memory, error)` | 更新记忆 |
| `DeleteMemory(id string) error` | 删除记忆 |
| `PropagateMemory(memoryID, mode string, tags, targetIDs []string, maxDepth int, publishUp bool) error` | 显式触发记忆传播 |

### 关系操作

| 方法 | 说明 |
|---|---|
| `GetRelations(worldID string, limit, offset int, relationType string) ([]Relation, error)` | 按条件列出关系 |
| `GetRelation(id string) (*Relation, error)` | 获取单条关系 |
| `CreateRelation(worldID, sourceID, targetID, relType string, weight int) (string, error)` | 创建关系 |
| `CreateRelationWithProps(worldID, sourceID, targetID, relType string, weight int, props string) (string, error)` | 创建带属性的关系 |
| `UpdateRelation(id string, sourceID, targetID, relationType, properties *string, weight *int) (*Relation, error)` | 更新关系 |
| `DeleteRelation(id string) error` | 删除关系 |

### 世界操作

| 方法 | 说明 |
|---|---|
| `GetWorlds() ([]Node, error)` | 列出世界根节点 |
| `UpdateWorld(worldID, name string) (*Node, error)` | 重命名世界或更新世界可变字段 |
| `ForkWorld(worldID, name string, lockWorld bool) (*Node, error)` | 创建可运行的工作副本 |
| `CreateWorldSnapshot(worldID, name string, lockWorld bool) (*Node, error)` | 创建存档快照 |
| `RestoreWorld(worldID, name string, lockWorld bool) (*Node, error)` | 将快照恢复为新世界 |
| `ValidateWorldSnapshot(worldID string) (*SnapshotValidationResult, error)` | 校验快照兼容性 |
| `GetWorldSnapshotMetadata(worldID string) (*WorldSnapshotInfo, error)` | 读取快照元数据 |
| `ListWorldSnapshots(worldID string) ([]WorldSnapshotInfo, error)` | 列出某个源世界的全部快照 |
| `DeleteWorldSnapshot(worldID string) error` | 删除快照世界及其元数据 |

### 运行时与推理操作

| 方法 | 说明 |
|---|---|
| `Invoke(req *InvokeRequest) (*InvokeResponse, error)` | 统一推理入口 |
| `AdvanceTick(worldID, tickType, gameTime string) (*TickResponse, error)` | 推进一个世界 Tick |
| `AdvanceTickWithAutonomousLimit(worldID, tickType, gameTime string, autonomousLimit *int) (*TickResponse, error)` | 推进一个世界 Tick，并限制自主行为数量 |
| `EventImpact(worldID string, event *WorldEvent) (*InvokeResponse, error)` | 评估事件影响 |
| `ScopeAdvance(worldID, scopeID string) (*InvokeResponse, error)` | 推进指定局部范围 |
| `TimelineReplan(worldID string) (*InvokeResponse, error)` | 重建世界未来大纲 |
| `ActionCallback(callbackID, status string, result any) error` | 完成异步动作回调 |
| `ListPendingPlans(worldID string) ([]PendingPlan, error)` | 列出待审批计划 |
| `ApprovePlan(worldID, planID string) (*PlanDecisionResponse, error)` | 批准一条待审批计划 |
| `RejectPlan(worldID, planID string) (*PlanDecisionResponse, error)` | 拒绝一条待审批计划 |

### 自主行为

| 方法 | 说明 |
|---|---|
| `GetAutonomousConfig(nodeID string) (*AutonomousConfigResponse, error)` | 读取自主行为配置 |
| `SetAutonomousConfig(nodeID string, cfg *AutonomousConfig) (*AutonomousConfigResponse, error)` | 创建或更新自主行为配置 |
| `RunAutonomousNode(worldID, nodeID string) (*InvokeResponse, error)` | 手动触发一次自主行为执行 |

### 设置、策略、日志与导入

| 方法 | 说明 |
|---|---|
| `GetWorldSettings(worldID string) (*WorldSettings, error)` | 读取世界设置 |
| `UpdateWorldSettings(worldID string, settings *WorldSettingsUpdate) (*WorldSettings, error)` | 部分更新世界设置 |
| `SetWorldSettings(worldID string, settings *WorldSettings) (*WorldSettings, error)` | 用完整载荷覆盖世界设置 |
| `GetWorldPolicy(worldID string) (*WorldPolicy, error)` | 读取世界策略 |
| `SetWorldPolicy(worldID string, blocked, safe []string) (*WorldPolicy, error)` | 更新世界策略 |
| `GetLogs(worldID string, limit, offset int, taskType string) ([]InferenceLog, error)` | 读取推理日志 |
| `GetDebugTraces(worldID string, limit int) (*DebugTraceList, error)` | 读取最近的调试轨迹 |
| `CreatorImport(format, content string, reset, dryRun bool) (*ImportResult, error)` | 导入世界配置 |

---

## 常用类型

### `WorldSettings`

```go
type WorldSettings struct {
    WorldID                  string `json:"world_id"`
    MemoryLimit              int    `json:"memory_limit"`
    MaxAnalysisRounds        int    `json:"max_analysis_rounds"`
    MaxContextDepth          int    `json:"max_context_depth"`
    AutoApply                bool   `json:"auto_apply"`
    RequireReviewAbove       string `json:"require_review_above"`
    PropagationMaxDepth      int    `json:"propagation_max_depth"`
    EnablePropagationMachine bool   `json:"enable_propagation_machine"`
    SubTaskMaxRetries        int    `json:"sub_task_max_retries"`
    SubTaskTimeoutSecs       int    `json:"sub_task_timeout_secs"`
    PipelineMode             string `json:"pipeline_mode"`
}
```

### `WorldSettingsUpdate`

```go
type WorldSettingsUpdate struct {
    MemoryLimit              *int    `json:"memory_limit,omitempty"`
    MaxAnalysisRounds        *int    `json:"max_analysis_rounds,omitempty"`
    MaxContextDepth          *int    `json:"max_context_depth,omitempty"`
    AutoApply                *bool   `json:"auto_apply,omitempty"`
    RequireReviewAbove       *string `json:"require_review_above,omitempty"`
    PropagationMaxDepth      *int    `json:"propagation_max_depth,omitempty"`
    EnablePropagationMachine *bool   `json:"enable_propagation_machine,omitempty"`
    SubTaskMaxRetries        *int    `json:"sub_task_max_retries,omitempty"`
    SubTaskTimeoutSecs       *int    `json:"sub_task_timeout_secs,omitempty"`
    PipelineMode             *string `json:"pipeline_mode,omitempty"`
}
```

### `TickResponse`

```go
type TickResponse struct {
    Tick           *Timeline             `json:"tick"`
    Invoke         *InvokeResponse       `json:"invoke"`
    AutonomousRuns []AutonomousRunResult `json:"autonomous_runs,omitempty"`
}
```

### `InvokeResponse`

```go
type InvokeResponse struct {
    RequestID       string               `json:"request_id"`
    TaskType        string               `json:"task_type"`
    ExecutionMode   string               `json:"execution_mode"`
    Reply           string               `json:"reply,omitempty"`
    ActionCalls     []ActionCall         `json:"action_calls,omitempty"`
    WorldChangePlan *WorldChangePlan     `json:"world_change_plan,omitempty"`
    MemoryUpdates   []MemoryUpdate       `json:"memory_updates,omitempty"`
    SubTasks        []SubTaskDeclaration `json:"sub_tasks,omitempty"`
    Metadata        *ResponseMeta        `json:"metadata,omitempty"`
}
```

### `ResponseMeta`

```go
type ResponseMeta struct {
    LLMModel               string `json:"llm_model"`
    TokensUsed             int    `json:"tokens_used"`
    ProcessingTimeMs       int64  `json:"processing_time_ms"`
    ConfiguredPipelineMode string `json:"configured_pipeline_mode,omitempty"`
    EffectivePipelineMode  string `json:"effective_pipeline_mode,omitempty"`
    MaxAnalysisRounds      int    `json:"max_analysis_rounds,omitempty"`
    RoundsUsed             int    `json:"rounds_used,omitempty"`
}
```
