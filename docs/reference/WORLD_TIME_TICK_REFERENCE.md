# 世界时间与 Tick 约束参考

**中文** | [**English**](./WORLD_TIME_TICK_REFERENCE_EN.md)

本文档描述 `world_time_settings`、`world_time_state`、`requested_ticks` 和 `advanced_ticks` 的当前行为。

## 1. 世界时间配置

世界级动态设置 `world_time_settings` 目前包含以下关键字段：

- `tick_scale_mode`：`fixed` 或 `flexible`
- `tick_min_unit`：最小 tick 单位，例如 `时辰`、`秒`
- `tick_step`：一次标准 tick 代表多少个最小单位
- `tick_units`：从大到小排列的时间单位列表
- `time_scale_carry`：从小到大的相邻进位规则
- `time_calendar`：可选日历模板，启用时必须提供 `calendar_name`
- `unit_value_sequences`：符号单位序列，例如 `时辰 = 子丑寅卯...`

约束规则：

- `tick_min_unit` 必须等于 `tick_units` 的最后一项。
- `time_scale_carry` 必须完整覆盖所有相邻单位。
- 启用 `time_calendar` 时，`calendar_name` 必填，且 `units` 必须与 `tick_units` 一一对应。
- 符号单位如果配置了进位规则，`unit_value_sequences` 的长度必须等于该进位基数。

## 2. fixed 与 flexible

### `fixed`

- 每次世界 tick 只能推进 1 个标准 tick。
- API/SDK/DevCli 传入的 `requested_ticks` 只能是 `1`。
- 模型即使返回其他 `advanced_ticks`，最终也会被引擎收敛为 `1`。

### `flexible`

- 每次世界 tick 可以推进多个标准 tick。
- 调用端可以先传入 `requested_ticks` 作为期望推进量。
- 模型可以在 `world_tick` 输出中返回 `advanced_ticks`，引擎会以最终采纳值持久化状态与时间线。

## 3. world_time_state

引擎会在 `world_tick` 后维护 `world_time_state` 连续性组件，包含：

- 当前时间规则镜像：`tick_scale_mode`、`tick_min_unit`、`tick_step`、`tick_units`
- 当前时间状态：`calendar_name`、`current_units`、`current_time_label`
- 推进记录：`total_ticks`、`last_tick_number`、`last_tick_type`、`last_advanced_ticks`
- 元数据：`advanced_min_units`、`external_time_label` 等

当存在多单位进位规则时，引擎会基于上一次 `world_time_state` 自动推进并进位。

## 4. world_tick 请求与响应

`POST /api/v1/worlds/{world_id}/ticks/advance` 现在支持：

```json
{
  "tick_type": "scheduled",
  "game_time": "太阴历 8 年 7 月 20 日 卯时",
  "requested_ticks": 3,
  "autonomous_limit": 10
}
```

响应新增：

- 顶层 `advanced_ticks`
- 顶层 `world_time_state`
- `invoke.advanced_ticks`

时间线归档 `data` 中新增：

- `advanced_ticks`
- `previous_world_time_state`
- `world_time_state`

## 5. SDK 与 DevCli

SDK：

- 新增 `AdvanceTickWithOptions(worldID, tickType, gameTime string, requestedTicks *int, autonomousLimit *int)`
- `TickResponse` 可直接读取 `AdvancedTicks` 与 `WorldTimeState`
- timeline 结构可直接读取 `AdvancedTicks`

DevCli：

- `GameAgentDevCli tick <world-id> --requested-ticks 3`
- `GameAgentDevCli world tick <world-id> --requested-ticks 3`

## 6. Creator

当前 Creator 至少可以通过连续性状态与时间线视图读取：

- `world_time_state` 组件
- timeline `data` 中的时间推进结果

如果后续为 Creator 增加专门的世界时间设置表单，应直接映射到 `world_time_settings`，而不是再引入单独的配置源。

## Creator 当前支持

- Creator 的 Settings 页面已经可以直接编辑 world_time_settings，而不是仅用于只读查看。
- Creator 的 Advance Tick 弹窗支持填写 requested_ticks、game_time 和 autonomous_limit。
- Tick 执行完成后，Creator 会显示 advanced_ticks、world_time_state，以及完整返回 JSON。
- State、Timelines、Continuity 页面都会显示 world_time_state 与 advanced_ticks，方便追踪世界时间推进链路。
- DevCli 的 `timeline latest`、`timeline list` 与 `debug continuity` 摘要也会直接显示 `advanced_ticks` 和世界时间标签；`debug continuity` 还会显示上一刻的世界时间标签，便于排查推进失真。
