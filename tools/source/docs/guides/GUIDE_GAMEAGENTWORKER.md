# GameAgentWorker 指南

`GameAgentWorker` 是随包提供的正式游戏侧 Worker 命令行工具。

它同时用于：

- 外部异步任务 push / pull / callback 联调
- YAML / JSON 权威状态承载
- `play` 文字游戏式 REPL
- 内置 Worker 场景测试

## 支持的命令

```bash
GameAgentWorker serve
GameAgentWorker push-receiver
GameAgentWorker pull-worker
GameAgentWorker pull-once
GameAgentWorker play
GameAgentWorker test <scenario>
```

当前支持的测试场景：

```bash
GameAgentWorker test base-data
GameAgentWorker test continuity
GameAgentWorker test runtime-tasks
GameAgentWorker test callback-resume
GameAgentWorker test tooling-smoke
GameAgentWorker test machine-scenario
GameAgentWorker test all
```

## 常用工作流

完整异步任务闭环：

```bash
GameAgentEngine serve
GameAgentWorker serve --verbose
```

单步处理一个 pull task：

```bash
GameAgentWorker pull-once --consumer game_client
```

进入试玩 REPL：

```bash
GameAgentEngine serve
GameAgentDevCli import demo-world.yaml
GameAgentWorker play --state-file demo-state.yaml --world-id demo_world --player-node-id player_001
```

执行内置测试：

```bash
GameAgentWorker test all
```

## `play` 模式说明

`play` 不是原始引擎壳，而是受约束的文字游戏入口。

当前已经支持：

- `/talk <npc>`：选择私聊目标
- 直接输入文本：向当前私聊目标发送自然语言
- `/say <message>`：房间公开发言
- `/ask <npc> <message>`：群聊语境点名 NPC
- `/act <text>`：先解释玩家意图，再做权威校验与执行
- `/gift <npc> <item>`：先在权威状态中转移物品，再请求 NPC 反馈
- `/show_item <npc> <item>`：校验物品存在后展示给 NPC
- `/trade [npc]` / `/threaten [npc]`

`play` 会通过 `game_client_request_data` 向游戏侧查询这些高频权威数据：

- HP
- 背包
- 金钱
- 玩家 / NPC 位置
- 场景即时状态
- 任务状态
- 物品是否真实存在

当前群聊仍采用单主响应 NPC，每回合不是多 NPC 并行推理。
