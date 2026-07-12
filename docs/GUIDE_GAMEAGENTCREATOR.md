# GameAgentCreator 指南

**中文** | [**English**](./GUIDE_GAMEAGENTCREATOR_EN.md)

GameAgentCreator 是 GameAgentEngine 附带的浏览器可视化编辑器。

---

## 当前页面

当前 Creator 已包含这些主页面：

- `Worlds`
- `Snapshots`
- `Plans`
- `Policy`
- `Settings`
- `Continuity`
- `State`
- `Timelines`
- `Logs`
- `Traces`
- `Tasks`

---

## 新手最先会用到什么

- 在 `Worlds` 创建和浏览世界
- 在 `Settings` 配置世界运行参数和 `world_time_settings`
- 在 `State` 查看 `world_time_state` 等状态组件
- 在 `Timelines` 和 `Continuity` 排查时间推进结果
- 在 `Tasks` 查看运行时任务列表和状态

---

## 世界时间相关

当前 Creator 已支持：

- 编辑 `world_time_settings`
- 校验最小时间单位和单位顺序
- 在 Tick 结果弹窗中查看 `advanced_ticks`
- 查看 `world_time_state.current_time_label`
- 在 `State`、`Timelines`、`Continuity` 页面展示世界时间状态

如果没有配置 `world_time_settings`，依赖世界时间推进的保存/推理流程会被阻塞，这个提示是设计的一部分。

---

## 连续性排查建议

建议按这个顺序排查：

1. `Timelines` 看最近一次 Tick 结果
2. `State` 看 `world_time_state`、`world_state`、`story_history`
3. `Continuity` 看聚合结果和差异
4. `Logs` / `Traces` 对齐同一个请求
