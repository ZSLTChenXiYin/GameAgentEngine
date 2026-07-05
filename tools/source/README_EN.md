# GameAgentEngine

[**中文**](./README.md) | **English**

**An AI Agent Creation & Runtime Engine for Game Developers.**

GameAgentEngine is a Go-based engine that sits between your game logic and LLM capabilities — responsible for world modeling, NPC intelligent behavior, memory management, and world timeline progression. Think of it as the **director system and intelligence layer** inside a game world.

> It does **not** replace Unity, Unreal, or Godot. It works _alongside_ them.

---

## Features

- **Unified World Modeling** — Nodes, Components, Memories, and Relations form a complete entity graph
- **NPC Intelligence** — LLM-powered dialogue with contextual awareness (profile, lore, memories, relations)
- **World Timeline** — Tick-based advancement with event impact assessment
- **Inference Pipeline** — Context assembly → Prompt generation → LLM call → Action parsing → Memory persistence
- **Action System** — Built-in sync/async actions (add_memory, update_mood, send_dialogue, adjust_relation, spawn_item)
- **Policy Engine** — Guardrails for safe action execution (blocked actions, review thresholds)
- **World Cloning** — Duplicate an entire world with all its data, with optional source world locking to prevent concurrent writes
- **Idempotency** — Safe retry for write operations
- **Full CRUD API** — 20+ RESTful endpoints for nodes, components, memories, and relations
- **Dual Storage** — SQLite (dev) and MySQL (production)
- **Go SDK** — Native Go client library with Agent builder API
- **GameAgentDevCli** — Developer CLI for scripted import, CRUD, world tick, verification
- **GameAgentCreator** — Web-based visual editor with node tree, inspector, logs, and import

---

## Quick Start

```bash
# 1. Clone and build
git clone <repo-url>
cd GameAgentEngine
go build ./...

# 2. Configure (copy default config and add LLM API key)
cp tools/source/gameagentengine.conf.yaml .
# Edit gameagentengine.conf.yaml — set llm.api_key

# 3. Start the engine service
go run ./cmd/gameagentengine serve

# 4. In another terminal, seed the demo world
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key --config gameagentengine.conf.yaml demo-seed

# 5. Open the visual editor
# web/GameAgentCreator/index.html
```

See [docs/GETTING_STARTED.md](docs/GETTING_STARTED_EN.md) for a detailed walkthrough.

---

## Project Structure

```
GameAgentEngine/
├── cmd/
│   ├── gameagentengine/      # Engine server + CLI (serve, validate, version, import, ...)
│   └── gameagentdevcli/      # Developer CLI (CRUD, world tick, import, verify, snapshot, ...)
├── internal/
│   ├── api/                  # HTTP API layer (router, handlers, middleware, error mapping)
│   ├── service/              # Domain rules & transaction boundaries
│   ├── engine/               # Inference pipeline, context builder, core types
│   ├── store/                # GORM persistence (Node, Component, Memory, Relation, Timeline, Log)
│   ├── llm/                  # LLM Provider (OpenAI-compatible + Mock)
│   ├── action/               # Action registry & callback system
│   ├── planner/              # Policy engine & world change plan evaluation
│   └── config/               # Viper configuration loading
├── sdk/                      # Go HTTP Client SDK
├── web/
│   ├── GameAgentCreator/     # Visual editor (node tree, inspector, logs, import)
│   └── Demo/                 # Demo showcase page (Gray Harbor Council)
├── tools/
│   ├── scripts/              # Build scripts, encoding checks
│   └── source/               # Default config
└── docs/                     # Documentation
```

---

## Tools Overview

| Tool | Purpose | Entry Point |
|---|---|---|
| **GameAgentEngine** | Backend engine service + CLI | `cmd/gameagentengine/main.go` |
| **GameAgentDevCli** | Developer command-line tool | `cmd/gameagentdevcli/main.go` |
| **GameAgentCreator** | Web-based visual editor | `web/GameAgentCreator/index.html` |
| **Web Demo** | Demo showcase page | `web/Demo/index.html` |

---

## Documentation

| Document | Description |
|---|---|
| [Getting Started](docs/GETTING_STARTED_EN.md) | Installation, configuration, first run |
| [Architecture](docs/ARCHITECTURE_EN.md) | System design, layers, data flow |
| [Core Concepts](docs/CORE_CONCEPTS_EN.md) | Nodes, Components, Memories, Relations, Tasks |
| [Autonomous Behavior](docs/AUTONOMOUS_BEHAVIOR_EN.md) | Optional node-level autonomous actions and capability allowlists |
| [API Reference](docs/API_REFERENCE_EN.md) | Full HTTP API endpoint reference |
| [GameAgentDevCli Guide](docs/GUIDE_GAMEAGENTDEVCLI_EN.md) | Complete CLI command reference |
| [GameAgentCreator Guide](docs/GUIDE_GAMEAGENTCREATOR_EN.md) | Web UI usage guide |
| [Configuration](docs/CONFIGURATION_EN.md) | Config file reference |
| [SDK Reference](docs/SDK_REFERENCE_EN.md) | Go SDK API reference |
| [Build & Deploy](docs/BUILD_AND_DEPLOY_EN.md) | Build, packaging, deployment |
| [Pipeline Internals](docs/PIPELINE_INTERNALS_EN.md) | Inference pipeline deep dive |
| [Demo World: Gray Harbor](docs/DEMO_WORLD_GRAY_HARBOR_EN.md) | Demo world guide |

---

## Tech Stack

| Layer | Tech | Purpose |
|---|---|---|
| Language | Go 1.25+ | Core engine |
| HTTP | net/http, http.ServeMux | Service interface |
| ORM | GORM v2 | Database access |
| Storage | SQLite / MySQL | Persistence |
| AI | OpenAI-compatible API | LLM inference |
| CLI | Cobra | Command-line framework |
| Config | Viper | Configuration management |

---

## License

MIT