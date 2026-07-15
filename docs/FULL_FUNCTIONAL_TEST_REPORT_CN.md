# 全功能测试报告

[**中文**] | [**English**](./FULL_FUNCTIONAL_TEST_REPORT.md)

## 运行元数据

- git 修订: `75e6ead8a15446514f0a21cf1be2623676b8db87`
- 配置文件: `C:\Users\808\AppData\Local\Temp\gae-s4-src-20260715122600\gameagentengine.conf.yaml`
- 数据库隔离: 使用独立临时 sqlite 数据库 `C:\Users\808\AppData\Local\Temp\gae-s4-src-20260715122600\gameagentengine.db`
- Engine 端口: `18080`
- worker 端口:
- API key 来源: 临时测试配置 `auth.api_key = dev-key`
- callback token 来源: 临时测试配置 `auth.callback_token = dev-callback-token`
- runtime task token 来源: 临时测试配置 `auth.runtime_task_token = dev-task-token`
- 运行模式: source-built
- 执行日期: `2026-07-15`

## 阶段摘要

| 阶段 | 状态 | 说明 |
|---|---|---|
| S0 基线 | 已完成 | 已补充计划文档 |
| S1 测试 worker | 已完成 | 已新增 `cmd/gameagentworker`（并保留旧 `cmd/gameagenttestworker` 兼容入口） |
| S2 Worker 自验证 | 已完成 | worker 单测与构建通过 |
| S3 自动化回归 | 已完成 | `go test ./...` 通过 |
| S4 基础数据面 | 已完成 | `docs/tests/full_functional_base_data.ps1` 在隔离的 source-built Engine 上通过 |
| S5 世界演化与连续性 | 已完成 | `docs/tests/full_functional_continuity.ps1` 在隔离的 mock-provider Engine 上通过 |
| S6 Runtime Task 投递 | 已完成 | `docs/tests/full_functional_runtime_tasks.ps1` 在隔离的 fixture-provider Engine、本地 push receiver 与 pull worker 环境上通过 |
| S7 Callback/Resume 编排 | 已完成 | `docs/tests/full_functional_callback_resume.ps1` 在隔离的 fixture-provider Engine 上通过 |
| S8 工具链冒烟 | 已完成 | `docs/tests/full_functional_tooling_smoke.ps1` 在隔离的 fixture-provider Engine 上通过 |
| S9 机器式场景 | 已完成 | `docs/tests/full_functional_machine_scenario.ps1` 在隔离的 fixture-provider Engine 上通过 |
| S10 最终报告 | 已完成 | 下文已汇总最终通过/失败矩阵、复现说明与操作性缺口 |

## 自动化回归

- 命令: `go test ./...`
- 结果: 通过
- 失败项: 无

## 优先回归验证

| 项目 | 状态 | 证据 |
|---|---|---|
| `task inspect` 返回完整字段 | 已完成 | `docs/tests/full_functional_runtime_tasks_result.json` 以及 `GameAgentDevCli task inspect <hybrid-task-id>` 显示了完整 `payload`、`dispatch_decision=fallback_to_pull` 和状态转换时间戳 |
| `nodes --world` 与直接 HTTP 查询一致 | 已完成 | `docs/tests/full_functional_base_data_result.json` 中 `legacy list parity` 通过，`count=5` |
| callback resume 不会重复发出重复 `data_request` | 已完成 | `docs/tests/full_functional_callback_resume_result.json` 以及 `/api/v1/logs?event_name=data_request_reused` 显示仅有一次 reuse 事件，恢复链路中只生成了一个 `game_client_request_data` runtime task |
| `POST /api/v1/components` 避免 `world_settings` 重复创建竞态 | 已完成 | 在新 world 上并发创建 component 成功，`docs/tests/full_functional_base_data_result.json` 中 `count=6` |

## 基础数据面结果

| 区域 | HTTP | DevCli | 说明 |
|---|---|---|---|
| Node CRUD | 通过 | 通过 | 完成 create/update/delete，并验证 world `5a9b0231-dc1e-4a48-8695-cd30990debb3` 上 legacy `nodes --world` 一致性 |
| Component CRUD | 通过 | 通过 | create/update/delete 通过；新 world 上的并发创建压力验证也通过 |
| Memory CRUD | 通过 | 通过 | create/update/delete 在 HTTP 与 DevCli 之间交叉验证 |
| Relation CRUD | 通过 | 通过 | create/update/delete 在 HTTP 与 DevCli 之间交叉验证 |
| World settings | 通过 | 通过 | DevCli 设置与 HTTP 获取结果一致，`pipeline_mode=polling`、`memory_limit=24` |
| World policy | 通过 | 通过 | HTTP 设置与 DevCli 获取结果一致，`blocked=spawn_item` |

