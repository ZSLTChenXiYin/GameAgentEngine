const fs = require('fs');
const path = require('path');

function repoPath(...parts) {
  return path.join(process.cwd(), ...parts);
}

function block(fn) {
  const match = fn.toString().match(/\/\*([\s\S]*?)\*\//);
  if (!match) {
    throw new Error('block() requires an inline comment body');
  }
  return match[1].replace(/^\n/, '');
}

function normalize(content) {
  return content.replace(/\r?\n/g, '\n').replace(/\n{3,}/g, '\n\n').trimEnd() + '\n';
}

function write(relPath, content) {
  const file = repoPath(relPath);
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, normalize(content), 'utf8');
}

function read(relPath) {
  return fs.readFileSync(repoPath(relPath), 'utf8');
}

function replaceOrThrow(content, searchValue, replacement, label) {
  if (!content.includes(searchValue)) {
    throw new Error(`missing expected block: ${label}`);
  }
  return content.replace(searchValue, replacement);
}

function replacePatternOrThrow(content, pattern, replacement, label) {
  if (!pattern.test(content)) {
    throw new Error(`missing expected pattern: ${label}`);
  }
  return content.replace(pattern, replacement);
}

const configTemplate = block(function () {/*
# ========================================
#  GameAgentEngine 配置文件
# ========================================
# 搜索路径：当前目录 ./gameagentengine.conf.yaml
#          ./config/gameagentengine.conf.yaml
# 可通过 --config <路径> 显式指定

server:
  # HTTP 服务监听地址。本机测试用 127.0.0.1，
  # 开放局域网访问用 0.0.0.0
  host: "0.0.0.0"
  # 服务端口号
  port: 8080

database:
  # 数据库驱动：sqlite / mysql / postgres
  driver: "sqlite"
  # SQLite：文件名；MySQL/PostgreSQL：连接串 / DSN
  dsn: "gameagentengine.db"
  # 是否在初始化时执行 schema/data migrations
  migrations_enabled: true
  # 是否启用短暂写冲突自动重试
  write_retry_enabled: true
  # 写冲突自动重试最大尝试次数
  write_retry_max_attempts: 3
  # 首次重试回退延迟（毫秒）
  write_retry_base_delay_ms: 40
  # 重试最大回退延迟（毫秒）
  write_retry_max_delay_ms: 250
  # 是否启用推理日志批量写入
  log_batch_enabled: true
  # 单次批量落库的最大日志条数
  log_batch_size: 32
  # 日志批量 flush 间隔（毫秒）
  log_batch_flush_ms: 750
  # 日志批处理内存队列大小
  log_batch_queue_size: 1024

auth:
  # API 鉴权密钥，客户端请求时需要附带 X-API-Key 请求头
  # 开发环境默认值 dev-key，生产环境务必修改
  api_key: "dev-key"

llm:
  # LLM 提供商：openai（兼容 OpenAI API 格式的供应商也可）
  provider: "openai"
  # 模型名称，按所接入的平台填写
  model: "deepseek-v4-flash"
  # API Key，留空则使用 mock Provider（仅用于离线调试）
  api_key: "sk-xxx"
  # API 端点地址
  base_url: "https://api.deepseek.com"

engine:
  # 执行模式：debug / review / production / full
  execution_mode: "debug"
  # 是否启用世界级重操作互斥锁
  world_lock_enabled: true
  # 后台自主行为调度器是服务级静态开关，默认模板保持关闭
  autonomous_scheduler_enabled: false
  # 后台调度器扫描间隔秒数
  autonomous_scheduler_interval_seconds: 300
  # 每次扫描每个世界最多触发的 scheduled 自主节点数
  autonomous_scheduler_max_nodes_per_world: 10
*/});

const autonomousZh = block(function () {/*
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
*/});

