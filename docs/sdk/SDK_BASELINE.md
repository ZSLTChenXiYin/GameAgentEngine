# SDK 基线规范

**中文** | [**English**](./SDK_BASELINE_EN.md)

本文件定义多语言 SDK 的统一最低交付标准。

## 1. 目标

所有外围语言 SDK 至少应支持当前项目的核心开发闭环：

1. 连接 Engine
2. 创建 / 查询世界与节点
3. 发起 `invoke`
4. 消费 runtime task
5. 回调任务结果
6. 查看状态、时间线、日志和调试信息

## 2. 统一目录结构

每个语言 SDK 目录建议至少包含：

```text
<lang>-sdk/
├── src/ 或等价源码目录
├── examples/
├── workerhome/       # 如果语言生态允许，Worker 共享数据可放在这里
└── package/build metadata
```

## 3. 能力层级

### S1：核心一致层（全部语言必须具备）

- HTTP client / auth / error handling
- health / version
- worlds / nodes / components / memories / relations
- invoke
- runtime task list / pending / claim / start / heartbeat / release / requeue / stats
- callback
- logs / traces
- state components / timelines

### S2：工具增强层（主流语言应具备）

- dynamic interfaces
- continuity bundle
- world policy / world settings
- snapshot / restore / validate / fork
- player interaction typed models

### S3：生态增强层（按语言价值补充）

- 包管理发布元数据
- 面向 Unity / Godot / Node.js / native plugin 的接入示例
- 更完整的 typed enums / constants / helper builders

## 4. 最低对象模型

每个 SDK 至少应提供这些核心对象或等价结构：

- `Node`
- `Component`
- `Relation`
- `Memory`
- `InvokeRequest`
- `InvokeResponse`
- `DynamicInterface`
- `RuntimeTask`
- `RuntimeTaskStats`
- `CallbackResponse`
- `WorldSettings`
- `StateComponent`
- `Timeline`
- `InferenceLog`
- `DebugTrace`

## 5. 最低方法能力矩阵

### 基础连接

- `health`
- `getVersion`

### 世界与节点

- `getWorlds`
- `createNode`
- `getNodes`
- `getNode`
- `updateNode`
- `deleteNode`

### 组件 / 记忆 / 关系

- `addComponent` / `getComponents` / `updateComponent` / `deleteComponent`
- `addMemory` / `getMemories` / `updateMemory` / `deleteMemory`
- `createRelation` / `getRelations` / `updateRelation` / `deleteRelation`

### 推理

- `invoke`
- `interpretPlayerInput`（主流语言优先）

### 世界运行时

- `advanceTick`
- `getWorldSettings`
- `setWorldSettings`
- `getWorldPolicy`
- `setWorldPolicy`

### 调试与观测

- `getLogs`
- `getDebugTraces`
- `getStateComponents`
- `getStateComponent`
- `putStateComponent`
- `getTimelines`
- `getLatestTimeline`
- `getContinuityBundle`（主流语言优先）

### 外部任务

- `listRuntimeTasks`
- `listPendingRuntimeTasks`
- `getRuntimeTask`
- `claimRuntimeTask`
- `startRuntimeTask`
- `heartbeatRuntimeTask`
- `releaseRuntimeTask`
- `requeueRuntimeTask`
- `getRuntimeTaskStats`
- `actionCallback`

### 快照与复制

- `forkWorld`
- `createWorldSnapshot`
- `restoreWorld`
- `validateWorldSnapshot`

## 6. 最低示例要求

每个 SDK 至少提供这些 `examples/`：

1. `health`：连通性检查
2. `world_bootstrap`：创建世界与基础节点
3. `invoke_dialogue`：发起一次 NPC 对话
4. `task_pull_once`：拉取并处理一个 runtime task
5. `callback_complete`：回填一次 callback
6. `continuity_inspect`：读取 state / timeline / logs

## 7. 与 Worker 的对接要求

主流语言 SDK 还应提供最小联调样例：

- 与 `GameAgentWorker serve` 配合的 pull / callback 示例
- 与 `GameAgentWorker play` 相关的 invoke / authority query 示例

对于原生侧或脚本侧的 request-builder SDK，如果当前不内置 HTTP 传输层，上述示例至少也应给出完整的请求构造序列与推荐的 Worker 配合方式，避免外围接入方自行猜测接口顺序。

## 8. 错误模型要求

每个 SDK 至少要统一处理：

- 非 2xx HTTP 错误
- API 返回的结构化错误消息
- JSON 反序列化失败
- 网络错误 / 超时

## 9. 验收要求

一个语言 SDK 只有在满足以下条件后才算达标：

1. 文档可独立指导接入
2. 示例可以覆盖最小闭环
3. 核心能力矩阵至少覆盖 S1
4. 与当前 Engine / Worker 工作流术语一致

## 10. 与 Go SDK 的关系

Go SDK 是当前能力和命名的语义基线：

- `sdk/client.go`
- `sdk/types.go`

其他语言 SDK 在命名上可以遵循各语言习惯，但不应擅自改变能力边界和对象语义。
