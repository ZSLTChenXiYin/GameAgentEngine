# 外部交互路线图

**中文**

本文档固化 Engine 在“异步接口、游戏端交互、定时调度、自主行动恢复链路”上的统一架构目标、当前实现进度与后续开发阶段。

目标是同时支持以下两大能力，并允许开发者按场景自由选择：

- Engine 内建 game client adapter，Engine 直接向游戏端发起 `HTTP` / `WebSocket` / `RPC` 请求。
- 任务拉取接口，外部 bridge、dispatcher 或游戏客户端主动从 Engine 拉取待执行任务。

这两种方案都必须保留，同时支持混合方案；不能强制所有开发者只使用其中一种。

---

## 一、问题边界

围绕自主行为与外部系统交互，当前需要统一解决四类问题：

| 维度 | 说明 |
|---|---|
| 交互类型 | 区分 `external_action` 与 `external_query` |
| 投递方式 | 支持 `push`、`pull`、`hybrid` |
| 完成方式 | 支持同步完成、异步完成、callback 回填、暂停后恢复 |
| 恢复语义 | 区分“仅记录结果”与“恢复原始推理链路继续跑” |

需要特别明确的一点：

- `callback` 在当前设计里是“外部完成结果回填到 Engine”的入站完成机制，不等同于“Engine 向游戏端发请求”的出站投递机制。
- 因此，`callback` 可以和 `push` 或 `pull` 任意组合使用。
- 定时调度模式不是游戏端主动发起调用，这并不妨碍它走 callback 链路；Engine 可以先通过 `push` 或把任务放入 `pull` 队列，等游戏端或 bridge 完成后再回调 `POST /api/v1/actions/callback`。

---

## 二、统一架构目标

### 2.1 外部交互分类

未来所有需要离开 Engine 边界的能力，都统一收敛为外部交互接口：

| 分类 | 说明 | 典型例子 |
|---|---|---|
| `external_query` | 向外部系统请求数据，再继续后续推理 | 查询游戏场景状态、角色实时位置、战斗上下文 |
| `external_action` | 请求外部系统执行动作，可能只记录结果，也可能影响后续业务链路 | 播放动画、生成实体、切换场景、驱动任务系统 |

### 2.2 投递策略模型

每个外部接口都不再只用“是否异步”描述，而是要具备显式投递策略：

| 字段 | 说明 |
|---|---|
| `delivery_mode` | `push` / `pull` / `hybrid` |
| `primary_transport` | 主投递通道，例如 `http_adapter`、`websocket_adapter`、`task_pull` |
| `fallback_transport` | `hybrid` 下的降级通道 |
| `consumer` | 任务消费方，例如 `bridge`、`game_client`、`server_worker` |
| `resume_policy` | 完成后是否恢复原执行，例如 `none`、`resume_paused_execution` |
| `timeout_ms` | 超时策略 |
| `retry_policy` | 重试规则 |
| `idempotency_key_strategy` | 幂等键生成策略 |

### 2.3 两层配置结构

建议将外部交互配置拆成两层：

| 层级 | 职责 |
|---|---|
| `external_integrations` | 描述接入方、凭证、基础连接、通道能力 |
| `external_interfaces` | 描述某个动作/查询接口的业务语义、投递方式、恢复策略 |

示意：

```yaml
external_integrations:
  game_http:
    type: http_adapter
    base_url: http://127.0.0.1:9000
    auth:
      mode: bearer
      token: ${GAME_TOKEN}

  local_bridge_pull:
    type: task_pull
    consumer: bridge

external_interfaces:
  fetch_scene_runtime:
    category: external_query
    delivery_mode: push
    primary_transport: game_http
    resume_policy: resume_paused_execution
    timeout_ms: 10000

  spawn_npc_runtime:
    category: external_action
    delivery_mode: hybrid
    primary_transport: game_http
    fallback_transport: local_bridge_pull
    consumer: game_client
    resume_policy: none
```

---

## 三、当前已实现能力

当前仓库已经完成的内容如下。

