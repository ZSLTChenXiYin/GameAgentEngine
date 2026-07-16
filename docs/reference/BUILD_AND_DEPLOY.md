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
├── GameAgentWorker(.exe)
├── gameagentengine.conf.yaml
├── demo-world.yaml
├── demo-state.yaml
├── README.md
├── sdks/
├── tests/
└── web/
    └── GameAgentCreator/
```

当前打包内容来自 `tools/source/` 的运行资产树，因此包含：

- 配置模板
- demo world / authority state 文件
- `tests/` 下的 Worker / SDK 共用测试数据
- `sdks/` 下的多语言 SDK 源码与示例
- Creator 静态资源

当前包内不再附带完整仓库 `docs/` 树；文档入口以包内 `README.md`、仓库根 README 和 GitHub 文档页面为准。

---

## 本地源码构建

前置条件：

- Go 1.25+

构建全部组件：

```bash
go build -o GameAgentEngine ./cmd/gameagentengine/
go build -o GameAgentDevCli ./cmd/gameagentdevcli/
go build -o GameAgentWorker ./cmd/gameagentworker/
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

- 编译 `GameAgentEngine`、`GameAgentDevCli` 与 `GameAgentWorker`
- 注入版本号
- 重新生成 `tools/source/web/GameAgentCreator/js/component-meta.js`
- 复制 `tools/source/` 下的运行资产树
- 在打包输出目录内回写 Creator 最低兼容版本号，不改动源码目录
- 输出 zip 包

---

## 启动打包产物

```bash
cd GameAgentEngine-windows-amd64-v0.4.6
GameAgentEngine serve
```

如果你需要管理世界、打开 Creator、推进 Tick，使用同目录下的 `GameAgentDevCli`。

如果你需要模拟游戏侧异步接口、跑内置集成测试场景，或进入 REPL 试玩，使用同目录下的 `GameAgentWorker`。

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
