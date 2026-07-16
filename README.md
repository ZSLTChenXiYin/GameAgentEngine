# GameAgentEngine

**中文** | [**English**](./README_EN.md)

面向游戏开发者的 AI Agent 引擎。

---

## 为什么游戏需要自己的 Agent Engine

如果你试过用常规 LLM Agent 做游戏，很可能遇到过以下问题：

> **NPC 每次对话都忘记你是谁。** 常规 Agent 没有世界 — 每次对话都是全新的上下文，NPC 不记得昨天见过玩家，不记得村口发生了什么事。

> **Agent 什么动作都能做。** LLM 什么都能说，但在游戏里 NPC 只能做游戏允许的事 — 卖道具、触发剧情、带路。Agent 需要校准的动作边界，而不只是提示词建议。

> **没有世界时间。** 游戏世界有自己的时钟：白天 NPC 开店，晚上打烊。常规 Agent 没有"世界时间"的概念。

> **集成靠手搓。** 每个 Agent 框架都只暴露 HTTP，但游戏需要推送、回调、WebSocket、重试、回退。每接一个游戏引擎都要重新造一遍轮子。

> **策划不会用。** 大部分 Agent 框架面向算法工程师。游戏策划需要的是可视化编辑器，能直接看 NPC 状态、调配置、跑测试。

GameAgentEngine 就是为解决这些问题设计的。

它是一个**专注于游戏的 AI Agent 运行时**，位于游戏逻辑与 LLM 之间，负责世界建模、NPC 行为、记忆管理、时间推进、外部任务调度和受控动作执行。不替代 Unity、Unreal 或 Godot，而是作为独立的 AI 世界层与它们协同工作。

---

## Engine 能做什么

### 世界建模 — 不只是一张对话表

Engine 用**节点、组件、记忆、关系**四条结构构建游戏世界：

- **节点** — 世界中的任何实体：NPC、地点、组织、物品、Quest
- **组件** — 节点承载的结构化数据：`autonomous`（自主行为配置）、`world_state`（世界概况）、`tick_policy`（策略规则）
- **记忆** — 节点的个人知识库，按重要等级分层（core / normal / trivial）
- **关系** — 节点间的社交图（ally / enemy / subordinate / kinship / located_at），Engine 自动组织上下文

```
[村庄] -> located_at -> [守门人 NPC]
[守门人 NPC] -> enemy -> [山贼 NPC]
[山贼 NPC] -> belongs_to -> [强盗组织]
```

当玩家与守门人对话时，Engine 会自动把"这个 NPC 和山贼是敌对关系、山贼属于强盗组织"写进提示词上下文。

### 受控的动作系统 — NPC 不能无法无天

LLM 输出的每一个动作都在**能力白名单**上校验：

- 同步动作立即执行（添加记忆、修改组件）
- 异步动作返回 `callback_id`，游戏端确认完成后再推进
- 动作越权会被拒绝并记录日志

这意味着你给一个 NPC 配置了 `["add_memory", "start_dialogue"]` 白名单，它就绝对不会擅自调 `delete_world`。

### 世界时间与连续性 — 游戏有自己的时钟

Engine 内置完整的世界时间系统：

- Tick 推进世界时间（支持 fixed / flexible 两种刻度模式）
- 每次 Tick 自动归档连续性状态：世界概况、故事状态、叙事历史、时间快照
- 下一次 Tick 时，上一轮的状态自动进入推理上下文

这就实现了"世界在持续演化，NPC 知道今天和昨天不同"。

### 外部交互 — 游戏端的异步任务

游戏端与 Engine 之间有三种任务投递方式：

| 模式 | 工作方式 |
|---|---|
| **Push** | Engine 直接通过 HTTP / WebSocket / RPC 推送任务给游戏端 |
| **Pull** | 游戏端轮询待办任务，认领，执行，心跳保活，回调结果 |
| **Hybrid** | 混合使用两者 |

支持 Fallback 路由和自动重试治理。

### 三档推理管线

| 模式 | 适用场景 |
|---|---|
| `vertical` | 单轮推理，低延迟，适合简单 NPC 对话 |
| `polling` | 多轮推理，支持数据请求与子任务，适合复杂叙事 |
| `full` | 完整推理，子任务 DAG、异步回调编排，适合世界 Tick |

### 后台自主行为

节点可以配置自主行为组件，NPC 在无人交互时也能主动行动：

- **手动触发** — 仅在需要时调用
- **Tick 同步** — 每次世界推进时自动触发
- **定时调度** — 后台调度器按配置周期运行

---

## 快速开始

### 构建

推荐使用打包脚本：

```
# Windows
tools\scripts\build.bat

# Linux / macOS
bash tools/scripts/build.sh
```

产物在 `dist/` 目录下，包含引擎、命令行工具、配置模板、demo / tests 数据、多语言 SDK 示例与 Creator Web 静态资源。

### 启动

```
# 复制默认配置
cp tools/source/gameagentengine.conf.yaml .

# 启动引擎
GameAgentEngine serve
```

### 创建世界并开始编辑

```
# 创建一个世界
GameAgentDevCli node create --type world --name "我的世界"

# 打开可视化编辑器
GameAgentDevCli creator
```

### 添加 NPC 并推进时间

```
# 在 Creator 中创建 NPC 节点，或使用命令行
GameAgentDevCli node create --world <world-id> --type npc --name "铁匠"

# 推进世界时间
GameAgentDevCli world tick <world-id>
```

> **注意：** 在跑 `world tick` 之前，需要先在 Creator 的 Settings 页面配置 `world_time_settings`。没有世界时间配置，时间推进会被刻意阻塞 — 这是 Engine 的设计约束，不是 BUG。

### 直接体验 Demo 世界与文字游戏壳

