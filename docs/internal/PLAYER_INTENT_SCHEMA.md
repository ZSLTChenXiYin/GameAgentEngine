# 玩家行为提案 Schema

**中文** | [**English**](./PLAYER_INTENT_SCHEMA_EN.md)

本文定义玩家自然语言解释后的结构化输出。

## 1. 顶层结构

```json
{
  "intent": {
    "type": "composite",
    "actor_node_id": "player_001",
    "scene_node_id": "scene_inn",
    "target_node_id": "npc_innkeeper",
    "summary": "玩家展示带血短刀并询问老板是否见过刀的主人",
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

## 2. 顶层字段

### 2.1 `intent`

结构化行为提案本体。

### 2.2 `missing_facts`

模型认为在正式校验/执行前仍然缺失的关键事实。

例如：

- 玩家当前位置未确认
- 目标 NPC 是否在场未确认
- 玩家是否持有某物品未确认

### 2.3 `suggested_interaction`

当行为执行成功后，推荐桥接的 interaction 类型。

## 3. PlayerIntent

### 3.1 字段

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

### 3.2 约束

- 单步 intent 可以不带 `steps`
- `composite` 必须带 `steps`
- `confidence` 只是解释可信度，不是执行许可

## 4. PlayerIntentStep

### 4.1 通用字段

- `type`
- `target_node_id`
- `scene_node_id`
- `item_id`
- `content`
- `args`
- `preconditions`

### 4.2 各类型最小字段要求

#### `speech`

- `content`
- 可选 `target_node_id`

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

- `target_node_id` 或 `args.destination_scene_id`

#### `inspect`

- `target_node_id` 或 `item_id`

#### `use_item`

- `item_id`

## 5. Preconditions

### 5.1 字段

- `type`
- `actor_node_id`
- `target_node_id`
- `scene_node_id`
- `item_id`
- `task_id`
- `expected`
- `args`

### 5.2 第一版支持的 precondition 类型

- `same_scene`
- `target_present`
- `item_present`
- `money_at_least`
- `task_status`
- `scene_flag`
- `location_accessible`

### 5.3 示例

```json
{
  "type": "item_present",
  "actor_node_id": "player_001",
  "item_id": "knife_bloody"
}
```

## 6. MissingFacts

### 6.1 字段

- `type`
- `node_id`
- `item_id`
- `task_id`
- `reason`

### 6.2 推荐的缺失事实类型

- `player_location`
- `target_location`
- `item_presence`
- `scene_state`
- `task_state`
- `wallet_state`

## 7. SuggestedInteraction

### 7.1 字段

- `mode`
- `event_type`
- `audience_scope`
- `target_node_id`

### 7.2 推荐映射

- `speech` -> `direct_dialogue` / `group_chat`
- `show_item` -> `direct_dialogue + event=show_item`
- `gift` -> `gift_response + event=gift`
- `trade_request` -> `trade_dialogue + event=trade_request`
- `threaten` -> `direct_dialogue + event=threaten`

## 8. Validator 输出建议

```json
{
  "status": "accepted",
  "steps": [
    {"index": 0, "status": "accepted"},
    {"index": 1, "status": "accepted"}
  ]
}
```

状态建议：

- `accepted`
- `rejected`
- `partially_accepted`

## 9. Executor 输出建议

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

## 10. 第一版边界

第一版 schema 不承诺：

- 任意自由文本都能可靠映射为复杂世界操作
- 多 NPC 并行响应
- 无校验直接落地复合动作

第一版 schema 重点解决的是：

- 玩家自然语言 -> 结构化提案
- 结构化提案 -> 游戏侧权威校验
- 校验通过 -> interaction bridge
