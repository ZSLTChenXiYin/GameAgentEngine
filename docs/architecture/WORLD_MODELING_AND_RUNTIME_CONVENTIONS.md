# 世界建模与运行约定

**中文** | [**English**](./WORLD_MODELING_AND_RUNTIME_CONVENTIONS_EN.md)

本文档用于统一 GameAgentEngine 在世界建模、运行基座、权威状态、world tick bootstrap 与结果写回上的推荐约定。

目标不是增加新的配置负担，而是明确以下边界：

- 开发者主要维护什么；
- Engine 应自动生成什么；
- 游戏侧 / Worker 应负责什么；
- 哪些数据可以进入 Engine 持久层，哪些数据只能按需查询。

---

## 1. 适用范围

本文档适用于以下场景：

- 世界 YAML / JSON 导入；
- Demo 世界设计；
- GameAgentWorker play / authority-state 工作流；
- world tick / scope tick 推理；
- Creator、DevCli、Worker 与 Engine 的职责划分；
- 后续 SDK 与游戏侧集成约定。

---

## 2. 三层数据模型

推荐将世界相关数据分为三层，而不是混合维护。

| 层级 | 作用 | 典型内容 | 主要维护者 | 是否持久化到 Engine |
|---|---|---|---|---|
| 设定层 | 描述稳定世界事实 | world、location、npc、item、profile、lore、稳定关系、基础 rule | 开发者 / Creator | 是 |
| 运行基座层 | 支撑 Engine 连续性与推理起点 | story_state、world_state、story_history、state_snapshot、初始 future outline | Engine 冷启动 | 是 |
| 权威动态层 | 描述高频、强权威、易变化事实 | HP、money、inventory、quest state、scene occupancy、live flags | 游戏侧 / Worker | 否 |

核心原则：

- 设定层描述“世界是什么”；
- 运行基座层描述“世界现在处于什么叙事状态”；
- 权威动态层描述“游戏侧当前真实值是什么”。

---

## 3. 设定层建模规范

开发者应优先维护稳定、低频、具有长期意义的世界事实。

推荐放入世界导入文件的内容包括：

- 世界根节点、场景节点、NPC 节点、物品节点；
- 稳定父子结构与稳定关系；
- profile、lore、基础 rule；
- 能作为初始叙事锚点的少量关键记忆；
- 不依赖实时权威状态也成立的组织、地理、阵营、社会结构。

不推荐把以下内容直接写成长期组件或长期节点数据：

- 当前 HP / max HP；
- 当前金钱；
- 当前背包明细；
- 当前任务即时阶段；
- 当前场景占用人数；
- 即时天气、即时风险标记、临时战斗状态；
- 高频变化且必须与游戏侧严格一致的运行时数值。

---

## 4. 运行基座层约定

运行基座层是 Engine 进行连续性推理所依赖的内部运行态，不建议由开发者手工维护大量细节。

推荐由 Engine 冷启动或重建机制生成的内容包括：

- `world_state`：当前世界叙事摘要、主要场景压力、全局局势简述；
- `story_state`：当前叙事阶段、主线张力、活跃冲突、关键角色摘要；
- `story_history`：初始世界历史条目与后续连续性条目；
- `state_snapshot`：运行基座初始化快照；
- 初始 `future_outline`：未来若干 tick 的粗粒度世界推进大纲；
- 必要时的默认 `world_time_state` 实例。

不建议要求开发者手工插入大量世界运行态组件，原因如下：

- 容易与世界设定输入不一致；
- 容易在世界结构改动后过时；
- 很多字段本质上是推理结果而不是原始设定；
- 会把建模流程变成“为了能跑而补配置”。

---

## 5. 权威动态层约定

高频、强权威、可瞬时变化的数据应保留在游戏侧或 Worker 侧，Engine 通过权威查询按需获取。

典型示例：

- 玩家与 NPC 的 HP / MP / money；
- 当前 inventory、equipment、item ownership；
- 当前任务阶段与可交互状态；
- 当前 scene 的 occupants、即时天气、即时事件标记；
- 当前地点占用、锁定、战斗、倒地、死亡等实时状态。

这些数据默认不应被长期固化到 Engine 组件中。

原因：

- 固化后极易失真；
- 推理前后容易与游戏侧真实数据不一致；
- 会制造双写状态；
- 会降低后续 callback / authority-query 的价值。

---

## 6. Demo 数据拆分约定

Demo 推荐拆成两部分：

| 文件 | 职责 |
|---|---|
| `demo-world.yaml` | Engine 可导入的世界骨架与稳定设定输入 |
| `demo-state.yaml` | Worker / 游戏侧 authority-state 示例 |

推荐约定：

- `demo-world.yaml` 只承载稳定设定；
- `demo-state.yaml` 只承载动态权威状态；
- 两者通过运行时查询链路接合，而不是通过静态同步合并。

反模式：

- 把 `demo-state.yaml` 同步成一组 Engine 长期组件；
- 用组件模拟实时 HP、inventory、occupants、quest state；
- 让 Demo 世界为了 world tick 收敛而破坏权威边界。

