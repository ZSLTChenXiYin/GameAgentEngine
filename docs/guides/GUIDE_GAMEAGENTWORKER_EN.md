# GameAgentWorker Guide

[**中文**](./GUIDE_GAMEAGENTWORKER.md) | **English**

`GameAgentWorker` is the canonical game-side worker CLI in this repository. It is no longer just a temporary test-script wrapper.

It serves two primary roles:

- close the Engine external async push / pull / callback loop during integration testing
- host YAML / JSON authoritative state and expose a text-game style `play` REPL during local development

---

## 1. Supported Commands

```bash
GameAgentWorker serve
GameAgentWorker push-receiver
GameAgentWorker pull-worker
GameAgentWorker pull-once
GameAgentWorker play
GameAgentWorker test <scenario>
```

What each command does:

- `serve`: run both the push receiver and the pull worker loop
- `push-receiver`: run only the HTTP push receiver on `/api/v1/runtime/dispatch`
- `pull-worker`: poll `/api/v1/runtime/tasks/pending` and process tasks
- `pull-once`: process at most one pending pull task and exit
- `play`: run a single-user text-game REPL backed by Engine and worker-side authoritative state
- `test <scenario>`: run packaged worker scenarios

Currently packaged scenarios:

```bash
GameAgentWorker test base-data
GameAgentWorker test continuity
GameAgentWorker test runtime-tasks
GameAgentWorker test callback-resume
GameAgentWorker test tooling-smoke
GameAgentWorker test machine-scenario
GameAgentWorker test all
```

---

## 2. Worker Positioning

Inside the current repository, `GameAgentWorker` is the standard game-side execution surface:

- for Engine, it is the runtime consumer of external interfaces
- for integration tests, it is the deterministic fixture worker
- for developers, it is the local game-side shell
- for `play`, it is the authority-state host rather than a raw invoke wrapper

This is the tool that should absorb future game-side async interface simulation, authority-state queries, play-mode workflows, and packaged test flows.

---

## 3. Default Ports and Tokens

Current implementation defaults:

- Engine base URL: `http://127.0.0.1:8080`
- push receiver port: `9000`
- runtime task token: `dev-task-token`
- callback token: `dev-callback-token`
- push bearer token: `local-test-token`
- default consumer: `game_client`
- default lease owner: `gameagentworker`

Common flags:

```bash
--engine-base-url
--engine-api-key
--runtime-task-token
--callback-token
--game-http-bearer-token
--state-file
--consumer
--lease-owner
--push-port
--poll-interval
--heartbeat-interval
--callback-delay
--long-task-duration
--fail-interface
--long-task-interface
--verbose
```

---

## 4. Integration-Test and Runtime Task Workflow

### 4.1 `serve`

Use this for full local loop validation:

```bash
GameAgentWorker serve --verbose
```

It will:

- receive Engine push dispatch
- poll pull-mode runtime tasks
- build deterministic fixture results by interface name
- send heartbeats when simulating long tasks
- callback results to `/api/v1/actions/callback`

### 4.2 `push-receiver`

Use this when you only want the push path:

```bash
GameAgentWorker push-receiver --push-port 9000
```

It currently accepts:

- `POST /api/v1/runtime/dispatch`
- `Authorization: Bearer <game-http-bearer-token>`

### 4.3 `pull-worker`

Use this when you only want the pull consumer:

```bash
GameAgentWorker pull-worker --consumer game_client
```

Its loop is:

1. fetch pending runtime task
2. claim task
3. start task
4. decide success / failure / long-running simulation by interface name
5. callback result

### 4.4 `pull-once`

Use this for scripting and single-step debugging:

```bash
GameAgentWorker pull-once --consumer game_client
```

If there is no pending task, it emits one `pull_noop` log event and exits.

---

## 5. Failure and Long-Task Simulation

Force one interface to fail:

```bash
GameAgentWorker serve --fail-interface spawn_item
```

Simulate one interface as long-running:

```bash
GameAgentWorker serve \
  --long-task-interface game_client_request_data \
  --long-task-duration 8s \
  --heartbeat-interval 2s
```

This is useful for validating:

- heartbeat renewal
- callback-resume stability
- Engine timeout and long-task handling

---

## 6. What `play` Really Is

`play` is not a thin raw-Engine shell. It is a constrained text-game entrypoint.

Current implementation already enforces these boundaries:

- high-frequency truth stays in worker-side authoritative state
- NPC responses are still Engine-driven
- Engine can fetch authoritative facts through `game_client_request_data` when needed
- player natural-language input can be interpreted into player intent, then validated and executed against authority state

So the current `play` mode is already a combined flow of game-side truth plus Engine reasoning, not just “send one text line to the model”.

---

## 7. Starting `play`

