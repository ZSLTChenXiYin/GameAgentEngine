# Getting Started

[**中文**](./GETTING_STARTED.md) | **English**

This guide is for new developers using the packaged build.

---

## Shortest Path

```bash
# 1. Start the service
GameAgentEngine serve

# 2. Create a world
GameAgentDevCli node create --type world --name "New World"

# 3. Open Creator
GameAgentDevCli creator
```

If you want world time advancement, configure `world_time_settings` first in Creator's `Settings` page.

---

## Config File

Edit the bundled `gameagentengine.conf.yaml` in the current directory.

Current packaged template highlights:

- `auth.api_key: dev-key`
- `llm.model: deepseek-v4-flash`
- `llm.base_url: https://api.deepseek.com`
- `engine.execution_mode: debug`
- `engine.autonomous_scheduler_enabled: false`
- `engine.world_lock_enabled: true`

If you omit these fields, the engine still falls back to code-level defaults such as:

- `llm.model: gpt-4o-mini`
- `llm.base_url: https://api.openai.com/v1`
- `engine.execution_mode: full`

---

## Common Commands

```bash
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world tick <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
```

---

## Diagnostics

Recommended endpoints:

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`
