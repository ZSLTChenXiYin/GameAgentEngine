# GameAgentEngine Development Plan Baseline

[**中文**](./DEVELOPMENT_PLAN.md) | **English**

This document is the authoritative development plan baseline for the GameAgentEngine project. It consolidates existing roadmaps, code analysis findings, and identified improvement gaps, ordered by priority with detailed implementation approaches for each item.

---

## 1. Plan Overview

### Status Legend

| Symbol | Meaning |
|---|---|
| [ ] | Not started |
| [->] | In progress |
| [x] | Completed |

### Priority Matrix

| Priority | ID | Item | Status | Reference |
|---|---:|---|:---:|---|
| P0 | F0 | Creator large-scale tree performance | [x] | CREATOR_TREE_PERFORMANCE_ROADMAP_EN.md |
| P0 | E1 | World Tick Bootstrap | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P0 | E2 | Convergence control | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P0 | E3 | World Tick prompt convergence rewrite | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P1 | E4 | Query result summarization | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P1 | E5 | Round context compression | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P1 | E6 | World cold-start interface | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P1 | E7 | Demo authority integration | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P1 | E8 | Store layer test coverage hardening | [x] | This doc |
| P1 | E9 | PipelineMode relation assembly enforcement | [x] | This doc |
| P2 | E10 | Callback bootstrap fallback | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P2 | E11 | Action system schema validation/extensions | [x] | This doc |
| P2 | E12 | PolicyEngine conflict resolution | [x] | This doc |
| P2 | E13 | Frontend interaction responsiveness | [x] | This doc |
| P2 | E14 | Autonomous scheduling roadmap | [x] | AUTONOMOUS_SCHEDULING_ROADMAP_EN.md |
| P2 | E15 | World Tick Context roadmap | [x] | WORLD_TICK_CONTEXT_ROADMAP_EN.md |
| P3 | E16 | SDK documentation and examples | [x] | This doc |
| P3 | E17 | Component system generalization | [x] | This doc |
| P3 | E18 | Telemetry observability enhancement | [x] | This doc |
| P3 | E19 | Tooling workflow integration | [x] | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| P3 | E20 | Frontend i18n completion | [x] | This doc |
| P4 | E21 | Multi-world isolation boundaries | [x] | This doc |
| P4 | E22 | Stress tests and benchmarks | [x] | This doc |
| P4 | E23 | Engine Kernelization | [x] | ENGINE_KERNELIZATION_MEMO_EN.md |
| P4 | E24 | Multi-language SDK expansion | [x] | This doc |

### Priority Definitions

| Level | Definition |
|---|---|
| P0 | Must-complete in current cycle; prerequisite for other work |
| P1 | High-value improvements to start right after P0 |
| P2 | Important but non-blocking; schedule as resources allow |
| P3 | Valuable enhancements with no time pressure |
| P4 | Long-term planning; requires further evaluation |

---

## 2. P0: Current Top Priority

### F0: Creator Large-Scale Tree Performance

See CREATOR_TREE_PERFORMANCE_ROADMAP_EN.md for the complete roadmap.

Key steps: 1k/5k/10k node baseline profiling, flat visible-row model, virtual scrolling, local updates replacing full-tree rebuilds, container-level event delegation, indexed search with debouncing.

---

### E1: World Tick Bootstrap

Prefetch authoritative facts before the main LLM loop enters its first round. Reuse existing request_data/handleDataRequest semantics. Inject results into request-scoped temporary context (not persistent components).

Pre-fetch types: scene_state, scene_occupants, player_state, player_inventory, task_state, item_presence, npc_state.

---

### E2: Convergence Control

Introduce world-tick-specific query budget, per-phase round limits, hard convergence rules near max_analysis_rounds (80% threshold), and a formal convergenceCheck() termination condition.

---

### E3: World Tick Prompt Convergence Rewrite

Rewrite buildWorldTickPrompt with completion-first directives. Grade queries as critical vs. nice-to-have. Differentiate prompt strategies for small vs. large worlds. Reorder: tick summary before future_outline.