const autonomousEn = block(function () {/*
# Autonomous Behavior System

[**中文**](./AUTONOMOUS_BEHAVIOR.md) | **English**

The autonomous behavior system allows a node to trigger its own reasoning and action loop without direct user input. It is intended for active entities such as NPCs, organizations, facilities, and world-scope controller nodes.

---

## Core Concepts

Autonomous behavior is configured through the `autonomous` component attached to a node.

Each node can independently define:

- whether autonomous behavior is enabled
- the trigger mode
- a capability allowlist
- its scheduling interval and last-run state

Current trigger modes:

| Trigger Mode | Value | Description |
|---|---|---|
| Manual | `manual` | Triggered only through API or DevCli |
| Tick Sync | `world_tick_sync` | Triggered after world tick advancement |
| Scheduled | `scheduled` | Triggered by the background scheduler |

---

## Execution Path

A typical autonomous execution goes through these stages:

1. Load the node's `autonomous` config and capability allowlist.
2. Build the `autonomous_act` task context.
3. Enter the shared pipeline reasoning loop.
4. Validate that emitted actions stay within the allowlist.
5. Execute sync actions, register async actions, and persist memory / propagation effects.
6. Persist structured logs into the unified `logs` table.

If the node belongs to a world, the flow also passes through the same-world exclusion boundary for critical heavy operations.

---

## Capability Allowlist

Each autonomous node can declare the actions it is allowed to invoke. The engine validates LLM output against this allowlist and rejects unauthorized actions.

Example:

```json
{
  "enabled": true,
  "trigger": "world_tick_sync",
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "Record a judgment",
      "schema": {
        "node_id": { "type": "string", "required": true },
        "content": { "type": "string", "required": true }
      }
    }
  ]
}
```

Validation failures are not silently ignored. The action is rejected and recorded for diagnosis.

---

## Configuration

### DevCli

```bash
# Read current config
GameAgentDevCli node autonomous get <node-id>

# Enable and switch to world tick sync mode
GameAgentDevCli node autonomous set <node-id> --enabled --trigger "world_tick_sync"

# Disable autonomous behavior
GameAgentDevCli node autonomous disable <node-id>

# Trigger one autonomous execution manually
GameAgentDevCli node autonomous run <node-id>
```

### API

Related endpoints:

- `GET /api/v1/nodes/{node_id}/autonomous`
- `PUT /api/v1/nodes/{node_id}/autonomous`
- `POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run`

Example request body:

```json
{
  "enabled": true,
  "trigger": "scheduled",
  "interval_seconds": 600,
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "Record an observation"
    }
  ]
}
```

---

## Scheduler

The background autonomous scheduler is a service-level static toggle controlled by config:

```yaml
engine:
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

The scheduler only attempts nodes that satisfy all of the following:

1. The node has an `autonomous` component.
2. `enabled = true`.
3. `trigger = "scheduled"`.
4. The elapsed time since the previous run exceeds `interval_seconds`.

Scheduled autonomous work in the same world still passes through the same-world exclusion boundary for critical heavy operations, which helps reduce SQLite contention while preserving business consistency.

---

## Relationship to the Database Pipeline

Autonomous write paths now use the shared database pipeline:

- SQLite: single writer, WAL, batched logs, batched memory writes
- MySQL / PostgreSQL: pooled concurrent writes with the shared retry layer
- critical same-world operations: protected by the world-level business lock

Because of this, autonomous behavior no longer benefits only from lower frequency. It directly benefits from shared transactions, batching, and retry recovery.

---

## Diagnostics

When investigating autonomous behavior issues, start with:

- `GET /api/v1/logs`
- `GET /api/v1/pipeline/stats`
- `GET /debug/traces`

Useful fields to watch:

- `task_type = autonomous_act`
- `event_name = autonomous_node_started|autonomous_node_completed|autonomous_node_failed`
- `request_id`
- `round`
- `execution_mode`

If you suspect scheduler congestion or lock contention, also inspect pipeline stats for:

- write retry counters
- transaction counters
- log sink queue depth
- world-lock contention stats
*/});

