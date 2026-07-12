# GameAgentEngine

[**中文**](./README.md) | **English**

This README is bundled with packaged builds.

---

## Why Games Need Their Own Agent Engine

General-purpose LLM Agents do not work well for games:

- **No world persistence** -- every conversation starts from scratch, NPCs have no memory
- **Uncontrolled actions** -- LLMs say anything; games need validated action boundaries
- **No world clock** -- games have their own time; agents have no time concept
- **High integration cost** -- every game engine needs custom push, callback, retry, fallback plumbing
- **No designer tools** -- most agent frameworks lack a visual editor

GameAgentEngine was designed to solve these problems.

---

## Core Capabilities

- **World Graph Model** -- nodes, components, memories, and relations
- **Controlled Actions** -- capability allowlist validation
- **World Time and Continuity** -- Tick advancement + continuity state archives
- **External Interaction** -- Push / Pull / Hybrid async task modes
- **Three Pipeline Modes** -- vertical / polling / full
- **Background Autonomous Behavior** -- NPCs act without player interaction
- **Database Pipeline** -- SQLite / MySQL / PostgreSQL

---

## Quick Start

```
GameAgentEngine serve
GameAgentDevCli node create --type world --name "New World"
GameAgentDevCli inspect
```

Configure `world_time_settings` in Creator's Settings page before running time advancement.

---

## Toolchain

- **GameAgentCreator** -- Visual Web editor (Worlds / Tasks / Plans / Settings / State / Timelines / Logs / Traces)
- **GameAgentDevCli** -- CLI dev tool (CRUD + task management + debugging)
- **Go SDK** -- Go server-side SDK (full API wrapper)

---

## Documentation

See the `docs/` directory:

- Getting Started / API Reference / Configuration
- DevCli Guide / Creator Guide / SDK Reference
- External Interaction Roadmap / External Interaction Examples
- Architecture / Core Concepts / Pipeline Internals
- Autonomous Behavior / World Time Tick Reference / Build and Deploy

---

## License

MIT
