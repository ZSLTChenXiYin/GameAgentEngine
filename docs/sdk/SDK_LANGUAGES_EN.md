# Language SDK Status

[**中文**](./SDK_LANGUAGES.md) | **English**

This document centralizes the positioning, current coverage, and attached examples for each language SDK, replacing the older scattered notes that used to live under `tools/source/sdks/*/README.md`.

## TypeScript SDK

- directory: `tools/source/sdks/ts-sdk`
- positioning: one of the most complete external scripting / tooling baselines at the moment
- current coverage:
  - health / version
  - invoke
  - player input interpretation
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback
  - world settings
  - continuity state components
  - timelines
  - logs / debug traces
  - world tick advance
- key structure:
  - `src/client.ts`
  - `src/types.ts`
  - `src/index.ts`

## JavaScript SDK

- directory: `tools/source/sdks/js-sdk`
- positioning: lightweight Node.js tooling and bridge-process baseline
- current coverage:
  - health / version
  - invoke
  - player input interpretation
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback
  - world settings
  - continuity state components
  - timelines
  - logs / debug traces
  - world tick advance
- note: depends on Node.js 18+ `fetch` by default

## C# SDK

- directory: `tools/source/sdks/cs-sdk`
- positioning: server-side C# / Unity-adjacent integration baseline
- current coverage:
  - health / version
  - invoke
  - player input interpretation
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback
  - world settings
  - continuity state components
  - timelines
  - logs / debug traces
  - world tick advance
- key structure:
  - `src/GameAgentEngineClient.cs`
  - `src/Models.cs`
  - `GameAgentEngine.SDK.csproj`

## GDScript SDK

- directory: `tools/source/sdks/gd-sdk`
- positioning: lightweight Godot-side integration baseline
- current coverage:
  - request builders for health / version
  - request builders for invoke / player input interpretation
  - request builders for runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - request builders for callback
  - request builders for world settings / state components / timelines / logs / debug traces / world policy
  - request builders for world tick advance
- attached examples: worker authority query / runtime roundtrip / continuity inspect
- note: returns Godot-friendly request dictionaries and leaves HTTP transport to the host project

## Java SDK

- directory: `tools/source/sdks/java-sdk`
- positioning: Java server-side or middleware integration baseline
- current coverage:
  - health / version
  - invoke
  - player input interpretation
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback completion
  - world settings
  - continuity state components
  - timelines
  - logs / debug traces
  - world policy
  - world tick advance
  - authority-query and worker roundtrip examples
  - continuity inspection example
- current status: now covers the minimum Engine / Worker integration loop plus a basic continuity-inspection loop directly, though typed-model depth is still weaker than TS / JS / C#

## C++ SDK

- directory: `tools/source/sdks/cpp-sdk`
- positioning: request-construction baseline for native-side integration
- current coverage:
  - request builders for health / version / invoke / player input interpretation
  - request builders for world tick advance / world settings / state components / timelines / logs / debug traces / world policy
  - request builders for runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - request builders for callback
  - worker authority-query / runtime roundtrip / continuity inspect examples
- current status: does not ship a full HTTP transport layer, but its request-builder layer now covers the minimum sequence needed for Worker integration and basic continuity inspection

## C SDK

- directory: `tools/source/sdks/c-sdk`
- positioning: lowest-dependency native-side integration baseline
- current coverage:
  - path helpers for health / version / invoke / player input interpretation
  - path helpers for world settings / state components / timelines / logs / debug traces / world policy
  - path / payload helpers for runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback payload construction
  - worker authority-query / runtime roundtrip / continuity inspect examples
- current status: does not ship an HTTP transport layer, but already covers the request assembly needed for Worker integration and basic continuity inspection

## Lua SDK

- directory: `tools/source/sdks/lua-sdk`
- positioning: lightweight script-side integration baseline
- current coverage:
  - health / version / invoke / player input interpretation path and request helpers
  - world tick advance / world settings / state components / timelines / logs / debug traces / world policy request helpers
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats helpers
  - callback payload / request helpers
  - worker authority-query / runtime roundtrip / continuity inspect examples
- current status: does not ship an HTTP transport layer, but the request-construction layer already matches the Worker integration sequence and a basic continuity inspection sequence directly

## Unified Requirements

- the Go SDK remains the semantic baseline: `sdk/client.go`, `sdk/types.go`
- capability naming and API semantics across all language SDKs should stay aligned with the Go SDK as much as possible
- when adding or backfilling SDK-facing documentation, write it into `docs/sdk/` rather than restoring scattered `tools/source/sdks/*/README.md` files
