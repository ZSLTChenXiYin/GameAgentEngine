# CloneWorld Snapshot Strategy

**中文** | [**English**](./CLONEWORLD_SNAPSHOT_STRATEGY_EN.md)

本文档描述 `CloneWorld` 在存档场景下的推荐演进方向。

当前 `CloneWorld` 的实现更接近“复制一个世界及其全部业务数据”。
如果它要承担游戏存档职责，那么优化目标不应该只包含复制性能，还需要同时覆盖：

- 快照兼容性
- 恢复可靠性
- 版本迁移能力
- 大世界复制性能
- 存档生命周期管理

---

## 1. 目标定位

在存档语义下，`CloneWorld` 的核心目标应为：

1. 保存某一时刻完整、可恢复的 Agent 世界状态
2. 允许后续 Engine 版本安全识别并恢复旧存档
3. 尽量减少“运行中数据结构变化导致旧存档不可用”的风险
4. 在中大型世界下维持可接受的存档速度

因此建议把内部设计逐步从“世界复制”提升为“世界快照”。

---

## 2. 推荐语义拆分

长期建议把两个需求拆开：

### A. CloneWorld

用于：

- 平行剧情分支
- 调试副本
- 世界模板复制
- 仿真分叉

特点：

- 保持当前 API 风格
- 优先优化复制吞吐
- 可接受轻量元数据

### B. SnapshotWorld / RestoreWorld

用于：

- 游戏存档
- 自动存档
- 关键剧情前保存
- 回档恢复

特点：

- 强调兼容性和恢复可靠性
- 必须带版本信息
- 允许在恢复前做 schema/data migration

如果短期不新增 API，也建议先在 `CloneWorld` 内部按 Snapshot 语义设计元数据。

---

## 3. 快照最小元数据

建议后续为每个存档引入独立元数据头，至少包含：

| 字段 | 说明 |
|---|---|
| `snapshot_id` | 快照唯一标识 |
| `source_world_id` | 来源世界 UUID |
| `source_world_name` | 来源世界名 |
| `created_at` | 快照创建时间 |
| `engine_version` | 生成快照时的 Engine 版本 |
| `schema_version` | 生成快照时的数据结构版本 |
| `content_version` | 可选，业务内容版本 |
| `reason` | 手动存档 / 自动存档 / 关卡切换 / 测试复制 |
| `node_count` | 节点数量 |
| `component_count` | 组件数量 |
| `memory_count` | 记忆数量 |
| `relation_count` | 关系数量 |
| `payload_hash` | 快照体校验哈希 |

这些字段的价值不在“好看”，而在于：

- 恢复前做兼容判断
- 检测快照是否损坏
- 用于存档列表展示
- 支持后续迁移工具

---

## 4. 推荐数据结构

建议把快照逻辑拆成两层：

### 快照头（Snapshot Header）

存版本、来源、统计信息、校验信息。

### 快照体（Snapshot Payload）

存世界实际数据：

- world node
- nodes
- components
- memories
- relations
- world_settings
- world_policy
- propagation chains（如果启用）
- 未来可能加入的扩展实体

这样做的好处是：

- 快照列表无需加载完整 payload
- 恢复前可先校验 header
- 后续可替换底层 payload 编码而不破坏外层结构

---

## 5. 兼容性策略

这是存档设计里最重要的一层。

### 5.1 必须记录版本

只复制业务表而不记录 `engine_version/schema_version`，后续无法可靠判断旧存档是否可恢复。

### 5.2 恢复前先做版本门禁

建议恢复流程改成：

1. 读取快照头
2. 检查 `schema_version`
3. 若与当前版本一致，直接恢复
4. 若旧于当前版本，尝试 migration
5. 若高于当前版本，拒绝恢复或标记为只读

### 5.3 优先做逻辑迁移，不做数据库文件绑定

如果未来要兼容 SQLite / MySQL / schema 演化，推荐保存“逻辑世界结构”，而不是数据库物理文件副本。

逻辑快照的优点：

- 更容易跨数据库驱动恢复
- 更容易做字段级迁移
- 更适合长期版本演进

---

## 6. CloneWorld 的性能优化路线

在不改变外部行为的前提下，建议按下面顺序优化：

### 6.1 一次性构建映射表

当前复制过程最容易退化成大量查询往返。

推荐先在内存中构建：

- `oldNodeID -> oldUUID`
- `oldUUID -> newUUID`
- `newUUID -> newNodeID`

后续组件、记忆、关系复制全部只走内存映射，不再逐条回查数据库。

### 6.2 分批读取 + 分批写入

对于大世界，建议：

- 节点批量读取
- 组件批量读取
- 记忆批量读取
- 关系批量读取
- 批量插入（按实体分组）

这样可以显著降低事务内 SQL 次数。

### 6.3 减少事务中的重复解析

例如：

- 不重复按 UUID 查 int64 ID
- 不重复按 parentID 查 parentUUID
- 不在每个循环里单独查 world settings / world policy

### 6.4 为存档流程保留只读锁语义

如果 `CloneWorld` 用于存档，建议保留源世界锁能力，但锁语义应明确：

- 是否要求强一致快照
- 是否允许读取进行中但未提交的数据
- 是否允许“近似快照”换取性能

默认建议：存档场景使用强一致快照。

---

## 7. 恢复流程建议

如果未来引入 `RestoreWorld`，推荐：

1. 读取 snapshot header
2. 校验 `payload_hash`
3. 校验 `schema_version`
4. 如有需要，执行 migration
5. 创建新世界或覆盖恢复目标
6. 批量重建 nodes / components / memories / relations
7. 恢复 settings / policy / propagation data
8. 写入恢复日志

建议不要直接覆盖生产中的活跃世界，除非 API 明确要求，并且已加锁。

---

## 8. 推荐短期落地顺序

### 第一阶段：不改 API，先增强内部结构

- 为 Clone 结果补充元信息
- 在复制流程中引入批量映射
- 降低事务内 N+1 查询

### 第二阶段：引入显式快照元数据

- 新增 snapshot header 表或逻辑对象
- 给每次 Clone/存档写入版本信息和统计信息

### 第三阶段：引入 Restore 能力

- 恢复前版本检查
- migration hook
- 完整恢复链路

### 第四阶段：将存档与普通复制正式拆分

- 保留 CloneWorld
- 新增 SnapshotWorld / RestoreWorld

---

## 9. 结论

如果 `CloneWorld` 主要服务于游戏存档，那么它的正确演化方向不是“只把复制变快”，而是：

- 让复制结果成为可校验、可识别版本、可迁移、可恢复的世界快照
- 再在这个前提下优化复制吞吐

建议下一步优先做两件事：

1. 设计 Snapshot 元数据模型
2. 重构 CloneWorld 的内部复制流程为批量映射 + 批量写入

这两步完成后，再做 Restore 设计会顺很多。
