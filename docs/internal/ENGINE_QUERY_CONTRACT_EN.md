# Engine-to-Game Query Contract

[**中文**](./ENGINE_QUERY_CONTRACT.md) | **English**

This document defines the contract by which Engine queries authoritative data from `gameagentworker` through the asynchronous `game_client` interface.

## 1. Query Principles

- Queries should only be triggered when Engine lacks key authoritative information.
- Query results are returned by the game side, and Engine continues reasoning through callback/resume.
- The query interface returns current authoritative facts rather than suggested actions.

## 2. Query Target

Use the following target uniformly:

- `target = "game_client"`

Default external interface name:

- `game_client_request_data`

Later versions may expose more fine-grained query aliases through request-scoped dynamic interfaces.

## 3. Recommended Query Types

The first version should standardize at least these query types:

- `player_state`
- `player_inventory`
- `player_wallet`
- `player_location`
- `npc_location`
- `scene_state`
- `room_state`
- `task_state`
- `item_presence`
- `relationship_hint`

Among them:

- `relationship_hint` is only for supplementing authoritative rule-side status from the game side and should not replace Engine's long-term relationship graph.

## 4. Example Request

```json
{
  "label": "fetch_player_scene_context",
  "target": "game_client",
  "external_interface": "game_client_request_data",
  "queries": [
    {
      "type": "player_location",
      "node_id": "player_001"
    },
    {
      "type": "scene_state",
      "node_id": "scene_inn"
    },
    {
      "type": "item_presence",
      "node_id": "player_001",
      "filter": "bloody_dagger"
    }
  ]
}
```

## 5. Example Callback Result

```json
{
  "callback_id": "cb-123",
  "status": "success",
  "result": {
    "player_location": {
      "node_id": "player_001",
      "scene_id": "scene_inn"
    },
    "scene_state": {
      "scene_id": "scene_inn",
      "occupants": ["player_001", "npc_innkeeper", "npc_messenger"],
      "counter_open": true
    },
    "item_presence": {
      "node_id": "player_001",
      "item_id": "bloody_dagger",
      "present": false
    }
  }
}
```

## 6. Safety and Granularity Limits

- Engine must not be allowed to pull an entire local save without limits.
- Each query should focus on the minimal facts required for the current turn.
- Highly sensitive internal script fields, hidden flags, drop tables, and similar fields should remain closed by default.
- Query types should follow a whitelist rather than temporary free-form assembly of low-level fields.

## 7. Requirements for Play Mode

- High-risk structured actions inferred from `/+gift`, `/+show_item`, `/+trade`, and `/+act` should land on the game side first.
- If Engine needs to confirm related authoritative facts, it may still query back through this contract.
- In group-chat mode, querying `room_state` is allowed, but room speaking order should still remain under game-side control.
