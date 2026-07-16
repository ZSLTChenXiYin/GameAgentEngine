# Go SDK Reference

[**中文**](./SDK_REFERENCE.md) | **English**

GameAgentEngine provides a Go SDK for calling the engine service from Go applications.

---

## Basic Usage

### Creating a Client

```go
import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

client := sdk.NewClient("http://127.0.0.1:8080", "dev-key")
```

## Relation and propagation semantics

The SDK should follow the same semantic boundaries as the Engine:

- `parent` is the only primary hierarchy and represents stable identity / ownership structure.
- `located_at` represents current environment position and does not replace `parent`.
- `belongs_to` / `subordinate` represent organization affiliation or control chain and do not replace `parent`.
- `ally` / `enemy` / `kinship` are social edges and should not enter default prompt context automatically.
- `external_parent` is an auxiliary scope attachment and is currently excluded from default context and default propagation.

Common SDK constants:

- Relation types: `sdk.RelationBelongsTo`, `sdk.RelationLocatedAt`, `sdk.RelationSubordinate`, `sdk.RelationExternalParent`, etc.
- Pipeline modes: `sdk.PipelineModeVertical`, `sdk.PipelineModePolling`, `sdk.PipelineModeFull`.
- Propagation modes: `sdk.PropagationModeUpward`, `sdk.PropagationModeEnvironment`, `sdk.PropagationModeOrganization`, `sdk.PropagationModeTagBroadcast`, `sdk.PropagationModeTargeted`, `sdk.PropagationModeManual`.

`InvokeContext.IncludeRelatedNodes` only enables bounded related-node supplements. It is not permission to expand every adjacent node into prompt context.

Use `InvokeContext.DynamicInterfaces` for request-scoped external capabilities. A good rule is:

- keep stable, global interfaces in engine config
- pass temporary or NPC-turn-specific interfaces through `dynamic_interfaces`
- prefer structured interface fields over hand-writing function specs into prompt text

Structured tool behavior:

- if the provider supports structured tools, the SDK request is forwarded as real tool definitions for built-in engine actions plus request-scoped dynamic interfaces
- if the provider does not support structured tools, the Engine keeps the same allowlist but exposes it through prompt instructions instead
- built-in engine capabilities are exposed by task type, so an NPC dialogue turn and a world tick do not automatically receive the same callable set
- dynamic action interfaces can add `ArgsSchema` for runtime argument validation before dispatch

### Invoking with Dynamic Interfaces

```go
resp, err := client.Invoke(&sdk.InvokeRequest{
    WorldID:  worldID,
    NodeID:   nodeID,
    TaskType: "npc_dialogue",
    Messages: []sdk.ChatMessage{{Role: "user", Content: "What do you see?"}},
    Context: &sdk.InvokeContext{
        PipelineMode:      sdk.PipelineModePolling,
        MaxAnalysisRounds: 4,
        DynamicInterfaces: []sdk.DynamicInterface{
            {
                ID:                "scene_facts",
                Kind:              sdk.DynamicInterfaceDataRequest,
                ExternalInterface: "game_client_request_data",
                Description:       "Query the current visible scene state",
                QueryTypes:        []string{"node_detail", "visible_entities"},
                MaxQueries:        2,
            },
            {
                ID:                "merchant_ops",
                Kind:              sdk.DynamicInterfaceAction,
                ExternalInterface: "npc_trade_action",
                Description:       "Perform trade-related external actions",
                MaxCalls:          1,
            },
        },
    },
})
```

### Health and Version

```go
if err := client.Health(); err != nil {
    panic(err)
}

engineVersion, minCompatibleVersion, err := client.GetVersion()
```

### Getting World Settings

```go
settings, err := client.GetWorldSettings(worldID)
// settings.PipelineMode, settings.PropagationMaxDepth, ...
```

### Updating World Settings

Use `UpdateWorldSettings` for partial updates so only the fields you set are changed:

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

Use `SetWorldSettings` when you want to send a complete `WorldSettings` payload in one call.

### Advancing World Time

```go
tick, err := client.AdvanceTick(worldID, "scheduled", "Day 3 - Evening")
// tick.Tick, tick.Invoke, tick.AutonomousRuns
```

Use `AdvanceTickWithAutonomousLimit` if you want to cap autonomous runs for the tick.

