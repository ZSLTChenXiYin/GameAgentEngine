# 自主行为系统

**中文** | [**English**](./AUTONOMOUS_BEHAVIOR_EN.md)

GameAgentEngine 的自主行为系统允许节点在没有直接用户输入时，按照节点本身的配置触发一次独立推理与动作执行循环。它主要面向 NPC、组织、设施、世界范围控制节点等“可以主动行动”的实体。

---

## 核心概念

自主行为通过挂载在节点上的 `autonomous` 组件配置。

每个节点可以独立声明：

- 是否启用自主行为
- 触发方式
- 允许调用的能力白名单
- 调度间隔与上次运行时间

当前支持的触发方式：

| 触发方式 | 值 | 说明 |
|---|---|---|
| 手动触发 | `manual` | 仅通过 API 或 DevCli 显式触发 |
| Tick 同步 | `world_tick_sync` | 在世界 Tick 推进后同步触发 |
| 定时调度 | `scheduled` | 由后台调度器按扫描周期触发 |

---

## 执行路径

一次自主行为执行通常包含以下阶段：

1. 读取节点的 `autonomous` 配置与能力白名单。
2. 根据任务类型 `autonomous_act` 构建上下文。
3. 进入统一 Pipeline 推理循环。
4. 校验 LLM 输出动作是否落在能力白名单内。
5. 执行同步动作、登记异步动作、写入记忆与传播。
6. 记录结构化日志到统一 `logs` 表。

如果该节点属于某个世界，执行会经过世界级业务互斥边界，以避免同一世界的重操作互相踩踏。

---

## 能力白名单

每个自主节点都可以声明一组允许调用的动作能力。引擎会校验 LLM 输出动作是否落在这组能力内，阻止节点越权执行未授权动作。

示例：

```json
{
  "enabled": true,
  "trigger": "world_tick_sync",
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "记录一次判断",
      "schema": {
        "node_id": { "type": "string", "required": true },
        "content": { "type": "string", "required": true }
      }
    }
  ]
}
```

能力校验失败时，请求不会静默放过；动作会被拒绝，并记录到日志中用于诊断。

---

## 配置方式

### DevCli

```bash
# 读取当前配置
GameAgentDevCli node autonomous get <node-id>

# 启用并配置为 world tick 同步触发
GameAgentDevCli node autonomous set <node-id> --enabled --trigger "world_tick_sync"

# 禁用自主行为
GameAgentDevCli node autonomous disable <node-id>

# 手动触发一次自主行为
GameAgentDevCli node autonomous run <node-id>
```

### API

相关接口：

- `GET /api/v1/nodes/{node_id}/autonomous`
- `PUT /api/v1/nodes/{node_id}/autonomous`
- `POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run`

示例请求体：

```json
{
  "enabled": true,
  "trigger": "scheduled",
  "interval_seconds": 600,
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "记录一次观察"
    }
  ]
}
```

---

## 调度器

后台自主调度器是服务级静态开关，由配置文件控制：

```yaml
engine:
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

调度器只会尝试触发同时满足以下条件的节点：

1. 节点挂载了 `autonomous` 组件。
2. `enabled = true`。
3. `trigger = "scheduled"`。
4. 距离上次运行已超过 `interval_seconds`。

同一个世界里的 scheduled 自主行为仍会通过世界级互斥边界串行化关键重操作，从而减少 SQLite 锁竞争并保持业务一致性。

---

## 与数据库管线的关系

当前自主行为相关写路径已经接入统一数据库管线：

- SQLite：单写者、WAL、批量日志、批量记忆写入
- MySQL / PostgreSQL：并发池化写入、统一重试层
- 关键同世界操作：世界级业务锁保护

因此，自主行为的优化不再只是“少跑一点”，而是会直接受益于统一的写事务、批量写和重试恢复层。

---

## 诊断建议

排查自主行为问题时，优先看这些入口：

- `GET /api/v1/logs`
- `GET /api/v1/pipeline/stats`
- `GET /debug/traces`

特别适合关注的字段：

- `task_type = autonomous_act`
- `event_name = autonomous_node_started|autonomous_node_completed|autonomous_node_failed`
- `request_id`
- `round`
- `execution_mode`

如果怀疑调度拥塞或锁竞争，可以同时观察 pipeline stats 中的：

- 写重试计数
- 事务计数
- log sink 队列深度
- world lock 争用统计