### 3.1 已完成

| 状态 | 能力 | 说明 |
|---|---|---|
| 已完成 | 异步 callback 持久化 | 持久化 `async_callback_records` |
| 已完成 | paused execution 持久化 | 持久化恢复所需的执行快照 |
| 已完成 | `game_client request_data` 暂停 | 多轮推理在请求外部数据时可暂停 |
| 已完成 | callback 自动恢复 | `POST /api/v1/actions/callback` 成功回填后自动恢复原始执行 |
| 已完成 | 恢复结果回注入上下文 | callback 结果可进入下一轮 LLM 推理 |
| 已完成 | 普通异步 action 回调处理 | 普通异步动作会更新 callback 记录并触发 `OnResult(...)` |
| 已完成 | 恢复链路测试 | 覆盖 engine 层与 API 层自动恢复测试 |
| 已完成 | callback / pipeline 文档 | 现有 API、SDK、Pipeline 文档已覆盖恢复行为 |
| 已完成 | runtime task 持久化模型 | 已引入统一 `runtime_tasks` 队列模型 |
| 已完成 | pull 基础 API | 已提供 `pending / claim / start / heartbeat / release` 五个接口 |
| 已完成 | pull 基础测试 | 覆盖 store 层与 API 层的核心领取链路 |
| 已完成 | game_client request_data 真实入队 | 暂停执行时会自动生成可拉取的 runtime task |
| 已完成 | runtime task callback 完成态回写 | callback 回填后会同步更新 runtime task 成功/失败状态 |
| 已完成 | 普通 async action 真实入队 | 普通异步动作也会同步生成 `external_action` 类型 runtime task |
| 已完成 | `running` 状态迁移基础能力 | 已支持 claim 后显式进入 `running` |
| 已完成 | `heartbeat_timeout` 基础回收能力 | 已支持将陈旧 claimed/running task 标记为 heartbeat_timeout |

### 3.2 当前真实边界

| 状态 | 能力 | 当前情况 |
|---|---|---|
| 部分完成 | callback 作为入站完成机制 | 已实现 |
| 未完成 | `push` 出站投递到游戏端 | 还没有内建 `http_adapter` / `websocket_adapter` / `rpc_adapter` |
| 部分完成 | `pull` 任务拉取接口 | 已有统一 runtime task queue API，并已接入 game_client request_data 与普通 async action 的真实生产及 callback 完成态闭环 |
| 未完成 | `hybrid` 自动降级 | 还没有策略编排 |
| 未完成 | 定时调度下主动出站调用 | 当前只完成暂停/恢复基础链路，未完成实际出站派发 |
| 未完成 | 普通异步 action 的 richer business resume | 目前主要是 `OnResult(...)`，还没有统一的后续编排策略 |

这意味着当前系统已经具备“结果回填并恢复推理”的下半段能力，但还缺“如何把任务稳定送到游戏端或 bridge”的上半段能力。

---

## 四、为什么要同时保留 push / pull / hybrid

| 方案 | 优势 | 风险/局限 | 适用场景 |
|---|---|---|---|
| `push` | Engine 直连快，链路短，低延迟 | 游戏端需暴露服务，连通性与鉴权复杂 | 受控局域网、服务端游戏逻辑、联机后端 |
| `pull` | 游戏端或 bridge 主动拉取，更适合客户端网络环境，失败恢复简单 | 存在轮询或 claim 协议复杂度 | 本地编辑器、单机客户端、复杂 NAT 环境 |
| `hybrid` | 主链路快，失败可降级，更稳健 | 配置和观测复杂度更高 | 生产环境、不同平台混布、需要高可用 |

稳定性上没有绝对唯一答案：

- 如果 Engine 与游戏逻辑服务处于同一受控网络，`push` 的执行链路通常更短。
- 如果游戏端更像一个会间歇在线、网络受限、不能长期开放入站端口的客户端，`pull` 往往更稳。
- 如果产品既有服务端逻辑又有客户端执行面，`hybrid` 会是更现实的生产方案。

