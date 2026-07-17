# Engine 内核化备忘录

**中文** | [**English**](./ENGINE_KERNELIZATION_MEMO_EN.md)

本文记录当前关于未来 Engine 内核化工作的架构判断。

它不是现在立刻启动实现的承诺。
它的目的，是在 Engine 功能集还没有成熟到足以真正实施之前，先把设计方向和边界决策保留下来，避免遗失。

## 1. 当前结论

GameAgentEngine 适合在未来推进内核化工作。

正确目标不是：

- 直接把当前 HTTP 服务编译成 DLL / so / dylib，并把它当成最终架构

正确目标是：

- 把 Engine 重构成与宿主无关的运行时核心
- 让宿主进程拥有通信、持久化和 LLM 调用路径
- 让 Engine 聚焦语义运行时、编排、连续性和执行规则

简而言之，目标是把 Engine 从“服务拥有的运行时”，变成“可嵌入的推理内核”。

## 2. 为什么当前仓库具备这个可行性

当前仓库已经有相对清晰的分层。

- `internal/engine`：推理管线、上下文构建、编排、world tick 连续性、action 与 memory 流
- `internal/service`：世界管理、事务工作流、持久化侧编排
- `internal/store`：数据库访问与持久化实现
- `internal/api`：HTTP 暴露层
- `internal/llm`：provider 侧 LLM 访问
- `internal/external`：外部派发适配器
- `cmd/gameagentengine`：进程引导与服务装配

这意味着未来的内核化主要是边界与所有权重构，而不是从零重写。

## 3. 主要动机

内核化的价值不只是把 HTTP 换成本地动态库调用。

### 3.1 通信从服务 IO 转向进程内 ABI

预期方向是：

- 把进程间 JSON request/response 交换，替换为进程内二进制消息交换
- 绕过 HTTP 路由、中间件以及 JSON 编解码在热路径上的开销
- 减少以字符串为主的大载荷拼装与重复对象重建

需要强调的是：

收益并不只是来自把 JSON 换成二进制。
收益还取决于是否把宿主-内核边界设计得足够粗粒度。

如果未来边界仍然暴露大量细碎的 CRUD 风格调用，那么系统只是把很多小 JSON 往返替换成很多小 ABI 往返，收益会被大幅稀释。

### 3.2 持久化从 Engine 自有 CRUD 转向宿主自有状态桥接

预期方向是：

- Engine 停止默认拥有数据库访问权
- Engine 声明它需要什么状态，以及它会产出什么状态变更
- 宿主进程决定数据从哪里来、如何缓存、如何持久化

这样宿主可以按照自己的存储模型桥接状态，例如：

- 内存运行时状态
- SQLite
- 自定义存档系统
- 基于 ECS 的状态
- 平台特定的权威服务

同样需要强调：

未来设计不应演变成宿主和内核之间的高频按需状态 RPC 系统。
更推荐的主路径是：

- snapshot-in
- patch-out

也就是说，由宿主向内核送入足够完整的状态切片或运行时快照，再由内核输出 patch、effect、plan、memory update 和待恢复状态。

按需桥接调用应作为补充，而不应成为主执行路径。

### 3.3 LLM 调用从 Engine 自有 provider 流转向宿主自有推理管线

预期方向是：

- Engine 准备 prompt、上下文、契约和预期结构化输出
- 宿主进程真正执行模型调用
- 宿主使用自己的网络路径、推理中间件、连接池、内部路由、专线以及 fallback 策略

对于宿主已经拥有私有推理链路或内部高速服务路径的环境，这一点尤其重要。

正确抽象不是“Engine 拥有模型调用”。
更正确的抽象更接近：

- Engine 产出 inference specification
- host 执行它
- host 返回结构化结果信封

这会降低部署耦合，在很多场景下，带来的真实时延改善甚至比单纯把 HTTP 服务本地化还更明显。

## 4. 目标所有权模型

### 4.1 当前倾向

当前系统更接近这样的所有权模型：

- Engine 拥有 API 暴露
- Engine 拥有持久化路径
- Engine 拥有 LLM provider 路径
- 外部系统围绕 Engine 服务集成

### 4.2 未来内核化倾向

未来的内核化模型应反转这种所有权：

- host 拥有 IO
- host 拥有持久化实现
- host 拥有 LLM 管线接入
- Engine 只拥有语义运行时规则和执行逻辑

这种反转才是内核化的核心。
动态库形态只是其中一种打包形式。

## 5. 哪些能力应保留在内核中

未来内核应保留 Engine 真正有产品价值的部分：

