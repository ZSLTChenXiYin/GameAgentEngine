# API 参考

**中文** | [**English**](./API_REFERENCE_EN.md)

除特别说明外，所有引擎 API 都通过 `X-API-Key` 请求头鉴权。

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

读取推理日志，支持 `world_id`、`task_type`、`limit`、`offset` 查询参数。

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
