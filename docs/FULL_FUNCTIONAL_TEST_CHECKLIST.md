# Full Functional Test Checklist

Use this checklist to execute the staged full-functional validation plan.

## S0 Baseline

- [x] record current git revision
- [x] record config file path
- [x] record database isolation strategy
- [x] verify Engine port availability
- [x] verify push receiver port availability
- [x] verify API key, callback token, and runtime task token

## S1-S2 Test Worker

- [x] build `cmd/gameagentworker`
- [x] run worker unit tests
- [x] verify push receiver startup
- [x] verify pull-once command startup
- [x] verify fixture-driven success path
- [x] verify forced failure path
- [x] verify long-task heartbeat path

## S3 Automated Regression

- [x] run `go test ./...`
- [x] re-run targeted runtime task / CLI packages if needed

## S4 Base Data Plane

- [x] node CRUD via HTTP
- [x] node CRUD via DevCli
- [x] component CRUD via HTTP
- [x] component CRUD via DevCli
- [x] memory CRUD via HTTP
- [x] memory CRUD via DevCli
- [x] relation CRUD via HTTP
- [x] relation CRUD via DevCli
- [x] world settings get/set
- [x] world policy get/set

## S5 World Evolution and Continuity

- [x] configure world time settings
- [x] execute one world tick
- [x] inspect timeline latest/list
- [x] inspect state list/get
- [x] inspect debug continuity
- [x] inspect logs and traces for one request_id

## S6 Runtime Task Delivery

- [x] push delivery success
- [x] pull delivery success
- [x] hybrid fallback transition
- [x] task claim/start/heartbeat
- [x] task release/requeue
- [x] task list/stats/inspect

## S7 Callback / Resume Orchestration

- [x] callback success path
- [x] callback failure path
- [x] paused execution auto-resume
- [x] `resume_policy = none`
- [x] callback replay protection
- [x] callback post-process `record_only`
- [x] callback post-process `write_memory`
- [x] duplicate `data_request` suppression after resume

## S8 Tooling Smoke

- [x] SDK runtime task helper smoke
- [x] DevCli node/task compatibility smoke
- [x] Creator Tasks page smoke
- [x] Creator Continuity page smoke
- [x] Creator Traces page smoke

## S9 Machine Scenario

- [x] isolated Engine runtime started
- [x] test worker started
- [x] NPC dialogue invoke with request-scoped dynamic interfaces
- [x] runtime task created
- [x] callback completed by worker
- [x] paused execution resumed
- [x] logs / traces / continuity confirmed

## Priority Regression Verification

- [x] `task inspect` shows populated fields
- [x] `nodes --world` matches direct HTTP output
- [x] callback resume does not re-emit duplicate `data_request`
- [x] `POST /api/v1/components` avoids duplicate `world_settings` write failure