const pipelineZh = block(function () {/*
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
3. 对 `target="game_client"` 的请求交给回调机制。
4. 将结果拼接回下一轮上下文。

这样可以把“先问、再查、再推理”的流程留在统一循环里，而不是把业务逻辑散落到调用方。

---

## 动作执行

动作执行遵循统一规则：

1. 解析 `action_calls`。
2. 校验能力与 schema。
3. 同步动作立即执行。
4. 异步动作返回 `callback_id`。
5. 调用方通过 `POST /api/v1/actions/callback` 回填结果。

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
*/});

const pipelineEn = block(function () {/*
# Inference Pipeline Internals

[**中文**](./PIPELINE_INTERNALS.md) | **English**

This document describes the current GameAgentEngine inference pipeline, including PipelineMode, multi-round execution, sub-task DAG orchestration, data-request loops, action execution, memory propagation, and the recently added database-pipeline and observability layers.

---

## Pipeline Modes

Each world can independently configure `pipeline_mode`:

| Mode | Value | Behavior |
|---|---|---|
| Vertical | `vertical` | Single-pass inference with minimal branching |
| Polling | `polling` | Multi-round inference with follow-up data requests |
| Full | `full` | Multi-round inference plus sub-task DAG orchestration |

`pipeline_mode` and `execution_mode` are independent:

- `pipeline_mode` controls how reasoning is performed
- `execution_mode` controls how outcomes are applied or reviewed

---

## Pipeline.Execute Flow

The unified execution entrypoint now performs the following stages:

1. Load world settings and world policy.
2. Build base context.
3. Dispatch by task type.
4. Enter the shared multi-round loop.
5. Execute actions, persist memories, and trigger propagation.
6. Persist structured logs and return the response.

Main task types currently include:

- `npc_dialogue`
- `world_tick`
- `world_event`
- `autonomous_act`
- `custom`

---

## Shared Multi-round Loop

The common loop drives all complex tasks:

1. Build the system prompt.
2. Call the LLM.
3. Parse the JSON response.
4. Process `request_data`.
5. Process `sub_tasks`.
6. Execute actions.
7. Persist memories and propagation side effects.
8. Log the result and decide whether another round is needed.

In `debug` and `review`, logs retain richer request / response / detail payloads. In `production`, the engine still writes structured summaries with a smaller footprint.

---

## Sub-task DAG

In `full` mode, the model can declare `sub_tasks`, which are registered as DAG nodes and scheduled according to dependency order.

Key responsibilities:

- sub-task registration
- dependency resolution
- ready-task execution
- retry and timeout handling
- result merging

Current merge modes:

- `append`
- `override`
- `summarize`

---

## Data Request Loop

The model can emit `request_data` queries inside a reasoning round:

1. The pipeline parses the query list.
2. `target="store"` queries are resolved directly against the database.
3. `target="game_client"` queries are delegated through the callback mechanism.
4. Resolved data is injected into the next round of context.

This keeps the “ask, fetch, continue reasoning” pattern inside the shared pipeline instead of scattering it across callers.

---

## Action Execution

Action execution follows a shared rule set:

1. Parse `action_calls`.
2. Validate capability and schema constraints.
3. Execute sync actions immediately.
4. Return `callback_id` for async actions.
5. Accept async completion via `POST /api/v1/actions/callback`.

For `autonomous_act`, the capability allowlist is an enforced constraint, not a soft suggestion.

---

## Memory and Propagation

`memory_updates` flow through the unified write path and may trigger propagation:

- `upward`
- `tag_broadcast`
- `targeted`
- `manual`

The optional propagation machine can add:

- rule-chain checks
- transform rules
- chained propagation actions

Recent database-pipeline work moved direct memory writes and part of propagation persistence into batched write paths, which reduces fragmented transactions on SQLite.

---

## world_tick Continuity Artifacts

`world_tick` does more than a normal inference call. It also persists continuity artifacts such as:

- `world_state`
- `story_state`
- `story_history`
- `world_time_state`
- `state_snapshot`
- timeline archive rows

These artifacts can be injected back into later prompts, creating a continuity loop across ticks.

---

## Database Pipeline Integration

The inference pipeline is now aligned with the unified database pipeline.

### SQLite

- WAL mode
- busy timeout
- single writer connection
- concurrent reads
- batched log writes
- batched memory / propagation writes
- same-world exclusion for critical heavy operations

### MySQL / PostgreSQL

- pooled concurrent writes
- shared transaction entrypoints
- shared migration runner
- shared retriable-write layer

As a result, bottleneck analysis is no longer limited to business logic. You can now directly observe retry, queue, and lock behavior.

---

## Observability

New read-only endpoint:

- `GET /api/v1/pipeline/stats`

Current observable indicators include:

- write retry attempts / retries / recoveries / failures
- transaction count and accumulated duration
- log sink queue depth, flush count, fallback writes
- world-lock acquisitions, contention count, active holders

For replay and detailed debugging, combine this with:

- `GET /api/v1/logs`
- `GET /debug/traces`

---

## Config and Fallback Controls

Static config fields most relevant to pipeline stability now include:

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

These flags are not meant for permanent degradation. They exist so staged rollouts, stress testing, and incident mitigation can temporarily fall back to safer behavior without reverting business-layer code.
*/});

