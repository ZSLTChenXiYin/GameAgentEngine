# GameAgentEngine

[**中文**](./README.md) | **English**

An AI agent creation and runtime engine for game developers.

GameAgentEngine sits between game logic and LLM capabilities. It handles world modeling, NPC behavior, memory management, world time progression, and controlled runtime actions. It does not replace Unity, Unreal, or Godot. It works alongside them as a dedicated AI world layer.

---

## Highlights

- Unified world graph based on nodes, components, memories, and relations
- LLM-driven NPC dialogue, world reasoning, and event impact evaluation
- World tick progression, continuity state, timeline archives, and world time advancement
- Three pipeline modes: `vertical`, `polling`, `full`
- Distinct semantics for working-copy fork, save snapshot, and restore
- Shared component validation metadata across Engine, Creator, and DevCli
- Runtime management of `world_settings`, `world_policy`, and continuity state components
- Bundled Web editor (`GameAgentCreator`), CLI (`GameAgentDevCli`), and Go SDK

---

## Quick Start

If this is your first time using the project, start with [Getting Started](./docs/GETTING_STARTED_EN.md).

Shortest path:

```bash
# 1. Build
go build ./...

# 2. Copy the default config
cp tools/source/gameagentengine.conf.yaml .

# 3. Start the engine
go run ./cmd/gameagentengine serve

# 4. Create a world root node
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --type world --name "New World"

# 5. Open Creator
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key inspect
```

If you plan to use `world tick`, timeline advancement, or worldline reasoning, configure `world_time_settings` for the world first. That configuration intentionally blocks dependent save flows when missing, so developers are reminded to define the world time system before relying on timeline inference.

---

## Current Tooling

### GameAgentCreator

The Creator currently supports:

- world selection and creation
- world rename
- node tree browsing, drag-to-parent, and drag-to-root
- node create, edit, delete, and copy
- relation creation and graph validation hints
- snapshot save, validation, restore, and delete
- world settings, world policy, and pending plan management
- continuity, state, timelines, logs, and traces
- `world_time_settings` editing and `world_time_state` inspection

### GameAgentDevCli

The CLI currently supports:

- node / component / memory / relation CRUD
- world tick, event impact, scope advance, and timeline replan
- world settings, world policy, and continuity state management
- working-copy fork, save snapshot, restore, snapshot validation, and snapshot metadata
- logs, traces, continuity debugging, and node graph debugging
- the `inspect` entry for Creator

### Go SDK

The SDK currently supports:

- basic service access, version, and health checks
- world settings and `world_time_settings` reads and partial updates
- continuity state component and timeline archive access
- world tick, event impact, plan approval, and snapshot flows

---

## Documentation

- [Getting Started](./docs/GETTING_STARTED_EN.md)
- [Architecture](./docs/ARCHITECTURE_EN.md)
- [Core Concepts](./docs/CORE_CONCEPTS_EN.md)
- [API Reference](./docs/API_REFERENCE_EN.md)
- [GameAgentDevCli Guide](./docs/GUIDE_GAMEAGENTDEVCLI_EN.md)
- [GameAgentCreator Guide](./docs/GUIDE_GAMEAGENTCREATOR_EN.md)
- [SDK Reference](./docs/SDK_REFERENCE_EN.md)
- [Configuration](./docs/CONFIGURATION_EN.md)
- [Build & Deploy](./docs/BUILD_AND_DEPLOY_EN.md)
- [World Time Tick Reference](./docs/WORLD_TIME_TICK_REFERENCE_EN.md)

---

## License

MIT
