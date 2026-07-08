# Engine Tooling Sync Plan

## Scope

This plan keeps `Engine`, `DevCli`, `Creator`, and `SDK` aligned by treating `Engine` as the single source of truth for runtime behavior, state persistence, and public data contracts.

## Current Baseline

`Engine` is already ahead in two important areas:

1. Execution observability across `debug`, `review`, and `production`.
2. `world_tick` continuity persistence through `logs`, `timelines`, and structured state components.

`DevCli`, `Creator`, and `SDK` already expose basic `logs` access, but they do not yet offer a unified surface for the continuity state carried by these component types:

- `world_state`
- `story_state`
- `story_history`
- `tick_policy`
- `state_snapshot`

## Sync Principles

1. `Engine` defines behavior and schema.
2. `SDK` mirrors the public API contract exactly.
3. `DevCli` is the first developer-facing debugging surface.
4. `Creator` is the final visual surface for inspection and editing.
5. No tool should invent fields or semantics that do not exist in `Engine`.

## Eight Phases

| Phase | Goal | Expected Outcome |
|---|---|---|
| 1 | Baseline sync | Sync docs and component metadata with current Engine capability |
| 2 | API surface | Add unified state component and timeline endpoints |
| 3 | SDK surface | Add state/timeline types and client methods |
| 4 | DevCli surface | Add state/timeline/debugging commands |
| 5 | Creator read-only surface | Add continuity inspection views |
| 6 | Creator editing surface | Add state component editing and validation |
| 7 | Docs and regression | Align docs, examples, and automated coverage |
| 8 | Wrap-up | Publish sync report and upgrade notes |

## Capability Matrix

| Capability | Engine | DevCli | Creator | SDK |
|---|---|---|---|---|
| Execution-mode logging policy | Done | Partial | Partial | Partial |
| Structured `logs` access | Done | Done | Done | Done |
| Timeline history archive | Done | Missing | Missing | Missing |
| `world_state` access | Done | Missing | Missing | Missing |
| `story_state` access | Done | Missing | Missing | Missing |
| `story_history` access | Done | Missing | Missing | Missing |
| `tick_policy` access | Done | Missing | Missing | Missing |
| `state_snapshot` access | Done | Missing | Missing | Missing |
| World tick continuity inspection | Done | Missing | Missing | Missing |

## Acceptance Standard

Each phase is only considered complete when all of the following are true:

1. `Engine` remains the semantic source.
2. Public API fields are stable and test-covered.
3. `SDK` can consume the capability without ad hoc JSON handling.
4. `DevCli` can inspect or operate the capability from the terminal.
5. `Creator` can at least view the capability once the UI phase is reached.
