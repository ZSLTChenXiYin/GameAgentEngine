# 打包资产目录说明

**中文** | [**English**](./README_EN.md)

本目录仅用于存放打包运行时资产，不是仓库的正式文档主树。

## 预期内容

当前本目录应只包含打包运行所需资产，例如：

- 位于根目录的共享运行配置模板，例如 `gameagentengine.conf.yaml`
- 位于 `workerhome/demo/` 下的 demo 资产，例如 `workerhome/demo/demo-world.yaml` 和 `workerhome/demo/demo-state.yaml`
- 位于 `workerhome/fixtures/` 下的 Worker / 集成测试工作数据
- 位于 `web/GameAgentCreator/` 下的 Creator 静态资源

## 推荐启动路径

```bash
GameAgentEngine serve
GameAgentDevCli import workerhome/demo/demo-world.yaml
GameAgentDevCli creator
GameAgentWorker play --state-file workerhome/demo/demo-state.yaml --world-id demo_world --player-node-id player_001
```

## 文档说明

本目录不再维护单独的正式文档副本。

- 如果你在源码仓库中工作，请使用根目录 `README.md` 与 `docs/` 文档树。
- 如果你正在使用打包产物目录，请优先参考 GitHub 上的项目文档，而不是期待本目录内存在完整的本地 `docs/` 文档树。

仓库入口：

- GitHub 仓库：<https://github.com/ZSLTChenXiYin/GameAgentEngine>
- 源码文档入口：`README.md` 与 `docs/`

## 许可

MIT
