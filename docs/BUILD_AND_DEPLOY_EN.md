# Build & Deploy

[**中文**](./BUILD_AND_DEPLOY.md) | **English**

GameAgentEngine v0.4.5 can be obtained either as a pre-compiled package for multiple platforms or built from source.

---

## Pre-compiled Packages

Each release provides compressed packages in `dist/` for the following platforms:

| Platform | File |
|---|---|
| Windows amd64 | `GameAgentEngine-windows-amd64-v0.4.5.zip` |
| Windows arm64 | `GameAgentEngine-windows-arm64-v0.4.5.zip` |
| Linux amd64 | `GameAgentEngine-linux-amd64-v0.4.5.zip` |
| Linux arm64 | `GameAgentEngine-linux-arm64-v0.4.5.zip` |
| macOS amd64 (Intel) | `GameAgentEngine-darwin-amd64-v0.4.5.zip` |
| macOS arm64 (Apple Silicon) | `GameAgentEngine-darwin-arm64-v0.4.5.zip` |

Each package contains:

```
GameAgentEngine-{os}-{arch}-v0.4.5/
├── GameAgentEngine.exe     # Backend engine service
├── GameAgentDevCli.exe     # CLI management tool
├── gameagentengine.conf.yaml  # Default configuration file
├── demo-world.yaml         # Demo world configuration
└── web/
    ├── Demo/               # Playable Demo showcase
    └── GameAgentCreator/   # Visual editor
```

### Using a Pre-compiled Package

```bash
# Extract
unzip GameAgentEngine-windows-amd64-v0.4.5.zip
cd GameAgentEngine-windows-amd64-v0.4.5

# Edit the configuration file, fill in LLM API Key
# Edit gameagentengine.conf.yaml

# Start the service
GameAgentEngine serve

# In another terminal, import the Demo world
GameAgentDevCli import demo-world.yaml --reset
```

---

## Building from Source

### Prerequisites

- Go 1.25+

### Local Build

```bash
# Build all components
go build -o GameAgentEngine ./cmd/gameagentengine/
go build -o GameAgentDevCli ./cmd/gameagentdevcli/
```

### Cross-compilation

Use `tools/scripts/build.sh` (Linux/macOS) or `tools/scripts/build.bat` (Windows):

```bash
# Build for all platforms
./tools/scripts/build.sh

# Build for a specific platform
GOOS=linux GOARCH=amd64 go build -o dist/GameAgentEngine-linux-amd64 ./cmd/gameagentengine/
GOOS=windows GOARCH=amd64 go build -o dist/GameAgentEngine-windows-amd64.exe ./cmd/gameagentengine/
GOOS=darwin GOARCH=amd64 go build -o dist/GameAgentEngine-darwin-amd64 ./cmd/gameagentengine/
```

The packaging script automatically creates version directories, regenerates `tools/source/web/GameAgentCreator/js/component-meta.js`, and copies resources such as the web/ directory.

---

## Deployment

### Basic Deployment

```bash
# 1. Copy the extracted directory to your server
# 2. Edit gameagentengine.conf.yaml
# 3. Run the service
GameAgentEngine serve
```

### Production Recommendations

1. **Change the default API Key** — set `auth.api_key` from `"dev-key"` to a random string
2. **Configure a valid LLM API Key**
3. **Use MySQL** (optional) — MySQL is recommended over SQLite for production
4. **Reverse proxy** — Use Nginx as a reverse proxy in production
5. **System service** — Configure as systemd (Linux) or a Windows Service

### Nginx Reverse Proxy Example

```nginx
server {
    listen 80;
    server_name engine.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### systemd Service Configuration (Linux)

```ini
[Unit]
Description=GameAgentEngine
After=network.target

[Service]
Type=simple
ExecStart=/opt/gameagent/GameAgentEngine serve
WorkingDirectory=/opt/gameagent
Restart=always
User=gameagent

[Install]
WantedBy=multi-user.target
```

---

## Ports & Security

- Default port: 8080
- API authentication: via `X-API-Key` request header
- Default dev key is `dev-key` — change it in production
- CORS allows all origins by default