### Evaluating Event Impact

```go
event := &sdk.WorldEvent{
    EventType:   "diplomatic_crisis",
    ScopeID:     scopeID,
    Description: "Neighboring nation is amassing troops at the border",
    Severity:    "critical",
}

resp, err := client.EventImpact(worldID, event)
```

### Loading a Continuity Bundle

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

### Managing Runtime Tasks


```go
// ListRuntimeTasks returns runtime tasks matching the given filters
func (c *Client) ListRuntimeTasks(category, status string, limit int) ([]RuntimeTask, error)

// GetRuntimeTask returns details for a single runtime task
func (c *Client) GetRuntimeTask(taskID string) (*RuntimeTask, error)

// ClaimRuntimeTask claims a pending task, returning the task with a lease token
func (c *Client) ClaimRuntimeTask(taskID, consumer, leaseOwner string) (*RuntimeTask, error)

// StartRuntimeTask marks a claimed task as running
func (c *Client) StartRuntimeTask(taskID, leaseToken string) (*RuntimeTask, error)

// HeartbeatRuntimeTask sends a heartbeat for a running task
func (c *Client) HeartbeatRuntimeTask(taskID, leaseToken string) error

// ReleaseRuntimeTask releases a claimed or running task
func (c *Client) ReleaseRuntimeTask(taskID, leaseToken, reason string) error
```


Runtime tasks support Push, Pull, and Hybrid delivery modes. In Push mode, the Engine directly dispatches tasks to the game client. In Pull mode, the game client polls for pending tasks and claims them. All modes ultimately report results through `ActionCallback`.


---

## Client Methods

### Service and Raw Access

| Method | Description |
|---|---|
| `Health() error` | Call the health endpoint |
| `GetVersion() (string, string, error)` | Read engine version and minimum compatible version |
| `RawGet(path string) ([]byte, error)` | Perform a raw authenticated GET |
| `WithIdempotency(key string) *Client` | Clone the client with an `Idempotency-Key` header |

### Node Operations

| Method | Description |
|---|---|
| `GetNodes(worldID string, limit, offset int, nodeType string) ([]Node, error)` | List nodes with optional filters |
| `GetNode(id string) (*NodeDetail, error)` | Get node details |
| `CreateNode(worldID, name, nodeType, parentID string) (string, error)` | Create a node |
| `UpdateNode(id string, name, nodeType string, parentID *string) (*Node, error)` | Update a node |
| `DeleteNode(id string) error` | Delete a node |
| `CopyNode(nodeID, name string, parentID *string, includeDescendants bool) (*Node, error)` | Copy a node, optionally with its subtree |

### Component Operations

| Method | Description |
|---|---|
| `AddComponent(nodeID, compType, data string) (string, error)` | Add a component |
| `GetComponents(nodeID string) ([]Component, error)` | List node components |
| `GetComponent(id string) (*Component, error)` | Get a component |
| `UpdateComponent(id string, componentType, data *string) (*Component, error)` | Update a component |
| `DeleteComponent(id string) error` | Delete a component |

### Memory Operations

| Method | Description |
|---|---|
| `AddMemory(nodeID, content, level, tags string) (string, error)` | Add a memory |
| `GetMemories(nodeID string) ([]Memory, error)` | List node memories |
| `GetMemory(id string) (*Memory, error)` | Get a memory |
| `UpdateMemory(id string, content, level, tags *string) (*Memory, error)` | Update a memory |
| `DeleteMemory(id string) error` | Delete a memory |
| `PropagateMemory(memoryID, mode string, tags, targetIDs []string, maxDepth int, publishUp bool) error` | Trigger explicit memory propagation |

### Relation Operations

| Method | Description |
|---|---|
| `GetRelations(worldID string, limit, offset int, relationType string) ([]Relation, error)` | List relations with optional filters |
| `GetRelation(id string) (*Relation, error)` | Get a relation |
| `CreateRelation(worldID, sourceID, targetID, relType string, weight int) (string, error)` | Create a relation |
| `CreateRelationWithProps(worldID, sourceID, targetID, relType string, weight int, props string) (string, error)` | Create a relation with properties |
| `UpdateRelation(id string, sourceID, targetID, relationType, properties *string, weight *int) (*Relation, error)` | Update a relation |
| `DeleteRelation(id string) error` | Delete a relation |

