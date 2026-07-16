# GameAgentDevCli Guide

[**中文**](./GUIDE_GAMEAGENTDEVCLI.md) | **English**

Key packaged DevCli capabilities:

- create worlds and nodes
- configure `world_settings` and `world_time_settings`
- advance world ticks
- inspect `world_time_state`, timelines, logs, and continuity
- open Creator

---

## Common Commands

```bash
GameAgentDevCli node create --type world --name "New World"
GameAgentDevCli world settings set <world-id> --world-time-settings-file world-time.json
GameAgentDevCli world tick <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli creator
```
