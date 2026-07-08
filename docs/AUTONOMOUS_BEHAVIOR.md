# 自主行为系统

**中文** | [**English**](./AUTONOMOUS_BEHAVIOR_EN.md)

GameAgentEngine v0.4.5 的自主行为系统允许 NPC 和节点在没有用户直接输入的情况下，自行决定并执行行为。

---

## 工作原理

自主行为通过挂载在节点上的 `autonomous` 组件配置。每个节点可以独立配置是否启用自主行为、触发方式、以及可调用的能力白名单。

### 触发方式

| 触发模式 | 值 | 说明 |
|---|---|---|
| 手动触发 | `manual` | 仅通过 API 或 DevCli 手动触发 |
| Tick 同步 | `world_tick_sync` | 在世界 Tick 推进时自动触发 |
| 定时调度 | `scheduled` | 按配置的时间间隔自动触发（需启用后台调度器） |

### 能力白名单

每个自主节点可以声明它有权调用的动作列表（capabilities）。引擎会校验 LLM 的输出动作是否在白名单内，阻止越权行为：

```json
{
  "capabilities": [
    {
      "id": "add_memory",
      "mode": "sync",
      "description": "记录短期判断",
      "schema": {
        "node_id": {"type": "string", "required": true},
        "content": {"type": "string", "required": true}
      }
    }
  ]
}
```

---

## 配置示例

### 通过 DevCli 配置

```bash
# 查看当前配置
GameAgentDevCli node autonomous get <node-id>

# 启用并配置为世界 Tick 同步触发
GameAgentDevCli node autonomous set <node-id> --enabled --trigger "world_tick_sync"

# 禁用
GameAgentDevCli node autonomous disable <node-id>

# 手动触发一次
GameAgentDevCli node autonomous run <node-id>
```

### 通过 API 配置

```json
// GET /api/v1/nodes/{node_id}/autonomous
// PUT /api/v1/nodes/{node_id}/autonomous

{
  "enabled": true,
  "trigger": "world_tick_sync",
  "capabilities": [
    {"id": "add_memory", "mode": "sync", "description": "记录判断"}
  ]
}
```

---

## 调度器配置

后台自主行为调度器是**服务级静态开关**，通过配置文件控制：

```yaml
engine:
  autonomous_scheduler_enabled: false           # 全局开关
  autonomous_scheduler_interval_seconds: 300    # 扫描间隔
  autonomous_scheduler_max_nodes_per_world: 10  # 每次扫描每世界最大触发数
```

调度器扫描时，只会触发满足以下条件的节点：
1. 挂载了 `autonomous` 组件
2. `enabled = true`
3. `trigger = "scheduled"`
4. 距离上次运行时间超过 `interval_seconds`
