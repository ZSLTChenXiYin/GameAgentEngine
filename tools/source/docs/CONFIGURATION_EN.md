# Configuration Reference

[**中文**](./CONFIGURATION.md) | **English**

GameAgentEngine v0.2.0 uses a two-layer configuration system: **static configuration** managed through YAML files, and **dynamic configuration** managed through the database WorldSettings.

---

## Static Configuration (gameagentengine.conf.yaml)

Managed by Viper using YAML format. The default configuration is at `tools/source/gameagentengine.conf.yaml`.

### Search Paths

1. Explicit path: `--config <path>` flag
2. Default search: `./gameagentengine.conf.yaml`
3. Fallback: `./config/gameagentengine.conf.yaml`

All values can also be overridden through environment variables.

### Full Static Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"         # sqlite / mysql
  dsn: "gameagentengine.db"

auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: ""              # leave empty for Mock Provider
  base_url: "https://api.deepseek.com/v1"

engine:
  execution_mode: "production"                    # debug / review / production
  autonomous_scheduler_enabled: false              # background autonomous scheduler (service-level toggle)
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

### Configuration Sections

#### server

| Field | Default | Description |
|---|---|---|
| `host` | `"0.0.0.0"` | Bind address |
| `port` | `8080` | HTTP service port |

#### database

| Field | Default | Description |
|---|---|---|
| `driver` | `"sqlite"` | `"sqlite"` or `"mysql"` |
| `dsn` | `"gameagentengine.db"` | SQLite: file path; MySQL: connection string |

MySQL DSN format: `user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True`

#### auth

| Field | Default | Description |
|---|---|---|
| `api_key` | `"dev-key"` | Sent via the `X-API-Key` request header |

#### llm

| Field | Default | Description |
|---|---|---|
| `provider` | `"openai"` | Compatible with any OpenAI-format API |
| `model` | `"gpt-4o-mini"` | Model identifier |
| `api_key` | `""` | Leave empty for Mock Provider |
| `base_url` | `"https://api.openai.com/v1"` | API endpoint |

Tested supported models:

- DeepSeek: `deepseek-chat` (set `base_url` to `https://api.deepseek.com/v1`)
- OpenAI: `gpt-4o-mini`, `gpt-4o` (set `base_url` to `https://api.openai.com/v1`)
- Alibaba Qwen: `qwen-turbo` (set `base_url` to `https://dashscope.aliyuncs.com/compatible-mode/v1`)

#### engine (static config)

| Field | Default | Description |
|---|---|---|
| `execution_mode` | `"production"` | `debug`, `review`, or `production` |
| `autonomous_scheduler_enabled` | `false` | Background autonomous behavior scheduler toggle |
| `autonomous_scheduler_interval_seconds` | `300` | Scan interval in seconds |
| `autonomous_scheduler_max_nodes_per_world` | `10` | Max triggers per scan |

### Environment Variable Overrides

| Config Path | Environment Variable |
|---|---|
| `server.host` | `SERVER_HOST` |
| `server.port` | `SERVER_PORT` |
| `database.driver` | `DATABASE_DRIVER` |
| `database.dsn` | `DATABASE_DSN` |
| `auth.api_key` | `AUTH_API_KEY` |
| `llm.provider` | `LLM_PROVIDER` |
| `llm.model` | `LLM_MODEL` |
| `llm.api_key` | `LLM_API_KEY` |
| `llm.base_url` | `LLM_BASE_URL` |
| `engine.execution_mode` | `ENGINE_EXECUTION_MODE` |

---

## Dynamic Configuration (Database WorldSettings)

The following configuration items are user-level dynamic configuration stored in the `world_settings` table, independent per world. Managed via DevCli or Creator:

| Field | Default | Description |
|---|---|---|
| `memory_limit` | 50 | Max memories loaded per inference |
| `max_analysis_rounds` | 5 | Max LLM polling rounds |
| `max_context_depth` | 3 | Max context traceback depth |
| `auto_apply` | true | Whether to auto-apply change plans |
| `require_review_above` | critical | Impact level above which review is required |
| `pipeline_mode` | full | Pipeline mode: vertical/polling/full |
| `propagation_max_depth` | 2 | Max upward memory propagation depth; 0 = unlimited |
| `sub_task_max_retries` | 2 | Max sub-task retries |
| `sub_task_timeout_secs` | 60 | Sub-task timeout in seconds |
| `enable_propagation_machine` | false | Enable tag propagation state machine |

### Configuring via DevCli

```bash
# View current settings
GameAgentDevCli world settings get <world-id>

# Modify settings
GameAgentDevCli world settings set <world-id> \
  --pipeline-mode "full" \
  --propagation-max-depth 3 \
  --sub-task-max-retries 3 \
  --sub-task-timeout-secs 120 \
  --enable-propagation-machine true
```

### Configuring via Creator

Adjust the above parameters directly in the Settings page of the Creator.

---

## Configuration Examples

### DeepSeek

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-your-key"
  base_url: "https://api.deepseek.com/v1"
```

### MySQL

```yaml
database:
  driver: "mysql"
  dsn: "gameuser:password@tcp(127.0.0.1:3306)/gameagent?charset=utf8mb4&parseTime=True"
```

### Mock (No API Key)

```yaml
llm:
  api_key: ""    # Engine uses Mock Provider
```

---

## Important Principle

> **The config file only holds static configuration** — service-level parameters that are not expected to change after startup (listen address, database connection, LLM access info, etc.).
>
> **User-level dynamic configuration** (inference parameters, pipeline mode, propagation rules, policies, etc.) should be managed through database WorldSettings and WorldPolicy, modified via DevCli or Creator, not by editing the config file.