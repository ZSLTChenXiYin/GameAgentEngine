# 推理管线内部实现

**中文** | [**English**](./PIPELINE_INTERNALS_EN.md)

本文档说明 GameAgentEngine 当前推理管线的核心结构，包括 PipelineMode、多轮循环、子任务 DAG、数据请求、动作执行、记忆传播，以及最近引入的数据库管线与观测能力。

---

## 管线模式

每个世界可以独立配置 `pipeline_mode`，当前支持：

| 模式 | 值 | 行为 |
|---|---|---|
| 垂直模式 | `vertical` | 单轮推理，尽量减少分支和补充轮次 |
| 轮询模式 | `polling` | 允许多轮补充上下文与数据请求 |
| 完整模式 | `full` | 启用多轮推理、子任务 DAG、更多编排能力 |

`pipeline_mode` 与 `execution_mode` 独立：

- `pipeline_mode` 决定“怎么推理”
- `execution_mode` 决定“结果如何落地和是否进入审核”

---

## Pipeline.Execute 主流程

统一入口会依次执行：

1. 读取世界设置与世界策略。
2. 构建基础上下文。
3. 按任务类型分发到专用 prompt / finalize 路径。
4. 进入公共多轮执行循环。
5. 执行动作、写入记忆、触发传播。
6. 记录结构化日志并返回响应。

当前主要任务类型包括：

- `npc_dialogue`
- `world_tick`
- `world_event`
- `autonomous_act`
- `custom`

---

## 公共多轮循环

多轮循环是所有复杂任务的统一骨架：

1. 生成系统提示词。
2. 调用 LLM。
3. 解析 JSON 响应。
4. 处理 `request_data`。
5. 处理 `sub_tasks`。
6. 执行动作。
7. 写入记忆与传播。
8. 写日志并判断是否结束。

在 `debug` 和 `review` 模式下，日志会保留更多 request / response / detail 载荷；在 `production` 下仍保留结构化摘要，但载荷更轻。

---

## 子任务 DAG

在 `full` 模式下，模型可以声明 `sub_tasks`，系统会将其注册为 DAG 节点并按依赖关系调度。

主要能力：

- 注册子任务
- 解析依赖
- 执行就绪任务
- 子任务失败后的重试与超时控制
- 合并子任务结果

合并模式目前包括：

- `append`
- `override`
- `summarize`

---

## 数据请求循环

模型可以在一轮响应中发出 `request_data` 查询：

1. 管线解析查询列表。
2. 对 `target="store"` 的请求直接读数据库。
3. 对 `target="game_client"` 的请求会暂停当前多轮执行，并生成持久化的 `callback_id` 与 paused execution 快照。
4. 外部通过 `POST /api/v1/actions/callback` 回填结果后，Engine 会自动恢复原始执行现场。
5. 回填结果会作为补充上下文注入后续轮次，再继续下一轮 LLM 推理。

这样可以把“先问、再查、再推理”的流程留在统一循环里，而不是把业务逻辑散落到调用方。

当前恢复链路会持久化以下信息，避免上下文压缩、服务重启或调用方丢失中间态时无法续跑：

- 原始 `InvokeRequest`
- `BuiltContext` 快照
- 当前 round state 与补充上下文
- 暂停时的 `request_data`
- `callback_id` 与恢复载荷

---

## 动作执行

动作执行遵循统一规则：

1. 解析 `action_calls`。
2. 校验能力与 schema。
3. 同步动作立即执行。
4. 异步动作返回 `callback_id`。
5. 调用方通过 `POST /api/v1/actions/callback` 回填结果。
6. 对普通异步动作，回调结果会更新持久化 callback 记录、触发 `OnResult(...)`，并按 runtime task payload 中固化的 `callback_post_process` 策略执行统一后处理。
7. 对暂停中的 `game_client request_data`，回调会自动恢复原执行并继续后续轮次。

当前统一后处理基础版已经支持：

- `record_only`：显式只记录 callback / task 结果
- `write_memory`：将 callback 结果按模板渲染后写入目标节点记忆

这里刻意使用 runtime task payload 中的策略快照，而不是在 callback 时重新读取最新配置。这样即使服务重启、发生上下文压缩，或配置在任务发出后被修改，callback 也仍然会按任务创建时的原始策略执行。

对 `autonomous_act`，能力白名单是硬约束，不是提示性建议。

---

## 记忆与传播

模型写出的 `memory_updates` 会进入统一写路径，并触发可选传播：

- `upward`
- `tag_broadcast`
- `targeted`
- `manual`

系统还支持可选传播状态机：

- `enable_propagation_machine`
- 规则链检查
- TransformRule
- 级联传播动作

最近的数据库管线优化已经把直接记忆写入与部分传播插入改成批处理，降低了 SQLite 上碎片化写事务的数量。

---

## world_tick 连续性工件

`world_tick` 不只是一次普通推理，它还会额外持久化连续性状态：

- `world_state`
- `story_state`
- `story_history`
- `world_time_state`
- `state_snapshot`
- 时间线归档行

这些工件会在后续 Tick 或调试流程中重新进入上下文，形成连续性闭环。

---

## 数据库管线集成

推理管线现在已经和统一数据库读写管线对齐：

### SQLite

- WAL 模式
- busy timeout
- 单写连接
- 并发读
- 日志批量写入
- 记忆/传播批量写入
- 同世界关键重操作互斥

### MySQL / PostgreSQL

- 并发连接池写入
- 共享事务入口
- 统一 migration runner
- 统一可重试写层

这意味着现在的瓶颈排查不再只看业务逻辑，还可以直接观察写重试、队列和锁统计。

---

## 可观测性

新增的只读入口：

- `GET /api/v1/pipeline/stats`

当前可观察的核心指标包括：

- 写重试 attempts / retries / recoveries / failures
- 事务次数与累计耗时
- log sink 队列深度、flush 次数、fallback 次数
- world lock 获取次数、争用次数、活动持有数

如果需要回放某次推理，还可以结合：

- `GET /api/v1/logs`
- `GET /debug/traces`

---

## 配置与降级开关

当前与管线稳定性强相关的静态配置包括：

```yaml
database:
  driver: "sqlite"                # sqlite / mysql / postgres
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

engine:
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

这些开关的目的不是长期关闭特性，而是在压测、灰度或问题定位时，可以快速退回更保守的运行方式，而不需要回滚业务代码。
