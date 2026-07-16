# GameAgentEngine SDKs

本目录存放面向外围开发者的多语言 SDK。

当前目标不是让每个语言 SDK 一开始就做到与 Go SDK 完全同等的工程深度，而是先统一建立可持续扩展的 baseline，确保所有外围语言都能接入当前项目的核心工作流：

- 调用 Engine HTTP API
- 发起 `invoke`
- 消费 runtime task
- 回调 `callback`
- 查询日志、轨迹、状态组件与时间线
- 与 `GameAgentWorker`、`GameAgentDevCli`、`GameAgentCreator` 形成完整闭环

## 当前语言目录

- `ts-sdk`
- `js-sdk`
- `cs-sdk`
- `gd-sdk`
- `cpp-sdk`
- `java-sdk`
- `lua-sdk`
- `c-sdk`

## 优先级

### 第一优先级

- `ts-sdk`
- `js-sdk`
- `cs-sdk`
- `gd-sdk`

### 第二优先级

- `cpp-sdk`
- `java-sdk`

### 第三优先级

- `lua-sdk`
- `c-sdk`

## 统一规范

每个语言 SDK 的最低要求由 [SDK_BASELINE.md](./SDK_BASELINE.md) 定义。

## Go SDK 作为语义基线

当前仓库中的 Go SDK 是能力基线，位于：

- `sdk/client.go`
- `sdk/types.go`

其它语言 SDK 不要求逐行复制 Go SDK 实现，但应在对外能力、对象模型和接入闭环上尽量保持一致。


## Shared Integration Inputs

- [SDK_FIXTURES.md](./SDK_FIXTURES.md) (shared fixture files and sample-data reuse rules)
- [SDK_CAPABILITY_MATRIX.md](./SDK_CAPABILITY_MATRIX.md) (current per-language implementation status and coverage matrix)
