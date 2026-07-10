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
  callback_token: ""
  runtime_task_token: ""
  callback_require_request_id: false

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
  runtime_task_governance_interval_seconds: 30
  runtime_task_heartbeat_timeout_seconds: 300
  runtime_task_auto_requeue_enabled: false
  runtime_task_auto_requeue_limit: 100
  runtime_task_auto_requeue_delay_ms: 1000

external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:9000"
    path: "/api/v1/runtime/dispatch"
    timeout_ms: 5000
    retry_max_attempts: 2
    retry_backoff_ms: 100
    idempotency_header: "Idempotency-Key"
    auth:
      mode: "bearer"
      token: "replace-me"

  game_ws:
    type: "websocket_adapter"
    base_url: "ws://127.0.0.1:9001"
    path: "/ws/runtime/dispatch"
    timeout_ms: 5000

  game_rpc:
    type: "rpc_adapter"
    base_url: "tcp://127.0.0.1:9002"
    path: "Runtime.Dispatch"
    timeout_ms: 5000

external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "push"
    primary_transport: "game_http"
    consumer: "game_client"
    resume_policy: "resume_paused_execution"

  spawn_item:
    category: "external_action"
    delivery_mode: "hybrid"
    primary_transport: "game_http"
    consumer: "bridge"
    max_attempts: 3
    resume_policy: "none"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"
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

如果你希望直接照着接入样例配置外部交互，可以继续看 [外部交互接入示例](./EXTERNAL_INTERACTION_EXAMPLES.md)。

---

## 外部交互静态配置

`auth` 下面当前除了通用 `api_key` 之外，还支持外部执行面的专用安全字段：

- `callback_token`：配置后，`POST /api/v1/actions/callback` 可以使用 `X-Callback-Token` 单独鉴权
- `runtime_task_token`：配置后，`/api/v1/runtime/tasks/*` 接口可以使用 `X-Runtime-Task-Token` 单独鉴权
- `callback_require_request_id`：开启后，callback 请求必须携带 `X-Callback-Request-Id`，用于请求级防重放

当前鉴权优先级是：

- callback 接口：优先接受 `X-Callback-Token`，否则回落到通用 `X-API-Key`
- runtime task 接口：优先接受 `X-Runtime-Task-Token`，否则回落到通用 `X-API-Key`
- 其他接口：继续使用通用 `X-API-Key`

当前已经引入 `external_integrations` 作为 Engine 内建 `push` adapter 的静态配置入口。

当前也已经引入 `external_interfaces` 作为接口级正式路由配置入口，用于把“某个业务接口该走 push/pull/hybrid、使用哪个 transport、由谁消费、是否需要恢复执行”等策略从临时参数提升为正式配置。

已支持字段：

- `type`：当前支持 `http_adapter`、`websocket_adapter`、`rpc_adapter`
- `base_url`：外部服务基础地址
- `path`：推送路径；`http_adapter` 缺省为 `/api/v1/runtime/dispatch`，`websocket_adapter` 缺省为 `/ws/runtime/dispatch`，`rpc_adapter` 缺省为 `Runtime.Dispatch`
- `timeout_ms`：HTTP 请求超时
- `retry_max_attempts`：push 派发最大尝试次数，默认 `1`
- `retry_backoff_ms`：重试退避毫秒数，默认 `100`
- `idempotency_header`：当外部协议支持头部时，Engine 会把 task 级幂等键写入该请求头
- `headers`：附加请求头
- `auth.mode`：当前支持 `bearer`、`header`
- `auth.token`：认证令牌
- `auth.header_name`：当 `mode = header` 时使用的请求头名

`external_interfaces` 当前已支持这些关键字段：

- `category`：`external_query` 或 `external_action`
- `delivery_mode`：`push`、`pull`、`hybrid`
- `primary_transport`：主通道 integration 名称
- `fallback_transport`：`hybrid` 下 push 失败后的回退 transport 标签
- `consumer`：pull/hybrid 下默认消费方
- `max_attempts`：pull/hybrid 生命周期内允许的最大领取尝试次数；超过后 release/requeue 不再回队，而是进入终态 `failed`
- `resume_policy`：当前支持 `none`、`resume_paused_execution`
- `callback_post_process`：callback 完成后的统一后处理策略，当前支持 `none`、`record_only`、`write_memory`
- `callback_memory_level`：当 `callback_post_process = write_memory` 时写入的记忆层级，默认 `short_term`
- `callback_memory_template`：当 `callback_post_process = write_memory` 时使用的模板，支持 `{action_id}`、`{status}`、`{result_json}`、`{callback_id}`、`{interface_name}`、`{task_id}`、`{node_id}`、`{world_id}`、`{request_id}`、`{delivery_mode}`、`{primary_transport}` 占位符
- `timeout_ms`：接口级默认超时

