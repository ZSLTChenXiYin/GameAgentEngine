# 配置参考

**中文** | [**English**](./CONFIGURATION_EN.md)

随包版本的配置重点：

- 静态配置文件：当前目录 `gameagentengine.conf.yaml`
- 动态世界配置：`world_settings`
- 时间规则：`world_time_settings`
- 时间结果：`world_time_state`

如果没有先配置 `world_time_settings`，世界时间推进相关流程会被阻塞。
