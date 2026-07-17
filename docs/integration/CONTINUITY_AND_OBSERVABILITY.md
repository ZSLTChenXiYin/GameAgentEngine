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

This page is now the canonical continuity workflow entrypoint.

Keep the following practical checks in one place when diagnosing continuity regressions:

- use `timeline latest`, `timeline list`, `state list`, `logs`, and `debug traces` first
- treat `state_snapshot` as an Engine-generated read-only checkpoint payload
- patch `tick_policy`, `world_state`, `story_state`, or `story_history` only when you are intentionally repairing continuity state
- re-run one tick after each targeted change instead of stacking multiple blind edits first

## 7. Minimal Regression Sample

When you want a quick regression sample, seed one stable canonical fact into both `world_state` and `story_history`, keep the same fact protected by `tick_policy`, then advance one tick and verify that:

1. the fact still exists in `world_state.canonical_facts`
2. the newest `story_history` entry still preserves the same fact rather than degrading into a vaguer paraphrase
3. logs, traces, and timelines still line up on the same request path

One practical example is a fixed underground facility fact such as:

- `地下52米量子谐振腔`

Treat the run as regressed if the fact disappears, loses its depth/status detail, or is no longer protected by the latest `tick_policy` path.