### World Operations

| Method | Description |
|---|---|
| `GetWorlds() ([]Node, error)` | List world root nodes |
| `UpdateWorld(worldID, name string) (*Node, error)` | Rename or update mutable world fields |
| `ForkWorld(worldID, name string, lockWorld bool) (*Node, error)` | Create a runnable working copy |
| `CreateWorldSnapshot(worldID, name string, lockWorld bool) (*Node, error)` | Create a save snapshot |
| `RestoreWorld(worldID, name string, lockWorld bool) (*Node, error)` | Restore a snapshot into a fresh world |
| `ValidateWorldSnapshot(worldID string) (*SnapshotValidationResult, error)` | Validate snapshot compatibility |
| `GetWorldSnapshotMetadata(worldID string) (*WorldSnapshotInfo, error)` | Read snapshot metadata |
| `ListWorldSnapshots(worldID string) ([]WorldSnapshotInfo, error)` | List snapshots created from a source world |
| `DeleteWorldSnapshot(worldID string) error` | Delete a snapshot world and its metadata |
| `GetStateComponents(worldID string) (*StateComponentsResponse, error)` | Read the world continuity state bundle |
| `GetStateComponent(worldID, componentType string) (*StateComponentResponse, error)` | Read one continuity state component |
| `PutStateComponent(worldID, componentType string, payload any) (*StateComponentResponse, error)` | Replace one continuity state component |
| `GetTimelines(worldID string, limit int) (*TimelinesResponse, error)` | Read recent world tick archives |
| `GetLatestTimeline(worldID string) (*LatestTimelineResponse, error)` | Read the latest world tick archive |

### Runtime and Inference Operations

| Method | Description |
|---|---|
| `Invoke(req *InvokeRequest) (*InvokeResponse, error)` | Unified inference entry point |
| `AdvanceTick(worldID, tickType, gameTime string) (*TickResponse, error)` | Advance one world tick |
| `AdvanceTickWithAutonomousLimit(worldID, tickType, gameTime string, autonomousLimit *int) (*TickResponse, error)` | Advance one world tick with an autonomous-run cap |
| `AdvanceTickWithOptions(worldID, tickType, gameTime string, requestedTicks *int, autonomousLimit *int) (*TickResponse, error)` | Advance one world tick with explicit requested base ticks |
| `EventImpact(worldID string, event *WorldEvent) (*InvokeResponse, error)` | Evaluate event impact |
| `ScopeAdvance(worldID, scopeID string) (*InvokeResponse, error)` | Advance a specific scope |
| `TimelineReplan(worldID string) (*InvokeResponse, error)` | Rebuild a world's future outline |
| `ActionCallback(callbackID, status string, result any) error` | Complete an async action callback |
| `ListPendingPlans(worldID string) ([]PendingPlan, error)` | List plans waiting for manual review |
| `ApprovePlan(worldID, planID string) (*PlanDecisionResponse, error)` | Approve one pending plan |
| `RejectPlan(worldID, planID string) (*PlanDecisionResponse, error)` | Reject one pending plan |

| `ListRuntimeTasks(category, status string, limit int) ([]RuntimeTask, error)` | List runtime tasks with optional filters |
| `GetRuntimeTask(taskID string) (*RuntimeTask, error)` | Get a single runtime task |
| `ClaimRuntimeTask(taskID, consumer, leaseOwner string) (*RuntimeTask, error)` | Claim a pending task |
| `StartRuntimeTask(taskID, leaseToken string) (*RuntimeTask, error)` | Start executing a claimed task |
| `HeartbeatRuntimeTask(taskID, leaseToken string) error` | Send a heartbeat for a running task |
| `ReleaseRuntimeTask(taskID, leaseToken, reason string) error` | Release a claimed or running task |

### Autonomous Behavior

| Method | Description |
|---|---|
| `GetAutonomousConfig(nodeID string) (*AutonomousConfigResponse, error)` | Read autonomous config |
| `SetAutonomousConfig(nodeID string, cfg *AutonomousConfig) (*AutonomousConfigResponse, error)` | Create or update autonomous config |
| `RunAutonomousNode(worldID, nodeID string) (*InvokeResponse, error)` | Manually trigger one autonomous run |

