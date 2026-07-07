# GameAgentEngine

[**中文**](./README.md) | **English**

An AI agent creation and runtime engine for game developers.

GameAgentEngine is a Go-based engine that sits between game logic and LLM capabilities. It handles world modeling, NPC behavior, memory management, world timeline progression, and controlled runtime actions.

It does not replace Unity, Unreal, or Godot. It is designed to work alongside them.

---

## Highlights

- Unified world graph based on nodes, components, memories, and relations
- LLM-driven NPC dialogue and world reasoning
- Tick advancement, event impact assessment, and scope-level evolution
- Three pipeline modes: `vertical`, `polling`, `full`
- World copy semantics split into working-copy fork, save snapshot, and restore
- World settings and world policy managed dynamically at runtime
- Web editor (`GameAgentCreator`) and CLI (`GameAgentDevCli`)
- Go SDK for integrating the engine into tools and services
- Observability through response metadata, logs, and debug traces

---

## Quick Start

```bash
# 1. Clone and build
git clone <repo-url>
cd GameAgentEngine
go build ./...

# 2. Copy the default config
cp tools/source/gameagentengine.conf.yaml .

# 3. Fill in llm.api_key in gameagentengine.conf.yaml

# 4. Start the engine
go run ./cmd/gameagentengine serve

# 5. Import the demo world
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml --reset

# 6. Open the Creator UI
# tools/source/web/GameAgentCreator/index.html
```

See [Getting Started](./docs/GETTING_STARTED_EN.md) for the full walkthrough.

---

## Current Tooling

### GameAgentCreator

The Creator currently supports:

- world selection and world creation
- world rename from the world page
- node tree browsing with collapse state
- drag-and-drop node reparenting
- drag-to-root reparenting
- node creation, edit, delete, and copy
- snapshot save, validation, restore, and delete flows
- world settings, world policy, logs, and traces

### GameAgentDevCli

The CLI currently supports:

- import and validation flows
- node / component / memory / relation CRUD
- world tick, event impact, scope advance, and timeline replan
- world fork, save snapshot, restore, validate snapshot, snapshot metadata, and snapshot deletion
- world runtime settings and world policy management
- world rename via `world update`
- node copy via `node copy`

---

## World Copy Semantics

GameAgentEngine distinguishes three related but different world-copy operations:

- `ForkWorld`: create a runnable working copy for branch simulation and editing
- `CreateWorldSnapshot`: create a save-oriented snapshot world with compatibility metadata
- `RestoreWorld`: validate a save snapshot and materialize a fresh runnable world from it

Snapshots preserve compatibility metadata separately from the live world graph so the engine can validate restore safety before rebuilding runtime state.

---

## Pipeline Modes

Each world can choose one of three pipeline modes:

- `vertical`: minimal, single-pass execution
- `polling`: multi-round reasoning without the full orchestration surface
- `full`: complete orchestration, including the heavier engine features

This lets games use only the amount of pipeline they need, improving response efficiency and reducing runtime cost.

---

## Project Structure

```text
GameAgentEngine/
|-- cmd/
|   |-- gameagentengine/
|   `-- gameagentdevcli/
|-- docs/
|-- internal/
|   |-- action/
|   |-- api/
|   |-- config/
|   |-- engine/
|   |-- llm/
|   |-- planner/
|   |-- service/
|   `-- store/
|-- sdk/
|-- tools/
|   `-- source/
|       `-- web/
|           `-- GameAgentCreator/
`-- web/
```

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

---

## License

MIT
