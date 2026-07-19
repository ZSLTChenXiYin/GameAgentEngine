# 未来开发计划

**中文** | [**English**](./FUTURE_DEVELOPMENT_PLAN_EN.md)

本文件记录当前清理后的路线图，避免未来实现阶段丢失此前已经形成的规划上下文。

## 规划更新

当前未完成事项的执行顺序调整为：

1. 优先解决 Creator 在大规模节点树下的性能问题
2. 在 Creator 可用性稳定后，再继续推进剩余 Engine 路线图事项
3. 其后再处理更广义的文档瘦身与后续扩展事项

Engine 内核补完与 Worker play 深化已经完成；当前最需要优先修复的是 Creator 左侧节点树在上万节点规模下的响应退化问题，因为它直接影响世界编辑与调试效率。

## F0 Creator 大规模节点树性能优化

- 目标是在万级以上节点规模下，保持 Creator 左侧层级树的可编辑性、可搜索性与可滚动性
- 不再接受“每次点击 / 搜索 / 折叠都整棵树重建”的渲染模式
- 先解决可见区域渲染与局部更新问题，再考虑更细的交互增强

### F0 checklist

1. 建立 1k / 5k / 10k 节点基线 profiling，量化首屏、滚动、展开、折叠、过滤、选中耗时
2. 拆分树数据准备与 DOM 渲染，缓存 `nodeMap`、`childMap`、可见行列表等中间结构
3. 将当前递归整树渲染改为可见行拍平模型
4. 为左侧节点树加入虚拟滚动，只渲染视口内行
5. 将展开 / 折叠 / 选中 / 拖拽改为局部刷新，避免整树 `renderTree()`
6. 将逐节点事件绑定改为容器级事件委托
7. 为名称 / 类型过滤加入索引化或增量过滤路径
8. 为超大树加入降级策略，例如默认折叠、按需展开、优先搜索定位
9. 补齐大树场景的验收标准与回归样例，纳入后续 Creator 回归基线

状态：已完成。

## F1 Engine 核心净化

- 移除已经有明确归属方的 Engine 侧开发者工具命令
- 让 Engine 聚焦服务运行时与版本元数据
- 在 DevCli 侧 coverage 存在后，再决定 validate 的最终去留

### F1 checklist

1. 当 DevCli 成为唯一入口后，移除 Engine `creator`
2. 当 DevCli / Creator 路径稳定后，移除 Engine `import`
3. 升级 DevCli `init`，然后移除 Engine `init`
4. 对比 Engine `test` 与 Worker `test`，然后做移除或合并
5. 迁移或删除 Engine `validate`
6. 让 Engine 最终只保留 `serve` 和 `version`，过渡期内可暂时保留 `validate`

状态：已完成。

## F2 Worker CLI 重组

- 将 Cobra 命令定义从 `internal/workercli` 移到 `cmd/gameagentworker`
- 可复用业务逻辑只保留在 internal 包中

状态：已完成。

## F3 Engine 内核补完

- 完成内核侧 interaction 模型，而不是继续把它散落在 Worker 独有的流程胶水里
- 在保持 Engine 适合嵌入的前提下，让 actor / target / scene / participant 语义成为一等公民
- 完成玩家意图解释、权威查询语义与 interaction 恢复流程的内核侧契约

### F3 checklist

1. 在 `invoke` 之上完成 interaction API 语义
2. 将 `interaction/*`、`player/input/interpret` 与普通 `invoke` 对齐为一个一致的内核契约
3. 稳定 direct dialogue 与 group chat 下的 actor / target / scene / participant 建模
4. 收紧内核侧的 player intent schema、校验词汇与响应契约
5. 验证 Engine 侧 authority-query / runtime-task / callback-resume 行为仍与目标嵌入边界一致
6. 减少那些本应属于 Engine 契约、却临时落在 Worker 的 ad-hoc glue

状态：已完成。

## F4 Worker play 深化

- 把 play 演进为真正的文字游戏外壳，而不是薄薄一层 engine wrapper
- 基于已经稳定的 Engine interaction 内核继续建设，而不是继续补偿缺失的内核语义
- 在优先级上，play 仍先于文档润色，但晚于内核补完

### F4 checklist

1. 改进 play 命令语义与 turn flow
2. 深化 room 反馈、目标切换与 interaction 呈现
3. 在内核侧 intent 契约稳定后，提升 `/act` bridge 质量
4. 重新评估 group-chat 行为，决定是继续保留 one-primary-responder 还是扩展

状态：已完成。

## F5 文档集中化与瘦身

