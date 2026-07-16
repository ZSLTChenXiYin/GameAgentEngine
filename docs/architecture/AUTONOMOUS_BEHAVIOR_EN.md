# Autonomous Behavior System

[**中文**](./AUTONOMOUS_BEHAVIOR.md) | **English**

The autonomous behavior system allows a node to trigger its own reasoning and action loop without direct user input. It is intended for active entities such as NPCs, organizations, facilities, and world-scope controller nodes.

---

## Core Concepts

Autonomous behavior is configured through the `autonomous` component attached to a node.

Each node can independently define:

- whether autonomous behavior is enabled
- the trigger mode
- a capability allowlist
- its scheduling interval and last-run state

Current trigger modes:

| Trigger Mode | Value | Description |
|---|---|---|
| Manual | `manual` | Triggered only through API or DevCli |
| Tick Sync | `world_tick_sync` | Triggered after world tick advancement |
| Scheduled | `scheduled` | Triggered by the background scheduler |

---

## Execution Path

A typical autonomous execution goes through these stages:

1. Load the node's `autonomous` config and capability allowlist.
2. Build the `autonomous_act` task context.
3. Enter the shared pipeline reasoning loop.
4. Validate that emitted actions stay within the allowlist.
5. Execute sync actions, register async actions, and persist memory / propagation effects.
6. Persist structured logs into the unified `logs` table.

If the node belongs to a world, the flow also passes through the same-world exclusion boundary for critical heavy operations.

---

## Capability Allowlist

Each autonomous node can declare the actions it is allowed to invoke. The engine validates LLM output against this allowlist and rejects unauthorized actions.

Example:

```json
{
  "enabled": true,
  "trigger": "world_tick_sync",
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "Record a judgment",
      "schema": {
        "node_id": { "type": "string", "required": true },
        "content": { "type": "string", "required": true }
      }
    }
  ]
}
```

Validation failures are not silently ignored. The action is rejected and recorded for diagnosis.

---

## Configuration

### DevCli

```bash
# Read current config
GameAgentDevCli node autonomous get <node-id>

# Enable and switch to world tick sync mode
GameAgentDevCli node autonomous set <node-id> --enabled --trigger "world_tick_sync"

# Disable autonomous behavior
GameAgentDevCli node autonomous disable <node-id>

# Trigger one autonomous execution manually
GameAgentDevCli node autonomous run <node-id>
```

### API

Related endpoints:

- `GET /api/v1/nodes/{node_id}/autonomous`
- `PUT /api/v1/nodes/{node_id}/autonomous`
- `POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run`

Example request body:

```json
{
  "enabled": true,
  "trigger": "scheduled",
  "interval_seconds": 600,
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "Record an observation"
    }
  ]
}
```

---

## Scheduler

The background autonomous scheduler is a service-level static toggle controlled by config:

```yaml
engine:
  autonomous_scheduler_enabled: false
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

The scheduler only attempts nodes that satisfy all of the following:

1. The node has an `autonomous` component.
2. `enabled = true`.
3. `trigger = "scheduled"`.
4. The elapsed time since the previous run exceeds `interval_seconds`.

Scheduled autonomous work in the same world still passes through the same-world exclusion boundary for critical heavy operations, which helps reduce SQLite contention while preserving business consistency.

---

## Relationship to the Database Pipeline

Autonomous write paths now use the shared database pipeline:

- SQLite: single writer, WAL, batched logs, batched memory writes
- MySQL / PostgreSQL: pooled concurrent writes with the shared retry layer
- critical same-world operations: protected by the world-level business lock

Because of this, autonomous behavior no longer benefits only from lower frequency. It directly benefits from shared transactions, batching, and retry recovery.

---

## Diagnostics

When investigating autonomous behavior issues, start with:

- `GET /api/v1/logs`
- `GET /api/v1/pipeline/stats`
- `GET /debug/traces`

Useful fields to watch:

- `task_type = autonomous_act`
- `event_name = autonomous_node_started|autonomous_node_completed|autonomous_node_failed`
- `request_id`
- `round`
- `execution_mode`

If you suspect scheduler congestion or lock contention, also inspect pipeline stats for:

- write retry counters
- transaction counters
- log sink queue depth
- world-lock contention stats
