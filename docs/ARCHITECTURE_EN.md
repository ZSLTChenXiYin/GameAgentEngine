# Architecture

[**中文**](./ARCHITECTURE.md) | **English**

GameAgentEngine v0.2.0 adopts a layered architecture with clear separation of concerns. The system is designed as a backend service with an HTTP API, SDK, CLI tools, and a web-based visual editor.

---

## High-Level Architecture

```mermaid
flowchart TB
    subgraph Tools ["Tools Layer"]
        Creator["GameAgentCreator\n(Web editor)"]
        DevCli["GameAgentDevCli\n(CLI tool)"]
        Demo["Web Demo"]
    end

    subgraph Backend ["Backend (GameAgentEngine)"]
        API["HTTP API\n(internal/api)"]
        SVC["Domain Service\n(internal/service)"]
        ENG["Inference Pipeline\n(internal/engine)"]
        STORE["Persistence Layer\n(internal/store)"]
        LLM["LLM Provider\n(internal/llm)"]
        ACT["Action System\n(internal/action)"]
        PLN["Policy Engine\n(internal/planner)"]
    end

    SDK["Go SDK\n(sdk/)"]

    Tools -->|HTTP| API
    DevCli -->|HTTP| SDK -->|HTTP| API
    Creator -->|HTTP| API
    API --> SVC
    SVC --> ENG
    ENG --> STORE
    ENG --> LLM
    ENG --> ACT
    ENG --> PLN
    STORE --> SQLite["SQLite / MySQL"]
```

---

## Layer Descriptions

### 1. API Layer (internal/api)

HTTP entry point. Routes requests to the appropriate handlers, validates input, and maps errors to HTTP status codes.

- **Router** (router.go) — registers all endpoints on `http.ServeMux`
- **Handlers** (invoke.go, world.go, world_settings.go, policy.go, etc.) — request parsing and response serialization
- **Middleware** (middleware.go) — API key authentication, CORS, idempotency
- **Service error mapping** (service_error.go) — maps 18 domain error codes to HTTP status codes

### 2. Domain Service Layer (internal/service)

Contains business rules and transaction boundaries. Prevents duplicated validation logic across HTTP/CLI/editor.

- **CRUD operations** — create/update/delete for nodes, components, memories, and relations with full validation
- **World import/export** (graph.go) — YAML/JSON world config import with dry-run support
- **World Tick** (world.go) — timeline advancement, autonomous node scheduling, event impact evaluation, scope advancement
- **World cloning** (clone.go) — duplicates a complete world with all its data, optionally locking the source world against concurrent writes
- **Autonomous behavior management** — configure, query, and manually trigger autonomous node behavior cycles

### 3. Engine Layer (internal/engine)

The core inference pipeline. Handles the entire inference lifecycle:

- **Three pipeline modes** — vertical (single-pass), polling (multi-round LLM), full (complete with DAG sub-tasks)
- **Context builder** (context.go) — loads node data, components, memories, and ancestor tree from storage
- **Prompt generation** (prompt_builders.go) — builds task-specific system prompts
- **Multi-round polling** (pipeline.go) — supports multiple LLM dialogue rounds, with request_data queries per round
- **Sub-task DAG** (dag.go) — orchestrates directed acyclic graphs of sub-tasks declared by the LLM, with retry, timeout, and merge modes
- **Task node tree** (tasktree.go) — records the complete inference trace for context inheritance
- **Memory propagation engine** (propagation_engine.go) — four propagation modes (upward/tag_broadcast/targeted/manual) with optional state machine
- **LLM invocation** — delegates to the configured LLM Provider
- **Action execution** — executes synchronous actions in-pipeline, returns async action callbacks
- **Memory persistence & propagation** — writes memory updates and propagates to target nodes

### 4. Storage Layer (internal/store)

GORM-based persistence. Handles database connection, auto-migration, and CRUD operations.

- **Models** (models.go) — 9 data models: Node, Component, Memory, Relation, Timeline, InferenceLog, IdempotencyKey, WorldPolicy, WorldSettings
- **Node operations** (nodes.go) — CRUD + paginated filtering
- **Component operations** (components.go) — get by node, by type, by world
- **Memory operations** (memories.go) — CRUD + level filtering, bulk creation, manual propagation
- **Relation operations** (relations.go) — CRUD + paginated filtering, get node-related relations
- **Timeline & logs** (timeline.go) — timeline ticks, inference logs
- **World settings** (world_settings.go) — per-world runtime settings with CRUD + defaults
- **World policy** (policy.go) — per-world blocked_actions / safe_actions policy
- **Memory propagation** (propagation.go) — propagation rules, propagation state machine state

### 5. LLM Provider (internal/llm)

Abstracts LLM API calls through a common interface:

- **OpenAI Provider** (openai.go) — compatible with any OpenAI-format API (OpenAI, DeepSeek, Qwen, etc.)
- **Mock Provider** (mock.go) — simulates LLM responses for offline development and testing

