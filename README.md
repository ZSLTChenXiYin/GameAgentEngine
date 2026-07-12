# GameAgentEngine

**中文** | [**English**](./README_EN.md)

面向游戏开发者的 AI Agent 创建与运行引擎。

GameAgentEngine 位于游戏逻辑与大模型能力之间，负责世界建模、NPC 行为、记忆管理、世界时间推进、运行时外部任务调度，以及受控的异步动作执行。它不替代 Unity、Unreal 或 Godot，而是作为独立 AI 世界层与它们协同工作。

---

## 核心能力

- **世界图模型** — 基于节点、组件、记忆、关系的统一世界建模，支持自主编排与共享组件校验元数据
- **LLM 推理管线** — 三档管线模式（vertical / polling / full），支持多轮推理、子任务 DAG、数据请求循环和异步动作回调
- **世界时间与连续性** — Tick 推进、连续性状态组件、时间线归档、世界时间刻度系统
- **外部交互与运行时任务** — Push / Pull / Hybrid 三种异步任务投递模式，支持 HTTP / RPC / WebSocket 适配器扩展
- **记忆传播** — 支持 upward / environment_scope / organization_scope / tag_broadcast / targeted 五种传播模式
- **工作副本与快照** — Fork（可运行工作副本）、Snapshot（存档快照）、Restore（恢复）三类世界复制语义
- **后台自主行为调度** — 节点级 autonomous 组件，支持手动触发、Tick 同步触发、定时调度三种模式
- **数据库管线** — 统一写事务、批量日志、批量记忆写入、可重试写层；支持 SQLite / MySQL / PostgreSQL
- **可观测性** — pipeline stats、推理日志、调试轨迹、world lock 争用统计
- **三套开发者工具** — GameAgentCreator（Web 编辑器）、GameAgentDevCli（命令行工具）、Go SDK

---

## 快速开始

### 编译构建

推荐使用打包脚本编译，产物会自动包含配置文件、文档和 Creator 编辑器：

```bash
# Windows
tools\scripts\build.bat

# Linux / macOS
bash tools/scripts/build.sh
```

编译产物输出到 `dist/` 目录，按平台分类的文件夹内包含：

- `GameAgentEngine` — 引擎服务端
- `GameAgentDevCli` — 命令行开发工具
- `gameagentengine.conf.yaml` — 默认配置文件
- `docs/` — 文档
- `web/GameAgentCreator/` — Web 编辑器

你也可以直接使用 Go 编译当前平台：

```bash
go build -o GameAgentEngine ./cmd/gameagentengine/
go build -o GameAgentDevCli ./cmd/gameagentdevcli/
```

### 启动引擎

```bash
# 复制默认配置
cp tools/source/gameagentengine.conf.yaml .

# 启动服务
GameAgentEngine serve
```

### 创建第一个世界

```bash
GameAgentDevCli node create --type world --name "新世界"
```

### 打开 Creator

```bash
GameAgentDevCli inspect
```

> 如果你计划使用 `world tick`、时间线推进和世界线推理，请先在 Creator 的 Settings 页面配置 `world_time_settings`。否则这部分能力虽然接口存在，但世界时间系统本身没有定义，开发流程会被刻意阻塞以提醒补配置。

---

## 工具链

### GameAgentCreator

Web 图形化编辑器，随引擎打包分发。已支持的页面：

- **Worlds** — 世界选择、创建、重命名，节点树拖拽浏览
- **Snapshots** — 快照保存、校验、恢复、删除
- **Tasks** — 运行时任务查看（状态、分类、重试次数）
- **Plans** — 待审批计划管理
- **Policy** — 世界策略配置
- **Settings** — 世界设置与 `world_time_settings` 编辑
- **Continuity** — 连续性调试包查看
- **State** — 连续性状态组件（world_state / story_state / story_history / tick_policy / world_time_state）
- **Timelines** — 时间线归档
- **Logs** — 推理日志
- **Traces** — 调试轨迹

### GameAgentDevCli

命令行开发工具，通过 HTTP API 操作 Engine：

```bash
# 节点管理
GameAgentDevCli node create --type world --name "新世界"
GameAgentDevCli node list --world <world-id>
GameAgentDevCli node get <node-id>
GameAgentDevCli node update <node-id> --name "重命名"

# 组件 / 记忆 / 关系
GameAgentDevCli component add <node-id> --type autonomous
GameAgentDevCli memory add <node-id> --content "..." --level normal
GameAgentDevCli relation create --source <id> --target <id> --type ally

# 世界推进
GameAgentDevCli world tick <world-id>
GameAgentDevCli world settings get <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>

# 任务管理（Pull 模式）
GameAgentDevCli task list --status pending --limit 20
GameAgentDevCli task claim <task-id> --consumer gamer
GameAgentDevCli task start <task-id> <lease-token>
GameAgentDevCli task heartbeat <task-id> <lease-token>
GameAgentDevCli task release <task-id> <lease-token>

# 调试
GameAgentDevCli logs --world <world-id> --limit 10
GameAgentDevCli debug traces --world <world-id>
GameAgentDevCli inspect
```

### Go SDK

Go SDK 封装了 Engine 全部 HTTP API，适用于在 Go 服务端集成。支持：

- 世界设置与 `world_time_settings` 的读取和部分更新
- 节点、组件、记忆、关系的 CRUD
- 世界 Tick、事件影响、计划审批、快照相关接口
- 连续性状态组件与时间线归档访问
- 运行时任务管理（ListRuntimeTasks / ClaimRuntimeTask / StartRuntimeTask / HeartbeatRuntimeTask / ReleaseRuntimeTask）
- 记忆传播（PropagateMemory）
- 自主行为配置（GetAutonomousConfig / SetAutonomousConfig / RunAutonomousNode）

```go
import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

client := sdk.NewClient("http://127.0.0.1:8080", "dev-key")
tasks, _ := client.ListRuntimeTasks("", "pending", 20)
```

---

## 文档索引

- [入门指南](./docs/GETTING_STARTED.md)
- [架构说明](./docs/ARCHITECTURE.md)
- [核心概念](./docs/CORE_CONCEPTS.md)
- [API 参考](./docs/API_REFERENCE.md)
- [外部交互路线图](./docs/EXTERNAL_INTERACTION_ROADMAP.md)
- [外部交互示例](./docs/EXTERNAL_INTERACTION_EXAMPLES.md)
- [GameAgentDevCli 指南](./docs/GUIDE_GAMEAGENTDEVCLI.md)
- [GameAgentCreator 指南](./docs/GUIDE_GAMEAGENTCREATOR.md)
- [SDK 参考](./docs/SDK_REFERENCE.md)
- [配置参考](./docs/CONFIGURATION.md)
- [自主行为系统](./docs/AUTONOMOUS_BEHAVIOR.md)
- [推理管线内部实现](./docs/PIPELINE_INTERNALS.md)
- [世界时间 Tick 参考](./docs/WORLD_TIME_TICK_REFERENCE.md)
- [构建与部署](./docs/BUILD_AND_DEPLOY.md)

---

## 许可证

MIT
