# Go SDK Reference

[**中文**](./SDK_REFERENCE.md) | **English**

GameAgentEngine v0.2.0 provides a Go SDK for communicating with the engine service from Go applications.

---

## Basic Usage

### Creating a Client

```go
import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

client := sdk.NewClient("http://127.0.0.1:8080", "dev-key")
```

### Getting World Settings

```go
settings, err := client.GetWorldSettings(worldID)
// settings contains pipeline_mode, propagation_max_depth, and other dynamic config
```

### Setting World Configuration

Use `UpdateWorldSettings` for partial updates so only the fields you set are changed:

```go
pipelineMode := "polling"
propagationDepth := 0
enableMachine := false

result, err := client.UpdateWorldSettings(worldID, &sdk.WorldSettingsUpdate{
    PipelineMode:             &pipelineMode,
    PropagationMaxDepth:      &propagationDepth,
    EnablePropagationMachine: &enableMachine,
})
```

Use `SetWorldSettings` when you want to submit a complete `WorldSettings` payload in one call.

### Advancing World Time

```go
result, err := client.AdvanceWorldTick(worldID, "scheduled", "Day 3 - Evening", nil)
// result.tick, result.invoke, result.autonomous_runs
```

### Evaluating Event Impact

```go
event := &sdk.WorldEvent{
    EventType:   "diplomatic_crisis",
    ScopeID:     scopeID,
    Description: "Neighboring nation is amassing troops at the border",
    Severity:    "critical",
}
resp, err := client.EvaluateWorldEvent(worldID, event)
```

---

## Client Struct

```go
type Client struct {
    ServerURL string
    APIKey    string
}
```

### Node Operations

| Method | Description |
|---|---|
| `CreateNode(worldID, name, nodeType, parentID string) (string, error)` | Create a node |
| `GetNode(id string) (*NodeDetail, error)` | Get node details |
| `UpdateNode(id string, name, nodeType, parentID *string, parentIDSet bool) (*NodeDetail, error)` | Update a node |
| `DeleteNode(id string) error` | Delete a node |
| `ListNodeByWorld(worldID string, limit, offset int) ([]NodeModel, error)` | List world nodes |
| `ListNodeAll(limit, offset int) ([]NodeModel, error)` | List all nodes |

### Component Operations

| Method | Description |
|---|---|
| `AddComponent(nodeID, componentType, data string) (string, error)` | Add a component |
| `GetComponent(id string) (*ComponentModel, error)` | Get a component |
| `GetComponents(nodeID string) ([]ComponentModel, error)` | Get node components |
| `UpdateComponent(id string, componentType, data *string) error` | Update a component |
| `DeleteComponent(id string) error` | Delete a component |

### Memory Operations

| Method | Description |
|---|---|
| `AddMemory(nodeID, content, level, tags string) (*MemoryModel, error)` | Add a memory |
| `GetMemory(id string) (*MemoryModel, error)` | Get a memory |
| `GetMemories(nodeID string) ([]MemoryModel, error)` | Get node memories |
| `UpdateMemory(id string, content, level, tags *string) (*MemoryModel, error)` | Update a memory |
| `DeleteMemory(id string) error` | Delete a memory |
| `PropagateMemory(memoryID, targetNode, mode string, tags []string, targetIDs []string, maxDepth int, publishUp bool) error` | Manually propagate a memory |

### Relation Operations

| Method | Description |
|---|---|
| `CreateRelation(worldID, sourceID, targetID, relationType string) (string, error)` | Create a relation |
| `CreateRelationWithProps(worldID, sourceID, targetID, relationType string, weight int, properties string) (string, error)` | Create a relation (with weight and properties) |
| `GetRelation(id string) (*RelationModel, error)` | Get a relation |
| `GetRelations(worldID string) ([]RelationModel, error)` | List world relations |
| `UpdateRelation(id string, sourceID, targetID, relationType, properties *string, weight *int) error` | Update a relation |
| `DeleteRelation(id string) error` | Delete a relation |

### World Operations

| Method | Description |
|---|---|
| `AdvanceWorldTick(worldID, tickType, gameTime string, autonomousLimit *int) (*TickAdvanceResult, error)` | Advance world time |
| `EvaluateWorldEvent(worldID string, event *WorldEvent) (*InvokeResponse, error)` | Evaluate event impact |
| `ReplanWorldTimeline(worldID string) (*InvokeResponse, error)` | Regenerate timeline |
| `AdvanceWorldScope(worldID, scopeID string) (*InvokeResponse, error)` | Advance a specific scope |
| `ForkWorld(worldID, name string, lockWorld bool) (*Node, error)` | Create a working-copy fork of a world (`lockWorld`: lock the source world during copying) |
| `CreateWorldSnapshot(worldID, name string, lockWorld bool) (*Node, error)` | Create a save snapshot of a world (`lockWorld`: lock the source world during snapshotting) |
| `RestoreWorld(worldID, name string, lockWorld bool) (*Node, error)` | Restore a saved snapshot into a new world (`lockWorld`: lock the snapshot source world during restore) |
| `ValidateWorldSnapshot(worldID string) (*SnapshotValidationResult, error)` | Validate whether a saved snapshot can still be safely restored and return a structured compatibility report |
| `GetWorldSnapshotMetadata(worldID string) (*WorldSnapshotInfo, error)` | Retrieve snapshot metadata for a copied world |
| `ListWorldSnapshots(worldID string) ([]WorldSnapshotInfo, error)` | List all save snapshots created from a source world |
| `DeleteWorldSnapshot(worldID string) error` | Delete a saved snapshot world and its persisted snapshot metadata |
| `GetWorldSettings(worldID string) (*WorldSettings, error)` | Get world settings |
| `UpdateWorldSettings(worldID string, settings *WorldSettingsUpdate) (*WorldSettings, error)` | Partially update world settings; only provided fields are changed |
| `SetWorldSettings(worldID string, settings *WorldSettings) (*WorldSettings, error)` | Submit a complete world settings payload in one call |
| `GetWorldPolicy(worldID string) (*WorldPolicy, error)` | Get world policy |
| `SetWorldPolicy(worldID string, blocked, safe []string) (*WorldPolicy, error)` | Set world policy |

### Autonomous Behavior Operations

| Method | Description |
|---|---|
| `GetAutonomousConfig(nodeID string) (*AutonomousConfig, error)` | Get autonomous behavior config |
| `SetAutonomousConfig(nodeID string, cfg *AutonomousConfig) (*AutonomousConfig, error)` | Set autonomous behavior config |
| `RunAutonomousNode(worldID, nodeID string) (*InvokeResponse, error)` | Manually trigger autonomous behavior |

### Logs & Import

| Method | Description |
|---|---|
| `GetInferenceLogs(worldID string, limit int) ([]InferenceLogModel, error)` | Read inference logs |
| `CreatorImport(format, content string, reset, dryRun bool) (*ImportResult, error)` | Import world configuration |
| `GetStatus() (*StatusResult, error)` | Get service status |

### Inference Operations

| Method | Description |
|---|---|
| `Invoke(req *InvokeRequest) (*InvokeResponse, error)` | Unified inference entry point |
| `ActionCallback(callbackID, status string, result any) error` | Async action callback |

---

## Common Type Definitions

### WorldSettings

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

### WorldPolicy

```go
type WorldPolicy struct {
    WorldID        string   `json:"world_id"`
    BlockedActions []string `json:"blocked_actions"`
    SafeActions    []string `json:"safe_actions"`
}
```
