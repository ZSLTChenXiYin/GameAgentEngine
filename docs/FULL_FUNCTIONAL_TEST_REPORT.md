# Full Functional Test Report

## Run Metadata

- git revision: `b0600c7c96750f8907f32ad618f12ee9364727f8`
- config file: `C:\Users\808\AppData\Local\Temp\gae-s4-src-20260715122600\gameagentengine.conf.yaml`
- database isolation: isolated temp sqlite db at `C:\Users\808\AppData\Local\Temp\gae-s4-src-20260715122600\gameagentengine.db`
- Engine port: `18080`
- worker port:
- API key source: temp test config `auth.api_key = dev-key`
- callback token source: temp test config `auth.callback_token = dev-callback-token`
- runtime task token source: temp test config `auth.runtime_task_token = dev-task-token`
- runtime mode: source-built
- execution date: `2026-07-15`

## Stage Summary

| Stage | Status | Notes |
|---|---|---|
| S0 Baseline | completed | plan document added |
| S1 Test worker | completed | `cmd/gameagenttestworker` added |
| S2 Worker self-validation | completed | worker unit tests and build passed |
| S3 Automated regression | completed | `go test ./...` passed |
| S4 Base data plane | completed | `docs/tests/full_functional_base_data.ps1` passed against isolated source-built Engine |
| S5 World evolution and continuity | completed | `docs/tests/full_functional_continuity.ps1` passed against isolated mock-provider Engine |
| S6 Runtime task delivery | completed | `docs/tests/full_functional_runtime_tasks.ps1` passed against isolated fixture-provider Engine plus local push receiver and pull worker |
| S7 Callback/resume orchestration | completed | `docs/tests/full_functional_callback_resume.ps1` passed against isolated fixture-provider Engine |
| S8 Tooling smoke | pending | |
| S9 Machine scenario | pending | |
| S10 Final report | pending | |

## Automated Regression

- command: `go test ./...`
- result: passed
- failures: none

## Priority Regression Verification

| Item | Status | Evidence |
|---|---|---|
| `task inspect` populated fields | completed | `docs/tests/full_functional_runtime_tasks_result.json` plus `GameAgentDevCli task inspect <hybrid-task-id>` showed populated `payload`, `dispatch_decision=fallback_to_pull`, and transition timestamps |
| `nodes --world` matches direct HTTP | completed | `docs/tests/full_functional_base_data_result.json` shows `legacy list parity` passed with `count=5` |
| callback resume avoids duplicate `data_request` | completed | `docs/tests/full_functional_callback_resume_result.json` plus `/api/v1/logs?event_name=data_request_reused` showed one reuse event and only one `game_client_request_data` runtime task for the resumed chain |
| `POST /api/v1/components` avoids duplicate `world_settings` race | completed | concurrent component create on fresh world passed with `count=6` in `docs/tests/full_functional_base_data_result.json` |

## Base Data Plane Results

| Area | HTTP | DevCli | Notes |
|---|---|---|---|
| Node CRUD | passed | passed | create/update/delete plus legacy `nodes --world` parity on world `5a9b0231-dc1e-4a48-8695-cd30990debb3` |
| Component CRUD | passed | passed | create/update/delete passed; concurrent create stress on fresh world also passed |
| Memory CRUD | passed | passed | create/update/delete cross-checked between HTTP and DevCli |
| Relation CRUD | passed | passed | create/update/delete cross-checked between HTTP and DevCli |
| World settings | passed | passed | DevCli set and HTTP get matched `pipeline_mode=polling`, `memory_limit=24` |
| World policy | passed | passed | HTTP set and DevCli get matched `blocked=spawn_item` |

## Base Data Plane Execution Notes

- script: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_base_data.ps1 -EngineBaseUrl http://127.0.0.1:18080 -DevCliPath <temp>\GameAgentDevCli.exe -OutFile docs\tests\full_functional_base_data_result.json`
- fixture: `docs/tests/full_functional_base_data_world.yaml`
- result artifact: `docs/tests/full_functional_base_data_result.json`
- run suffix: `20260715122604`
- primary world id: `5a9b0231-dc1e-4a48-8695-cd30990debb3`
- stress world id: `8f44662c-3b84-4c44-b65f-2a42d5fb00f0`

## World Evolution and Continuity Results

- world tick: passed; `advanced_ticks=2`, request id `6c33c53b-ec44-4423-8144-8f841920cf91`
- timeline latest/list: passed; latest tick `#1`, latest/list head matched, timeline payload carried `world_time_state`
- state list/get: passed; continuity state included `world_state`, `story_state`, `story_history`, `tick_policy`, `world_time_state`
- debug continuity: passed; `logs=2`, `traces=1` in request-scoped continuity bundle
- logs/traces correlation: passed; request-scoped `logs` and `debug traces` both matched `6c33c53b-ec44-4423-8144-8f841920cf91`

## World Evolution Execution Notes

- script: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_continuity.ps1 -EngineBaseUrl http://127.0.0.1:18081 -DevCliPath <temp>\GameAgentDevCli.exe -BaseDataResultPath docs\tests\full_functional_base_data_result_s5.json -OutFile docs\tests\full_functional_continuity_result.json`
- temp config: `C:\Users\808\AppData\Local\Temp\gae-s5-src-20260715123313\gameagentengine.conf.yaml`
- runtime mode: source-built Engine + mock provider
- result artifact: `docs/tests/full_functional_continuity_result.json`
- world id: `7be14a95-387d-4280-8011-b02ed444c0c1`
- request id: `6c33c53b-ec44-4423-8144-8f841920cf91`
- latest world time label: `Cycle历 12day 10hour`

