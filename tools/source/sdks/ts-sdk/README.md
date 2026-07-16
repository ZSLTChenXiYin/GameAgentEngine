# GameAgentEngine TypeScript SDK

This SDK is the primary TypeScript client baseline for GameAgentEngine.

## Current Status

This is now a practical first version rather than a pure placeholder scaffold.

It still does not provide full Go SDK parity, but it already covers the main integration loop needed by external tools and worker-side orchestration.

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

- `src/client.ts` — typed HTTP client
- `src/types.ts` — shared request / response models
- `src/index.ts` — public exports
- `package.json`
- `tsconfig.json`

It is designed to work with Node.js 18+ and uses the built-in `fetch` implementation by default.

## Build

```bash
npm install
npm run check
npm run build
```

If your runtime does not expose global `fetch`, inject one through the client constructor.

## Example Usage

```ts
import { GameAgentEngineClient } from './src/index';

const client = new GameAgentEngineClient('http://127.0.0.1:8080', 'dev-key');

const version = await client.getVersion();
console.log(version.version, version.min_compatible);
```

## Recommended Use

Use this SDK when you need to:

- connect a TypeScript runtime to GameAgentEngine
- trigger `invoke`
- consume runtime tasks
- callback task results
- inspect continuity state, timelines, logs, and traces
- build Node.js-side tools around `GameAgentWorker` and Engine integration flows

## Included Examples

- `examples/health.ts`
- `examples/invoke_dialogue.ts`
- `examples/task_pull_once.ts`
- `examples/worker_runtime_roundtrip.ts`
- `examples/worker_authority_query.ts`

## GameAgentWorker Integration Flows

### 1. Pull task claim / start / callback roundtrip

When Engine has already produced a pending runtime task and you want a TypeScript tool to complete the worker-side callback path:

```bash
node dist/examples/worker_runtime_roundtrip.js
```

Recommended environment variables:

```bash
GAE_SERVER=http://127.0.0.1:8080
GAE_KEY=dev-key
GAE_CONSUMER=game_client
GAE_OWNER=ts-sdk-roundtrip
GAE_CALLBACK_STATUS=success
```

This example:

- lists one pending runtime task
- claims it
- starts it
- callbacks a deterministic result payload
- prints whether resume / post-process happened

### 2. Authority query / resume preparation flow

When you want to trigger one Engine invoke that emits `game_client_request_data`, then hand the pending runtime task to `GameAgentWorker`:

```bash
node dist/examples/worker_authority_query.js
```

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
- prints the exact consumer and task metadata for `GameAgentWorker pull-once`

Typical follow-up command:

```bash
GameAgentWorker pull-once --consumer game_client
```

## Not Yet Included

Not yet at Go SDK parity:

- full worlds / nodes / components / memories / relations CRUD surface
- snapshot / restore / fork helpers
- higher-level typed builders for dynamic interfaces and interaction flows

Those can be added in later phases without changing the current core client shape.
