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

### Runtime and Inference Operations

| Method | Description |
|---|---|
| `Invoke(req *InvokeRequest) (*InvokeResponse, error)` | Unified inference entry point |
| `AdvanceTick(worldID, tickType, gameTime string) (*TickResponse, error)` | Advance one world tick |
| `AdvanceTickWithAutonomousLimit(worldID, tickType, gameTime string, autonomousLimit *int) (*TickResponse, error)` | Advance one world tick with an autonomous-run cap |
| `EventImpact(worldID string, event *WorldEvent) (*InvokeResponse, error)` | Evaluate event impact |
| `ScopeAdvance(worldID, scopeID string) (*InvokeResponse, error)` | Advance a specific scope |
| `TimelineReplan(worldID string) (*InvokeResponse, error)` | Rebuild a world's future outline |
| `ActionCallback(callbackID, status string, result any) error` | Complete an async action callback |

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