因此 Engine 需要把它们设计成可选策略，而不是单一架构押注。

---

## 五、统一任务状态机目标

未来需要把外部交互收敛到统一状态机，避免各类异步逻辑散落在不同分支里。

建议状态：

| 状态 | 说明 |
|---|---|
| `pending` | 已生成任务，等待投递或等待被拉取 |
| `dispatched` | 已通过 `push` 发出 |
| `claimed` | 已被某个 pull consumer 领取 |
| `running` | 外部系统确认执行中 |
| `heartbeat_timeout` | claim 后长时间无心跳 |
| `succeeded` | 执行成功 |
| `failed` | 执行失败 |
| `released` | 被 consumer 放回队列 |
| `cancelled` | 被内部取消 |
| `resumed` | 对应 paused execution 已恢复 |

普通异步动作与 `game_client request_data` 的关键区别要在状态机上显式体现：

- 普通异步动作完成后，通常进入业务回调处理，不一定恢复当前推理现场。
- `request_data.target = game_client` 对应的是“补数据后继续想”，因此完成后需要走 `resume_paused_execution`。

---

## 六、上下文压缩不可丢信息的要求

为了保证长链路、重试、服务重启、上下文压缩后仍可继续执行，恢复所需信息必须以持久化快照为准，而不是只依赖当前会话上下文。

至少要确保以下信息可恢复：

- 原始 `InvokeRequest`
- `BuiltContext` 快照
- 当前 round state
- 已累积补充上下文
- 暂停时的 `request_data`
- 外部交互接口标识与投递策略
- `callback_id` / task id / execution id
- 恢复策略与幂等信息
- 最近一次外部回填结果或错误

当前已完成前五项的核心持久化；后续要把接口策略、任务态、consumer 领取态也一起纳入持久化边界。

---

## 七、分阶段实施计划

### 阶段 P0：恢复链路基础设施

目标：让 `game_client request_data` 支持暂停、回填、自动恢复。

| 子项 | 状态 |
|---|---|
| callback 记录持久化 | 已完成 |
| paused execution 快照持久化 | 已完成 |
| callback API 自动恢复 | 已完成 |
| 自动恢复测试 | 已完成 |
| 文档说明 | 已完成 |

### 阶段 P1：任务拉取接口基础设施

目标：建立统一 runtime task queue，支撑 `pull` 方案，并为 `hybrid` 打底。

计划内容：

| 子项 | 状态 |
|---|---|
| runtime task 数据模型 | 已完成 |
| 任务持久化 store 能力 | 已完成 |
| `GET /api/v1/runtime/tasks/pending` | 已完成 |
| `POST /api/v1/runtime/tasks/claim` | 已完成 |
| `POST /api/v1/runtime/tasks/start` | 已完成 |
| `POST /api/v1/runtime/tasks/heartbeat` | 已完成 |
| `POST /api/v1/runtime/tasks/release` | 已完成 |
| pull 基础测试 | 已完成 |
| runtime task 与真实外部任务生产对接 | 已完成（当前先覆盖 game_client request_data） |
| task 成功/失败完成态闭环 | 已完成（当前先覆盖 callback 驱动的 game_client request_data） |
| 普通 async action 接入 runtime task queue | 已完成 |
| 更细粒度 task 状态迁移（如 running / heartbeat_timeout） | 已完成基础能力 |
| heartbeat_timeout 后续回收/重派策略 | 未开始 |

### 阶段 P2：内建 push adapter

目标：让 Engine 可以主动向游戏端投递外部交互。

计划内容：

| 子项 | 状态 |
|---|---|
| adapter 抽象层 | 未开始 |
| `http_adapter` | 未开始 |
| `websocket_adapter` | 未开始 |
| `rpc_adapter` | 未开始 |
| dispatch 失败重试与幂等 | 未开始 |
| push 基础观测 | 未开始 |

### 阶段 P3：hybrid 与策略编排

目标：支持主通道失败后自动降级，并统一恢复策略。

计划内容：

