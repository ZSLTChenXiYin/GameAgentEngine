# 构建与部署

**中文** | [**English**](./BUILD_AND_DEPLOY_EN.md)

GameAgentEngine v0.4.6 提供源码构建和预编译打包两种使用方式。

---

## 当前打包内容

每个版本的打包目录结构以 `dist/GameAgentEngine-{os}-{arch}-v0.4.6/` 为准。当前包内包含：

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

当前打包内容不包含 `demo-world.yaml` 或 `web/Demo/`。

---

## 本地源码构建

前置条件：

- Go 1.25+

构建全部组件：

```bash
go build -o GameAgentEngine ./cmd/gameagentengine/
go build -o GameAgentDevCli ./cmd/gameagentdevcli/
```

---

## 使用打包脚本

Windows：

```bash
tools\scripts\build.bat windows/amd64
```

Linux 或 macOS：

```bash
./tools/scripts/build.sh linux/amd64
```

全部平台：

```bash
./tools/scripts/build.sh all
```

打包脚本会自动：

- 编译 `GameAgentEngine` 与 `GameAgentDevCli`
- 注入版本号
- 重新生成 `tools/source/web/GameAgentCreator/js/component-meta.js`
- 复制 `tools/source/` 下的配置、文档和 Creator 静态资源
- 输出 zip 包

---

## 启动打包产物

```bash
cd GameAgentEngine-windows-amd64-v0.4.6
GameAgentEngine serve
```

如果你需要管理世界、打开 Creator、推进 Tick，使用同目录下的 `GameAgentDevCli`。

---

## 部署建议

生产环境至少检查以下事项：

1. 修改 `auth.api_key`，不要保留 `dev-key`
2. 配置真实可用的 LLM API Key
3. 评估是否切换到 MySQL
4. 用反向代理暴露 HTTP 服务
5. 用系统服务托管进程

### Nginx 示例

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

### systemd 示例

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

- 默认端口：`8080`
- API 认证：`X-API-Key`
- 默认开发密钥：`dev-key`
- 默认 CORS：允许所有来源
