# GameAgentEngine

**中文** | [**English**](./README_EN.md)

面向游戏开发的 AI Agent 运行时。

它不是一个“会聊天的 NPC 包装器”，而是一层专门给游戏准备的世界运行时：负责世界连续性、权威状态接入、受控动作执行、世界时间推进，以及可调试、可打包、可联调的开发工作流。

---

## 它解决什么问题

如果你试过把通用 LLM Agent 直接塞进游戏，通常会遇到这些问题：

| 游戏开发痛点 | 通用 Agent 的常见问题 | GameAgentEngine 的处理方式 |
|---|---|---|
| NPC 每次对话都像失忆 | 没有持久世界，上一轮发生了什么很快就丢 | 节点 / 记忆 / 关系 / 连续性状态把世界串起来 |
| AI 经常胡编玩家状态 | 不知道玩家是否真的有刀、钱、背包、任务阶段 | 通过权威查询按需读游戏侧真实数据 |
| 动作输出不可控 | 模型会说很多“做了”的事，但游戏里不能直接执行 | 受控动作白名单 + 同步 / 异步执行 + callback 确认 |
| 世界没有时间感 | 只有对话，没有白天黑夜、剧情推进和历史归档 | world tick + world time + timeline + future outline |
| 异步联调很碎 | 只会 HTTP，不会面对 push / pull / callback / retry | Engine / Worker / runtime task / resume 链路统一接入 |
| 策划和程序都不好用 | 只有 prompt 和代码，缺少可视化编辑与调试入口 | Creator + DevCli + Worker + docs + demo workflow |
| 集成测试难稳定复现 | 状态模拟和外部回调容易散成脚本碎片 | Worker CLI + fixtures + packaged smoke / integration 流程 |

---

## 它是什么

GameAgentEngine 是一个位于游戏逻辑与 LLM 之间的独立运行时，负责：

- 世界建模
- NPC 行为
- 记忆管理
- 时间推进
- 外部任务调度
- 受控动作执行
- 权威状态查询
- 可观测、可复现的调试链路

它不替代 Unity、Unreal 或 Godot，而是作为游戏世界之上的 AI 层与它们协同工作。

---

## 核心能力

### 1. 世界建模

世界不是一张对话表，而是由四类结构组成：

- 节点：NPC、地点、组织、物品、任务线、世界
- 组件：结构化状态，如 `world_state`、`story_state`、`tick_policy`、`autonomous`
- 记忆：节点的知识库，按层级区分
- 关系：`located_at`、`belongs_to`、`subordinate`、`enemy` 等图结构

### 2. 权威状态

高频、强权威、会变化的数据保留在游戏侧，Engine 按需查询，不靠猜。

例如：

- HP
- 金钱
- 背包
- 任务阶段
- 场景占用
- 即时天气
- 临时事件状态

### 3. 受控动作

LLM 的输出会先经过能力白名单校验：

- 同步动作：立即执行
- 异步动作：返回 `callback_id`，由游戏侧确认后再推进
- 越权动作：拒绝并记录日志

### 4. 世界时间推进

Engine 内置世界时间系统：

- 支持 `fixed` / `flexible` 两种刻度模式
- 每次 tick 归档连续性状态
- 下一次 tick 自动带上上一次的世界历史

### 5. 可联调的工作流

Engine 配套提供：

- GameAgentCreator：可视化编辑器
- GameAgentDevCli：命令行与 CI 工具
- GameAgentWorker：游戏侧 worker、REPL、集成测试入口
- 多语言 SDK：给外围系统和游戏服务提供统一接入面
### 6. 玩家交互与意图解释

玩家输入通过 Engine 侧的意图解释管道转化为结构化意图：

- POST /api/v1/player/input/interpret 将自然语言转为意图体
- 支持 speech、move、gift、trade_request、threaten、inspect、use_item 等意图类型
- 交互模式包括 direct_dialogue、group_chat、gift_response
- 交互回合追踪 actor / target / scene / participant 语义

### 7. 可观测性与度量

Engine 内置结构化日志与运行时观测：

- 分层日志：按 execution mode（debug / review / production）控制输出粒度
- TraceID 在 pipeline 上下游传播，支持请求链路追踪
- GET /metrics 端点暴露 Go 运行时指标（goroutines、内存、GC 暂停）
- GET /api/v1/pipeline/stats 返回 store、lock、log sink 等运行时统计

---

## 适合谁

- 想做文字游戏、类 DOL、剧情驱动 RPG、可持续 NPC 行为的游戏团队
- 需要让 AI 读取真实游戏状态，而不是靠 prompt 猜状态的团队
- 需要把异步回调、外部任务、测试和打包流程统一起来的团队
- 需要让策划和程序都能参与世界调试的团队

---

## 工作流长什么样

```text
世界设定 / 世界骨架导入
    -> Engine 进行世界建模与冷启动
    -> world tick 推进世界时间、剧情与记忆
    -> 按需向游戏侧查询权威数据
    -> 执行受控动作、回调确认、状态更新
    -> Worker / Creator / DevCli / SDK 完成联调、编辑、测试和打包
```

更贴近实际的流程是：

1. 先导入世界骨架或基础配置。
2. Engine 根据世界设定完成冷启动，建立可推理的运行基座。
3. world tick 负责推进世界时间、人物状态、剧情和记忆。
4. 当遇到高频变化且必须权威的数据时，Engine 按需向游戏侧查询。
5. Worker 负责模拟游戏侧响应、REPL 体验和集成测试。
6. Creator / DevCli / SDK 负责编辑、联调、回归和打包。

一句话总结：

> 先把世界搭起来，再让 Engine 推着世界往前走，必要时去游戏侧拿权威数据，最后由 Worker、Creator、DevCli 和 SDK 把开发、测试和发布串起来。

---

## 快速开始

### 构建

```bash
# Windows
tools\scripts\build.bat

# Linux / macOS
bash tools/scripts/build.sh
```

### 启动引擎

```bash
GameAgentEngine serve
```

### 创建世界并打开编辑器

```bash
GameAgentDevCli node create --type world --name "我的世界"
GameAgentDevCli creator
```

### 体验 Demo 世界与文字游戏壳

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/workerhome/demo/demo-world.yaml
GameAgentWorker play --state-file tools/source/workerhome/demo/demo-state.yaml --world-id demo_world --player-node-id player_001
```

---

## 工具链入口

| 工具 | 作用 |
|---|---|
| GameAgentCreator | 世界编辑、状态查看、调试与回归入口 |
| GameAgentDevCli | 导入、初始化、调试、日志、时间推进、打包验收 |
| GameAgentWorker | 权威状态、REPL、push/pull/callback、集成测试 |
| Go SDK | Engine REST API 的 Go 客户端（语义基线） |
| TypeScript SDK | Engine REST API 的 TypeScript 客户端（`sdk/typescript/`） |

---

## 文档入口

- [中文文档索引](./docs/README.md)
- [English documentation index](./docs/README_EN.md)

---

## 许可证

MIT