## 基础数据面执行说明

- 脚本: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_base_data.ps1 -EngineBaseUrl http://127.0.0.1:18080 -DevCliPath <temp>\GameAgentDevCli.exe -OutFile docs\tests\full_functional_base_data_result.json`
- 夹具: `docs/tests/full_functional_base_data_world.yaml`
- 结果产物: `docs/tests/full_functional_base_data_result.json`
- 运行后缀: `20260715122604`
- 主 world id: `5a9b0231-dc1e-4a48-8695-cd30990debb3`
- 压测 world id: `8f44662c-3b84-4c44-b65f-2a42d5fb00f0`

## 世界演化与连续性结果

- world tick: 通过；`advanced_ticks=2`，request id 为 `6c33c53b-ec44-4423-8144-8f841920cf91`
- timeline latest/list: 通过；最新 tick 为 `#1`，latest 与 list 首项一致，timeline payload 包含 `world_time_state`
- state list/get: 通过；continuity state 包含 `world_state`、`story_state`、`story_history`、`tick_policy`、`world_time_state`
- debug continuity: 通过；按 request 范围查询的 continuity bundle 返回 `logs=2`、`traces=1`
- logs/traces 关联性: 通过；按 request 范围查询的 `logs` 与 `debug traces` 都匹配 `6c33c53b-ec44-4423-8144-8f841920cf91`

## 世界演化执行说明

- 脚本: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_continuity.ps1 -EngineBaseUrl http://127.0.0.1:18081 -DevCliPath <temp>\GameAgentDevCli.exe -BaseDataResultPath docs\tests\full_functional_base_data_result_s5.json -OutFile docs\tests\full_functional_continuity_result.json`
- 临时配置: `C:\Users\808\AppData\Local\Temp\gae-s5-src-20260715123313\gameagentengine.conf.yaml`
- 运行模式: source-built Engine + mock provider
- 结果产物: `docs/tests/full_functional_continuity_result.json`
- world id: `7be14a95-387d-4280-8011-b02ed444c0c1`
- request id: `6c33c53b-ec44-4423-8144-8f841920cf91`
- 最新世界时间标签: `Cycle历 12day 10hour`

## Runtime Task 投递结果

- push: 通过；`spawn_item` push task `9e6067e9-2c9b-4d4e-a56d-cc01c1dfa3bd` 成功派发到本地 push receiver，并成功完成 callback
- pull: 通过；`npc_trade_action` pull task `f3dc112a-0066-40a6-b6a9-4ea1dcf690a2` 被 `GameAgentWorker pull-once` 成功 claim 并完成
- hybrid fallback: 通过；`spawn_item` hybrid task `bffec0a1-fefc-4388-8041-d354f87743f8` 在 push 派发网络失败后回落为 `released` 且 `transport=task_pull`
- claim/start/heartbeat: 通过；手工任务 `021a7acb-7318-44e5-a10c-448c30e1b1d7` 完成 `claimed -> running` 转换，并成功接收显式 heartbeat
- release/requeue: 通过；手工 release 将任务退回 `released`，超时任务 `6abc550e-53f6-4000-851e-c204d3d4d691` 在显式 requeue 后完成 `heartbeat_timeout -> released`
- stats/inspect: 通过；`task stats` 输出中包含 `fallback_to_pull`，`task inspect` 显示 hybrid task 的完整 payload/dispatch/timestamp 字段

## Runtime Task 投递执行说明

