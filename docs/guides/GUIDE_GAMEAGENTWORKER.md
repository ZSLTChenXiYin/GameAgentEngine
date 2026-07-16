# GameAgentWorker 指南

**中文** | [**English**](./GUIDE_GAMEAGENTWORKER_EN.md)

`GameAgentWorker` 是项目内正式的游戏侧 Worker 命令行工具，不再只是临时测试脚本集合。

它承担两类职责：

- 集成测试时，闭合 Engine 的外部异步任务 push / pull / callback 链路
- 本地开发与试玩时，承载 YAML / JSON 权威状态，并提供文字游戏式 `play` REPL

---

## 1. 当前支持的命令

```bash
GameAgentWorker serve
GameAgentWorker push-receiver
GameAgentWorker pull-worker
GameAgentWorker pull-once
GameAgentWorker play
GameAgentWorker test <scenario>
```

各命令作用如下：

- `serve`：同时运行 push receiver 和 pull worker
- `push-receiver`：只启动 HTTP 推送接收端，监听 `/api/v1/runtime/dispatch`
- `pull-worker`：只轮询 `/api/v1/runtime/tasks/pending` 并处理任务
- `pull-once`：只处理一个待执行 pull task，适合脚本化调试
- `play`：启动单人文字游戏 REPL，由 Engine 驱动 NPC 响应，由 Worker 承载游戏侧权威状态
- `test <scenario>`：运行内置 Worker 场景测试

当前内置测试场景：

```bash
GameAgentWorker test base-data
GameAgentWorker test continuity
GameAgentWorker test runtime-tasks
GameAgentWorker test callback-resume
GameAgentWorker test tooling-smoke
GameAgentWorker test machine-scenario
GameAgentWorker test all
```

---

## 2. Worker 的定位

当前仓库中，`GameAgentWorker` 是外部游戏侧行为的标准承接面：

- 对 Engine 来说，它是运行时外部接口的消费端
- 对集成测试来说，它是稳定、可重复的 fixture worker
- 对开发者来说，它是本地游戏侧壳层
- 对 `play` 模式来说，它是权威状态宿主，而不是原始引擎外壳

这意味着后续涉及游戏侧异步接口、权威状态读取、试玩 REPL、打包测试工作流时，都应该优先围绕 `GameAgentWorker` 收口，而不是再新增零散脚本工具。

---

## 3. 默认端口与令牌

默认值来自当前实现：

- Engine base URL：`http://127.0.0.1:8080`
- push receiver port：`9000`
- runtime task token：`dev-task-token`
- callback token：`dev-callback-token`
- push bearer token：`local-test-token`
- 默认 consumer：`game_client`
- 默认 lease owner：`gameagentworker`

常用公共参数：

```bash
--engine-base-url
--engine-api-key
--runtime-task-token
--callback-token
--game-http-bearer-token
--state-file
--consumer
--lease-owner
--push-port
--poll-interval
--heartbeat-interval
--callback-delay
--long-task-duration
--fail-interface
--long-task-interface
--verbose
```

---

## 4. 集成测试与异步任务工作流

### 4.1 `serve`

适合本地联调或完整闭环测试：

```bash
GameAgentWorker serve --verbose
```

它会同时：

- 接收 Engine 主动 push 过来的 runtime dispatch
- 轮询 pull 模式待处理任务
- 根据 interface 名称构造确定性 fixture 结果
- 在需要时发送 heartbeat
- 最终向 `/api/v1/actions/callback` 回调结果

### 4.2 `push-receiver`

适合只测 push 分发链路：

```bash
GameAgentWorker push-receiver --push-port 9000
```

当前只接受：

- `POST /api/v1/runtime/dispatch`
- `Authorization: Bearer <game-http-bearer-token>`

### 4.3 `pull-worker`

适合只测 pull 消费：

```bash
GameAgentWorker pull-worker --consumer game_client
```

它会循环执行：

