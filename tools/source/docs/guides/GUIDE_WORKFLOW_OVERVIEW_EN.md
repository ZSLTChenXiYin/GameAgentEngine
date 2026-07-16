# Workflow Overview

[**中文**](./GUIDE_WORKFLOW_OVERVIEW.md) | **English**

This document explains the current responsibility split between `Engine`, `DevCli`, `Worker`, `Creator`, and the language SDKs, and how they fit together in real development and integration workflows.

It is not a deep architecture document. It is a practical workflow-facing overview.

---

## 1. Five Primary Roles

| Component | Responsibility |
| --- | --- |
| `GameAgentEngine` | world modeling, NPC reasoning, memory, relations, time progression, external async task orchestration |
| `GameAgentDevCli` | config import, world management, state inspection, task and timeline debugging, opening Creator |
| `GameAgentWorker` | game-side worker, authority-state host, push/pull/callback loop, play REPL, packaged test scenarios |
| `GameAgentCreator` | browser-based visual editor and observability UI |
| `SDKs` | programmatic integration path into the same Engine / Worker workflow |

The current boundary is simple:

- reasoning, world evolution, and async task orchestration belong to Engine
- game-side authoritative truth, high-frequency queries, and external interface consumption belong to Worker / the game side
- operational and diagnostic entrypoints belong to DevCli / Creator / SDKs

---

## 2. Shortest Development Loop

If you just want to verify the project end to end, the recommended shortest loop is:

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/demo-world.yaml
GameAgentDevCli creator
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

That means:

1. `Engine` starts the service
2. `DevCli` imports the world
3. `Creator` handles visual inspection and editing
4. `Worker play` handles game-side authority state plus NPC interaction experience

---

## 3. Integration-Test Loop

If the goal is validating external async interfaces instead of the REPL experience, use:

```bash
GameAgentEngine serve
GameAgentWorker serve --verbose
```

Or run the packaged worker scenarios directly:

```bash
GameAgentWorker test all
```

At this layer:

- Engine produces runtime tasks
- Worker consumes push / pull tasks and callbacks results
- DevCli / SDKs drive and inspect the concrete flow

---

## 4. `play` Is Not a Raw Shell

The current `GameAgentWorker play` mode is not just a CLI text wrapper around Engine. It is:

- the text-game entrypoint for player and NPC interaction
- the game-side authority-state query boundary
- the place where natural-language player intent can be validated and executed
- the integration surface for Engine-driven NPC and room-chat feedback

So:

- `play` belongs to Worker, not to Engine itself
- `play` depends on authority-state files such as `demo-state.yaml`
- player input must still pass game-side truth validation before becoming real state changes

---

## 5. Where SDKs Fit

The multi-language SDKs are not intended to replace DevCli. They exist so that external programs, scripts, bridge layers, and engine plugins can integrate into the same Engine / Worker loop.

Current practical SDKs:

- `ts-sdk`
- `js-sdk`
- `cs-sdk`
- `gd-sdk` (request-construction level)

Their main current use is:

- invoke requests
- consume runtime tasks
- callback results
- inspect state / timeline / logs / traces
- form the smallest useful integration loop with `GameAgentWorker`

Shared SDK references live in:

- `tools/source/sdks/README.md`
- `tools/source/sdks/SDK_FIXTURES.md`
- `tools/source/sdks/SDK_CAPABILITY_MATRIX.md`

---

## 6. Recommended Tool Split

### Config and import

- use `GameAgentDevCli` first

### Visual editing and observability

- use `GameAgentCreator` first

### Programmatic integration

- use the language SDKs first

### Async task integration

- use `GameAgentWorker` first

### Text-game experience and authority-state validation

- use `GameAgentWorker play` first

---

## 7. Recommended Documentation Path

If you want to read the docs in workflow order, use this path:

1. `docs/getting-started/GETTING_STARTED.md`
2. `docs/guides/GUIDE_GAMEAGENTDEVCLI.md`
3. `docs/guides/GUIDE_GAMEAGENTWORKER.md`
4. `docs/gameplay/PLAYER_INTERACTION.md`
5. `docs/gameplay/GAME_STATE_AUTHORITY.md`
6. `docs/integration/EXTERNAL_INTERACTION.md`
7. `tools/source/sdks/README.md`

That sequence matches the current real project workflow better than reading purely by subsystem.
