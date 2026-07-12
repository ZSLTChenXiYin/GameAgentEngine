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
# Windows: tools\scripts\build.bat
# Linux/macOS: bash tools/scripts/build.sh
#
# 如果只是使用打包产物，也可以直接进入解压后的目录，不必重新编译。
```

如果你只是使用打包产物，跳过本步骤，直接进入解压后的目录即可。

---

## 第二步：准备配置文件

复制默认配置：

```bash
cp tools/source/gameagentengine.conf.yaml .
```

当前随包模板的关键点如下：

- 默认监听地址：`0.0.0.0:8080`
- 默认 API Key：`dev-key`
- 模板示例模型名：`deepseek-v4-flash`
- 模板示例 `base_url`：`https://api.deepseek.com`
- 模板执行模式：`debug`
- 默认后台自主调度器：关闭

额外需要知道的是，代码级保底默认值与模板示例并不完全相同；如果配置缺项，Engine 会回退到内部默认值，例如：

- `llm.model = gpt-4o-mini`
- `llm.base_url = https://api.openai.com/v1`
- `engine.execution_mode = full`

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
GameAgentEngine serve
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
GameAgentDevCli node create --type world --name "新世界"
```

这条命令会直接创建一个 `world` 类型节点，它就是整个世界树的根节点。

你也可以继续创建子节点，例如：

```bash
GameAgentDevCli node create --world <world-id> --type location --name "起始村庄"
```

---

## 第五步：打开 Creator

推荐直接用 DevCli 打开：

```bash
GameAgentDevCli inspect
```

如果你的环境不方便走这个入口，也可以直接打开：

`tools/source/web/GameAgentCreator/index.html`

进入后你可以看到这些核心页面：

- `Worlds`
- `Settings`
- `Policy`
- `Plans`
- `State`
- `Timelines`
- `Continuity`
- `Logs` / `Traces`

---

## 第六步：先配置世界时间，再跑 Tick

如果你要使用以下能力：

- `world tick`
- 时间线推进
- 连续性状态中的世界时间演化
- 世界线推理

那你应该先在 `Settings` 页面配置 `world_time_settings`。

这是当前设计中的强约束，不是可有可无的补充信息。没有世界时间系统，Engine 无法可靠地做时间推进和世界线连续性推理，所以相关保存/推进流程会故意阻塞，提醒开发者先完成配置。

---

## 第七步：观察管线状态

当你开始压测、跑 Tick 或启用自主行为时，推荐顺手观察：

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`

这三个入口可以帮助你快速判断：

- 是否出现频繁写重试
- 日志批量队列是否积压
- 世界级锁是否出现争用
