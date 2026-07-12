# GameAgentEngine

[**中文**](./README.md) | **English**

This README is bundled with packaged builds.

---

## Core Capabilities

- **World Graph Model** — Unified world modelling with nodes, components, memories, and relations
- **LLM Inference Pipeline** — Three pipeline modes (vertical / polling / full), multi-round inference, sub-task DAG, async action callbacks
- **World Time and Continuity** — Tick advancement, continuity state components, timeline archives, world-time scale system
- **External Interaction and Runtime Tasks** — Push / Pull / Hybrid async task delivery modes, HTTP / RPC / WebSocket adapters
- **Memory Propagation** — Five propagation modes (upward / environment_scope / organization_scope / tag_broadcast / targeted)
- **Working Copies and Snapshots** — Fork, Snapshot, Restore three world-copy semantics
- **Background Autonomous Scheduling** — Node-level autonomous component with manual, tick-sync, and scheduled triggers
- **Database Pipeline** — Unified write transactions, batched logging, retriable writes; supports SQLite / MySQL / PostgreSQL
- **Three Developer Tools** — GameAgentCreator (Web editor), GameAgentDevCli (CLI), Go SDK

---

## Shortest Path

```bash
# 1. Start the engine
GameAgentEngine serve

# 2. Create a world
GameAgentDevCli node create --type world --name "New World"

# 3. Open Creator
GameAgentDevCli inspect
```

> If you want time advancement and worldline reasoning, configure `world_time_settings` first.

---

## Toolchain

### GameAgentCreator

Web-based graphical editor. Pages: Worlds / Snapshots / Tasks / Plans / Policy / Settings / Continuity / State / Timelines / Logs / Traces

### GameAgentDevCli

CLI development tool: node / component / memory / relation CRUD; world tick, event impact, plan approval, snapshot management; task management (task command); debugging and observability

### Go SDK

Go server-side SDK: full API wrapper including runtime task management, memory propagation, autonomous behavior configuration

---

## Documents

All docs are located under the `docs/` directory:

- Getting Started / API Reference / Configuration
- DevCli Guide / Creator Guide / SDK Reference
- Architecture / Core Concepts / Pipeline Internals
- Autonomous Behavior / World Time Tick Reference / External Interaction Roadmap
- Build and Deploy

---

## License

MIT
