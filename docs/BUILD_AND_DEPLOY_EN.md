# Build & Deploy

[**中文**](./BUILD_AND_DEPLOY.md) | **English**

GameAgentEngine v0.4.6 supports both source builds and packaged releases.

---

## Current Package Layout

Each release package is structured as `dist/GameAgentEngine-{os}-{arch}-v0.4.6/`.

```text
GameAgentEngine-{os}-{arch}-v0.4.6/
├── GameAgentEngine(.exe)
├── GameAgentDevCli(.exe)
├── gameagentengine.conf.yaml
├── README.md
├── README_EN.md
├── docs/
└── web/
    └── GameAgentCreator/
```

Packaged builds now include the docs, config, and static assets from `tools/source/`, which means they also ship with:

- `demo-world.yaml`
- `demo-state.yaml`

Packaged builds still do not include a standalone `web/Demo/` page.

---

## Build From Source

Prerequisite:

- Go 1.25+

```bash
go build -o GameAgentEngine ./cmd/gameagentengine/
go build -o GameAgentDevCli ./cmd/gameagentdevcli/
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

- build `GameAgentEngine` and `GameAgentDevCli`
- inject version values
- regenerate `tools/source/web/GameAgentCreator/js/component-meta.js`
- copy config, docs, and Creator static assets from `tools/source/`
- emit zip archives

---

## Start a Packaged Build

```bash
cd GameAgentEngine-windows-amd64-v0.4.6
GameAgentEngine serve
```

Use the bundled `GameAgentDevCli` to manage worlds, open Creator, and run ticks.

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
