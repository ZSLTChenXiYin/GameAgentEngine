# Full Functional Test Report

## Run Metadata

- git revision: `0885daaff08de89f5b0049d0bdaf012927e8e4ce`
- config file:
- database isolation:
- Engine port:
- worker port:
- API key source:
- callback token source:
- runtime task token source:
- runtime mode: source-built / packaged / mixed
- execution date:

## Stage Summary

| Stage | Status | Notes |
|---|---|---|
| S0 Baseline | completed | plan document added |
| S1 Test worker | completed | `cmd/gameagenttestworker` added |
| S2 Worker self-validation | completed | worker unit tests and build passed |
| S3 Automated regression | completed | `go test ./...` passed |
| S4 Base data plane | pending | |
| S5 World evolution and continuity | pending | |
| S6 Runtime task delivery | pending | |
| S7 Callback/resume orchestration | pending | |
| S8 Tooling smoke | pending | |
| S9 Machine scenario | pending | |
| S10 Final report | pending | |

## Automated Regression

- command: `go test ./...`
- result: passed
- failures: none

## Priority Regression Verification

| Item | Status | Evidence |
|---|---|---|
| `task inspect` populated fields | pending | |
| `nodes --world` matches direct HTTP | pending | |
| callback resume avoids duplicate `data_request` | pending | |
| `POST /api/v1/components` avoids duplicate `world_settings` race | pending | |

## Base Data Plane Results

| Area | HTTP | DevCli | Notes |
|---|---|---|---|
| Node CRUD | pending | pending | |
| Component CRUD | pending | pending | |
| Memory CRUD | pending | pending | |
| Relation CRUD | pending | pending | |
| World settings | pending | pending | |
| World policy | pending | pending | |

## World Evolution and Continuity Results

- world tick:
- timeline latest/list:
- state list/get:
- debug continuity:
- logs/traces correlation:

## Runtime Task Delivery Results

- push:
- pull:
- hybrid fallback:
- claim/start/heartbeat:
- release/requeue:
- stats/inspect:

## Callback and Resume Results

- callback success:
- callback failure:
- paused execution auto-resume:
- `resume_policy = none`:
- replay protection:
- `record_only`:
- `write_memory`:
- duplicate query suppression:

## Tooling Smoke Results

- SDK:
- DevCli:
- Creator Tasks:
- Creator Continuity:
- Creator Traces:

## Machine-Style Scenario Results

- invoke:
- runtime task creation:
- worker callback:
- paused execution resume:
- observability artifacts:

## Failures and Follow-Ups

| Severity | Area | Symptom | Reproduction | Notes |
|---|---|---|---|---|

## Final Assessment

- overall status:
- blocking issues:
- non-blocking issues:
- operational gaps:
