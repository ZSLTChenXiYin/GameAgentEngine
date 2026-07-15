# Full Functional Test Plan

This document defines the staged full-functional validation plan for the current GameAgentEngine repository.

## Goal

Build one repeatable validation workflow that covers:

- source-level regression safety
- base data CRUD and world configuration
- world tick, continuity, timeline, and debug surfaces
- runtime task push / pull / hybrid delivery
- callback, paused execution resume, and callback post-process
- SDK / DevCli / Creator smoke validation
- machine-style end-to-end NPC dialogue scenarios

## Test Layers

| Layer | Scope | Primary Tools | Exit Condition |
|---|---|---|---|
| L0 | Automated regression baseline | `go test ./...` | all automated tests pass |
| L1 | Base data plane | HTTP, DevCli | CRUD and world configuration behave consistently |
| L2 | World evolution plane | DevCli, logs, traces, continuity | tick/state/timeline surfaces stay coherent |
| L3 | External interaction plane | Engine, test worker, runtime tasks | push/pull/hybrid delivery paths are usable |
| L4 | Resume and orchestration plane | Engine, callback chain, worker | callback/resume/post-process semantics match design |
| L5 | Tooling and machine scenario | SDK, DevCli, Creator, packaged or isolated runtime | one real end-to-end scenario is repeatable |

## Execution Order

1. Establish the environment baseline.
2. Implement and validate a dedicated external test worker.
3. Run source-level automated regression.
4. Run L1 base data plane checks.
5. Run L2 world evolution and continuity checks.
6. Run L3 runtime task delivery checks.
7. Run L4 callback/resume orchestration checks.
8. Run L5 tooling smoke and machine-style scenario checks.
9. Produce a final pass/fail report and issue list.

## Environment Baseline

The full-functional run should explicitly record:

- current git revision
- config file path
- database isolation strategy
- ports used by Engine and any external worker
- auth tokens used for API, callback, and runtime task access
- whether the run uses source-built binaries, packaged binaries, or both

## Stage Breakdown

### S0. Baseline and documentation

Purpose:

- unify the test scope
- define environment rules
- avoid ad hoc one-off checks

Deliverables:

- this full test plan
- one execution checklist
- one result report template

### S1. Dedicated external test worker

Purpose:

- provide a deterministic game-side execution surface
- remove manual callback steps from the critical validation path

Minimum worker capabilities:

- HTTP push receiver
- runtime task pull consumer
- callback completion client
- heartbeat support for long-running task simulation
- fixture-driven response generation
- structured logs for task and callback correlation

### S2. Worker self-validation

Purpose:

- verify the worker before mixing it into Engine scenarios

Minimum checks:

- accepts push payloads
- claims and starts pull tasks
- posts callback successfully
- emits stable structured logs

### S3. Automated regression

Purpose:

- verify source baseline before scenario testing

Minimum checks:

- `go test ./...`
- focused reruns for recently changed runtime task and CLI paths if needed

### S4. Base data plane checks

Scope:

- node CRUD
- component CRUD
- memory CRUD
- relation CRUD
- world settings
- world policy

Validation rule:

- HTTP and DevCli should agree on the observed state

### S5. World evolution and continuity checks

Scope:

- world tick
- timeline latest/list
- continuity state component read/write
- debug continuity
- logs and traces correlation

Validation rule:

- one world tick should produce coherent timeline, state, and observability artifacts

### S6. Runtime task delivery checks

Scope:

- push delivery
- pull delivery
- hybrid fallback
- claim / start / heartbeat / release / requeue
- task inspect / stats / list

Validation rule:

- task state transitions should match the documented lifecycle

### S7. Callback/resume orchestration checks

Scope:

- callback success and failure
- paused execution auto-resume
- `resume_policy = none`
- callback replay protection
- callback post-process `record_only`
- callback post-process `write_memory`
- duplicate data-request suppression after resume

Validation rule:

- callback behavior must match persisted task payload semantics, not only current config state

### S8. Tooling smoke checks

Scope:

- SDK runtime task helpers
- DevCli task and node compatibility paths
- Creator Tasks / Continuity / Traces smoke path

Validation rule:

- tooling views should reflect the same data the HTTP APIs expose

### S9. Machine-style integrated scenario

Scope:

- invoke one NPC dialogue with request-scoped dynamic interfaces
- generate one runtime task
- let the worker complete callback
- confirm Engine resumes paused execution
- inspect logs, traces, and continuity bundle

Validation rule:

- one full invoke -> task -> callback -> resume loop is repeatable in an isolated environment

### S10. Final report

Deliverables:

- pass/fail matrix
- exact reproduction steps for failed cases
- issue severity and priority notes
- known gaps that are operational rather than code defects

## Current Priority Regressions

The first full-functional run must explicitly verify these paths:

- `task inspect` returns populated task details instead of an empty shell
- legacy `nodes --world <world-id>` matches direct `/api/v1/nodes?world_id=...` output
- callback resume does not repeatedly re-emit the same `data_request`
- `POST /api/v1/components` does not fail due to duplicate `world_settings` creation races

## Suggested Artifacts

- `docs/FULL_FUNCTIONAL_TEST_PLAN.md`
- `docs/FULL_FUNCTIONAL_TEST_CHECKLIST.md`
- `docs/FULL_FUNCTIONAL_TEST_REPORT.md`
- one dedicated test worker under `cmd/` or `tools/`

