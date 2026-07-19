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
