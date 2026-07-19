# GameAgentEngine 开发计划基准

**中文** | [**English**](./DEVELOPMENT_PLAN_EN.md)

本文档是 GameAgentEngine 项目的权威开发计划基准。它整合了现有的路线图文档、代码分析发现和已识别的改进缺口，按优先级排列，每项附有详细的实施方案。

---

## 1. 规划总览

### 完成状态图例

| 符号 | 含义 |
|---|---|
| [ ] | 未开始 |
| [->] | 进行中 |
| [x] | 已完成 |

### 优先级总表

| 优先级 | 编号 | 改进项 | 状态 | 参考文档 |
|---|---:|---|:---:|---|
| P0 | F0 | Creator 大规模节点树性能优化 | [x] | CREATOR_TREE_PERFORMANCE_ROADMAP.md |
| P0 | E1 | World Tick Bootstrap | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P0 | E2 | 收束控制 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P0 | E3 | World Tick Prompt 收束重写 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P1 | E4 | 查询结果摘要化 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P1 | E5 | Round Context 压缩 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P1 | E6 | 世界冷启动接口 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P1 | E7 | Demo Authority 接入 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P1 | E8 | Store 数据层覆盖测试加固 | [x] | 本文档 E8 |
| P1 | E9 | PipelineMode 关系装配策略硬化 | [x] | 本文档 E9 |
| P2 | E10 | Callback Bootstrap 兜底 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P2 | E11 | 动作系统 Schema 验证与扩展 | [x] | 本文档 E11 |
| P2 | E12 | PolicyEngine 冲突解析 | [x] | 本文档 E12 |
| P2 | E13 | 前端交互响应优化 | [x] | 本文档 E13 |
| P2 | E14 | Autonomous 调度路线图 | [x] | AUTONOMOUS_SCHEDULING_ROADMAP.md |
| P2 | E15 | World Tick Context 路线图 | [x] | WORLD_TICK_CONTEXT_ROADMAP.md |
| P3 | E16 | SDK 文档与示例完善 | [x] | 本文档 E16 |
| P3 | E17 | 组件系统通用化 | [x] | 本文档 E17 |
| P3 | E18 | Telemetry 可观测性增强 | [x] | 本文档 E18 |
| P3 | E19 | 工具链工作流接入 | [x] | ENGINE_IMPROVEMENT_ROADMAP.md |
| P3 | E20 | 前端 i18n 补齐 | [x] | 本文档 E20 |
| P4 | E21 | 多世界隔离边界明确 | [x] | 本文档 E21 |
| P4 | E22 | 压力测试与 Benchmark | [x] | 本文档 E22 |
| P4 | E23 | Engine Kernelization | [x] | ENGINE_KERNELIZATION_MEMO.md |
| P4 | E24 | 多语言 SDK 扩展 | [x] | 本文档 E24 |

### 优先级定义

| 级别 | 定义 |
|---|---|
| P0 | 当前开发周期必须完成，阻碍其他工作的前提性任务 |
| P1 | 高价值改进，应在 P0 完成后立即开始 |
| P2 | 重要但非阻塞性改进，可在资源允许时安排 |
| P3 | 有价值的增强，无时间压力 |
| P4 | 长期规划，需要进一步评估后再启动 |

---

## 2. P0：当前最高优先级

### F0：Creator 大规模节点树性能优化

**背景**：当前 Creator 左侧节点树在节点规模上万后出现明显卡顿。每次 `renderTree()` 做整树重建：清空整个容器、递归生成全部可见 DOM、为每个节点绑定独立事件。万级节点下交互响应退化严重。

**已有完整路线图**：CREATOR_TREE_PERFORMANCE_ROADMAP.md

**实施方案**：

