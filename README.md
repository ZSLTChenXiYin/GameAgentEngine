# GameAgentEngine

**中文** | [**English**](./README_EN.md)

**面向游戏开发者的 AI Agent 制作与运行引擎。**

GameAgentEngine 是一个基于 Go 语言的引擎，位于游戏逻辑与大模型能力之间——负责世界建模、NPC 智能行为、记忆管理和世界时间线推进。可以理解为游戏世界中的**导演系统与智能运行层**。

> 它并**不替代** Unity、Unreal 或 Godot，而是与它们**协同工作**。

---

## 特性

- **统一世界建模** — 节点、组件、记忆、关系构成完整的实体图
- **NPC 智能** — 基于 LLM 的对话，具备上下文感知（身份、背景、记忆、关系）
- **世界时间线** — 基于 Tick 的推进机制与事件影响评估
- **推理管线** — 上下文组装 → Prompt 生成 → LLM 调用 → 动作解析 → 记忆持久化
- **动作系统** — 内置同步/异步动作（add_memory, update_mood, send_dialogue, adjust_relation, spawn_item）
- **策略引擎** — 安全执行护栏（禁止动作、审核阈值）
- **世界复制** — 完整复制世界及其全部数据，支持锁定源世界防止并发写入
- **幂等支持** — 写入操作的安全重试
- **完整 CRUD API** — 20+ RESTful 端点
- **双存储支持** — SQLite（开发）/ MySQL（生产）
- **Go SDK** — 原生 Go 客户端库，含 Agent 构建器
- **GameAgentDevCli** — 开发者 CLI，支持脚本化导入、CRUD、世界推进、验证
- **GameAgentCreator** — 基于 Web 的可视化编辑器（节点树、检视器、日志、导入）

---

## 快速开始

```bash
# 1. 克隆并构建
git clone <仓库地址>
cd GameAgentEngine
go build ./...

# 2. 配置（复制默认配置并填入 LLM API Key）
cp tools/source/gameagentengine.conf.yaml .
# 编辑 gameagentengine.conf.yaml — 设置 llm.api_key

# 3. 启动引擎服务
go run ./cmd/gameagentengine serve

# 4. 另开终端，填充 Demo 世界
go run ./cmd/gameagentdevcli --server http://127.0.0.1:8080 --key dev-key --config gameagentengine.conf.yaml demo-seed

# 5. 打开可视化编辑器
# web/GameAgentCreator/index.html
```

详见[入门指南](docs/GETTING_STARTED.md)。

---

## 项目结构

```
GameAgentEngine/
├── cmd/
│   ├── gameagentengine/      # 引擎服务 + CLI（serve, validate, version, import ...）
│   └── gameagentdevcli/      # 开发者 CLI（CRUD, world tick, import, verify, snapshot ...）
├── internal/
│   ├── api/                  # HTTP API 层（路由、处理器、中间件、错误映射）
│   ├── service/              # 领域规则与事务边界
│   ├── engine/               # 推理管线、上下文构建器、核心类型
│   ├── store/                # GORM 持久化层
│   ├── llm/                  # LLM Provider（兼容 OpenAI + Mock）
│   ├── action/               # 动作注册与回调系统
│   ├── planner/              # 策略引擎与世界变更计划评估
│   └── config/               # Viper 配置加载
├── sdk/                      # Go HTTP 客户端 SDK
├── web/
│   ├── GameAgentCreator/     # 可视化编辑器（节点树、检视器、日志、导入）
│   └── Demo/                 # Demo 展示页面（灰港议会）
├── tools/
│   ├── scripts/              # 构建脚本、编码检查
│   └── source/               # 默认配置文件
└── docs/                     # 文档（支持中英双语）
```

---

## 工具一览

| 工具 | 用途 | 入口 |
|---|---|---|
| **GameAgentEngine** | 后端引擎服务 + CLI | `cmd/gameagentengine/main.go` |
| **GameAgentDevCli** | 开发者命令行工具 | `cmd/gameagentdevcli/main.go` |
| **GameAgentCreator** | 前端可视化编辑器 | `web/GameAgentCreator/index.html` |
| **Web Demo** | Demo 展示页面 | `web/Demo/index.html` |

---

## 文档（中英文双语）

| 文档 | 中文 | English |
|---|---|---|
| 入门指南 | [GETTING_STARTED.md](docs/GETTING_STARTED.md) | [EN](docs/GETTING_STARTED_EN.md) |
| 架构设计 | [ARCHITECTURE.md](docs/ARCHITECTURE.md) | [EN](docs/ARCHITECTURE_EN.md) |
| 核心概念 | [CORE_CONCEPTS.md](docs/CORE_CONCEPTS.md) | [EN](docs/CORE_CONCEPTS_EN.md) |
| 自主行为 | [AUTONOMOUS_BEHAVIOR.md](docs/AUTONOMOUS_BEHAVIOR.md) | [EN](docs/AUTONOMOUS_BEHAVIOR_EN.md) |
| API 参考 | [API_REFERENCE.md](docs/API_REFERENCE.md) | [EN](docs/API_REFERENCE_EN.md) |
| GameAgentDevCli 指南 | [GUIDE_GAMEAGENTDEVCLI.md](docs/GUIDE_GAMEAGENTDEVCLI.md) | [EN](docs/GUIDE_GAMEAGENTDEVCLI_EN.md) |
| GameAgentCreator 指南 | [GUIDE_GAMEAGENTCREATOR.md](docs/GUIDE_GAMEAGENTCREATOR.md) | [EN](docs/GUIDE_GAMEAGENTCREATOR_EN.md) |
| 配置参考 | [CONFIGURATION.md](docs/CONFIGURATION.md) | [EN](docs/CONFIGURATION_EN.md) |
| SDK 参考 | [SDK_REFERENCE.md](docs/SDK_REFERENCE.md) | [EN](docs/SDK_REFERENCE_EN.md) |
| 构建与部署 | [BUILD_AND_DEPLOY.md](docs/BUILD_AND_DEPLOY.md) | [EN](docs/BUILD_AND_DEPLOY_EN.md) |
| 管线内部 | [PIPELINE_INTERNALS.md](docs/PIPELINE_INTERNALS.md) | [EN](docs/PIPELINE_INTERNALS_EN.md) |
| Demo 世界：灰港 | [DEMO_WORLD_GRAY_HARBOR.md](docs/DEMO_WORLD_GRAY_HARBOR.md) | [EN](docs/DEMO_WORLD_GRAY_HARBOR_EN.md) |

---

## 技术栈

| 层次 | 技术 | 用途 |
|---|---|---|
| 语言 | Go 1.25+ | 核心引擎 |
| HTTP | net/http, http.ServeMux | 服务接口 |
| ORM | GORM v2 | 数据库访问 |
| 存储 | SQLite / MySQL | 持久化 |
| AI | OpenAI-compatible API | LLM 推理 |
| CLI | Cobra | 命令行框架 |
| 配置 | Viper | 配置管理 |

---

## 许可证

MIT