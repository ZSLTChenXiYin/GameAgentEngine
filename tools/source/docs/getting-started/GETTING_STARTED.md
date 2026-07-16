# 入门指南

**中文** | [**English**](./GETTING_STARTED_EN.md)

这份文档面向使用打包产物的新手开发者。

---

## 最短上手路径

```bash
# 1. 启动服务
GameAgentEngine serve

# 2. 创建世界
GameAgentDevCli node create --type world --name "新世界"

# 3. 打开 Creator
GameAgentDevCli creator
```

如果你要做世界时间推进，请先在 Creator 的 `Settings` 页面配置 `world_time_settings`。

---

## 配置文件

直接编辑当前目录里的 `gameagentengine.conf.yaml`。

当前随包模板重点：

- `auth.api_key: dev-key`
- `llm.model: deepseek-v4-flash`
- `llm.base_url: https://api.deepseek.com`
- `engine.execution_mode: debug`
- `engine.autonomous_scheduler_enabled: false`
- `engine.world_lock_enabled: true`

如果你删掉这些字段，Engine 仍会回退到代码级默认值，例如：

- `llm.model: gpt-4o-mini`
- `llm.base_url: https://api.openai.com/v1`
- `engine.execution_mode: full`

---

## 常用命令

```bash
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world tick <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
```

---

## 诊断入口

推荐同时关注：

- `GET /api/v1/pipeline/stats`
- `GET /api/v1/logs`
- `GET /debug/traces`
