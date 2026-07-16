# 玩家自然语言输入管线

本文定义玩家自然语言输入在 GameAgentEngine 中的执行边界。

## 1. 目标

玩家自然语言输入的目标不是直接修改世界真值，而是通过玩家镜像节点完成：

- 意图理解
- 缺失事实识别
- 权威数据查询请求
- 结构化行为提案生成

最终状态是否落地，必须由游戏侧权威系统决定。

## 2. 设计原则

### 2.1 玩家节点只负责提案，不负责定案

玩家镜像节点可以像普通 Agent 一样推理玩家输入，但只能输出：

- `intent`
- `preconditions`
- `missing_facts`
- `risk_level`

不能把任何未校验的动作视为已经成功执行。

### 2.2 游戏侧负责权威校验与执行

游戏侧/worker 负责：

- 场景邻近性校验
- 物品持有校验
- 金钱/HP/背包/任务状态校验
- 执行成功后的真值修改
- 将执行结果桥接回 interaction invoke

### 2.3 解释链路复用 invoke

玩家输入解释不另起新的执行核心。

推荐方式：

- 对外暴露友好入口：`/api/v1/player/input/interpret`
- 内部仍构造 `InvokeRequest`
- `task_type=custom`
- `node_id=player mirror node`

### 2.4 对话输入和行为输入分开

为了避免把“说话”和“行动”混为一谈：

- 普通对话仍走 `npc_dialogue` / interaction
- 行为型自然语言走 player input interpret
- play 模式中建议通过 `/act` 或 `/do` 显式进入行为解释链

## 3. 总体执行链

```text
玩家自然语言输入
    -> player input interpret
    -> Engine 产出 PlayerIntent
    -> 游戏侧 validator
    -> 游戏侧 executor
    -> interaction bridge
    -> NPC / 群聊 / 场景响应
```

## 4. 分层职责

### 4.1 Engine

负责：

- 玩家输入语义拆解
- 复合动作分步
- 缺失事实识别
- request_data 查询请求
- 结构化 intent proposal 输出

不负责：

- 最终状态落地
- 非法动作的权威拒绝
- 事务回滚

### 4.2 游戏侧 / worker

负责：

- 权威状态读取
- validator
- executor
- interaction bridge

### 4.3 play / REPL

负责：

- 开发者输入入口
- 调试输出
- authority state 文件驱动试玩

## 5. 输入分类

### 5.1 对话型输入

例子：

- “老板，今晚谁最后一个从码头回来？”
- “你刚才看见谁进门了？”

特征：

- 主要是说话
- 一般不直接引发高风险真值变更

推荐：

- 直接走 interaction invoke

### 5.2 行为型输入

例子：

- “我把沾血的短刀拍在柜台上，问老板今晚有没有见过这把刀的主人”
- “我把银戒指塞给老板，试探她会不会松口”
- “我转身往后门走，同时示意守卫别出声”

特征：

- 包含动作步骤
- 需要权威状态验证
- 可能是复合动作

推荐：

- 先走 player input interpret

## 6. 复合动作约束

复合动作必须拆成结构化 steps。

例如：

`我把沾血的短刀拍在柜台上，问老板今晚有没有见过这把刀的主人`

不应直接当作已经成功的叙事事实，而应拆成：

1. `show_item`
2. `speech`

如果第一步权威校验失败，则第二步不能建立在“已经展示了刀”的错误前提之上。

## 7. 第一版支持范围

建议第一版支持的 intent / step 类型：

- `speech`
- `show_item`
- `gift`
- `trade_request`
- `threaten`
- `move`
- `inspect`
- `use_item`
- `composite`

## 8. 第一版暂不支持

第一版不建议直接支持：

- 无边界自由世界修改
- 未建模动作直接落地
- 多 NPC 并行链式推理
- 把玩家输入视为既成事实写入世界

## 9. 与 interaction 的关系

player input interpret 不是 interaction 的替代品，而是 interaction 的前置层。

典型桥接关系：

- `speech` -> `direct_dialogue` 或 `group_chat`
- `show_item` -> `event=show_item`
- `gift` -> `gift_response`
- `trade_request` -> `trade_dialogue`
- `threaten` -> `event=threaten`

## 10. 核心结论

玩家节点可以承担“玩家自然语言的行为推理”，但不能承担“世界真值的最终写入”。

工程上必须坚持：

- Engine 负责提案
- 游戏侧负责定案

