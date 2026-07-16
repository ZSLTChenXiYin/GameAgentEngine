# GameAgentEngine C# SDK

This SDK is the primary C# baseline for GameAgentEngine server-side and Unity-adjacent integration.

## Current Status

This is now a practical first version rather than a raw string-only scaffold.

It still does not provide full Go SDK parity, but it already covers the main Engine / Worker integration surface needed by external C# tooling.

## Current Capability Scope

- health and version
- invoke
- player input interpretation
- runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
- callback
- world settings
- continuity state components
- timelines
- logs and debug traces
- world tick advance

## Packaging

This SDK now includes:

- `GameAgentEngine.SDK.csproj`
- `src/GameAgentEngineClient.cs`
- `src/Models.cs`
- example entry files under `examples/`

The client is based on `HttpClient` plus typed POCO models and `System.Text.Json` serialization.

## Recommended Use

Use this SDK when you need to:

- connect a C# runtime to GameAgentEngine
- trigger `invoke`
- consume runtime tasks
- callback task results
- inspect continuity state, timelines, logs, and traces
- build Unity-friendly or .NET tool-side integrations around `GameAgentWorker`

## Included Examples

- `examples/HealthExample.cs`
- `examples/InvokeDialogueExample.cs`
- `examples/TaskPullOnceExample.cs`
- `examples/WorkerRuntimeRoundtripExample.cs`
- `examples/WorkerAuthorityQueryExample.cs`

## GameAgentWorker Integration Flows

### 1. Pull task claim / start / callback roundtrip

When Engine has already produced a pending runtime task and you want a C# tool to complete the worker-side callback path, use:

- `examples/WorkerRuntimeRoundtripExample.cs`

Recommended environment variables:

```bash
GAE_SERVER=http://127.0.0.1:8080
GAE_KEY=dev-key
GAE_CONSUMER=game_client
GAE_OWNER=cs-sdk-roundtrip
GAE_CALLBACK_STATUS=success
```

This example:

- lists one pending runtime task
- claims it
- starts it
- callbacks a deterministic result payload
- prints whether resume / post-process happened

### 2. Authority query / resume preparation flow

When you want to trigger one Engine invoke that emits `game_client_request_data`, then hand the pending runtime task to `GameAgentWorker`, use:

- `examples/WorkerAuthorityQueryExample.cs`

Recommended environment variables:

```bash
GAE_SERVER=http://127.0.0.1:8080
GAE_KEY=dev-key
GAE_WORLD_ID=demo_world
GAE_NODE_ID=innkeeper_001
GAE_DYNAMIC_INTERFACES_FILE=tools/source/tests/runtime_task_dynamic_interfaces.json
GAE_PIPELINE_MODE=full
```

This example:

- invokes one dialogue request with request-scoped dynamic interfaces
- prints the Engine response summary
- finds the pending runtime task created for `game_client_request_data`
- prints the exact consumer for `GameAgentWorker pull-once`

Typical follow-up command:

```bash
GameAgentWorker pull-once --consumer game_client
```

## Not Yet Included

Not yet at Go SDK parity:

- full worlds / nodes / components / memories / relations CRUD surface
- snapshot / restore / fork helpers
- higher-level dynamic interface builders and interaction helpers

Those can be layered on later without changing the current client shape.
