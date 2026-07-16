# GameAgentEngine Packaged Assets

这个目录用于承载打包产物随附的运行资产，而不是完整仓库文档。

## 本目录当前应只包含

- 运行配置模板：`gameagentengine.conf.yaml`
- demo 资产：`demo-world.yaml`、`demo-state.yaml`
- Worker / 集成测试工作数据：`tests/`
- Creator 静态资源：`web/GameAgentCreator/`

## 推荐启动路径

```bash
GameAgentEngine serve
GameAgentDevCli import demo-world.yaml
GameAgentDevCli creator
GameAgentWorker play --state-file demo-state.yaml --world-id demo_world --player-node-id player_001
```

## 文档说明

这个目录不再维护独立的多语言文档副本。

如果你在源码仓库中工作，完整文档请查看仓库根目录 `README.md` 与 `docs/` 目录。

如果你当前使用的是打包产物目录，请直接查看项目 GitHub 上对应的文档页面；打包产物本身不再附带完整 `docs/` 树。

## 许可证

MIT
