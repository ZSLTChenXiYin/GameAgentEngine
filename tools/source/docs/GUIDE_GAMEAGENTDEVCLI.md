# GameAgentDevCli 指南

**中文** | [**English**](./GUIDE_GAMEAGENTDEVCLI_EN.md)

当前随包版本的 DevCli 重点能力：

- 创建世界与节点
- 配置 `world_settings` 和 `world_time_settings`
- 推进世界 Tick
- 查看 `world_time_state`、时间线、日志与连续性
- 打开 Creator

---

## 常用命令

```bash
GameAgentDevCli node create --type world --name "新世界"
GameAgentDevCli import demo-world.yaml
GameAgentDevCli verify demo
GameAgentDevCli world settings set <world-id> --world-time-settings-file world-time.json
GameAgentDevCli world tick <world-id>
GameAgentDevCli state get <world-id> world_time_state
GameAgentDevCli timeline latest <world-id>
GameAgentDevCli creator
```

如果你使用随包附带的 `demo-world.yaml`，可以先导入它，再配合 `GameAgentWorker play --state-file demo-state.yaml --world-id demo_world --player-node-id player_001` 直接体验文字游戏式交互。
