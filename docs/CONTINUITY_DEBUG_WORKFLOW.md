# World Tick Continuity Debug Workflow

This workflow is the recommended path for inspecting and adjusting `world_tick` continuity after the Engine observability refactor.

## 1. Inspect the latest timelines

Use DevCli:

```bash
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli timeline list <world-id> --limit 5
GameAgentDevCli debug continuity <world-id>
```

Use Creator:

- Open the `Continuity` page first for the aggregated world-tick bundle
- Use `Continuity Diff` to compare the latest tick against the previous one
- Open the `Timelines` page
- Inspect the latest tick summary
- Expand `Timeline Payload` to compare `reply`, `future_outline`, `memory_updates`, and `action_calls`

## 2. Inspect continuity state components

Use DevCli:

```bash
GameAgentDevCli state list <world-id>
GameAgentDevCli state get <world-id> world_state
GameAgentDevCli state get <world-id> story_state
GameAgentDevCli state get <world-id> story_history
GameAgentDevCli state get <world-id> tick_policy
```

Use Creator:

- Open the `State` page
- Compare `world_state`, `story_state`, `story_history`, and `tick_policy`
- Treat `state_snapshot` as a read-only checkpoint payload

## 3. Inspect pipeline and execution logs

Use DevCli:

```bash
GameAgentDevCli logs --world <world-id> --task-type world_tick --mode debug --details
GameAgentDevCli logs --world <world-id> --category pipeline --event raw_llm_response_received --details
GameAgentDevCli logs --world <world-id> --request-id <request-id> --round 1 --details
GameAgentDevCli debug traces --world <world-id> --limit 10
```

Use Creator:

- Use the `Continuity` page request filter to isolate one `request_id`
- Open the `Logs` page for structured request/response/detail payloads
- Open the `Traces` page for debug-mode prompt and parsed-output inspection

## 4. Apply targeted continuity fixes

Use DevCli to patch continuity state directly:

```bash
GameAgentDevCli state set <world-id> tick_policy --data '{"continuity_rules":["Do not discard established underground reactor facts."]}'
```

Use Creator to:

- Edit `world_state`
- Edit `story_state`
- Edit `story_history`
- Edit `tick_policy`

Avoid editing `state_snapshot` unless you are intentionally reconstructing an engine-generated checkpoint.

## 5. Re-run and compare

After editing state:

1. Advance one tick.
2. Re-open the latest `Timelines` entry.
3. Re-check `world_state.canonical_facts` and `story_history.entries`.
4. Re-check `logs` or `debug traces` if the model still drops context.
