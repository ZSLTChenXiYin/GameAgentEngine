# GameAgentDevCli 指南

**中文** | [**English**](./GUIDE_GAMEAGENTDEVCLI_EN.md)

GameAgentDevCli 是 GameAgentEngine v0.2.0 的命令行管理工具。它通过 HTTP API 与引擎服务通信，支持世界管理、节点 CRUD、组件、记忆、关系操作，以及世界推理（Tick、事件影响、局部推进）等功能。

---

## 全局参数

| 参数 | 简写 | 说明 |
|---|---|---|
| `--server <url>` | `-s` | 引擎服务地址（默认 `http://127.0.0.1:8080`） |
| `--key <key>` | `-k` | API 密钥（默认 `dev-key`） |
| `--config <path>` | | 本地配置文件路径（用于 reset 等本地操作） |
| `--memory-limit <n>` | | 推理记忆上限（0=使用服务端配置） |
| `--max-analysis-rounds <n>` | | LLM 最大轮询次数 |
| `--max-context-depth <n>` | | 上下文追溯最大深度 |
| `--include-related-nodes` | | 加载关联节点数据 |
| `--idempotency-key <key>` | | 幂等 key |

---

## 命令概览

### status — 检查服务状态

```bash
GameAgentDevCli status
```

### reset — 清空本地数据库

```bash
GameAgentDevCli reset --config gameagentengine.conf.yaml
```

### import — 导入世界配置

导入 YAML/JSON 格式的世界配置（支持 `--dry-run` 纯校验和 `--reset` 清空后导入）：

```bash
# 从文件导入
GameAgentDevCli import demo-world.yaml

# 校验导入内容（不写入数据库）
GameAgentDevCli import demo-world.yaml --dry-run

# 清空数据库后导入
GameAgentDevCli import demo-world.yaml --reset

# 从标准输入导入
cat world.yaml | GameAgentDevCli import - --format yaml
```

导入文件格式详见 Demo 世界示例和核心概念文档。

---

### node — 节点管理

```bash
# 创建世界节点
GameAgentDevCli node create --name "MyWorld" --type "world"

# 创建子节点
GameAgentDevCli node create --world <world-id> --name "议事厅" --type "location" --parent <parent-id>

# 列出所有节点
GameAgentDevCli node list --world <world-id>

# 查看节点详情（含组件、记忆、关系）
GameAgentDevCli node get <node-id>

# 更新节点
GameAgentDevCli node update <node-id> --name "新名称" --type "npc"

# 删除节点（叶子节点才能删除）
GameAgentDevCli node delete <node-id>
```

#### 自主行为管理

```bash
# 查看节点自主行为配置
GameAgentDevCli node autonomous get <node-id>

# 配置自主行为
GameAgentDevCli node autonomous set <node-id> --enabled --trigger "world_tick_sync"

# 禁用自主行为
GameAgentDevCli node autonomous disable <node-id>

# 手动触发一次自主行为
GameAgentDevCli node autonomous run <node-id>
```

---

### world — 世界级运行时操作

#### 世界管理命令

```bash
# 复制世界
GameAgentDevCli world clone <world-id> [name] [--lock]
```

- `--lock` / `-l`：复制期间锁定源世界，阻止并发写入（可选，默认不锁定）
- `name`：为新世界指定名称，留空则自动生成"原名 (副本)"

#### 世界设置

```bash
# 查看世界运行设置
GameAgentDevCli world settings get <world-id>

# 设置运行参数
GameAgentDevCli world settings set <world-id> \
  --memory-limit 100 \
  --analysis-rounds 8 \
  --context-depth 5 \
  --pipeline-mode "full" \
  --propagation-max-depth 3 \
  --sub-task-max-retries 3 \
  --sub-task-timeout-secs 120 \
  --enable-propagation-machine true
```

#### 世界策略

```bash
# 查看策略
GameAgentDevCli world policy get <world-id>

# 设置阻止/安全动作
GameAgentDevCli world policy set <world-id> \
  --blocked "kill_character,nuclear_strike" \
  --safe "add_memory,send_dialogue"
```

#### 世界推理

```bash
# 推进世界时间（Tick）
GameAgentDevCli world tick <world-id> --type "scheduled" --time "第2天-中午"

# 评估事件影响
GameAgentDevCli world event-impact <world-id> \
  --type "diplomatic_crisis" \
  --scope <scope-id> \
  --description "邻国在边境集结军队..." \
  --severity "critical"

# 局部范围推进
GameAgentDevCli world scope-advance <world-id> <scope-id>

# 重新生成世界大纲
GameAgentDevCli world replan <world-id>
```

#### 快照与导出

```bash
# 输出世界运行时快照
GameAgentDevCli world snapshot <world-id>

# 导出世界配置（用于备份或迁移）
GameAgentDevCli world export <world-id> --format yaml --out myworld.yaml
```

---

### component — 组件管理

```bash
# 列出节点组件
GameAgentDevCli component list --node <node-id>

# 获取单个组件
GameAgentDevCli component get <component-id>

# 创建组件（data 建议传 JSON 字符串）
GameAgentDevCli component create --node <node-id> --type "profile" --data '{"name":"艾琳"}'

# 更新组件
GameAgentDevCli component update <component-id> --data '{"name":"艾琳议长"}'

# 删除组件
GameAgentDevCli component delete <component-id>
```

---

### memory — 记忆管理

```bash
# 列出节点记忆
GameAgentDevCli memory list --node <node-id>

# 创建记忆
GameAgentDevCli memory create --node <node-id> --content "..." --level "long_term" --tags "history"

# 获取记忆
GameAgentDevCli memory get <memory-id>

# 更新记忆
GameAgentDevCli memory update <memory-id> --content "..." --level "shared"

# 删除记忆
GameAgentDevCli memory delete <memory-id>
```

---

### relation — 关系管理

```bash
# 列出关系
GameAgentDevCli relation list --world <world-id>

# 创建关系
GameAgentDevCli relation create --world <world-id> --source <node-id> --target <node-id> --type "ally" --weight 50

# 获取关系
GameAgentDevCli relation get <relation-id>

# 更新关系
GameAgentDevCli relation update <relation-id> --weight 80

# 删除关系
GameAgentDevCli relation delete <relation-id>
```

---

### logs — 推理日志

```bash
# 读取最近的推理日志
GameAgentDevCli logs --world <world-id> --limit 10
```

---

### verify — 验证

```bash
# 验证导入内容
GameAgentDevCli verify import demo-world.yaml
```

---

### inspect — 打开 Creator

```bash
GameAgentDevCli inspect
```

---

## 使用示例：完整工作流

```bash
# 1. 启动引擎（终端 1）
GameAgentEngine serve

# 2. 导入 Demo 世界（终端 2）
GameAgentDevCli import tools/source/demo-world.yaml --reset

# 3. 查看世界状态
GameAgentDevCli status

# 4. 查看运行设置
GameAgentDevCli world settings get <world-id>

# 5. 推进世界时间
GameAgentDevCli world tick <world-id>

# 6. 查看推理日志
GameAgentDevCli logs --world <world-id>
```