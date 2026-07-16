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

## Not Yet Included

Not yet at Go SDK parity:

- full worlds / nodes / components / memories / relations CRUD surface
- snapshot / restore / fork helpers
- higher-level dynamic interface builders and interaction helpers

Those can be layered on later without changing the current client shape.