const configZh = block(function () {/*
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
*/});

const configEn = block(function () {/*
# Configuration

[**中文**](./CONFIGURATION.md) | **English**

GameAgentEngine uses a two-layer configuration model:

- static config: `gameagentengine.conf.yaml`
- dynamic world config: `world_settings`, `world_policy`, and state components stored in the database

---

## Two Kinds of Defaults

The repository currently exposes two different sources of “defaults”:

- code-level defaults registered in `internal/config/config.go`
- packaged template values defined in `tools/source/gameagentengine.conf.yaml`

For example, the code-level default for `engine.autonomous_scheduler_enabled` is currently `false`, while an older template revision still showed `true`. This documentation pass also updates the packaged template so both now point to the disabled state.

---

## Static Config File

The default template lives at `tools/source/gameagentengine.conf.yaml`.

Lookup order:

1. `--config <path>`
2. `./gameagentengine.conf.yaml`
3. `./config/gameagentengine.conf.yaml`

Current recommended template:

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
```

The code-level fallback defaults also include:

```yaml
llm:
  provider: "openai"
  model: "gpt-4o-mini"
  base_url: "https://api.openai.com/v1"

engine:
  execution_mode: "full"
```

This means that if you omit these fields from a config file, the engine still falls back to the internal defaults.

`database.driver` currently supports `sqlite`, `mysql`, and `postgres`.

---

## Key Static Toggles

### `database.migrations_enabled`

Controls whether schema / data migrations run during initialization.

### `database.write_retry_enabled`

Controls the shared retriable-write layer. When disabled, transient write conflicts surface directly to callers.

### `database.log_batch_enabled`

Controls whether inference logs are buffered and flushed in batches. When disabled, log writes fall back to direct persistence.

### `engine.world_lock_enabled`

Controls the business-level same-world exclusion boundary for critical heavy operations.

### `engine.autonomous_scheduler_enabled`

Controls the service-level autonomous scheduler. The recommended default is currently `false`.

---

## Dynamic Config: world_settings

`world_settings` is stored per world and commonly includes:

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

These settings affect one world's runtime behavior without changing other worlds.

---

## World Time Config

`world_time_settings` defines the rules, while `world_time_state` stores the resulting timeline state.

If valid `world_time_settings` are missing, world-time-dependent flows are intentionally blocked.

Core constraints:

- `tick_scale_mode` must be `fixed` or `flexible`
- `tick_units` must be non-empty and unique
- `tick_min_unit` must match the smallest configured unit

---

## Operational Guidance

When diagnosing pipeline or database issues, start with:

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`

Together they help reveal:

- whether write retries are spiking
- how many transactions are being executed
- whether the log queue is backing up
- whether world-level lock contention is growing
*/});

