# 核心概念

**中文** | [**English**](./CORE_CONCEPTS_EN.md)

GameAgentEngine v0.4.5 使用四种基本抽象来建模游戏世界：**节点（Node）**、**组件（Component）**、**记忆（Memory）** 和 **关系（Relation）**。它们共同构成一个有向图，表示每个实体、其属性、历史以及与其他实体的连接。引擎还提供推理管线、记忆传播系统和子任务 DAG 编排能力。

---

## 1. 节点（Node）

节点是统一的实体抽象。游戏世界中的一切都是节点。

### 节点类型

| 类型 | 说明 |
|---|---|
| `world` | 游戏世界的根容器 |
| `faction` | 有组织的群体或阵营 |
| `location` | 地点或区域 |
| `npc` | 非玩家角色（Agent） |
| `item` | 物品或对象 |
| `quest_line` | 任务或故事情节 |
| `event` | 已安排或进行中的事件 |

### 节点结构

```json
{
  "id": "uuid字符串",
  "world_id": "世界uuid",
  "name": "艾琳议长",
  "node_type": "npc",
  "parent_id": "阵营uuid",
  "created_at": "2026-07-03T...",
  "updated_at": "2026-07-03T..."
}
```

### 节点层级

节点通过 `parent_id` 形成树状结构：

- `world` 节点没有父节点
- 世界节点是根节点，其 `world_id` 等于自身的 `id`
- 非世界节点必须通过 `world_id` 属于某个世界
- 非叶节点（有子节点的节点）不能被删除

```
World: 灰港边境
├── Faction: 灰港议会
│   ├── NPC: 艾琳议长 (belongs_to 灰港议会)
│   └── NPC: 布莱姆总管 (belongs_to 灰港议会)
├── Faction: 铁潮商会
│   └── NPC: 米拉代表 (belongs_to 铁潮商会)
├── Location: 议事厅
│   └── Location: 圆桌议会
├── Location: 雾湾矿场
│   └── Location: 雾湾征购站
└── Location: 北门要塞
    └── Location: 北门军营
```

---

## 2. 组件（Component）

组件是挂载在节点上的结构化数据。它们包含关于实体的所有描述性信息。

### 内置组件类型

| 类型 | 用途 | 典型内容 |
|---|---|---|
| `profile` | 身份与角色 | `{"name":"艾琳议长","personality":"冷静、克制"}` |
| `lore` | 世界或阵营背景 | 背景故事摘要 |
| `rule` | 玩法规则 | 回合制规则定义 |
| `timeline` | 当前回合上下文 | 回合阶段信息 |
| `action_policy` | 动作偏好 | AI 行为偏向 |
| `relations` | 关系摘要 | 缓存的关系概览 |
| `prompt_profile` | LLM 行为约束 | 角色语气和风格 |
| `autonomous` | 自主行为配置 | 触发条件、能力白名单 |
| `world_state` | 持续的世界连续性摘要 | 摘要、canonical facts、活跃剧情线 |
| `story_state` | 当前故事连续性状态 | 局势、近期变化、待推进线索 |
| `story_history` | 最近 tick 的连续性历史 | tick 条目、摘要、保留事实 |
| `tick_policy` | 连续性约束 | 规则、偏好事实、Prompt 指导 |
| `state_snapshot` | 引擎生成的检查点 | 最近一次结构化 world tick 载荷 |
| `memory` | 旧版记忆（已弃用） | 建议改用 Memory 实体 |

### 自定义组件

你可以创建与模式 `^[a-z][a-z0-9_:-]{1,63}$` 匹配的自定义组件类型：

来自 Demo 世界的示例：

- `resource_state` — `{"food":62,"order":58,"defense":49,"morale":55,"treasury":46}`
- `district_state` — `{"stability":48,"pressure":68,"output":72}`
- `demo_state` — 跟踪 Demo UI 的游戏状态

### 组件校验模式

Engine、GameAgentCreator 与 GameAgentDevCli 现在共享同一份组件校验元数据，以保证导入、API 写入与可视化编辑时的规则一致。

| 组件类型 | 校验模式 | 数据格式 | 当前约束 |
|---|---|---|---|
| `autonomous` | 强类型（strong） | JSON object | 会读取特定字段，当前至少要求 `enabled`、`trigger` 存在，并校验 `trigger` 枚举值；当 `trigger=scheduled` 时还会要求 `interval_seconds` 为正数 |
| `profile` | 弱类型（weak） | JSON object | 只要求是合法 JSON 对象，不限制具体字段集合 |
| `rule` / `timeline` / `action_policy` / `relations` / `prompt_profile` / `lore` | 无类型（free） | text | 当前按纯文本处理，不做结构化字段校验 |