| 步骤 | 动作 | 产出 |
|---|---:|---|
| F0.1 | 建立 1k/5k/10k 节点基线 profiling | 量化首屏、滚动、展开、折叠、过滤、选中耗时 |
| F0.2 | 重组树数据层 | 常驻缓存 nodeMap/childMap/visibleRows |
| F0.3 | 引入拍平可见行模型 | 递归 DOM 改为树语义+行渲染模型 |
| F0.4 | 接入虚拟滚动 | 只渲染视口内+缓冲区节点行 |
| F0.5 | 局部更新替代整树刷新 | 展开/折叠/选中/拖拽改为局部状态更新 |
| F0.6 | 改为容器级事件委托 | 减少逐节点绑定成本 |
| F0.7 | 搜索优化与超大树降级 | 索引化搜索、防抖、默认折叠深层分支 |
| F0.8 | 补齐验收标准与回归样例 | 纳入 Creator 回归基线 |

**关键文件**：`tools/source/web/GameAgentCreator/js/ui.js`（树渲染）、`core.js`（数据状态）、`layout.js`（布局）

---

### E1：World Tick Bootstrap

**背景**：world tick 在进入主 LLM 循环前缺少开局权威数据预取。模型靠 request_data 逐个查询 scene_state / npc_state 等基础事实。即使小世界也可能十几轮才开始收敛。

**实施方案**：

| 步骤 | 动作 | 涉及模块 |
|---|---:|---|
| E1.1 | 定义 bootstrap 预取查询类型 | scene_state, scene_occupants, player_state, player_inventory, task_state, item_presence, npc_state |
| E1.2 | executeWorldTick 入口的 bootstrap phase | pipeline.go - PromptBuilder 之前插入预取阶段 |
| E1.3 | 复用现有 request_data / handleDataRequest 语义 | pipeline.go - 不另造协议 |
| E1.4 | 结果注入请求级临时上下文 | context.go - BuiltContext 扩展临时 bootstrap block |
| E1.5 | 预取结果参与 Prompt 拼接 | prompt_builders.go - buildWorldTickPrompt 增加 bootstrap section |
| E1.6 | 同步预取优先，callback 复杂场景兜底 | external/dispatcher.go |

**验收**：纯净 demo world 下 rounds_used 显著降低；不破坏现有非权威查询链路。

---

### E2：收束控制

**背景**：多轮循环缺少收束机制。request_data 返回后直接继续下一轮，没有"事实是否已足够"的判断，也没有接近上限时的硬收束规则。

**实施方案**：

| 步骤 | 动作 | 涉及模块 |
|---|---:|---|
| E2.1 | world tick 专用查询预算控制 | types.go - executionConfig 增加 queryBudget |
| E2.2 | 连续查询轮次阶段上限 | pipeline.go - executeMultiTurnLoop 增加阶段计数器 |
| E2.3 | 接近上限时的强制收束 prompt | rounds_used > maxRounds*0.8 时启用收束指令 |
| E2.4 | 定义 convergenceCheck() 终止条件 | pipeline.go - 基础事实已足够时优先闭环 |
| E2.5 | vertical 管线独立收束策略 | pipeline.go - executeVertical 轮次控制优化 |

**验收**：小世界轮次稳定在 2-4 轮以内；不因收束阻断合法 request_data。

---

### E3：World Tick Prompt 收束重写

**背景**：当前 prompt 对"优先完成当前 tick"约束不足。模型先补齐 scene/npc/relation/memory 信息，再写 future_outline。小世界也偏高轮次。

**实施方案**：

| 步骤 | 动作 | 涉及模块 |
|---|---:|---|
| E3.1 | buildWorldTickPrompt 收束重写 | prompt_builders.go - completion-first 指令 |
| E3.2 | 关键缺失 vs 锦上添花分级查询 | prompt 结构增加必需/可选查询分级 |
| E3.3 | 小世界 vs 大世界差异化提示策略 | 基于节点数量和 scope 深度选择 prompt variant |
| E3.4 | 时序调整：tick summary 优先于 future_outline | 降低"先写大纲再补细节"的冲动 |

**验收**：相同 world 下 rounds_used 降低 30%+；输出完整性不下降。

---

## 3. P1：高优先级

### E4：查询结果摘要化

**背景**：handleDataRequest 返回节点详情/记忆/关系/timeline 的文本碎片拼接。模型需要自己做信息归并，容易串行补查。

**方案**：在 pipeline.go 中新增 `summarizeQueryResults()` 摘要层，将 node/memory/relation 原始结果按 scope 聚合、去重、精简为结构化摘要。风格参考 `buildWorldTickRelationSummary()`，并按 PipelineMode 选择摘要粒度。