const gettingStartedZh = block(function () {/*
# 入门指南

**中文** | [**English**](./GETTING_STARTED_EN.md)

这份文档面向第一次接触 GameAgentEngine 的开发者，目标是带你从零完成三件事：启动 Engine、创建第一个世界、打开 Creator 开始编辑。

---

## 你会得到什么

完成本指南后，你应该可以：

- 启动本地 Engine 服务
- 使用 DevCli 创建世界根节点
- 在 Creator 中继续编辑节点、组件、关系
- 理解什么时候必须先配置 `world_time_settings`

---

## 前置条件

- Go 1.25+
- 可运行的终端环境
- 如果要接真实模型，需要一个兼容 OpenAI 协议的 API Key

如果 `llm.api_key` 留空，引擎会自动回退到 Mock Provider，这适合做本地功能联调，但不适合验证真实世界推理质量。

---

## 第一步：构建项目

```bash
git clone <仓库地址>
cd GameAgentEngine
go build ./...
```

如果你只是使用打包产物，也可以直接进入解压后的目录，不必重新编译。

---

## 第二步：准备配置文件

复制默认配置：

```bash
cp tools/source/gameagentengine.conf.yaml .
```

当前随包模板的关键点如下：

- 默认监听地址：`0.0.0.0:8080`
- 默认 API Key：`dev-key`
- 模板示例模型名：`deepseek-v4-flash`
- 模板示例 `base_url`：`https://api.deepseek.com`
- 模板执行模式：`debug`
- 默认后台自主调度器：关闭

额外需要知道的是，代码级保底默认值与模板示例并不完全相同；如果配置缺项，Engine 会回退到内部默认值，例如：

- `llm.model = gpt-4o-mini`
- `llm.base_url = https://api.openai.com/v1`
- `engine.execution_mode = full`

最少要检查这几个字段：

```yaml
auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-v4-flash"
  api_key: "sk-xxx"
  base_url: "https://api.deepseek.com"
```

如果你不想连接真实模型，可以把 `llm.api_key` 留空。

---

## 第三步：启动 Engine

```bash
go run ./cmd/gameagentengine serve
```

确认服务正常：

```bash
curl http://127.0.0.1:8080/health
```

预期结果：

```json
{"status":"ok"}
```

---

## 第四步：创建第一个世界

当前新手流程从直接创建一个世界根节点开始。

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --type world --name "新世界"
```

这条命令会直接创建一个 `world` 类型节点，它就是整个世界树的根节点。

你也可以继续创建子节点，例如：

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --world <world-id> --type location --name "起始村庄"
```

---

## 第五步：打开 Creator

推荐直接用 DevCli 打开：

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key inspect
```

如果你的环境不方便走这个入口，也可以直接打开：

`tools/source/web/GameAgentCreator/index.html`

进入后你可以看到这些核心页面：

- `Worlds`
- `Settings`
- `Policy`
- `Plans`
- `State`
- `Timelines`
- `Continuity`
- `Logs` / `Traces`

---

## 第六步：先配置世界时间，再跑 Tick

如果你要使用以下能力：

- `world tick`
- 时间线推进
- 连续性状态中的世界时间演化
- 世界线推理

那你应该先在 `Settings` 页面配置 `world_time_settings`。

这是当前设计中的强约束，不是可有可无的补充信息。没有世界时间系统，Engine 无法可靠地做时间推进和世界线连续性推理，所以相关保存/推进流程会故意阻塞，提醒开发者先完成配置。

---

## 第七步：观察管线状态

当你开始压测、跑 Tick 或启用自主行为时，推荐顺手观察：

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`

这三个入口可以帮助你快速判断：

- 是否出现频繁写重试
- 日志批量队列是否积压
- 世界级锁是否出现争用
*/});

