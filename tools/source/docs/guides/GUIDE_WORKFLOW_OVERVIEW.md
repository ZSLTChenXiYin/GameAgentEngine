# 工作流总览

**中文** | [**English**](./GUIDE_WORKFLOW_OVERVIEW_EN.md)

这份文档用于统一说明当前项目中 `Engine`、`DevCli`、`Worker`、`Creator` 和各语言 SDK 的职责边界，以及它们如何在真实工作流里串起来。

它不是架构设计细节文档，而是面向开发与联调的操作视角总览。

---

## 1. 五个主角色

| 组件 | 职责 |
| --- | --- |
| `GameAgentEngine` | 世界建模、NPC 推理、记忆、关系、时间推进、外部异步任务编排 |
| `GameAgentDevCli` | 导入配置、管理世界、查看状态、调试时间线与任务、打开 Creator |
| `GameAgentWorker` | 游戏侧 Worker、权威状态宿主、push/pull/callback 闭环、play REPL、内置测试场景 |
| `GameAgentCreator` | 浏览器可视化编辑与观察界面 |
| `SDKs` | 让外部程序以代码方式接入 Engine / Worker 工作流 |

当前项目的核心判断标准很简单：

- 推理、世界状态演化、异步任务编排属于 Engine
- 游戏侧权威真值、高频查询、外部接口消费属于 Worker / 游戏侧
- 诊断和操作入口属于 DevCli / Creator / SDK

---

## 2. 最短开发闭环

如果只是验证项目能跑通，当前推荐的最短闭环是：

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/demo-world.yaml
GameAgentDevCli creator
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

对应关系是：

1. `Engine` 启动服务
2. `DevCli` 导入世界
3. `Creator` 负责可视化查看与编辑
4. `Worker play` 负责游戏侧权威状态与 NPC 互动体验

---

## 3. 集成测试闭环

如果目标是验证外部异步接口，而不是试玩 REPL，推荐闭环是：

```bash
GameAgentEngine serve
GameAgentWorker serve --verbose
```

或者只跑内置 Worker 测试：

```bash
GameAgentWorker test all
```

这一层里：

- Engine 负责产出 runtime task
- Worker 负责消费 push / pull 任务并回调
- DevCli / SDK 负责观察和驱动具体测试步骤

---

## 4. `play` 不是原始壳层

当前 `GameAgentWorker play` 的定位不是“给 Engine 套一个命令行输入框”，而是：

- 玩家与 NPC 的文字游戏入口
- 游戏侧权威状态读取入口
- 玩家自然语言意图校验与执行的落点
- 体验 Engine 驱动 NPC / 群聊反馈的联调环境

因此：

- `play` 应该归 Worker，不归 Engine 本体
- `play` 依赖 `demo-state.yaml` 这类权威状态文件
- 玩家输入如果要变成真实行为，必须经过游戏侧真值校验

---

## 5. SDK 在工作流中的位置

当前多语言 SDK 不是为了替代 DevCli，而是为了给外围程序、脚本、桥接层和引擎插件接入同一套 Engine / Worker 流程。

目前已经进入“实用级”的 SDK：

- `ts-sdk`
- `js-sdk`
- `cs-sdk`
- `gd-sdk`（请求构造级）

它们现在的主要用途是：

- 发起 `invoke`
- 消费 runtime task
- 回调 callback
- 查看 state / timeline / logs / traces
- 与 `GameAgentWorker` 组成最小联调闭环

共享 SDK 说明见：

- `tools/source/sdks/README.md`
- `tools/source/sdks/SDK_FIXTURES.md`
- `tools/source/sdks/SDK_CAPABILITY_MATRIX.md`

---

## 6. 当前推荐的工具分工

### 配置与导入

- 优先用 `GameAgentDevCli`
- 大批量结构化内容可以直接导入 YAML / JSON

### 可视化编辑与观察

- 优先用 `GameAgentCreator`

### 程序化接入

- 优先用各语言 SDK

### 异步任务联调

- 优先用 `GameAgentWorker`

### 文字游戏体验与权威状态联调

- 优先用 `GameAgentWorker play`

---

## 7. 当前文档导航建议

如果你是按工作流阅读文档，建议顺序是：

1. `docs/getting-started/GETTING_STARTED.md`
2. `docs/guides/GUIDE_GAMEAGENTDEVCLI.md`
3. `docs/guides/GUIDE_GAMEAGENTWORKER.md`
4. `docs/gameplay/PLAYER_INTERACTION.md`
5. `docs/gameplay/GAME_STATE_AUTHORITY.md`
6. `docs/integration/EXTERNAL_INTERACTION.md`
7. `tools/source/sdks/README.md`

这条路径更贴近当前项目的真实使用顺序。
