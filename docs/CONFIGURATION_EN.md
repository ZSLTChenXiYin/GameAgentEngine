# Configuration

[**中文**](./CONFIGURATION.md) | **English**

GameAgentEngine v0.4.6 uses a two-layer configuration model:

- static config: `gameagentengine.conf.yaml`
- dynamic config: `world_settings`, `world_policy`, and state components stored in the database

---

## Static Config File

The default config file lives at `tools/source/gameagentengine.conf.yaml`.

Lookup order:

1. `--config <path>`
2. `./gameagentengine.conf.yaml`
3. `./config/gameagentengine.conf.yaml`

Current defaults:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"
  dsn: "gameagentengine.db"

auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-v4-flash"
  api_key: "sk-xxx"
  base_url: "https://api.deepseek.com"

engine:
  execution_mode: "debug"
  autonomous_scheduler_enabled: true
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

`database.driver` currently supports `sqlite`, `mysql`, and `postgres`.

---

## Dynamic Config: world_settings

`world_settings` is stored per world and currently includes:

- `memory_limit`
- `max_analysis_rounds`
- `max_context_depth`
- `auto_apply`
- `require_review_above`
- `pipeline_mode`
- `propagation_max_depth`
- `enable_propagation_machine`
- `sub_task_max_retries`
- `sub_task_timeout_secs`
- `world_time_settings`

---

## world_time_settings

`world_time_settings` is part of `world_settings`. It defines the world's time system and is not a static config field.

Important rules:

- `tick_scale_mode` must be `fixed` or `flexible`
- `tick_units` must be non-empty and unique
- `tick_units` must be ordered from large to small
- `tick_min_unit` must equal the last unit
- when `time_calendar` is enabled, `calendar_name` is required
- when `time_calendar` is enabled, `time_calendar.units` must exactly match `tick_units`

If this is missing, world-time-dependent flows are intentionally blocked as a developer reminder.

---

## world_time_state

`world_time_state` is not configuration. It is an Engine-maintained runtime state component representing the current world time result.

Think of them as:

- `world_time_settings`: rules
- `world_time_state`: result
