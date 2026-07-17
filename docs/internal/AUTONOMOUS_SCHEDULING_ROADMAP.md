# Autonomous 调度路线图

**中文** | [**English**](./AUTONOMOUS_SCHEDULING_ROADMAP_EN.md)

本文记录未来如何把 autonomous 执行从基础触发扫描，演进为一个有边界的调度系统。

## 1. 当前基线

当前 autonomous 执行支持三种触发模式：

- `manual`
- `world_tick_sync`
- `scheduled`

当前行为有意保持简单：

- 加载一个 world 的 autonomous 组件
- 顺序扫描它们
- 按 enabled 标志和 trigger 过滤
- 对 `scheduled` 按时间间隔检查是否到期
- 到达固定运行上限后停止

这足以支撑一个基线版本，但它还不是一个真正的调度模型。

## 2. 未来目标

未来目标不是让 autonomous 行为变成全局常驻、永远在线。

真正的目标是：

- 让正确的节点在正确的时间运行
- 避免把全世界扫描当作主要机制
- 支持有边界的优先级与批处理
- 保持 autonomous 行为与外部权威回调和恢复流程兼容

## 3. 为什么未来需要调度

当前模型仍有这些限制：

- 没有优先级
- 没有显式排队状态
- 没有事件驱动唤醒路径
- 没有面对异构节点的公平批处理策略
- 无法区分“重要但未到期”和“已到期但价值很低”

## 4. 未来调度模型

### 4.1 派发分层

未来的 autonomous 派发在概念上应拆分为：

- trigger 准入
- 唤醒 / 到期状态判定
- 优先级评分
- 批次选择
- 执行
- 运行后状态更新

### 4.2 Trigger 准入

Trigger 准入应继续支持当前模式：

- manual
- world tick sync
- scheduled

但后续也可以在不破坏这些粗粒度模式的前提下，增加事件驱动唤醒。

## 5. 优先级与批处理

### 5.1 为什么优先级必须独立于 Trigger

Trigger 回答的是“这个节点什么时候可以运行”。

Priority 回答的是“哪些已可运行的节点应该优先运行”。

这两个字段必须保持分离。

### 5.2 建议的未来优先级信号

未来优先级可以考虑：

- 配置中的显式 autonomous priority
- 与当前活跃 world scope 的关联度
- 最近玩家交互相关性
- 最近世界事件相关性
- 唤醒原因的严重度
- 对长期未运行节点的防饥饿补偿

### 5.3 建议的批处理规则

未来批次选择应受以下边界控制：

- 每个 world 每次派发的最大节点数
- 可选的每-trigger 上限
- 可选的每-scope 上限
- 针对重复唤醒的冷却窗口

## 6. 事件驱动唤醒

### 6.1 含义

事件驱动唤醒意味着，系统不再仅通过每个周期扫描所有 autonomous 组件来发现可运行节点。

取而代之的是，内部或外部事件会显式地把节点加入队列，或标记为新近相关。

### 6.2 示例唤醒来源

未来可能的唤醒来源包括：

- 玩家对某节点发起对话
- 玩家动作影响某节点或其所在场景
- scene-state 或 room-state 变化
- quest-state 或 item-ownership 变化
- authority callback 完成
- world tick 生成的事件触及某个节点或 scope

### 6.3 为什么它不替代游戏侧驱动

游戏侧驱动与事件驱动唤醒不是同一件事。

游戏侧驱动意味着外部系统可以调用或恢复 Engine 工作。

事件驱动唤醒意味着 Engine / service 会更精确地维护“哪些 autonomous 节点现在值得运行”的内部模型。

这两种机制应该共存。

## 7. Autonomous 运行时状态机

### 7.1 建议范围

未来工作应优先增加调度生命周期状态机，而不是游戏人格状态机。

建议的运行时生命周期状态：

- `idle`
- `queued`
- `running`
- `waiting_external`
- `cooled_down`
- `blocked`
- `failed`

这个状态机只描述执行生命周期。

它不应与巡逻、交易、恐惧、战斗等 gameplay 行为状态混为一谈。

### 7.2 为什么重要

它有助于：

- 队列可见性
- 重试控制
- 重复唤醒抑制
- callback 恢复正确性
- 后续可观测性与公平性

## 8. 与 World Tick 的关系

未来的 autonomous 调度应与 world-tick scope 选择协同，但保持分离。

- world tick = 摘要与世界推进
- autonomous = 节点级后续执行

未来可用的整合点包括：

- world tick 可以产出一份值得唤醒的节点短名单
- `world_focus` 与活跃度评分可以影响 autonomous priority
- autonomous 唤醒可以反向影响后续 scope 摘要

## 9. 实施顺序

建议的未来实施顺序：

1. 增加显式 autonomous priority 与 cooldown 语义
2. 增加可运行节点评分与有边界批处理
3. 增加执行生命周期状态跟踪
4. 增加事件驱动唤醒队列或 wake-mark 模型
5. 将唤醒原因与 world-tick 输出及 interaction 流整合
6. 仅在之后再评估是否需要更丰富的行为状态框架
