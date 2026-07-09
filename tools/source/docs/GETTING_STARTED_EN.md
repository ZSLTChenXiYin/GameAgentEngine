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
GameAgentDevCli inspect
```

If you want world time advancement, configure `world_time_settings` first in Creator's `Settings` page.

---

## Config File

Edit the bundled `gameagentengine.conf.yaml` in the current directory.

Key defaults:

- `auth.api_key: dev-key`
- `llm.model: deepseek-v4-flash`
- `llm.base_url: https://api.deepseek.com`
- `engine.execution_mode: debug`
- `engine.autonomous_scheduler_enabled: true`
