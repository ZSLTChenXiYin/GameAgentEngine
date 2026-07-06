# API 参考

**中文** | [**English**](./API_REFERENCE_EN.md)

GameAgentEngine v0.2.0 提供 RESTful HTTP API。所有 API 通过 `X-API-Key` 请求头鉴权。

---

## 基础路径

所有接口以 `/api/v1/` 为前缀。服务默认监听 `0.0.0.0:8080`。

---

## 验证与健康检查

### GET /health

服务健康检查。

```bash
curl http://127.0.0.1:8080/health
# {"status":"ok"}
```

### GET /api/v1/status

服务状态及世界列表。

```bash
curl -H "X-API-Key: dev-key" http://127.0.0.1:8080/api/v1/status
```

---

## 推理

### POST /api/v1/invoke

统一推理入口。支持所有任务类型。

```json
// 请求
{
  "world_id": "世界ID",
  "node_id": "目标节点ID",
  "task_type": "npc_dialogue",
  "context": {
    "messages": [{"role": "user", "content": "你好"}]
  }
}
```

```json
// 响应
{
  "request_id": "uuid",
  "task_type": "npc_dialogue",
  "reply": "角色回应...",
  "action_calls": [],
  "memory_updates": [],
  "world_change_plan": {},
  "metadata": {
    "llm_model": "deepseek-chat",
    "processing_time_ms": 3980,
    "configured_pipeline_mode": "polling",
    "effective_pipeline_mode": "vertical",
    "max_analysis_rounds": 4,
    "rounds_used": 1
  }
}
```

### POST /api/v1/actions/callback

异步动作回调接口。

```json
{
  "callback_id": "回调ID",
  "status": "completed",
  "result": {}
}
```

---

## 节点

### POST /api/v1/nodes

创建节点。

```json
{
  "world_id": "世界ID",
  "name": "节点名称",
  "node_type": "npc",
  "parent_id": "父节点ID（可选）"
}
```

### GET /api/v1/nodes/{id}

获取节点详情（含组件、记忆、关系）。

### PUT /api/v1/nodes/{id}

更新节点。

### DELETE /api/v1/nodes/{id}

删除节点（仅叶子节点）。

### GET /api/v1/nodes

列出节点，支持分页和过滤。

| 参数 | 说明 |
|---|---|
| `world_id` | 按世界过滤 |
| `node_type` | 按类型过滤 |
| `limit` | 返回条数上限 |
| `offset` | 偏移量 |

---

## 组件

### POST /api/v1/components

创建组件。

```json
{
  "node_id": "节点ID",
  "component_type": "profile",
  "data": "{\"name\":\"艾琳\"}"
}
```

### GET /api/v1/components

获取节点组件列表。参数：`node_id`。

### GET /api/v1/components/{id}

获取单个组件。

### PUT /api/v1/components/{id}

更新组件。

### DELETE /api/v1/components/{id}

删除组件。

---

## 记忆

### POST /api/v1/memories

创建记忆。

```json
{
  "node_id": "节点ID",
  "content": "记忆内容",
  "level": "long_term",
  "tags": "tag1,tag2"
}
```

### GET /api/v1/memories

获取节点记忆列表。参数：`node_id`。

### GET /api/v1/memories/{id}

获取单条记忆。

### PUT /api/v1/memories/{id}

更新记忆。

### DELETE /api/v1/memories/{id}

删除记忆。

### POST /api/v1/memories/propagate

手动传播记忆。

```json
{
  "memory_id": "记忆ID",
  "target_node": "目标节点ID",
  "mode": "upward",
  "tags": [],
  "target_ids": [],
  "max_depth": 0,
  "publish_up": false
}
```

---

## 关系

### POST /api/v1/relations

创建关系。

```json
{
  "world_id": "世界ID",
  "source_id": "源节点ID",
  "target_id": "目标节点ID",
  "relation_type": "ally",
  "weight": 50,
  "properties": "{}"
}
```

### GET /api/v1/relations

列出关系。参数：`world_id`、`limit`、`offset`。

### GET /api/v1/relations/{id}

获取单条关系。

### PUT /api/v1/relations/{id}

更新关系。

### DELETE /api/v1/relations/{id}

删除关系。

---

## 世界操作

### POST /api/v1/worlds/{world_id}/ticks/advance

推进世界时间（Tick）。

```json
{
  "tick_type": "scheduled",
  "game_time": "第2天-中午",
  "autonomous_limit": 10
}
```

响应包含 tick 记录、推理响应和自主行为运行结果。

### POST /api/v1/worlds/{world_id}/events/impact

评估事件影响。

```json
{
  "event_type": "diplomatic_crisis",
  "scope_id": "作用范围节点ID",
  "description": "事件描述",
  "severity": "critical"
}
```

### POST /api/v1/worlds/{world_id}/timeline/replan

重新生成世界未来大纲。

### POST /api/v1/worlds/{world_id}/scopes/{scope_id}/advance