1. 拉取 pending runtime task
2. claim task
3. start task
4. 根据 interface 决定成功 / 失败 / 长任务模拟
5. callback

### 4.4 `pull-once`

适合脚本和单步调试：

```bash
GameAgentWorker pull-once --consumer game_client
```

如果当前没有 pending task，会输出一次 `pull_noop` 日志然后退出。

---

## 5. 失败与长任务模拟

### 强制某接口失败

```bash
GameAgentWorker serve --fail-interface spawn_item
```

### 将某接口模拟为长任务

```bash
GameAgentWorker serve \
  --long-task-interface game_client_request_data \
  --long-task-duration 8s \
  --heartbeat-interval 2s
```

这适合验证：

- heartbeat 是否正常续租
- callback-resume 是否稳定
- Engine 对长任务和超时的处理是否符合预期

---

## 6. play 模式的真实职责

`play` 不是裸 `invoke` 命令包装层，而是一个受约束的文字游戏入口。

当前实现已经具备这些核心边界：

- 游戏高频真值保留在 Worker 侧权威状态文件中
- NPC 响应仍由 Engine 驱动
- Engine 可在必要时通过异步接口 `game_client_request_data` 读取权威数据
- 玩家自然语言输入可以先解释为玩家意图，再做权威校验与执行

也就是说，`play` 当前已经是“游戏侧状态 + Engine 推理”的组合，而不是单纯把一句文本直接丢给引擎。

---

## 7. play 模式启动方式

先启动 Engine 并导入世界：

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/demo-world.yaml
```

再启动 `play`：

```bash
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

`play` 相关参数：

```bash
--state-file
--world-id
--player-node-id
--session-id
--pipeline-mode
--include-related-nodes
--auto-worker
```

其中：

- `--state-file` 必填，用于加载 YAML / JSON 权威状态
- `--world-id` 可从状态文件继承；未提供时会尝试读取 `world_id`
- `--player-node-id` 不填时会自动寻找 `kind=player` 的 actor
- `--auto-worker` 默认开启，会在 play 内嵌一个 pull worker，用来自动完成权威查询 callback

如果关闭 `--auto-worker`，而当前回合又触发了 `game_client_request_data`，那么响应会停在 paused 状态并报错提醒。

---

## 8. play 命令说明

当前 `play` 支持：

```text
/+help
/+look
/+who
/+state
/+inventory
/+quests
/+talk <npc>
/+next_target
/+prev_target
/+target
/+room
/+move <scene>
/+inspect [target]
/+use_item <item>
/+clear_target
/+say <message>
/+ask <npc> <message>
/+act <text>
/+gift <npc> <item>
/+show_item <npc> <item>
/+trade [npc]
/+threaten [npc]
/+exit

# 兼容旧写法
/help
/look
...
```

各命令职责如下：

| 命令 | 作用 |
| --- | --- |
| `/+help` | 查看帮助 |
| `/+look` | 查看当前场景摘要、提示和同场角色 |
| `/+who` | 列出当前场景角色 |
| `/+state` | 查看玩家权威状态摘要，如 HP、金钱、背包、位置 |
| `/+inventory` | 查看背包详情 |
| `/+quests` | 查看任务 / 剧情状态摘要 |
| `/+talk <npc>` | 选择当前私聊对象 |
| `/+next_target` / `/+prev_target` | 在当前场景的可对话 NPC 之间切换 |
| `/+target` | 查看当前对话目标 |
| `/+room` | 查看当前房间参与者和当前群聊主响应者 |
| `/+move <scene>` | 在游戏侧权威状态里执行确定性移动 |
| `/+inspect [target]` | 查看当前场景、角色或可见物品的权威摘要 |
| `/+use_item <item>` | 对当前持有物品执行确定性使用校验 |
| `/+clear_target` | 清除当前对话目标 |
| `/+say <message>` | 面向当前房间公开发言，由当前群聊主响应 NPC 回应 |
| `/+ask <npc> <message>` | 在群聊语境下点名某个 NPC 回应 |
| `/+act <text>` | 将自然语言先映射为玩家意图，再做权威校验、执行与后续 NPC / 群聊响应 |
| `/+gift <npc> <item>` | 先在权威状态里完成赠礼落地，再请求 NPC 给出反馈 |
| `/+show_item <npc> <item>` | 校验玩家确实持有该物品后，再向 NPC 展示 |
| `/+trade [npc]` | 发起交易 / 议价对话 |
| `/+threaten [npc]` | 发起威胁式对话 |
| `/+exit` | 退出 play 模式 |

