# 入门指南

**中文** | [**English**](./GETTING_STARTED_EN.md)

这份文档面向第一次接触 GameAgentEngine 的开发者，目标是带你从零完成三件事：启动 Engine、创建第一个世界、打开 Creator 开始编辑。

---

## 你会得到什么

完成本指南后，你应该可以：

- 启动本地 Engine 服务
- 使用 DevCli 创建世界根节点
- 在 Creator 中继续编辑节点、组件、关系
- 理解什么时候必须先配置 `world_time_settings`

---

## 前置条件

- Go 1.25+
- 可运行的终端环境
- 如果要接真实模型，需要一个兼容 OpenAI 协议的 API Key

如果 `llm.api_key` 留空，引擎会自动回退到 Mock Provider，这适合做本地功能联调，但不适合验证真实世界推理质量。

---

## 第一步：构建项目

```bash
git clone <仓库地址>
cd GameAgentEngine
go build ./...
```

如果你只是使用打包产物，也可以直接进入解压后的目录，不必重新编译。

---

## 第二步：准备配置文件

复制默认配置：

```bash
cp tools/source/gameagentengine.conf.yaml .
```

当前默认配置的关键点如下：

- 默认监听地址：`0.0.0.0:8080`
- 默认 API Key：`dev-key`
- 默认模型名：`deepseek-v4-flash`
- 默认 `base_url`：`https://api.deepseek.com`
- 默认执行模式：`debug`
- 默认后台自主调度器：已开启

最少要检查这几个字段：

```yaml
auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-v4-flash"
  api_key: "sk-xxx"
  base_url: "https://api.deepseek.com"
```

如果你不想连接真实模型，可以把 `llm.api_key` 留空。

---

## 第三步：启动 Engine

```bash
go run ./cmd/gameagentengine serve
```

确认服务正常：

```bash
curl http://127.0.0.1:8080/health
```

预期结果：

```json
{"status":"ok"}
```

---

## 第四步：创建第一个世界

当前新手流程从直接创建一个世界根节点开始。

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --type world --name "新世界"
```

这条命令会直接创建一个 `world` 类型节点，它就是整个世界树的根节点。

你也可以继续创建子节点，例如：

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node create --world <world-id> --type location --name "起始村庄"
```

如果你更喜欢代码方式建世界，也可以用 SDK 的 `Agent.CreateWorld()`。

---

## 第五步：打开 Creator

推荐直接用 DevCli 打开：

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key inspect
```

如果你的环境不方便走这个入口，也可以直接打开：

`tools/source/web/GameAgentCreator/index.html`

进入后你可以看到这些核心页面：

- `Worlds`：世界和节点树
- `Settings`：世界运行设置
- `Policy`：世界策略
- `Plans`：待审批计划
- `State`：连续性状态组件
- `Timelines`：时间线归档
- `Continuity`：连续性排查入口
- `Logs` / `Traces`：观测与调试

---

## 第六步：先配置世界时间，再跑 Tick

如果你要使用以下能力：

- `world tick`
- 时间线推进
- 连续性状态中的世界时间演化
- 世界线推理

那你应该先在 `Settings` 页面配置 `world_time_settings`。

这是当前设计中的强约束，不是可有可无的补充信息。没有世界时间系统，Engine 无法可靠地做时间推进和世界线连续性推理，所以相关保存/推进流程会故意阻塞，提醒开发者先完成配置。

你也可以通过 DevCli 配置：

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world settings set <world-id> --world-time-settings-json '{"tick_scale_mode":"flexible","tick_min_unit":"时","tick_step":1,"tick_units":["日","时"]}'
```

最小规则：

- `tick_units` 必须按从大到小排列
- `tick_min_unit` 必须等于最后一个单位
- `tick_scale_mode` 目前只能是 `fixed` 或 `flexible`

---

## 第七步：执行一次 Tick

完成世界时间配置后，就可以推进一次世界 Tick：

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world tick <world-id> --type manual --time "day-1" --requested-ticks 1
```

如果你要限制这次 Tick 最多触发多少个自主节点：

```bash
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world tick <world-id> --autonomous-limit 2
```

推进之后，你可以在 Creator 的这些页面里查看结果：

- `State`：查看 `world_time_state`
- `Timelines`：查看最新 tick 归档
- `Continuity`：查看连续性汇总与差异

---

## 常用后续命令

```bash
# 查看世界列表
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node list --type world

# 查看世界设置
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world settings get <world-id>

# 查看连续性状态组件
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key state list <world-id>

# 查看最新时间线
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key timeline latest <world-id>

# 查看最近日志
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key logs --world <world-id> --limit 10
```

---

## 下一步读什么

- [配置参考](./CONFIGURATION.md)
- [GameAgentCreator 指南](./GUIDE_GAMEAGENTCREATOR.md)
- [GameAgentDevCli 指南](./GUIDE_GAMEAGENTDEVCLI.md)
- [世界时间 Tick 参考](./WORLD_TIME_TICK_REFERENCE.md)
- [SDK 参考](./SDK_REFERENCE.md)
