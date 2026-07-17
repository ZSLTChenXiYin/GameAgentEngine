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

---

## 适合谁

- 想做文字游戏、类 DOL、剧情驱动 RPG、可持续 NPC 行为的游戏团队
- 需要让 AI 读取真实游戏状态，而不是靠 prompt 猜状态的团队
- 需要把异步回调、外部任务、测试和打包流程统一起来的团队
- 需要让策划和程序都能参与世界调试的团队

---

## 工作流长什么样

```text
World Definition -> Engine world modeling -> cold start baseline
                     -> world tick -> authority query -> controlled actions
                     -> timeline / memories / story state
                     -> Worker / Creator / DevCli / SDK integration
```

更贴近实际的流程是：

1. 先导入世界骨架
2. 冷启动生成运行基座
3. world tick 在权威数据与世界连续性上推进剧情
4. Worker 负责回应游戏侧真实状态
5. Creator / DevCli / SDK 负责编辑、联调、回归、打包

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
| SDKs | 外围系统和其他语言的统一接入面 |

---

## 文档入口

- [中文文档索引](./docs/README.md)
- [English documentation index](./docs/README_EN.md)

---

## 许可证

MIT