---

### E5：Round Context 压缩

**背景**：每轮分析与查询全量进入 round-state 上下文，prompt 越跑越像"研究过程记录"，历史噪声压过当前目标。

**方案**：RoundState.SupplementalContext 改为摘要压缩策略：旧轮次压缩为阶段性摘要、只保留关键事实/决策/缺口。tasktree.go 中树节点摘要化，summary-first 的 round history。

---

### E6：世界冷启动接口

**背景**：导入后缺少正式的世界运行基座初始化接口。开发者容易误把 world tick 当成初始化。

**方案**：pipeline.go 新增 `ColdStartWorld(worldID string) (*ColdStartResult, error)`，纯内存计算（不触发 LLM），模式参数 initial/rebuild。输出：运行基座组件列表、缺失设定警告。DevCli 接入 `world cold-start` 命令。

---

### E7：Demo Authority 接入

**背景**：demo-state.yaml 的 authority facts 只在 Worker play 中生效，未注入 world tick。Demo 链路与真实权威查询流程有偏差。

**方案**：Worker 作为 Demo authority 响应器，world tick bootstrap 通过 Worker 用 demo-state.yaml 回答。不把 demo-state 镜像为 Engine 组件，保持 authority 边界。

---

### E8：Store 数据层覆盖测试加固

**背景**：核心持久化层有 41 个测试文件，但 pause/resume、state-components、world_settings 等路径覆盖率不足。

**方案**：

| 步骤 | 测试目标 | 覆盖文件 |
|---|---:|---|
| E8.1 | PausedExecution 路径 | async_callbacks.go |
| E8.2 | 组件 CRUD 事务边界 | components.go |
| E8.3 | 关系批量操作 | relations.go |
| E8.4 | 迁移与写重试 | migrations.go, write_retry.go |
| E8.5 | 快照创建/验证/恢复 | snapshots.go |
| E8.6 | 世界设置与策略一致性 | world_settings.go, policy.go |

---

### E9：PipelineMode 关系装配策略硬化

**背景**：pipeline.go 头部注释定义了五条任务级关系装配策略约束，但 ContextBuilder 尚未完全对齐。Vertical/Polling/Full 的差异不仅是轮次，还包括关系子图装配强度。

**方案**：ContextBuilder.Build 按 TaskType 差异化装配。NPC 对话优先环境关系，world tick 优先摘要性关系。IncludeRelatedNodes 受限扩图。vertical 最小闭环子图，full 结构化扩图。

---

## 4. P2：中优先级

### E10：Callback Bootstrap 兜底

同步预取超时或无法快速返回时，按配置降级为 callback/paused-execution/resume。不另造第二套 bootstrap 专用异步协议。详见 ENGINE_IMPROVEMENT_ROADMAP.md。

### E11：动作系统 Schema 验证与扩展

内置 5 个动作（UpdateMood/AddMemory/SendDialogue/AdjustRelation/SpawnItem）。方案：每个 Action 暴露 Schema() 方法、validateActionCallsBySchema 推广到所有路径、action.Registry 增加 RegisterExternal() 外部注入接口、异步动作超时控制。

### E12：PolicyEngine 冲突解析

规则明确为：blocked > allowed > safe。冲突时记录 warning。增加作用域优先级（world > scope > node）和参数级策略。

### E13：前端交互响应优化

API 请求缓存与去重、乐观更新（修改立即响应，失败回滚）、大列表懒加载与骨架屏、日志和 traces 页面虚拟滚动。

### E14：Autonomous 调度路线图

按 AUTONOMOUS_SCHEDULING_ROADMAP.md 顺序：1. priority/cooldown 2. 节点评分与批处理 3. 生命周期状态跟踪 4. 事件驱动唤醒。

### E15：World Tick Context 路线图

按 WORLD_TICK_CONTEXT_ROADMAP.md 顺序：1. world_focus 组件契约 2. 候选节点评分 3. scope refinement 4. tick 摘要与 refined scope 连接。

---

## 5. P3：低优先级

### E16：SDK 文档与示例完善