---

## 7. 世界冷启动约定

世界冷启动（cold start）是指：

- 在世界设定输入导入完成后；
- 由 Engine 基于静态世界结构生成初始运行基座；
- 使世界进入“可继续 tick”的状态。

冷启动不等于：

- world tick；
- import 的隐式副作用；
- 权威动态状态同步。

冷启动推荐完成的工作包括：

1. 校验世界骨架是否可运行；
2. 生成初始 `story_state` / `world_state`；
3. 生成初始 `story_history` 条目；
4. 生成初始化快照；
5. 生成初始 `future_outline`；
6. 标记当前世界已完成运行基座初始化。

冷启动不应固化以下信息：

- 当前 HP；
- 当前 money；
- 当前 inventory；
- 当前 scene occupancy；
- 任何必须由游戏侧实时回答的事实。

---

## 8. World Tick Bootstrap 约定

world tick bootstrap 是 world tick 进入主 LLM 推理前的权威预取阶段。

其目的不是替代 `request_data`，而是减少低价值、重复性的开局补查。

推荐流程：

1. Engine 构建静态世界上下文；
2. Engine 根据当前 tick 的 scope、scene、参与者和任务类型，生成一组关键权威查询；
3. 游戏侧 / Worker 返回权威快照；
4. Engine 将快照作为本次请求的临时上下文注入；
5. LLM 在“世界骨架 + 权威快照”的基础上开始 world tick；
6. 仅在仍缺关键事实时再继续 `request_data`。

推荐的 world tick bootstrap 查询类型包括：

- `scene_state`
- `scene_occupants`
- `player_state`
- `player_inventory`
- `task_state`
- `item_presence`
- `npc_state`

---

## 9. Bootstrap 查询语义与执行方式

推荐只有一套查询语义，但允许两种执行方式。

| 层级 | 推荐做法 |
|---|---|
| 查询语义层 | 统一使用同一套 `request_data` / authority-query 语义 |
| 执行方式层 | 支持同步优先、callback 兜底 |
| 传输层 | 允许 HTTP / WS / RPC / Worker 本地实现 |

推荐原则：

- 同步优先：适用于 Demo、Worker、本地联调、低延迟服务侧；
- 异步 callback 兜底：适用于复杂聚合、客户端在线不稳定或高延迟场景；
- 不要为 bootstrap 另造一套完全不同的接口语义。

---

## 10. 写回规则

Engine 应只沉淀叙事结果，不沉淀高频权威原始状态。

允许写回的内容：

- `story_history` 条目；
- `story_state` 的阶段变化；
- `world_state` 的稳定叙事变化；
- timeline 摘要；
- 经过推理确认的短期记忆或共享记忆；
- 已经在游戏侧确认并可长期引用的剧情结果。

不允许默认写回的内容：

- raw HP / money / inventory；
- raw quest live state；
- raw scene occupancy；
- 单次请求中的临时 authority snapshot；
- 仅为当轮推理服务的动态快照。

---

## 11. Scope 与上下文边界

无论是 world tick、scope tick、dialogue 还是 play，都不应默认展开整张世界图。

推荐边界：

- 以当前 scope / scene / participant 为第一层；
- 以环境节点与关键祖先链为第二层；
- 只在任务确有需要时补充受控 descendants；
- 对 deeper descendants 使用受控提升策略，而不是递归全展开。

后续如引入 `world_focus` 或活跃节点筛选，也应遵守以下原则：

- 先摘要，后细化；
- 先挑活跃节点，再做局部展开；
- 不把 world tick 变成“默认吃下整图”的大 prompt。

---

## 12. Creator / DevCli / Worker / Engine 职责边界

| 模块 | 主要职责 |
|---|---|
| Creator | 编辑世界设定输入、查看运行态、触发冷启动与调试工作流 |
| DevCli | 导入、初始化、调试、回归、打包验收、脚本化操作 |
| Worker | 提供 authority-state、外部任务回调、play / Demo 联调 |
| Engine | 世界建模运行时、冷启动、world tick 推理、连续性维护、受控写回 |

推荐原则：

- Creator 面向设定编辑；
- DevCli 面向工作流控制；
- Worker 面向权威响应与联调；
- Engine 面向内核推理与连续性。

---

## 13. 反模式清单

以下做法不推荐：

1. 把高频权威状态长期固化到 Engine 组件；
2. 让开发者手工维护大量 `story_state` / `world_state` 细节；
3. 把 world tick bootstrap 做成另一套割裂的查询体系；
4. 让 world tick 默认展开全量关系图；
5. 用 Demo 的静态 world 替代真实 authority 链路；
6. 把 world tick 当成冷启动；
7. 将临时 authority snapshot 直接落入长期持久层。

---

## 14. 推荐总原则

建议用以下一句话概括本约定：

> 开发者负责世界设定输入，Engine 负责世界运行基座与连续性推理，游戏侧负责动态权威状态；高频权威数据按需查询，只将叙事结果沉淀回 Engine。
