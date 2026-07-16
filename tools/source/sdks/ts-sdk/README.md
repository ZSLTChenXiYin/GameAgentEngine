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

## Not Yet Included

Not yet at Go SDK parity:

- full worlds / nodes / components / memories / relations CRUD surface
- snapshot / restore / fork helpers
- higher-level typed builders for dynamic interfaces and interaction flows

Those can be added in later phases without changing the current core client shape.
