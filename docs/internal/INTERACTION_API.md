# 交互接口设计

本文定义基于现有 `invoke` 执行内核扩展出的交互语义模型。

## 1. 设计目标

当前 Engine 的推理入口是单焦点节点模型。交互接口的目标不是替换 `invoke`，而是在 `invoke` 请求上增加正式的扮演语义：

- 谁在说话
- 对谁说
- 在哪里说
- 是单聊、群聊还是结构化事件反馈

## 2. 对外分层

- 执行内核：仍然使用 `InvokeRequest -> Pipeline.Execute(...)`
- 业务 API：允许新增 `interaction/*` 路由或 SDK 包装方法

换句话说：

- `invoke` 是统一执行协议
- `interaction/*` 是更友好的角色交互入口

## 3. InteractionContext

建议在 `InvokeContext` 下新增：

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

## 4. 推荐字段

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

## 5. 事件类型

第一版建议支持：

- `speech`
- `gift`
- `show_item`
- `trade_request`
- `threaten`

示例：

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

## 6. 上下文装配原则

- `target_node_id` 是主视角节点
- `speaker_node_id` 是正式参与方
- `scene_node_id` 提供共享场景上下文
- `participant_node_ids` 中除主目标外，其余节点只注入轻量摘要
- 其他参与者不应默认注入完整组件、记忆和全关系图

## 7. 群聊策略

第一版群聊不做多节点并行推理。

推荐策略：

- 游戏侧维护 room 权威状态
- worker 决定谁先回应
- Engine 一次只扮演一个 NPC
- 其他参与者仅以房间上下文摘要方式进入 prompt

## 8. play 模式映射

- `/+talk innkeeper`
  - `mode=direct_dialogue`
- `直接输入文本`（在已选择对话目标时）
  - `mode=direct_dialogue`
- `/+say 大家今晚都看见了什么？`
  - `mode=group_chat`
- `/+ask guard 你看见谁进门了？`
  - `mode=group_chat`，但 `target_node_id=guard`
- `/+gift innkeeper silver_ring`
  - 游戏侧先落地，再调用 `mode=gift_response`

正式文档面现在统一使用 `/+cmd + 参数` 风格。旧写法 `/talk`、`/ask` 等别名仅作为兼容输入存在。
