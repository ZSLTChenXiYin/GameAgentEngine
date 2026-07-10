# 配置参考

**中文** | [**English**](./CONFIGURATION_EN.md)

GameAgentEngine 采用双层配置：

- 静态配置：`gameagentengine.conf.yaml`
- 动态世界配置：数据库中的 `world_settings`、`world_policy` 和状态组件

---

## 两类默认值要分开理解

仓库里当前同时存在两种“默认值”：

- 代码级默认值：由 `internal/config/config.go` 注册，在缺失配置项时生效
- 随包模板值：写在 `tools/source/gameagentengine.conf.yaml`，更偏向本地演示与开箱体验

例如，代码级默认的 `engine.autonomous_scheduler_enabled` 当前是 `false`，而旧模板曾经写成 `true`。本次文档和模板都已经统一到关闭状态。

---

## 静态配置文件

默认模板位于 `tools/source/gameagentengine.conf.yaml`。

搜索顺序：

1. `--config <路径>`
2. `./gameagentengine.conf.yaml`
3. `./config/gameagentengine.conf.yaml`

当前推荐模板：

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

external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:9000"
    path: "/api/v1/runtime/dispatch"
    timeout_ms: 5000
    auth:
      mode: "bearer"
      token: "replace-me"
```

代码级缺省值中，还包括这些重要字段：

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  base_url: "https://api.openai.com/v1"

engine:
  execution_mode: "full"
```

这表示如果你不显式提供模板文件中的值，Engine 会回落到代码里的保底默认值。

`database.driver` 当前支持 `sqlite`、`mysql`、`postgres`。

---

## 外部交互静态配置

当前已经引入 `external_integrations` 作为 Engine 内建 `push` adapter 的静态配置入口。

已支持字段：

- `type`：当前首个实现是 `http_adapter`
- `base_url`：外部服务基础地址
- `path`：推送路径，缺省为 `/api/v1/runtime/dispatch`
- `timeout_ms`：HTTP 请求超时
- `headers`：附加请求头
- `auth.mode`：当前支持 `bearer`、`header`
- `auth.token`：认证令牌
- `auth.header_name`：当 `mode = header` 时使用的请求头名

当前实现边界：

- `game_client request_data` 可以通过 `delivery_mode: push|hybrid` + `primary_transport` 使用内建 `http_adapter`
- 普通 async action 也可以通过动作参数中的 `delivery_mode` / `primary_transport` 走相同 push 链路
- `external_interfaces` 的完整双层业务配置模型仍在后续阶段继续补齐；当前属于过渡期可用实现

---

## 关键静态开关

### `database.migrations_enabled`

控制初始化时是否执行 schema / data migrations。

### `database.write_retry_enabled`

控制统一可重试写层是否启用。关闭后，数据库写冲突将直接暴露给调用方。

### `database.log_batch_enabled`

控制推理日志是否走内存队列批量落库。关闭后会回退为直接写入。

### `engine.world_lock_enabled`

控制同世界关键重操作是否经过业务级互斥边界。默认建议保持开启。

### `engine.autonomous_scheduler_enabled`

控制服务级后台自主行为调度器。当前推荐默认值是 `false`。

---

## 动态配置：world_settings

`world_settings` 是每个世界独立的运行时配置，常见字段包括：

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

这些配置影响的是单个世界如何推理，不影响其他世界。

---

## 时间系统配置

`world_time_settings` 定义时间规则，`world_time_state` 保存时间结果。

如果没有先配置有效的 `world_time_settings`，依赖世界时间推进的流程会被显式阻塞。

核心约束：

- `tick_scale_mode` 必须是 `fixed` 或 `flexible`
- `tick_units` 不能为空且不能重复
- `tick_min_unit` 必须等于最小单位

---

## 运维建议

排查数据库或管线问题时，优先结合以下入口：

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`

这样可以同时看到：

- 写重试是否频繁
- 事务累计次数与耗时
- 日志队列是否堆积
- 世界级锁是否争用
