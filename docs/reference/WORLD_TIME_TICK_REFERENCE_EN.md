# World Time and Tick Constraint Reference

[**中文**](./WORLD_TIME_TICK_REFERENCE.md) | **English**

This document describes the current behavior of `world_time_settings`, `world_time_state`, `requested_ticks`, and `advanced_ticks`.

## 1. World Time Settings

`world_time_settings` currently exposes these key fields:

- `tick_scale_mode`: `fixed` or `flexible`
- `tick_min_unit`: the smallest tick unit, such as `时辰` or `second`
- `tick_step`: how many minimum units one standard tick represents
- `tick_units`: ordered from large to small
- `time_scale_carry`: adjacent carry rules ordered from small to large
- `time_calendar`: optional calendar template; requires `calendar_name` when enabled
- `unit_value_sequences`: symbolic ordered sequences such as `子丑寅卯...`

Validation rules:

- `tick_min_unit` must equal the last item in `tick_units`.
- `time_scale_carry` must fully cover every adjacent unit pair.
- When `time_calendar` is enabled, `calendar_name` is required and `units` must match `tick_units` exactly.
- If a symbolic unit participates in carry, its `unit_value_sequences` length must match the carry base.

## 2. `fixed` vs `flexible`

### `fixed`

- Each world tick can advance exactly one standard tick.
- `requested_ticks` must be `1`.
- Even if the model returns a different `advanced_ticks`, the engine collapses it to `1`.

### `flexible`

- Each world tick may advance multiple standard ticks.
- The caller can send `requested_ticks` as the intended progression amount.
- The model may return `advanced_ticks` in the `world_tick` output, and the engine persists the effective adopted value.

## 3. `world_time_state`

After every `world_tick`, the engine maintains a persistent `world_time_state` continuity component containing:

- mirrored rule fields: `tick_scale_mode`, `tick_min_unit`, `tick_step`, `tick_units`
- current time state: `calendar_name`, `current_units`, `current_time_label`
- progression bookkeeping: `total_ticks`, `last_tick_number`, `last_tick_type`, `last_advanced_ticks`
- metadata such as `advanced_min_units` and `external_time_label`

When multi-unit carry rules are configured, the engine advances and carries from the previous `world_time_state` automatically.

## 4. `world_tick` Request and Response

`POST /api/v1/worlds/{world_id}/ticks/advance` now supports:

```json
{
  "tick_type": "scheduled",
  "game_time": "Lunar Calendar Year 8 Month 7 Day 20 Mao Watch",
  "requested_ticks": 3,
  "autonomous_limit": 10
}
```

The response now includes:

- top-level `advanced_ticks`
- top-level `world_time_state`
- `invoke.advanced_ticks`

Persisted timeline `data` now includes:

- `advanced_ticks`
- `previous_world_time_state`
- `world_time_state`

## 5. SDK and DevCli

SDK:

- new `AdvanceTickWithOptions(worldID, tickType, gameTime string, requestedTicks *int, autonomousLimit *int)`
- `TickResponse` now exposes `AdvancedTicks` and `WorldTimeState`
- timeline envelopes now expose `AdvancedTicks`

DevCli:

- `GameAgentDevCli tick <world-id> --requested-ticks 3`
- `GameAgentDevCli world tick <world-id> --requested-ticks 3`

## 6. Creator

Creator can currently inspect world time progression through:

- the `world_time_state` continuity component
- timeline `data` that includes world time progression snapshots

If Creator later adds a dedicated world time settings form, it should map directly to `world_time_settings` instead of introducing a second configuration source.
