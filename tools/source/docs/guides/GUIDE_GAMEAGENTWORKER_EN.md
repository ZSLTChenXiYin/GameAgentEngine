# GameAgentWorker Guide

`GameAgentWorker` is the packaged canonical game-side worker CLI.

It is used for:

- external async task push / pull / callback validation
- YAML / JSON authority-state hosting
- `play` text-game style REPL
- packaged worker scenario tests

## Supported Commands

```bash
GameAgentWorker serve
GameAgentWorker push-receiver
GameAgentWorker pull-worker
GameAgentWorker pull-once
GameAgentWorker play
GameAgentWorker test <scenario>
```

Supported scenarios:

```bash
GameAgentWorker test base-data
GameAgentWorker test continuity
GameAgentWorker test runtime-tasks
GameAgentWorker test callback-resume
GameAgentWorker test tooling-smoke
GameAgentWorker test machine-scenario
GameAgentWorker test all
```

## Common Workflows

Full async loop:

```bash
GameAgentEngine serve
GameAgentWorker serve --verbose
```

Process one pull task:

```bash
GameAgentWorker pull-once --consumer game_client
```

Start play REPL:

```bash
GameAgentEngine serve
GameAgentDevCli import demo-world.yaml
GameAgentWorker play --state-file demo-state.yaml --world-id demo_world --player-node-id player_001
```

Run packaged tests:

```bash
GameAgentWorker test all
```

## `play` Mode Notes

`play` is not a raw Engine shell. It is a constrained text-game entrypoint.

Current support includes:

- `/talk <npc>` for private dialogue target selection
- plain text input for direct dialogue with the current target
- `/say <message>` for room-wide public speech
- `/ask <npc> <message>` for named NPC reply inside group-chat context
- `/act <text>` for natural-language player intent interpretation plus authority validation and execution
- `/gift <npc> <item>` for authority-state item transfer before NPC feedback
- `/show_item <npc> <item>` for item-presence validation before showing the item
- `/trade [npc]` and `/threaten [npc]`

During `play`, Engine can query high-frequency authoritative data through `game_client_request_data`, including:

- HP
- inventory
- money
- player / NPC location
- scene immediate state
- quest state
- item presence

Current group chat still uses one primary NPC responder per turn rather than parallel multi-NPC reasoning.
