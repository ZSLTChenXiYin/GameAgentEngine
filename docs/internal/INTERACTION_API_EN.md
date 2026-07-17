# Interaction API Design

[**中文**](./INTERACTION_API.md) | **English**

This document defines the interaction semantics model built on top of the existing `invoke` execution kernel.

## 1. Design Goal

The current Engine inference entrypoint follows a single-focus-node model. The goal of the interaction API is not to replace `invoke`, but to add formal roleplay semantics on top of an `invoke` request:

- who is speaking
- who they are speaking to
- where they are speaking
- whether the turn is direct chat, group chat, or structured event feedback

## 2. External Layering

- execution kernel: continue to use `InvokeRequest -> Pipeline.Execute(...)`
- business API: allow `interaction/*` routes or SDK wrapper methods

In other words:

- `invoke` is the unified execution protocol
- `interaction/*` is the more ergonomic role-interaction entrypoint

## 3. InteractionContext

It is recommended to add the following structure under `InvokeContext`:

```json
{
  "interaction": {
    "mode": "direct_dialogue",
    "speaker_node_id": "player_001",
    "target_node_id": "npc_innkeeper",
    "scene_node_id": "scene_inn",
    "room_id": "room_inn_mainhall",
    "participant_node_ids": ["player_001", "npc_innkeeper"],
    "audience_scope": "public",
    "turn_index": 3,
    "event": {
      "type": "speech"
    }
  }
}
```

## 4. Recommended Fields

- `mode`
  - `direct_dialogue`
  - `group_chat`
  - `gift_response`
  - `trade_dialogue`
- `speaker_node_id`
- `target_node_id`
- `scene_node_id`
- `room_id`
- `participant_node_ids`
- `audience_scope`
  - `public`
  - `private`
  - `whisper`
- `turn_index`
- `event`

## 5. Event Types

The first version should support at least:

- `speech`
- `gift`
- `show_item`
- `trade_request`
- `threaten`

Example:

```json
{
  "mode": "gift_response",
  "speaker_node_id": "player_001",
  "target_node_id": "npc_innkeeper",
  "scene_node_id": "scene_inn",
  "participant_node_ids": ["player_001", "npc_innkeeper"],
  "event": {
    "type": "gift",
    "item_id": "silver_ring"
  }
}
```

## 6. Context Assembly Principles

- `target_node_id` is the primary viewpoint node.
- `speaker_node_id` is the formal acting participant.
- `scene_node_id` provides shared scene context.
- In `participant_node_ids`, nodes other than the primary target should enter only as lightweight summaries.
- Other participants should not receive full components, memories, or the complete relationship graph by default.

## 7. Group-Chat Strategy

The first version should not perform parallel multi-node reasoning for group chat.

Recommended strategy:

- the game side maintains authoritative room state
- worker decides who responds first
- Engine plays only one NPC per turn
- other participants enter the prompt only through room-context summaries

## 8. Mapping to Play Mode

- `/+talk innkeeper`
  - `mode=direct_dialogue`
- plain text input after a dialogue target has already been selected
  - `mode=direct_dialogue`
- `/+say 大家今晚都看见了什么？`
  - `mode=group_chat`
- `/+ask guard 你看见谁进门了？`
  - `mode=group_chat`, but `target_node_id=guard`
- `/+gift innkeeper silver_ring`
  - land on the game side first, then call `mode=gift_response`

Formal documentation now uses the `/+cmd + args` style consistently. Legacy aliases such as `/talk` and `/ask` remain only as compatibility input forms.
