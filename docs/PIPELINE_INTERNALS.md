# 推理管线内部实现

**中文** | [**English**](./PIPELINE_INTERNALS_EN.md)

本文档详细说明 GameAgentEngine v0.4.5 推理管线的内部机制，包括管线模式、多轮轮询、子任务 DAG、数据请求循环和记忆传播。

---

## 管线模式（PipelineMode）

每个世界可以独立配置管线模式，存储在数据库 WorldSettings 中：

| 模式 | 值 | 功能 |
|---|---|---|
| 垂直模式 | `vertical` | 单轮 LLM 调用，不创建任务节点树，不轮询，最少功能 |
| 轮询模式 | `polling` | 多轮 LLM 轮询，支持 request_data 数据查询，不创建子任务 DAG |
| 完整模式 | `full` | 完整功能：多轮轮询 + DAG 子任务编排 |

模式通过 DevCli 或 Creator 的 WorldSettings 配置，与 ExecutionMode（debug/review/production）独立互不影响。

---

## Pipeline.Execute 主流程

1. 加载世界设置（memory_limit, max_analysis_rounds, pipeline_mode 等）
2. 加载世界策略（blocked_actions / safe_actions）
3. 构建初始上下文（BuiltContext）
4. 按任务类型分发：
   - `npc_dialogue` → executeDialogue（含第一轮分析）
   - `world_tick` → executeWorldTick
   - `world_event_impact` → executeWorldEvent
   - `autonomous_act` → executeAutonomousAct（含 capability 过滤）
   - `custom` → executeCustom
5. 每种任务类型最终进入 executeMultiTurnLoop 公共循环

---

## executeMultiTurnLoop 公共循环

多轮推理循环是所有任务类型的核心引擎：

```
第 1 轮：
  1. 构建系统 Prompt（含上下文 + 任务节点树 + 指令）
  2. 调用 LLM
  3. 解析 JSON 响应
  4. 如果是 full 模式，检查 raw_sub_tasks → 创建 DAGInstance
  5. 处理 request_data → 异步数据请求等待
  6. 执行同步动作
  7. 写入记忆更新 → 传播
  8. 构建下一个 TaskNode
  9. 记录推理日志

第 2+ 轮（需要时）：
  - 将上一轮分析结果加入上下文
  - 再次调用 LLM
  - 同上处理

当达到 max_analysis_rounds 或 LLM 标记 decision=stop 时结束
```

---

## 子任务 DAG 编排

在 full 模式下，LLM 可以在 JSON 响应中声明 `sub_tasks` 数组：

```json
{
  "reply": "需要并行调查几个方面。",
  "sub_tasks": [
    {"label": "investigate_market", "task_type": "custom", "node_id": "...", "depends_on": []},
    {"label": "assess_military", "task_type": "custom", "node_id": "...", "depends_on": []},
    {"label": "make_plan", "task_type": "custom", "node_id": "...", "depends_on": ["investigate_market", "assess_military"]}
  ]
}
```

DAGInstance 负责：

- **注册**子任务声明
- **依赖解析** — depends_on 为空 = 立即就绪；非空 = 等待前置完成
- **并发执行**就绪子任务（goroutine）
- **重试与超时** — 每个子任务最多重试 MaxRetries 次，超时时间为 TimeoutDuration
- **结果合并** — 支持三种合并模式：
  - `append`（默认）：全部追加
  - `override`：后完成的覆盖前结果
  - `summarize`：LLM 语义摘要
- **失败处理** — 失败不阻塞依赖链，失败信息附加到最终回复

---

## 数据请求循环

LLM 在响应中可以发起 `request_data` 查询请求，管线执行以下逻辑：

1. 解析 DataRequest 的 queries
2. 对 `target="store"` 的查询执行数据加载（节点组件、记忆、关系）
3. 对 `target="game_client"` 的查询通过 callback 机制等待外部响应
4. 加载完成后将数据注入下轮上下文
5. 循环最多执行 max_analysis_rounds 轮

---

## 动作执行流程

1. 管线解析 LLM 输出的 action_calls
2. 对 autonomous_act 任务进行 capability 校验和 schema 校验
3. 同步动作在管线内立即执行
4. 异步动作返回 callback_id 给调用方
5. 调用方通过 POST /api/v1/actions/callback 上报执行结果
6. ActionRegistry 匹配 callback_id 并存储结果

---

## 记忆处理与传播

1. LLM 声明 memory_updates（含 propagation 规则）
2. 管线创建 MemoryModel 并持久化
3. 按 PropagationRule 执行传播：
   - upward：沿父链递归上传（由 propagation_max_depth 限制）
   - tag_broadcast：匹配同世界拥有相同 tags 的节点
   - targeted：写入指定的 NodeID 列表
   - manual：不自动传播
4. 可选启用状态机模式（enable_propagation_machine）：
   - 每轮传播后检查规则链
   - 满足触发条件时执行 PropagateAction
   - 支持 TransformRule（内容前缀、层级提升、标签追加）

---

## 配置参考

### 静态配置（gameagentengine.conf.yaml）

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"    # sqlite / mysql
  dsn: "gameagentengine.db"

auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: ""
  base_url: "https://api.deepseek.com/v1"

engine:
  execution_mode: "production"    # debug / review / production
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

### 动态配置（数据库 WorldSettings）

可通过 DevCli 或 Creator 配置：

| 字段 | 默认值 | 说明 |
|---|---|---|
| memory_limit | 50 | 每次推理加载的最大记忆条数 |
| max_analysis_rounds | 5 | LLM 多轮轮询最大次数 |
| max_context_depth | 3 | 上下文向上追溯最大深度 |
| auto_apply | true | 是否自动执行变更计划 |
| require_review_above | critical | 超过此等级需审核 |
| pipeline_mode | full | 管线模式：vertical/polling/full |
| propagation_max_depth | 2 | 记忆向上传播最大层数 |
| sub_task_max_retries | 2 | 子任务最大重试次数 |
| sub_task_timeout_secs | 60 | 子任务超时秒数 |
| enable_propagation_machine | false | 是否启用标签传播状态机 |