const gettingStartedEn = block(function () {/*
# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide is for developers who are new to GameAgentEngine. The goal is to get you from zero to three things quickly: start the Engine, create your first world, and open Creator to continue editing.

---

## What You Will Have After This

By the end of this guide, you should be able to:

- start a local Engine service
- create a world root node with DevCli
- continue editing nodes, components, and relations in Creator
- understand when `world_time_settings` must be configured first

---

## Prerequisites

- Go 1.25+
- a working terminal environment
- an OpenAI-compatible API key if you want real model calls

If `llm.api_key` is empty, the engine falls back to the Mock Provider. That is fine for local flow verification, but not for validating real reasoning quality.

---

## Step 1: Build the Project

```bash
git clone <repo-url>
cd GameAgentEngine
go build ./...
```

If you are using a packaged build, you can work directly from the extracted directory instead of rebuilding.

---

## Step 2: Prepare the Config File

Copy the default config:

```bash
cp tools/source/gameagentengine.conf.yaml .
```

Important points in the packaged template:

- default listen address: `0.0.0.0:8080`
- default API key: `dev-key`
- sample model in the template: `deepseek-v4-flash`
- sample `base_url` in the template: `https://api.deepseek.com`
- template execution mode: `debug`
- background autonomous scheduler: disabled by default

There is also an important distinction between template values and code-level fallback defaults. If fields are omitted, the engine falls back to internal defaults such as:

- `llm.model = gpt-4o-mini`
- `llm.base_url = https://api.openai.com/v1`
- `engine.execution_mode = full`

At minimum, check these fields:

```yaml
auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-v4-flash"
  api_key: "sk-xxx"
  base_url: "https://api.deepseek.com"
```

If you do not want to call a real model yet, leave `llm.api_key` empty.

---

## Step 3: Start the Engine

```bash
go run ./cmd/gameagentengine serve
```

Confirm the service is healthy:

```bash
curl http://127.0.0.1:8080/health
```

Expected result:

```json
{"status":"ok"}
```

---

## Step 4: Create Your First World

The beginner flow now starts by creating a world root node directly.

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --type world --name "New World"
```

This creates a `world` node that acts as the root of the world tree.

You can then create child nodes, for example:

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --world <world-id> --type location --name "Starter Village"
```

---

## Step 5: Open Creator

The simplest path is through DevCli:

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key inspect
```

If that launcher path is inconvenient in your environment, you can also open:

`tools/source/web/GameAgentCreator/index.html`

Key pages include:

- `Worlds`
- `Settings`
- `Policy`
- `Plans`
- `State`
- `Timelines`
- `Continuity`
- `Logs` / `Traces`

---

## Step 6: Configure World Time Before Running Tick

If you plan to use:

- `world tick`
- timeline advancement
- world-time continuity state
- worldline reasoning

you should configure `world_time_settings` first in the `Settings` page.

This is a deliberate hard requirement in the current design. Without a valid world-time system, the engine cannot reliably advance time or maintain continuity reasoning, so related save / advance flows intentionally stop and ask you to finish the configuration first.

---

## Step 7: Watch Pipeline State Early

Once you start load-testing, running ticks, or enabling autonomous behavior, it helps to monitor:

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`

These endpoints quickly show whether:

- write retries are becoming frequent
- the batched log queue is backing up
- world-level lock contention is increasing
*/});

