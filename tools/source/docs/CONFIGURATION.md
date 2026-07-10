# 配置参考

**中文** | [**English**](./CONFIGURATION_EN.md)

随包版本的配置重点：

- 静态配置文件：当前目录 `gameagentengine.conf.yaml`
- 动态世界配置：`world_settings`
- 时间规则：`world_time_settings`
- 时间结果：`world_time_state`

---

## 随包模板当前重点

```yaml
database:
  driver: "sqlite"
  dsn: "gameagentengine.db"
  migrations_enabled: true
  write_retry_enabled: true
  log_batch_enabled: true

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
```

这份模板偏向本地演示与开箱体验；如果缺失字段，Engine 仍会回退到代码级默认值。

---

## 额外提醒

- `database.driver` 支持 `sqlite`、`mysql`、`postgres`
- `GET /api/v1/pipeline/stats` 适合排查锁竞争、批量写与重试情况
- 如果没有先配置 `world_time_settings`，世界时间推进相关流程会被阻塞