当前实现边界：

- `game_client request_data` 可以通过 `delivery_mode: push|hybrid` + `primary_transport` 使用内建 `http_adapter`、`websocket_adapter` 或 `rpc_adapter`
- 普通 async action 也可以通过动作参数中的 `delivery_mode` / `primary_transport` 走相同 push 链路
- `game_client request_data` 默认读取 `game_client_request_data` 接口配置，也可以通过 `request_data.external_interface` 显式指定
- 普通 async action 默认读取同名 `action_id` 接口配置，也可以通过动作参数 `external_interface` 显式改绑
- callback 自动恢复当前会读取 runtime task payload 中的 `resume_policy`；为空或 `resume_paused_execution` 时自动恢复，`none` 时只回填结果不自动恢复
- `hybrid` 下如果 push 派发失败且配置了 `fallback_transport`，runtime task 会转为 `released`，并把 `transport` 写成该 fallback 值，供 pull consumer 后续领取
- 当前 `fallback_transport` 还不是“自动切到第二个 push adapter 再发一次”的意思，而是“明确进入 pull 风格回退态”的持久化语义
- 当前 `max_attempts` 已接入 runtime task 治理：claim 会累加 `attempt_count`，当达到上限后，后续 `release` 或 `heartbeat_timeout` requeue 会把任务稳定落到 `failed`，避免无限回队
- 当前 `hybrid` push 失败后还会把失败分类与回退决策写入 task 字段，例如 `last_dispatch_failure_class`、`last_dispatch_decision`、`fallback_from_transport`、`last_transition_reason`，方便后续治理与排查
- 普通 async action 当前已经支持统一 callback 后处理基础版；Engine 会把 `callback_post_process` 策略快照写入 runtime task payload，再由 callback handler 按持久化快照执行后处理，避免配置漂移或上下文压缩导致行为变化
- 当前已经提供基础 runtime task 管理面，包括聚合统计、条件查询、单任务详情和 `heartbeat_timeout` sweep 入口，方便运维排查与后续自动治理接线

`rpc_adapter` 当前最小实现约束：

- `base_url` 需要使用 `tcp://` 或 `unix://`
- `path` 表示 RPC 方法名，默认 `Runtime.Dispatch`
- 当前使用标准库 `net/rpc/jsonrpc` 做一次请求-响应式派发
- `auth` 与 `headers` 会作为 RPC 参数的一部分一起传入，不走独立传输层头部
- 当前 push 链路已经支持基础重试与幂等透传；runtime task 会记录 `idempotency_key`、`dispatch_attempts`、`last_dispatch_at`、`last_dispatch_error`
- `external_interfaces` 的完整双层业务配置模型仍在后续阶段继续补齐；当前属于过渡期可用实现

`resume_policy` 当前需要特别注意：

- `resume_paused_execution`：callback 成功后会尝试自动恢复原 paused execution
- `none`：callback 只更新 callback/runtime task 结果，不自动恢复
- 对普通 async action 而言，当前 dispatch request 仍固定以 `none` 语义投递；但 callback 完成后，除了原有 `OnResult(...)` 外，还可以通过 `callback_post_process` 走统一后处理层
- 对 `game_client request_data` 而言，若配置为 `none`，当前 paused execution 会保持 `paused`，这是现阶段刻意保留的边界，后续还会补显式恢复或统一后处理能力

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

### `engine.runtime_task_governance_interval_seconds`

控制后台 runtime task governor 的扫描间隔。大于 `0` 时，Engine 会周期性处理 heartbeat timeout 治理逻辑。

### `engine.runtime_task_heartbeat_timeout_seconds`

控制将 `claimed` / `running` task 判定为 `heartbeat_timeout` 的阈值秒数。

### `engine.runtime_task_auto_requeue_enabled`

控制 governor 在标记 `heartbeat_timeout` 后，是否继续自动批量 requeue 这些任务。

### `engine.runtime_task_auto_requeue_limit`

控制每轮自动 requeue 的最大任务数，避免一次性回收过多任务。

### `engine.runtime_task_auto_requeue_delay_ms`

控制自动 requeue 后，任务重新出现在 pending 列表前的延迟。

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
