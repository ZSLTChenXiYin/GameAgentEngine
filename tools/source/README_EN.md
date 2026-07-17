# Packaged Assets Directory Notes

[**中文**](./README.md) | **English**

This directory is for packaged runtime assets only. It is not the formal repository documentation tree.

## Intended Contents

This directory should currently contain only packaged runtime assets such as:

- shared runtime config templates at the source root, for example `gameagentengine.conf.yaml`
- demo assets under `workerhome/demo/`, for example `workerhome/demo/demo-world.yaml` and `workerhome/demo/demo-state.yaml`
- Worker / integration-test working data under `workerhome/fixtures/`
- Creator static assets under `web/GameAgentCreator/`

## Recommended Startup Path

```bash
GameAgentEngine serve
GameAgentDevCli import workerhome/demo/demo-world.yaml
GameAgentDevCli creator
GameAgentWorker play --state-file workerhome/demo/demo-state.yaml --world-id demo_world --player-node-id player_001
```

## Documentation Note

This directory no longer maintains a separate formal documentation copy.

- If you are working in the source repository, use the root `README.md` and the `docs/` tree.
- If you are using a packaged artifact directory, refer to the project documentation on GitHub instead of expecting a full local `docs/` tree here.

Repository entrypoints:

- GitHub repository: <https://github.com/ZSLTChenXiYin/GameAgentEngine>
- Source docs entry: `README.md` and `docs/`

## License

MIT