Go SDK 集成入门示例、tick/play 工作流 SDK 调用示例、错误码文档化、各语言 SDK 能力状态更新。

### E17：组件系统通用化

运行时注册自定义组件类型、组件数据验证器接口、组件依赖声明。

### E18：Telemetry 可观测性增强

结构化日志字段标准化、关键路径指标收集（轮次/耗时/token/查询命中率）、Trace ID 传播、可选 Prometheus /metrics 端点。

### E19：工具链工作流接入

DevCli: `world cold-start`, `world bootstrap inspect`。Creator: 世界初始化操作、运行基座状态查看。Worker: Demo authority query 标准响应模式。

### E20：前端 i18n 补齐

扫描所有用户可见文本，补齐 zh/en 翻译表，新页面自动要求翻译键值。

---

## 6. P4：长期规划

### E21：多世界隔离边界明确

跨世界引用检查、快照范围限定、API 请求 worldID 强制校验。

### E22：压力测试与 Benchmark

多 NPC 并发 autonomous_act、万级节点 world tick、高并发 invoke benchmark、打包产物 smoke 自动化。

### E23：Engine Kernelization

契约成熟后进一步分离运行时内核与外围工具。详见 ENGINE_KERNELIZATION_MEMO.md。

### E24：多语言 SDK 扩展

非 Go 生态 SDK，优先 TypeScript/Python。

---

## 7. 验收与回归

### 常规验收基线

| 基线 | 目标 |
|---|---|
| 纯净 demo world | world tick 低轮次收敛且输出完整 |
| 带 authority bootstrap 的 demo world | world tick 正确利用权威快照 |
| 重复导入污染世界 | world tick 不因历史噪声撞满轮次 |
| callback authority 场景 | paused execution/resume 链路稳定 |

### 推荐测量指标

| 指标 | 目标值 |
|---|---|
| rounds_used / max_analysis_rounds 比值 | < 0.6 |
| bootstrap 命中率 | > 80% |
| future_outline 完整性 | 每次 tick 必生成 |
| advanced_ticks 合理性 | 与 TickScaleMode 一致 |
| 重复低价值查询次数 | 0 |
| 测试通过率 | 100% |
| Creator 10k 节点首屏耗时 | < 2s |

---

## 8. 现有路线图文档索引

| 文档 | 中文章节 |
|---|---|
| Engine 改进路线图 | [ENGINE_IMPROVEMENT_ROADMAP.md](./ENGINE_IMPROVEMENT_ROADMAP.md) |
| Creator 树性能路线图 | [CREATOR_TREE_PERFORMANCE_ROADMAP.md](./CREATOR_TREE_PERFORMANCE_ROADMAP.md) |
| Autonomous 调度路线图 | [AUTONOMOUS_SCHEDULING_ROADMAP.md](./AUTONOMOUS_SCHEDULING_ROADMAP.md) |
| World Tick Context 路线图 | [WORLD_TICK_CONTEXT_ROADMAP.md](./WORLD_TICK_CONTEXT_ROADMAP.md) |
| Engine Kernelization 备忘录 | [ENGINE_KERNELIZATION_MEMO.md](./ENGINE_KERNELIZATION_MEMO.md) |
| 未来开发计划 | [FUTURE_DEVELOPMENT_PLAN.md](./FUTURE_DEVELOPMENT_PLAN.md) |
| Roleplay Interaction 路线图 | [ROLEPLAY_INTERACTION_ROADMAP.md](./ROLEPLAY_INTERACTION_ROADMAP.md) |
| Player Input Pipeline | [PLAYER_INPUT_PIPELINE.md](./PLAYER_INPUT_PIPELINE.md) |
| Interaction API | [INTERACTION_API.md](./INTERACTION_API.md) |
| Engine Query Contract | [ENGINE_QUERY_CONTRACT.md](./ENGINE_QUERY_CONTRACT.md) |
| Engine Graph Semantics | [ENGINE_GRAPH_SEMANTICS.md](./ENGINE_GRAPH_SEMANTICS.md) |

---

*本文档整合了项目中所有现有的路线图和规划，是团队开发决策的单一参照。开发团队应在每个 sprint 结束时更新本文档的状态列。*



