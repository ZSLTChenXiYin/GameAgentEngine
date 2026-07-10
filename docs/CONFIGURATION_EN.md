# Configuration

[**中文**](./CONFIGURATION.md) | **English**

GameAgentEngine uses a two-layer configuration model:

- static config: `gameagentengine.conf.yaml`
- dynamic world config: `world_settings`, `world_policy`, and state components stored in the database

---

## Two Kinds of Defaults

The repository currently exposes two different sources of “defaults”:

- code-level defaults registered in `internal/config/config.go`
- packaged template values defined in `tools/source/gameagentengine.conf.yaml`

For example, the code-level default for `engine.autonomous_scheduler_enabled` is currently `false`, while an older template revision still showed `true`. This documentation pass also updates the packaged template so both now point to the disabled state.

---

## Static Config File

The default template lives at `tools/source/gameagentengine.conf.yaml`.

Lookup order:

1. `--config <path>`
2. `./gameagentengine.conf.yaml`
3. `./config/gameagentengine.conf.yaml`

Current recommended template:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"
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

auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-v4-flash"
  api_key: "sk-xxx"
  base_url: "https://api.deepseek.com"

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

The code-level fallback defaults also include:

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  base_url: "https://api.openai.com/v1"

engine:
  execution_mode: "full"
```

This means that if you omit these fields from a config file, the engine still falls back to the internal defaults.

`database.driver` currently supports `sqlite`, `mysql`, and `postgres`.

---

## Key Static Toggles

### `database.migrations_enabled`

Controls whether schema / data migrations run during initialization.

### `database.write_retry_enabled`

Controls the shared retriable-write layer. When disabled, transient write conflicts surface directly to callers.

### `database.log_batch_enabled`

Controls whether inference logs are buffered and flushed in batches. When disabled, log writes fall back to direct persistence.

### `engine.world_lock_enabled`

Controls the business-level same-world exclusion boundary for critical heavy operations.

### `engine.autonomous_scheduler_enabled`

Controls the service-level autonomous scheduler. The recommended default is currently `false`.

---

## Dynamic Config: world_settings

`world_settings` is stored per world and commonly includes:

- `memory_limit`
- `max_analysis_rounds`
- `max_context_depth`
- `auto_apply`
- `require_review_above`
- `pipeline_mode`
- `propagation_max_depth`
- `sub_task_max_retries`
- `sub_task_timeout_secs`
- `enable_propagation_machine`
- `world_time_settings`

These settings affect one world's runtime behavior without changing other worlds.

---

## World Time Config

`world_time_settings` defines the rules, while `world_time_state` stores the resulting timeline state.

If valid `world_time_settings` are missing, world-time-dependent flows are intentionally blocked.

Core constraints:

- `tick_scale_mode` must be `fixed` or `flexible`
- `tick_units` must be non-empty and unique
- `tick_min_unit` must match the smallest configured unit

---

## Operational Guidance

When diagnosing pipeline or database issues, start with:

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`

Together they help reveal:

- whether write retries are spiking
- how many transactions are being executed
- whether the log queue is backing up
- whether world-level lock contention is growing
