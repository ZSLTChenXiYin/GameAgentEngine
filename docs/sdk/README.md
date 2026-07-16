# SDK Overview

本目录用于集中存放 GameAgentEngine 多语言 SDK 的正式文档。

目标是把 SDK 的总体说明、基线规范、共享夹具与能力矩阵从源码目录收回到统一的 `docs/` 文档体系中。

## 当前文档

- [SDK Baseline](./SDK_BASELINE.md)
- [SDK Shared Fixtures](./SDK_FIXTURES.md)
- [SDK Capability Matrix](./SDK_CAPABILITY_MATRIX.md)
- [Language SDK Status](./SDK_LANGUAGES.md)

## 当前语言目录

代码仍位于下列目录，但这些目录应尽量只保留源码、示例与构建文件，不再承载正式文档：

- `tools/source/sdks/ts-sdk`
- `tools/source/sdks/js-sdk`
- `tools/source/sdks/cs-sdk`
- `tools/source/sdks/gd-sdk`
- `tools/source/sdks/cpp-sdk`
- `tools/source/sdks/java-sdk`
- `tools/source/sdks/lua-sdk`
- `tools/source/sdks/c-sdk`

Go SDK 仍然是当前语义基线，位于：

- `sdk/client.go`
- `sdk/types.go`

## 文档约束

- 正式文档统一收口到 `docs/`
- `tools/source/sdks/*` 不再放置分散 README
- 如需补 SDK 说明，优先补到 `docs/sdk/`