## Runtime Task Delivery Results

- push: passed; `spawn_item` push task `9e6067e9-2c9b-4d4e-a56d-cc01c1dfa3bd` dispatched to local push receiver and completed callback successfully
- pull: passed; `npc_trade_action` pull task `f3dc112a-0066-40a6-b6a9-4ea1dcf690a2` was claimed and completed by `GameAgentTestWorker pull-once`
- hybrid fallback: passed; `spawn_item` hybrid task `bffec0a1-fefc-4388-8041-d354f87743f8` fell back to `released` + `transport=task_pull` after push dispatch network failure
- claim/start/heartbeat: passed; manual task `021a7acb-7318-44e5-a10c-448c30e1b1d7` transitioned through `claimed -> running`, then accepted explicit heartbeat
- release/requeue: passed; manual release returned task to `released`, and timeout task `6abc550e-53f6-4000-851e-c204d3d4d691` transitioned through `heartbeat_timeout -> released` after explicit requeue
- stats/inspect: passed; `task stats` output included `fallback_to_pull`, and `task inspect` showed populated payload/dispatch/timestamp fields for the hybrid task

## Runtime Task Delivery Execution Notes

- script: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_runtime_tasks.ps1 -EngineExePath .\tmp\s6\GameAgentEngine.exe -DevCliPath .\tmp\s6\GameAgentDevCli.exe -WorkerExePath .\tmp\s6\GameAgentTestWorker.exe -OutFile docs\tests\full_functional_runtime_tasks_result.json`
- temp config: `C:\Users\808\AppData\Local\Temp\gae-s6-src-20260715125246\gameagentengine.conf.yaml`
- temp db: `C:\Users\808\AppData\Local\Temp\gae-s6-src-20260715125246\gameagentengine.db`
- runtime mode: source-built Engine + fixture provider + local push receiver + pull worker
- result artifact: `docs/tests/full_functional_runtime_tasks_result.json`
- world id: `35a1fc1d-73c0-4292-9615-6b6d55351890`
- engine / worker ports: `18082` / `19000`

## Callback and Resume Results

- callback success: passed; callback `3eb9fea7-5b5b-44d1-acd5-21e9f6af4701` returned `resumed.reply=scene-resumed-final`, runtime task `f513873b-d582-407a-a346-ab8fc4fe4db4` completed to `succeeded`
- callback failure: passed; callback `bf36c9d7-34e6-4e42-8642-6c678e63cf7c` completed runtime task `225cea56-2b56-480e-8ced-34607c49546c` to `failed` with persisted callback payload in `error_message`
- paused execution auto-resume: passed; callback success created one `resume_completed` log and returned non-empty `resume_execution_id`
- `resume_policy = none`: passed; callback `63000dec-5d06-4941-844b-7e8fb6646fa9` completed task `2b40a732-6ef6-44d0-b09d-dae8c75925d0` without `resumed` payload and without a second `resume_completed` log
- replay protection: passed; second callback request with `X-Callback-Request-Id=s7-scene-1` returned `X-Callback-Replayed=true` and did not duplicate resume logs
- `record_only`: passed; callback `78ed85ca-0e37-43bd-a08d-f7b3c76eb310` completed task `e7fed65d-df9f-4013-9cad-d402f1a45fd7` with `post_process_applied=false` and wrote no memory rows
- `write_memory`: passed; callback `e1d0dda5-a23e-4e43-9b7b-c55174c022e1` completed task `633d032f-5190-4c91-9eb6-145e22389088` and wrote long-term memory `8e790632-f162-4dc5-ad56-3d9ffa97dbad`
- duplicate query suppression: passed; resumed chain emitted one `data_request_reused` log and kept `game_client_request_data` runtime task count at `1`

## Callback and Resume Execution Notes

- script: `powershell -NoProfile -ExecutionPolicy Bypass -File .\docs\tests\full_functional_callback_resume.ps1 -EngineExePath .\tmp\s7\GameAgentEngine.exe -DevCliPath .\tmp\s7\GameAgentDevCli.exe -WorkerExePath .\tmp\s7\GameAgentTestWorker.exe -OutFile docs\tests\full_functional_callback_resume_result.json`
- temp config: `C:\Users\808\AppData\Local\Temp\gae-s7-src-20260715130743\gameagentengine.conf.yaml`
- temp db: `C:\Users\808\AppData\Local\Temp\gae-s7-src-20260715130743\gameagentengine.db`
- runtime mode: source-built Engine + fixture provider + direct callback HTTP + pull worker
- result artifact: `docs/tests/full_functional_callback_resume_result.json`
- world id: `936a0329-27ef-492d-9dd0-1ce5ef277ea7`
- engine port: `18083`

## Tooling Smoke Results

- SDK:
- DevCli:
- Creator Tasks:
- Creator Continuity:
- Creator Traces:

## Machine-Style Scenario Results

- invoke:
- runtime task creation:
- worker callback:
- paused execution resume:
- observability artifacts:

## Failures and Follow-Ups

| Severity | Area | Symptom | Reproduction | Notes |
|---|---|---|---|---|

## Final Assessment

- overall status:
- blocking issues:
- non-blocking issues:
- operational gaps:
