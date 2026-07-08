# GameAgentDevCli 指南

**中文** | [**English**](./GUIDE_GAMEAGENTDEVCLI_EN.md)

GameAgentDevCli 是通过 HTTP API 操作 GameAgentEngine 的命令行工具。

---

## 全局参数

- `--server`, `-s`：引擎服务地址
- `--key`, `-k`：API Key
- `--config`：本地配置文件路径，用于 reset 等本地操作
- `--idempotency-key`：写请求使用的幂等键
- `--memory-limit`
- `--max-analysis-rounds`
- `--max-context-depth`
- `--include-related-nodes`

---

## 导入

```bash
GameAgentDevCli import tools/source/demo-world.yaml --reset
GameAgentDevCli import tools/source/demo-world.yaml --dry-run
```

---

## 节点命令

```bash
# 创建
GameAgentDevCli node create --world <world-id> --name "议事厅" --type location

# 查询
GameAgentDevCli node get <node-id>
GameAgentDevCli node list --world <world-id>

# 更新
GameAgentDevCli node update <node-id> --name "新名称"

# 移动
GameAgentDevCli node update <node-id> --parent <new-parent-id>
GameAgentDevCli node update <node-id> --clear-parent

# 复制
GameAgentDevCli node copy <node-id>
GameAgentDevCli node copy <node-id> --name "复制节点"
GameAgentDevCli node copy <node-id> --with-children=false

# 删除
GameAgentDevCli node delete <node-id>
```

`node copy` 默认复制整棵子树。

---

## 世界命令

```bash
# 修改世界名称
GameAgentDevCli world update <world-id> --name "新的世界名称"

# 创建工作副本
GameAgentDevCli world fork <world-id> [name] [--lock-world]

# 保存存档快照
GameAgentDevCli world save <world-id> [name] [--lock-world]

# 从快照恢复
GameAgentDevCli world restore <snapshot-world-id> [name] [--lock-world]

# 快照检查
GameAgentDevCli world validate-snapshot <snapshot-world-id>
GameAgentDevCli world snapshot-info <snapshot-world-id>
GameAgentDevCli world list-snapshots <world-id>
GameAgentDevCli world delete-snapshot <snapshot-world-id>

# 审批计划
GameAgentDevCli world plan pending
GameAgentDevCli world plan pending <world-id>
GameAgentDevCli world plan approve <world-id> <plan-id>
GameAgentDevCli world plan reject <world-id> <plan-id>
```

---

## 运行时命令

```bash
GameAgentDevCli world tick <world-id>
GameAgentDevCli world event-impact <world-id> --type crisis --description "..."
GameAgentDevCli world scope-advance <world-id> <scope-id>
GameAgentDevCli world replan <world-id>
```

---

## 世界设置与策略

```bash
# 设置
GameAgentDevCli world settings get <world-id>
GameAgentDevCli world settings set <world-id> --pipeline-mode polling

# 策略
GameAgentDevCli world policy get <world-id>
GameAgentDevCli world policy set <world-id> --blocked spawn_item --safe add_memory
```

`world settings set` 是部分更新命令，只有显式传入的字段会被修改。

---

## 记忆传播

```bash
GameAgentDevCli memory propagate <memory-id>
GameAgentDevCli memory propagate <memory-id> --mode tag_broadcast --tags rumor,politics
GameAgentDevCli memory propagate <memory-id> --mode targeted --target node-a,node-b
GameAgentDevCli memory propagate <memory-id> --max-depth 2 --publish-up
```

---

## 日志与调试轨迹

```bash
GameAgentDevCli logs --world <world-id> --limit 10
GameAgentDevCli logs --world <world-id> --limit 10 --json

GameAgentDevCli debug traces --world <world-id> --limit 10
GameAgentDevCli debug traces --world <world-id> --limit 10 --json
```

---

## 打开 Creator

```bash
GameAgentDevCli inspect
```

当你的运行环境暴露了 Creator 检查入口时，可以使用这个命令。
