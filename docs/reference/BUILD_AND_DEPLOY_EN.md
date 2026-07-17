# Build & Deploy

[**中文**](./BUILD_AND_DEPLOY.md) | **English**

GameAgentEngine v0.5.0 supports both source builds and packaged releases.

---

## Current Package Layout

Each release package is structured as `dist/GameAgentEngine-{os}-{arch}-v0.5.0/`.

```text
GameAgentEngine-{os}-{arch}-v0.5.0/
├── GameAgentEngine(.exe)
├── GameAgentDevCli(.exe)
├── GameAgentWorker(.exe)
├── gameagentengine.conf.yaml
├── README.md
├── sdks/
├── workerhome/
│   ├── demo/
│   │   ├── demo-world.yaml
│   │   └── demo-state.yaml
│   └── fixtures/
└── web/
    └── GameAgentCreator/
```

Packaged builds now come from the runtime-asset tree under `tools/source/`, so they include:

- the config template
- demo world / authority-state files under `workerhome/demo/`
- shared Worker / SDK test data under `workerhome/fixtures/`
- multi-language SDK source and examples under `sdks/`
- Creator static assets

Packaged builds no longer ship the full repository `docs/` tree. Use the bundled `README.md`, the repository root README, and the GitHub-hosted docs as the primary documentation entrypoints.

---

## Build From Source

Prerequisite:

- Go 1.25+

```bash
go build -o GameAgentEngine ./cmd/gameagentengine/
go build -o GameAgentDevCli ./cmd/gameagentdevcli/
go build -o GameAgentWorker ./cmd/gameagentworker/
```

---

## Use the Packaging Scripts

Windows:

```bash
tools\scripts\build.bat windows/amd64
```

Linux or macOS:

```bash
./tools/scripts/build.sh linux/amd64
```

All platforms:

```bash
./tools/scripts/build.sh all
```

The packaging scripts automatically:

- build `GameAgentEngine`, `GameAgentDevCli`, and `GameAgentWorker`
- inject version values
- regenerate `tools/source/web/GameAgentCreator/js/component-meta.js`
- copy the runtime asset tree from `tools/source/`
- rewrite the packaged Creator minimum-compatible version in the output directory without mutating source files
- emit zip archives

---

## Start a Packaged Build

```bash
cd GameAgentEngine-windows-amd64-v0.5.0
GameAgentEngine serve
```

Use the bundled `GameAgentDevCli` to manage worlds, open Creator, and run ticks.

Use the bundled `GameAgentWorker` to simulate game-side async interfaces, run packaged integration scenarios, and enter play-mode REPL.

---

## Deployment Notes

At minimum for production:

1. replace `auth.api_key` instead of keeping `dev-key`
2. configure a real LLM API key
3. evaluate switching to MySQL
4. put a reverse proxy in front of the HTTP service
5. supervise the process with a system service

---

## Ports & Security

- default port: `8080`
- API auth header: `X-API-Key`
- default dev key: `dev-key`
- default CORS behavior: allow all origins
