# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide walks you through setting up GameAgentEngine v0.2.0 from scratch, configuring an LLM provider, importing the Demo world, and interacting with NPCs.

---

## Prerequisites

- **Go 1.25+** — [Download Go](https://go.dev/dl/)
- **LLM API Key** — An OpenAI-compatible API key (DeepSeek, OpenAI, Qwen, etc.)
- **Git** — To clone the repository

---

## Installation

```bash
# Clone the repository
git clone <repo-url>
cd GameAgentEngine

# Build all components
go build ./...

# Verify the build
GameAgentEngine version
# Output: GameAgentEngine version v0.2.0
```

---

## Configuration

Copy the default config file and fill in your LLM API key:

```bash
cp tools/source/gameagentengine.conf.yaml .
```

Edit `gameagentengine.conf.yaml` and set your LLM API key:

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"         # or gpt-4o-mini, qwen-turbo, etc.
  api_key: "sk-your-key-here"   # <-- set your key here
  base_url: "https://api.deepseek.com/v1"
```

> **No API Key?** If `api_key` is left empty, the engine defaults to the Mock LLM Provider, which returns fixed responses — ideal for testing the pipeline without real API calls.

---

## Start the Engine Service

```bash
GameAgentEngine serve
```

Expected output:
```
DB: sqlite (gameagentengine.db)
LLM: deepseek-chat (https://api.deepseek.com/v1)
listen on 0.0.0.0:8080
```

Verify the service is running:

```bash
curl http://127.0.0.1:8080/health
# {"status":"ok"}
```

---

## Import the Demo World

Open another terminal and import the Demo world "Gray Harbor Border":

```bash
GameAgentDevCli import tools/source/demo-world.yaml --reset
```

The `--reset` flag clears the database before importing. Use `--dry-run` to validate without writing.

This imports a complete world with:
- 4 factions (Gray Harbor Council, Iron Tide Merchant Guild, etc.)
- 8 locations (Council Hall, Round Table Chamber, Mist Bay Mine, etc.)
- 4 NPCs (Speaker Elrin, Steward Brahm, Commander Cyllo, Representative Mira)
- 14 weighted relations
- Custom components (resource_state, district_state, demo_state)

Verify the world was created:

```bash
GameAgentDevCli status
```

---

## Quick Commands

### List all nodes

```bash
GameAgentDevCli node list
```

### Interact with the world

```bash
# Advance world time
GameAgentDevCli world tick <world-id>

# Talk to an NPC (via REST API)
curl -X POST http://127.0.0.1:8080/api/v1/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key" \
  -d '{"world_id":"<world-id>","node_id":"<npc-id>","task_type":"npc_dialogue","context":{"messages":[{"role":"user","content":"Hello"}]}}'

# Evaluate event impact
GameAgentDevCli world event-impact <world-id> --type "crisis" --description "..." --severity "critical"

# Clone a world
GameAgentDevCli world save <world-id> "My save slot 1" --lock-world

# View world runtime settings
GameAgentDevCli world settings get <world-id>

# Switch pipeline mode only
GameAgentDevCli world settings set <world-id> --pipeline-mode "polling"

# Remove the upward propagation depth limit
GameAgentDevCli world settings set <world-id> --propagation-max-depth 0
```

*(See the [GameAgentDevCli Guide](GUIDE_GAMEAGENTDEVCLI_EN.md) for complete command reference.)*

---

## Open the Visual Editor

Open `web/GameAgentCreator/index.html` in a browser.

---

## Open the Demo Showcase

Open `web/Demo/index.html` in a browser.

This is the playable "Gray Harbor Border" Demo: talk to different NPCs, make decisions each round, and advance the world timeline.

---

## Next Steps

| Topic | Document |
|---|---|
| Understand core concepts | [Core Concepts](CORE_CONCEPTS_EN.md) |
| Explore all API endpoints | [API Reference](API_REFERENCE_EN.md) |
| Master the CLI | [GameAgentDevCli Guide](GUIDE_GAMEAGENTDEVCLI_EN.md) |
| Use the Web editor | [GameAgentCreator Guide](GUIDE_GAMEAGENTCREATOR_EN.md) |
| Learn about the Demo world | [Demo World: Gray Harbor](DEMO_WORLD_GRAY_HARBOR_EN.md) |
| Build and package for distribution | [Build & Deploy](BUILD_AND_DEPLOY_EN.md) |
| Use the Go SDK in your project | [SDK Reference](SDK_REFERENCE_EN.md) |
