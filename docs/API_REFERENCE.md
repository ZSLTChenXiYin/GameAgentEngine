# API 参考

**中文** | [**English**](./API_REFERENCE_EN.md)

除特别说明外，所有引擎 API 都通过 `X-API-Key` 请求头鉴权。

外部执行面相关接口还有两个可选专用头：

- `X-Callback-Token`：可单独用于 `POST /api/v1/actions/callback`
- `X-Runtime-Task-Token`：可单独用于 `/api/v1/runtime/tasks/*`

基础前缀：`/api/v1/`

多数写操作接口同时支持 `Idempotency-Key` 请求头，便于对 Tick、复制、导入等长流程做安全重试。

---

## 健康检查与版本

### `GET /health`

健康检查，不需要 API Key。

### `GET /api/v1/version`

返回当前引擎版本和最低兼容版本。

---

## 推理与计划

### `POST /api/v1/invoke`

统一推理入口。

典型请求体：

```json
{
  "world_id": "world-id",
  "node_id": "node-id",
  "task_type": "npc_dialogue",
  "messages": [
    { "role": "user", "content": "你好" }
  ],
  "context": {
    "pipeline_mode": "polling",
    "max_analysis_rounds": 4
  }
}
```

### `POST /api/v1/actions/callback`

完成异步动作回调。

典型请求体：

```json
{
  "callback_id": "callback-id",
  "status": "success",
  "result": {
    "scene": "tavern"
  }
}
```

当前行为：

- 对普通异步动作：更新 callback 记录，并将结果交给对应 action 的 `OnResult(...)`。
- 对由 `request_data.target = "game_client"` 触发的暂停执行：当该 task payload 的 `resume_policy` 为空或 `resume_paused_execution` 时，Engine 会自动恢复原始多轮推理，并在响应中返回 `resumed` 字段，携带恢复后的最终推理结果。
- 如果该 paused execution 对应的 `resume_policy = none`，当前 callback 只会回填结果，不会自动恢复，原执行会继续保持 `paused`。
- 如果该 callback 对应某个 `runtime task`，Engine 还会同步将该任务标记为 `succeeded`、`failed` 或 `cancelled`，写回完成结果。

安全与防重放：

- 如果配置了 `auth.callback_token`，该接口可以使用 `X-Callback-Token` 单独鉴权
- 如果配置了 `auth.callback_require_request_id`，请求必须携带 `X-Callback-Request-Id`
- 当携带 `X-Callback-Request-Id` 时，Engine 会把它作为 callback 请求级幂等键处理；相同请求会直接回放结果，不同 payload 复用同一 ID 会返回冲突

### `GET /api/v1/runtime/tasks/pending`

列出当前可被 `pull` consumer 领取的 runtime task。

如果配置了 `auth.runtime_task_token`，该接口以及其他 `/api/v1/runtime/tasks/*` 接口都可以使用 `X-Runtime-Task-Token` 单独鉴权。

查询参数：

- `consumer` - 可选，例如 `bridge`、`game_client`
- `limit` - 可选，默认 `20`，最大 `200`

当前返回的任务来自统一 `runtime_tasks` 队列，只包含状态为 `pending` 或 `released` 且已经到达可领取时间的任务。

当前 `game_client request_data` 已经接入该队列：当多轮推理因为请求游戏端数据而暂停时，Engine 会自动生成对应的 pull task，供游戏端或 bridge 领取执行。

普通 async action 也已经接入该队列：当模型输出异步动作调用时，Engine 会为对应 callback 同步生成 `external_action` 类型的 runtime task，默认由 `bridge` 消费；如果动作参数中显式提供 `consumer`，则按该值路由。

当前如果外部交互配置为 `push` 或 `hybrid`，Engine 也可能先通过内建 adapter 主动派发；当前已支持 `http_adapter`、`websocket_adapter` 与 `rpc_adapter`。此时任务状态会进入 `dispatched`，不再出现在 pending 列表里，等待后续 callback 完成。

当前 push 基础能力还包括：

- 可按 integration 配置执行基础重试
- 会为每个 runtime task 生成稳定幂等键并向支持头部的外部协议透传
- 任务侧会记录 dispatch 尝试次数与最近一次派发错误
- 当 `hybrid` task 的 push 失败且路由配置了 `fallback_transport` 时，任务会被显式转为 `released`，后续重新出现在 pending 列表里，供 pull consumer 继续消费

