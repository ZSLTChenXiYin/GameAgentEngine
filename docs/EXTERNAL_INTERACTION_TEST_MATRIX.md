# 外部交互测试矩阵

**中文**

本文档固化 Engine 外部交互链路当前已经纳入自动化回归的关键场景，作为 Stage 4 的测试矩阵基线。

它的目标不是替代测试代码，而是让开发者和运维能快速知道：哪些路径已经有自动化保障，哪些仍属于后续增强。

---

## 一、矩阵维度

当前矩阵围绕这些维度组织：

| 维度 | 说明 |
|---|---|
| 投递模式 | `push` / `pull` / `hybrid` |
| 完成方式 | callback 完成、自动恢复、仅记录结果 |
| 治理路径 | `max_attempts`、`heartbeat_timeout`、auto requeue |
| 稳定性问题 | 配置漂移、callback replay、重复 timeout |
| 排障能力 | 诊断视图与扩展统计 |

---

## 二、自动化覆盖矩阵

| 场景 | 当前状态 | 主要覆盖点 | 代表测试 |
|---|---|---|---|
| `push` 成功派发 | 已覆盖 | task 进入 `dispatched`，记录 transport / attempts / idempotency | `internal/engine/pipeline_test.go` |
| `pull` 待领取入队 | 已覆盖 | task 保持 `pending`，支持 claim/start/heartbeat/release | `internal/api/runtime_tasks_test.go`、`internal/store/runtime_tasks_test.go` |
| `hybrid` push 失败后回退 | 已覆盖 | failure class / decision 持久化，task 转 `released` | `internal/engine/pipeline_test.go`、`internal/store/runtime_tasks_test.go` |
| `max_attempts` 终态阈值 | 已覆盖 | release/requeue 超限后进入 `failed` | `internal/store/runtime_tasks_test.go` |
| `heartbeat_timeout` 标记 | 已覆盖 | stale claimed/running 转 `heartbeat_timeout` | `internal/store/runtime_tasks_test.go`、`internal/service/runtime_task_governor_test.go` |
| task 级 heartbeat timeout 策略快照 | 已覆盖 | auto requeue 是否开启、延迟和原因按 payload 快照执行 | `internal/service/runtime_task_governor_test.go`、`internal/engine/pipeline_test.go` |
| 普通 async action callback 完成态 | 已覆盖 | callback 后 runtime task 进入 `succeeded`/`failed` | `internal/api/debug_handlers_test.go` |
| paused execution callback 自动恢复 | 已覆盖 | callback 回填后自动恢复原推理现场 | `internal/api/debug_handlers_test.go`、`internal/engine/pipeline_test.go` |
| `resume_policy = none` 不自动恢复 | 已覆盖 | callback 成功但原执行保持 `paused` | `internal/api/debug_handlers_test.go` |
| callback 后处理快照 | 已覆盖 | `write_memory` 按 payload 快照执行，不受后续配置漂移影响 | `internal/api/debug_handlers_test.go` |
| `resume_policy` 配置漂移后仍按快照恢复 | 已覆盖 | task 创建后把接口配置改成 `none`，callback 仍按原快照恢复 | `internal/api/debug_handlers_test.go` |
| callback replay 不重复恢复 | 已覆盖 | 同一 `X-Callback-Request-Id` 重放时不重复触发恢复链路 | `internal/api/debug_handlers_test.go`、`internal/api/middleware_test.go` |
| 重复 heartbeat timeout 计数 | 已覆盖 | task 经 requeue 再次 timeout 时 `heartbeat_timeout_count` 累加 | `internal/store/runtime_tasks_test.go` |
| 诊断视图筛选 | 已覆盖 | `retry_exhausted`、`stale_dispatched`、`repeated_timeout` 视图与查询参数 | `internal/api/runtime_tasks_test.go`、`internal/store/runtime_tasks_test.go` |
| 诊断聚合统计 | 已覆盖 | `retry_exhausted_tasks`、`dispatched_without_callback`、`repeated_heartbeat_timeouts` | `internal/api/runtime_tasks_test.go`、`internal/store/runtime_tasks_test.go` |

---

## 三、当前仍未完全自动化的增强项

以下内容目前已经有架构方向或部分基础能力，但还没有完整矩阵化自动回归：

| 场景 | 当前情况 |
|---|---|
| 按 `consumer` / `category` 的更细粒度治理策略 | 仍属于后续能力，当前主要覆盖 interface/task 快照级策略 |
| hybrid 多阶段回切与更复杂 fallback state machine | 当前只覆盖最小回退闭环和观测字段 |
| callback replay 的 TTL / 签名 / nonce 策略 | 当前只覆盖请求 ID 级 replay 防护 |
| 管理面按诊断视图直接批量人工干预 | 当前已有筛选与统计，批量人工治理流仍未落地 |

---

## 四、阅读顺序建议

如果你要理解完整外部交互能力，推荐按这个顺序读：

1. [外部交互路线图](./EXTERNAL_INTERACTION_ROADMAP.md)
2. [外部交互接入示例](./EXTERNAL_INTERACTION_EXAMPLES.md)
3. [API 参考](./API_REFERENCE.md)
4. 本测试矩阵文档

这样可以依次建立：架构目标、接入方式、接口语义、自动化保障边界。
