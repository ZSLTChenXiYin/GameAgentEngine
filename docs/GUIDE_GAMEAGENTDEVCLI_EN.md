# GameAgentDevCli Guide

[**中文**](./GUIDE_GAMEAGENTDEVCLI.md) | **English**

GameAgentDevCli is the command-line management tool for GameAgentEngine v0.2.0. It communicates with the engine service via the HTTP API and supports world management, node CRUD, component, memory, and relation operations, as well as world inference (Tick, event impact, scope advancement).

---

## Global Parameters

| Parameter | Short | Description |
|---|---|---|
| `--server <url>` | `-s` | Engine service address (default `http://127.0.0.1:8080`) |
| `--key <key>` | `-k` | API key (default `dev-key`) |
| `--config <path>` | | Local config file path (for local operations like reset) |
| `--memory-limit <n>` | | Inference memory limit (0 = use server-side config) |
| `--max-analysis-rounds <n>` | | Max LLM polling rounds |
| `--max-context-depth <n>` | | Max context traceback depth |
| `--include-related-nodes` | | Load related node data |
| `--idempotency-key <key>` | | Idempotency key |

---

## Command Overview

### status — Check service status

```bash
GameAgentDevCli status
```

### reset — Clear the local database

```bash
GameAgentDevCli reset --config gameagentengine.conf.yaml
```

### import — Import world configuration

Import world config in YAML/JSON format (supports `--dry-run` for validation-only and `--reset` for clearing before import):

```bash
# Import from file
GameAgentDevCli import demo-world.yaml

# Validate import content (no database write)
GameAgentDevCli import demo-world.yaml --dry-run

# Clear database before import
GameAgentDevCli import demo-world.yaml --reset

# Import from stdin
cat world.yaml | GameAgentDevCli import - --format yaml
```

See the Demo world example and Core Concepts documentation for import file format details.

---

### node — Node management

```bash
# Create a world node
GameAgentDevCli node create --name "MyWorld" --type "world"

# Create a child node
GameAgentDevCli node create --world <world-id> --name "Council Hall" --type "location" --parent <parent-id>

# List all nodes
GameAgentDevCli node list --world <world-id>

# View node details (includes components, memories, relations)
GameAgentDevCli node get <node-id>

# Update a node
GameAgentDevCli node update <node-id> --name "New Name" --type "npc"

# Delete a node (leaf nodes only)
GameAgentDevCli node delete <node-id>
```

#### Autonomous behavior management

```bash
# View autonomous behavior config for a node
GameAgentDevCli node autonomous get <node-id>

# Configure autonomous behavior
GameAgentDevCli node autonomous set <node-id> --enabled --trigger "world_tick_sync"

# Disable autonomous behavior
GameAgentDevCli node autonomous disable <node-id>

# Manually trigger autonomous behavior once
GameAgentDevCli node autonomous run <node-id>
```

---

### world — World-level runtime operations

#### World management commands

```bash
# Clone a world
GameAgentDevCli world fork <world-id> [name] [--lock-world]

GameAgentDevCli world save <world-id> [name] [--lock-world]

GameAgentDevCli world restore <snapshot-world-id> [name] [--lock-world]

GameAgentDevCli world validate-snapshot <snapshot-world-id>

GameAgentDevCli world snapshot-info <snapshot-world-id>

GameAgentDevCli world list-snapshots <world-id>

GameAgentDevCli world delete-snapshot <snapshot-world-id>
```

- `--lock` / `-l`: Lock the source world during cloning to prevent concurrent writes (optional, defaults to unlocked)
- `name`: Specify a name for the new world; if omitted, it is auto-generated as "original_name (copy)"
- `validate-snapshot`: Check restore compatibility before attempting to restore a saved snapshot
- `snapshot-info`: Show metadata for a snapshot world
- `list-snapshots`: List all saved snapshots created from a source world
- `delete-snapshot`: Delete a saved snapshot world together with its persisted snapshot metadata

#### World settings

```bash
# View world runtime settings
GameAgentDevCli world settings get <world-id>

# Change only the flags you provide
GameAgentDevCli world settings set <world-id> \
  --pipeline-mode "polling" \
  --propagation-max-depth 0 \
  --sub-task-max-retries 0 \
  --sub-task-timeout-secs 0 \
  --enable-propagation-machine false
```