- 脚本: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_runtime_tasks.ps1 -EngineExePath .\tmp\s6\GameAgentEngine.exe -DevCliPath .\tmp\s6\GameAgentDevCli.exe -WorkerExePath .\tmp\s6\GameAgentWorker.exe -OutFile docs\tests\full_functional_runtime_tasks_result.json`
- 临时配置: `C:\Users\808\AppData\Local\Temp\gae-s6-src-20260715125246\gameagentengine.conf.yaml`
- 临时数据库: `C:\Users\808\AppData\Local\Temp\gae-s6-src-20260715125246\gameagentengine.db`
- 运行模式: source-built Engine + fixture provider + 本地 push receiver + pull worker
- 结果产物: `docs/tests/full_functional_runtime_tasks_result.json`
- world id: `35a1fc1d-73c0-4292-9615-6b6d55351890`
- engine / worker 端口: `18082` / `19000`

## Callback 与 Resume 结果

- callback success: 通过；callback `3eb9fea7-5b5b-44d1-acd5-21e9f6af4701` 返回 `resumed.reply=scene-resumed-final`，runtime task `f513873b-d582-407a-a346-ab8fc4fe4db4` 最终为 `succeeded`
- callback failure: 通过；callback `bf36c9d7-34e6-4e42-8642-6c678e63cf7c` 将 runtime task `225cea56-2b56-480e-8ced-34607c49546c` 置为 `failed`，且持久化 callback payload 到 `error_message`
- paused execution auto-resume: 通过；callback success 生成一条 `resume_completed` 日志，并返回非空 `resume_execution_id`
- `resume_policy = none`: 通过；callback `63000dec-5d06-4941-844b-7e8fb6646fa9` 完成任务 `2b40a732-6ef6-44d0-b09d-dae8c75925d0` 时未返回 `resumed` payload，也没有第二条 `resume_completed` 日志
- replay protection: 通过；第二次带 `X-Callback-Request-Id=s7-scene-1` 的 callback 请求返回 `X-Callback-Replayed=true`，且没有重复生成 resume 日志
- `record_only`: 通过；callback `78ed85ca-0e37-43bd-a08d-f7b3c76eb310` 完成任务 `e7fed65d-df9f-4013-9cad-d402f1a45fd7` 时 `post_process_applied=false`，且没有写入 memory
- `write_memory`: 通过；callback `e1d0dda5-a23e-4e43-9b7b-c55174c022e1` 完成任务 `633d032f-5190-4c91-9eb6-145e22389088`，并写入长期记忆 `8e790632-f162-4dc5-ad56-3d9ffa97dbad`
- duplicate query suppression: 通过；恢复后的链路只产生一条 `data_request_reused` 日志，并将 `game_client_request_data` runtime task 数量保持在 `1`

## Callback 与 Resume 执行说明

- 脚本: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_callback_resume.ps1 -EngineExePath .\tmp\s7\GameAgentEngine.exe -DevCliPath .\tmp\s7\GameAgentDevCli.exe -WorkerExePath .\tmp\s7\GameAgentWorker.exe -OutFile docs\tests\full_functional_callback_resume_result.json`
- 临时配置: `C:\Users\808\AppData\Local\Temp\gae-s7-src-20260715130743\gameagentengine.conf.yaml`
- 临时数据库: `C:\Users\808\AppData\Local\Temp\gae-s7-src-20260715130743\gameagentengine.db`
- 运行模式: source-built Engine + fixture provider + 直接 callback HTTP + pull worker
- 结果产物: `docs/tests/full_functional_callback_resume_result.json`
- world id: `936a0329-27ef-492d-9dd0-1ce5ef277ea7`
- engine 端口: `18083`

## 工具链冒烟结果

- SDK: 通过；`docs/tests/sdk_tooling_smoke.go` 针对在线 Engine 验证了 pending runtime task `51995f94-8cb5-4762-b867-adf437c685a1`、latest timeline tick `1`、continuity logs 与 traces
- DevCli: 通过；`node list` 与 legacy `nodes --world` 均保持 `count=2`，`task get 51995f94-8cb5-4762-b867-adf437c685a1 --json` 与 HTTP 任务字段一致
- Creator Tasks: 通过；页面使用的数据源 `/api/v1/runtime/tasks` 与 `/api/v1/runtime/tasks/stats` 返回了同一个 pending task `51995f94-8cb5-4762-b867-adf437c685a1`，统计总数为 `1`
- Creator Continuity: 通过；页面使用的数据源 `timelines/latest`、`timelines`、`state-components`、`logs` 与 `debug/traces` 返回了一次 world tick，`state_components=6`
- Creator Traces: 通过；页面使用的数据源 `/debug/traces?world_id=<world>&limit=30` 返回 `trace_count=2`

## 工具链冒烟执行说明