### 6. Action System (internal/action)

A registry-based action system supporting both synchronous and asynchronous modes:

- **Sync actions** — executed immediately within the pipeline (add_memory, update_mood, send_dialogue)
- **Async actions** — return a callback ID for the game side to execute (adjust_relation, spawn_item)

### 7. Planner & Policy (internal/planner)

Evaluates world change plans against configured policy:

- **PolicyEngine** — blocks dangerous actions, requires review for high-impact changes
- **ExecutionMode** — debug (verbose logging), review (high-impact requires confirmation), production (auto-apply)

---

## Data Flow: NPC Dialogue

```mermaid
sequenceDiagram
    participant Client as GameAgentCreator / DevCli
    participant API as HTTP API
    participant SVC as Domain Service
    participant ENG as Inference Pipeline
    participant LLM as LLM Provider
    participant DB as Database

    Client->>API: POST /api/v1/invoke {world_id, task_type, node_id, messages}
    API->>API: Validate request
    API->>ENG: Pipeline.Execute(request)
    ENG->>DB: Load nodes, components, memories, relations, ancestors
    ENG->>ENG: Load WorldSettings (PipelineMode, etc.)
    ENG->>ENG: Build system prompt
    alt vertical mode
        ENG->>LLM: Chat() — single call
    else polling mode
        loop per round
            ENG->>LLM: Chat()
            LLM-->>ENG: JSON response
            opt request_data not empty
                ENG->>DB: Load additional data
            end
        end
    else full mode
        loop per round
            ENG->>LLM: Chat()
            LLM-->>ENG: JSON {sub_tasks, ...}
            opt sub_tasks not empty
                ENG->>ENG: Create DAGInstance
                loop ready sub-tasks
                    ENG->>LLM: Execute sub-task
                    ENG->>ENG: Aggregate results
                end
            end
        end
    end
    ENG->>ENG: Execute sync actions, handle memory propagation
    ENG->>DB: Write memories, propagation records, inference logs
    ENG-->>API: InvokeResponse
    API-->>Client: JSON response
```

---

## Data Flow: World Tick

```mermaid
sequenceDiagram
    participant Client as GameAgentCreator / DevCli
    participant API as HTTP API
    participant SVC as Domain Service
    participant ENG as Inference Pipeline
    participant DB as Database

    Client->>API: POST /api/v1/worlds/{id}/ticks/advance
    API->>SVC: AdvanceWorldTickWithAutonomous(pipeline, worldID, tickType)
    SVC->>ENG: Pipeline.Execute(world_tick task)
    ENG->>DB: Load world context, WorldSettings
    ENG->>ENG: Execute inference per PipelineMode
    LLM-->>ENG: JSON {world_change_plan, future_outline, memory_updates}
    ENG->>ENG: Policy engine evaluates plan
    ENG->>DB: Write memories, logs, propagation
    SVC->>DB: Create TimelineModel (Tick record)
    SVC->>ENG: RunWorldTickAutonomous (trigger world_tick_sync autonomous nodes)
    SVC-->>API: {tick, invoke, autonomous_runs}
    API-->>Client: JSON response
```

---

## Memory Propagation

The engine supports four memory propagation modes to help memories flow between node levels:

| Mode | Description |
|---|---|
| upward | Propagate up the parent chain (default); depth limited by max_depth |
| tag_broadcast | Spread to nodes matching given tags |
| targeted | Direct propagation to a specified list of nodes |
| manual | No automatic propagation; user triggers manually |

Propagation can be configured as a state machine (enable_propagation_machine), which automatically executes propagation actions according to preset rule chains.

---

## Configuration System

Configuration is divided into two layers:

- **Static config** (gameagentengine.conf.yaml): server address, database connection, LLM access info, execution mode, autonomous scheduler parameters
- **Dynamic config** (database WorldSettings): pipeline mode, memory limit, analysis rounds, context depth, sub-task retry/timeout, propagation parameters

See [Configuration Reference](CONFIGURATION_EN.md).

---

## Database Schema

Nine tables managed by GORM AutoMigrate:

- **nodes** — id, world_id, name, node_type, parent_id, timestamps
- **components** — id, node_id, component_type, data, timestamps
- **memories** — id, node_id, content, level, tags, created_at
- **relations** — id, world_id, source_id, target_id, relation_type, weight, properties, created_at
- **timelines** — id, world_id, tick_number, tick_type, game_time, summary, data, future_outline, created_at
- **inference_logs** — id, world_id, task_type, node_id, request_data, response_data, llm_model, tokens_used, duration_ms, created_at
- **idempotency_keys** — id, result, created_at
- **world_policies** — world_id, blocked_actions, safe_actions, timestamps
- **world_settings** — world_id, memory_limit, max_analysis_rounds, max_context_depth, auto_apply, require_review_above, pipeline_mode, propagation_max_depth, sub_task_max_retries, sub_task_timeout_secs, enable_propagation_machine, timestamps