# External Interaction Overview

This page is the primary documentation entrypoint for external asynchronous interaction in the current project.

## 1. Capability Boundary

Engine currently supports three external task delivery modes:

- `push`
- `pull`
- `hybrid`

External interaction is organized around two interface families:

- `external_query`: query authoritative data from the game side or another external system, then continue inference
- `external_action`: request that an external system execute an action, optionally record only the result, or feed it back to trigger resume or post-processing

## 2. Current Recommended Responsibility Split

- Engine: runtime task creation, delivery governance, callback write-back, resume orchestration, and unified post-processing
- Worker / game side: actual task execution, maintenance of high-frequency authoritative state, and callback of execution results
- DevCli / Creator: task diagnostics, callback/resume inspection, and troubleshooting
- SDK: programmatic integration against these HTTP contracts

## 3. Recommended Integration Modes

| Scenario | Recommended Mode | Notes |
| --- | --- | --- |
| Engine and the game logic service are on the same controlled network | `push` | shortest path and lowest latency |
| the game client should not expose an inbound service | `pull` | the client or bridge actively claims work |
| push should be preferred but fallback is required | `hybrid` | fallback to pull queue when push fails |

## 4. Minimal Closed Loop

The minimal external-interaction loop is:

1. the caller sends a request through `invoke`
2. Engine creates a runtime task
3. Worker or the game side receives the task through `push` / `pull`
4. the external system completes the query or action
5. the external system calls `POST /api/v1/actions/callback`
6. Engine updates task state and resumes the original execution path when needed

## 5. Key Capabilities Already Landed

- runtime task queue model
- `pending / claim / start / heartbeat / release / requeue / stats`
- callback completion write-back
- paused execution callback auto-resume
- request-scoped `dynamic_interfaces`
- baseline callback post-process (`none` / `record_only` / `write_memory`)
- Worker push / pull / callback closed loop and built-in test scenarios

## 6. Recommended Debugging Order

When external interaction does not behave as expected, inspect in this order:

1. `GameAgentDevCli task inspect <task-id>`
2. `GameAgentDevCli task stats`
3. `GameAgentDevCli debug continuity <world-id>`
4. Creator `Tasks` page
5. Creator `Continuity` / `Logs` / `Traces` pages

## 7. Supplemental Material

This page is now the canonical external-interaction workflow entrypoint.

Keep implementation-only notes under `docs/internal/` only when they still define a live contract or an unfinished future roadmap.

## 8. Recommended Integration Patterns

Treat the current external-interaction baseline as three stable patterns:

- push: Engine dispatches through a configured adapter and the game side completes through callback
- pull: the game side or a bridge claims runtime tasks, executes them, then reports completion through callback
- hybrid: Engine prefers push first, then falls back into pull-style queue consumption when dispatch fails

Important current boundaries:

- `fallback_transport` currently means falling back into pull-style queue consumption, not switching automatically to another push adapter
- `max_attempts` constrains pull / hybrid claim-retry behavior rather than acting as a full dead-letter subsystem
- callback post-process behavior such as `record_only` or `write_memory` should be treated as task-snapshot behavior

## 9. Current Automated Coverage Boundary

The current automated external-interaction baseline covers:

- push dispatch state transition and observability fields
- pull queue claim / start / heartbeat / release / requeue paths
- hybrid push-failure fallback into released pull tasks
- callback completion, paused-execution auto-resume, and `resume_policy = none` behavior
- heartbeat-timeout marking, auto-requeue snapshot policy, retry exhaustion, and repeated-timeout diagnostics
- request-id-based callback replay protection

Still future enhancement areas:

- finer governance policies by `consumer` or `category`
- richer multi-stage hybrid fallback state machines
- stronger callback replay protection beyond request-id occupation
- batch operator intervention flows on top of diagnostic views
