# Future Development Plan

[**中文**](./FUTURE_DEVELOPMENT_PLAN.md) | **English**

This file records the current post-cleanup roadmap so future implementation does not lose earlier planning context.

## Planning Update

The execution order for unfinished work is now:

1. fix Creator large-tree performance first
2. continue remaining Engine roadmap work only after Creator usability is stable at scale
3. handle broader documentation slimming and later expansion work after that

Engine kernel completion and Worker play deepening are already done. The highest-priority unfinished problem now is Creator outline responsiveness once the world tree reaches ten-thousand-scale nodes, because it directly impacts editing and debugging efficiency.

## F0 Creator large-tree performance optimization

- keep the Creator left outline usable, searchable, scrollable, and editable at 10k+ scale
- stop relying on a rendering model where every click, filter change, and collapse rebuilds the full tree
- solve visible-region rendering and local-update behavior first, then layer in finer interaction improvements

### F0 checklist

1. establish 1k / 5k / 10k profiling baselines for first paint, scroll, expand, collapse, filter, and selection latency
2. split tree data preparation from DOM rendering and cache structures such as `nodeMap`, `childMap`, and flattened visible rows
3. replace recursive full-tree rendering with a flattened visible-row model
4. add virtual scrolling so only viewport rows are rendered
5. convert expand / collapse / selection / drag feedback into local refreshes instead of full `renderTree()` rebuilds
6. replace per-node event binding with container-level event delegation
7. add indexed or incremental filtering paths for name/type search
8. add large-tree degradation rules such as default collapse, on-demand expansion, and search-first navigation
9. add acceptance criteria and regression samples for large-tree Creator scenarios

Status: not started. This is now the highest-priority future development item.

## F1 Engine core purification

- remove Engine-side developer-tool commands that already have clear owners
- keep Engine focused on service runtime and version metadata
- decide the final fate of validate after DevCli-side coverage exists

### F1 checklist

1. remove Engine `creator` once DevCli remains the single entry
2. remove Engine `import` once DevCli / Creator paths remain intact
3. upgrade DevCli `init`, then remove Engine `init`
4. compare Engine `test` against Worker `test`, then remove or merge
5. migrate or delete Engine `validate`
6. leave Engine with `serve` and `version` only, optionally `validate` during transition

Status: completed.

## F2 Worker CLI restructuring

- move Cobra command definitions from `internal/workercli` into `cmd/gameagentworker`
- keep reusable business logic in internal packages only

Status: completed.

## F3 Engine kernel completion

- finish the kernel-side interaction model instead of leaving it scattered across Worker-only flow glue
- keep Engine suitable for embedding while still making actor/target/scene/participant semantics first-class
- complete the kernel-side contract for player-intent interpretation, authority-query semantics, and interaction resume flow

### F3 checklist

1. finish interaction API semantics on top of `invoke`
2. align `interaction/*`, `player/input/interpret`, and plain `invoke` around one coherent kernel contract
3. stabilize actor / target / scene / participant modeling for direct dialogue and group chat
4. tighten the kernel-side player intent schema, validation vocabulary, and response contract
5. verify Engine-side authority-query / runtime-task / callback-resume behavior still matches the intended embedding boundary
6. reduce Worker-side ad-hoc glue where the contract really belongs to Engine

Status: completed.

## F4 Worker play deepening

- evolve play into a real text-game shell instead of a thin engine wrapper
- build on top of the stabilized Engine interaction kernel instead of compensating for missing kernel semantics
- keep play ahead of documentation polish, but after kernel completion

### F4 checklist

1. improve play command semantics and turn flow
2. deepen room feedback, target switching, and interaction presentation
3. improve `/act` bridge quality after kernel-side intent contract is stable
4. revisit group-chat behavior and decide whether to keep one-primary-responder mode or extend it

Status: completed.

## F5 Documentation centralization and slimming

- delete `tools/source/docs`
- keep only packaged runtime assets under `tools/source`
- move SDK overview / baseline / capability docs into `docs/`
- remove scattered READMEs under command and SDK folders after migration
- normalize formal documentation filenames to uppercase naming
- require every formal documentation page to have both Chinese and English counterparts

### F5 checklist

1. continue removing outdated internal rollout and sync documents
2. trim remaining low-value historical docs that no longer describe the live workflow
3. keep only docs that still define active contracts, boundaries, or design rationale
4. make sure `docs/` remains the only formal documentation tree
5. normalize surviving formal doc filenames to uppercase naming
6. add or backfill English/Chinese counterparts for every surviving formal doc page

Status: completed.

## F6 tests consolidation

- keep `tools/source/workerhome/fixtures` data-only
- move remaining procedural flows into Worker commands

Status: completed for the current worker-driven test workflow baseline.

## F7 SDK doc and responsibility reorganization

- centralize SDK documentation into `docs/`
- keep SDK folders code-and-examples only
- continue aligning SDK outward-facing responsibilities with the Go SDK baseline
- apply the same uppercase naming and bilingual-document rule to SDK-facing formal docs

Status: completed.

## F8 Packaged artifact acceptance

- verify Engine / DevCli / Worker / Creator packaged workflow against local release packages
- keep GitHub release automation out of the current completion gate

### F8 checklist

1. verify the 6 target-platform packages and their zip archives
2. verify packaged Engine startup and DevCli connectivity
3. verify packaged Worker tooling-smoke completion
4. verify one baseline Go SDK smoke scenario
5. align README / docs with release package paths

Status: completed.

## Deferred But Tracked

- Creator large-tree performance roadmap: `docs/internal/CREATOR_TREE_PERFORMANCE_ROADMAP_EN.md`

- recommended conventions for world modeling, runtime baselines, authoritative dynamic state, and world-tick bootstrap; see `docs/architecture/WORLD_MODELING_AND_RUNTIME_CONVENTIONS_EN.md`
- the current Engine improvement roadmap derived from real world-tick convergence issues; see `docs/internal/ENGINE_IMPROVEMENT_ROADMAP_EN.md`

- world-tick context roadmap, including `world_focus`, active-node selection, and staged scope refinement; see `docs/internal/WORLD_TICK_CONTEXT_ROADMAP.md`
- autonomous scheduling roadmap, including priority, batching, lifecycle state, and event-driven wake-up; see `docs/internal/AUTONOMOUS_SCHEDULING_ROADMAP.md`
- roleplay interaction roadmap, including direct single-chat, group-chat, interaction-session modeling, and player-intent bridging; see `docs/internal/ROLEPLAY_INTERACTION_ROADMAP.md`

- broader documentation slimming beyond active contract cleanup
- deeper multi-NPC group-chat reasoning, if still needed after play/kernel stabilization
- later SDK expansion work for non-Go ecosystems
- future Engine kernelization work after contract maturity; see `docs/internal/ENGINE_KERNELIZATION_MEMO.md`
