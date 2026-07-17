# Engine Improvement Roadmap

[**中文**](./ENGINE_IMPROVEMENT_ROADMAP.md) | **English**

This document captures the current Engine improvement roadmap based on real code inspection, packaged-artifact behavior, and demo world-tick runs.

The goal is not to list every future feature. The goal is to lock down the currently confirmed problems, upgrade directions, priorities, and validation targets so future work does not drift away from the established conclusions.

---

## 1. Background and Current State

In the current implementation:

- a clean demo world can converge during world tick;
- but even a very small world may still require many rounds before completion;
- duplicated imports or polluted world context can push world tick into the `max_analysis_rounds` limit;
- world tick currently starts mostly from imported static world state;
- Worker / authority-state data is not yet a default opening input for world tick.

This leads to two conclusions:

1. the world-tick pipeline is usable but not efficient enough at convergence;
2. the demo workflow still does not fully connect the world skeleton and the authority half of runtime state.

---

## 2. Confirmed Problems

### 2.1 The world-tick prompt leans toward “query more before writing”

The current prompt allows continued `request_data` when facts are missing, but it does not strongly enforce “finish the current tick once the baseline facts are already sufficient.”

As a result:

- the model often keeps querying scene, NPC, relation, and memory details first;
- only later does it commit to `future_outline` and current-tick narration;
- even small worlds can consume a high number of rounds.

### 2.2 `request_data` resolution always loops into the next round

In the current multi-turn loop, once a `request_data` result is resolved, the result is appended and the pipeline continues into the next round.

Missing controls:

- no sufficiency check for “the current facts are enough to finish”;
- no hard convergence rule near `max_analysis_rounds`;
- no world-tick-specific query budget.

### 2.3 Round history keeps accumulating and prompts grow

Each round’s analysis and query result is appended into task-tree / round-state context and carried forward.

That means:

- the prompt increasingly resembles a running research log;
- the model becomes more likely to continue elaborating and querying;
- even small worlds suffer from historical noise competing with the current objective.

### 2.4 Query results are too fragment-oriented

`handleDataRequest` currently returns node detail, memory, relation, and timeline fragments.

The problem is not correctness. The problem is that:

- there is no world-tick-focused middle summary layer;
- the model must still integrate the fragments itself;
- this encourages sequences like “query detail, then query memory, then query relation.”

### 2.5 World-tick opening context is still thin

The demo world skeleton is sufficient for import, dialogue, and play demonstrations, but for world tick:

- the authority facts in `demo-state.yaml` are not automatically injected;
- the current context builder mostly consumes world/nodes/components/memories/relations/state blocks;
- dynamic authority state is not yet part of the default opening input.

### 2.6 Cold start is not yet explicitly separated from world tick

The project still lacks a clear, formal, stable “runtime baseline initialization” capability after import.

As a result:

- developers can easily confuse world tick with initialization;
- developers can also assume they should manually maintain large runtime-state components.

---

## 3. Upgrade Principles

Future Engine upgrades should follow these principles:

1. do not persist high-frequency authority data inside Engine by default;
2. do not require developers to hand-maintain large runtime-state components;
3. world tick should prefer “world skeleton + authority snapshot” as its opening basis;
4. convergence control matters more than simply raising `max_analysis_rounds`;
5. kernel capabilities and workflow entrypoints may be separate, but their semantics must remain unified.

---

## 4. P0: World Tick Bootstrap

### 4.1 Goal

Before the main LLM world-tick loop begins, prefetch a small set of authoritative facts to reduce low-value opening rounds.

### 4.2 Recommended approach

- keep one shared authority-query / `request_data` semantics layer;
- use sync prefetch as the preferred path;
- use callback / paused-execution as the fallback path;
- inject authority results into request-scoped temporary context rather than long-lived components.

### 4.3 Recommended initial query types

- `scene_state`
- `scene_occupants`
- `player_state`
- `player_inventory`
- `task_state`
- `item_presence`
- `npc_state`

### 4.4 Expected gains

- fewer low-value opening query rounds;
- a smaller gap between demo worlds and real authority-driven integration;
- better alignment for play, demos, and integration tests.

---

## 5. P0: Convergence Control

### 5.1 Goal

Prevent world tick from consuming too many rounds in small worlds or repeatedly querying until it hits the round limit in larger worlds.

### 5.2 Recommended changes

- introduce a world-tick-specific query budget;
- cap consecutive query phases;
- enforce hard convergence rules near the maximum round limit;
- define explicit “baseline facts are sufficient” finish conditions.

### 5.3 Expected gains

- less dependence on high `max_analysis_rounds` values;
- more stable round-count distribution;
- faster closure in small worlds.

---

## 6. P0: World Tick Prompt Convergence Rewrite

### 6.1 Goal

Shift the world-tick prompt from “keep refining whenever possible” toward “finish the current tick once the necessary facts are available.”

### 6.2 Recommended direction

- explicitly state that once scene, participants, primary tension, and minimal authority facts exist, the current tick should be completed;
- distinguish critical missing facts from optional enrichment;
- reduce the urge to over-query before producing the first usable world-tick result;
- adapt prompting strategy for small worlds versus larger scopes.

---

## 7. P1: Query Result Summarization

### 7.1 Goal

Reduce the model’s need to manually assemble meaning from fragmented query results.

### 7.2 Recommended changes

- add a world-tick-focused summary layer on top of `handleDataRequest`;
- convert raw node/memory/relation results into higher-level summaries better suited for tick reasoning;
- define a stable authority snapshot block format for bootstrap responses.

### 7.3 Expected gains

- fewer “query detail, then query memory, then query relation” behaviors;
- less dependence on repeated data pulls;
- more consistent and explainable final outputs.

---

## 8. P1: Round Context Compression

### 8.1 Goal

Prevent linear prompt growth and reduce the effect of stale round history.

### 8.2 Recommended direction

- stop carrying every previous query verbatim;
- compress old rounds into staged summaries;
- retain only key facts, key decisions, and unresolved gaps;
- provide a summary-first round-history path for world tick.

---

## 9. P1: World Cold-Start Interface

### 9.1 Goal

Turn “post-import runtime baseline initialization” into a formal Engine capability rather than an implicit workflow assumption.

### 9.2 Recommended capability boundary

- cold start should be separate from import;
- cold start should be separate from world tick;
- it should support both first-time initialization and later rebuild;
- it should generate runtime baselines only, not synchronize dynamic authority state.

### 9.3 Suggested outputs

- success / failure status;
- generated vs reused component list;
- initialization version and source markers;
- warnings about missing anchors, weak modeling, or structural conflicts.

---

## 10. P1: Demo Authority Integration

### 10.1 Goal

Make `demo-state.yaml` participate in world tick through the authority flow instead of only powering Worker play.

### 10.2 Recommended approach

- let Worker act as the demo authority responder;
- let both bootstrap and later authority queries use Worker-backed `demo-state.yaml` answers;
- do not mirror `demo-state.yaml` into Engine components.

### 10.3 Expected gains

- a demo path closer to real integration;
- better test coverage for authority-backed world tick;
- easier detection of query-contract mismatches between Engine and Worker.

---

## 11. P2: Callback Bootstrap Fallback

### 11.1 Goal

Allow expensive or high-latency authority fetches to reuse the existing runtime-task / paused-execution / resume mechanisms.

### 11.2 Recommended principles

- bootstrap and normal authority queries should use the same query semantics;
- prefer synchronous prefetch;
- fall back to callback / resume when a fast answer is unavailable;
- avoid inventing a second bootstrap-only async protocol.

---

## 12. P2: Tooling Workflow Integration

### 12.1 DevCli

Recommended additions:

- `world cold-start`
- `world cold-start --mode rebuild`
- `world bootstrap inspect` or an equivalent debug entrypoint

### 12.2 Creator

Recommended additions:

- an “initialize world” action after import;
- runtime-baseline inspection views;
- bootstrap / authority snapshot debug views.

### 12.3 Worker

Recommended additions:

- a standard demo authority-query responder mode;
- direct support for bootstrap query packs;
- observability for both sync and callback-based authority paths.

---

## 13. Validation and Regression

Each stage of the roadmap should be validated at least against the following baselines:

| Baseline | Target |
|---|---|
| clean demo world | low-round world-tick convergence with complete output |
| demo world with authority bootstrap | correct use of Worker-provided authority snapshots |
| polluted / duplicated world | world tick should not collapse into round-limit failure too easily |
| callback authority scenario | paused execution / resume behavior remains stable |

Recommended key metrics:

- `rounds_used`
- `max_analysis_rounds`
- bootstrap hit rate
- complete `future_outline` generation
- reasonable `advanced_ticks`
- repeated low-value query detection

---

## 14. Priority Summary

| Priority | ID | Upgrade | Goal |
|---|---|---|---|
| P0 | E1 | world tick bootstrap | ensure authoritative baseline facts before the main loop |
| P0 | E2 | convergence control | stop high-round repeated querying even in small worlds |
| P0 | E3 | prompt convergence rewrite | finish the current tick earlier and more reliably |
| P1 | E4 | query result summarization | reduce fragment integration cost |
| P1 | E5 | round context compression | reduce prompt growth and historical noise |
| P1 | E6 | cold-start interface | formalize runtime baseline generation |
| P1 | E7 | demo authority integration | make demo paths exercise the real authority chain |
| P2 | E8 | callback bootstrap fallback | support more complex authority aggregation |
| P2 | E9 | tooling workflow integration | expose the new capabilities through Creator / DevCli / Worker |
| P2 | E10 | regression and acceptance hardening | lock in convergence and integration quality |

---

## 15. Summary

The main problem is not that world tick is unusable. The main problem is that:

- opening authoritative facts are too thin;
- query continuation is too unconstrained;
- convergence control is weak;
- the authority half of the demo workflow is not fully connected yet.

The roadmap should therefore prioritize better opening facts and stronger convergence control before deeper cold-start, tooling, and callback-heavy authority work.
