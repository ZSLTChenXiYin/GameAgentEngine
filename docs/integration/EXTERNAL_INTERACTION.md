# 外部交互总览

**中文** | [**English**](./EXTERNAL_INTERACTION_EN.md)

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
| --- | --- | --- |
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

本页现在是外部交互工作流的规范主入口。

仅当实现细节仍构成活跃契约或未完成未来路线图时，才把实现性笔记继续保留在 `docs/internal/`。

## 8. 推荐接入模式扩展

当前外部交互基线可视为三种稳定模式：

- push：Engine 通过已配置的适配器派发，游戏侧通过 callback 完成
- pull：游戏侧或 bridge 领取 runtime task、执行，再通过 callback 报告完成
- hybrid：Engine 优先 push，派发失败后回落到 pull 队列消费

当前边界注意事项：

- `fallback_transport` 目前表示回落到 pull 队列消费，而不是自动切换到另一种 push 适配器
- `max_attempts` 约束的是 pull / hybrid 的领取重试行为，而不是完整死信系统
- `record_only`、`write_memory` 等 callback post-process 行为应被视为任务快照行为

## 9. 当前自动化覆盖边界

当前自动化外部交互基线覆盖：

- push 派发状态迁移和可观测字段
- pull 队列 claim / start / heartbeat / release / requeue 路径
- hybrid push 失败后回落到已释放的 pull 任务
- callback 完成、暂停执行自动恢复，以及 `resume_policy = none` 行为
- heartbeat 超时标记、自动 requeue 快照策略、重试耗尽和重复超时诊断
- 基于 request-id 的 callback 重放保护

仍待增强的部分：

- 按 `consumer` 或 `category` 做更细的治理策略
- 更丰富的多阶段 hybrid 回退状态机
- 超出 request-id 占用之外的更强 callback 重放保护
- 在诊断视图之上增加批量人工干预流程
