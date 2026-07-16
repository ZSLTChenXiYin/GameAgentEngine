# GameAgentDevCli Guide

[**中文**](./GUIDE_GAMEAGENTDEVCLI.md) | **English**

GameAgentDevCli is the command-line tool for operating GameAgentEngine through the HTTP API.

---

## Current Scope

- node / component / memory / relation CRUD
- world import, export, snapshot, restore, and validation
- world settings, world policy, and plan approval
- world tick, event impact, scope advance, and timeline replan
- continuity state components and timeline archive access
- logs, traces, continuity debugging, and node graph debugging
- opening Creator
- task management (`task` commands)
- verify and action-callback helper entrypoints

---

## Start From Scratch

The current recommended first step is to create a `world` root node:

```bash
GameAgentDevCli node create --type world --name "New World"
```

Then add child nodes:

```bash
GameAgentDevCli node create --world <world-id> --type location --name "Starter Village"
GameAgentDevCli node create --world <world-id> --type npc --name "Gatekeeper"
```

---

## World Settings

```bash
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world settings set <world-id> --pipeline-mode polling
GameAgentDevCli world settings set <world-id> --world-time-settings-file world-time.json
GameAgentDevCli world settings set <world-id> --world-time-settings-json '{"tick_scale_mode":"flexible","tick_min_unit":"hour","tick_step":1,"tick_units":["day","hour"]}'
```

When `world_time_settings` is missing, world-time-dependent flows are intentionally blocked in the current design.

---

## Run a Tick

```bash
GameAgentDevCli world tick <world-id>
GameAgentDevCli world tick <world-id> --type manual --time "day-1" --requested-ticks 1
GameAgentDevCli world tick <world-id> --autonomous-limit 2
```

If `tick_scale_mode` is `fixed`, `requested_ticks` must remain 1.

---

## Continuity and Timelines

```bash
GameAgentDevCli state list <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli timeline list <world-id> --limit 5
GameAgentDevCli debug continuity <world-id>
```

When debugging paused multi-round execution or task resumption, check these in order:

1. `GameAgentDevCli task inspect <task-id>`
2. `GameAgentDevCli debug continuity <world-id>`
3. `GameAgentDevCli debug traces --world <world-id> --limit 10`

---

## Invoke Requests

`invoke` is the lowest-friction way to send one reasoning request without writing a custom client.

```bash
GameAgentDevCli invoke <world-id> <node-id> \
  --task-type npc_dialogue \
  --message "What happened at the south gate?"
```

Request-scoped pipeline mode and dynamic interfaces can be passed directly:

```bash
GameAgentDevCli invoke <world-id> <node-id> \
  --task-type npc_dialogue \
  --message "Check the nearby scene before answering." \
  --pipeline-mode full \
  --dynamic-interfaces-file dynamic-interfaces.json
```

`dynamic-interfaces.json` should contain a JSON array. Use it for request-local interface whitelists such as temporary NPC dialogue query tools.

---

## Task Management

Runtime Tasks manage external interactions between the Engine and the game client. Three delivery modes are supported: Push, Pull, and Hybrid. The commands below target Pull mode workflows:

```bash
GameAgentDevCli task list --status pending --limit 20
GameAgentDevCli task list --pending-only --consumer game_client --limit 20
GameAgentDevCli task get <task-id>
GameAgentDevCli task get <task-id> --json
GameAgentDevCli task inspect <task-id>
GameAgentDevCli task stats
GameAgentDevCli task claim <task-id> --consumer gamer --owner devcli
GameAgentDevCli task start <task-id> <lease-token>
GameAgentDevCli task heartbeat <task-id> <lease-token>
GameAgentDevCli task release <task-id> <lease-token> --reason "completed"
GameAgentDevCli task requeue <task-id> --retry-delay-ms 1500 --reason "manual requeue"
```

Use `task inspect` when you need callback and resume details in one view. It prints `callback_id`, `resume_execution_id`, compact payload preview, dispatch decision fields, and timing fields.

Use `task stats` when diagnosing queue health. It summarizes total tasks, ready pull tasks, in-flight tasks, heartbeat timeouts, retry exhaustion, stale dispatched tasks, and aggregated counters.

---

## Minimal Pull Worker Loop

If your game side consumes pull tasks, the smallest stable loop is:

1. `task list --pending-only --consumer <consumer>` or call `/api/v1/runtime/tasks/pending`
2. `task claim`
3. `task start`
4. execute the game-side query or action
5. post callback result
6. keep heartbeat if execution is long-running

The callback response now includes post-process metadata from the Engine, so SDK or custom worker code can tell whether the callback only completed, resumed execution, or triggered memory write post-processing.

---

## Basic Operational Checks

When you only need to verify service reachability or version alignment, start with:

```bash
GameAgentDevCli status
GameAgentDevCli version
```

These are better first checks than jumping directly into world or task commands when you are not yet sure the service endpoint and API key are correct.

---

## Import, Export, and Verification

World-asset lifecycle commands are now expected to live in DevCli:

```bash
GameAgentDevCli import tools/source/demo-world.yaml
GameAgentDevCli world export <world-id> --format yaml --out exported-world.yaml
GameAgentDevCli world snapshot <world-id> --out runtime-snapshot.json
GameAgentDevCli world save <world-id> demo-save
GameAgentDevCli world restore <snapshot-world-id> restored-world
GameAgentDevCli world validate-snapshot <snapshot-world-id>
GameAgentDevCli verify import tools/source/demo-world.yaml
GameAgentDevCli verify demo
```

The intended boundary is:

- Engine remains the runtime kernel
- DevCli owns import/export/snapshot/verification workflows
- Worker owns game-side async loops and play-mode REPL

---

## Opening Creator

```bash
GameAgentDevCli creator
```

`creator` is now the canonical DevCli browser entrypoint. Older naming ideas around “inspect/open editor” are no longer the main workflow surface.
