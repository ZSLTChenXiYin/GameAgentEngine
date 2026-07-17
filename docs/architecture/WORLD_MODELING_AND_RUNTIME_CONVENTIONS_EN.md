# World Modeling and Runtime Conventions

[**中文**](./WORLD_MODELING_AND_RUNTIME_CONVENTIONS.md) | **English**

This document defines the recommended conventions for world modeling, runtime baseline generation, authoritative state handling, world-tick bootstrap, and write-back rules in GameAgentEngine.

The goal is not to create more configuration burden. The goal is to make the following boundaries explicit:

- what developers should maintain directly;
- what Engine should derive automatically;
- what the game side / Worker should own;
- which facts may enter Engine persistence and which facts must stay request-scoped.

---

## 1. Scope

This document applies to:

- world YAML / JSON imports;
- demo world design;
- GameAgentWorker play / authority-state workflows;
- world tick / scope tick reasoning;
- Creator, DevCli, Worker, and Engine responsibility boundaries;
- future SDK and game-side integration contracts.

---

## 2. Three-Layer Data Model

World-related data should be split into three layers instead of being maintained as one mixed state.

| Layer | Role | Typical content | Primary owner | Persisted in Engine |
|---|---|---|---|---|
| World-definition layer | Stable world facts | world, location, npc, item, profile, lore, stable relations, base rules | developer / Creator | yes |
| Runtime-baseline layer | Engine continuity and reasoning starting state | story_state, world_state, story_history, state_snapshot, initial future outline | Engine cold start | yes |
| Authoritative dynamic layer | High-frequency, strongly authoritative facts | HP, money, inventory, quest state, scene occupancy, live flags | game side / Worker | no |

Core principle:

- the world-definition layer describes what the world is;
- the runtime-baseline layer describes the current narrative baseline Engine should continue from;
- the authoritative dynamic layer describes what is currently true on the game side.

---

## 3. World-Definition Modeling Rules

Developers should primarily maintain stable, low-frequency, long-lived world facts.

Recommended contents for importable world files:

- world roots, scene nodes, NPC nodes, item nodes;
- stable parent/child structure and stable relations;
- profile, lore, and base rule components;
- a small number of key memories that act as narrative anchors;
- organizations, geography, factions, and social structures that do not depend on real-time authority values.

Not recommended as long-lived Engine-side component data:

- current HP / max HP;
- current money;
- current inventory details;
- current quest live stage;
- current scene occupancy;
- live weather, live risk flags, temporary combat state;
- any fact that changes frequently and must remain game-authoritative.

---

## 4. Runtime-Baseline Layer

The runtime-baseline layer is Engine-owned runtime state used to support continuity and world reasoning. Developers should not be required to hand-maintain large amounts of it.

Recommended cold-start-generated content:

- `world_state`: global narrative summary, scene pressure, high-level world situation;
- `story_state`: current arc stage, primary tension, active conflict, key actor summary;
- `story_history`: initial history entries and later continuity entries;
- `state_snapshot`: initialization snapshot of the runtime baseline;
- initial `future_outline`: coarse-grained future world progression outline;
- a default `world_time_state` instance where needed.

Why this should not be heavily hand-authored:

- it easily drifts from the actual world-definition input;
- it becomes stale when the world structure changes;
- much of it is derived state rather than source-of-truth world design;
- it turns the workflow into “fill enough state so the world can run.”

---

## 5. Authoritative Dynamic Layer

High-frequency, strongly authoritative, fast-changing facts should remain on the game side or Worker side and be queried on demand by Engine.

Typical examples:

- HP / MP / money for players and NPCs;
- current inventory, equipment, and item ownership;
- current quest phase and interaction availability;
- current scene occupants, live weather, and live event flags;
- current location occupancy, locks, combat, downed, death, or other live runtime status.

These facts should not be persisted as long-lived Engine components by default.

Why:

- they drift quickly;
- they can diverge from the game side before and after reasoning;
- they create double-write state;
- they weaken the purpose of callback / authority-query integration.

---

## 6. Demo Data Split

Demo data should be split into two files with different responsibilities.

| File | Responsibility |
|---|---|
| `demo-world.yaml` | importable Engine world skeleton and stable world-definition input |
| `demo-state.yaml` | Worker / game-side authority-state sample |

Recommended rule:

- `demo-world.yaml` carries stable world definition;
- `demo-state.yaml` carries dynamic authoritative state;
- the two are joined through runtime query flow rather than static synchronization.

Anti-patterns:

- mirroring `demo-state.yaml` into long-lived Engine components;
- simulating live HP, inventory, occupants, or quest state as persistent components;
- weakening authority boundaries just to make demo world-tick converge faster.

