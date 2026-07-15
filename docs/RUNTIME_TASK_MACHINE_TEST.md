# Runtime Task Machine Test

This document records the repeatable local machine-test flow for validating:

- request-scoped `dynamic_interfaces`
- runtime task production
- push dispatch to a local game-side receiver
- callback completion
- paused execution resume

## Goal

Validate one real end-to-end loop:

1. invoke one NPC dialogue request
2. expose one temporary game-side query interface through `dynamic_interfaces`
3. let the LLM emit a `data_request`
4. let Engine persist and dispatch a runtime task
5. let a local receiver post callback
6. confirm Engine resumes the paused execution

## Test Files

- `docs/tests/runtime_task_push_receiver_example.js`
- `docs/tests/runtime_task_pull_worker_example.js`
- `docs/tests/runtime_task_dynamic_interfaces.json`

## Pre-Check

Before starting Engine, verify the target ports are not occupied by another instance:

```powershell
Get-NetTCPConnection -LocalPort 8080,9000 -State Listen
```

If another Engine instance is already listening on `127.0.0.1:8080`, stop it first. Otherwise requests may hit the wrong database and produce false results.

## Minimal Config Expectations

The machine-test config should include:

- `auth.api_key`
- `auth.callback_token`
- `external_integrations.game_http`
- `external_interfaces.game_client_request_data`

Suggested values for local testing:

- `callback_token = dev-callback-token`
- `runtime_task_token = dev-task-token`
- `game_http.auth.token = local-test-token`

## Test Flow

### 1. Start the local push receiver

```powershell
node docs/tests/runtime_task_push_receiver_example.js
```

### 2. Start Engine

```powershell
GameAgentEngine.exe serve --config gameagentengine.conf.yaml
```

### 3. Create a minimal world

```powershell
GameAgentDevCli.exe node create --type world --name MachineTestWorld
GameAgentDevCli.exe node create --world <world-id> --type npc --name TestMerchant
GameAgentDevCli.exe node create --world <world-id> --type location --name StarterInn
```

### 4. Configure world settings

At minimum, set:

- valid `world_time_settings`
- `pipeline_mode = full`

### 5. Invoke with temporary dynamic interfaces

```powershell
GameAgentDevCli.exe invoke <world-id> <node-id> \
  --task-type npc_dialogue \
  --pipeline-mode full \
  --dynamic-interfaces-file docs/tests/runtime_task_dynamic_interfaces.json \
  --message "Before answering, query the game side for nearby scene facts and world state."
```

### 6. Inspect runtime tasks

```powershell
GameAgentDevCli.exe task list --limit 20
GameAgentDevCli.exe task stats
```

### 7. Inspect logs and continuity

```powershell
GameAgentDevCli.exe logs --world <world-id> --limit 20 --details
GameAgentDevCli.exe debug continuity <world-id>
```

## Expected Success Signals

- invoke returns one async `data_request`
- one runtime task is created
- the local receiver logs one dispatch
- callback succeeds
- logs show `data_request_paused_for_client`
- logs show `resume_completed`

## Known Failure Modes This Flow Detects

- wrong Engine instance on port `8080`
- malformed `dynamic_interfaces` JSON
- runtime task dispatch/auth mismatch
- callback token mismatch
- resume after callback repeatedly issuing the same query
- DevCli task inspection mismatch with actual runtime task state
