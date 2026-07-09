# GameAgentCreator Guide

[**中文**](./GUIDE_GAMEAGENTCREATOR.md) | **English**

GameAgentCreator is the bundled browser-based visual editor for GameAgentEngine.

---

## Current Pages

Creator currently includes these main pages:

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

---

## What New Users Touch First

- use `Worlds` to create and browse worlds
- use `Settings` to configure runtime parameters and `world_time_settings`
- use `State` to inspect `world_time_state` and other state components
- use `Timelines` and `Continuity` to inspect time advancement results

---

## World Time Support

Creator currently supports:

- editing `world_time_settings`
- validating minimum units and unit ordering
- showing `advanced_ticks` in the tick result dialog
- showing `world_time_state.current_time_label`
- displaying world time state in `State`, `Timelines`, and `Continuity`

If `world_time_settings` is missing, world-time-dependent save and reasoning flows are intentionally blocked by design.
