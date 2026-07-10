# Inference Pipeline Internals

[**中文**](./PIPELINE_INTERNALS.md) | **English**

This document describes the current GameAgentEngine inference pipeline, including PipelineMode, multi-round execution, sub-task DAG orchestration, data-request loops, action execution, memory propagation, and the recently added database-pipeline and observability layers.

---

## Pipeline Modes

Each world can independently configure `pipeline_mode`:

| Mode | Value | Behavior |
|---|---|---|
| Vertical | `vertical` | Single-pass inference with minimal branching |
| Polling | `polling` | Multi-round inference with follow-up data requests |
| Full | `full` | Multi-round inference plus sub-task DAG orchestration |

`pipeline_mode` and `execution_mode` are independent:

- `pipeline_mode` controls how reasoning is performed
- `execution_mode` controls how outcomes are applied or reviewed

---

## Pipeline.Execute Flow

The unified execution entrypoint now performs the following stages:

1. Load world settings and world policy.
2. Build base context.
3. Dispatch by task type.
4. Enter the shared multi-round loop.
5. Execute actions, persist memories, and trigger propagation.
6. Persist structured logs and return the response.

Main task types currently include:

- `npc_dialogue`
- `world_tick`
- `world_event`
- `autonomous_act`
- `custom`

---

## Shared Multi-round Loop

The common loop drives all complex tasks:

1. Build the system prompt.
2. Call the LLM.
3. Parse the JSON response.
4. Process `request_data`.
5. Process `sub_tasks`.
6. Execute actions.
7. Persist memories and propagation side effects.
8. Log the result and decide whether another round is needed.

In `debug` and `review`, logs retain richer request / response / detail payloads. In `production`, the engine still writes structured summaries with a smaller footprint.

---

## Sub-task DAG

In `full` mode, the model can declare `sub_tasks`, which are registered as DAG nodes and scheduled according to dependency order.

Key responsibilities:

- sub-task registration
- dependency resolution
- ready-task execution
- retry and timeout handling
- result merging

Current merge modes:

- `append`
- `override`
- `summarize`

---

## Data Request Loop

The model can emit `request_data` queries inside a reasoning round:

1. The pipeline parses the query list.
2. `target="store"` queries are resolved directly against the database.
3. `target="game_client"` queries are delegated through the callback mechanism.
4. Resolved data is injected into the next round of context.

This keeps the “ask, fetch, continue reasoning” pattern inside the shared pipeline instead of scattering it across callers.

---

## Action Execution

Action execution follows a shared rule set:

1. Parse `action_calls`.
2. Validate capability and schema constraints.
3. Execute sync actions immediately.
4. Return `callback_id` for async actions.
5. Accept async completion via `POST /api/v1/actions/callback`.

For `autonomous_act`, the capability allowlist is an enforced constraint, not a soft suggestion.

---

## Memory and Propagation

`memory_updates` flow through the unified write path and may trigger propagation:

- `upward`
- `tag_broadcast`
- `targeted`
- `manual`

The optional propagation machine can add:

- rule-chain checks
- transform rules
- chained propagation actions

Recent database-pipeline work moved direct memory writes and part of propagation persistence into batched write paths, which reduces fragmented transactions on SQLite.

---

## world_tick Continuity Artifacts

`world_tick` does more than a normal inference call. It also persists continuity artifacts such as:

- `world_state`
- `story_state`
- `story_history`
- `world_time_state`
- `state_snapshot`
- timeline archive rows

These artifacts can be injected back into later prompts, creating a continuity loop across ticks.

---

## Database Pipeline Integration

The inference pipeline is now aligned with the unified database pipeline.

### SQLite

- WAL mode
- busy timeout
- single writer connection
- concurrent reads
- batched log writes
- batched memory / propagation writes
- same-world exclusion for critical heavy operations

### MySQL / PostgreSQL

- pooled concurrent writes
- shared transaction entrypoints
- shared migration runner
- shared retriable-write layer

As a result, bottleneck analysis is no longer limited to business logic. You can now directly observe retry, queue, and lock behavior.

---

## Observability

New read-only endpoint:

- `GET /api/v1/pipeline/stats`

Current observable indicators include:

- write retry attempts / retries / recoveries / failures
- transaction count and accumulated duration
- log sink queue depth, flush count, fallback writes
- world-lock acquisitions, contention count, active holders

For replay and detailed debugging, combine this with:

- `GET /api/v1/logs`
- `GET /debug/traces`

---

## Config and Fallback Controls

Static config fields most relevant to pipeline stability now include:

```yaml
database:
  driver: "sqlite"                # sqlite / mysql / postgres
  dsn: "gameagentengine.db"
  migrations_enabled: true
  write_retry_enabled: true
  write_retry_max_attempts: 3
  write_retry_base_delay_ms: 40
  write_retry_max_delay_ms: 250
  log_batch_enabled: true
  log_batch_size: 32
  log_batch_flush_ms: 750
  log_batch_queue_size: 1024

engine:
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

These flags are not meant for permanent degradation. They exist so staged rollouts, stress testing, and incident mitigation can temporarily fall back to safer behavior without reverting business-layer code.
