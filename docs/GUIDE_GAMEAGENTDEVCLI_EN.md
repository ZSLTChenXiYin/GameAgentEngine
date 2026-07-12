# GameAgentDevCli Guide

[**中文**](./GUIDE_GAMEAGENTDEVCLI.md) | **English**

GameAgentDevCli is the command-line tool for operating GameAgentEngine through the HTTP API.

---

## Current Scope

- node / component / memory / relation CRUD
- world settings, world policy, and plan approval
- world tick, event impact, scope advance, and timeline replan
- continuity state components and timeline archive access
- logs, traces, continuity debugging, and node graph debugging
- snapshot save, validation, restore, and deletion
- opening Creator
- task management (`task` commands)

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
---

## Task Management

Runtime Tasks manage external interactions between the Engine and the game client. Three delivery modes are supported: Push, Pull, and Hybrid. The commands below target Pull mode workflows:

```bash
GameAgentDevCli task list --status pending --limit 20
GameAgentDevCli task get <task-id>
GameAgentDevCli task claim <task-id> --consumer gamer --owner devcli
GameAgentDevCli task start <task-id> <lease-token>
GameAgentDevCli task heartbeat <task-id> <lease-token>
GameAgentDevCli task release <task-id> <lease-token> --reason "completed"
```
