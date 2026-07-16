# Core Concepts

[**中文**](./CORE_CONCEPTS.md) | **English**

GameAgentEngine models a game world with four base abstractions: nodes, components, memories, and relations, then builds reasoning pipelines, continuity state, and world time progression on top.

---

## Nodes

Common node types:

- `world`
- `faction`
- `location`
- `npc`
- `item`
- `quest_line`
- `event`

A `world` node is the root of a world tree.

---

## Components

Common built-in components:

- `profile`
- `lore`
- `rule`
- `timeline`
- `action_policy`
- `relations`
- `prompt_profile`
- `autonomous`
- `world_state`
- `story_state`
- `story_history`
- `tick_policy`
- `state_snapshot`
- `world_time_state`

`world_time_state` is an Engine-maintained runtime state component that stores the current world time result.

---

## Memories

Common levels:

- `short_term`
- `long_term`
- `shared`
- `world`

Current propagation modes:

- `upward`
- `environment_scope`
- `organization_scope`
- `tag_broadcast`
- `targeted`
- `manual`

---

## Relations

Common relation types:

- `belongs_to`
- `ally`
- `enemy`
- `subordinate`
- `kinship`
- `located_at`
- `external_parent`

Current semantics:

- `parent_id` is the primary hierarchy and stable ownership chain
- `located_at` is current location
- `belongs_to` / `subordinate` are organization or control links
- `external_parent` is only an auxiliary scope edge

---

## world_settings, world_time_settings, world_time_state

This is the most important grouping to keep straight:

- `world_settings`: the world-level runtime settings container
- `world_time_settings`: time rules
- `world_time_state`: time result

There is no role overlap between the last two, but the former determines how the latter is generated.
