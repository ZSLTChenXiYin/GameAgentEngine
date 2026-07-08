# Project Roadmap

[**中文**](./ROADMAP.md) | **English**

Updated: 2026-07-08

Current baseline version: `v0.4.5`

---

## 1. Purpose

This roadmap explains the current stage of GameAgentEngine, the next development priorities, and the synchronization principles across the four product lines: Engine, DevCli, Creator, and SDK.

The roadmap follows `Engine` as the primary track:

- `Engine` defines the real capability boundary first
- `DevCli` and `Creator` expose those capabilities to developers
- `SDK` packages those capabilities into stable integration interfaces
- docs, examples, and packaging artifacts follow afterward to prevent drift

---

## 2. Project Positioning

GameAgentEngine is a runtime engine between game logic and large-model capabilities. Its goal is to provide game developers with:

- world modeling based on nodes, components, memories, and relations
- LLM-driven NPC reasoning, dialogue, and world progression
- a controllable runtime action system
- an observable, debuggable, and integratable toolchain

Supporting tool responsibilities are:

- `Engine`: reasoning, state evolution, persistence, policy, logging, continuity, and world tick
- `DevCli`: a development tool for command-line and AI-assisted workflows
- `Creator`: a visual tool for world inspection, editing, and troubleshooting
- `SDK`: Go integration wrappers for games or backend services

---

## 3. Current Baseline Capabilities

### Engine

Currently available:

- world graph model: nodes, components, memories, relations
- three pipeline modes: `vertical`, `polling`, `full`
- `world_tick`, `event_impact`, `scope_advance`, `timeline_replan`
- world settings, world policy, pending review plans
- working copies, save snapshots, snapshot restore
- continuity state components: `world_state`, `story_state`, `story_history`, `tick_policy`, `state_snapshot`
- timeline archives, structured logs, debug traces

### DevCli

Currently available:

- world import and validation
- CRUD for nodes / components / memories / relations
- world advance, event impact evaluation, scope advance, timeline replan
- world rename, fork, snapshot, restore
- log, continuity, state component, and timeline read capabilities

### Creator

Currently available:

- visual editing for worlds and nodes
- node copy, relation editing, drag-and-drop reparenting
- world settings, world policy, and pending-plan inspection
- `Continuity`, `State`, `Timelines`, `Logs`, `Traces` pages
- snapshot management and continuity troubleshooting entry points

### SDK

Currently available:

- basic inference calls
- world settings and world policy access
- log queries, continuity bundle loading, debug trace loading
- wrappers for state components, timelines, snapshots, approval plans, and related interfaces

---

## 4. Roadmap Principles

Future development follows these principles:

1. Land changes in `Engine` first, then synchronize `SDK`, `DevCli`, and `Creator`.
2. Observability takes priority over black-box intelligence; the debugging chain must be traceable.
3. `world_tick` continuity stability takes priority over surface-level new gameplay features.
4. Structured state takes priority over temporary stitched text; if something can become a component, it should.
5. Docs, examples, mirrored docs, and packaging scripts must be updated together with version changes.

---

## 5. Near-Term Priorities

### P0: Fill in pipeline observability

Goal: let developers inspect the full internal path of a single inference run instead of only seeing a black-box result.

Focus items:

- in `debug` and `review` modes, write complete internal pipeline runtime information into the `logs` table
- in `production` mode, keep only the minimum necessary production logs
- define different console logging granularity for `debug`, `review`, and `production`
- establish a unified `request_id` view that connects logs, traces, timelines, and results
- make it easier for `Creator` and `DevCli` to filter by `request_id`, event name, and mode

Expected result:

- a single inference run can be fully replayed across input, rounds, data requests, parse results, action decisions, and final output
- pipeline tuning no longer depends on guesswork, but on structured evidence

### P0: Govern `world_tick` continuity drift

Goal: reduce story-context drift, fact loss, and summary inconsistency during world progression.