如果后续某类组件开始被 Engine 结构化读取字段，应将其升级到弱类型或强类型，而不是继续沿用纯文本约定。

其中这些连续性状态组件属于 world tick 的世界级持久化工件：`world_state`、`story_state`、`story_history`、`tick_policy`、`state_snapshot`。它们可以通过 API、SDK、DevCli 和 Creator 读取，其中 `state_snapshot` 更适合作为只读检查点。

---

## 3. 记忆（Memory）

记忆存储 AI Agent（NPC）可以读取的信息。它们是引擎实现持久化上下文的机制。

### 记忆层级

| 层级 | 可见性 | 说明 |
|---|---|---|
| `short_term` | 仅该 NPC | 短暂，本轮推理结束后通常不保留 |
| `long_term` | 仅该 NPC | 永久，NPC 个人历史 |
| `shared` | 阵营/组织级 | 同阵营节点可见 |
| `world` | 世界中所有实体 | 世界级公共知识 |

### 记忆传播

引擎支持四种记忆传播模式，帮助记忆在节点层级之间流动：

| 模式 | 说明 |
|---|---|
| `upward` | 沿父链向上传播（默认），可通过 `propagation_max_depth` 限制层数 |
| `tag_broadcast` | 按 tags 匹配目标节点扩散 |
| `targeted` | 定向传播到指定节点列表 |
| `manual` | 不自动传播，用户手动触发 |

传播可配置为**状态机模式**（`enable_propagation_machine`），按预设规则链自动执行传播动作，包括内容转换（前缀添加、层级提升、标签追加）。

传播规则通过 WorldSettings 在数据库中进行配置，属于用户态动态配置。

### 记忆结构

```json
{
  "id": "uuid字符串",
  "node_id": "npc的uuid",
  "content": "她记得上一次边境暴乱，是因为税负和军需同时加压。",
  "level": "long_term",
  "tags": "npc,history,war",
  "created_at": "2026-07-03T..."
}
```

---

## 4. 关系（Relation）

关系是节点之间的有向边，构成图结构。

### 内置关系类型

| 类型 | 语义 | 示例 |
|---|---|---|
| `belongs_to` | 成员/归属 | NPC 属于某个阵营 |
| `ally` | 联盟或合作 | 阵营之间的联盟 |
| `enemy` | 敌对关系 | 阵营之间的敌对 |
| `subordinate` | 层级隶属 | 分支机构隶属于总部 |
| `kinship` | 家族或血亲关系 | NPC 之间的亲属关系 |
| `located_at` | 物理位置 | NPC 位于某个地点 |

### 关系结构

```json
{
  "id": "uuid字符串",
  "world_id": "世界的uuid",
  "source_id": "来源节点uuid",
  "target_id": "目标节点uuid",
  "relation_type": "belongs_to",
  "weight": 92,
  "properties": "{\"role\":\"chair\"}",
  "created_at": "2026-07-03T..."
}
```

### 权重

`weight` 是表示关系强度或情感的整数。正值代表正面关系，负值代表负面关系。

---

## 5. 任务类型（Task Types）

引擎支持五种推理任务类型：

| 任务类型 | 用途 |
|---|---|
| `npc_dialogue` | NPC 对话——以角色身份响应，携带完整上下文 |
| `world_tick` | 时间线推进——生成世界状态变化和未来大纲 |
| `world_event_impact` | 事件影响评估——评估事件如何影响世界 |
| `autonomous_act` | 节点自主行为——在 capabilities 白名单内决定是否行动 |
| `custom` | 用户自定义——使用上下文的原始系统 Prompt |

---

## 6. 管线模式（PipelineMode）

每个世界可以独立配置推理管线模式，存储在数据库 WorldSettings 中，与 ExecutionMode 互不干扰：

| 模式 | 值 | 功能 |
|---|---|---|
| 垂直模式 | `vertical` | 单轮 LLM 调用，不创建任务节点树，不轮询 |
| 轮询模式 | `polling` | 多轮 LLM 轮询，支持 request_data 数据查询 |
| 完整模式 | `full`（默认） | 完整功能：多轮轮询 + DAG 子任务编排 |

