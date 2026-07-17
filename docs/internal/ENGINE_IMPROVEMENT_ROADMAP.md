# Engine 改进路线图

**中文** | [**English**](./ENGINE_IMPROVEMENT_ROADMAP_EN.md)

本文档记录当前基于真实代码、打包产物和 Demo world tick 运行结果整理出的 Engine 后续改进路线图。

目标不是列出所有可能的未来功能，而是把当前已经确认的问题、改造方向、优先级与验收目标固定下来，避免后续实现偏离讨论结论。

---

## 1. 背景与现状

在当前实现中：

- 纯净 Demo world 可以在 world tick 中收敛；
- 但即使世界规模很小，仍可能需要十几轮查询才完成；
- 如果世界中存在重复导入或上下文污染，world tick 容易直接撞到 `max_analysis_rounds` 上限；
- world tick 当前主要依赖静态导入世界作为开局上下文；
- Worker / authority-state 并未默认成为 world tick 的开局权威数据来源。

由此得到两个结论：

1. 现有 world tick 管线具备基本可用性，但收束效率不足；
2. 当前 Demo 工作流中，“世界骨架”与“权威状态”之间的接合仍然不完整。

---

## 2. 已确认问题

### 2.1 world tick prompt 导向偏“先查全再写”

当前 world tick prompt 允许在缺事实时继续 `request_data`，但缺少足够强的“已有基础事实后优先完成当前 tick”的收束约束。

结果是：

- 模型经常先补齐 scene、NPC、relation、memory 等信息；
- 再开始写 `future_outline` 与当前 tick 事件；
- 对小世界也会出现偏高轮次。

### 2.2 `request_data` 结果回灌后直接继续下一轮

当前多轮循环中，只要 `request_data` 返回结果，就会把结果回灌并 `continue` 下一轮。

缺失点：

- 没有判断“当前事实是否已经足够收束”；
- 没有接近 `max_analysis_rounds` 时的硬收束机制；
- 没有 world tick 专用的查询预算控制。

### 2.3 round history 持续累积，prompt 易膨胀

每轮分析与数据查询结果都会进入 task-tree / round-state 上下文，后续 prompt 会继续携带这批历史。

结果是：

- prompt 越跑越像“研究过程记录”；
- 模型更容易继续做补充说明和补充查询；
- 小世界也会出现“历史噪声压过当前目标”的情况。

### 2.4 查询结果偏碎片化

当前 `handleDataRequest` 返回内容以节点详情、记忆、关系、timeline 等文本碎片为主。

问题不在于结果错误，而在于：

- 缺少面向 world tick 的中间摘要层；
- 模型需要自己做信息归并；
- 因而容易出现“先查 detail，再查 memory，再查 relation”的串行补查行为。

### 2.5 world tick 开局上下文仍偏薄

Demo world 的静态骨架足以支撑导入、对话与 play 演示，但对 world tick 而言：

- `demo-state.yaml` 中的 authority facts 未自动注入 world tick；
- 当前 context builder 主要消费的是 world/nodes/components/memories/relations/state blocks；
- 权威动态状态没有成为默认开局输入。

### 2.6 冷启动与 world tick 语义尚未显式分离

世界导入完成后，当前项目还缺少清晰、正式、稳定的“世界运行基座初始化”接口与工作流。

结果是：

- 开发者容易误把 world tick 当成初始化动作；
- 也容易误以为应该手工维护大量世界运行态组件。

---

## 3. 改造总原则

后续 Engine 改造应遵守以下原则：

1. 不把高频权威数据长期固化进 Engine；
2. 不要求开发者手工维护大量世界运行态组件；
3. world tick 优先基于“世界骨架 + 权威快照”完成当前 tick；
4. 收束控制优先于盲目增加 `max_analysis_rounds`；
5. 引擎能力与工作流入口分离，但语义保持统一。

---

## 4. P0：world tick bootstrap

### 4.1 目标

在 world tick 进入主 LLM 循环前，先按规则预取一小批关键权威数据，降低开局补查轮次。

### 4.2 推荐做法

- 保持一套统一的 authority query / `request_data` 语义；
- 同步预取作为优先路径；
- callback / paused-execution 作为复杂场景兜底；
- 把权威结果注入请求级临时上下文，而不是持久化成长期组件。

### 4.3 首批推荐查询类型

- `scene_state`
- `scene_occupants`
- `player_state`
- `player_inventory`
- `task_state`
- `item_presence`
- `npc_state`

### 4.4 预期收益

- 降低 world tick 开局低价值查询轮次；
- 缩小 Demo world 与真实 authority 流程之间的偏差；
- 为 play / Demo / 集成测试提供更接近真实线上链路的行为。

---

## 5. P0：收束控制

### 5.1 目标

避免 world tick 在小世界中也大量消耗轮次，或在复杂世界中反复补查直至撞上上限。

### 5.2 推荐改造点

- 引入 world tick 专用查询预算；
- 对连续查询轮次设置阶段上限；
- 在接近最大轮次时启用强制收束规则；
- 明确“基础事实已足够时优先完成当前 tick”的终止条件。

### 5.3 预期收益

- world tick 对 `max_analysis_rounds` 的依赖降低；
- 轮次分布更稳定；
- 更容易在小世界上形成快速闭环。

---

## 6. P0：world tick prompt 收束重写

### 6.1 目标

把 world tick prompt 从“允许无限细化”改为“优先完成当前 tick，仅在缺关键事实时补查”。

### 6.2 推荐方向

