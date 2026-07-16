# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide is for developers who are new to GameAgentEngine. The goal is to get you from zero to three things quickly: start the Engine, create your first world, and open Creator to continue editing.

---

## What You Will Have After This

By the end of this guide, you should be able to:

- start a local Engine service
- create a world root node with DevCli
- continue editing nodes, components, and relations in Creator
- use Worker for local play REPL or game-side async simulation
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
# Windows: tools\scripts\build.bat
# Linux/macOS: bash tools/scripts/build.sh
#
# If you are using a packaged build, skip this step and work directly from the extracted directory.
```

If you are using a packaged build, skip this step and work directly from the extracted directory.

---

## Step 2: Prepare the Config File

Copy the default config:

```bash
cp tools/source/gameagentengine.conf.yaml .
```

Important points in the packaged template:

- default listen address: `0.0.0.0:8080`
- default API key: `dev-key`
- sample model in the template: `deepseek-v4-flash`
- sample `base_url` in the template: `https://api.deepseek.com`
- template execution mode: `debug`
- background autonomous scheduler: disabled by default

There is also an important distinction between template values and code-level fallback defaults. If fields are omitted, the engine falls back to internal defaults such as:

- `llm.model = gpt-4o-mini`
- `llm.base_url = https://api.openai.com/v1`
- `engine.execution_mode = full`

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
GameAgentEngine serve
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
GameAgentDevCli node create --type world --name "New World"
```

This creates a `world` node that acts as the root of the world tree.

You can then create child nodes, for example:

```bash
GameAgentDevCli node create --world <world-id> --type location --name "Starter Village"
```

---

## Step 5: Open Creator

The simplest path is through DevCli:

```bash
GameAgentDevCli creator
```

If that launcher path is inconvenient in your environment, you can also open:

`tools/source/web/GameAgentCreator/index.html`

Key pages include:

- `Worlds`
- `Settings`
- `Policy`
- `Plans`
- `State`
- `Timelines`
- `Continuity`
- `Logs` / `Traces`

---

## Step 6: Configure World Time Before Running Tick

If you plan to use:

- `world tick`
- timeline advancement
- world-time continuity state
- worldline reasoning

you should configure `world_time_settings` first in the `Settings` page.

This is a deliberate hard requirement in the current design. Without a valid world-time system, the engine cannot reliably advance time or maintain continuity reasoning, so related save / advance flows intentionally stop and ask you to finish the configuration first.

---

## Step 7: Watch Pipeline State Early

Once you start load-testing, running ticks, or enabling autonomous behavior, it helps to monitor:

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`

These endpoints quickly show whether:

- write retries are becoming frequent
- the batched log queue is backing up
- world-level lock contention is increasing

---

## Step 8: Use Worker to Validate the Game-Side Loop

If you want to verify the external worker, authority state file, NPC dialogue flow, and callback path together, the shortest path is to start `GameAgentWorker` directly:

```bash
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

This command:

- loads YAML / JSON authority state
- selects the player node
- lets you use `/talk`, `/ask`, `/gift`, `/trade`, and related commands in the text-game REPL
- serves high-frequency authoritative facts such as HP, inventory, money, quest state, and scene occupancy during dialogue

If you are not testing play mode and only want the Runtime Task push / pull / callback loop, use:

```bash
GameAgentWorker serve --verbose
```

If you only want to validate the packaged scenarios in the current build, run:

```bash
GameAgentWorker test all
```
