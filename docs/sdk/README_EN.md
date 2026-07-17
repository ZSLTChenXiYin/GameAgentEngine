# SDK Overview

This directory centralizes the formal documentation for the GameAgentEngine multi-language SDK set.

The goal is to pull SDK overview material, baseline requirements, shared fixtures, and capability matrices back into the unified `docs/` documentation tree instead of leaving them scattered across source directories.

## Current Documents

- [SDK Baseline Specification](./SDK_BASELINE_EN.md)
- [SDK Shared Fixtures and Inputs](./SDK_FIXTURES_EN.md)
- [SDK Capability Matrix](./SDK_CAPABILITY_MATRIX_EN.md)
- [Language SDK Status](./SDK_LANGUAGES_EN.md)

## Current Language Directories

The code still lives in the following directories, but those locations should now contain source code, examples, and build metadata only rather than formal documentation:

- `tools/source/sdks/ts-sdk`
- `tools/source/sdks/js-sdk`
- `tools/source/sdks/cs-sdk`
- `tools/source/sdks/gd-sdk`
- `tools/source/sdks/cpp-sdk`
- `tools/source/sdks/java-sdk`
- `tools/source/sdks/lua-sdk`
- `tools/source/sdks/c-sdk`

The Go SDK remains the semantic baseline and lives at:

- `sdk/client.go`
- `sdk/types.go`

## Documentation Rules

- formal SDK documentation should live under `docs/sdk/`
- `tools/source/sdks/*` should no longer accumulate scattered README files
- when SDK-facing documentation needs to be added, add it to `docs/sdk/` first
