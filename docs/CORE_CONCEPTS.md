# 核心概念

**中文** | [**English**](./CORE_CONCEPTS_EN.md)

GameAgentEngine 以节点、组件、记忆、关系四类基础抽象建模游戏世界，并在其上提供推理管线、连续性状态和世界时间推进能力。

---

## 节点

常见节点类型：

- `world`
- `faction`
- `location`
- `npc`
- `item`
- `quest_line`
- `event`

`world` 节点是一个世界的根节点。当前零起步流程就是先创建它。

---

## 组件

常见内置组件：

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

`world_time_state` 是 Engine 运行时维护的状态组件，用于保存当前世界时间结果。

---

## 记忆

常见层级：

- `short_term`
- `long_term`
- `shared`
- `world`

当前传播模式：

- `upward`
- `environment_scope`
- `organization_scope`
- `tag_broadcast`
- `targeted`
- `manual`

---

## 关系

常见关系类型：

- `belongs_to`
- `ally`
- `enemy`
- `subordinate`
- `kinship`
- `located_at`
- `external_parent`

当前语义约定：

- `parent_id` 表示主层级和稳定归属链
- `located_at` 表示当前位置
- `belongs_to` / `subordinate` 表示组织归属或控制链
- `external_parent` 仅作辅助作用域挂接

---

## world_settings、world_time_settings、world_time_state

这是当前最重要的一组概念：

- `world_settings`：世界级运行配置容器
- `world_time_settings`：时间规则
- `world_time_state`：时间结果

两者没有职责重叠，但前者决定后者如何生成。

---

## 管线模式

每个世界可独立选择：

- `vertical`
- `polling`
- `full`

它属于 `world_settings`，不是静态配置里的 `execution_mode`。
