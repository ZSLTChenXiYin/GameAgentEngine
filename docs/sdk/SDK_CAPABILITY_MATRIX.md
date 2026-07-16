# SDK Capability Matrix

This matrix summarizes the current implementation level of each SDK directory relative to the current project workflow.

Status levels:

- `practical`: already usable for real Engine / Worker integration work
- `baseline`: basic scaffold exists, but still shallow
- `planned`: not yet upgraded in the current refactor sequence

## 1. Current Matrix

| SDK | Current Level | Health / Version | Invoke | Runtime Task Loop | Callback | State / Timeline / Logs | Worker Examples | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `ts-sdk` | practical | yes | yes | yes | yes | yes | yes | strongest non-Go SDK so far |
| `js-sdk` | practical | yes | yes | yes | yes | yes | yes | plain Node.js / script-friendly twin to ts-sdk |
| `cs-sdk` | practical | yes | yes | yes | yes | yes | yes | Unity / .NET oriented, typed POCO client |
| `gd-sdk` | practical-request-builder | request builders | request builders | request builders | request builders | request builders | yes | Godot-side request construction now includes worker authority-query and roundtrip examples; execution wrapper still absent |
| `cpp-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | request-builder surface now covers pull/callback helper sequence plus worker examples |
| `java-sdk` | baseline+worker-loop | yes | yes | yes | yes | no | yes | real HTTP client plus authority-query / callback roundtrip examples; observability surface still shallow |
| `lua-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | lightweight request helper now mirrors the worker loop and authority-query sequence |
| `c-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | path/payload helper set plus worker examples; caller still owns transport |

## 2. What “practical” Means Here

For this project, a practical SDK is expected to cover the current outer-loop workflow:

1. connect to Engine
2. invoke one request
3. inspect or consume one runtime task
4. callback one result
5. inspect continuity-oriented runtime artifacts when needed
6. integrate cleanly with `GameAgentWorker`

## 3. Current Mainline SDKs

The current mainline SDK set is:

- `ts-sdk`
- `js-sdk`
- `cs-sdk`
- `gd-sdk`

These are the SDKs that already map directly onto the current Engine / Worker development workflow.

## 4. Upgrade Priority for Remaining SDKs

Remaining order in the current plan:

1. `lua-sdk`
2. `c-sdk`
3. `cpp-sdk` observability / typed-model depth

The next upgrade target after this matrix phase is still observability / typed-model depth for the lower-tier SDK set, not a rework of the already practical SDKs.
