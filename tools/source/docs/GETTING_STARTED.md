# 入门指南

**中文** | [**English**](./GETTING_STARTED_EN.md)

本指南将带你从零开始搭建 GameAgentEngine v0.2.0，配置 LLM 提供商，导入 Demo 世界，并与 NPC 交互。

---

## 前置条件

- **Go 1.25+** — [下载 Go](https://go.dev/dl/)
- **LLM API Key** — 兼容 OpenAI 的 API 密钥（DeepSeek、OpenAI、Qwen 等）
- **Git** — 用于克隆仓库

---

## 安装

```bash
# 克隆仓库
git clone <仓库地址>
cd GameAgentEngine

# 构建所有组件
go build ./...

# 验证构建
GameAgentEngine version
# 输出: GameAgentEngine version v0.2.0
```

---

## 配置

复制默认配置文件并填入你的 LLM API Key：

```bash
cp tools/source/gameagentengine.conf.yaml .
```

编辑 `gameagentengine.conf.yaml`，设置 LLM API Key：

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"         # 或 gpt-4o-mini、qwen-turbo 等
  api_key: "sk-your-key-here"   # <-- 在这里设置
  base_url: "https://api.deepseek.com/v1"
```

> **没有 API Key？** 如果 `api_key` 留空，引擎会使用 Mock LLM Provider 返回固定响应——适合在无需真实 API 调用的情况下测试管线。

---

## 启动引擎服务

```bash
GameAgentEngine serve
```

预期输出：
```
DB: sqlite (gameagentengine.db)
LLM: deepseek-chat (https://api.deepseek.com/v1)
listen on 0.0.0.0:8080
```

验证服务运行：

```bash
curl http://127.0.0.1:8080/health
# {"status":"ok"}
```

---

## 导入 Demo 世界

另开终端，导入 Demo 世界「灰港边境」：

```bash
GameAgentDevCli import tools/source/demo-world.yaml --reset
```

`--reset` 参数会先清空数据库再导入。如果只想校验不写入，使用 `--dry-run`。

这将导入一个完整的世界，包含：
- 4 个阵营（灰港议会、铁潮商会、铁潮商会·雾湾矿场分部、北境守军）
- 8 个地点（议事厅、圆桌议会、雾湾矿场等）
- 4 个 NPC（艾琳议长、布莱姆总管、赛洛指挥官、米拉代表）
- 14 条带权重的关系
- 自定义组件（resource_state、district_state、demo_state）

验证世界创建成功：

```bash
GameAgentDevCli status
```

---

## 快速命令

### 列出所有节点

```bash
GameAgentDevCli node list
```

### 与世界交互

```bash
# 推进世界时间
GameAgentDevCli world tick <world-id>

# 与 NPC 对话（通过 REST API）
curl -X POST http://127.0.0.1:8080/api/v1/invoke \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key" \
  -d '{"world_id":"<world-id>","node_id":"<npc-id>","task_type":"npc_dialogue","context":{"messages":[{"role":"user","content":"你好"}]}}'

# 评估事件影响
GameAgentDevCli world event-impact <world-id> --type "crisis" --description "..." --severity "critical"

# 复制世界
GameAgentDevCli world clone <world-id> "我的存档副本" --lock

# 查看世界运行设置
GameAgentDevCli world settings get <world-id>

# 切换管线模式
GameAgentDevCli world settings set <world-id> --pipeline-mode "full"
```

*(具体命令详见 [GameAgentDevCli 指南](GUIDE_GAMEAGENTDEVCLI.md)。)*

---

## 打开可视化编辑器

在浏览器中打开 `web/GameAgentCreator/index.html`。

---

## 打开 Demo 展示

在浏览器中打开 `web/Demo/index.html`。

这是「灰港边境」的可玩 Demo：与不同 NPC 对话、每回合做出决策、推进世界时间线。

---

## 后续步骤

| 主题 | 文档 |
|---|---|
| 理解核心概念 | [核心概念](CORE_CONCEPTS.md) |
| 探索所有 API 端点 | [API 参考](API_REFERENCE.md) |
| 掌握 CLI | [GameAgentDevCli 指南](GUIDE_GAMEAGENTDEVCLI.md) |
| 使用 Web 编辑器 | [GameAgentCreator 指南](GUIDE_GAMEAGENTCREATOR.md) |
| 了解 Demo 世界 | [Demo 世界：灰港](DEMO_WORLD_GRAY_HARBOR.md) |
| 构建与打包分发 | [构建与部署](BUILD_AND_DEPLOY.md) |
| 在你的项目中使用 Go SDK | [SDK 参考](SDK_REFERENCE.md) |