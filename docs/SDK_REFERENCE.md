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

## 关系与传播语义

SDK 应与 Engine 保持同一套语义边界：

- `parent` 是唯一主层级结构，表示稳定身份/归属链。
- `located_at` 表示当前环境位置，不替代 `parent`。
- `belongs_to` / `subordinate` 表示组织归属或控制链，不替代 `parent`。
- `ally` / `enemy` / `kinship` 是社会关系，默认不会自动进入主上下文。
- `external_parent` 是额外作用域挂接，当前不参与默认上下文和默认传播。

常用 SDK 常量：

- 关系类型：`sdk.RelationBelongsTo`、`sdk.RelationLocatedAt`、`sdk.RelationSubordinate`、`sdk.RelationExternalParent` 等。
- 管线模式：`sdk.PipelineModeVertical`、`sdk.PipelineModePolling`、`sdk.PipelineModeFull`。
- 传播模式：`sdk.PropagationModeUpward`、`sdk.PropagationModeEnvironment`、`sdk.PropagationModeOrganization`、`sdk.PropagationModeTagBroadcast`、`sdk.PropagationModeTargeted`、`sdk.PropagationModeManual`。

`InvokeContext.IncludeRelatedNodes` 只是“受控关系补充”开关，不是“把所有邻接节点全部展开”的开关。

`InvokeContext.DynamicInterfaces` 用于传递请求级的动态外部能力。一个实用原则是：

- 稳定、全局存在的接口继续放在 Engine 配置里
- 只在当前回合或当前 NPC 对话中临时开放的接口，通过 `dynamic_interfaces` 传入
- 给模型提供可调用接口时，优先使用结构化字段，不要把函数定义整段手写进提示词

### 携带动态接口发起 Invoke

```go
resp, err := client.Invoke(&sdk.InvokeRequest{
    WorldID:  worldID,
    NodeID:   nodeID,
    TaskType: "npc_dialogue",
    Messages: []sdk.ChatMessage{{Role: "user", Content: "你现在看到了什么？"}},
    Context: &sdk.InvokeContext{
        PipelineMode:      sdk.PipelineModePolling,
        MaxAnalysisRounds: 4,
        DynamicInterfaces: []sdk.DynamicInterface{
            {
                ID:                "scene_facts",
                Kind:              sdk.DynamicInterfaceDataRequest,
                ExternalInterface: "game_client_request_data",
                Description:       "查询当前玩家可见的场景状态",
                QueryTypes:        []string{"node_detail", "visible_entities"},
                MaxQueries:        2,
            },
            {
                ID:                "merchant_ops",
                Kind:              sdk.DynamicInterfaceAction,
                ExternalInterface: "npc_trade_action",
                Description:       "执行与交易相关的外部动作",
                MaxCalls:          1,
            },
        },
    },
})
```

如果不想手写整段 `DynamicInterface` 结构，也可以用 SDK helper 组装：

```go
req := &sdk.InvokeRequest{
    WorldID:  worldID,
    NodeID:   nodeID,
    TaskType: "npc_dialogue",
    Messages: []sdk.ChatMessage{{Role: "user", Content: "你现在看到了什么？"}},
    Context:  sdk.NewInvokeContext(),
}

req.Context.PipelineMode = sdk.PipelineModePolling
req.Context.MaxAnalysisRounds = 4

req.AddDynamicInterfaces(
    sdk.NewDynamicDataRequest(
        "scene_facts",
        "game_client_request_data",
        sdk.WithDescription("查询当前玩家可见的场景状态"),
        sdk.WithQueryTypes("node_detail", "visible_entities"),
        sdk.WithMaxQueries(2),
    ),
    sdk.NewDynamicAction(
        "merchant_ops",
        "npc_trade_action",
        sdk.WithActionDescription("执行与交易相关的外部动作"),
        sdk.WithMaxCalls(1),
    ),
)

resp, err := client.Invoke(req)
```

推荐约定：

- 稳定存在、投递治理固定的接口继续放在 Engine 正式 `external_interfaces` 配置里
- 只在当前回合、当前 NPC 对话或当前场景临时开放的接口，使用 `AddDynamicInterfaces(...)`
- SDK helper 只负责构造请求级白名单；真正的 delivery、consumer、resume_policy 仍由 Engine 配置决定

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

