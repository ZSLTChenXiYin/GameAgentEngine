# GameAgentWorker

`GameAgentWorker` is the canonical standalone worker CLI for GameAgentEngine.

It currently provides a deterministic fixture-driven runtime task worker that is suitable for:

- integration test orchestration
- callback / resume validation
- local developer-side engine interaction simulation

The legacy `GameAgentTestWorker` command remains available as a compatibility wrapper, but new scripts and docs should prefer `GameAgentWorker`.

## Supported Modes

- `serve`: run both the push receiver and the pull worker loop
- `push-receiver`: run only the HTTP push receiver
- `pull-worker`: run only the pull worker polling loop
- `pull-once`: process at most one pending pull task and exit

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