- 世界语义模型
- 上下文构建与裁剪规则
- 关系与记忆选择逻辑
- pipeline mode 与多轮编排
- 结构化输出解析与校验
- action planning 与执行规则
- world tick 连续性与叙事状态迁移
- pending continuation 与 resume 状态机
- 可嵌入的交互契约

## 6. 哪些能力应变成宿主桥接

以下能力应逐步转向显式的宿主桥接接口：

- LLM 调用
- 状态读写
- 外部 action 执行
- runtime task 派发
- scheduler 与 clock 访问
- logging、metrics 与 trace sink
- lock 与 transaction policy

这些能力不应继续隐藏在“默认是 server runtime”的假设中。

## 7. 宿主-内核边界应偏好的形态

未来边界应优先暴露粗粒度的语义操作，而不是细碎的持久化式调用。

正确形态示例：

- `invoke_dialogue`
- `advance_world_tick`
- `apply_world_snapshot`
- `run_inference_round`
- `resume_pending_effect`
- `commit_state_patch`

错误的默认形态示例：

- 大量 per-field 更新
- 大量 per-component 拉取
- relation-by-relation 远程访问
- 正常执行期内 memory-by-memory 的宿主回调

内核应消费结构化运行时状态，并返回结构化运行时结果。

## 8. 建议的消息模型方向

未来接口应逐步脱离当前 HTTP DTO 或数据库模型，不应把它们直接暴露成稳定的内核契约。

更稳定的方向应围绕这些运行时消息展开：

- `WorldSnapshot`
- `RuntimeContext`
- `InferenceSpec`
- `InferenceResult`
- `StatePatch`
- `ActionEffect`
- `PendingContinuation`

这些名称只是方向性建议，不是最终 API 承诺。

## 9. 主要预期收益

如果实现得当，内核化可以带来这些收益：

- 热路径通信开销更低
- 与引擎侧权威系统和存档系统对齐得更自然
- 可直接复用宿主已有的 LLM 基础设施与私有服务链路
- 更容易在 Unity、Unreal Engine 和 Godot 等不同宿主间复用同一运行时核心
- 在嵌入式场景中降低部署摩擦

## 10. 主要风险与约束

内核化是可行的，但它绝不是低成本工作。

### 10.1 接口设计风险

最大的风险，是把宿主-内核协议设计错。

如果未来契约过于细粒度，项目会以另一种形式保留大部分当前边界成本。
如果契约只是简单镜像当前 HTTP DTO 或 GORM 模型，那么结果会是“嵌入式服务”，而不是真正的内核。

### 10.2 可观测性回退风险

一旦 LLM 调用和持久化所有权转移到宿主，内核将不再自动拥有以下完整链路追踪：

- request path
- retry behavior
- latency
- token usage
- provider fallback
- environment-specific failures

这意味着宿主必须在推理结果和状态应用结果中返回足够的元数据，才能支撑有意义的诊断。

### 10.3 运行时模型迁移成本

当前围绕服务模型已经内建了不少便利能力，例如：

- 自动迁移
- callback 持久化
- runtime task 恢复
- 中央日志汇聚
- world lock 处理
- database retry policy

这些能力不会消失，但它们不再是 Engine 的隐式所有权，而会变成宿主显式桥接策略的一部分。

### 10.4 多引擎打包风险

即使运行时核心设计得很好，Unity、Unreal Engine 和 Godot 依然需要各自独立的宿主绑定和生命周期集成。
这项工作应被视为后续阶段，而不是第一阶段成功的定义。

## 11. 建议时序

内核化不应立即启动。

当前正确时序是：

1. 先继续打磨 Engine 内核功能和契约
2. 继续稳定 interaction、intent、continuity、callback 和 runtime task 语义
3. 等这些契约足够成熟后，再启动内核化实现
4. 真正开始实现时，先验证宿主无关的运行时边界，再做多引擎绑定

第一个真正的实现里程碑，应是本地嵌入式运行时边界的 proof of concept，而不是立刻推进 Unity + UE + Godot 的生产级绑定。

## 12. 实施前检查点

在内核化开始前，项目至少应重新审视以下问题：

- 当前真正的主瓶颈时延是什么？
- 哪些调用是真正高频的？
- 哪些状态必须保持宿主权威？
- 哪些状态应该继续保持 Engine 权威？
- 哪些流程必须同步，哪些可以继续做成可恢复或延迟执行？
- 对于推理与持久化结果，宿主必须返回哪些可观测字段？

如果这些问题仍不清楚，就不应开始实现。

## 13. 当前最终决定

当前的固定结论是：

- 先把本文保留为设计指导
- 暂不启动内核化实现
- 等 Engine 功能和契约足够成熟后，再回到本文

到那时，项目再把本文转化为具体的架构与实施计划。
