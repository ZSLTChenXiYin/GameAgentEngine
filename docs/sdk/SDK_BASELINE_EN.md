# SDK Baseline Specification

This document defines the unified minimum delivery standard for the multi-language SDK set.

## 1. Goal

Every external-language SDK should support at least the current core development loop of the project:

1. connect to Engine
2. create / query worlds and nodes
3. send `invoke`
4. consume runtime tasks
5. callback task results
6. inspect state, timelines, logs, and debug information

## 2. Unified Directory Layout

Each language SDK directory should contain at least:

```text
<lang>-sdk/
├── src/ or equivalent source directory
├── examples/
├── workerhome/       # when the language ecosystem supports it, shared Worker data can live here
└── package/build metadata
```

## 3. Capability Tiers

### S1: Core Consistency Tier (required for all languages)

- HTTP client / auth / error handling
- health / version
- worlds / nodes / components / memories / relations
- invoke
- runtime task list / pending / claim / start / heartbeat / release / requeue / stats
- callback
- logs / traces
- state components / timelines

### S2: Tooling Enhancement Tier (expected for mainstream languages)

- dynamic interfaces
- continuity bundle
- world policy / world settings
- snapshot / restore / validate / fork
- player interaction typed models

### S3: Ecosystem Enhancement Tier (added according to language value)

- package-management publishing metadata
- integration examples for Unity / Godot / Node.js / native plugins
- richer typed enums / constants / helper builders

## 4. Minimum Object Model

Each SDK should provide at least these core objects or equivalent structures:

- `Node`
- `Component`
- `Relation`
- `Memory`
- `InvokeRequest`
- `InvokeResponse`
- `DynamicInterface`
- `RuntimeTask`
- `RuntimeTaskStats`
- `CallbackResponse`
- `WorldSettings`
- `StateComponent`
- `Timeline`
- `InferenceLog`
- `DebugTrace`

## 5. Minimum Method Surface

### Basic Connectivity

- `health`
- `getVersion`

### Worlds and Nodes

- `getWorlds`
- `createNode`
- `getNodes`
- `getNode`
- `updateNode`
- `deleteNode`

### Components / Memories / Relations

- `addComponent` / `getComponents` / `updateComponent` / `deleteComponent`
- `addMemory` / `getMemories` / `updateMemory` / `deleteMemory`
- `createRelation` / `getRelations` / `updateRelation` / `deleteRelation`

### Inference

- `invoke`
- `interpretPlayerInput` (mainstream languages first)

### World Runtime

- `advanceTick`
- `getWorldSettings`
- `setWorldSettings`
- `getWorldPolicy`
- `setWorldPolicy`

### Debugging and Observability

- `getLogs`
- `getDebugTraces`
- `getStateComponents`
- `getStateComponent`
- `putStateComponent`
- `getTimelines`
- `getLatestTimeline`
- `getContinuityBundle` (mainstream languages first)

### External Tasks

- `listRuntimeTasks`
- `listPendingRuntimeTasks`
- `getRuntimeTask`
- `claimRuntimeTask`
- `startRuntimeTask`
- `heartbeatRuntimeTask`
- `releaseRuntimeTask`
- `requeueRuntimeTask`
- `getRuntimeTaskStats`
- `actionCallback`

### Snapshots and Duplication

- `forkWorld`
- `createWorldSnapshot`
- `restoreWorld`
- `validateWorldSnapshot`

## 6. Minimum Example Set

Each SDK should provide at least these `examples/`:

1. `health`: connectivity check
2. `world_bootstrap`: create a world and base nodes
3. `invoke_dialogue`: start one NPC dialogue
4. `task_pull_once`: pull and process one runtime task
5. `callback_complete`: report one callback result
6. `continuity_inspect`: read state / timeline / logs

## 7. Integration Requirements With Worker

Mainstream-language SDKs should also provide minimal integration examples for:

- pull / callback flow alongside `GameAgentWorker serve`
- invoke / authority-query flow related to `GameAgentWorker play`

For native-side or script-side request-builder SDKs that do not ship an HTTP transport layer yet, these examples should at least provide the full request construction sequence and the recommended Worker pairing, so outside integrators do not have to guess the API order.

## 8. Error Model Requirements

Every SDK should consistently handle at least:

- non-2xx HTTP errors
- structured API error payloads
- JSON deserialization failures
- network errors / timeouts

## 9. Acceptance Requirements

A language SDK should only be treated as meeting baseline after all of the following are true:

1. documentation can guide integration independently
2. examples cover the minimal closed loop
3. the core capability matrix covers at least S1
4. terminology is aligned with the current Engine / Worker workflow

## 10. Relationship to the Go SDK

The Go SDK is the semantic baseline for current capability and naming:

- `sdk/client.go`
- `sdk/types.go`

Other language SDKs may follow each language's idioms in naming, but should not silently change capability boundaries or object semantics.
