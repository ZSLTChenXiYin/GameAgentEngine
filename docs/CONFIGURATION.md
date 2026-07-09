# 配置参考

**中文** | [**English**](./CONFIGURATION_EN.md)

GameAgentEngine v0.4.6 采用双层配置体系：

- 静态配置：`gameagentengine.conf.yaml`
- 动态配置：数据库中的 `world_settings`、`world_policy` 与状态组件

---

## 静态配置文件

默认配置文件位于 `tools/source/gameagentengine.conf.yaml`。

搜索顺序：

1. `--config <路径>`
2. `./gameagentengine.conf.yaml`
3. `./config/gameagentengine.conf.yaml`

当前默认内容：

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "sqlite"
  dsn: "gameagentengine.db"

auth:
  api_key: "dev-key"

llm:
  provider: "openai"
  model: "deepseek-v4-flash"
  api_key: "sk-xxx"
  base_url: "https://api.deepseek.com"

engine:
  execution_mode: "debug"
  autonomous_scheduler_enabled: true
  autonomous_scheduler_interval_seconds: 300
  autonomous_scheduler_max_nodes_per_world: 10
```

### 关键说明

- `execution_mode` 是服务级静态配置
- `pipeline_mode` 是每个世界独立的动态配置
- 后台自主调度器总开关在静态配置中
- 节点是否真正参与自主调度，取决于节点上的 `autonomous` 组件

---

## 动态配置：world_settings

`world_settings` 是每个世界独立持久化的运行时设置，当前包括：

- `memory_limit`
- `max_analysis_rounds`
- `max_context_depth`
- `auto_apply`
- `require_review_above`
- `pipeline_mode`
- `propagation_max_depth`
- `enable_propagation_machine`
- `sub_task_max_retries`
- `sub_task_timeout_secs`
- `world_time_settings`

---

## world_time_settings

`world_time_settings` 属于 `world_settings`，用于定义世界时间系统规则。它不是静态配置文件字段。

常见字段：

- `tick_scale_mode`
- `tick_min_unit`
- `tick_step`
- `tick_units`
- `time_scale_carry`
- `time_calendar`
- `unit_value_sequences`

核心约束：

- `tick_scale_mode` 必须是 `fixed` 或 `flexible`
- `tick_units` 至少包含一个单位，且不能重复
- `tick_units` 必须按从大到小排列
- `tick_min_unit` 必须等于最后一个单位
- 启用 `time_calendar` 时，`calendar_name` 必填
- 启用 `time_calendar` 时，`time_calendar.units` 必须与 `tick_units` 完全一致

如果没有先设置 `world_time_settings`，依赖世界时间推进的保存和推理流程会被阻塞。这是当前设计中的提醒机制。

---

## world_time_state

`world_time_state` 不是配置，而是 Engine 运行时写出的状态组件，用于保存当前时间结果。

可以这样理解：

- `world_time_settings`：时间规则
- `world_time_state`：时间结果

`world_time_state` 会出现在：

- 状态组件列表
- 时间线归档
- 连续性聚合结果
- Tick 响应

---

## 常用入口

### DevCli

```bash
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world settings set <world-id> --pipeline-mode polling
GameAgentDevCli world settings set <world-id> --world-time-settings-file world-time.json
```

### Creator

在 `Settings` 页面直接编辑普通运行参数和 `world_time_settings`。

### SDK

SDK 的 `WorldSettings` 结构已经包含 `WorldTimeSettings`。