Target: 30%+ reduction in rounds_used for the same world.

---

## 3. P1: High Priority

### E4-E7

See ENGINE_IMPROVEMENT_ROADMAP_EN.md for details.

- E4: Query result summarization layer
- E5: Round context compression (summary-first, trim historical noise)
- E6: ColdStartWorld() interface (no LLM call, initial/rebuild modes)
- E7: Worker as Demo authority responder

### E8: Store Layer Test Coverage Hardening

Targets: PausedExecution paths, component CRUD transaction boundaries, relation batch operations, migrations and write-retry, snapshot create/validate/restore, world settings and policy consistency.

### E9: PipelineMode Relation Assembly Enforcement

ContextBuilder.Build must respect per-TaskType relation assembly strategies. NPC dialogue prioritizes environment relations; world tick prioritizes summary relations. IncludeRelatedNodes must enforce hop limits. Vertical mode uses minimum closed-loop subgraphs.

---

## 4. P2: Medium Priority

See the individual roadmap docs for E10 (Callback Bootstrap), E14 (Autonomous Scheduling), and E15 (World Tick Context).

### E11: Action System Schema Validation

Standardize Action.Schema() declarations, promote validateActionCallsBySchema to all paths, add RegisterExternal() for third-party injection, async action timeout control.

### E12: PolicyEngine Conflict Resolution

Explicit rule ordering: blocked > allowed > safe. Scope-level priority (world > scope > node). Parameter-level restrictions.

### E13: Frontend Interaction Responsiveness

API request caching and deduplication, optimistic updates with rollback, lazy loading with skeleton screens, virtual scrolling for logs/traces.

---

## 5. P3: Low Priority

SDK documentation and examples (E16), component system generalization (E17), structured telemetry (E18), tooling workflow integration in DevCli/Creator/Worker (E19), i18n completion (E20).

---

## 6. P4: Long-term Planning

Multi-world isolation (E21), stress testing/benchmarking (E22), Engine Kernelization (E23), multi-language SDKs starting with TypeScript/Python (E24).

---

## 7. Acceptance and Regression

| Baseline | Goal |
|---|---|
| Clean demo world | Low-round world tick with complete output |
| Demo world with authority bootstrap | Correct use of authority snapshots |
| Re-imported polluted world | No round-limit exhaustion from noise |
| Callback authority scenarios | Stable pause/resume lifecycle |

| Metric | Target |
|---|---|
| rounds_used / max_analysis_rounds ratio | < 0.6 |
| Bootstrap hit rate | > 80% |
| future_outline completeness | Every tick |
| Repetitive low-value queries | 0 |
| Test pass rate | 100% |
| Creator 10k node first-paint | < 2s |

---

## 8. Roadmap Document Index

| Document | English link |
|---|---|
| Engine Improvement Roadmap | ENGINE_IMPROVEMENT_ROADMAP_EN.md |
| Creator Tree Performance Roadmap | CREATOR_TREE_PERFORMANCE_ROADMAP_EN.md |
| Autonomous Scheduling Roadmap | AUTONOMOUS_SCHEDULING_ROADMAP_EN.md |
| World Tick Context Roadmap | WORLD_TICK_CONTEXT_ROADMAP_EN.md |
| Engine Kernelization Memo | ENGINE_KERNELIZATION_MEMO_EN.md |
| Future Development Plan | FUTURE_DEVELOPMENT_PLAN_EN.md |
| Roleplay Interaction Roadmap | ROLEPLAY_INTERACTION_ROADMAP_EN.md |
| Player Input Pipeline | PLAYER_INPUT_PIPELINE_EN.md |
| Interaction API | INTERACTION_API_EN.md |
| Engine Query Contract | ENGINE_QUERY_CONTRACT_EN.md |
| Engine Graph Semantics | ENGINE_GRAPH_SEMANTICS_EN.md |

---

*This document consolidates all existing roadmaps and planning into a single reference for team development decisions. The development team should update the status column at the end of each sprint.*