此外：

- 旧写法 `/help`、`/talk` 这类输入仍兼容。
- 直接输入普通文本时，会把文本发给当前 `/+talk` 选中的 NPC，作为 `direct_dialogue` 处理。

---

## 9. `/act` 的意义

`/act` 是当前实现里最接近“自然语言控制玩家行为”的入口。

它的流程不是“文本即真相”，而是：

1. 先调用 Engine 的玩家输入解释接口
2. 生成玩家意图结构
3. 用 Worker 当前加载的权威状态做校验
4. 仅当校验通过时，才在游戏侧状态中执行
5. 在 play 中显式展示解释结果、缺失事实、建议交互
6. 再将执行后的互动桥接为 NPC 对话或群聊响应

这条链路正是后续玩家自然语言控制的正确基础，因为它保留了：

- 玩家节点在 Engine 中作为正式节点参与建模
- 游戏侧状态仍然是最终权威
- Engine 负责理解、组织和生成交互反馈

---

## 10. 当前权威查询类型

`play` 模式会通过动态接口 `game_client_request_data` 暴露这些查询能力：

- `player_state`
- `player_inventory`
- `player_wallet`
- `player_location`
- `npc_location`
- `scene_state`
- `room_state`
- `task_state`
- `item_presence`

这些查询覆盖当前讨论里最关键的一批高频权威事实：

- HP
- 背包
- 金钱
- 玩家 / NPC 所在地点
- 房间 / 场景即时状态
- 任务状态
- 某物品是否真实存在于玩家身上

---

## 11. 当前实现边界

`play` 已经支持群聊入口，但目前仍有明确边界：

- 群聊回合当前仍由一个主响应 NPC 回应，不是多 NPC 并行推理集群
- 群聊主响应者规则当前是显式的：优先当前 target，否则回退到当前场景中稳定可预测的默认 NPC
- `action_calls` 当前只展示，不在本地直接自动落地
- 高风险自然语言动作仍必须经过权威校验，不能绕过游戏侧真值
- `/act` 当前已显式展示解释出的 intent、steps、missing facts 和 suggested interaction，便于调试自然语言控制链路

所以当前版本适合做：

- 体验 Engine 驱动的类文字游戏交互
- 验证玩家输入解释与权威校验链路
- 验证 NPC 对话、群聊、送礼、展示物品等交互闭环

但还不等同于完整的多角色并行世界模拟器。

---

## 12. 推荐工作流

### 最短试玩路径

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/demo-world.yaml
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

### 最短异步任务联调路径

```bash
GameAgentEngine serve
GameAgentWorker serve --verbose
```

### 最短内置测试路径

```bash
GameAgentWorker test all
```

---

## 13. 与其他工具的职责分工

| 工具 | 当前职责 |
| --- | --- |
| `GameAgentEngine` | 世界建模、NPC 推理、记忆、关系、时间推进、外部异步任务编排 |
| `GameAgentDevCli` | 配置导入、世界管理、运行时调试、任务与时间线诊断、打开 Creator |
| `GameAgentWorker` | 游戏侧 Worker、权威状态承载、异步接口闭环、play REPL、内置集成测试 |
| `GameAgentCreator` | 浏览器可视化编辑与观测界面 |

这也是当前建议长期维持的职责边界。
