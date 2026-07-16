# GameAgentWorker

`GameAgentWorker` is the canonical standalone worker CLI for GameAgentEngine.

It currently provides a deterministic fixture-driven runtime task worker that is suitable for:

- integration test orchestration
- callback / resume validation
- local developer-side engine interaction simulation

Packaged integration-test fixtures and helper scripts now live under `tools/source/tests`.

## Supported Modes

- `serve`: run both the push receiver and the pull worker loop
- `push-receiver`: run only the HTTP push receiver
- `pull-worker`: run only the pull worker polling loop
- `pull-once`: process at most one pending pull task and exit
- `play`: run a single-user text-game REPL backed by Engine invoke and worker-side authority state
- `test <scenario>`: run packaged worker-side full-functional scenarios; current subcommands are `base-data`, `continuity`, `runtime-tasks`, `callback-resume`, `tooling-smoke`, `machine-scenario`, and `all`

## Capabilities

- receive push dispatch requests on `/api/v1/runtime/dispatch`
- poll `/api/v1/runtime/tasks/pending`
- claim and start tasks
- send task heartbeats for long-running task simulation
- post callback results to `/api/v1/actions/callback`
- emit structured logs for task, callback, and decision correlation
- return deterministic fixture payloads by interface name

## Default Ports and Tokens

- push receiver port: `9000`
- Engine base URL: `http://127.0.0.1:8080`
- runtime task token: `dev-task-token`
- callback token: `dev-callback-token`
- expected push bearer token: `local-test-token`

## Example Usage

Run both push and pull behaviors:

```powershell
GameAgentWorker.exe serve --verbose
```

Run only the push receiver:

```powershell
GameAgentWorker.exe push-receiver --push-port 9000
```

Run one pull step:

```powershell
GameAgentWorker.exe pull-once --consumer game_client
```

Run play mode with a local authority state file:

```powershell
GameAgentWorker.exe play --state-file .\tools\source\demo-state.yaml --player-node-id player_001 --world-id demo_world
```

## Play Mode

`play` is the first step toward a real text-game shell instead of a raw engine wrapper.

Current behavior:

- loads authoritative game-side state from a YAML/JSON `--state-file`
- selects one player actor via `--player-node-id` or `kind=player`
- uses `/talk <npc>` to lock the current dialogue target
- sends plain text input to Engine as `npc_dialogue` with `interaction.mode=direct_dialogue`
- exposes request-scoped `game_client_request_data` so the NPC can query authoritative state during dialogue
- can run an embedded pull worker with `--auto-worker` enabled so play-mode invoke calls can resolve authority callbacks automatically

Current commands:

- `/help`
- `/look`
- `/who`
- `/room`
- `/state`
- `/talk <npc>`
- `/say <message>`
- `/ask <npc> <message>`
- `/target`
- `/clear_target`
- `/gift <npc> <item>`
- `/show_item <npc> <item>`
- `/trade [npc]`
- `/threaten [npc]`
- `/exit`

Notes:

- `play` now supports direct dialogue plus a first group-chat pass.
- group chat still uses a single primary responder per turn; it does not run multi-NPC parallel reasoning.
- the state file remains the authority source for high-frequency facts like HP, money, inventory, and scene occupancy.
- the repository ships a matching pair of demo assets in `tools/source/demo-world.yaml` and `tools/source/demo-state.yaml`.

## Failure and Long-Task Simulation

Force callback failure for one interface:

```powershell
GameAgentWorker.exe serve --fail-interface spawn_item
```

Simulate long-running execution with heartbeats:

```powershell
GameAgentWorker.exe serve --long-task-interface game_client_request_data --long-task-duration 8s --heartbeat-interval 2s
```

## Recommended Test Usage

Use this worker in the full-functional plan for:

- runtime task push delivery checks
- runtime task pull consumer checks
- hybrid fallback checks
- callback and paused execution resume validation
- callback post-process validation

Current migration status:

- `test base-data`, `test continuity`, `test runtime-tasks`, `test callback-resume`, and `test tooling-smoke` are implemented in the worker CLI.
- Remaining test scenarios are being migrated from legacy scripts into `gameagentworker test ...`.
