# GameAgentDevCli Guide

[**中文**](./GUIDE_GAMEAGENTDEVCLI.md) | **English**

GameAgentDevCli is the command-line tool for operating GameAgentEngine over its HTTP API.

---

## Global Flags

- `--server`, `-s`: engine base URL
- `--key`, `-k`: API key
- `--config`: local config file path for local-only operations such as reset
- `--idempotency-key`: idempotency header for write requests
- `--memory-limit`
- `--max-analysis-rounds`
- `--max-context-depth`
- `--include-related-nodes`

---

## Import

```bash
GameAgentDevCli import tools/source/demo-world.yaml --reset
GameAgentDevCli import tools/source/demo-world.yaml --dry-run
```

---

## Node Commands

```bash
# Create
GameAgentDevCli node create --world <world-id> --name "Council Hall" --type location

# Read
GameAgentDevCli node get <node-id>
GameAgentDevCli node list --world <world-id>

# Update
GameAgentDevCli node update <node-id> --name "New Name"

# Move
GameAgentDevCli node update <node-id> --parent <new-parent-id>
GameAgentDevCli node update <node-id> --clear-parent

# Copy
GameAgentDevCli node copy <node-id>
GameAgentDevCli node copy <node-id> --name "Copied Node"
GameAgentDevCli node copy <node-id> --with-children=false

# Delete
GameAgentDevCli node delete <node-id>
```

`node copy` uses subtree copy by default.

---

## World Commands

```bash
# Rename world
GameAgentDevCli world update <world-id> --name "Renamed World"

# Fork working copy
GameAgentDevCli world fork <world-id> [name] [--lock-world]

# Save snapshot
GameAgentDevCli world save <world-id> [name] [--lock-world]

# Restore from snapshot
GameAgentDevCli world restore <snapshot-world-id> [name] [--lock-world]

# Snapshot inspection
GameAgentDevCli world validate-snapshot <snapshot-world-id>
GameAgentDevCli world snapshot-info <snapshot-world-id>
GameAgentDevCli world list-snapshots <world-id>
GameAgentDevCli world delete-snapshot <snapshot-world-id>

# Plan review
GameAgentDevCli world plan pending
GameAgentDevCli world plan pending <world-id>
GameAgentDevCli world plan approve <world-id> <plan-id>
GameAgentDevCli world plan reject <world-id> <plan-id>
```

---

## Runtime Commands

```bash
GameAgentDevCli world tick <world-id>
GameAgentDevCli world event-impact <world-id> --type crisis --description "..."
GameAgentDevCli world scope-advance <world-id> <scope-id>
GameAgentDevCli world replan <world-id>
```

---

## Continuity State and Timeline

```bash
GameAgentDevCli state list <world-id>
GameAgentDevCli state get <world-id> world_state
GameAgentDevCli state get <world-id> story_state
GameAgentDevCli state get <world-id> story_history
GameAgentDevCli state get <world-id> tick_policy
GameAgentDevCli state set <world-id> tick_policy --data '{"continuity_rules":["Do not discard established underground reactor facts."]}'

GameAgentDevCli timeline latest <world-id>
GameAgentDevCli timeline list <world-id> --limit 5
```

Use these commands when you want to inspect or patch the continuity state that feeds the next `world_tick`.

---

## World Settings and Policy

```bash
# Settings
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world settings set <world-id> --pipeline-mode polling

# Policy
GameAgentDevCli world policy get <world-id>
GameAgentDevCli world policy set <world-id> --blocked spawn_item --safe add_memory
```

`world settings set` is a partial update command. Only explicitly passed flags are changed.

World tick progression controls:

```bash
GameAgentDevCli tick <world-id> --requested-ticks 3
GameAgentDevCli world tick <world-id> --requested-ticks 3
```

Use `--requested-ticks` together with `world_time_settings.tick_scale_mode`.
In `fixed` mode, the engine only accepts `1`. In `flexible` mode, the final adopted value is returned as `advanced_ticks`.

---

## Memory Propagation

```bash
GameAgentDevCli memory propagate <memory-id>
GameAgentDevCli memory propagate <memory-id> --mode tag_broadcast --tags rumor,politics
GameAgentDevCli memory propagate <memory-id> --mode targeted --target node-a,node-b
GameAgentDevCli memory propagate <memory-id> --max-depth 2 --publish-up
```

---

## Async Action Callback

```bash
GameAgentDevCli action callback <callback-id>
GameAgentDevCli action callback <callback-id> --status failed
GameAgentDevCli action callback <callback-id> --status success --result '{"item_id":"sword-01","quality":"rare"}'
```

`--result` is parsed as JSON first; if parsing fails, it is reported as raw text.

---

## Logs and Debug Traces

```bash
GameAgentDevCli logs --world <world-id> --limit 10
GameAgentDevCli logs --world <world-id> --limit 10 --json
GameAgentDevCli logs --world <world-id> --task-type world_tick --category pipeline --event llm_response_received --mode debug --request-id <request-id> --details

GameAgentDevCli debug traces --world <world-id> --limit 10
GameAgentDevCli debug traces --world <world-id> --limit 10 --json
GameAgentDevCli debug continuity <world-id>
GameAgentDevCli debug continuity <world-id> --mode debug --request-id <request-id> --log-limit 20 --trace-limit 10
```

`logs` now supports server-side filters such as `--node`, `--category`, `--event`, `--mode`, `--request-id`, and `--round`.

`debug continuity` is the fastest way to load the latest timeline, recent timeline history, continuity state components, recent `world_tick` logs, and debug traces in one summary view.
The summary now expands `advanced_ticks`, the current `world_time` label, and the `previous_world_time` label when the timeline payload carries world time state.

---

## Creator Entry

```bash
GameAgentDevCli inspect
```

Use this when your environment exposes the Creator inspection flow.