### `GET /api/v1/runtime/tasks`

按条件查询 runtime task 管理视图。

查询参数：

- `consumer` - 可选
- `category` - 可选，例如 `external_query`、`external_action`
- `interface_name` - 可选
- `transport` - 可选，例如 `game_http`、`task_pull`
- `world_id` - 可选
- `status` - 可选，逗号分隔，例如 `pending,released`
- `available_only` - 可选，传 `true` 时只返回已经到达可领取时间的任务
- `limit` - 可选，默认 `20`，最大 `200`

该接口用于运维排查和管理面查询，不替代 `pending` 拉取接口。

### `GET /api/v1/runtime/tasks/{task_id}`

读取单个 runtime task 详情。

### `GET /api/v1/runtime/tasks/stats`

读取 runtime task 聚合统计。

当前包括：

- 总任务数
- ready pull / in flight / terminal 数量
- `heartbeat_timeout` 数量
- 存在 `last_dispatch_error` 的任务数量
- 按 `status`、`category`、`consumer`、`delivery_mode`、`transport`、`interface_name` 聚合的计数
- 最老 ready pull 任务的等待秒数

### `POST /api/v1/runtime/tasks/claim`

领取一个 runtime task。

典型请求体：

```json
{
  "task_id": "task-id",
  "consumer": "game_client",
  "lease_owner": "client-1"
}
```

当前行为：

- 只有 `pending` / `released` 且已到达可领取时间的任务可以被 claim。
- claim 成功后，任务状态会变成 `claimed`，并生成 `lease_token`。
- `dispatched` 状态任务表示已经由 Engine 主动向外部系统发出，不参与 pull claim。
- 如果任务已被其他 consumer 领取，会返回 `409`，错误码 `runtime_task_not_claimable`。

### `POST /api/v1/runtime/tasks/heartbeat`

为已 claim 的 runtime task 上报心跳。

典型请求体：

```json
{
  "task_id": "task-id",
  "lease_token": "lease-token"
}
```

当前行为：

- 只有 `claimed` 或 `running` 且 `lease_token` 匹配的任务可以更新心跳。
- lease 不匹配时返回 `409`，错误码 `runtime_task_lease_mismatch`。

### `POST /api/v1/runtime/tasks/start`

将一个已 claim 的 runtime task 显式标记为开始执行。

典型请求体：

```json
{
  "task_id": "task-id",
  "lease_token": "lease-token"
}
```

当前行为：

- 只有处于 `claimed` 状态且 `lease_token` 匹配的任务可以进入 `running` 状态。
- start 成功后，任务状态会变成 `running`，并刷新 `last_heartbeat_at`。
- lease 不匹配时返回 `409`，错误码 `runtime_task_lease_mismatch`。

### `POST /api/v1/runtime/tasks/release`

释放一个已 claim 的 runtime task，并可选地延迟重新入队。

典型请求体：

```json
{
  "task_id": "task-id",
  "lease_token": "lease-token",
  "retry_delay_ms": 2500,
  "error_message": "temporary failure"
}
```

当前行为：

- release 成功后，任务状态会变成 `released`。
- `retry_delay_ms` 可控制任务何时再次出现在 pending 列表里。
- lease 不匹配时返回 `409`，错误码 `runtime_task_lease_mismatch`。

### `POST /api/v1/runtime/tasks/requeue`

将一个 `heartbeat_timeout` 状态的 runtime task 显式重新放回队列。

典型请求体：

```json
{
  "task_id": "task-id",
  "retry_delay_ms": 1500,
  "error_message": "manual requeue"
}
```

当前行为：

- 只有 `heartbeat_timeout` 状态的任务可以被 requeue。
- requeue 成功后，任务状态会变成 `released`，并清除超时标记与租约。
- `retry_delay_ms` 可控制任务何时重新进入 pending 列表。
- 非 `heartbeat_timeout` 任务会返回 `409`，错误码 `runtime_task_not_requeueable`。

### `POST /api/v1/runtime/tasks/heartbeat-timeout/sweep`

按给定超时阈值，将长时间无心跳的 `claimed` / `running` 任务批量标记为 `heartbeat_timeout`。

典型请求体：

```json
{
  "timeout_seconds": 60
}
```

当前行为：