仓库现在附带一套最小 demo 资产：

- `tools/source/demo-world.yaml`：导入到 Engine 的 demo 世界
- `tools/source/demo-state.yaml`：供 `GameAgentWorker play` 使用的权威状态文件

最短体验路径：

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/demo-world.yaml
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

这样你可以一边让 Engine 扮演 NPC，一边让 worker 侧状态文件提供 HP、背包、金钱、任务和场景即时状态这类权威数据。

---

## 工具链

Engine 当前形成四个直接配套的开发工具入口：

### GameAgentCreator — Web 可视化编辑器

面向游戏策划和开发者，随引擎打包分发。所有页面：

| 页面 | 用途 |
|---|---|
| **Worlds** | 世界选择、创建、节点树拖拽编辑 |
| **Snapshots** | 快照管理（保存、校验、恢复、删除） |
| **Tasks** | 运行时任务监控（状态、分类、重试次数） |
| **Plans** | 待审批计划管理 |
| **Policy** | 世界策略配置 |
| **Settings** | 世界设置与 `world_time_settings` 编辑 |
| **Continuity** | 连续性调试包 |
| **State** | 连续性状态组件查看与编辑 |
| **Timelines** | 时间线归档 |
| **Logs** | 推理日志 |
| **Traces** | 调试轨迹 |

### GameAgentDevCli — 命令行工具

适用于脚本化和 CI 集成：

```
# CRUD
GameAgentDevCli node create --type world --name "世界"
GameAgentDevCli node list --world <id>
GameAgentDevCli component create --node <node-id> --type autonomous
GameAgentDevCli memory create --node <node-id> --content "..."
GameAgentDevCli relation create --source <a> --target <b> --type ally

# 世界运行
GameAgentDevCli world tick <world-id>
GameAgentDevCli world settings get <world-id>

# 任务管理（Pull 模式）
GameAgentDevCli task list --status pending --limit 20
GameAgentDevCli task claim <task-id> --consumer gamer
GameAgentDevCli task start <task-id> <lease-token>

# 调试与观测
GameAgentDevCli logs --world <world-id>
GameAgentDevCli debug traces --world <world-id>
GameAgentDevCli creator
```

### GameAgentWorker — 独立 Worker / REPL / 集成测试入口

适用于外部异步回调模拟、游戏侧本地状态承载、REPL 试玩和集成测试：

```bash
# 同时运行 push receiver 和 pull worker
GameAgentWorker serve --verbose

# 单次处理一个 pull task
GameAgentWorker pull-once --consumer game_client

# 进入文字游戏 REPL
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001

# 运行打包内置的集成测试场景
GameAgentWorker test all
```

它的定位不是临时测试脚本集合，而是项目内正式的游戏侧 worker：

- 集成测试时，承担外部 worker 的 push / pull / callback 闭环
- 开发时，承载 YAML / JSON 权威状态并模拟游戏侧异步接口
- 试玩时，提供 `/+talk`、`/+say`、`/+ask`、`/+act`、`/+gift`、`/+trade` 等 REPL 入口，验证 Engine 驱动的文字游戏体验

`play` 的直接文本输入会发给当前对话目标；`/+say` 用于房间公开发言；`/+act` 先做玩家意图解释和权威校验，再决定是否触发后续 NPC / 场景反馈。旧写法 `/talk`、`/ask` 仍兼容，但文档以 `/+cmd` 形式为准。

### Go SDK

Go 服务端集成用：

```go
import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

client := sdk.NewClient("http://127.0.0.1:8080", "dev-key")

// 列出待处理任务
tasks, _ := client.ListRuntimeTasks("", "pending", 20)

// 推进世界
tick, _ := client.AdvanceTick(worldID, "scheduled", "Day 2")
```

支持世界管理、节点 CRUD、记忆传播、自主行为配置、运行时任务管理、事件影响评估等全部 API。

---

## 详细介绍

- [入门指南](./docs/getting-started/GETTING_STARTED.md) — 从零搭建第一个世界
- [架构说明](./docs/architecture/ARCHITECTURE.md) — 整体设计与模块边界
- [核心概念](./docs/architecture/CORE_CONCEPTS.md) — 节点、组件、记忆、关系详解
- [外部交互总览](./docs/integration/EXTERNAL_INTERACTION.md) — Push / Pull / Hybrid、callback 与 worker 闭环
- [玩家交互总览](./docs/gameplay/PLAYER_INTERACTION.md) — 玩家输入、群聊、自然语言意图与校验边界
- [游戏状态权威边界](./docs/gameplay/GAME_STATE_AUTHORITY.md) — Engine 与游戏侧的数据所有权约束
- [API 参考](./docs/reference/API_REFERENCE.md) — 全部 HTTP 接口
- [GameAgentDevCli 指南](./docs/guides/GUIDE_GAMEAGENTDEVCLI.md) — 命令行工具完整手册
- [GameAgentCreator 指南](./docs/guides/GUIDE_GAMEAGENTCREATOR.md) — Web 编辑器使用说明
- [SDK 参考](./docs/reference/SDK_REFERENCE.md) — Go SDK 方法列表与类型说明
- [配置参考](./docs/reference/CONFIGURATION.md) — 静态配置与动态世界设置
- [自主行为系统](./docs/architecture/AUTONOMOUS_BEHAVIOR.md) — 节点级自主调度
- [推理管线内部实现](./docs/architecture/PIPELINE_INTERNALS.md) — Pipeline 执行流程详解
- [世界时间 Tick 参考](./docs/reference/WORLD_TIME_TICK_REFERENCE.md) — 时间刻度与推进规则
- [构建与部署](./docs/reference/BUILD_AND_DEPLOY.md) — 多平台编译与部署

---

## 许可证

MIT