---

## 7. 子任务 DAG

在 full 模式下，LLM 可以在 JSON 响应中声明 `sub_tasks` 数组，引擎自动编排执行：

- **依赖解析**：`depends_on` 字段控制执行顺序
- **重试与超时**：每个子任务可配置重试次数和超时时间（通过 WorldSettings）
- **合并模式**：支持 `append`（追加）、`override`（覆盖）、`summarize`（LLM 摘要）

---

## 8. 动作（Actions）

动作是 LLM 在其响应中提出的操作。它们可以同步执行（在管线内）或异步执行（返回回调给游戏客户端）。

### 当前内置动作示例

| 动作 | 模式 | 说明 |
|---|---|---|
| `update_mood` | 同步 | 更新 NPC 情绪状态 |
| `add_memory` | 同步 | 直接向节点写入记忆 |
| `send_dialogue` | 同步 | 记录对话内容 |
| `adjust_relation` | 异步 | 请求调整关系权重 |
| `spawn_item` | 异步 | 请求创建物品 |

以上为当前引擎内置动作的典型示例，实际可用动作集合会随版本演进而扩展。

### 动作策略

通过数据库 WorldPolicy 配置 blocked_actions 和 safe_actions，控制 LLM 可执行的动作范围。

---

## 9. 世界时间线（World Timeline）

世界时间线跟踪各 Tick 之间的游戏状态变化。

### Tick 生命周期

1. 客户端请求 Tick 推进
2. 管线按 PipelineMode 执行 world_tick 任务
3. 策略引擎评估计划
4. 写入并传播记忆
5. 持久化时间线记录和连续性状态组件
6. 触发 world_tick_sync 自主节点
7. 创建 TimelineModel 记录

### 连续性持久化

`world_tick` 现在不再只依赖最近一次 `future_outline`。引擎会把连续性信息结构化落到：

- `world_state`
- `story_state`
- `story_history`
- `tick_policy`
- `state_snapshot`

为了减少剧情漂移，高价值事实会被提升进入 `world_state.canonical_facts` 和 `story_history.entries[].facts`，并在后续 tick 中配合连续性约束再次注入。

---

## 10. 世界分叉与存档

Engine 将世界复制能力拆分为三类正式语义：工作副本（ForkWorld）、存档快照（CreateWorldSnapshot）和从快照恢复（RestoreWorld）。这三类操作都会复制完整的世界数据，包括节点、组件、记忆、关系、设置和策略，并通过 API 触发为服务器端原子操作。

### API 端点

```
POST /api/v1/worlds/{world_id}/fork
POST /api/v1/worlds/{world_id}/snapshots
POST /api/v1/worlds/{world_id}/restore
```

### 可选参数

| 参数 | 类型 | 说明 |
|---|---|---|
| `lock_world` | bool | 复制期间锁定源世界，阻止并发写入（可选，默认不锁定） |

### 锁定机制

当 `lock_world` 为 `true` 时，引擎使用世界粒度的互斥锁（mutex）保护源世界或源快照世界。该锁仅影响 fork / snapshot / restore 操作本身，不会阻止正常的节点/组件/记忆/关系读写。锁定只持续复制或恢复过程，完成后立即释放。

### 行为说明

- 复制操作会创建源世界的完整副本，包括所有节点、组件、记忆、关系、世界设置（WorldSettings）和世界策略（WorldPolicy）
- 新世界以 `name` 参数命名，若未指定则自动生成"原名 (副本)"
- 复制是同步操作，大世界中可能需要较长时间完成

### 常见用途

- **保存/加载**：将游戏世界复制作为存档快照
- **世界分支**：基于现有世界创建分支，在不同方向上进行演化测试
- **A/B 测试**：复制同一世界并在不同配置下分别推进，比较结果差异
- **并行开发**：多个开发者基于同一世界快照独立工作

### 状态资产化边界

当前 Engine 有意将三类“带状态资产”严格分开：

- `ForkWorld`：用于编辑、调试、仿真和分支演化的可运行工作副本
- `CreateWorldSnapshot`：用于存档、兼容性检查和安全恢复的存档型快照
- `RestoreWorld`：从已校验快照恢复出的可运行世界

未来如果要支持 Agent 运行态存档、局部 scope 存档、纯结构导出等能力，不应继续复用这三类接口，而应引入新的资产类型，并为它们定义独立的兼容性和恢复规则。
