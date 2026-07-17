# GitHub Release Automation Plan

This document captures the post-v0.5.0 GitHub automation plan for GameAgentEngine. At this stage it is design-only and does not add workflow files yet.

## 1. Goal

The future GitHub automation should be split into two layers:

1. CI build validation
2. Tag-driven formal release packaging and publishing

The plan assumes the packaged artifact layout has already stabilized around:

- `gameagentengine.conf.yaml`
- `workerhome/demo/`
- `workerhome/fixtures/`
- `sdks/`
- `web/GameAgentCreator/`

## 2. Recommended Workflow Split

### 2.1 ci.yml

Triggers:

- push to the main branch
- pull requests

Responsibilities:

- Go build
- baseline tests
- packaging smoke check
- packaged artifact structure validation

Suggested checks:

- `go build ./...`
- required unit tests
- run the packaging script for at least one primary platform
- verify the packaged output contains:
  - `gameagentengine.conf.yaml`
  - `workerhome/demo/demo-world.yaml`
  - `workerhome/demo/demo-state.yaml`
  - `workerhome/fixtures/runtime_task_dynamic_interfaces.json`
  - `web/GameAgentCreator/index.html`

### 2.2 release.yml

Triggers:

- push tags such as `v0.5.0`

Responsibilities:

- multi-platform matrix builds
- run packaging scripts
- generate zip artifacts
- create GitHub Releases
- upload release assets

Suggested target matrix:

- windows/amd64
- windows/arm64
- linux/amd64
- linux/arm64
- darwin/amd64
- darwin/arm64

## 3. Release Preconditions

Before adding GitHub release automation, the repository should first satisfy these conditions:

1. versioning has one formal primary source and can be injected by packaging scripts
2. the packaged directory layout is stable
3. Worker demo and fixture directories are stable
4. packaging docs match the real artifact layout
5. no remaining Demo-only workflow logic depends on local path guessing

## 4. Recommended Execution Flow

### CI Layer

1. checkout
2. setup Go
3. build all
4. run tests
5. run one packaging smoke build
6. validate the packaged asset tree

### Release Layer

1. checkout the tag
2. setup Go
3. run matrix builds by platform
4. execute packaging scripts
5. archive outputs
6. create the GitHub Release
7. upload the archives

## 5. Explicitly Out of Scope for Now

This stage does not yet add:

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- automated GitHub Release notes generation
- Gitee or other mirror-release synchronization

Those will be implemented later as a separate task.

## 6. Suggested Later Implementation Order

When implementation starts, the recommended order is:

1. add `ci.yml` first
2. stabilize it, then add `release.yml`
3. only after that add release notes, checksums, and mirror publishing enhancements
