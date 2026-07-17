# 未来开发计划

本文件记录当前清理后的路线图，避免未来实现阶段丢失此前已经形成的规划上下文。

## 规划更新

在内核 / play 重构完成后，当前执行顺序再次调整为：

1. 继续完成文档清理与工作流对齐
2. 重组面向 SDK 的文档与职责边界
3. 在结构清理稳定后，执行打包产物验收

Engine 内核补完与 Worker play 深化，已不再是当前路线图中处于未完成状态的主焦点。

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

- 在结构清理完成后，验证 Engine / DevCli / Worker / Creator 的打包工作流

状态：待开始。

## 已延期但持续跟踪

- world-tick context 路线图，包括 `world_focus`、active-node selection 和 staged scope refinement；见 `docs/internal/WORLD_TICK_CONTEXT_ROADMAP.md`
- autonomous scheduling 路线图，包括 priority、batching、lifecycle state 与 event-driven wake-up；见 `docs/internal/AUTONOMOUS_SCHEDULING_ROADMAP.md`
- roleplay interaction 路线图，包括 direct single-chat、group-chat、interaction-session 建模与 player-intent bridge；见 `docs/internal/ROLEPLAY_INTERACTION_ROADMAP.md`

- 在活跃契约清理之外继续做更广义的文档瘦身
- 如果 play / kernel 稳定后仍然需要，再深化 multi-NPC group-chat 推理
- 面向非 Go 生态的后续 SDK 扩展工作
- 在契约成熟之后再推进 future Engine kernelization work；见 `docs/internal/ENGINE_KERNELIZATION_MEMO.md`