- 删除 `tools/source/docs`
- `tools/source` 只保留打包运行时资产
- 将 SDK overview / baseline / capability 文档迁移到 `docs/`
- 迁移完成后，删除命令目录和 SDK 目录下分散的 README
- 将正式文档文件名统一为大写命名风格
- 要求每一页正式文档都具备中英文双版本

### F5 checklist

1. 继续移除过时的内部 rollout 与 sync 文档
2. 清理不再描述当前工作流、价值较低的历史文档
3. 只保留仍在定义活跃契约、边界或设计理由的文档
4. 确保 `docs/` 仍然是唯一正式文档树
5. 将存活下来的正式文档文件名统一为大写命名
6. 为每一页存活的正式文档补齐中英文对应版本

状态：已完成。

## F6 tests 收敛

- 保持 `tools/source/workerhome/fixtures` 为纯数据目录
- 将剩余过程型流程迁移到 Worker 命令中

状态：对当前以 worker 驱动的测试工作流基线而言，已完成。

## F7 SDK 文档与职责重组

- 将 SDK 文档集中到 `docs/`
- 让 SDK 目录只保留代码与示例
- 继续让 SDK 的对外职责与 Go SDK 基线对齐
- 对面向 SDK 的正式文档同样施加大写命名与双语规则

状态：已完成。

## F8 打包产物验收

- 以本地发布包为基准，验证 Engine / DevCli / Worker / Creator 的打包工作流
- 先完成包结构、基础运行链路、Creator、Go SDK smoke 和文档对齐的本地验收，不把 GitHub 链路纳入当前完成条件

### F8 checklist

1. 验证 6 个目标平台的包结构与 zip 产物
2. 验证包内 Engine 可启动，DevCli 可连通
3. 验证包内 Worker 的 tooling-smoke 可完成
4. 验证一个基准 Go SDK smoke 场景
5. 对齐 README / docs 与发布包路径

状态：已完成。

## 已延期但持续跟踪

- Creator 大规模节点树性能路线图，见 `docs/internal/CREATOR_TREE_PERFORMANCE_ROADMAP.md`

- 世界建模、运行基座、权威动态状态与 world tick bootstrap 的推荐约定，见 `docs/architecture/WORLD_MODELING_AND_RUNTIME_CONVENTIONS.md`
- 当前基于真实 world tick 收束问题整理出的 Engine 改进路线图，见 `docs/internal/ENGINE_IMPROVEMENT_ROADMAP.md`

- world-tick context 路线图，包括 `world_focus`、active-node selection 和 staged scope refinement；见 `docs/internal/WORLD_TICK_CONTEXT_ROADMAP.md`
- autonomous scheduling 路线图，包括 priority、batching、lifecycle state 与 event-driven wake-up；见 `docs/internal/AUTONOMOUS_SCHEDULING_ROADMAP.md`
- roleplay interaction 路线图，包括 direct single-chat、group-chat、interaction-session 建模与 player-intent bridge；见 `docs/internal/ROLEPLAY_INTERACTION_ROADMAP.md`

- 在活跃契约清理之外继续做更广义的文档瘦身
- 如果 play / kernel 稳定后仍然需要，再深化 multi-NPC group-chat 推理
- 面向非 Go 生态的后续 SDK 扩展工作
- 在契约成熟之后再推进 future Engine kernelization work；见 `docs/internal/ENGINE_KERNELIZATION_MEMO.md`


## 下一阶段: 可选方向实施

以下为各路线图文档中整理出的剩余可选方向的整合实施计划。此计划从 Engine 核心出发，逐步向外扩展。

### 阶段分区

| 阶段 | 代号 | 聚焦 |
|---|---|---|
| Phase 1 | Engine Core | world_focus、候选节点评分、事件驱动唤醒、interaction session |
| Phase 2 | Worker & Service | 房间 authority、world-tick 整合、scope 精细化 |
| Phase 3 | SDK & Bench | TypeScript SDK、压力测试、文档同步 |

### AS.2: 自治行为生命周期状态跟踪 [DONE]

- [x] 在 `AutonomousConfig` 中增加 `Status` 字段
- [x] 在执行前设置 `running` 状态并持久化
- [x] 在执行成功后设置 `completed` 状态并持久化
- [x] 提交: `265dc6e`

### Phase 1: Engine Core 增强

#### P1.1 (WTC.1): world_focus 组件契约 [PENDING]

为 Engine 增加 `world_focus` 组件，使 world tick 可以显式地将特定子节点提升到推理上下文中。

