# Player Intent Proposal Schema

This document defines the structured output produced after interpreting player natural language.

## 1. Top-Level Structure

```json
{
  "intent": {
    "type": "composite",
    "actor_node_id": "player_001",
    "scene_node_id": "scene_inn",
    "target_node_id": "npc_innkeeper",
    "summary": "The player shows a blood-stained dagger and asks the innkeeper whether they have seen its owner.",
    "risk_level": "medium",
    "confidence": 0.88,
    "steps": [
      {
        "type": "show_item",
        "target_node_id": "npc_innkeeper",
        "item_id": "knife_bloody",
        "preconditions": [
          {"type": "same_scene", "actor_node_id": "player_001", "target_node_id": "npc_innkeeper"},
          {"type": "item_present", "actor_node_id": "player_001", "item_id": "knife_bloody"}
        ]
      },
      {
        "type": "speech",
        "target_node_id": "npc_innkeeper",
        "content": "今晚有没有见过这把刀的主人？"
      }
    ]
  },
  "missing_facts": [],
  "suggested_interaction": {
    "mode": "direct_dialogue",
    "event_type": "show_item"
  }
}
```

## 2. Top-Level Fields

### 2.1 `intent`

The structured action proposal body.

### 2.2 `missing_facts`

Key facts that the model believes are still missing before formal validation/execution.

Examples:

- the player's current location is not yet confirmed
- whether the target NPC is present is not yet confirmed
- whether the player actually holds a specific item is not yet confirmed

### 2.3 `suggested_interaction`

The recommended interaction type to bridge into after the action executes successfully.

## 3. PlayerIntent

### 3.1 Fields

- `type`
  - `speech`
  - `show_item`
  - `gift`
  - `trade_request`
  - `threaten`
  - `move`
  - `inspect`
  - `use_item`
  - `composite`
- `actor_node_id`
- `scene_node_id`
- `target_node_id`
- `summary`
- `risk_level`
  - `low`
  - `medium`
  - `high`
- `confidence`
- `steps`

### 3.2 Constraints

- a single-step intent may omit `steps`
- `composite` must include `steps`
- `confidence` represents interpretation confidence only, not execution permission

## 4. PlayerIntentStep

### 4.1 Common Fields

- `type`
- `target_node_id`
- `scene_node_id`
- `item_id`
- `content`
- `args`
- `preconditions`

### 4.2 Minimum Field Requirements by Type

#### `speech`

- `content`
- optional `target_node_id`

#### `show_item`

- `item_id`
- `target_node_id`

#### `gift`

- `item_id`
- `target_node_id`

#### `trade_request`

- `target_node_id`

#### `threaten`

- `target_node_id`

#### `move`

- `target_node_id` or `args.destination_scene_id`

#### `inspect`

- `target_node_id` or `item_id`

#### `use_item`

- `item_id`

## 5. Preconditions

### 5.1 Fields

- `type`
- `actor_node_id`
- `target_node_id`
- `scene_node_id`
- `item_id`
- `task_id`
- `expected`
- `args`

### 5.2 Supported Precondition Types in the First Version

- `same_scene`
- `target_present`
- `item_present`
- `money_at_least`
- `task_status`
- `scene_flag`
- `location_accessible`

### 5.3 Example

```json
{
  "type": "item_present",
  "actor_node_id": "player_001",
  "item_id": "knife_bloody"
}
```

## 6. MissingFacts

### 6.1 Fields

- `type`
- `node_id`
- `item_id`
- `task_id`
- `reason`

### 6.2 Recommended Missing-Fact Types

- `player_location`
- `target_location`
- `item_presence`
- `scene_state`
- `task_state`
- `wallet_state`

## 7. SuggestedInteraction

### 7.1 Fields

- `mode`
- `event_type`
- `audience_scope`
- `target_node_id`

### 7.2 Recommended Mapping

- `speech` -> `direct_dialogue` / `group_chat`
- `show_item` -> `direct_dialogue + event=show_item`
- `gift` -> `gift_response + event=gift`
- `trade_request` -> `trade_dialogue + event=trade_request`
- `threaten` -> `direct_dialogue + event=threaten`

## 8. Recommended Validator Output

```json
{
  "status": "accepted",
  "steps": [
    {"index": 0, "status": "accepted"},
    {"index": 1, "status": "accepted"}
  ]
}
```

Recommended statuses:

- `accepted`
- `rejected`
- `partially_accepted`

## 9. Recommended Executor Output

```json
{
  "status": "accepted",
  "applied_steps": [0, 1],
  "world_updates": [
    {"type": "inventory_transfer", "item_id": "silver_ring", "from": "player_001", "to": "npc_innkeeper"}
  ],
  "triggered_interaction": {
    "mode": "gift_response",
    "target_node_id": "npc_innkeeper",
    "event_type": "gift"
  }
}
```

## 10. First-Version Boundary

The first-version schema does not promise:

- reliable mapping from arbitrary free text to complex world operations
- parallel response from multiple NPCs
- direct landing of composite actions without validation

The first-version schema is primarily intended to solve:

- player natural language -> structured proposal
- structured proposal -> game-side authoritative validation
- validation passed -> interaction bridge