- 明确声明：已有 scene / participants / primary tension / minimal authority facts 后，应优先完成本 tick；
- 明确区分“关键缺失事实”和“锦上添花的补充事实”；
- 降低“先生成未来大纲、再补所有细节”的冲动；
- 对小 world 和大 world 提供不同的提示策略。

---

## 7. P1：查询结果摘要化

### 7.1 目标

减少模型对碎片化查询结果的二次拼装成本。

### 7.2 推荐改造点

- 在 `handleDataRequest` 之上增加 world tick 专用摘要层；
- 把原始 node/memory/relation 结果整理成更适合 world tick 读取的结构化摘要；
- 对 authority bootstrap 结果提供统一格式的 snapshot block。

### 7.3 预期收益

- 减少“先查 detail，再查 memory，再查 relation”的行为；
- 降低对多轮补查的依赖；
- 提升最终输出的一致性与可解释性。

---

## 8. P1：round context 压缩

### 8.1 目标

避免 round history 线性膨胀，降低“模型不断受旧查询记录影响”的副作用。

### 8.2 推荐方向

- 不再原样保留所有旧查询文本；
- 将旧轮次压缩成阶段性摘要；
- 只保留关键事实、关键决策、关键未解缺口；
- 为 world tick 提供 summary-first 的 round history 机制。

---

## 9. P1：世界冷启动接口

### 9.1 目标

把“导入后运行基座初始化”从隐式工作流变成正式 Engine 能力。

### 9.2 推荐能力边界

- 冷启动接口独立于 import；
- 语义独立于 world tick；
- 支持首次初始化与后续重建；
- 只生成运行基座，不同步权威动态状态。

### 9.3 建议输出

- 是否初始化成功；
- 生成或重用的组件列表；
- 初始化版本与来源标记；
- 缺失设定、弱建模、冲突结构等 warning。

---

## 10. P1：Demo authority 接入

### 10.1 目标

让 Demo world 的 `demo-state.yaml` 真正通过 authority 流程参与 world tick，而不是只在 Worker play 中生效。

### 10.2 推荐做法

- 让 Worker 作为 Demo authority 响应器；
- world tick bootstrap 和后续 authority query 均可由 Worker 用 `demo-state.yaml` 回答；
- 不把 `demo-state.yaml` 镜像为 Engine 组件。

### 10.3 预期收益

- Demo 更接近真实集成链路；
- world tick 对真实 authority 接入的测试覆盖更完整；
- 更容易发现 Engine 与 Worker 在 query contract 上的缺口。

---

## 11. P2：callback bootstrap 兜底

### 11.1 目标

让复杂或高延迟 authority 查询也能复用 Engine 当前已有的 runtime-task / paused-execution / resume 机制。

### 11.2 推荐原则

- world tick bootstrap 与普通 authority query 使用同一套语义；
- 优先同步预取；
- 超时或无法快速返回时，按配置降级为 callback / resume；
- 不再额外创造第二套“bootstrap 专用”异步协议。

---

## 12. P2：工具链工作流接入

### 12.1 DevCli

建议补齐：

- `world cold-start`
- `world cold-start --mode rebuild`
- `world bootstrap inspect` 或等价调试入口

### 12.2 Creator

建议补齐：

- 世界导入后的“初始化世界”操作；
- 运行基座状态查看；
- bootstrap / authority snapshot 调试视图。

### 12.3 Worker

建议补齐：

- Demo authority query 的标准响应模式；
- bootstrap query pack 的直接支持；
- 对同步与 callback 两条链路的可观测支持。

---

## 13. 验收与回归建议

后续每个阶段改造都应至少覆盖以下回归基线：

| 基线 | 目标 |
|---|---|
| 纯净 demo world | world tick 低轮次收敛且输出完整 |
| 带 authority bootstrap 的 demo world | world tick 正确利用 Worker 返回的权威快照 |
| 重复导入污染世界 | world tick 不因历史噪声轻易撞满轮次 |
| callback authority 场景 | paused execution / resume 链路行为稳定 |

推荐重点指标：

- `rounds_used`
- `max_analysis_rounds`
- bootstrap 是否命中
- 是否生成完整 `future_outline`
- 是否返回合理 `advanced_ticks`
- 是否发生重复低价值查询

---

## 14. 当前优先级总表

| 优先级 | 编号 | 改进项 | 目标 |
|---|---|---|---|
| P0 | E1 | world tick bootstrap | world tick 开局获得关键权威事实 |
| P0 | E2 | 收束控制 | 防止小世界也高轮次补查 |
| P0 | E3 | prompt 收束重写 | 优先完成当前 tick，而不是先查全 |
| P1 | E4 | 查询结果摘要化 | 降低模型对碎片信息的拼装成本 |
| P1 | E5 | round context 压缩 | 减少 prompt 膨胀与历史噪声 |
| P1 | E6 | 世界冷启动接口 | 正式生成运行基座 |
| P1 | E7 | Demo authority 接入 | 让 Demo 链路覆盖真实权威查询 |
| P2 | E8 | callback bootstrap 兜底 | 兼容复杂 authority 聚合场景 |
| P2 | E9 | 工具链工作流接入 | 把改造能力接入 Creator / DevCli / Worker |
| P2 | E10 | 回归与验收完善 | 固化 world tick 收敛能力与联调能力 |

---

## 15. 总结

当前 Engine 的主要问题不是“world tick 不能用”，而是：

- 开局权威事实不足；
- 查询补查过于自由；
- 收束控制偏弱；
- Demo authority 半边尚未接上。

后续改造应优先解决 world tick 的开局事实供给与收束控制，再推进 cold start、工具链与更复杂的 authority callback 路线。