- 脚本: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_tooling_smoke.ps1 -EngineExePath .\tmp\s8\GameAgentEngine.exe -DevCliPath .\tmp\s8\GameAgentDevCli.exe -WorkerExePath .\tmp\s8\GameAgentWorker.exe -OutFile docs\tests\full_functional_tooling_smoke_result.json`
- SDK helper: `go run .\docs\tests\sdk_tooling_smoke.go --server http://127.0.0.1:18084 --key dev-key --world 6327c45d-bec7-4cbc-b7fa-3b94250e59d7 --node 088baebd-ca32-4a04-9626-b4a939089d42`
- 临时配置: `C:\Users\808\AppData\Local\Temp\gae-s8-src-20260715131652\gameagentengine.conf.yaml`
- 临时数据库: `C:\Users\808\AppData\Local\Temp\gae-s8-src-20260715131652\gameagentengine.db`
- 运行模式: source-built Engine + fixture provider + SDK/DevCli/API 工具链冒烟
- 结果产物: `docs/tests/full_functional_tooling_smoke_result.json`
- world id: `6327c45d-bec7-4cbc-b7fa-3b94250e59d7`
- engine 端口: `18084`

## 机器式场景结果

- invoke: 通过；对 NPC `83443764-bfaf-475b-97dd-80c78664adcc` 发起的 `POST /api/v1/invoke` 在 request `d6f5343f-0633-4873-a3ab-050ba781f5fb` 下生成了 request-scoped callback `df578b37-f1ac-4908-aa8d-3b47f2053f78`
- runtime task creation: 通过；Engine 为接口 `game_client_request_data` 持久化了 pull task `3532df57-9cab-4e12-b760-b2088362f667`
- worker callback: 通过；`GameAgentWorker pull-once --consumer game_client` 成功 claim 该任务，并完成 callback，`resume_execution_id=651ed3c1-b50e-46a3-902d-d13841c6c55d`
- paused execution resume: 通过；按 request 范围查询的日志中有一条 `data_request_paused_for_client` 和一条 `resume_completed`
- observability artifacts: 通过；`debug continuity --request-id` 返回 `logs=1`、`traces=1`、`state_components=6`；由于该场景是直接写入 `world_time_state` 而没有执行 world tick，因此 `latest_timeline` 不存在

## 机器式场景执行说明

- 脚本: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_machine_scenario.ps1 -EngineExePath .\tmp\s9\GameAgentEngine.exe -DevCliPath .\tmp\s9\GameAgentDevCli.exe -WorkerExePath .\tmp\s9\GameAgentWorker.exe -OutFile docs\tests\full_functional_machine_scenario_result.json`
- 临时配置: `C:\Users\808\AppData\Local\Temp\gae-s9-src-20260715132417\gameagentengine.conf.yaml`
- 临时数据库: `C:\Users\808\AppData\Local\Temp\gae-s9-src-20260715132417\gameagentengine.db`
- 运行模式: source-built Engine + fixture provider + pull worker
- 结果产物: `docs/tests/full_functional_machine_scenario_result.json`
- world id: `9abbb007-17c9-4de8-9367-6a72ae32b0d4`
- engine 端口: `18085`

## 失败项与后续跟进

| 严重度 | 区域 | 现象 | 复现方式 | 说明 |
|---|---|---|---|---|
| P3 | 机器场景可观测性 | 当场景直接写入 `world_time_state` 且跳过 `world tick` 时，`debug continuity --request-id` 不包含 `latest_timeline` | 运行 `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_machine_scenario.ps1 -EngineExePath .\tmp\s9\GameAgentEngine.exe -DevCliPath .\tmp\s9\GameAgentDevCli.exe -WorkerExePath .\tmp\s9\GameAgentWorker.exe -OutFile docs\tests\full_functional_machine_scenario_result.json`，再检查 `latest_timeline_present` | 这不是 resume 流程的代码缺陷；当前机器场景为了保持 fixture 的确定性，选择用 state/logs/traces 验证 continuity，而不是依赖 timeline 行 |

## 最终评估

- overall status: `S0-S10` 全部完成。当前 source-built 全功能验证流程可以端到端重复执行，计划中列出的所有优先回归路径均已覆盖且通过。
- blocking issues: 在本次验证过的 source-built 流程中，没有阻塞性问题。
- non-blocking issues: 机器式 NPC 场景当前不会生成 timeline 行，因为它直接写入 `world_time_state`，以避免在 request-scoped callback/resume 之前消耗 fixture 响应。
- operational gaps: Creator 冒烟验证目前仍是 API 层验证，因为本轮没有可用的浏览器自动化能力；打包发行版的端到端验证仍需单独跑一轮 bundled runtime，而不是只验证 source-built 隔离二进制。