局部范围推进演化。

### POST /api/v1/worlds/{world_id}/fork

创建世界工作副本。请求体可选 `name` 和 `lock_world`；当 `lock_world` 为 `true` 时，复制期间锁定源世界。

### POST /api/v1/worlds/{world_id}/snapshots

创建世界存档快照。请求体可选 `name` 和 `lock_world`；当 `lock_world` 为 `true` 时，存档期间锁定源世界。

### POST /api/v1/worlds/{world_id}/restore

从存档快照恢复一个新世界。`world_id` 应为快照世界 ID。请求体可选 `name` 和 `lock_world`；当 `lock_world` 为 `true` 时，恢复期间锁定源快照世界。

### DELETE /api/v1/worlds/{world_id}/snapshot

删除某个存档快照世界，以及与之关联的快照元数据。`world_id` 必须是 `save_snapshot` 类型的快照世界 ID。

---

## 世界设置

### GET /api/v1/worlds/{world_id}/settings

获取世界运行设置。

```json
{
  "world_id": "...",
  "memory_limit": 50,
  "max_analysis_rounds": 5,
  "max_context_depth": 3,
  "auto_apply": true,
  "require_review_above": "critical",
  "pipeline_mode": "full",
  "propagation_max_depth": 2,
  "sub_task_max_retries": 2,
  "sub_task_timeout_secs": 60,
  "enable_propagation_machine": false
}
```

### PUT /api/v1/worlds/{world_id}/settings

更新世界运行设置。支持部分更新（只传需要修改的字段）。

```json
{
  "pipeline_mode": "polling",
  "propagation_max_depth": 0,
  "sub_task_max_retries": 0,
  "sub_task_timeout_secs": 0
}
```

说明：

- `memory_limit`、`max_analysis_rounds`、`max_context_depth` 在传入时必须大于 `0`。
- `propagation_max_depth`、`sub_task_max_retries`、`sub_task_timeout_secs` 支持显式设置为 `0`。
- `pipeline_mode` 只能是 `vertical`、`polling`、`full` 之一。

### GET /api/v1/worlds/{world_id}/snapshot-validation

校验某个存档快照当前是否仍可安全恢复。返回结构化校验结果，包含版本兼容性、组件类型漂移、世界设置漂移、策略漂移等信息。

```json
{
  "snapshot_world_id": "snapshot-world-id",
  "source_world_id": "source-world-id",
  "snapshot_name": "Save Slot 1",
  "reason": "save_snapshot",
  "valid": true,
  "schema_version": "world_snapshot/v1",
  "engine_version": "v0.4.3",
  "min_compatible_version": "v0.4.3",
  "current_engine_version": "v0.4.3",
  "current_min_compatible_version": "v0.4.3",
  "saved_component_types": ["profile", "rule"],
  "current_component_types": ["profile", "rule"],
  "issues": []
}
```

### GET /api/v1/worlds/{world_id}/snapshot-metadata

查询某个快照世界的存档元数据。

### GET /api/v1/worlds/{world_id}/snapshots

列出某个源世界创建过的全部存档快照，仅返回 `save_snapshot` 类型。

### Invoke Response Metadata

`metadata` 除了模型与耗时字段外，还可能包含以下管线可观测性字段：

- `configured_pipeline_mode`：世界设置中的默认管线模式
- `effective_pipeline_mode`：本次请求实际生效的管线模式
- `max_analysis_rounds`：本次请求解析后的最大轮次数
- `rounds_used`：本次请求实际消耗的轮次数

---

## 世界策略

### GET /api/v1/worlds/{world_id}/policy

获取世界动作策略。

### PUT /api/v1/worlds/{world_id}/policy

更新世界动作策略。

```json
{
  "blocked_actions": ["kill_character"],
  "safe_actions": ["add_memory"]
}
```

---

## 自主行为

### GET /api/v1/nodes/{node_id}/autonomous

获取节点自主行为配置。

### PUT /api/v1/nodes/{node_id}/autonomous

设置节点自主行为配置。

### POST /api/v1/nodes/{node_id}/autonomous/run

手动触发节点自主行为。

### POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run

（兼容路径）手动触发自主行为。

---

## 日志

### GET /api/v1/logs

读取推理日志。

| 参数 | 说明 |
|---|---|
| `world_id` | 按世界过滤 |
| `limit` | 返回条数上限（默认 10） |

---

## 导入

### POST /api/v1/creator/import

导入世界配置（YAML 或 JSON）。

```json
{
  "format": "yaml",
  "content": "YAML或JSON字符串",
  "reset": false,
  "dry_run": false
}
```

---

## 错误码

| HTTP 状态 | 说明 |
|---|---|
| 400 | 请求参数无效 |
| 404 | 资源不存在 |
| 409 | 冲突（如删除非叶子节点） |
| 500 | 服务器内部错误 |

错误响应格式：

```json
{
  "error": "人类可读的错误信息",
  "code": "机器可读的错误码"
}
```
