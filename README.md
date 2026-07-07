# GameAgentEngine

**中文** | [**English**](./README_EN.md)

面向游戏开发者的 AI Agent 创建与运行引擎。

GameAgentEngine 是一个基于 Go 的引擎，位于游戏逻辑与大模型能力之间，负责世界建模、NPC 行为、记忆管理、世界时间线推进，以及受控的运行时动作执行。

它不替代 Unity、Unreal 或 Godot，而是与这些游戏引擎协同工作。

---

## 核心能力

- 基于节点、组件、记忆、关系的统一世界图模型
- LLM 驱动的 NPC 对话与世界推理
- Tick 推进、事件影响评估、局部范围演化
- 三档管线模式：`vertical`、`polling`、`full`
- 将世界复制语义拆分为工作副本、存档快照、快照恢复三类操作
- Engine、Creator、DevCli 共享组件校验元数据，减少前后端规则漂移
- 运行时动态调整世界设置与世界策略
- 提供 Web 编辑器 `GameAgentCreator` 与命令行工具 `GameAgentDevCli`
- 提供 Go SDK，便于集成到自定义工具或服务
- 通过响应元数据、日志与调试轨迹提供可观测性

---

## 快速开始

```bash
# 1. 克隆并构建
git clone <repo-url>
cd GameAgentEngine
go build ./...

# 2. 复制默认配置
cp tools/source/gameagentengine.conf.yaml .

# 3. 在 gameagentengine.conf.yaml 中填写 llm.api_key

# 4. 启动引擎
go run ./cmd/gameagentengine serve

# 5. 导入示例世界
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml --reset

# 6. 打开 Creator
# tools/source/web/GameAgentCreator/index.html
```

完整流程见[入门指南](./docs/GETTING_STARTED.md)。

---

## 当前工具能力

### GameAgentCreator

当前 Creator 已支持：

- 世界选择与世界创建
- 在世界页中修改世界名称
- 节点树浏览与折叠状态保持
- 拖拽修改父子关系
- 拖拽到根级
- 节点创建、编辑、删除、复制
- 从节点操作中创建通用“被指向关系”
- 快照保存、校验、恢复、删除
- 世界设置、世界策略、日志、调试轨迹查看
- 按组件类型给出结构化校验提示

### GameAgentDevCli

当前 DevCli 已支持：

- 导入与验证流程
- 节点 / 组件 / 记忆 / 关系 CRUD
- 世界 Tick、事件影响评估、局部推进、时间线重规划
- 世界 fork、存档快照、恢复、快照校验、快照元数据、快照删除
- 世界运行时设置与世界策略管理
- 通过 `world update` 修改世界名称
- 通过 `node copy` 复制节点
- 复用 Engine 侧组件校验规则进行导入与写入校验

---

## 世界复制语义

GameAgentEngine 当前把世界复制拆分为三种不同但相关的操作：

- `ForkWorld`：创建可运行的工作副本，用于分支演化、编辑与调试
- `CreateWorldSnapshot`：创建面向存档的快照世界，并写入兼容性元数据
- `RestoreWorld`：先校验存档快照，再从快照恢复出新的可运行世界

快照元数据与运行中的世界图分离持久化，这样引擎可以在真正恢复前先完成兼容性校验，避免直接把不兼容状态写回运行世界。

---

## 管线模式

每个世界都可以独立选择三种推理管线模式之一：

- `vertical`：最轻量的单轮执行
- `polling`：支持多轮推理，但不启用最重的完整编排能力
- `full`：启用完整编排能力，包括更重的引擎特性

这样不同游戏可以只使用自己真正需要的那一档能力，以提高响应效率并降低运行成本。

---

## 项目结构

```text
GameAgentEngine/
|-- cmd/
|   |-- gameagentengine/
|   `-- gameagentdevcli/
|-- docs/
|-- internal/
|   |-- action/
|   |-- api/
|   |-- config/
|   |-- engine/
|   |-- llm/
|   |-- planner/
|   |-- service/
|   `-- store/
|-- sdk/
|-- tools/
|   `-- source/
|       `-- web/
|           `-- GameAgentCreator/
`-- web/
```

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

---

## 许可证

MIT
