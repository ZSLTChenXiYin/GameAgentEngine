# Future Development Plan

This file records the current post-cleanup roadmap so future implementation does not lose earlier planning context.

## Planning Update

Current execution order has changed again after the kernel / play restructuring completed:

1. finish the remaining documentation cleanup and workflow alignment
2. reorganize SDK-facing documentation and responsibility boundaries
3. run packaged-artifact acceptance after structural cleanup stabilizes

Engine kernel completion and Worker play deepening are no longer the active unfinished focus areas in this roadmap.

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

### F5 checklist

1. continue removing outdated internal rollout and sync documents
2. trim remaining low-value historical docs that no longer describe the live workflow
3. keep only docs that still define active contracts, boundaries, or design rationale
4. make sure `docs/` remains the only formal documentation tree

Status: in progress.

## F6 tests consolidation

- keep `tools/source/tests` data-only
- move remaining procedural flows into Worker commands

Status: completed for the current worker-driven test workflow baseline.

## F7 SDK doc and responsibility reorganization

- centralize SDK documentation into `docs/`
- keep SDK folders code-and-examples only
- continue aligning SDK outward-facing responsibilities with the Go SDK baseline

Status: pending.

## F8 Packaged artifact acceptance

- verify Engine / DevCli / Worker / Creator packaged workflow after structural cleanup

Status: pending.

## Deferred But Tracked
- world-tick context roadmap, including `world_focus`, active-node selection, and staged scope refinement; see `docs/internal/WORLD_TICK_CONTEXT_ROADMAP.md`
- autonomous scheduling roadmap, including priority, batching, lifecycle state, and event-driven wake-up; see `docs/internal/AUTONOMOUS_SCHEDULING_ROADMAP.md`
- roleplay interaction roadmap, including direct single-chat, group-chat, interaction-session modeling, and player-intent bridging; see `docs/internal/ROLEPLAY_INTERACTION_ROADMAP.md`

- broader documentation slimming beyond active contract cleanup
- deeper multi-NPC group-chat reasoning, if still needed after play/kernel stabilization
- later SDK expansion work for non-Go ecosystems
- future Engine kernelization work after contract maturity; see `docs/internal/ENGINE_KERNELIZATION_MEMO.md`
