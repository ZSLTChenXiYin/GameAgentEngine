# SDK 共享夹具与集成输入

本文件定义 SDK 示例应优先复用的共享夹具文件与集成输入。

目标是让各语言 SDK 示例都对齐到同一套 Engine / Worker 语义，而不是各自漂移到独立的临时样例数据。

## 1. 主夹具目录

共享的 worker 侧与 runtime-task 夹具数据位于：

```text
tools/source/workerhome/fixtures/
```

SDK 示例在新增样例数据前，应优先复用这些文件。

## 2. 核心共享文件

| 文件 | 用途 | 典型使用方 |
| --- | --- | --- |
| `runtime_task_dynamic_interfaces.json` | 用于触发 authority-query runtime task 的 request-scoped `game_client_request_data` 示例 | ts/js/cs SDK worker authority-query 示例，以及后续 Java/Lua/C++ 示例 |
| `runtime_task_dynamic_action_trade.json` | pull 模式 external action 示例 | Worker runtime-task 场景、后续 pull 示例 |
| `runtime_task_delivery_fixture.json` | worker runtime-task 集成测试使用的 LLM 输出夹具 | `GameAgentWorker test runtime-tasks`、后续 SDK 编排文档 |
| `machine_scenario_fixture.json` | 端到端 worker + continuity + callback-resume 夹具 | `GameAgentWorker test machine-scenario`、后续 SDK smoke loop |
| `callback_resume_fixture.json` | callback-resume 基础夹具 | Worker callback-resume 场景 |
| `callback_resume_dynamic_actions.json` | callback-resume 的 dynamic action 后续夹具 | Worker callback-resume 场景 |
| `full_functional_base_data_world.yaml` | worker 侧 full-functional 场景的基础可导入 world 夹具 | Engine / Worker 测试引导 |
| `world_time_settings_flexible.json` | 灵活 world-time settings 样例 | continuity / machine scenario / 后续 SDK settings 示例 |
| `state_world_state.json` | `world_state` 连续性载荷示例 | SDK state-component 示例 |
| `state_story_state.json` | `story_state` 连续性载荷示例 | SDK state-component 示例 |
| `state_story_history.json` | `story_history` 连续性载荷示例 | SDK continuity 示例 |
| `state_tick_policy.json` | `tick_policy` 连续性载荷示例 | SDK continuity 示例 |

## 3. 仓库级 Demo 资产

以下文件不在 `tools/source/workerhome/fixtures` 下，但仍属于共享集成资产：

| 文件 | 用途 |
| --- | --- |
| `tools/source/workerhome/demo/demo-world.yaml` | Engine / DevCli / Worker 快速开始使用的 demo world 导入文件 |
| `tools/source/workerhome/demo/demo-state.yaml` | Worker play 模式的 authority-state 样例 |

## 4. 夹具使用规则

SDK 示例应遵守以下规则：

1. authority-query 示例优先复用 `runtime_task_dynamic_interfaces.json`；
2. 面向 play 模式的演示优先复用 `tools/source/workerhome/demo/demo-world.yaml` 和 `tools/source/workerhome/demo/demo-state.yaml`；
3. 如果共享文件已存在，不要在每种语言示例里继续内嵌大段 JSON；
4. 若某个 SDK 需要新增夹具，仅当它至少还能复用于另一个 SDK 或 Worker 场景时，才加入 `tools/source/workerhome/fixtures/`。

## 5. 当前共享示例模式

当前 practical SDK 示例主要收敛到两种共享模式：

- runtime task pull / claim / start / callback roundtrip
- 通过 `game_client_request_data` 触发 authority query，再移交给 `GameAgentWorker pull-once`

这两种模式是跨语言一致性的当前基线。