Focus items:

- clarify which world tick outputs should be persisted long-term and which should remain transient inference artifacts
- continue moving high-value continuity information into structured components instead of scattering it across logs and free text
- strengthen the responsibility boundaries among `world_state`, `story_state`, `story_history`, `tick_policy`, and `state_snapshot`
- establish a troubleshooting loop around continuity diff, fact verification, and anomaly backtracking
- analyze drift sources across pre-tick context construction, in-tick reasoning, and post-tick writeback

Expected result:

- key settings, facts, and storylines remain more stable across multi-round progression in the same world
- developers can more quickly judge whether a problem comes from the prompt, state writeback, or context assembly

### P1: Expand the state component system

Goal: pull more high-value runtime information out of implicit text and into manageable, validatable, editable components.

Candidate directions:

- extract stage conclusions from story progression into new components
- extract world-level decision constraints into clearer policy components
- structure intermediate reasoning summaries, key fact sets, and conflict clue sets
- keep improving Creator and DevCli validation and editing for these components

### P1: Fill in toolchain synchronization

Goal: avoid a situation where Engine capabilities exist but are only indirectly reachable through raw API access.

Focus items:

- `SDK` provides stable interfaces for new Engine capabilities
- `DevCli` adds read/write, query, and filtering commands for new interfaces
- `Creator` adds visual entry points for new state and log views
- keep field names, event names, and mode names fully aligned across all four lines

### P2: Version and documentation governance

Goal: control documentation drift and outdated distribution information after feature evolution.

Focus items:

- keep root docs synchronized with mirrored docs under `tools/source/docs`
- keep version numbers, example commands, field descriptions, and event names aligned
- update build scripts, packaging scripts, and release notes together with version changes

---

## 6. Suggested Delivery Phases

### Phase A: Logging and mode stratification

- unify logging policy for `debug` / `review` / `production`
- define the responsibility boundary between console logs and database logs
- fill in event names and structured detail payloads for key reasoning stages

### Phase B: `world_tick` persistence audit

- audit where every generated data artifact lands during tick execution
- mark which data should enter the node / component system
- mark which data should only go into `timelines`, `logs`, and `traces`

### Phase C: Continuity component refactor

- merge duplicated semantics
- add missing state component types
- adjust state write and read flows

### Phase D: Troubleshooting workflow formation

- use `request_id` as the center point connecting continuity, timelines, logs, and traces
- form an actionable troubleshooting path in Creator
- form scriptable troubleshooting commands in DevCli

### Phase E: SDK / DevCli / Creator synchronization

- expose new interfaces in SDK
- add commands and help text in DevCli
- add pages and explanatory copy in Creator

### Phase F: Documentation and release wrap-up

- align docs
- update examples
- update packaging scripts and version numbers

---

## 7. Mid-Term Goals

Once the near-term work is stable, the project's mid-term goals are:

- form a more reliable `world_tick` continuity system
- form a standard observability scheme for debug, review, and production environments
- make `Creator` the front-line tool for continuity troubleshooting and world editing
- make `DevCli` the main development entry point for AI collaboration and automation scripts
- make `SDK` the stable boundary layer for external game integration

---

## 8. Long-Term Direction

Over the long term, GameAgentEngine needs to move from "can run" to "can be developed against reliably, tuned continuously, and integrated into production projects."

Long-term directions include:

- a more mature world-state modeling standard
- a more stable boundary between policy approval and automatic execution
- stronger inference explainability and replay capability
- a more complete multi-tool collaborative development experience
- a clearer version compatibility and migration strategy

---

## 9. Maintenance Notes

This roadmap is not a history log. It is a statement of the current development direction. After each version iteration, at minimum the following should be updated together:

- current version number
- current completed capability baseline
- near-term priorities
- phased delivery status
- any tool or documentation items that no longer match the real `Engine` implementation

It is recommended to review this file once for every `minor` release.
