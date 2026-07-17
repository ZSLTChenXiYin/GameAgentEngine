# GameAgentDevCli 指南

[**中文**] | [**English**](./GUIDE_GAMEAGENTDEVCLI_EN.md)

GameAgentDevCli 是通过 HTTP API 操作 GameAgentEngine 的命令行工具。

---

## 当前能力

- 节点、组件、记忆、关系 CRUD
- world 导入、导出、快照、恢复与验证
- 世界设置、世界策略、计划审批
- world tick、事件影响、scope advance、timeline replan
- continuity 状态组件和 timeline 归档访问
- logs、traces、continuity 调试、node graph 调试
- 打开 Creator
- runtime task 管理
- verify / action callback 等运行时辅助入口

---

## 从零开始

推荐先创建一个 `world` 根节点：

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
GameAgentDevCli world settings set <world-id> --world-time-settings-json '{"tick_scale_mode":"flexible","tick_min_unit":"hour","tick_step":1,"tick_units":["day","hour"]}'
```

如果 `world_time_settings` 缺失，依赖世界时间推进的流程会被阻断，这是当前设计的有意约束。

---

## 推进 Tick

```bash
GameAgentDevCli world tick <world-id>
GameAgentDevCli world tick <world-id> --type manual --time "day-1" --requested-ticks 1
GameAgentDevCli world tick <world-id> --autonomous-limit 2
```

如果 `tick_scale_mode` 是 `fixed`，`requested_ticks` 必须保持为 `1`。

---

## Continuity 与时间线

```bash
GameAgentDevCli state list <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli timeline list <world-id> --limit 5
GameAgentDevCli debug continuity <world-id>
```

排查 paused 多轮执行或 callback 恢复链路时，建议按下面顺序看：

1. `GameAgentDevCli task inspect <task-id>`
2. `GameAgentDevCli debug continuity <world-id>`
3. `GameAgentDevCli debug traces --world <world-id> --limit 10`

---

## 调试与观察

```bash
GameAgentDevCli logs --world <world-id> --limit 10
GameAgentDevCli debug traces --world <world-id> --limit 10
GameAgentDevCli debug node-graph <node-id>
GameAgentDevCli status
GameAgentDevCli version
```

如果只是先确认服务是否在线、鉴权是否正确，优先跑：

```bash
GameAgentDevCli status
GameAgentDevCli version
```

---

## 发起 Invoke

`invoke` 是最直接的单次推理入口，适合不用自己写客户端时快速发请求。

```bash
GameAgentDevCli invoke <world-id> <node-id> \
  --task-type npc_dialogue \
  --message "南门刚刚发生了什么？"
```

也可以直接传请求级 pipeline mode 和动态接口白名单：

```bash
GameAgentDevCli invoke <world-id> <node-id> \
  --task-type npc_dialogue \
  --message "回答前先查询附近场景和任务状态。" \
  --pipeline-mode full \
  --dynamic-interfaces-file dynamic-interfaces.json
```

`dynamic-interfaces.json` 需要是 JSON 数组。它适合描述当前回合、当前 NPC 对话、当前场景下临时开放给 LLM 的查询或动作接口。

---

## Runtime Task 管理

Runtime Task 用于承载 Engine 与游戏侧之间的外部交互，支持 `push`、`pull`、`hybrid` 三种投递模式。下面这些命令主要面向 `pull` 工作流：

```bash
GameAgentDevCli task list --status pending --limit 20
GameAgentDevCli task list --pending-only --consumer game_client --limit 20
GameAgentDevCli task get <task-id>
GameAgentDevCli task get <task-id> --json
GameAgentDevCli task inspect <task-id>
GameAgentDevCli task stats
GameAgentDevCli task claim <task-id> --consumer gamer --owner devcli
GameAgentDevCli task start <task-id> <lease-token>
GameAgentDevCli task heartbeat <task-id> <lease-token>
GameAgentDevCli task release <task-id> <lease-token> --reason "completed"
GameAgentDevCli task requeue <task-id> --retry-delay-ms 1500 --reason "manual requeue"
```

`task inspect` 适合排查 callback / resume 相关问题，它会集中打印：

- `callback_id`
- `resume_execution_id`
- payload 摘要
- dispatch decision / failure class / transition reason
- dispatched / claimed / heartbeat / completed 时间线索

`task stats` 适合看队列健康度，它会汇总：

- total / ready pull / in flight / terminal
- heartbeat timeout
- retry exhausted
- stale dispatched
- 聚合分布统计

---

## 最小 Pull Worker 闭环

如果游戏侧采用 `pull` 模式消费任务，最小闭环一般是：

1. `task list --pending-only --consumer <consumer>` 或直接调 `/api/v1/runtime/tasks/pending`
2. `task claim`
3. `task start`
4. 执行游戏侧查询或动作
5. 回调结果
6. 如果执行时间较长，期间持续 `heartbeat`

当前 callback 响应已经带有结构化 post-process 信息。SDK 或自定义 worker 可以据此判断：

- 这次 callback 只是完成了任务
- 这次 callback 触发了 paused execution 自动恢复
- 这次 callback 还触发了统一后处理，例如 `write_memory`

---

## 打开 Creator

```bash
GameAgentDevCli creator
```

`creator` 现在就是 DevCli 侧正式的浏览器入口。此前一些“inspect/打开编辑器”类历史命名已经不再作为主工作流保留。

---

## 导入、导出与校验

当前围绕世界资产的主命令已经集中在 DevCli：

```bash
GameAgentDevCli import tools/source/workerhome/demo/demo-world.yaml
GameAgentDevCli world export <world-id> --format yaml --out exported-world.yaml
GameAgentDevCli world snapshot <world-id> --out runtime-snapshot.json
GameAgentDevCli world save <world-id> demo-save
GameAgentDevCli world restore <snapshot-world-id> restored-world
GameAgentDevCli world validate-snapshot <snapshot-world-id>
```

约束上应这样理解：

- Engine 保持运行时内核，不承担这些外围工作流入口
- 资产导入导出、快照校验、演示验证属于 DevCli 职责
- 游戏侧异步闭环和 REPL 试玩属于 Worker 职责


---

## Supplementary Notes

### Legacy Alias Compatibility

The following two commands should be treated as equivalent list entrypoints:

```bash
GameAgentDevCli node list --world <world-id>
GameAgentDevCli nodes --world <world-id>
```

`nodes` is a legacy root alias. It is expected to honor the same `--world`, `--limit`, `--offset`, and `--type` flags as `node list`.

### Dynamic Interface Input Guidance

For game-side callable interfaces, keep the boundary simple:

- stable and globally available interfaces belong in Engine `external_interfaces` config
- temporary per-turn or per-NPC-turn interfaces belong in `dynamic_interfaces`
- prefer structured `dynamic_interfaces` / function fields over writing callable interface contracts directly into prompt text

### When DevCli Output Differs From Direct HTTP

If direct HTTP results look correct but DevCli output looks empty or incomplete, verify in this order:

1. use raw-oriented views first, such as `task get <task-id> --json` or `nodes --world <world-id>`
2. then compare the human-oriented view, such as `task inspect <task-id>` or `node list --world <world-id>`

If step 1 is correct and step 2 is wrong, the problem is usually tooling-side field mapping or flag wiring instead of missing Engine data.
