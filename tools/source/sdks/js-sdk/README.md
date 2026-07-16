# GameAgentEngine JavaScript SDK

This SDK is the primary plain-JavaScript baseline for lightweight Node.js tools and bridge-side integration.

## Current Status

This is now a practical first version rather than a raw placeholder wrapper.

It still does not provide full Go SDK parity, but it already covers the main Engine / Worker integration surface needed by external JavaScript tools.

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

## Notes

- best fit for lightweight scripts, Node.js tools, and bridge processes
- keeps the API surface close to the TypeScript baseline without TS-only types
- uses built-in `fetch` by default on Node.js 18+

## Included Examples

- `examples/health.js`
- `examples/invoke_dialogue.js`
- `examples/task_pull_once.js`
- `examples/worker_runtime_roundtrip.js`
- `examples/worker_authority_query.js`

## GameAgentWorker Integration Flows

### 1. Pull task claim / start / callback roundtrip

Use `examples/worker_runtime_roundtrip.js` when Engine has already produced a pending runtime task and you want a JavaScript tool to complete the worker-side callback path.

Recommended environment variables:

```bash
GAE_SERVER=http://127.0.0.1:8080
GAE_KEY=dev-key
GAE_CONSUMER=game_client
GAE_OWNER=js-sdk-roundtrip
GAE_CALLBACK_STATUS=success
```

### 2. Authority query / resume preparation flow

Use `examples/worker_authority_query.js` when you want to trigger one Engine invoke that emits `game_client_request_data`, then hand the pending runtime task to `GameAgentWorker`.

Recommended environment variables:

```bash
GAE_SERVER=http://127.0.0.1:8080
GAE_KEY=dev-key
GAE_WORLD_ID=demo_world
GAE_NODE_ID=innkeeper_001
GAE_DYNAMIC_INTERFACES_FILE=tools/source/tests/runtime_task_dynamic_interfaces.json
GAE_PIPELINE_MODE=full
```

Typical follow-up command:

```bash
GameAgentWorker pull-once --consumer game_client
```

## Not Yet Included

Not yet at Go SDK parity:

- full worlds / nodes / components / memories / relations CRUD surface
- snapshot / restore / fork helpers
- higher-level dynamic interface builders and interaction helpers
