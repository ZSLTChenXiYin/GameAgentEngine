# GameAgentEngine

[**中文**](./README.md) | **English**

An AI Agent runtime built for games.

This is not just a “chatty NPC wrapper.” It is a dedicated game-side runtime for world continuity, authoritative state access, controlled action execution, world-time progression, and a development workflow that can be debugged, packaged, and integrated reliably.

---

## What Problems It Solves

If you have tried dropping a general-purpose LLM Agent directly into a game, you have probably run into these problems:

| Game dev pain point | Common problem with generic agents | What GameAgentEngine does |
|---|---|---|
| NPCs forget everything every conversation | No persistent world, so yesterday’s events disappear fast | Nodes / memories / relations / continuity state keep the world connected |
| AI makes up player state | It does not know whether the player really has a knife, money, inventory, or a quest stage | It queries the game side for authoritative state on demand |
| Actions are not safe to execute | The model can claim to do things that the game should never allow | Capability allowlists + sync/async execution + callback confirmation |
| The world has no time | You get dialogue, but not day/night, progression, or archived history | world tick + world time + timeline + future outline |
| Async integration becomes glue hell | HTTP alone is not enough for push / pull / callback / retry flows | Engine / Worker / runtime task / resume are unified |
| Designers and programmers cannot both use it | Prompt plus code is not a usable workflow | Creator + DevCli + Worker + docs + demo workflow |
| Integration tests are hard to reproduce | State simulation and external callbacks fragment into scripts | Worker CLI + fixtures + packaged smoke / integration flows |

---

## What It Is

GameAgentEngine is a standalone runtime between game logic and LLMs, responsible for:

- world modeling
- NPC behavior
- memory management
- time progression
- external task dispatch
- controlled action execution
- authoritative state queries
- observable, reproducible debugging flows

It does not replace Unity, Unreal, or Godot. It works with them as an AI world layer.

---

## Core Capabilities

### 1. World Modeling

The world is not a chat table. It is built from four structural primitives:

- Nodes: NPCs, locations, factions, items, quest lines, and the world itself
- Components: structured state such as `world_state`, `story_state`, `tick_policy`, and `autonomous`
- Memories: per-node knowledge, tiered by importance
- Relations: graph edges such as `located_at`, `belongs_to`, `subordinate`, and `enemy`

### 2. Authoritative State

High-frequency facts that must stay correct live on the game side, and Engine queries them instead of guessing.

Examples:

- HP
- money
- inventory
- quest stage
- scene occupancy
- live weather
- temporary event state

### 3. Controlled Actions

Every LLM output is validated against a capability allowlist:

- Sync actions execute immediately
- Async actions return a `callback_id` and wait for the game side to confirm
- Unauthorized actions are rejected and logged

### 4. World Time Progression

The Engine includes a built-in world-time system:

- supports `fixed` and `flexible` tick scales
- archives continuity state after each tick
- automatically feeds the previous world state into the next tick

### 5. A Workflow You Can Actually Integrate

GameAgentEngine ships with:

- GameAgentCreator: visual editor
- GameAgentDevCli: CLI and CI tool
- GameAgentWorker: game-side worker, REPL, and integration-test entrypoint
- multi-language SDKs: a unified integration surface for external systems

---

## Who It Is For

- teams building text games, DOL-like games, narrative RPGs, or persistent NPC behavior
- teams that need AI to read true game state instead of guessing from prompts
- teams that need async callbacks, external tasks, tests, and packaging to work together
- teams where designers and programmers both need to inspect and debug the world

---

## What the Workflow Looks Like

```text
World Definition -> Engine world modeling -> cold start baseline
                     -> world tick -> authority query -> controlled actions
                     -> timeline / memories / story state
                     -> Worker / Creator / DevCli / SDK integration
```

A more practical sequence is:

1. import the world skeleton
2. cold-start the runtime baseline
3. let world tick advance the story on top of authority data and continuity
4. let Worker answer live game-side state
5. use Creator / DevCli / SDKs for editing, integration, regression, and packaging

---

## Quick Start

### Build

```bash
# Windows
tools\scripts\build.bat

# Linux / macOS
bash tools/scripts/build.sh
```

### Start the Engine

```bash
GameAgentEngine serve
```

### Create a World and Open the Editor

```bash
GameAgentDevCli node create --type world --name "My World"
GameAgentDevCli creator
```

### Try the Demo World and Text-Game Shell

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/workerhome/demo/demo-world.yaml
GameAgentWorker play --state-file tools/source/workerhome/demo/demo-state.yaml --world-id demo_world --player-node-id player_001
```

---

## Tooling Entry Points

| Tool | Purpose |
|---|---|
| GameAgentCreator | world editing, runtime-state inspection, debugging, regression |
| GameAgentDevCli | import, initialization, debugging, logs, time advancement, package acceptance |
| GameAgentWorker | authoritative state, REPL, push/pull/callback, integration tests |
| SDKs | the integration surface for external systems and other languages |

---

## Documentation Entry Points

- [Chinese documentation index](./docs/README.md)
- [English documentation index](./docs/README_EN.md)

---

## License

MIT
