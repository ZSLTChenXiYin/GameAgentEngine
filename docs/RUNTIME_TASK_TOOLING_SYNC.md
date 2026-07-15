# Runtime Task Tooling Sync

This document records the tooling-side sync work that aligned `SDK`, `DevCli`, and `Creator` with the current Engine runtime-task and callback-resume model.

## Scope

The sync target was the Engine capability set around:

- request-scoped `dynamic_interfaces`
- runtime task queue inspection
- callback completion and post-process observation
- paused execution resume diagnostics

## Delivered

### SDK

- Added request-scoped dynamic interface helper builders
- `ActionCallback(...)` now returns `(*CallbackResponse, error)`
- Added runtime task helper methods:
  - `ListPendingRuntimeTasks`
  - `RequeueRuntimeTask`
  - `GetRuntimeTaskStats`
- Expanded `RuntimeTask` and `RuntimeTaskStats` typed fields

### DevCli

- Added `invoke <world-id> <node-id>`
- Added request-scoped dynamic interface input flags:
  - `--pipeline-mode`
  - `--dynamic-interfaces-json`
  - `--dynamic-interfaces-file`
- Added runtime task diagnosis commands:
  - `task inspect <task-id>`
  - `task stats`
  - `task requeue <task-id>`
- Extended `task list` and `task get`

### Creator

- Upgraded the `Tasks` page from a flat table to a diagnostic view
- Added filters for:
  - status
  - category
  - consumer
  - diagnostic view
  - current world only
- Added runtime task health and distribution cards
- Added callback / resume / route / payload / result inspection
- Added jump from runtime task request ID into `Continuity` tracked request analysis

## Agreed Responsibility Boundary

The current recommended boundary is:

1. Keep stable delivery, transport, consumer, retry, callback post-process, and resume policy in Engine configuration.
2. Pass only request-local temporary game-side interfaces through `dynamic_interfaces`.
3. Treat runtime task payload snapshots as the execution-time contract for workers and management tooling.

Two practical rules follow from that boundary:

1. Prefer structured function/tool fields over hand-writing interface definitions into prompt text.
2. Use prompt text only for behavioral instruction, not as the primary contract for callable game-side interfaces.

This means the game side does not need to stuff every possible interface into prompt text or every request. It only needs to provide temporary per-turn capabilities when they are actually relevant.

## Minimal End-to-End Loop

The currently supported minimal loop is:

1. Game side invokes one request with optional `dynamic_interfaces`
2. Engine exposes those interfaces as structured tool definitions for the current invocation
3. LLM emits a query/action that routes to a runtime task
4. Worker or game client claims / starts / executes the task
5. Worker posts callback result
6. Engine updates runtime task status, applies callback post-process, and resumes paused execution when policy requires it

## Operational Checks

When the loop does not behave as expected, check in this order:

1. `GameAgentDevCli task inspect <task-id>`
2. `GameAgentDevCli task stats`
3. `GameAgentDevCli debug continuity <world-id>`
4. Creator `Tasks` page
5. Creator `Continuity` and `Traces` pages

## Recent Source Fixes Covered By Tooling

The current source tree already includes fixes for several issues seen in real machine-style testing:

- repeated `data_request` emissions after callback resume now reuse the resolved callback payload instead of generating duplicate runtime tasks
- `GameAgentDevCli task inspect <task-id>` now reads the real runtime task envelope returned by Engine, instead of showing mostly empty fields
- the legacy root command `GameAgentDevCli nodes` now shares the same filter flags as `GameAgentDevCli node list`, so `--world` and related filters reach `/api/v1/nodes` correctly
- `world_settings` creation/update paths are hardened against duplicate-row races so runtime-task and component-adjacent flows do not fail on duplicate `world_settings` inserts

## Verification Completed In This Sync

Validated locally with:

```bash
go test ./...
```

Validated additionally with Creator JS syntax checks for the updated task diagnostics page.

## Remaining Gap

The remaining gap is not Engine-side wiring. It is real game-side integration rehearsal:

- a real worker process
- a real callback token / runtime task token setup
- a real invoke -> task -> callback -> resume run against a configured world

That gap is operational, not architectural.
