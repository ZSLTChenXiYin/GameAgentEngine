# 入门指南

**中文** | [**English**](./GETTING_STARTED_EN.md)

本文档面向随包使用 GameAgentEngine 资源的用户，帮助你完成配置、导入示例世界，并打开 Creator。

---

## 你需要准备的内容

- 可运行的 `GameAgentEngine`
- 可运行的 `GameAgentDevCli`
- Go 兼容的 OpenAI 风格 API Key（如果你希望使用真实 LLM）

如果 `llm.api_key` 留空，引擎会自动使用 Mock Provider 做本地测试。

---

## 配置引擎

编辑当前目录随包提供的 `gameagentengine.conf.yaml`，至少填写：

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-your-key"
  base_url: "https://api.deepseek.com/v1"
```

更多字段说明见[配置参考](./CONFIGURATION.md)。

---

## 启动引擎

```bash
GameAgentEngine serve
```

用健康检查确认服务已启动：

```bash
curl http://127.0.0.1:8080/health
```

预期结果：

```json
{"status":"ok"}
```

---

## 导入示例世界

```bash
GameAgentDevCli import demo-world.yaml --reset
```

常用变体：

```bash
# 只校验，不写入数据库
GameAgentDevCli import demo-world.yaml --dry-run

# 保留现有数据并继续导入
GameAgentDevCli import demo-world.yaml
```

---

## 常用运行命令

```bash
# 列出世界
GameAgentDevCli node list --type world

# 推进一个世界 Tick
GameAgentDevCli world tick <world-id>

# 为某个世界创建存档快照
GameAgentDevCli world save <world-id> "Save Slot 1"

# 修改世界名称
GameAgentDevCli world update <world-id> --name "新的世界名称"

# 复制节点
GameAgentDevCli node copy <node-id> --name "复制节点"
```

完整命令见 [GameAgentDevCli 指南](./GUIDE_GAMEAGENTDEVCLI.md)。

---

## 打开 Creator

在浏览器中打开：

`web/GameAgentCreator/index.html`

当前 Creator 支持：

- 世界创建与世界重命名
- 节点树拖拽改父级与拖到根级
- 节点创建、编辑、删除、复制
- 快照保存、校验、恢复、删除
- 世界设置、世界策略、日志与调试轨迹查看

详细说明见 [GameAgentCreator 指南](./GUIDE_GAMEAGENTCREATOR.md)。

---

## 下一步阅读

- [核心概念](./CORE_CONCEPTS.md)
- [API 参考](./API_REFERENCE.md)
- [SDK 参考](./SDK_REFERENCE.md)
- [Demo 世界：灰港](./DEMO_WORLD_GRAY_HARBOR.md)
