# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide is for users consuming the bundled GameAgentEngine resource package. It helps you configure the engine, import the demo world, and open Creator.

---

## What You Need

- a runnable `GameAgentEngine`
- a runnable `GameAgentDevCli`
- an OpenAI-compatible API key if you want real LLM responses

If `llm.api_key` is empty, the engine automatically falls back to the mock provider for local testing.

---

## Configure the Engine

Edit the bundled `gameagentengine.conf.yaml` in the current directory and set at least:

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-your-key"
  base_url: "https://api.deepseek.com/v1"
```

See [Configuration](./CONFIGURATION_EN.md) for the rest of the settings.

---

## Start the Engine

```bash
GameAgentEngine serve
```

Verify the service with the health check:

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
GameAgentDevCli import demo-world.yaml --reset
```

Useful variants:

```bash
# Validate only
GameAgentDevCli import demo-world.yaml --dry-run

# Keep existing data
GameAgentDevCli import demo-world.yaml
```

---

## Common Runtime Commands

```bash
# List worlds
GameAgentDevCli node list --type world

# Advance one world tick
GameAgentDevCli world tick <world-id>

# Create a save snapshot
GameAgentDevCli world save <world-id> "Save Slot 1"

# Rename a world
GameAgentDevCli world update <world-id> --name "Renamed World"

# Copy a node
GameAgentDevCli node copy <node-id> --name "Copied Node"
```

See the [GameAgentDevCli Guide](./GUIDE_GAMEAGENTDEVCLI_EN.md) for the full command set.

---

## Open Creator

Open this file in a browser:

`web/GameAgentCreator/index.html`

The current Creator supports:

- world creation and rename
- drag-to-parent and drag-to-root node reparenting
- node create, edit, delete, and copy
- snapshot save, validate, restore, and delete
- world settings, world policy, logs, and traces

See the [GameAgentCreator Guide](./GUIDE_GAMEAGENTCREATOR_EN.md) for details.

---

## Read Next

- [Core Concepts](./CORE_CONCEPTS_EN.md)
- [API Reference](./API_REFERENCE_EN.md)
- [SDK Reference](./SDK_REFERENCE_EN.md)
- [Demo World: Gray Harbor](./DEMO_WORLD_GRAY_HARBOR_EN.md)