---

## 7. World Cold Start

World cold start means:

- after the world-definition input is imported;
- Engine derives the initial runtime baseline from the static world structure;
- the world becomes ready for continued ticking.

Cold start is not:

- a world tick;
- an implicit import side effect by definition;
- authority-state synchronization.

Recommended cold-start responsibilities:

1. validate that the world skeleton is runnable;
2. derive initial `story_state` / `world_state`;
3. create initial `story_history` entries;
4. create an initialization snapshot;
5. derive an initial `future_outline`;
6. mark the world as runtime-initialized.

Cold start must not persist:

- current HP;
- current money;
- current inventory;
- current scene occupancy;
- any fact that must be answered by the authoritative game side at runtime.

---

## 8. World Tick Bootstrap

World tick bootstrap is an authority-prefetch phase that runs before the main LLM world-tick reasoning loop starts.

Its purpose is not to replace `request_data`. Its purpose is to eliminate low-value, repetitive opening queries.

Recommended flow:

1. Engine builds the static world context;
2. Engine generates a small set of key authority queries based on the current scope, scene, participants, and task type;
3. the game side / Worker returns an authority snapshot;
4. Engine injects that snapshot into the current request as temporary context;
5. the LLM starts world tick on top of “world skeleton + authority snapshot”;
6. only remaining fact gaps should use additional `request_data`.

Recommended bootstrap query types:

- `scene_state`
- `scene_occupants`
- `player_state`
- `player_inventory`
- `task_state`
- `item_presence`
- `npc_state`

---

## 9. Bootstrap Query Semantics and Execution Modes

There should be one query semantics layer and two possible execution modes.

| Layer | Recommendation |
|---|---|
| query semantics | keep one shared `request_data` / authority-query contract |
| execution mode | allow sync-first with callback fallback |
| transport | allow HTTP / WS / RPC / local Worker implementations |

Recommended rule:

- sync-first for demos, Worker-driven tests, local integration, and low-latency service environments;
- async callback fallback for expensive aggregation, unstable clients, or higher latency scenarios;
- do not create a second incompatible bootstrap-only query language.

---

## 10. Write-Back Rules

Engine should persist narrative results, not raw high-frequency authority facts.

Allowed write-back targets:

- `story_history` entries;
- `story_state` stage changes;
- `world_state` narrative changes that remain valid beyond the current request;
- timeline summaries;
- confirmed short-term or shared memories;
- game-side-confirmed narrative results suitable for later continuity.

Not allowed as default write-back:

- raw HP / money / inventory;
- raw live quest state;
- raw scene occupancy;
- temporary authority snapshots from a single request;
- transient facts used only for the current inference round.

---

## 11. Scope and Context Boundaries

World tick, scope tick, dialogue, and play should not default to expanding the entire world graph.

Recommended boundary rules:

- current scope / scene / participants first;
- environment node and key ancestor chains second;
- deeper descendants only when task-relevant;
- bounded promotion of deeper nodes instead of recursive full expansion.

If future `world_focus` or active-node selection is introduced, it should still follow:

- summarize first, refine second;
- select active nodes before expanding sub-scopes;
- keep world tick bounded instead of turning it into full-graph prompt stuffing.

---

## 12. Creator / DevCli / Worker / Engine Responsibility Split

| Module | Primary responsibility |
|---|---|
| Creator | edit world-definition input, inspect runtime state, trigger cold-start and debug workflows |
| DevCli | import, initialize, debug, regress, package, and script workflows |
| Worker | provide authority-state, external callback behavior, and play/demo integration |
| Engine | runtime world reasoning, cold start, world tick continuity, and controlled write-back |

Recommended rule:

- Creator is for world-definition editing;
- DevCli is for workflow control;
- Worker is for authoritative response and integration;
- Engine is for kernel reasoning and continuity.

---

## 13. Anti-Patterns

The following are not recommended:

1. persisting high-frequency authoritative state as long-lived Engine components;
2. forcing developers to hand-maintain large `story_state` / `world_state` bodies;
3. creating a second unrelated bootstrap-only query contract;
4. letting world tick expand the entire world graph by default;
5. using only the static demo world to stand in for a real authority chain;
6. treating world tick as cold start;
7. writing temporary authority snapshots directly into long-lived persistence.

---

## 14. Summary Principle

This document can be summarized as:

> Developers provide world-definition input, Engine derives runtime baselines and continuity state, and the game side owns authoritative dynamic state. High-frequency authority data is queried on demand, while only narrative results are persisted back into Engine.
