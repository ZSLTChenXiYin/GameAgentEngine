# 外部交互接入示例

**中文**

本文档给出 Engine 外部交互的三种推荐接入方式：`push`、`pull`、`hybrid`。

目标不是把所有项目强行收敛到同一种模式，而是帮助开发者根据自己游戏端的网络形态、部署方式和容错要求做选择。

---

## 一、推荐选择

| 场景 | 推荐模式 | 原因 |
|---|---|---|
| Engine 与游戏逻辑服务在同一受控网络 | `push` | 链路最短，时延最低 |
| 游戏客户端不方便暴露入站服务 | `pull` | 由客户端或 bridge 主动领取，网络适配更稳 |
| 既有服务端执行面又有客户端执行面 | `hybrid` | 主链路快，失败时可回落到队列 |

---

## 二、Push 示例

适合 Engine 直接向游戏端服务发 HTTP / WebSocket / RPC 请求。

```yaml
external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:9000"
    path: "/api/v1/runtime/dispatch"
    timeout_ms: 5000
    retry_max_attempts: 2
    retry_backoff_ms: 100
    idempotency_header: "Idempotency-Key"

external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "push"
    primary_transport: "game_http"
    consumer: "game_client"
    resume_policy: "resume_paused_execution"
```

链路：

1. Engine 生成 runtime task。
2. Engine 通过内建 adapter 主动派发。
3. 游戏端完成后调用 `POST /api/v1/actions/callback`。
4. 如果该任务属于 paused execution，Engine 自动恢复原推理。

---

## 三、Pull 示例

适合游戏客户端、编辑器插件或本地 bridge 主动领取任务。

```yaml
external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "pull"
    consumer: "game_client"
    max_attempts: 5
    resume_policy: "resume_paused_execution"
```

bridge / 游戏端的典型流程：

1. `GET /api/v1/runtime/tasks/pending?consumer=game_client`
2. `POST /api/v1/runtime/tasks/claim`
3. `POST /api/v1/runtime/tasks/start`
4. 执行真实外部逻辑
5. `POST /api/v1/actions/callback`

如果执行中断：

1. `POST /api/v1/runtime/tasks/release`
2. 或长时间无心跳后由 governor 标记为 `heartbeat_timeout`
3. 再由人工或自动逻辑执行 `requeue`

---

## 四、Hybrid 示例

适合优先走主动派发，但需要稳定回退到队列的生产环境。

```yaml
external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:9000"
    path: "/api/v1/runtime/dispatch"

external_interfaces:
  spawn_item:
    category: "external_action"
    delivery_mode: "hybrid"
    primary_transport: "game_http"
    fallback_transport: "task_pull"
    consumer: "bridge"
    max_attempts: 3
    resume_policy: "none"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"
```

当前语义要点：

- `fallback_transport` 表示 push 失败后，任务进入 pull 风格的 `released` 状态，并写入该 transport 标签。
- 它当前不表示“自动切换到第二个 push adapter 再发一次”。
- `max_attempts` 约束的是 pull/hybrid 阶段的领取重试次数；达到上限后，release 或 timeout requeue 会把任务落到终态 `failed`。

---

## 五、Callback 后处理示例

普通 async action 当前除了原有 `OnResult(...)` 之外，还支持统一后处理基础版。

```yaml
external_interfaces:
  spawn_item:
    category: "external_action"
    delivery_mode: "pull"
    consumer: "bridge"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"
```

当前支持：

- `none`
- `record_only`
- `write_memory`

`write_memory` 模式下可用占位符：

- `{action_id}`
- `{status}`
- `{result_json}`
- `{callback_id}`
- `{interface_name}`
- `{task_id}`
- `{node_id}`
- `{world_id}`
- `{request_id}`
- `{delivery_mode}`
- `{primary_transport}`

这里使用的是 runtime task payload 中的策略快照，所以即使接口配置在任务发出后被修改，callback 仍会按原策略执行。

---

## 六、推荐运维观察点

排查外部交互链路时，优先看这些信息：

- `GET /api/v1/runtime/tasks`
- `GET /api/v1/runtime/tasks/stats`
- `GET /api/v1/runtime/tasks/{task_id}`
- `GET /api/v1/logs`

重点字段：

- `status`
- `delivery_mode`
- `transport`
- `attempt_count`
- `max_attempts`
- `dispatch_attempts`
- `last_dispatch_error`
- `callback_id`
- `resume_execution_id`

---

## 七、当前边界

当前已经完成生产可用基础版，但仍要注意：

- `callback` 仍然是入站完成机制，不是出站投递机制。
- `fallback_transport` 还不是多 push adapter 串联重试。
- `max_attempts` 已支持终态阈值，但还不是完整 dead-letter 管理面。
- 按 consumer/category 的差异化阈值策略、自动回切策略仍属于后续增强。
