# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide is for developers who are new to GameAgentEngine. The goal is to get you from zero to three things quickly: start the Engine, create your first world, and open Creator to continue editing.

---

## What You Will Have After This

By the end of this guide, you should be able to:

- start a local Engine service
- create a world root node with DevCli
- continue editing nodes, components, and relations in Creator
- understand when `world_time_settings` must be configured first

---

## Prerequisites

- Go 1.25+
- a working terminal environment
- an OpenAI-compatible API key if you want real model calls

If `llm.api_key` is empty, the engine falls back to the Mock Provider. That is fine for local flow verification, but not for validating real reasoning quality.

---

## Step 1: Build the Project

```bash
git clone <repo-url>
cd GameAgentEngine
go build ./...
```

If you are using a packaged build, you can work directly from the extracted directory instead of rebuilding.

---

## Step 2: Prepare the Config File

Copy the default config:

```bash
cp tools/source/gameagentengine.conf.yaml .
```

Important defaults in the current config:

- default listen address: `0.0.0.0:8080`
- default API key: `dev-key`
- default model: `deepseek-v4-flash`
- default `base_url`: `https://api.deepseek.com`
- default execution mode: `debug`
- background autonomous scheduler: enabled by default

At minimum, check these fields:

```yaml
auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-v4-flash"
  api_key: "sk-xxx"
  base_url: "https://api.deepseek.com"
```

If you do not want to call a real model yet, leave `llm.api_key` empty.

---

## Step 3: Start the Engine

```bash
go run ./cmd/gameagentengine serve
```

Confirm the service is healthy:

```bash
curl http://127.0.0.1:8080/health
```

Expected result:

```json
{"status":"ok"}
```

---

## Step 4: Create Your First World

The beginner flow now starts by creating a world root node directly.

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --type world --name "New World"
```

This creates a `world` node that acts as the root of the world tree.

You can then create child nodes, for example:

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --world <world-id> --type location --name "Starter Village"
```

If you prefer code-driven world setup, you can also use `Agent.CreateWorld()` from the SDK.

---

## Step 5: Open Creator

Recommended path:

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key inspect
```

If that entry is not convenient in your environment, open this file directly:

`tools/source/web/GameAgentCreator/index.html`

Core pages you will use first:

- `Worlds`: world and node tree
- `Settings`: world runtime settings
- `Policy`: world policy
- `Plans`: pending plan approvals
- `State`: continuity state components
- `Timelines`: archived world ticks
- `Continuity`: continuity debugging overview
- `Logs` / `Traces`: observability and debugging

---

## Step 6: Configure World Time Before Running Tick

If you want to use any of the following:

- `world tick`
- timeline advancement
- world time evolution inside continuity state
- worldline reasoning

configure `world_time_settings` first in the `Settings` page.

This is an intentional guardrail in the current design. Without a defined world time system, the Engine cannot reliably advance time or maintain timeline continuity, so dependent flows are intentionally blocked to remind developers to configure it first.

You can also configure it through DevCli:

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world settings set <world-id> --world-time-settings-json '{"tick_scale_mode":"flexible","tick_min_unit":"hour","tick_step":1,"tick_units":["day","hour"]}'
```

Minimum rules:

- `tick_units` must be ordered from large to small
- `tick_min_unit` must match the last unit
- `tick_scale_mode` currently must be `fixed` or `flexible`

---

## Step 7: Run a Tick

Once world time is configured, you can advance a world tick:

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world tick <world-id> --type manual --time "day-1" --requested-ticks 1
```

To limit how many autonomous nodes can run during this tick:

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world tick <world-id> --autonomous-limit 2
```

After that, inspect the result in Creator:

- `State` for `world_time_state`
- `Timelines` for the latest tick archive
- `Continuity` for the aggregated continuity bundle and diff

---

## Common Follow-Up Commands

```bash
# List worlds
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node list --type world

# Read world settings
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world settings get <world-id>

# Read continuity state components
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key state list <world-id>

# Read the latest timeline archive
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key timeline latest <world-id>

# Read recent logs
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key logs --world <world-id> --limit 10
```

---

## What To Read Next

- [Configuration](./CONFIGURATION_EN.md)
- [GameAgentCreator Guide](./GUIDE_GAMEAGENTCREATOR_EN.md)
- [GameAgentDevCli Guide](./GUIDE_GAMEAGENTDEVCLI_EN.md)
- [World Time Tick Reference](./WORLD_TIME_TICK_REFERENCE_EN.md)
- [SDK Reference](./SDK_REFERENCE_EN.md)
