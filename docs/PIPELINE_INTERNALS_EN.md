# Inference Pipeline Internals

[**中文**](./PIPELINE_INTERNALS.md) | **English**

This document details the internal mechanisms of the GameAgentEngine v0.2.0 inference pipeline, including pipeline modes, multi-round polling, sub-task DAG, data request loops, and memory propagation.

---

## PipelineMode

Each world can independently configure its pipeline mode, stored in the database WorldSettings:

| Mode | Value | Behavior |
|---|---|---|
| Vertical | `vertical` | Single LLM call, no task node tree, no polling, minimal features |
| Polling | `polling` | Multi-round LLM polling, supports request_data queries, no sub-task DAG |
| Full | `full` | Full features: multi-round polling + DAG sub-task orchestration |

The mode is configured through DevCli or Creator via WorldSettings, independently from ExecutionMode (debug/review/production).

---

## Pipeline.Execute Main Flow

1. Load world settings (memory_limit, max_analysis_rounds, pipeline_mode, etc.)
2. Load world policy (blocked_actions / safe_actions)
3. Build initial context (BuiltContext)
4. Dispatch by task type:
   - `npc_dialogue` → executeDialogue (includes first-round analysis)
   - `world_tick` → executeWorldTick
   - `world_event_impact` → executeWorldEvent
   - `autonomous_act` → executeAutonomousAct (with capability filtering)
   - `custom` → executeCustom
5. Each task type ultimately enters the executeMultiTurnLoop common loop

---

## executeMultiTurnLoop Common Loop

The multi-round inference loop is the core engine for all task types:

```
Round 1:
  1. Build system prompt (with context + task node tree + instructions)
  2. Call LLM
  3. Parse JSON response
  4. If full mode, check raw_sub_tasks → create DAGInstance
  5. Process request_data → async data request wait
  6. Execute sync actions
  7. Write memory updates → propagate
  8. Build next TaskNode
  9. Log inference

Round 2+ (when needed):
  - Include previous round analysis in context
  - Call LLM again
  - Same processing as above

Ends when max_analysis_rounds is reached or LLM marks decision=stop
```

---

## Sub-task DAG Orchestration

In `full` mode, the LLM can declare a `sub_tasks` array in its JSON response:

```json
{
  "reply": "Need to investigate several aspects in parallel.",
  "sub_tasks": [
    {"label": "investigate_market", "task_type": "custom", "node_id": "...", "depends_on": []},
    {"label": "assess_military", "task_type": "custom", "node_id": "...", "depends_on": []},
    {"label": "make_plan", "task_type": "custom", "node_id": "...", "depends_on": ["investigate_market", "assess_military"]}
  ]
}
```

DAGInstance is responsible for:

- **Registering** sub-task declarations
- **Dependency resolution** — empty depends_on = immediately ready; non-empty = wait for predecessors
- **Concurrent execution** of ready sub-tasks (goroutines)
- **Retry & timeout** — each sub-task retries up to MaxRetries times with TimeoutDuration
- **Result merging** — supports three merge modes:
  - `append` (default): concatenates all results
  - `override`: later results replace earlier ones
  - `summarize`: LLM semantic summary
- **Failure handling** — failures do not block the dependency chain; failure info is appended to the final reply

---

## Data Request Loop

The LLM can issue `request_data` queries in its response. The pipeline executes the following logic:

1. Parse DataRequest queries
2. For `target="store"` queries, execute data loading (node components, memories, relations)
3. For `target="game_client"` queries, wait for an external response via the callback mechanism
4. After loading, inject data into the next round's context
5. The loop runs at most max_analysis_rounds times

---

## Action Execution Flow

1. Pipeline parses the LLM's output action_calls
2. For autonomous_act tasks, perform capability validation and schema validation
3. Sync actions execute immediately within the pipeline
4. Async actions return a callback_id to the caller
5. The caller reports results via POST /api/v1/actions/callback
6. ActionRegistry matches the callback_id and stores the result

---

## Memory Processing & Propagation

1. LLM declares memory_updates (with propagation rules)
2. Pipeline creates MemoryModel entries and persists them
3. Executes propagation per PropagationRule:
   - upward: recursively upload along the parent chain (limited by propagation_max_depth)
   - tag_broadcast: match nodes in the same world with the same tags
   - targeted: write to a specified list of NodeIDs
   - manual: no automatic propagation
4. Optionally enable state machine mode (enable_propagation_machine):
   - Check rule chain after each propagation round
   - Execute PropagateAction when trigger conditions are met
   - Supports TransformRule (content prefix, level promotion, tag appending)

---

## Configuration Reference

### Static Configuration (gameagentengine.conf.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"    # sqlite / mysql
  dsn: "gameagentengine.db"

auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: ""
  base_url: "https://api.deepseek.com/v1"

engine:
  execution_mode: "production"    # debug / review / production
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

### Dynamic Configuration (Database WorldSettings)

Configurable via DevCli or Creator:

| Field | Default | Description |
|---|---|---|
| memory_limit | 50 | Max memories loaded per inference |
| max_analysis_rounds | 5 | Max LLM polling rounds |
| max_context_depth | 3 | Max context traceback depth |
| auto_apply | true | Auto-apply change plans |
| require_review_above | critical | Impact level requiring review |
| pipeline_mode | full | Pipeline mode: vertical/polling/full |
| propagation_max_depth | 2 | Max upward memory propagation depth |
| sub_task_max_retries | 2 | Max sub-task retries |
| sub_task_timeout_secs | 60 | Sub-task timeout in seconds |
| enable_propagation_machine | false | Enable tag propagation state machine |