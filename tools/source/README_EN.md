# GameAgentEngine

[**中文**](./README.md) | **English**

This directory is the bundled resource set distributed with GameAgentEngine. It contains the default config, demo world data, web tools, and user-facing reference documents.

If you already have a runnable engine build or integration package, this is the usual starting point:

- `gameagentengine.conf.yaml`: default configuration template
- `demo-world.yaml`: demo world import file
- `web/GameAgentCreator/`: Creator UI that can be opened directly in a browser
- `web/Demo/`: demo showcase page
- `docs/`: user-facing reference documents

---

## Quick Use

### 1. Configure the Engine

Edit `gameagentengine.conf.yaml` in this directory and set at least:

```yaml
llm:
  provider: "openai"
  model: "deepseek-chat"
  api_key: "sk-your-key"
  base_url: "https://api.deepseek.com/v1"
```

If `llm.api_key` is empty, the engine falls back to the mock provider for local testing.

### 2. Import the Demo World

```bash
GameAgentDevCli import demo-world.yaml --reset
```

Useful variants:

```bash
GameAgentDevCli import demo-world.yaml --dry-run
GameAgentDevCli import demo-world.yaml
```

### 3. Open Creator

Open this file in a browser:

`web/GameAgentCreator/index.html`

The current Creator supports:

- world creation and rename
- drag-to-parent and drag-to-root node reparenting
- node create, edit, delete, and copy
- snapshot save, validate, restore, and delete
- world settings, world policy, logs, and traces

---

## Directory Layout

```text
source/
|-- docs/
|-- web/
|   |-- Demo/
|   `-- GameAgentCreator/
|-- demo-world.yaml
|-- gameagentengine.conf.yaml
|-- README.md
`-- README_EN.md
```

---

## Document Index

- [Getting Started](./docs/GETTING_STARTED_EN.md)
- [Core Concepts](./docs/CORE_CONCEPTS_EN.md)
- [Configuration](./docs/CONFIGURATION_EN.md)
- [GameAgentCreator Guide](./docs/GUIDE_GAMEAGENTCREATOR_EN.md)
- [GameAgentDevCli Guide](./docs/GUIDE_GAMEAGENTDEVCLI_EN.md)
- [API Reference](./docs/API_REFERENCE_EN.md)
- [SDK Reference](./docs/SDK_REFERENCE_EN.md)
- [Demo World: Gray Harbor](./docs/DEMO_WORLD_GRAY_HARBOR_EN.md)

---

## Note

The `docs/` folder intentionally keeps only the documents that are useful to end users of the packaged engine resources. Internal architecture notes, build-chain details, and deeper implementation materials remain in the main repository documentation.