### 读取连续性状态组件

```go
state, err := client.GetStateComponents(worldID)
single, err := client.GetStateComponent(worldID, "world_state")

_, err = client.PutStateComponent(worldID, "tick_policy", map[string]any{
    "continuity_rules": []string{"Do not discard established underground reactor facts."},
})
```

### 读取时间线归档

```go
timelines, err := client.GetTimelines(worldID, 10)
latest, err := client.GetLatestTimeline(worldID)
```

### 推进世界时间

```go
tick, err := client.AdvanceTick(worldID, "scheduled", "第 3 天 - 傍晚")
// tick.Tick, tick.Invoke, tick.AutonomousRuns
```

如果你需要限制本次 Tick 可触发的自主节点数量，可使用 `AdvanceTickWithAutonomousLimit`。


### 运行时任务管理

```go
// ListRuntimeTasks 按分类和状态过滤运行时任务
func (c *Client) ListRuntimeTasks(category, status string, limit int) ([]RuntimeTask, error)

// GetRuntimeTask 获取单个运行时任务详情
func (c *Client) GetRuntimeTask(taskID string) (*RuntimeTask, error)

// ClaimRuntimeTask 认领一个待处理任务，返回含 lease_token 的任务
func (c *Client) ClaimRuntimeTask(taskID, consumer, leaseOwner string) (*RuntimeTask, error)

// StartRuntimeTask 开始执行一个已认领的任务
func (c *Client) StartRuntimeTask(taskID, leaseToken string) (*RuntimeTask, error)

// HeartbeatRuntimeTask 为运行中的任务发送心跳
func (c *Client) HeartbeatRuntimeTask(taskID, leaseToken string) error

// ReleaseRuntimeTask 释放一个已认领或运行中的任务
func (c *Client) ReleaseRuntimeTask(taskID, leaseToken, reason string) error
```

运行时任务支持 Push、Pull、Hybrid 三种投递模式。Push 模式下 Engine 直接推送任务到游戏端；Pull 模式需要游戏端主动轮询认领。所有模式最终都通过 `ActionCallback` 汇报结果。

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

### 加载连续性调试包