实施:
1. 在 `types.go` 中定义 `WorldFocusConfig` 结构体（enabled、tasks、priority、max_parent_distance 等字段）
2. 在 `store` 中注册 `CompWorldFocus` 组件类型
3. 在 context builder 中实现 world_focus 扫描逻辑
4. 在 world tick prompt 中注入被提升的观察节点摘要
5. 设置硬限制：扫描深度 ≤ 5，选中节点数 ≤ 10

验收:
- world_focus 组件可正常创建/读取/更新
- world tick context 中包含被提升的子节点摘要
- 超限节点被正确截断且不报错

#### P1.2 (WTC.2): 候选节点选择与评分 [PENDING]

为 world tick 增加基于多信号的候选节点评分系统。

实施:
1. 定义评分信号接口（结构紧密度、近期变更、事件影响、玩家交互、autonomous 到期状态）
2. 实现评分函数，各信号加权汇总
3. 在 world tick context 构建阶段调用评分，产出 top-K 候选节点列表
4. 候选列表作为额外观察上下文注入 world tick prompt

验收:
- 有活跃变更的节点获得更高评分
- 候选列表长度不超过 K（默认 10）
- 评分不影响现有 world tick 行为

#### P1.3 (AS.3): 事件驱动唤醒队列 [PENDING]

为 autonomous 调度增加事件驱动唤醒路径。

实施:
1. 在 `internal/service/` 中增加 `AutonomousWakeQueue`
2. 定义 `WakeEvent`（node_id、reason、priority、timestamp）
3. 暴露 `EnqueueWake(nodeID, reason)` 和 `PendingWakeEvents(worldID)` 接口
4. 在 autonomous 调度扫描时先消费唤醒队列
5. 在对话完成、action 执行、world tick 事件生成时触发 enqueue

验收:
- 唤醒事件可正常入队和消费
- 消费后的节点优先于普通扫描被调度
- 队列支持按 world 隔离

#### P1.4 (RI.1): Interaction Session 标准化 [PENDING]

稳定 interaction 中 actor/target/scene/room 的上下文模型。

实施:
1. 定义 `InteractionContext` 结构体（ActorID、TargetID、SceneID、RoomID、TurnIndex、EventType）
2. 在 pipeline invoke 时传递 InteractionContext
3. Worker play direct-chat 和 room-chat 使用同一套语义
4. 统一 invoke 和 play 的交互上下文构建路径

验收:
- interaction invoke 输出包含一致的 actor/target/scene 标识
- Worker play 正确消费 interaction context

### Phase 2: Worker & Service 深化

#### P2.1 (RI.2): 房间 Authority 所有权模型 [PENDING]

改进房间级的 authority 所有权和参与者可见性规则。

实施:
1. 定义 room-level authority owner 规则
2. 改进参与者 visibility 规则
3. 在 Worker play room-chat 中加入参与者摘要注入

#### P2.2 (AS.4): 唤醒原因与 World-Tick 整合 [PENDING]

将 event-driven wake 与 world-tick 输出和 interaction 流整合。

实施:
1. 在 world tick 输出中记录被唤醒的 autonomous 节点
2. 在 interaction 完成时自动触发临近 autonomous 节点的唤醒
3. 建立唤醒来源跟踪：world_tick / interaction / external

#### P2.3 (WTC.3): 分阶段 Scope 精细化 [PENDING]

支持 world tick 先粗粒度摘要、再选择子 scope 深入。

实施:
1. 在 world tick 中增加粗摘要阶段
2. 从粗摘要中选择变更显著的子 scope
3. 对被选中的子 scope 进行第二轮深入
4. 设置轮次预算控制

### Phase 3: SDK & 测试扩展

#### P3.1 (SDK.1): TypeScript SDK [PENDING]

为 Engine API 提供 TypeScript 客户端 SDK。

实施:
1. 在 `sdk/` 下创建 `typescript/` 目录
2. 封装 REST API 客户端（world、node、component、invoke、autonomous 等资源）
3. 提供 TypeScript 类型定义
4. 编写 README 和示例

#### P3.2 (BENCH.1): 压力测试与基准测试 [PENDING]

为 Engine 核心路径建立基准测试。

实施:
1. 创建 `internal/bench/` 或 `test/bench/`
2. 为以下路径建立 Go 基准测试：world tick、invoke、data request、autonomous scheduling
3. 记录基准线并输出到 CI
4. 提供 `go test -bench=. ./internal/bench/` 运行方式

---

P1.1-P1.4 优先级相同，按 WTC.1 → WTC.2 → AS.3 → RI.1 顺序执行。P2.x 在 Phase 1 完成后开始。P3.x 为扩展阶段，可在任意空闲时间穿插。