const packagedGettingStartedZh = block(function () {/*
# 入门指南

**中文** | [**English**](./GETTING_STARTED_EN.md)

这份文档面向使用打包产物的新手开发者。

---

## 最短上手路径

```bash
# 1. 启动服务
GameAgentEngine serve

# 2. 创建世界
GameAgentDevCli node create --type world --name "新世界"

# 3. 打开 Creator
GameAgentDevCli inspect
```

如果你要做世界时间推进，请先在 Creator 的 `Settings` 页面配置 `world_time_settings`。

---

## 配置文件

直接编辑当前目录里的 `gameagentengine.conf.yaml`。

当前随包模板重点：

- `auth.api_key: dev-key`
- `llm.model: deepseek-v4-flash`
- `llm.base_url: https://api.deepseek.com`
- `engine.execution_mode: debug`
- `engine.autonomous_scheduler_enabled: false`
- `engine.world_lock_enabled: true`

如果你删掉这些字段，Engine 仍会回退到代码级默认值，例如：

- `llm.model: gpt-4o-mini`
- `llm.base_url: https://api.openai.com/v1`
- `engine.execution_mode: full`

---

## 常用命令

```bash
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world tick <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
```

---

## 诊断入口

推荐同时关注：

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`
*/});

const packagedGettingStartedEn = block(function () {/*
# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide is for new developers using the packaged build.

---

## Shortest Path

```bash
# 1. Start the service
GameAgentEngine serve

# 2. Create a world
GameAgentDevCli node create --type world --name "New World"

# 3. Open Creator
GameAgentDevCli inspect
```

If you want world time advancement, configure `world_time_settings` first in Creator's `Settings` page.

---

## Config File

Edit the bundled `gameagentengine.conf.yaml` in the current directory.

Current packaged template highlights:

- `auth.api_key: dev-key`
- `llm.model: deepseek-v4-flash`
- `llm.base_url: https://api.deepseek.com`
- `engine.execution_mode: debug`
- `engine.autonomous_scheduler_enabled: false`
- `engine.world_lock_enabled: true`

If you omit these fields, the engine still falls back to code-level defaults such as:

- `llm.model: gpt-4o-mini`
- `llm.base_url: https://api.openai.com/v1`
- `engine.execution_mode: full`

---

## Common Commands

```bash
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world tick <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
```

---

## Diagnostics

Recommended endpoints:

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`
*/});

const packagedConfigZh = block(function () {/*
# 配置参考

**中文** | [**English**](./CONFIGURATION_EN.md)

随包版本的配置重点：

- 静态配置文件：当前目录 `gameagentengine.conf.yaml`
- 动态世界配置：`world_settings`
- 时间规则：`world_time_settings`
- 时间结果：`world_time_state`

---

## 随包模板当前重点

```yaml
database:
  driver: "sqlite"
  dsn: "gameagentengine.db"
  migrations_enabled: true
  write_retry_enabled: true
  log_batch_enabled: true

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
```

这份模板偏向本地演示与开箱体验；如果缺失字段，Engine 仍会回退到代码级默认值。

---

## 额外提醒

- `database.driver` 支持 `sqlite`、`mysql`、`postgres`
- `GET /api/v1/pipeline/stats` 适合排查锁竞争、批量写与重试情况
- 如果没有先配置 `world_time_settings`，世界时间推进相关流程会被阻塞
*/});

const packagedConfigEn = block(function () {/*
# Configuration

[**中文**](./CONFIGURATION.md) | **English**

Packaged-build configuration focus:

- static config file: local `gameagentengine.conf.yaml`
- dynamic world config: `world_settings`
- time rules: `world_time_settings`
- time result: `world_time_state`

---

## Current Packaged Template Highlights

```yaml
database:
  driver: "sqlite"
  dsn: "gameagentengine.db"
  migrations_enabled: true
  write_retry_enabled: true
  log_batch_enabled: true

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
```

This template is optimized for local demos and packaged onboarding. If fields are omitted, the engine still falls back to code-level defaults.

---

## Additional Notes

- `database.driver` supports `sqlite`, `mysql`, and `postgres`
- `GET /api/v1/pipeline/stats` is useful for lock contention, batching, and retry diagnosis
- if `world_time_settings` is missing, world-time-dependent flows are intentionally blocked
*/});

const pipelineStatsSectionZh = block(function () {/*
## 管线观测

### `GET /api/v1/pipeline/stats`

返回共享数据库/推理管线的轻量统计信息。

当前包括：

- 写重试 attempts / retries / recoveries / failures
- 事务次数与累计耗时
- log sink 队列深度、flush 次数、fallback 写入次数
- world lock 获取与争用统计

---
*/});