| 子项 | 状态 |
|---|---|
| `delivery_mode` / `primary_transport` / `fallback_transport` | 未开始 |
| `consumer` 路由策略 | 未开始 |
| `resume_policy` 扩展 | 未开始 |
| 普通 async action 的统一后处理编排 | 未开始 |
| hybrid fallback 状态迁移 | 未开始 |

### 阶段 P4：安全、观测、管理能力

目标：补齐生产可用性。

计划内容：

| 子项 | 状态 |
|---|---|
| callback / task claim 鉴权模型 | 未开始 |
| 幂等键与防重放 | 未开始 |
| external task metrics | 未开始 |
| admin / management endpoints | 未开始 |
| 故障注入与测试矩阵 | 未开始 |
| 开发者文档与示例 | 未开始 |

---

## 八、当前建议的开发顺序

基于当前代码状态，推荐按以下顺序继续实施：

1. 先完成 `pull` 任务队列与 API 基础设施。
2. 再补首个内建 `push` 实现，优先 `http_adapter`。
3. 再把配置层升级到 `external_integrations` + `external_interfaces` 双层模型。
4. 最后补 `hybrid` 自动降级、安全、观测和管理接口。

这样排序的原因：

- 当前系统已经具备 callback 恢复下半段，先补 `pull` 可以最快闭合一条完整的外部执行链路。
- `pull` 方案天然适配 bridge 和游戏端主动取任务，也更容易与定时调度模式结合。
- 等 queue 与状态机稳定后，再加入 `push` adapter，架构会更清晰，不容易返工。

---

## 九、与当前代码实现的对应关系

当前可以把现状理解为：

| 能力层 | 当前状态 |
|---|---|
| 推理循环中的 `request_data` | 已支持 |
| `game_client` 数据请求暂停 | 已支持 |
| callback 入站结果回填 | 已支持 |
| paused execution 自动恢复 | 已支持 |
| runtime task queue | 已支持基础队列、拉取接口，以及 game_client request_data / 普通 async action 真实任务生产 |
| built-in push adapter | 尚未支持 |
| scheduled 自主行动真实出站派发 | 尚未支持 |
| hybrid 策略与 consumer 路由 | 尚未支持 |

因此，定时调度模式下如果未来要触发异步游戏接口，完整链路应该是以下两种之一：

### 方案 A：push

1. 定时调度触发自主行动。
2. Engine 决定调用某个 `external_query` 或 `external_action`。
3. Engine 通过内建 adapter 主动发往游戏端。
4. 游戏端完成后回调 `POST /api/v1/actions/callback`。
5. 若该调用属于 paused execution，则 Engine 自动恢复。

### 方案 B：pull

1. 定时调度触发自主行动。
2. Engine 生成 runtime task 并放入待领取队列。
3. bridge 或游戏端主动拉取并 claim。
4. 外部执行完成后回调 `POST /api/v1/actions/callback`。
5. 若该调用属于 paused execution，则 Engine 自动恢复。

所以，定时调度并不会阻止 callback；它只会改变“任务如何离开 Engine”的上游投递方式。

当前这一阶段已经具备第 2、3 步所需的基础队列与领取协议，并且 `game_client request_data` 已经能够真实入队；外部完成后也能通过 callback 驱动 task 完成态回写。

后续还需要继续补的内容是：

- 为 `heartbeat_timeout` 补自动回收、重派或人工介入策略
- 在 scheduled 自主行为里把更多外部交互统一走 external interface 配置层，而不是只覆盖当前的 `game_client request_data`

当前普通 async action 的 consumer 路由仍然是过渡方案：默认使用 `bridge`，并允许通过动作参数显式覆盖；后续仍需要升级到 `external_interfaces` 配置层驱动的正式路由策略。

---

## 十、文档维护规则

后续每完成一个阶段，必须同步更新本文档中的：

- 当前已实现能力
- 当前真实边界
- 分阶段实施计划状态
- 与当前代码实现的对应关系

这样即使发生上下文压缩，也能依赖仓库文档快速恢复完整架构意图与阶段进度。
