# GameAgentDevCli 指南

**中文** | [**English**](./GUIDE_GAMEAGENTDEVCLI_EN.md)

GameAgentDevCli 是通过 HTTP API 操作 GameAgentEngine 的命令行工具。

---

## 当前能力

- 节点、组件、记忆、关系 CRUD
- 世界设置、世界策略、计划审批
- 世界 Tick、事件影响评估、局部推进、时间线重规划
- 连续性状态组件与时间线归档查看
- 日志、调试轨迹、连续性调试、节点图调试
- 快照保存、校验、恢复、删除
- 打开 Creator
- 任务管理（task 命令）

---

## 零起步创建世界

当前建议从创建一个 `world` 根节点开始：

```bash
GameAgentDevCli node create --type world --name "新世界"
```

然后继续创建子节点：

```bash
GameAgentDevCli node create --world <world-id> --type location --name "起始村庄"
GameAgentDevCli node create --world <world-id> --type npc --name "守门人"
```

---

## 世界设置

```bash
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world settings set <world-id> --pipeline-mode polling
GameAgentDevCli world settings set <world-id> --world-time-settings-file world-time.json
GameAgentDevCli world settings set <world-id> --world-time-settings-json '{"tick_scale_mode":"flexible","tick_min_unit":"时","tick_step":1,"tick_units":["日","时"]}'
```

`world_time_settings` 不设置时，世界时间推进相关流程会被阻塞，这是当前的有意设计。

---

## 推进 Tick

```bash
GameAgentDevCli world tick <world-id>
GameAgentDevCli world tick <world-id> --type manual --time "day-1" --requested-ticks 1
GameAgentDevCli world tick <world-id> --autonomous-limit 2
```

如果 `tick_scale_mode` 是 `fixed`，则 `requested_ticks` 必须等于 1。

---

## 连续性与时间线

```bash
GameAgentDevCli state list <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli timeline list <world-id> --limit 5
GameAgentDevCli debug continuity <world-id>
```

排查时间推进或世界线问题时，优先看：

1. `timeline latest`
2. `state get <world-id> world_time_state`
3. `debug continuity`

---

## 调试与观测

```bash
GameAgentDevCli logs --world <world-id> --limit 10
GameAgentDevCli debug traces --world <world-id> --limit 10
GameAgentDevCli debug node-graph <node-id>
```

---

## 任务管理


运行时任务（Runtime Task）用于管理 Engine 与游戏端之间的外部交互。支持 Push、Pull、Hybrid 三种投递模式。以下命令主要面向 Pull 模式：


```bash
GameAgentDevCli task list --status pending --limit 20
GameAgentDevCli task get <task-id>
GameAgentDevCli task claim <task-id> --consumer gamer --owner devcli
GameAgentDevCli task start <task-id> <lease-token>
GameAgentDevCli task heartbeat <task-id> <lease-token>
GameAgentDevCli task release <task-id> <lease-token> --reason "completed"
```


---

## 打开 Creator