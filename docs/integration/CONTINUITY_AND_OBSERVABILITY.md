# 连续性与可观测性

**中文** | [**English**](./CONTINUITY_AND_OBSERVABILITY_EN.md)

本页是当前 `world_tick` 连续性检查与可观测性排障的主入口文档。

## 1. 重点检查对象

当前连续性检查主要围绕四类产物：

- 时间线（timelines）
- 连续性状态组件
- 结构化日志
- 调试追踪

## 2. 推荐排查顺序

调试连续性问题时，建议按以下顺序进行：

1. 检查最新时间线
2. 检查 `world_state`、`story_state`、`story_history` 和 `tick_policy`
3. 围绕同一个 `request_id` 检查日志与追踪
4. 做有针对性的连续性修复
5. 再推进下一次 tick 并重新比对

## 3. 工具入口

使用 DevCli：

```bash
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli timeline list <world-id> --limit 5
GameAgentDevCli state list <world-id>
GameAgentDevCli debug continuity <world-id>
GameAgentDevCli logs --world <world-id> --details
GameAgentDevCli debug traces --world <world-id> --limit 10
```

使用 Creator：

- `Continuity`
- `State`
- `Timelines`
- `Logs`
- `Traces`

## 4. 当前持久化模型

当前 `world_tick` 持久化有意拆分为：

- `logs`：运行时可观测性
- `timelines`：按 tick 排序的归档
- 连续性状态组件：承载结构化继承状态

## 5. 常见回归判定

出现以下情况时，应视为连续性回归：

- 已知规范事实从 `world_state` 中消失
- 最新历史条目丢失重要保留事实
- 模型不再遵守 `tick_policy` 的连续性约束
- 同一路请求的日志、追踪和时间线无法对齐

## 6. 历史材料与补充说明

本页现在是连续性工作流的规范主入口。

诊断连续性回归时，应集中保留这些实践检查点：

- 优先使用 `timeline latest`、`timeline list`、`state list`、`logs` 和 `debug traces`
- 将 `state_snapshot` 视为 Engine 生成的只读检查点载荷
- 仅当你明确在修复连续性状态时，才修改 `tick_policy`、`world_state`、`story_state` 或 `story_history`
- 每次定点修改后只推进一个 tick 再检查，不要先叠加多次盲改

## 7. 最小回归样例

当你需要一个快速回归样例时，可以先把一个稳定规范事实同时写入 `world_state` 和 `story_history`，并通过 `tick_policy` 保护该事实，再推进一个 tick，并验证：

1. 该事实仍存在于 `world_state.canonical_facts`
2. 最新的 `story_history` 条目仍保留同一事实，而不是退化成更模糊的改写
3. 日志、追踪和时间线仍能对齐到同一路请求

一个可用的固定事实示例是：

- `地下52米量子谐振腔`

如果该事实消失、丢失深度/状态细节，或不再受最新 `tick_policy` 路径保护，则应视为回归。
