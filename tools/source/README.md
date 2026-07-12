# GameAgentEngine

**中文** | [**English**](./README_EN.md)

这份 README 随打包产物一起分发。

---

## 核心能力

- **世界图模型** — 基于节点、组件、记忆、关系的统一世界建模
- **LLM 推理管线** — 三档管线模式（vertical / polling / full），支持多轮推理、子任务 DAG、异步动作回调
- **世界时间与连续性** — Tick 推进、连续性状态组件、时间线归档、世界时间刻度系统
- **外部交互与运行时任务** — Push / Pull / Hybrid 三种异步任务投递模式，支持 HTTP / RPC / WebSocket
- **记忆传播** — 五种传播模式（upward / environment_scope / organization_scope / tag_broadcast / targeted）
- **工作副本与快照** — Fork、Snapshot、Restore 三类世界复制语义
- **后台自主行为调度** — 节点级 autonomous 组件，支持手动触发、Tick 同步、定时调度
- **数据库管线** — 统一写事务、批量日志、可重试写层；支持 SQLite / MySQL / PostgreSQL
- **三套开发者工具** — GameAgentCreator（Web 编辑器）、GameAgentDevCli（命令行工具）、Go SDK

---

## 最短上手路径

```bash
# 1. 启动引擎
GameAgentEngine serve

# 2. 创建世界
GameAgentDevCli node create --type world --name "新世界"

# 3. 打开 Creator
GameAgentDevCli inspect
```

> 如果你要使用时间推进和世界线推理，请先配置 `world_time_settings`。

---

## 工具链

### GameAgentCreator

Web 图形化编辑器，已支持的页面：Worlds / Snapshots / Tasks / Plans / Policy / Settings / Continuity / State / Timelines / Logs / Traces

### GameAgentDevCli

命令行开发工具：节点、组件、记忆、关系 CRUD；世界 Tick、事件影响、计划审批、快照管理；任务管理（task 命令）；调试与观测

### Go SDK

Go 服务端 SDK：完整 API 封装，包括运行时任务管理、记忆传播、自主行为配置等

---

## 文档

各文档位于 `docs/` 目录下：

- 入门指南 / API 参考 / 配置参考
- DevCli 指南 / Creator 指南 / SDK 参考
- 架构说明 / 核心概念 / 推理管线内部实现
- 自主行为系统 / 世界时间 Tick 参考 / 外部交互路线图
- 构建与部署

---

## 许可证

MIT
