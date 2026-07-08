# 构建与部署

**中文** | [**English**](./BUILD_AND_DEPLOY_EN.md)

GameAgentEngine v0.4.5 提供跨平台预编译包和源码构建两种获取方式。

---

## 预编译包

每个版本在 dist/ 目录下提供以下平台的压缩包：

| 平台 | 文件 |
|---|---|
| Windows amd64 | `GameAgentEngine-windows-amd64-v0.4.5.zip` |
| Windows arm64 | `GameAgentEngine-windows-arm64-v0.4.5.zip` |
| Linux amd64 | `GameAgentEngine-linux-amd64-v0.4.5.zip` |
| Linux arm64 | `GameAgentEngine-linux-arm64-v0.4.5.zip` |
| macOS amd64 (Intel) | `GameAgentEngine-darwin-amd64-v0.4.5.zip` |
| macOS arm64 (Apple Silicon) | `GameAgentEngine-darwin-arm64-v0.4.5.zip` |

每个包包含：

```
GameAgentEngine-{os}-{arch}-v0.4.5/
├── GameAgentEngine.exe     # 后端引擎服务
├── GameAgentDevCli.exe     # 命令行管理工具
├── gameagentengine.conf.yaml  # 默认配置文件
├── demo-world.yaml         # Demo 世界配置
└── web/
    ├── Demo/               # 可玩 Demo 展示
    └── GameAgentCreator/   # 可视化编辑器
```

### 使用预编译包

```bash
# 解压
unzip GameAgentEngine-windows-amd64-v0.4.5.zip
cd GameAgentEngine-windows-amd64-v0.4.5

# 编辑配置文件，填入 LLM API Key
# 编辑 gameagentengine.conf.yaml

# 启动服务
GameAgentEngine serve

# 另开终端，导入 Demo 世界
GameAgentDevCli import demo-world.yaml --reset
```

---

## 源码构建

### 前置条件

- Go 1.25+

### 本地构建

```bash
# 构建全部组件
go build -o GameAgentEngine ./cmd/gameagentengine/
go build -o GameAgentDevCli ./cmd/gameagentdevcli/
```

### 跨平台交叉编译

使用 `tools/scripts/build.sh`（Linux/macOS）或 `tools/scripts/build.bat`（Windows）：

```bash
# 构建所有平台
./tools/scripts/build.sh

# 构建特定平台
GOOS=linux GOARCH=amd64 go build -o dist/GameAgentEngine-linux-amd64 ./cmd/gameagentengine/
GOOS=windows GOARCH=amd64 go build -o dist/GameAgentEngine-windows-amd64.exe ./cmd/gameagentengine/
GOOS=darwin GOARCH=amd64 go build -o dist/GameAgentEngine-darwin-amd64 ./cmd/gameagentengine/
```

打包脚本会自动创建版本目录、重新生成 `tools/source/web/GameAgentCreator/js/component-meta.js`，并复制 `web/` 等资源文件。

---

## 部署

### 基础部署

```bash
# 1. 复制解压后的目录到服务器
# 2. 修改 gameagentengine.conf.yaml 配置
# 3. 运行服务
GameAgentEngine serve
```

### 生产环境建议

1. **修改默认 API Key** — 将 `auth.api_key` 从 `"dev-key"` 改为随机字符串
2. **配置正确的 LLM API Key**
3. **使用 MySQL**（可选）— 生产环境推荐 MySQL 而非 SQLite
4. **反向代理** — 生产环境建议使用 Nginx 反向代理
5. **系统服务** — 建议配置为 systemd（Linux）或 Windows Service

### Nginx 反向代理示例

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

### systemd 服务配置（Linux）

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

## 端口与安全

- 默认端口：8080
- API 认证：通过 `X-API-Key` 请求头
- 默认开发密钥为 `dev-key`，生产环境务必修改
- CORS 默认允许所有来源
