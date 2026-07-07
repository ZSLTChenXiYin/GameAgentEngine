# GameAgentEngine

**中文** | [**English**](./README_EN.md)

这是随 GameAgentEngine 一起分发的资源目录，包含默认配置、示例世界、Web 工具和可直接查阅的用户文档。

如果你已经拿到了可运行的 Engine 或集成包，通常会从这里开始：

- `gameagentengine.conf.yaml`：默认配置模板
- `demo-world.yaml`：示例世界导入文件
- `web/GameAgentCreator/`：可直接在浏览器打开的 Creator
- `web/Demo/`：演示世界页面
- `docs/`：面向使用者的参考文档

---

## 快速使用

### 1. 配置引擎

编辑当前目录下的 `gameagentengine.conf.yaml`，至少填写：

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-your-key"
  base_url: "https://api.deepseek.com/v1"
```

如果 `llm.api_key` 为空，引擎会退回到 Mock Provider，适合本地联调。

### 2. 导入示例世界

```bash
GameAgentDevCli import demo-world.yaml --reset
```

常用变体：

```bash
GameAgentDevCli import demo-world.yaml --dry-run
GameAgentDevCli import demo-world.yaml
```

### 3. 打开 Creator

在浏览器中打开：

`web/GameAgentCreator/index.html`

当前 Creator 已支持：

- 世界创建与世界重命名
- 节点树拖拽改父级与拖到根级
- 节点创建、编辑、删除、复制
- 快照保存、校验、恢复、删除
- 世界设置、世界策略、日志与调试轨迹查看

---

## 目录结构

```text
source/
|-- docs/
|-- web/
|   |-- Demo/
|   `-- GameAgentCreator/
|-- demo-world.yaml
|-- gameagentengine.conf.yaml
|-- README.md
`-- README_EN.md
```

---

## 文档索引

- [入门指南](./docs/GETTING_STARTED.md)
- [核心概念](./docs/CORE_CONCEPTS.md)
- [配置参考](./docs/CONFIGURATION.md)
- [GameAgentCreator 指南](./docs/GUIDE_GAMEAGENTCREATOR.md)
- [GameAgentDevCli 指南](./docs/GUIDE_GAMEAGENTDEVCLI.md)
- [API 参考](./docs/API_REFERENCE.md)
- [SDK 参考](./docs/SDK_REFERENCE.md)
- [Demo 世界：灰港](./docs/DEMO_WORLD_GRAY_HARBOR.md)

---

## 说明

`docs/` 目录只保留了适合最终使用者查阅的文档，不包含全部内部设计文档。更偏实现细节、构建链路或内部架构的材料应以仓库主文档为准。
