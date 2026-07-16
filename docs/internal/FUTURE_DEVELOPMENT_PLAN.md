# Future Development Plan

This file records the current post-cleanup roadmap so future implementation does not lose earlier planning context.

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

## F2 Worker CLI restructuring

- move Cobra command definitions from `internal/workercli` into `cmd/gameagentworker`
- keep reusable business logic in internal packages only

## F3 Documentation centralization

- delete `tools/source/docs`
- keep only packaged runtime assets under `tools/source`
- move SDK overview / baseline / capability docs into `docs/`
- remove scattered READMEs under command and SDK folders after migration

## F4 Engine interaction API track

- formalize actor / target / scene / audience interaction semantics on top of invoke
- support direct dialogue and group chat flows

## F5 Player natural-language intent mapping

- map player natural language into a player-node intent proposal
- validate against authority state before committing game-side truth

## F6 Worker play deepening

- evolve play into a real text-game shell instead of a thin engine wrapper
- unify slash-command and natural-language act flows

## F7 tests consolidation

- keep `tools/source/tests` data-only
- move remaining procedural flows into Worker commands

## F8 SDK doc reorganization

- centralize SDK documentation into `docs/`
- keep SDK folders code-and-examples only

## F9 Packaged artifact acceptance

- verify Engine / DevCli / Worker / Creator packaged workflow after structural cleanup
