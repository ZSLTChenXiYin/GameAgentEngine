# Continuity and Observability

This page is the main entrypoint for `world_tick` continuity inspection and observability.

## 1. What to Inspect

Current continuity inspection revolves around four artifacts:

- timelines
- continuity state components
- structured logs
- debug traces

## 2. Recommended Workflow

Use this order when debugging continuity issues:

1. Inspect the latest timelines
2. Inspect `world_state`, `story_state`, `story_history`, and `tick_policy`
3. Inspect logs and traces for one `request_id`
4. Apply targeted continuity fixes
5. Re-run the next tick and compare again

## 3. Tooling Entry Points

Use DevCli:

```bash
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli timeline list <world-id> --limit 5
GameAgentDevCli state list <world-id>
GameAgentDevCli debug continuity <world-id>
GameAgentDevCli logs --world <world-id> --details
GameAgentDevCli debug traces --world <world-id> --limit 10
```

Use Creator:

- `Continuity`
- `State`
- `Timelines`
- `Logs`
- `Traces`

## 4. Current Persistence Model

Current `world_tick` persistence is intentionally split across:

- `logs` for runtime observability
- `timelines` for ordered per-tick archives
- continuity state components for inherited structured carry-over

## 5. Common Regression Check

Treat continuity as regressed when:

- known canonical facts disappear from `world_state`
- the latest history entry loses important retained facts
- the model stops honoring `tick_policy` continuity rules
- linked logs and traces no longer line up on the same request path

## 6. Historical and Supplemental Material

Detailed historical documents now live under `docs/internal/`:

- [Continuity Debug Workflow](../internal/CONTINUITY_DEBUG_WORKFLOW.md)
- [Continuity Regression Sample](../internal/CONTINUITY_REGRESSION_SAMPLE.md)
- [Engine Tooling Sync Report](../internal/ENGINE_TOOLING_SYNC_REPORT.md)
- [Engine Observability And World Tick Refactor](../internal/engine-observability-and-worldtick.md)
