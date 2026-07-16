# 外部交互总览

本页是当前项目对外部异步交互的主入口文档。

## 1. 能力边界

Engine 当前支持三种对外任务投递模式：

- `push`
- `pull`
- `hybrid`

外部交互统一围绕两类接口组织：

- `external_query`：向游戏侧或外部系统查询权威数据，再继续推理
- `external_action`：请求外部系统执行动作，可只记录结果，也可回填后触发恢复或后处理

## 2. 当前推荐职责分层

- Engine：负责 runtime task 创建、投递治理、callback 回填、恢复编排、统一后处理
- Worker / 游戏侧：负责真实执行任务、维护高频权威状态、回调执行结果
- DevCli / Creator：负责诊断 task、观察 callback / resume 状态与排障
- SDK：负责程序化接入这些 HTTP 契约

## 3. 推荐接入模式

| 场景 | 推荐模式 | 说明 |
|---|---|---|
| Engine 与游戏逻辑服务位于同一受控网络 | `push` | 链路最短，时延最低 |
| 游戏客户端不方便暴露入站服务 | `pull` | 由客户端或 bridge 主动领取任务 |
| 希望优先主动派发，但必须保留失败回退 | `hybrid` | push 失败后回落到 pull 队列 |

## 4. 最小闭环

最小外部交互闭环如下：

1. 调用方通过 `invoke` 发起请求
2. Engine 生成 runtime task
3. Worker 或游戏侧通过 `push` / `pull` 接收任务
4. 外部系统完成查询或动作
5. 外部系统调用 `POST /api/v1/actions/callback`
6. Engine 更新任务状态，并在需要时恢复原执行链路

## 5. 当前已落地的关键能力

- runtime task 队列模型
- `pending / claim / start / heartbeat / release / requeue / stats`
- callback 完成态回写
- paused execution callback 自动恢复
- request-scoped `dynamic_interfaces`
- callback post-process 基础版（`none` / `record_only` / `write_memory`）
- Worker 的 push / pull / callback 闭环与内置测试场景

## 6. 调试顺序建议

当外部交互不符合预期时，按这个顺序排查：

1. `GameAgentDevCli task inspect <task-id>`
2. `GameAgentDevCli task stats`
3. `GameAgentDevCli debug continuity <world-id>`
4. Creator `Tasks` 页面
5. Creator `Continuity` / `Logs` / `Traces` 页面

## 7. 补充材料

以下历史/补充文档已移动到 `docs/internal/`：

- [外部交互路线图](../internal/EXTERNAL_INTERACTION_ROADMAP.md)
- [外部交互接入示例](../internal/EXTERNAL_INTERACTION_EXAMPLES.md)
- [外部交互测试矩阵](../internal/EXTERNAL_INTERACTION_TEST_MATRIX.md)
- [Runtime Task Tooling Sync](../internal/RUNTIME_TASK_TOOLING_SYNC.md)
