# GameAgentEngine

**中文** | [**English**](./README_EN.md)

这份 README 随打包产物一起分发。

---

## 为什么游戏需要自己的 Agent Engine

通用 LLM Agent 不适合做游戏：

- **没有世界** — 每次对话都是全新上下文，NPC 没有记忆
- **动作不受控** — LLM 什么都能说，游戏需要校准的动作边界
- **没有世界时间** — 游戏有自己的时钟，Agent 没有时间概念
- **集成成本高** — 每次对接游戏引擎都要重造推送、回调、重试、回退
- **没有策划工具** — 大部分 Agent 框架没有可视化编辑器

GameAgentEngine 就是为解决这些问题设计的。

---

## 核心能力

- **世界图模型** — 节点、组件、记忆、关系四条结构构建游戏世界
- **受控动作系统** — 能力白名单校验，禁止越权操作
- **世界时间与连续性** — Tick 推进 + 连续性状态归档
- **外部交互** — Push / Pull / Hybrid 三种异步任务模式
- **三档推理管线** — vertical / polling / full
- **后台自主行为** — NPC 无人交互时也能主动行动
- **数据库管线** — SQLite / MySQL / PostgreSQL

---

## 最短上手路径

```
GameAgentEngine serve
GameAgentDevCli node create --type world --name "新世界"
GameAgentDevCli inspect
```

时间推进前需要在 Creator 的 Settings 页面配置 `world_time_settings`。

如果你想直接体验仓库附带的文字游戏壳与外部 worker 权威状态配合，可以使用：

```bash
GameAgentEngine serve
GameAgentDevCli import demo-world.yaml
GameAgentWorker play --state-file demo-state.yaml --world-id demo_world --player-node-id player_001
```

---

## 工具链

- **GameAgentCreator** — Web 可视化编辑器（Worlds / Tasks / Plans / Settings / State / Timelines / Logs / Traces）
- **GameAgentDevCli** — 命令行开发工具（CRUD + 任务管理 + 调试）
- **Go SDK** — Go 服务端集成（全部 API 封装）

---

## 文档

详见 `docs/` 目录：

- 入门指南 / API 参考 / 配置参考
- DevCli 指南 / Creator 指南 / SDK 参考
- 外部交互路线图 / 外部交互示例
- 架构说明 / 核心概念 / 推理管线内部实现
- 自主行为系统 / 世界时间 Tick 参考 / 构建与部署

---

## 许可证

MIT
