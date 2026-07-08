# Autonomous Behavior System

[**中文**](./AUTONOMOUS_BEHAVIOR.md) | **English**

GameAgentEngine v0.4.5's autonomous behavior system allows NPCs and nodes to decide and execute actions on their own, without direct user input.

---

## How It Works

Autonomous behavior is configured through an `autonomous` component attached to a node. Each node can independently configure whether autonomous behavior is enabled, its trigger mode, and a capability allowlist.

### Trigger Modes

| Trigger Mode | Value | Description |
|---|---|---|
| Manual | `manual` | Triggered only via API or DevCli |
| Tick Sync | `world_tick_sync` | Automatically triggered during world Tick advancement |
| Scheduled | `scheduled` | Automatically triggered at configured intervals (requires the background scheduler) |

### Capability Allowlist

Each autonomous node can declare a list of actions it is allowed to invoke (capabilities). The engine validates whether the LLM's output actions are within the allowlist, preventing unauthorized behavior:

```json
{
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "Record a short-term judgment",
      "schema": {
        "node_id": {"type": "string", "required": true},
        "content": {"type": "string", "required": true}
      }
    }
  ]
}
```

---

## Configuration Examples

### Configuring via DevCli

```bash
# View current configuration
GameAgentDevCli node autonomous get <node-id>

# Enable and configure as world Tick sync trigger
GameAgentDevCli node autonomous set <node-id> --enabled --trigger "world_tick_sync"

# Disable
GameAgentDevCli node autonomous disable <node-id>

# Manually trigger once
GameAgentDevCli node autonomous run <node-id>
```

### Configuring via API

```json
// GET /api/v1/nodes/{node_id}/autonomous
// PUT /api/v1/nodes/{node_id}/autonomous

{
  "enabled": true,
  "trigger": "world_tick_sync",
  "capabilities": [
    {"id": "add_memory", "mode": "sync", "description": "Record a judgment"}
  ]
}
```

---

## Scheduler Configuration

The background autonomous behavior scheduler is a **service-level static toggle** controlled through the config file:

```yaml
engine:
  autonomous_scheduler_enabled: false           # Global toggle
  autonomous_scheduler_interval_seconds: 300    # Scan interval
  autonomous_scheduler_max_nodes_per_world: 10  # Max triggers per world per scan
```

When the scheduler scans, it only triggers nodes that meet all of the following conditions:
1. Has an `autonomous` component attached
2. `enabled = true`
3. `trigger = "scheduled"`
4. Time since last run exceeds `interval_seconds`