const pipelineStatsSectionEn = block(function () {/*
## Pipeline Observability

### `GET /api/v1/pipeline/stats`

Returns lightweight structured stats for the shared data pipeline.

Current fields cover:

- write retry attempts / retries / recoveries / failures
- transaction count and accumulated duration
- log sink queue depth, flush count, fallback writes
- world lock acquisition and contention stats

---
*/});

const copyBlockZhOld = block(function () {/*
上述复制类接口都接受如下可选请求体：

```json
{
  "name": "Optional Name",
  "lock_world": true
}
```
*/});

const copyBlockZhNew = block(function () {/*
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
*/});

const copyBlockEnOld = block(function () {/*
All copy routes accept an optional body like:

```json
{
  "name": "Optional Name",
  "lock_world": true
}
```
*/});

const copyBlockEnNew = block(function () {/*
All copy routes accept an optional body like:

```json
{
  "name": "Optional Name",
  "lock_world": true
}
```

Notes:

- `lock_world` is retained for compatibility
- the current implementation already enforces same-world exclusion for critical copy / restore flows
- treat this field as a caller intent signal rather than a way to bypass the safety boundary
*/});

write('tools/source/gameagentengine.conf.yaml', configTemplate);
write('docs/AUTONOMOUS_BEHAVIOR.md', autonomousZh);
write('docs/AUTONOMOUS_BEHAVIOR_EN.md', autonomousEn);
write('docs/PIPELINE_INTERNALS.md', pipelineZh);
write('docs/PIPELINE_INTERNALS_EN.md', pipelineEn);
write('docs/CONFIGURATION.md', configZh);
write('docs/CONFIGURATION_EN.md', configEn);
write('docs/GETTING_STARTED.md', gettingStartedZh);
write('docs/GETTING_STARTED_EN.md', gettingStartedEn);
write('tools/source/docs/CONFIGURATION.md', packagedConfigZh);
write('tools/source/docs/CONFIGURATION_EN.md', packagedConfigEn);
write('tools/source/docs/GETTING_STARTED.md', packagedGettingStartedZh);
write('tools/source/docs/GETTING_STARTED_EN.md', packagedGettingStartedEn);

for (const file of ['docs/API_REFERENCE.md', 'tools/source/docs/API_REFERENCE.md']) {
  let content = read(file);
  content = replacePatternOrThrow(
    content,
    /上述复制类接口都接受如下可选请求体：\r?\n\r?\n```json\r?\n\{\r?\n  "name": "Optional Name",\r?\n  "lock_world": true\r?\n\}\r?\n```/,
    copyBlockZhNew,
    `${file} copy block zh`
  );
  if (!content.includes('GET /api/v1/pipeline/stats')) {
    content = replacePatternOrThrow(
      content,
      /## 世界设置与世界策略\r?\n/,
      `${pipelineStatsSectionZh}\n## 世界设置与世界策略\n`,
      `${file} pipeline stats zh`
    );
  }
  write(file, content);
}

for (const file of ['docs/API_REFERENCE_EN.md', 'tools/source/docs/API_REFERENCE_EN.md']) {
  let content = read(file);
  content = replacePatternOrThrow(
    content,
    /All copy routes accept an optional body like:\r?\n\r?\n```json\r?\n\{\r?\n  "name": "Optional Name",\r?\n  "lock_world": true\r?\n\}\r?\n```/,
    copyBlockEnNew,
    `${file} copy block en`
  );
  if (!content.includes('GET /api/v1/pipeline/stats')) {
    content = replacePatternOrThrow(
      content,
      /## World Settings and Policy\r?\n/,
      `${pipelineStatsSectionEn}\n## World Settings and Policy\n`,
      `${file} pipeline stats en`
    );
  }
  write(file, content);
}

console.log('documentation rewrite complete');
