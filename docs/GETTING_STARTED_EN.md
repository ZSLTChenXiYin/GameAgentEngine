# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide walks through bringing up GameAgentEngine, importing a demo world, and opening the Creator UI.

---

## Prerequisites

- Go 1.25+
- Git
- An OpenAI-compatible API key if you want real LLM responses

If `llm.api_key` is empty, the engine falls back to the mock provider for local testing.

---

## Build

```bash
git clone <repo-url>
cd GameAgentEngine
go build ./...
```

---

## Configure

```bash
cp tools/source/gameagentengine.conf.yaml .
```

Edit `gameagentengine.conf.yaml` and set at least:

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-your-key"
  base_url: "https://api.deepseek.com/v1"
```

---

## Start the Engine

```bash
go run ./cmd/gameagentengine serve
```

Check the health endpoint:

```bash
curl http://127.0.0.1:8080/health
```

Expected result:

```json
{"status":"ok"}
```

---

## Import the Demo World

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml --reset
```

Useful variants:

```bash
# Validate only
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml --dry-run

# Keep existing data
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml
```

---

## Basic Runtime Commands

```bash
# List worlds
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node list --type world

# Advance a world tick
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world tick <world-id>

# Save a snapshot
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world save <world-id> "Save Slot 1"

# Rename a world
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world update <world-id> --name "Renamed World"
```

---

## Open Creator

Open this file in a browser:

`tools/source/web/GameAgentCreator/index.html`

Creator currently supports:

- world creation and rename
- node tree editing and drag-and-drop reparenting
- node copy
- snapshot save / validate / restore / delete
- world settings and policy editing
- logs and traces

---

## Next Documents

- [GameAgentCreator Guide](./GUIDE_GAMEAGENTCREATOR_EN.md)
- [GameAgentDevCli Guide](./GUIDE_GAMEAGENTDEVCLI_EN.md)
- [API Reference](./API_REFERENCE_EN.md)
- [SDK Reference](./SDK_REFERENCE_EN.md)
