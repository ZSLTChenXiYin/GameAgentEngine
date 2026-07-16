# SDK Shared Fixtures and Integration Inputs

This document defines the shared fixture files and integration inputs that SDK examples should reuse whenever possible.

The goal is to keep all SDK examples aligned to the same Engine / Worker semantics instead of letting each language drift into its own ad-hoc sample data.

## 1. Primary Fixture Directory

Shared worker-side and runtime-task fixture data lives in:

```text
tools/source/tests/
```

SDK examples should prefer these files before inventing new sample payloads.

## 2. Core Shared Files

| File | Purpose | Typical Consumers |
| --- | --- | --- |
| `runtime_task_dynamic_interfaces.json` | Request-scoped `game_client_request_data` example used to trigger authority-query runtime tasks | ts/js/cs SDK worker authority-query examples, future Java/Lua/C++ examples |
| `runtime_task_dynamic_action_trade.json` | Pull-mode external action example | Worker runtime-task scenarios, future pull examples |
| `runtime_task_delivery_fixture.json` | Fixture LLM output used by worker runtime-task integration testing | `GameAgentWorker test runtime-tasks`, future SDK orchestration docs |
| `machine_scenario_fixture.json` | End-to-end worker + continuity + callback-resume fixture | `GameAgentWorker test machine-scenario`, future SDK smoke loops |
| `callback_resume_fixture.json` | Callback-resume base fixture | Worker callback-resume scenario |
| `callback_resume_dynamic_actions.json` | Dynamic action follow-up fixture for callback-resume | Worker callback-resume scenario |
| `full_functional_base_data_world.yaml` | Base importable world fixture for worker-side full-functional scenarios | Engine / Worker test bootstrap |
| `world_time_settings_flexible.json` | Flexible world-time settings sample | continuity / machine scenario / future SDK settings examples |
| `state_world_state.json` | Example `world_state` continuity payload | SDK state-component examples |
| `state_story_state.json` | Example `story_state` continuity payload | SDK state-component examples |
| `state_story_history.json` | Example `story_history` continuity payload | SDK continuity examples |
| `state_tick_policy.json` | Example `tick_policy` continuity payload | SDK continuity examples |

## 3. Repository-Level Demo Assets

These are not under `tools/source/tests`, but they are still shared integration assets:

| File | Purpose |
| --- | --- |
| `tools/source/demo-world.yaml` | Demo world import used by Engine / DevCli / Worker quick-start |
| `tools/source/demo-state.yaml` | Worker play-mode authority-state sample |

## 4. Fixture Usage Rules

SDK examples should follow these rules:

1. reuse `runtime_task_dynamic_interfaces.json` for authority-query examples;
2. reuse `tools/source/demo-world.yaml` and `tools/source/demo-state.yaml` for play-mode-facing walkthroughs;
3. avoid embedding large inline JSON blobs in every language example when a shared file already exists;
4. if one SDK needs a new fixture, add it under `tools/source/tests/` only when it is reusable by at least one other SDK or Worker scenario.

## 5. Current Shared Example Patterns

At the moment, the practical SDK examples converge around two shared patterns:

- runtime task pull / claim / start / callback roundtrip
- authority query trigger via `game_client_request_data`, then hand-off to `GameAgentWorker pull-once`

These two patterns are the baseline for language-to-language consistency.