### Settings, Policy, Logs, and Import

| Method | Description |
|---|---|
| `GetWorldSettings(worldID string) (*WorldSettings, error)` | Read world settings |
| `UpdateWorldSettings(worldID string, settings *WorldSettingsUpdate) (*WorldSettings, error)` | Partially update world settings |
| `SetWorldSettings(worldID string, settings *WorldSettings) (*WorldSettings, error)` | Replace world settings with a full payload |
| `GetWorldPolicy(worldID string) (*WorldPolicy, error)` | Read world policy |
| `SetWorldPolicy(worldID string, blocked, safe []string) (*WorldPolicy, error)` | Update world policy |
| `GetLogs(worldID string, limit, offset int, taskType string) ([]InferenceLog, error)` | Read inference logs |
| `GetLogsByQuery(query InferenceLogQuery) ([]InferenceLog, error)` | Read inference logs with structured server-side filters |
| `GetDebugTraces(worldID string, limit int) (*DebugTraceList, error)` | Read recent debug traces |
| `GetContinuityBundle(worldID string, options *ContinuityBundleOptions) (*ContinuityBundle, error)` | Load timelines, continuity state, logs, and traces together |
| `CreatorImport(format, content string, reset, dryRun bool) (*ImportResult, error)` | Import world configuration |

---

## Common Types

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
    AdvancedTicks  int                   `json:"advanced_ticks,omitempty"`
    WorldTimeState *WorldTimeState       `json:"world_time_state,omitempty"`
    AutonomousRuns []AutonomousRunResult `json:"autonomous_runs,omitempty"`
}
```

### `RuntimeTask`

```go
type RuntimeTask struct {
    TaskID       string `json:"task_id"`
    Category     string `json:"category,omitempty"`
    Status       string `json:"status"`
    Consumer     string `json:"consumer,omitempty"`
    WorldID      string `json:"world_id,omitempty"`
    NodeID       string `json:"node_id,omitempty"`
    RequestID    string `json:"request_id,omitempty"`
    CallbackID   string `json:"callback_id,omitempty"`
    LeaseToken   string `json:"lease_token,omitempty"`
    LeaseOwner   string `json:"lease_owner,omitempty"`
    AttemptCount int    `json:"attempt_count,omitempty"`
    MaxAttempts  int    `json:"max_attempts,omitempty"`
    Priority     int    `json:"priority,omitempty"`
    PayloadJSON  string `json:"payload_json,omitempty"`
    CreatedAt    string `json:"created_at,omitempty"`
    UpdatedAt    string `json:"updated_at,omitempty"`
}
```

### Continuity helpers

The SDK continuity bundle now includes recent timeline history in addition to `LatestTimeline`:

```go
type ContinuityBundle struct {
    WorldID         string
    LatestTimeline  *TimelineEnvelope
    Timelines       []TimelineEnvelope
    StateComponents []StateComponentEnvelope
    Logs            []InferenceLog
    Traces          []DebugTrace
}
```

Use these helpers to avoid re-parsing timeline and state payloads by hand:

- `bundle.FindStateComponent("world_time_state")`
- `bundle.LatestWorldTimeState()`
- `timeline.WorldTimeState()`
- `timeline.PreviousWorldTimeState()`
- `timeline.EffectiveAdvancedTicks()`

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
    TargetTags    []string `json:"target_tags,omitempty"`     // tag_broadcast mode
    TargetNodeIDs []string `json:"target_node_ids,omitempty"` // targeted mode
    MaxDepth      int      `json:"max_depth,omitempty"`       // for structural propagation modes, ancestor expansion depth starting from the target node
    PublishUp     bool     `json:"publish_up,omitempty"`      // only affects higher-level publication behavior for upward mode
}
```

Propagation mode meanings:

- `upward`: walk only the primary `parent` chain.
- `environment_scope`: use the `located_at` target and that environment node's scene ancestors.
- `organization_scope`: use `belongs_to` / `subordinate` targets and then each target's primary `parent` chain.
- `tag_broadcast`: explicit tag-based broadcast.
- `targeted`: explicit point-to-point propagation.
- `manual`: disable automatic propagation.

Data-query filter semantics must also stay aligned with the Engine:

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
