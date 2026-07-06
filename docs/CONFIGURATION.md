# 配置参考

**中文** | [**English**](./CONFIGURATION_EN.md)

GameAgentEngine v0.2.0 采用双层配置体系：**静态配置**通过 YAML 文件管理，**动态配置**通过数据库 WorldSettings 管理。

---

## 静态配置（gameagentengine.conf.yaml）

由 Viper 管理，使用 YAML 格式。默认配置位于 `tools/source/gameagentengine.conf.yaml`。

### 搜索路径

1. 显式路径：`--config <路径>` 标志
2. 默认搜索：`./gameagentengine.conf.yaml`
3. 回退：`./config/gameagentengine.conf.yaml`

所有值也可以通过环境变量覆盖。

### 完整静态配置

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
  api_key: ""              # 留空使用 Mock Provider
  base_url: "https://api.deepseek.com/v1"

engine:
  execution_mode: "production"                    # debug / review / production
  autonomous_scheduler_enabled: false              # 后台自主调度器（服务级开关）
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

### 配置段说明

#### server

| 字段 | 默认值 | 说明 |
|---|---|---|
| `host` | `"0.0.0.0"` | 绑定地址 |
| `port` | `8080` | HTTP 服务端口 |

#### database

| 字段 | 默认值 | 说明 |
|---|---|---|
| `driver` | `"sqlite"` | `"sqlite"` 或 `"mysql"` |
| `dsn` | `"gameagentengine.db"` | SQLite：文件路径；MySQL：连接字符串 |

MySQL DSN 格式：`user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True`

#### auth

| 字段 | 默认值 | 说明 |
|---|---|---|
| `api_key` | `"dev-key"` | 通过 `X-API-Key` 请求头发送 |

#### llm

| 字段 | 默认值 | 说明 |
|---|---|---|
| `provider` | `"openai"` | 兼容任何 OpenAI 格式的 API |
| `model` | `"gpt-4o-mini"` | 模型标识 |
| `api_key` | `""` | 留空使用 Mock Provider |
| `base_url` | `"https://api.openai.com/v1"` | API 端点 |

已测试的支持模型：

- DeepSeek：`deepseek-chat`（`base_url` 设为 `https://api.deepseek.com/v1`）
- OpenAI：`gpt-4o-mini`、`gpt-4o`（`base_url` 设为 `https://api.openai.com/v1`）
- 阿里通义千问：`qwen-turbo`（`base_url` 设为 `https://dashscope.aliyuncs.com/compatible-mode/v1`）

#### engine（静态配置）

| 字段 | 默认值 | 说明 |
|---|---|---|
| `execution_mode` | `"production"` | `debug`、`review` 或 `production` |
| `autonomous_scheduler_enabled` | `false` | 后台自主行为调度器开关 |
| `autonomous_scheduler_interval_seconds` | `300` | 扫描间隔秒数 |
| `autonomous_scheduler_max_nodes_per_world` | `10` | 每次扫描最大触发数 |

### 环境变量覆盖

| 配置路径 | 环境变量 |
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

## 动态配置（数据库 WorldSettings）

以下配置项属于用户态动态配置，存储在数据库 `world_settings` 表中，每个世界独立。通过 DevCli 或 Creator 管理：

| 字段 | 默认值 | 说明 |
|---|---|---|
| `memory_limit` | 50 | 每次推理加载的最大记忆条数 |
| `max_analysis_rounds` | 5 | LLM 多轮轮询最大次数 |
| `max_context_depth` | 3 | 上下文向上追溯最大深度 |
| `auto_apply` | true | 是否自动执行变更计划 |
| `require_review_above` | critical | 超过此等级需审核 |
| `pipeline_mode` | full | 管线模式：vertical/polling/full |
| `propagation_max_depth` | 2 | 记忆向上传播最大层数；0 为不限制 |
| `sub_task_max_retries` | 2 | 子任务最大重试次数；0 表示禁用自动重试 |
| `sub_task_timeout_secs` | 60 | 子任务超时秒数；0 表示关闭超时保护 |
| `enable_propagation_machine` | false | 是否启用标签传播状态机 |

### 通过 DevCli 配置

```bash
# 查看当前设置
GameAgentDevCli world settings get <world-id>

# 只修改需要变更的字段
GameAgentDevCli world settings set <world-id> \
  --pipeline-mode "polling" \
  --propagation-max-depth 0 \
  --sub-task-max-retries 0 \
  --sub-task-timeout-secs 0 \
  --enable-propagation-machine false
```

CLI 发送的是部分更新请求，未传入的 flag 会保留当前值。

### 通过 Creator 配置

在 Creator 的 Settings 页面中直接调整上述参数。

---

## 配置示例

### DeepSeek

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-你的密钥"
  base_url: "https://api.deepseek.com/v1"
```

### MySQL

```yaml
database:
  driver: "mysql"
  dsn: "gameuser:password@tcp(127.0.0.1:3306)/gameagent?charset=utf8mb4&parseTime=True"
```

### Mock（无 API Key）

```yaml
llm:
  api_key: ""    # 引擎使用 Mock Provider
```

---

## 重要原则

> **配置文件仅保存静态配置**，即启动后不会轻易修改的服务级参数（监听地址、数据库连接、LLM 接入信息等）。
>
> **用户态动态配置**（推理参数、管线模式、传播规则、策略等）应通过数据库 WorldSettings 和 WorldPolicy 管理，通过 DevCli 或 Creator 修改，而非写入配置文件。
