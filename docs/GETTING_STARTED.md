# 入门指南

**中文** | [**English**](./GETTING_STARTED_EN.md)

这份文档带你从零启动 GameAgentEngine、导入示例世界，并打开 Creator 可视化编辑器。

---

## 前置条件

- Go 1.25+
- Git
- 如果需要真实 LLM 响应，需要准备一个 OpenAI 兼容 API Key

如果 `llm.api_key` 留空，引擎会自动回退到 mock provider，适合本地流程验证。

---

## 构建项目

```bash
git clone <仓库地址>
cd GameAgentEngine
go build ./...
```

---

## 配置引擎

```bash
cp tools/source/gameagentengine.conf.yaml .
```

然后编辑 `gameagentengine.conf.yaml`，至少填写：

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-your-key"
  base_url: "https://api.deepseek.com/v1"
```

---

## 启动引擎

```bash
go run ./cmd/gameagentengine serve
```

检查健康状态：

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
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml --reset
```

常见变体：

```bash
# 只校验，不写入
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml --dry-run

# 保留现有数据，直接导入
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key import tools/source/demo-world.yaml
```

---

## 基础运行命令

```bash
# 列出所有世界节点
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key node list --type world

# 推进世界 Tick
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world tick <world-id>

# 保存一个存档快照
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world save <world-id> "存档槽 1"

# 修改世界名称
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key world update <world-id> --name "新的世界名称"
```

---

## 打开 Creator

在浏览器中打开：

`tools/source/web/GameAgentCreator/index.html`

当前 Creator 已支持：

- 世界创建与世界重命名
- 节点树编辑与拖拽改父节点
- 节点复制
- 快照保存 / 校验 / 恢复 / 删除
- 世界设置与世界策略编辑
- 日志与调试轨迹查看

---

## 后续阅读

- [GameAgentCreator 指南](./GUIDE_GAMEAGENTCREATOR.md)
- [GameAgentDevCli 指南](./GUIDE_GAMEAGENTDEVCLI.md)
- [API 参考](./API_REFERENCE.md)
- [SDK 参考](./SDK_REFERENCE.md)
