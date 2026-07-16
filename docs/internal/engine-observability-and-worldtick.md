# Engine Observability And World Tick Refactor

## Scope

This document records the recent refactor that addressed two production issues:

1. Pipeline execution was too black-box for debugging and review.
2. `world_tick` could advance plot beats but failed to consistently carry narrative continuity forward.

The work was delivered in 8 incremental phases on branch `clean-main`.

## Observability Model

### Execution Modes

- `debug`
  Full console logs and full DB logs.
- `review`
  Summary console logs and full DB logs.
- `production`
  Summary console logs and summary DB logs.

### Log Storage

The legacy `inference_logs` persistence model was upgraded to a unified `logs` table through `InferenceLogModel`.

Added fields:

- `category`
- `event_name`
- `log_level`
- `message`
- `request_id`
- `execution_mode`
- `configured_pipeline_mode`
- `effective_pipeline_mode`
- `round`
- `detail_data`

Migration behavior:

- `store.Init()` now migrates old `inference_logs` rows into `logs` when needed.
- Legacy rows are tagged as `category=pipeline`, `event_name=legacy_inference`.

### Logged Pipeline Stages

The engine now records structured events for:

- pipeline request start
- context build success/failure
- prompt preparation
- raw LLM response receipt
- parse results and parse failures
- interim memory updates
- data request emission/resolution
- round completion
- sub-task declaration
- review-mode pending plan creation
- action execution outcomes
- memory write outcomes
- memory propagation outcomes
- world tick service lifecycle
- autonomous scan/run lifecycle

## World Tick Persistence Split

`world_tick` persistence is now intentionally split across three targets.

### 1. `logs`

Purpose:

- runtime observability
- full replay/debug of the pipeline
- service-layer lifecycle tracing

### 2. `timelines`

Purpose:

- ordered world history by tick
- compact summary and canonical per-tick archive

Current `timelines.data` payload stores:

- `reply`
- `world_change_plan`
- `future_outline`
- `memory_updates`
- `action_calls`

### 3. State Components

Purpose:

- inheritable continuity state for the next tick
- structured world/story carry-over

New component types:

- `world_state`
- `story_state`
- `story_history`
- `tick_policy`
- `state_snapshot`

Current usage:

- `world_state`
  Stores summary, key facts, canonical facts, active arcs, metadata.
- `story_state`
  Stores current situation, recent changes, pending threads.
- `story_history`
  Stores recent tick summaries and extracted facts.
- `state_snapshot`
  Stores the latest structured world tick snapshot.
- `tick_policy`
  Currently available for continuity constraints and prompt injection.

## Continuity Injection

`world_tick` no longer depends only on `future_outline`.

Prompt input now explicitly includes:

- context graph and memories
- persistent continuity blocks from state components
- recent timeline summaries
- a continuity guard that tells the model not to reset or casually discard established facts

This is the main mechanism used to prevent one-off story facts from disappearing on the next tick.

## Fact Retention

To reduce drift, a lightweight fact extraction layer now promotes high-value facts into `world_state.canonical_facts` and `story_history.entries[].facts`.

The current extractor is rule-based and biased toward story-critical tokens such as:

- named facilities
- underground structures
- resonance / quantum artifacts
- conflict escalation facts
- key NPC names
- stability metrics and threat trends

This is intentionally conservative and easy to revise.

## Refactor Report

### Table / Schema Changes

- `InferenceLogModel.TableName()` changed from `inference_logs` to `logs`.
- `logs` gained structured observability fields listed above.
- legacy `inference_logs` rows are migrated forward on startup.
- no destructive table drop is performed automatically.

### Persistence Behavior Changes

- `CreateInferenceLog()` now resolves `world_id` / `node_id` from UUIDs before insert.
- `world_tick` now persists structured state components in addition to timeline rows.
- `timelines.data` is now actively used as a tick archive payload.

### Component Surface Changes

Added engine-recognized component types:

- `world_state`
- `story_state`
- `story_history`
- `tick_policy`
- `state_snapshot`

Added helper APIs:

- `store.GetSingleComponentByType()`
- `store.UpsertComponentByType()`
- `service.GetStateComponent()`
- `service.UpsertStateComponent()`

## Regression Coverage

Tests added/expanded for:

- log migration to `logs`
- structured pipeline logs
- full debug detail persistence
- review-mode pending plan logs
- world service log persistence
- state component upsert behavior
- world tick continuity prompt injection
- canonical fact retention in world state

## Follow-On Work

Suggested next steps after this refactor:

- expose state components and full logs more directly in Creator and DevCli
- add explicit timeline diff views for review mode
- evolve fact extraction from rule-based to schema-guided extraction
- add tick-policy authoring UX for continuity rules
