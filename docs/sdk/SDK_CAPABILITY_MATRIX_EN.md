# SDK Capability Matrix

This matrix summarizes the current implementation maturity of each SDK directory relative to the active project workflow.

Status levels:

- `practical`: already usable for real Engine / Worker integration work
- `baseline`: basic scaffold exists, but capability depth is still shallow
- `planned`: not yet upgraded to the target level in the current refactor sequence

## 1. Current Matrix

| SDK | Current Level | Health / Version | Invoke | Runtime Task Loop | Callback | State / Timeline / Logs | Worker Examples | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `ts-sdk` | practical | yes | yes | yes | yes | yes | yes | strongest non-Go SDK at the moment |
| `js-sdk` | practical | yes | yes | yes | yes | yes | yes | lightweight plain Node.js / scripting counterpart |
| `cs-sdk` | practical | yes | yes | yes | yes | yes | yes | typed client aimed at Unity / .NET integration |
| `gd-sdk` | practical-request-builder | request builders | request builders | request builders | request builders | request builders | yes | worker authority-query and roundtrip examples are present, but the execution wrapper is still absent |
| `cpp-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | request-builder surface now covers pull/callback helper sequences and worker examples |
| `java-sdk` | practical | yes | yes | yes | yes | yes | yes | real HTTP client now covers world settings, state, timelines, logs, debug traces, world policy, and a continuity-inspection example |
| `lua-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | lightweight request helpers now mirror the worker loop and authority-query sequence |
| `c-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | path/payload helpers and worker examples exist, while the caller still owns transport |

## 2. What practical Means Here

In this project, a practical SDK is expected to cover at least the current outer integration loop:

1. connect to Engine
2. send one invoke
3. inspect or consume one runtime task
4. callback one result
5. inspect continuity-related runtime artifacts when needed
6. work cleanly with `GameAgentWorker`

## 3. Current Mainline SDKs

The current mainline SDK set is:

- `ts-sdk`
- `js-sdk`
- `cs-sdk`
- `gd-sdk`
- `java-sdk`

These SDKs already map directly to the active Engine / Worker development workflow.

## 4. Upgrade Priority for Remaining SDKs

The remaining order in the current plan is:

1. `lua-sdk`
2. `c-sdk`
3. observability / typed-model depth for `cpp-sdk`

Java SDK has now moved from a worker-loop baseline into the practical tier for continuity-oriented integration work.

The next upgrade target after this phase is still observability and typed-model depth for the lower-tier SDK set, not a rework of the SDKs that are already practical.