Start Engine and import a world first:

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/workerhome/demo/demo-world.yaml
```

Then run `play`:

```bash
GameAgentWorker play --state-file tools/source/workerhome/demo/demo-state.yaml --world-id demo_world --player-node-id player_001
```

Relevant flags:

```bash
--state-file
--world-id
--player-node-id
--session-id
--pipeline-mode
--include-related-nodes
--auto-worker
```

Notes:

- `--state-file` is required and loads YAML / JSON authoritative state
- `--world-id` can be inherited from the state file `world_id`
- `--player-node-id` falls back to the actor marked `kind=player`
- `--auto-worker` is enabled by default and runs an embedded pull worker for authority callbacks

If `--auto-worker` is disabled and a turn pauses on `game_client_request_data`, `play` will stop and tell you the authority callback was not resumed.

---

## 8. `play` Commands

Current commands:

```text
/+help
/+look
/+who
/+state
/+inventory
/+quests
/+talk <npc>
/+next_target
/+prev_target
/+target
/+room
/+move <scene>
/+inspect [target]
/+use_item <item>
/+clear_target
/+say <message>
/+ask <npc> <message>
/+act <text>
/+gift <npc> <item>
/+show_item <npc> <item>
/+trade [npc]
/+threaten [npc]
/+exit

# legacy aliases remain supported
/help
/look
...
```

What they do:

| Command | Purpose |
| --- | --- |
| `/+help` | show help |
| `/+look` | show current scene summary, prompt, and occupants |
| `/+who` | list actors in the current scene |
| `/+state` | show authoritative player summary such as HP, money, inventory, and location |
| `/+inventory` | show detailed inventory |
| `/+quests` | show quest and story-state summary |
| `/+talk <npc>` | set the current private dialogue target |
| `/+next_target` / `/+prev_target` | cycle between dialogue-capable NPCs in the current scene |
| `/+target` | show the current dialogue target |
| `/+room` | show room participants and the current group-chat primary responder |
| `/+move <scene>` | execute deterministic player movement against authoritative state |
| `/+inspect [target]` | inspect the current scene, one actor, or one visible item |
| `/+use_item <item>` | execute deterministic item-use validation for an owned item |
| `/+clear_target` | clear the current dialogue target |
| `/+say <message>` | speak publicly in the room; the current group-chat primary responder answers |
| `/+ask <npc> <message>` | ask one named NPC to answer inside group-chat context |
| `/+act <text>` | map natural language into player intent, then validate, execute, and bridge into NPC or group-chat response |
| `/+gift <npc> <item>` | commit the gift in authoritative state first, then ask the NPC to react |
| `/+show_item <npc> <item>` | verify the player actually has the item, then show it to the NPC |
| `/+trade [npc]` | start trade / bargaining dialogue |
| `/+threaten [npc]` | start threat-based dialogue |
| `/+exit` | leave play mode |

Also note:

- legacy `/help` and `/talk` style aliases still work.
- If you type plain text directly, it is sent to the current `/+talk` target as `direct_dialogue`.

---

## 9. Why `/act` Matters

`/act` is the closest current implementation to natural-language player control.

Its flow is not “text equals truth”. It is:

1. call the Engine player-input interpretation endpoint
2. produce a structured player intent
3. validate that intent against worker authority state
4. execute only validated changes in game-side state
5. surface the interpreted intent, missing facts, and suggested follow-up interaction inside play
6. bridge the result into NPC dialogue or group-chat response

This is the correct base for future natural-language player control because it preserves:

- the player as a first-class node inside Engine world modeling
- the game side as final authority over fast-changing state
- the Engine as the interpreter and narrative response layer

---

## 10. Current Authority Query Types

`play` exposes these dynamic query types through `game_client_request_data`:

- `player_state`
- `player_inventory`
- `player_wallet`
- `player_location`
- `npc_location`
- `scene_state`
- `room_state`
- `task_state`
- `item_presence`

These cover the main high-frequency authoritative facts discussed so far:

- HP
- inventory
- money
- player / NPC location
- room / scene immediate state
- quest state
- whether the player actually has a specific item

---

## 11. Current Implementation Boundaries

`play` already supports group-chat entrypoints, but still has explicit limits:

- group chat still selects one primary responder per turn; it is not multi-NPC parallel cluster reasoning
- the primary group-chat responder is now explicit: current target first, otherwise a stable scene-default NPC
- `action_calls` are displayed but not automatically committed locally
- high-risk natural-language actions must still pass authority validation
- `/act` now prints interpreted intent, steps, missing facts, and suggested interaction details for debugging

So the current version is suitable for:

- experiencing Engine-driven text-game interaction
- validating player-intent interpretation plus authority validation
- validating NPC dialogue, room chat, gifting, and item-showing loops

It is not yet a full multi-character parallel world simulator.

---

## 12. Recommended Workflows

### Shortest play path

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/workerhome/demo/demo-world.yaml
GameAgentWorker play --state-file tools/source/workerhome/demo/demo-state.yaml --world-id demo_world --player-node-id player_001
```

### Shortest async task loop path

```bash
GameAgentEngine serve
GameAgentWorker serve --verbose
```

### Shortest packaged test path

```bash
GameAgentWorker test all
```

---

## 13. Responsibility Split with Other Tools

| Tool | Current responsibility |
| --- | --- |
| `GameAgentEngine` | world modeling, NPC reasoning, memory, relations, time progression, external async task orchestration |
| `GameAgentDevCli` | config import, world management, runtime debugging, task and timeline diagnostics, open Creator |
| `GameAgentWorker` | game-side worker, authority-state host, async interface loop, play REPL, packaged integration tests |
| `GameAgentCreator` | browser-based visual editor and observability UI |

This is also the recommended long-term responsibility boundary.
