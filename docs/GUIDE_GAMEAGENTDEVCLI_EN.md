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

---

## Memory Propagation

```bash
GameAgentDevCli memory propagate <memory-id>
GameAgentDevCli memory propagate <memory-id> --mode tag_broadcast --tags rumor,politics
GameAgentDevCli memory propagate <memory-id> --mode targeted --target node-a,node-b
GameAgentDevCli memory propagate <memory-id> --max-depth 2 --publish-up
```

---

## Logs and Debug Traces

```bash
GameAgentDevCli logs --world <world-id> --limit 10
GameAgentDevCli logs --world <world-id> --limit 10 --json

GameAgentDevCli debug traces --world <world-id> --limit 10
GameAgentDevCli debug traces --world <world-id> --limit 10 --json
```

---

## Creator Entry

```bash
GameAgentDevCli inspect
```

Use this when your environment exposes the Creator inspection flow.