- `world settings set` performs a partial update: omitted flags keep their existing values.
- `propagation-max-depth 0` means no upward propagation depth limit.
- `sub-task-max-retries 0` disables automatic retries for sub-tasks.
- `sub-task-timeout-secs 0` disables the timeout guard for sub-tasks.

#### World policy

```bash
# View policy
GameAgentDevCli world policy get <world-id>

# Set blocked/safe actions
GameAgentDevCli world policy set <world-id> \
  --blocked "kill_character,nuclear_strike" \
  --safe "add_memory,send_dialogue"
```

#### World inference

```bash
# Advance world time (Tick)
GameAgentDevCli world tick <world-id> --type "scheduled" --time "Day 2 - Noon"

# Evaluate event impact
GameAgentDevCli world event-impact <world-id> \
  --type "diplomatic_crisis" \
  --scope <scope-id> \
  --description "Neighboring nation is amassing troops at the border..." \
  --severity "critical"

# Advance evolution within a specific scope
GameAgentDevCli world scope-advance <world-id> <scope-id>

# Regenerate world outline
GameAgentDevCli world replan <world-id>
```

#### Snapshots & Export

```bash
# Output world runtime snapshot
GameAgentDevCli world snapshot <world-id>

# Export world configuration (for backup or migration)
GameAgentDevCli world export <world-id> --format yaml --out myworld.yaml
```

---

### component — Component management

```bash
# List node components
GameAgentDevCli component list --node <node-id>

# Get a single component
GameAgentDevCli component get <component-id>

# Create a component (data should be a JSON string)
GameAgentDevCli component create --node <node-id> --type "profile" --data '{"name":"Elrin"}'

# Update a component
GameAgentDevCli component update <component-id> --data '{"name":"Speaker Elrin"}'

# Delete a component
GameAgentDevCli component delete <component-id>
```

---

### memory — Memory management

```bash
# List node memories
GameAgentDevCli memory list --node <node-id>

# Create a memory
GameAgentDevCli memory create --node <node-id> --content "..." --level "long_term" --tags "history"

# Get a memory
GameAgentDevCli memory get <memory-id>

# Update a memory
GameAgentDevCli memory update <memory-id> --content "..." --level "shared"

# Delete a memory
GameAgentDevCli memory delete <memory-id>
```

---

### relation — Relation management

```bash
# List relations
GameAgentDevCli relation list --world <world-id>

# Create a relation
GameAgentDevCli relation create --world <world-id> --source <node-id> --target <node-id> --type "ally" --weight 50

# Get a relation
GameAgentDevCli relation get <relation-id>

# Update a relation
GameAgentDevCli relation update <relation-id> --weight 80

# Delete a relation
GameAgentDevCli relation delete <relation-id>
```

---

### logs ? Inference logs

```bash
# Read recent inference logs (summary view by default)
GameAgentDevCli logs --world <world-id> --limit 10

# Output raw JSON for scripting
GameAgentDevCli logs --world <world-id> --limit 10 --json

# Filter logs by task type
GameAgentDevCli logs --world <world-id> --task-type world_tick
```

- The default output is a readable summary that includes world / node, pipeline mode, round usage, reply preview, and action / memory previews.
- `--json` prints the raw inference log array for scripting or deeper analysis.
- `--task-type` filters logs by task category.

### debug traces ? Debug traces

```bash
# Show recent debug traces (summary view by default)
GameAgentDevCli debug traces --world <world-id> --limit 10

# Output raw JSON
GameAgentDevCli debug traces --world <world-id> --limit 10 --json
```

- The summary view shows request id, pipeline mode, round usage, and errors for quick inspection.
- `--json` preserves the full debug payload for deeper troubleshooting.

---

### verify — Verification

```bash
# Verify imported content
GameAgentDevCli verify import demo-world.yaml
```

---

### inspect — Open Creator

```bash
GameAgentDevCli inspect
```

---

## Example: Complete Workflow

```bash
# 1. Start the engine (terminal 1)
GameAgentEngine serve

# 2. Import the Demo world (terminal 2)
GameAgentDevCli import tools/source/demo-world.yaml --reset

# 3. Check world status
GameAgentDevCli status

# 4. View runtime settings
GameAgentDevCli world settings get <world-id>

# 5. Advance world time
GameAgentDevCli world tick <world-id>

# 6. View inference logs
GameAgentDevCli logs --world <world-id>
```