- `timeout_seconds` 必须大于 `0`
- 成功后返回本次 sweep 影响的任务数量
- 这是最小管理入口，后续还会继续补自动治理与批量 requeue 策略

### `POST /api/v1/runtime/tasks/heartbeat-timeout/requeue`

按筛选条件批量将 `heartbeat_timeout` 任务重新放回队列。

典型请求体：

```json
{
  "consumer": "bridge",
  "category": "external_action",
  "transport": "task_pull",
  "retry_delay_ms": 500,
  "limit": 100,
  "error_message": "auto requeue"
}
```

当前行为：

- 可按 `consumer`、`category`、`transport` 过滤批量目标
- `limit` 最大 `500`，默认按最早创建的 timeout task 优先处理
- requeue 后任务会回到 `released`，并清理 timeout/lease 状态
- 这是 `heartbeat_timeout` 自动治理前的批量运维入口，后续还会继续补策略化自动回收能力

### `GET /api/v1/plans/pending`

列出处于人工审核等待中的计划。

### `POST /api/v1/worlds/{world_id}/plan/approve`

批准待审核的世界变更计划。

### `POST /api/v1/worlds/{world_id}/plan/reject`

拒绝待审核的世界变更计划。

---

## 节点

### `GET /api/v1/nodes`

列出节点。

查询参数：

- `world_id`
- `node_type`
- `limit`
- `offset`

### `GET /api/v1/nodes/{id}`

返回节点详情，包括节点本体、子节点、组件、记忆与关系。

### `POST /api/v1/nodes`

创建节点。

典型请求体：

```json
{
  "world_id": "world-id",
  "name": "村口守卫",
  "node_type": "npc",
  "parent_id": "optional-parent-id"
}
```

### `PUT /api/v1/nodes/{id}`

更新节点的可变字段，例如 `name`、`node_type`、`parent_id`。

### `POST /api/v1/nodes/{id}/copy`

在同一世界内复制节点。

请求体：

```json
{
  "name": "Copied Node",
  "parent_id": "optional-parent-id",
  "include_descendants": true
}
```

行为说明：

- 该接口不能复制世界根节点
- 默认支持整棵子树复制
- 若关系两端都在复制集合内，内部关系也会一起复制

### `DELETE /api/v1/nodes/{id}`

删除叶子节点。

### `GET /api/v1/nodes/{node_id}/autonomous`

读取节点的自主行为配置。

### `PUT /api/v1/nodes/{node_id}/autonomous`

创建或更新节点的自主行为配置。

### `POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run`

手动触发一次节点自主行为执行。

---

## 组件

- `POST /api/v1/components`
- `GET /api/v1/components`
- `GET /api/v1/components/{id}`
- `PUT /api/v1/components/{id}`
- `DELETE /api/v1/components/{id}`

组件写入接口现在会按组件类型执行数据校验：

- `autonomous`：必须是合法 JSON 对象，并满足 Engine 当前的强类型字段约束
- `profile`：必须是合法 JSON 对象，但字段集合保持开放
- 其他当前内置类型：默认按纯文本处理，除非后续为该类型补充结构化规则

当组件数据不符合对应规则时，接口会返回 `400`，错误码为 `invalid_component_data`。

---

## 记忆

- `POST /api/v1/memories`
- `GET /api/v1/memories`
- `GET /api/v1/memories/{id}`
- `PUT /api/v1/memories/{id}`
- `DELETE /api/v1/memories/{id}`
- `POST /api/v1/memories/propagate`

传播接口用于显式驱动记忆传播规则，例如 `mode`、`target_ids`、`tags`、`max_depth`、`publish_up`。

---

## 关系

- `POST /api/v1/relations`
- `GET /api/v1/relations`
- `GET /api/v1/relations/{id}`
- `PUT /api/v1/relations/{id}`
- `DELETE /api/v1/relations/{id}`

---

## 世界管理

### `GET /api/v1/worlds`

列出所有世界根节点。

### `PUT /api/v1/worlds/{world_id}`

更新世界的可变字段。

当前请求体：

```json
{
  "name": "Renamed World"
}
```

---

## 世界运行时操作

### `POST /api/v1/worlds/{world_id}/ticks/advance`

推进一个世界 Tick。

典型请求体：

```json
{
  "tick_type": "scheduled",
  "game_time": "第 3 天 - 傍晚",
  "autonomous_limit": 10
}
```

### `POST /api/v1/worlds/{world_id}/events/impact`