```go
bundle, err := client.GetContinuityBundle(worldID, &sdk.ContinuityBundleOptions{
    LogLimit:   20,
    TraceLimit: 10,
    LogQuery: &sdk.InferenceLogQuery{
        TaskType:      "world_tick",
        ExecutionMode: "debug",
    },
})
// bundle.LatestTimeline, bundle.StateComponents, bundle.Logs, bundle.Traces
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
| `GetTimelines(worldID string, limit int) (*TimelinesResponse, error)` | 读取最近的 world tick 时间线归档 |
| `GetLatestTimeline(worldID string) (*LatestTimelineResponse, error)` | 读取最近一条 world tick 时间线归档 |

### 自主行为

| 方法 | 说明 |
|---|---|
| `GetAutonomousConfig(nodeID string) (*AutonomousConfigResponse, error)` | 读取自主行为配置 |
| `SetAutonomousConfig(nodeID string, cfg *AutonomousConfig) (*AutonomousConfigResponse, error)` | 创建或更新自主行为配置 |
| `RunAutonomousNode(worldID, nodeID string) (*InvokeResponse, error)` | 手动触发一次自主行为执行 |

说明：

- `ActionCallback(...)` 当前用于两类场景：普通异步动作回调，以及 `request_data.target = "game_client"` 的数据回填。
- 当某次回调对应的是 paused execution 时，Engine 会在服务端自动恢复原始多轮推理；调用方不需要再单独发一次 resume 请求。

### 设置、策略、日志与导入

| 方法 | 说明 |
|---|---|
| `GetWorldSettings(worldID string) (*WorldSettings, error)` | 读取世界设置 |
| `UpdateWorldSettings(worldID string, settings *WorldSettingsUpdate) (*WorldSettings, error)` | 部分更新世界设置 |
| `SetWorldSettings(worldID string, settings *WorldSettings) (*WorldSettings, error)` | 用完整载荷覆盖世界设置 |
| `GetWorldPolicy(worldID string) (*WorldPolicy, error)` | 读取世界策略 |
| `SetWorldPolicy(worldID string, blocked, safe []string) (*WorldPolicy, error)` | 更新世界策略 |
| `GetLogs(worldID string, limit, offset int, taskType string) ([]InferenceLog, error)` | 读取推理日志 |
| `GetLogsByQuery(query InferenceLogQuery) ([]InferenceLog, error)` | 使用结构化服务端过滤条件读取推理日志 |
| `GetDebugTraces(worldID string, limit int) (*DebugTraceList, error)` | 读取最近的调试轨迹 |
| `GetContinuityBundle(worldID string, options *ContinuityBundleOptions) (*ContinuityBundle, error)` | 一次性加载时间线、连续性状态、日志和调试轨迹 |
| `GetStateComponents(worldID string) (*StateComponentsResponse, error)` | 读取全部连续性状态组件 |
| `GetStateComponent(worldID, componentType string) (*StateComponentResponse, error)` | 读取单个连续性状态组件 |
| `PutStateComponent(worldID, componentType string, payload any) (*StateComponentResponse, error)` | 创建或更新连续性状态组件 |
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

### `StateComponentsResponse`

```go
type StateComponentsResponse struct {
    WorldID    string                   `json:"world_id"`
    Components []StateComponentEnvelope `json:"components"`
}
```

### `TimelinesResponse`

```go
type TimelinesResponse struct {
    WorldID   string             `json:"world_id"`
    Timelines []TimelineEnvelope `json:"timelines"`
}
```

### `InferenceLogQuery`

```go
type InferenceLogQuery struct {
    WorldID       string `json:"world_id,omitempty"`
    NodeID        string `json:"node_id,omitempty"`
    TaskType      string `json:"task_type,omitempty"`
    Category      string `json:"category,omitempty"`
    EventName     string `json:"event_name,omitempty"`
    ExecutionMode string `json:"execution_mode,omitempty"`
    RequestID     string `json:"request_id,omitempty"`
    Round         int    `json:"round,omitempty"`
    Limit         int    `json:"limit,omitempty"`
    Offset        int    `json:"offset,omitempty"`
}
```

### `ContinuityBundleOptions`

```go
type ContinuityBundleOptions struct {
    SkipLogs      bool               `json:"skip_logs,omitempty"`
    SkipTraces    bool               `json:"skip_traces,omitempty"`
    LogLimit      int                `json:"log_limit,omitempty"`
    TraceLimit    int                `json:"trace_limit,omitempty"`
    LogQuery      *InferenceLogQuery `json:"log_query,omitempty"`
}
```

### `ContinuityBundle`

```go
type ContinuityBundle struct {
    WorldID         string                   `json:"world_id"`
    LatestTimeline  *LatestTimelineResponse  `json:"latest_timeline,omitempty"`
    StateComponents *StateComponentsResponse `json:"state_components,omitempty"`
    Logs            []InferenceLog           `json:"logs,omitempty"`
    Traces          []DebugTrace             `json:"traces,omitempty"`
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

### `PropagationRule`

```go
type PropagationRule struct {
    Mode          string   `json:"mode,omitempty"`            // upward / environment_scope / organization_scope / tag_broadcast / targeted / manual
    TargetTags    []string `json:"target_tags,omitempty"`     // tag_broadcast 模式
    TargetNodeIDs []string `json:"target_node_ids,omitempty"` // targeted 模式
    MaxDepth      int      `json:"max_depth,omitempty"`       // 结构传播模式下表示目标节点起始后的祖先展开深度
    PublishUp     bool     `json:"publish_up,omitempty"`      // 仅影响 upward 的更高层发布行为
}
```

传播模式说明：

- `upward`：只沿主 `parent` 链传播。
- `environment_scope`：沿 `located_at` 指向的环境节点及其场景祖先传播。
- `organization_scope`：沿 `belongs_to` / `subordinate` 指向的组织或控制节点及其主 `parent` 链传播。
- `tag_broadcast`：按标签广播。
- `targeted`：定向传播。
- `manual`：不自动传播。

`node_relations` 与 `node_memories` 的数据查询过滤语义也需要与 Engine 保持一致：

- `node_relations.filter` = `relation_type`
- `node_memories.filter` = `memory_level`

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
