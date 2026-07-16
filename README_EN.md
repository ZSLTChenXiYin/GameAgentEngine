# GameAgentEngine

[**中文**](./README.md) | **English**

An AI Agent engine built for games.

---

## Why Games Need Their Own Agent Engine

If you have tried using general-purpose LLM Agents in a game, you have likely run into these problems:

> **"Who are you again?" every single time.** Regular agents have no persistent world. Every conversation starts from scratch. The NPC has no idea it met the player yesterday or what happened at the village gate.

> **Agents can do anything.** LLMs will say anything, but in a game, NPCs should only do what the game allows -- sell items, trigger quests, lead the way. Agents need validated action boundaries, not just prompt-level suggestions.

> **No world clock.** Game worlds have their own time: shops open during the day and close at night. Regular agents have no concept of "world time."

> **Integration is a DIY mess.** Every agent framework exposes HTTP and calls it done. Games need push, callbacks, WebSocket, retry, and fallback. Every engine integration means rebuilding the same plumbing.

> **Designers cannot use it.** Most agent frameworks are built for ML engineers. Game designers need a visual editor to inspect NPC state, tweak configs, and run tests without touching code.

GameAgentEngine was designed to solve these problems.

It is a **game-focused AI Agent runtime** that sits between game logic and LLMs, handling world modeling, NPC behavior, memory management, time progression, external task dispatch, and controlled action execution. It does not replace Unity, Unreal, or Godot. It works alongside them as a dedicated AI world layer.

---

## What the Engine Does

### World Modeling -- More Than a Chat Table

The Engine builds game worlds using four structural primitives:

- **Nodes** -- any entity in the world: NPC, location, organization, item, quest
- **Components** -- structured data attached to nodes: `autonomous` (behavior config), `world_state` (world overview), `tick_policy` (rules)
- **Memories** -- personal knowledge base per node, tiered by importance (core / normal / trivial)
- **Relations** -- social graph between nodes (ally / enemy / subordinate / kinship / located_at), automatically organized into prompt context

```
[village] -> located_at -> [gatekeeper NPC]
[gatekeeper NPC] -> enemy -> [bandit NPC]
[bandit NPC] -> belongs_to -> [bandit faction]
```

When the player talks to the gatekeeper, the Engine automatically includes "this NPC is hostile to the bandits, who belong to the bandit faction" in the prompt context.

### Controlled Actions -- NPCs Cannot Go Rogue

Every LLM-produced action is validated against a **capability allowlist**:

- Sync actions execute immediately (add memory, update component)
- Async actions return a `callback_id` -- the game client confirms completion before proceeding
- Unauthorized actions are rejected and logged

Give an NPC `["add_memory", "start_dialogue"]` as its allowlist, and it will never call `delete_world`.

### World Time and Continuity -- Games Have Their Own Clock

The Engine has a built-in world time system:

- Tick advancement (supports `fixed` and `flexible` scale modes)
- Each Tick archives continuity state: world overview, story state, narrative history, time snapshot
- The next Tick automatically loads the previous state into inference context

This means "the world keeps evolving" and "NPCs know today is different from yesterday" work out of the box.

### External Interaction -- Async Tasks for Game Clients

Three task delivery modes between the game client and the Engine:

| Mode | How It Works |
|---|---|
| **Push** | Engine pushes tasks directly via HTTP / WebSocket / RPC |
| **Pull** | Game client polls pending tasks, claims, executes, heartbeats, and reports results |
| **Hybrid** | Combination of both |

Supports fallback routing and automatic retry governance.

### Three Pipeline Modes

| Mode | Best For |
|---|---|
| `vertical` | Single-pass, low latency, simple NPC dialogue |
| `polling` | Multi-round with data requests and sub-tasks, complex narratives |
| `full` | Full inference with sub-task DAG and async callback orchestration, world Ticks |

### Background Autonomous Behavior

Nodes can be configured with autonomous components so NPCs act without player interaction:

- **Manual** -- trigger only when called
- **Tick Sync** -- trigger automatically after each world Tick
- **Scheduled** -- background scheduler runs at configured intervals

---

## Quick Start

### Build

Recommended: use the build scripts.

```
# Windows
tools\scripts\build.bat

# Linux / macOS
bash tools/scripts/build.sh
```

Output goes to `dist/`, containing the engine binary, CLI, config file, docs, and Web editor.

### Start

```
# Copy the default config
cp tools/source/gameagentengine.conf.yaml .

# Start the engine
GameAgentEngine serve
```

### Create a World and Open the Editor

```
# Create a world
GameAgentDevCli node create --type world --name "My World"

# Open the visual editor
GameAgentDevCli creator
```

### Add an NPC and Advance Time

```
GameAgentDevCli node create --world <world-id> --type npc --name "Blacksmith"

# Advance world time
GameAgentDevCli world tick <world-id>
```

> **Note:** Before running `world tick`, configure `world_time_settings` in Creator's Settings page. Without a world-time configuration, time advancement is intentionally blocked -- this is a design constraint, not a bug.

### Try the Demo World and Text-Game Shell Directly

The repository ships with a minimal demo asset pair:

- `tools/source/demo-world.yaml` -- demo world imported into the Engine
- `tools/source/demo-state.yaml` -- authority state file consumed by `GameAgentWorker play`

Shortest path:

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/demo-world.yaml
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

This lets the Engine drive NPC behavior while the worker-side state file remains the authority source for HP, inventory, money, quest state, and scene occupancy.

---

## Toolchain

The Engine now ships with four first-class developer tool entrypoints:

### GameAgentCreator -- Visual Web Editor

For game designers and developers, bundled with every build. Available pages:

| Page | Purpose |
|---|---|
| **Worlds** | World selection, creation, drag-and-drop node tree |
| **Snapshots** | Snapshot management (save, validate, restore, delete) |
| **Tasks** | Runtime task monitoring (status, category, retry count) |
| **Plans** | Pending plan approval |
| **Policy** | World policy configuration |
| **Settings** | World settings and `world_time_settings` editing |
| **Continuity** | Continuity debug bundle |
| **State** | Continuity state component inspection and editing |
| **Timelines** | Timeline archives |
| **Logs** | Inference logs |
| **Traces** | Debug traces |

### GameAgentDevCli -- Command Line Tool

For scripting and CI integration:

```
# CRUD
GameAgentDevCli node create --type world --name "My World"
GameAgentDevCli node list --world <id>
GameAgentDevCli component add <node-id> --type autonomous
GameAgentDevCli memory add <node-id> --content "..."
GameAgentDevCli relation create --source <a> --target <b> --type ally

# World operations
GameAgentDevCli world tick <world-id>
GameAgentDevCli world settings get <world-id>

# Task management (Pull mode)
GameAgentDevCli task list --status pending --limit 20
GameAgentDevCli task claim <task-id> --consumer gamer
GameAgentDevCli task start <task-id> <lease-token>

# Debugging and observability
GameAgentDevCli logs --world <world-id>
GameAgentDevCli debug traces --world <world-id>
GameAgentDevCli creator
```

### GameAgentWorker -- Standalone Worker / REPL / Integration Test Entry

Use it for external async callback simulation, local game-side authority state, play-mode REPL, and packaged integration tests:

```bash
# Run push receiver and pull worker together
GameAgentWorker serve --verbose

# Process one pending pull task
GameAgentWorker pull-once --consumer game_client

# Enter text-game REPL
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001

# Run packaged integration scenarios
GameAgentWorker test all
```

It is not just a temporary test script wrapper. It is the canonical game-side worker in this repository:

- for integration tests, it closes the push / pull / callback loop
- for local development, it hosts YAML / JSON authority state and simulates async game-side interfaces
- for play mode, it exposes `/talk`, `/ask`, `/gift`, `/trade`, and related REPL flows to validate Engine-driven text-game interaction

### Go SDK

For Go server-side integration:

```go
import "github.com/ZSLTChenXiYin/GameAgentEngine/sdk"

client := sdk.NewClient("http://127.0.0.1:8080", "dev-key")

// List pending runtime tasks
tasks, _ := client.ListRuntimeTasks("", "pending", 20)

// Advance world time
tick, _ := client.AdvanceTick(worldID, "scheduled", "Day 2")
```

Covers world management, node CRUD, memory propagation, autonomous configuration, runtime task management, event impact evaluation, and all other API methods.

---

## Documentation

- [Getting Started](./docs/getting-started/GETTING_STARTED_EN.md) -- build your first world from scratch
- [Architecture](./docs/architecture/ARCHITECTURE_EN.md) -- overall design and module boundaries
- [Core Concepts](./docs/architecture/CORE_CONCEPTS_EN.md) -- nodes, components, memories, relations explained
- [External Interaction Roadmap](./docs/EXTERNAL_INTERACTION_ROADMAP.md) -- Push / Pull / Hybrid task modes
- [External Interaction Examples](./docs/EXTERNAL_INTERACTION_EXAMPLES.md) -- full game-client integration flows
- [API Reference](./docs/reference/API_REFERENCE_EN.md) -- all HTTP endpoints
- [GameAgentDevCli Guide](./docs/guides/GUIDE_GAMEAGENTDEVCLI_EN.md) -- complete CLI reference
- [GameAgentCreator Guide](./docs/guides/GUIDE_GAMEAGENTCREATOR_EN.md) -- Web editor walkthrough
- [SDK Reference](./docs/reference/SDK_REFERENCE_EN.md) -- Go SDK methods and types
- [Configuration](./docs/reference/CONFIGURATION_EN.md) -- static config and dynamic world settings
- [Autonomous Behavior System](./docs/architecture/AUTONOMOUS_BEHAVIOR_EN.md) -- node-level autonomous scheduling
- [Pipeline Internals](./docs/architecture/PIPELINE_INTERNALS_EN.md) -- execution flow details
- [World Time Tick Reference](./docs/reference/WORLD_TIME_TICK_REFERENCE_EN.md) -- scale modes and advancement rules
- [Build and Deploy](./docs/reference/BUILD_AND_DEPLOY_EN.md) -- cross-platform compilation and deployment

---

## License

MIT
