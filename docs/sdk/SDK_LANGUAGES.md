# 各语言 SDK 状态

本文档集中记录各语言 SDK 的定位、当前覆盖范围与附带示例，替代原先散落在 `tools/source/sdks/*/README.md` 下的说明文件。

## TypeScript SDK

- 目录：`tools/source/sdks/ts-sdk`
- 定位：当前最完整的外围脚本 / 工具链基线之一
- 当前覆盖：
  - health / version
  - invoke
  - player input interpretation
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback
  - world settings
  - continuity state components
  - timelines
  - logs / debug traces
  - world tick advance
- 目录结构重点：
  - `src/client.ts`
  - `src/types.ts`
  - `src/index.ts`

## JavaScript SDK

- 目录：`tools/source/sdks/js-sdk`
- 定位：轻量 Node.js 工具与桥接进程基线
- 当前覆盖：
  - health / version
  - invoke
  - player input interpretation
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback
  - world settings
  - continuity state components
  - timelines
  - logs / debug traces
  - world tick advance
- 备注：默认依赖 Node.js 18+ 的 `fetch`

## C# SDK

- 目录：`tools/source/sdks/cs-sdk`
- 定位：服务端 C# / Unity 邻接集成基线
- 当前覆盖：
  - health / version
  - invoke
  - player input interpretation
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats
  - callback
  - world settings
  - continuity state components
  - timelines
  - logs / debug traces
  - world tick advance
- 目录结构重点：
  - `src/GameAgentEngineClient.cs`
  - `src/Models.cs`
  - `GameAgentEngine.SDK.csproj`

## GDScript SDK

- 目录：`tools/source/sdks/gd-sdk`
- 定位：Godot 侧轻量接入基线
- 当前覆盖：
  - health / version 请求构造
  - invoke / player input interpretation 请求构造
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats 请求构造
  - callback 请求构造
  - world settings / state components / timelines / logs / debug traces 请求构造
  - world tick advance 请求构造
- 附带示例：worker authority query / runtime roundtrip
- 备注：返回 Godot 友好的请求字典，由项目自行接 HTTP 层

## Java SDK

- 目录：`tools/source/sdks/java-sdk`
- 定位：Java 服务端或中间层集成基线
- 当前覆盖：
  - health / version
  - invoke
  - player input interpretation
  - pending runtime task list
  - runtime task list / get / claim / start / heartbeat / release / requeue / stats
  - callback completion
  - authority query 与 worker roundtrip 示例
- 现状：已可直接覆盖当前 Engine / Worker 最小联调闭环，但状态、时间线、日志与更完整 typed model 仍弱于 TS / JS / C#

## C++ SDK

- 目录：`tools/source/sdks/cpp-sdk`
- 定位：原生侧集成的请求构造基线
- 当前覆盖：
  - health / version / invoke / player input interpret 请求构造
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats 请求构造
  - callback 请求构造
  - worker authority query / runtime roundtrip 示例
- 现状：不内置完整 HTTP 传输层，但已给出与 Worker 对接所需的最小请求序列

## C SDK

- 目录：`tools/source/sdks/c-sdk`
- 定位：最低依赖的原生侧集成基线
- 当前覆盖：
  - health / version / invoke / player input interpret 路径辅助
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats 路径 / payload 辅助
  - callback payload 构造
  - worker authority query / runtime roundtrip 示例
- 现状：不内置 HTTP 传输层，但已覆盖与 Worker 联调的基础请求拼装

## Lua SDK

- 目录：`tools/source/sdks/lua-sdk`
- 定位：轻量脚本侧集成基线
- 当前覆盖：
  - health / version / invoke / player input interpret path 和 request helper
  - runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats helper
  - callback payload / request helper
  - worker authority query / runtime roundtrip 示例
- 现状：不内置 HTTP 传输层，但请求构造层已能直接承接 Worker 联调顺序

## 统一要求

- Go SDK 仍是语义基线：`sdk/client.go`、`sdk/types.go`
- 所有语言 SDK 的能力命名与接口语义应尽量对齐 Go SDK
- 新增或补齐说明时，统一补到 `docs/sdk/`，不要回写到 `tools/source/sdks/*/README.md`
