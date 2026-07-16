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
| `gd-sdk` | practical-request-builder | request builders | request builders | request builders | request builders | request builders | partial | Godot-side request construction is in place; execution wrapper still absent |
| `cpp-sdk` | baseline | partial | partial | no | no | no | no | next-wave upgrade target |
| `java-sdk` | baseline | partial | partial | no | no | no | no | next-wave upgrade target |
| `lua-sdk` | baseline | partial | partial | no | no | no | no | later-stage lightweight integration target |
| `c-sdk` | baseline | partial | partial | no | no | no | no | later-stage native integration target |

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

1. `cpp-sdk`
2. `java-sdk`
3. `lua-sdk`
4. `c-sdk`

The next upgrade target after this matrix phase is still the lower-tier SDK set, not a rework of the already practical SDKs.
