# GameAgentEngine

**中文** | [**English**](./README_EN.md)

面向游戏开发者的 AI Agent 创建与运行引擎。

GameAgentEngine 位于游戏逻辑与大模型能力之间，负责世界建模、NPC 行为、记忆管理、世界时间推进，以及受控的运行时动作执行。它不替代 Unity、Unreal 或 Godot，而是作为独立 AI 世界层与它们协同工作。

---

## 核心能力

- 基于节点、组件、记忆、关系的统一世界图模型
- LLM 驱动的 NPC 对话、世界推理与事件影响评估
- 世界 Tick、连续性状态、时间线归档与世界时间刻度推进
- 三档推理管线：`vertical`、`polling`、`full`
- 工作副本、存档快照、恢复三类世界复制语义
- Engine、Creator、DevCli 共享组件校验元数据
- 运行时动态管理 `world_settings`、`world_policy` 和连续性状态组件
- 提供 Web 编辑器 `GameAgentCreator`、命令行工具 `GameAgentDevCli` 和 Go SDK

---

## 快速开始

如果你是第一次接触这个项目，推荐直接看[入门指南](./docs/GETTING_STARTED.md)。

最短路径如下：

```bash
# 1. 构建
go build ./...

# 2. 复制默认配置
cp tools/source/gameagentengine.conf.yaml .

# 3. 启动引擎
go run ./cmd/gameagentengine serve

# 4. 创建一个世界根节点
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --type world --name "新世界"

# 5. 打开 Creator
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key inspect
```

如果你计划使用 `world tick`、时间线推进和世界线推理，请先为该世界配置 `world_time_settings`。否则这部分能力虽然接口存在，但世界时间系统本身没有定义，开发流程会被刻意阻塞以提醒补配置。

---

## 当前工具能力

### GameAgentCreator

当前 Creator 已支持：

- 世界选择与世界创建
- 世界重命名
- 节点树浏览、拖拽改父节点、拖到根级
- 节点创建、编辑、删除、复制
- 关系创建与节点图校验提示
- 快照保存、校验、恢复、删除
- 世界设置、世界策略、待审批计划管理
- 连续性、状态、时间线、日志、调试轨迹查看
- `world_time_settings` 编辑与 `world_time_state` 查看

### GameAgentDevCli

当前 DevCli 已支持：

- 节点、组件、记忆、关系 CRUD
- 世界 Tick、事件影响评估、局部范围推进、时间线重规划
- 世界设置、世界策略、连续性状态组件管理
- 工作副本、存档快照、恢复、快照校验与快照元数据查询
- 日志、调试轨迹、连续性调试、节点图调试
- 调用 Creator 的 `inspect` 入口

### Go SDK

当前 SDK 已支持：

- 基础服务访问、版本读取、健康检查
- 世界设置与 `world_time_settings` 的读取和部分更新
- 连续性状态组件与时间线归档访问
- 世界 Tick、事件影响、计划审批、快照相关接口

---

## 文档索引

- [入门指南](./docs/GETTING_STARTED.md)
- [架构说明](./docs/ARCHITECTURE.md)
- [核心概念](./docs/CORE_CONCEPTS.md)
- [API 参考](./docs/API_REFERENCE.md)
- [GameAgentDevCli 指南](./docs/GUIDE_GAMEAGENTDEVCLI.md)
- [GameAgentCreator 指南](./docs/GUIDE_GAMEAGENTCREATOR.md)
- [SDK 参考](./docs/SDK_REFERENCE.md)
- [配置参考](./docs/CONFIGURATION.md)
- [构建与部署](./docs/BUILD_AND_DEPLOY.md)
- [世界时间 Tick 参考](./docs/WORLD_TIME_TICK_REFERENCE.md)

---

## 许可证

MIT
