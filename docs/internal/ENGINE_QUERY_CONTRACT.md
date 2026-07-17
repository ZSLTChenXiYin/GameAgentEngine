# Engine 到游戏侧查询契约

**中文** | [**English**](./ENGINE_QUERY_CONTRACT_EN.md)

本文定义 Engine 通过 `game_client` 异步接口向 `gameagentworker` 查询权威数据的契约。

## 1. 查询原则

- 查询只在 Engine 缺失关键权威信息时触发。
- 查询结果由游戏侧返回，Engine 通过 callback/resume 继续推理。
- 查询接口返回的是当前权威事实，而不是建议动作。

## 2. 查询目标

统一使用：

- `target = "game_client"`

默认外部接口名：

- `game_client_request_data`

后续允许通过 request-scoped dynamic interface 暴露更细化的查询接口别名。

## 3. 推荐查询类型

第一版建议标准化支持以下 query types：

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

其中：

- `relationship_hint` 仅用于从游戏侧补充权威规则状态，不替代 Engine 的长期关系图。

## 4. 请求示例

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

## 5. 回调结果示例

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

## 6. 安全与粒度限制

- 不允许 Engine 无限制拉取整个本地存档。
- 单次查询应聚焦当前回合所需的最小事实。
- 高敏感内部脚本字段、隐藏标志、掉率表等默认不开放。
- 查询类型采用白名单，拒绝临时自由拼接底层字段。

## 7. play 模式要求

- `/+gift`、`/+show_item`、`/+trade` 以及 `/+act` 解释出的高风险结构化动作，应先由游戏侧落地。
- Engine 如需确认相关权威事实，仍可通过本契约回查。
- 群聊模式下允许查询 `room_state`，但房间发言顺序仍由游戏侧控制。
