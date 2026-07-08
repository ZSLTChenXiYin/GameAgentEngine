# Engine Tooling Sync Report

## Scope

This report records the completed synchronization work that brought `DevCli`, `Creator`, and `SDK` up to the current `Engine` baseline for observability, world tick continuity, and persistent state inspection.

## Baseline Used

`Engine` remained the semantic source of truth throughout this sync effort.

The baseline capability set came from two already-completed engine tracks:

1. Execution observability across `debug`, `review`, and `production`
2. `world_tick` continuity persistence through `logs`, `timelines`, and structured state components

## Delivered Across 8 Phases

| Phase | Delivery |
|---|---|
| 1 | Synced Creator component metadata with Engine component registry and established the tooling sync baseline |
| 2 | Added unified HTTP APIs for `state-components` and `timelines` |
| 3 | Exposed these contracts through the Go SDK with typed responses and tests |
| 4 | Added DevCli commands for `state`, `timeline`, and enhanced `logs` filtering/detail output |
| 5 | Added Creator read-only pages for `State`, `Timelines`, and richer log details |
| 6 | Added Creator editing for continuity state components, while keeping `state_snapshot` read-only |
| 7 | Updated SDK / DevCli / Creator documentation and added regression coverage |
| 8 | Published this final sync report and upgrade summary |

## Current Capability Matrix

| Capability | Engine | DevCli | Creator | SDK |
|---|---|---|---|---|
| Execution-mode logging policy | Done | Done | Done | Done |
| Structured `logs` access | Done | Done | Done | Done |
| Timeline history archive | Done | Done | Done | Done |
| `world_state` access | Done | Done | Done | Done |
| `story_state` access | Done | Done | Done | Done |
| `story_history` access | Done | Done | Done | Done |
| `tick_policy` access | Done | Done | Done | Done |
| `state_snapshot` access | Done | Done | Done | Done |
| World tick continuity inspection | Done | Done | Done | Done |

## New Public API Surface

Added endpoints:

- `GET /api/v1/worlds/{world_id}/state-components`
- `GET /api/v1/worlds/{world_id}/state-components/{component_type}`
- `PUT /api/v1/worlds/{world_id}/state-components/{component_type}`
- `GET /api/v1/worlds/{world_id}/timelines`
- `GET /api/v1/worlds/{world_id}/timelines/latest`

These endpoints are intentionally world-oriented rather than generic component queries, because the continuity state is logically owned by the world tick lifecycle, not by arbitrary node browsing.

## SDK Additions

Added typed SDK support for:

- `StateComponentsResponse`
- `StateComponentResponse`
- `TimelineTick`
- `TimelineEnvelope`
- `TimelinesResponse`
- `LatestTimelineResponse`

Added client methods:

- `GetStateComponents`
- `GetStateComponent`
- `PutStateComponent`
- `GetTimelines`
- `GetLatestTimeline`

## DevCli Additions

Added command groups:

- `GameAgentDevCli state list|get|set`
- `GameAgentDevCli timeline list|latest`

Enhanced:

- `GameAgentDevCli logs` now supports `--category`, `--event`, `--mode`, `--details`

## Creator Additions

Added pages:

- `State`
- `Timelines`

Enhanced:

- `Logs` now exposes `detail_data` and mode/category/event context more clearly
- continuity state components are editable directly in Creator, except `state_snapshot`

## Non-Destructive Upgrade Notes

This sync did not introduce destructive schema changes.

What changed instead:

- new world-level API routes were added
- SDK types and methods were expanded
- DevCli added new command groups
- Creator gained new read/write continuity views

No existing route was removed, and no legacy CLI or SDK method was intentionally broken.

## Recommended Usage Order For Debugging Continuity

1. Check `Timelines` for the latest tick archive.
2. Check `world_state` and `story_history` for retained facts.
3. Check `logs` and `debug traces` for prompt / response / detail data.
4. Adjust `tick_policy` or continuity state as needed.
5. Run the next tick and compare again.

See also:

- `docs/CONTINUITY_DEBUG_WORKFLOW.md`
- `docs/engine-observability-and-worldtick.md`

## Regression Status

Final regression completed successfully with:

```bash
go test ./...
```

## Follow-Up Enhancements

The next enhancement wave added deeper continuity tooling on top of the original sync baseline:

- server-side structured log filtering by `world_id`, `node_id`, `task_type`, `category`, `event_name`, `execution_mode`, `request_id`, and `round`
- SDK continuity bundle loading through `GetContinuityBundle`
- DevCli continuity diagnosis through `GameAgentDevCli debug continuity`
- Creator continuity aggregation and continuity diff views
- stronger structured validation for `world_state`, `story_state`, `story_history`, and `tick_policy`
