# SDK 能力矩阵

本矩阵按当前项目工作流，总结各 SDK 目录的实现成熟度。

状态等级：

- `practical`：已经可用于真实的 Engine / Worker 联调工作
- `baseline`：已有基础脚手架，但能力仍偏浅
- `planned`：在当前重构序列中尚未提升到目标层级

## 1. 当前矩阵

| SDK | 当前级别 | Health / Version | Invoke | Runtime Task Loop | Callback | State / Timeline / Logs | Worker 示例 | 备注 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `ts-sdk` | practical | yes | yes | yes | yes | yes | yes | 当前最强的非 Go SDK |
| `js-sdk` | practical | yes | yes | yes | yes | yes | yes | 面向纯 Node.js / 脚本工具链的轻量同构版本 |
| `cs-sdk` | practical | yes | yes | yes | yes | yes | yes | 面向 Unity / .NET 的 typed client |
| `gd-sdk` | practical-request-builder | request builders | request builders | request builders | request builders | request builders | yes | 已补齐 worker authority-query 与 roundtrip 示例，但执行包装层仍缺失 |
| `cpp-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | 请求构造层已覆盖 pull/callback 辅助序列与 worker 示例 |
| `java-sdk` | baseline+worker-loop | yes | yes | yes | yes | no | yes | 已有真实 HTTP client 和 authority-query / callback roundtrip 示例，但观测面仍偏浅 |
| `lua-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | 轻量 request helper 已镜像 worker loop 与 authority-query 顺序 |
| `c-sdk` | baseline+worker-examples | partial | partial | partial | partial | no | yes | 路径/payload helper 与 worker 示例已具备，调用方仍自管 transport |

## 2. 这里的 practical 含义

在本项目里，一个 practical SDK 预期至少覆盖当前外围联调工作流：

1. 连接 Engine
2. 发起一次 invoke
3. 检查或消费一个 runtime task
4. 回填一次 callback 结果
5. 在需要时检查连续性相关运行产物
6. 能与 `GameAgentWorker` 顺利配合

## 3. 当前主线 SDK

当前主线 SDK 集合为：

- `ts-sdk`
- `js-sdk`
- `cs-sdk`
- `gd-sdk`

这些 SDK 已经直接映射到当前 Engine / Worker 开发工作流。

## 4. 剩余 SDK 的升级优先级

当前计划中的剩余顺序：

1. `lua-sdk`
2. `c-sdk`
3. `cpp-sdk` 的观测面 / typed-model 深度

这一阶段之后的下一个升级目标，仍然是补齐低层级 SDK 的观测面与 typed-model 深度，而不是重做已经 practical 的 SDK。
