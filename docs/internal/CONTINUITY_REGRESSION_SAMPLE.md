# Continuity Regression Sample

Use this sample when you want to verify that `world_tick` does not drop previously established facts during the next tick.

## Target Fact

- `地下52米量子谐振腔`

## Baseline State Patch

```bash
GameAgentDevCli state set <world-id> world_state --data '{
  "summary": "A non-human reactor chamber has been confirmed below the city.",
  "canonical_facts": ["地下52米量子谐振腔", "The chamber remains active below the lower vault."],
  "active_arcs": ["Council debates whether to seal the chamber."],
  "metadata": {"scenario": "continuity_regression_sample"}
}'

GameAgentDevCli state set <world-id> story_history --data '{
  "entries": [
    {
      "tick_number": 12,
      "summary": "Investigators confirmed the underground resonance chamber.",
      "facts": ["地下52米量子谐振腔", "The chamber remains active below the lower vault."],
      "game_time": "Day 12 - Night"
    }
  ],
  "metadata": {"scenario": "continuity_regression_sample"}
}'

GameAgentDevCli state set <world-id> tick_policy --data '{
  "continuity_rules": [
    "Do not discard established underground reactor facts.",
    "If the chamber is mentioned again, preserve its known depth and status."
  ],
  "focus_scopes": ["lower_vault"],
  "metadata": {"scenario": "continuity_regression_sample"}
}'
```

## Execution Pass

```bash
GameAgentDevCli world tick <world-id>
GameAgentDevCli debug continuity <world-id> --mode debug --log-limit 20 --trace-limit 10
```

## Expected Checks

1. `world_state.canonical_facts` still contains `地下52米量子谐振腔`.
2. The latest `story_history.entries[0].facts` still contains the same fact.
3. The latest timeline summary and future outline still reference the underground chamber when the new tick touches that plotline.
4. `debug continuity` shows the same `request_id` across the linked logs and traces for the inspected tick.
5. Creator `Continuity Diff` shows the chamber fact as stable rather than removed.

## Failure Pattern To Watch

Treat the run as regressed if any of these appear:

- the chamber fact disappears from `canonical_facts`
- the latest history entry replaces the fact with a vaguer phrase such as only `reactor` or `underground facility`
- the latest prompt or trace no longer includes the continuity constraints from `tick_policy`
