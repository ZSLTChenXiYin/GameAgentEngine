# GitHub 自动发布方案

本文件用于固化 GameAgentEngine v0.5.0 之后的 GitHub 自动构建与发布方案。当前阶段只做设计，不直接提交 workflow 实现。

## 1. 目标

后续 GitHub 自动化分为两层：

1. CI 构建验证
2. Tag 驱动的正式 Release 打包发布

当前方案假设仓库主线产物结构已经稳定为：

- `gameagentengine.conf.yaml`
- `workerhome/demo/`
- `workerhome/fixtures/`
- `sdks/`
- `web/GameAgentCreator/`

## 2. 建议的 workflow 划分

### 2.1 ci.yml

触发条件：

- push 到主分支
- pull request

职责：

- Go build
- 基础测试
- 打包 smoke check
- 产物结构检查

建议检查项：

- `go build ./...`
- 必要单元测试
- 执行打包脚本（至少一个主平台）
- 检查产物中是否存在：
  - `gameagentengine.conf.yaml`
  - `workerhome/demo/demo-world.yaml`
  - `workerhome/demo/demo-state.yaml`
  - `workerhome/fixtures/runtime_task_dynamic_interfaces.json`
  - `web/GameAgentCreator/index.html`

### 2.2 release.yml

触发条件：

- push tag，格式如 `v0.5.0`

职责：

- 多平台矩阵构建
- 运行打包脚本
- 生成 zip 产物
- 创建 GitHub Release
- 上传 Release Assets

建议平台矩阵：

- windows/amd64
- windows/arm64
- linux/amd64
- linux/arm64
- darwin/amd64
- darwin/arm64

## 3. 发布前提

在引入 GitHub 自动发布前，仓库应先满足这些前提：

1. 版本号只存在一个正式主来源，并能被打包脚本注入
2. 打包目录结构稳定
3. Worker Demo / fixtures 目录稳定
4. 打包文档与真实产物一致
5. 不再存在依赖本地手工路径猜测的 Demo 专属逻辑

## 4. 推荐执行流程

### CI 层

1. checkout
2. setup Go
3. build all
4. run tests
5. run package smoke build
6. validate packaged asset tree

### Release 层

1. checkout tag
2. setup Go
3. matrix build by platform
4. run packaging script
5. archive outputs
6. create GitHub Release
7. upload archives

## 5. 当前明确不做的事

当前阶段暂不直接提交：

- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- GitHub Release notes 自动生成逻辑
- Gitee / 其他镜像站同步发布逻辑

这些工作留到后续单独实施。

## 6. 后续实施建议

真正开始实现时，建议顺序为：

1. 先做 `ci.yml`
2. 稳定后再做 `release.yml`
3. 最后再补 release note、checksum、镜像发布等增强项
