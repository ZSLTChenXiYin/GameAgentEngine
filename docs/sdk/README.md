# SDK 文档总览

**中文** | [**English**](./README_EN.md)

本目录集中存放 GameAgentEngine 多语言 SDK 的正式文档。

目标是把 SDK 的总体说明、基线规范、共享夹具与能力矩阵，从源码目录回收到统一的 `docs/` 文档体系中。

## 当前文档

- [SDK 基线规范](./SDK_BASELINE.md)
- [SDK 共享夹具与输入](./SDK_FIXTURES.md)
- [SDK 能力矩阵](./SDK_CAPABILITY_MATRIX.md)
- [各语言 SDK 状态](./SDK_LANGUAGES.md)

## 当前语言目录

代码仍位于下列目录，但这些目录应尽量只保留源码、示例与构建文件，不再承载正式说明文档：

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

新增 TypeScript SDK：

- `sdk/typescript/src/client.ts`
- `sdk/typescript/src/types.ts`
- `sdk/typescript/src/interaction.ts`
- `sdk/typescript/src/index.ts`

可通过 `cd sdk/typescript && npm run build` 构建。

## 文档约束

- 正式 SDK 文档统一收口到 `docs/sdk/`
- `tools/source/sdks/*` 不再放置分散 README
- 如需补 SDK 说明，优先补到 `docs/sdk/`