评估一个事件对世界的影响。

### `POST /api/v1/worlds/{world_id}/scopes/{scope_id}/advance`

推进某个非世界根节点范围的局部演化。

### `POST /api/v1/worlds/{world_id}/timeline/replan`

重建世界的未来大纲。

---

## 快照与世界复制流程

### `POST /api/v1/worlds/{world_id}/fork`

创建可运行的工作副本。

### `GET /api/v1/worlds/{world_id}/snapshots`

列出某个源世界创建过的存档快照。

### `POST /api/v1/worlds/{world_id}/snapshots`

创建一个存档快照世界。

### `DELETE /api/v1/worlds/{world_id}/snapshot`

删除存档快照世界及其元数据。

### `GET /api/v1/worlds/{world_id}/snapshot-metadata`

读取某个复制世界对应的快照元数据。

### `GET /api/v1/worlds/{world_id}/snapshot-validation`

校验一个快照当前是否仍可恢复。

### `POST /api/v1/worlds/{world_id}/restore`

将存档快照恢复为新的可运行世界。

上述复制类接口都接受如下可选请求体：

```json
{
  "name": "Optional Name",
  "lock_world": true
}
```

说明：

- `lock_world` 仍然保留为兼容字段
- 当前实现已经默认对同世界关键复制/恢复操作启用业务级互斥
- 这个字段更适合作为客户端语义声明，而不是绕过安全边界的开关

---

## 管线观测

### `GET /api/v1/pipeline/stats`

返回共享数据库/推理管线的轻量统计信息。

当前包括：

- 写重试 attempts / retries / recoveries / failures
- 事务次数与累计耗时
- log sink 队列深度、flush 次数、fallback 写入次数
- world lock 获取与争用统计

---

## 世界设置与世界策略

### `GET /api/v1/worlds/{world_id}/settings`
### `PUT /api/v1/worlds/{world_id}/settings`

获取或部分更新世界运行时设置。

关键字段：

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

### `GET /api/v1/worlds/{world_id}/policy`
### `PUT /api/v1/worlds/{world_id}/policy`

获取或更新世界动作策略。

---

## 连续性状态与时间线

### `GET /api/v1/worlds/{world_id}/state-components`

读取一个世界当前全部由 Engine 识别的 world tick 连续性状态组件。

当前组件类型包括：

- `world_state`
- `story_state`
- `story_history`
- `tick_policy`
- `state_snapshot`

### `GET /api/v1/worlds/{world_id}/state-components/{component_type}`

读取某一个连续性状态组件。

如果 `component_type` 不属于当前受支持的连续性组件类型，接口会返回 `400`，错误码为 `invalid_component_type`。

### `PUT /api/v1/worlds/{world_id}/state-components/{component_type}`

创建或整体替换某一个连续性状态组件。

请求体必须是合法 JSON。像 `world_state`、`story_state`、`story_history`、`tick_policy` 这类结构化组件会按 Engine 当前的字段形状规则做校验。

### `GET /api/v1/worlds/{world_id}/timelines`

读取最近持久化的 world tick 时间线归档。

查询参数：

- `limit` - 可选，默认 `20`，最大 `200`

### `GET /api/v1/worlds/{world_id}/timelines/latest`

读取最近一条持久化的 world tick 时间线归档。

---

## Creator 导入

### `POST /api/v1/creator/import`

从 YAML 或 JSON 文本导入世界配置。

典型请求体：

```json
{
  "format": "yaml",
  "content": "...",
  "reset": false,
  "dry_run": false
}
```

---

## 日志与调试

### `GET /api/v1/logs`

读取推理日志。

查询参数：

- `world_id`
- `task_type`
- `node_id`
- `category`
- `event_name`
- `execution_mode`
- `request_id`
- `round`
- `limit`
- `offset`

### `GET /debug/traces`

读取调试轨迹。

---

## 响应元数据

推理响应可能包含：

- `configured_pipeline_mode`
- `effective_pipeline_mode`
- `max_analysis_rounds`
- `rounds_used`

当不同游戏选择不同 `pipeline_mode` 档位时，这些字段尤其有用。

---

## 错误模型

常见状态码：

- `400` 请求非法
- `404` 资源不存在
- `409` 资源冲突
- `500` 内部错误

错误响应格式：

```json
{
  "error": "message",
  "code": "machine_code"
}
```
