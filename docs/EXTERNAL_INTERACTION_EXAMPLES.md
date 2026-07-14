# 外部交互接入示例

本文档给出 Engine 外部交互的三种推荐接入方式：`push`、`pull`、`hybrid`。
目标不是强行统一到一种模式，而是帮助游戏侧根据网络形态、部署方式和容错要求做选择。

如果你要确认哪些链路已经进入自动化回归，可以继续查看 [EXTERNAL_INTERACTION_TEST_MATRIX.md](./EXTERNAL_INTERACTION_TEST_MATRIX.md)。

---

## 一、推荐选择

| 场景 | 推荐模式 | 原因 |
|---|---|---|
| Engine 与游戏逻辑服务处于同一受控网络 | `push` | 链路最短，时延最低 |
| 游戏客户端不方便暴露入站服务 | `pull` | 由客户端或 bridge 主动领取，网络适配更稳 |
| 希望优先主动派发，但需要失败回退 | `hybrid` | 主链路快，失败时可回落到队列 |

---

## 二、Push 示例

适合 Engine 直接向游戏侧服务发送 HTTP / WebSocket / RPC 请求。

```yaml
external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:9000"
    path: "/api/v1/runtime/dispatch"
    timeout_ms: 5000
    retry_max_attempts: 2
    retry_backoff_ms: 100
    idempotency_header: "Idempotency-Key"

external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "push"
    primary_transport: "game_http"
    consumer: "game_client"
    resume_policy: "resume_paused_execution"
```

链路：

1. Engine 生成 runtime task。
2. Engine 通过内建 adapter 主动派发。
3. 游戏侧完成后调用 `POST /api/v1/actions/callback`。
4. 如果该任务属于 paused execution，Engine 自动恢复原推理。

---

## 三、Pull 示例

适合游戏客户端、编辑器插件或本地 bridge 主动领取任务。

```yaml
external_interfaces:
  game_client_request_data:
    category: "external_query"
    delivery_mode: "pull"
    consumer: "game_client"
    max_attempts: 5
    resume_policy: "resume_paused_execution"
```

典型流程：

1. `GET /api/v1/runtime/tasks/pending?consumer=game_client`
2. `POST /api/v1/runtime/tasks/claim`
3. `POST /api/v1/runtime/tasks/start`
4. 执行真实游戏侧查询或动作
5. `POST /api/v1/actions/callback`

如果执行中断：

1. `POST /api/v1/runtime/tasks/release`
2. 或长时间无 heartbeat 后由 governor 标记成 `heartbeat_timeout`
3. 再由人工或自动治理执行 `requeue`

---

## 四、Hybrid 示例

适合优先走主动派发，但要求有稳定回退语义的生产环境。

```yaml
external_integrations:
  game_http:
    type: "http_adapter"
    base_url: "http://127.0.0.1:9000"
    path: "/api/v1/runtime/dispatch"

external_interfaces:
  spawn_item:
    category: "external_action"
    delivery_mode: "hybrid"
    primary_transport: "game_http"
    fallback_transport: "task_pull"
    consumer: "bridge"
    max_attempts: 3
    heartbeat_timeout_auto_requeue: true
    heartbeat_timeout_requeue_delay_ms: 2500
    heartbeat_timeout_reason: "spawn_item selective auto requeue"
    resume_policy: "none"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"
```

当前语义要点：

- `fallback_transport` 表示 push 失败后，任务进入 pull 风格的 `released` 状态，并写入该 transport 标签。
- 它当前不是“自动切换到第二个 push adapter 再发一次”。
- `max_attempts` 约束的是 pull/hybrid 阶段的领取重试次数；达到上限后，release 或 timeout requeue 会把任务落到终态 `failed`。
- `heartbeat_timeout_*` 会在任务创建时固化进 payload 快照，避免后续配置漂移影响已经发出的任务。

---

## 五、Callback 后处理示例

普通 async action 现在除了原始 callback 完成语义外，还支持统一后处理基础版：

```yaml
external_interfaces:
  spawn_item:
    category: "external_action"
    delivery_mode: "pull"
    consumer: "bridge"
    callback_post_process: "write_memory"
    callback_memory_level: "long_term"
    callback_memory_template: "spawn callback {status}: {result_json}"
```

当前支持：

- `none`
- `record_only`
- `write_memory`

`write_memory` 模式可用占位符：

- `{action_id}`
- `{status}`
- `{result_json}`
- `{callback_id}`
- `{interface_name}`
- `{task_id}`
- `{node_id}`
- `{world_id}`
- `{request_id}`
- `{delivery_mode}`
- `{primary_transport}`

这里使用的是 runtime task payload 里的策略快照，所以即便接口配置在任务发出后被修改，callback 仍会按原策略执行。

---

## 六、推荐观察点

排查外部交互链路时，优先看这些接口：

- `GET /api/v1/runtime/tasks`
- `GET /api/v1/runtime/tasks/stats`
- `GET /api/v1/runtime/tasks/{task_id}`
- `GET /api/v1/logs`

重点字段：

- `status`
- `delivery_mode`
- `transport`
- `attempt_count`
- `max_attempts`
- `dispatch_attempts`
- `last_dispatch_error`
- `callback_id`
- `resume_execution_id`

---

## 七、最小接入样例

下面给出两个最小化样例，帮助快速接一个 `pull` worker 或 `push` 接收端。

### 7.1 Bridge / 游戏侧 Pull Worker

这个样例适合本地 bridge、编辑器插件或游戏客户端主动领取任务。

```js
const baseUrl = process.env.ENGINE_BASE_URL || 'http://127.0.0.1:8080';
const runtimeTaskToken = process.env.RUNTIME_TASK_TOKEN || 'dev-task-token';
const callbackToken = process.env.CALLBACK_TOKEN || 'dev-callback-token';

async function engineFetch(path, init = {}) {
  const headers = {
    'Content-Type': 'application/json',
    'X-Runtime-Task-Token': runtimeTaskToken,
    ...(init.headers || {}),
  };
  return fetch(baseUrl + path, { ...init, headers });
}

async function callbackFetch(body, requestId) {
  return fetch(baseUrl + '/api/v1/actions/callback', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Callback-Token': callbackToken,
      'X-Callback-Request-Id': requestId,
    },
    body: JSON.stringify(body),
  });
}

async function pollOnce() {
  const pendingResp = await engineFetch('/api/v1/runtime/tasks/pending?consumer=game_client&limit=1');
  const pending = await pendingResp.json();
  const task = pending.tasks?.[0];
  if (!task) return false;

  const claimResp = await engineFetch('/api/v1/runtime/tasks/claim', {
    method: 'POST',
    body: JSON.stringify({ task_id: task.task_id, consumer: 'game_client', lease_owner: 'local-worker-1' }),
  });
  const claimBody = await claimResp.json();
  const claimed = claimBody.task;

  await engineFetch('/api/v1/runtime/tasks/start', {
    method: 'POST',
    body: JSON.stringify({ task_id: claimed.task_id, lease_token: claimed.lease_token }),
  });

  const result = { scene: 'tavern', source: 'pull-worker' };
  await callbackFetch({ callback_id: claimed.callback_id, status: 'success', result }, `cb-${claimed.task_id}`);
  return true;
}

pollOnce().catch((err) => {
  console.error('pull worker failed', err);
  process.exitCode = 1;
});
```

这个最小 worker 的几个关键点：

- `callback_id` 是完成 paused query / action 的稳定句柄。
- 如果任务同时带有 `resume_execution_id`，说明 callback 完成后 Engine 还会继续恢复原来的多轮推理。
- `payload_json` 里可能已经带有 `resume_policy`、`external_interface`、`delivery_mode`、`primary_transport` 等路由快照，worker 侧应以它为执行时契约。

### 7.1.1 请求级动态接口 invoke

如果某段 NPC 对话只想临时暴露一小组查询接口给 LLM，不要把所有函数定义直接塞进 prompt，而是通过请求级 `dynamic_interfaces` 传入白名单。

使用 DevCli：

```bash
GameAgentDevCli invoke <world-id> <node-id> \
  --task-type npc_dialogue \
  --message "回答前先查询附近场景和当前任务状态。" \
  --dynamic-interfaces-json '[
    {
      "id": "scene_query",
      "type": "data_request",
      "description": "从游戏侧查询附近场景和世界状态",
      "query_types": ["node_detail", "world_state"],
      "max_queries": 2,
      "external_interface": {
        "name": "game_client_request_data"
      }
    }
  ]'
```

这里的职责边界是：

- Engine 配置里继续固化稳定的 delivery / consumer / resume_policy
- 游戏侧按请求决定本回合到底开放哪些临时接口
- 模型通过结构化工具定义来发起调用，而不是靠 prompt 里手写函数文本

### 7.2 游戏侧 Push Receiver

这个样例适合 Engine 通过 `http_adapter` 主动向外派发时，游戏侧暴露一个接收端点。

```js
import http from 'node:http';

const port = Number(process.env.GAME_PORT || 9000);

function readJson(req) {
  return new Promise((resolve, reject) => {
    let data = '';
    req.setEncoding('utf8');
    req.on('data', (chunk) => { data += chunk; });
    req.on('end', () => {
      try {
        resolve(data ? JSON.parse(data) : {});
      } catch (err) {
        reject(err);
      }
    });
    req.on('error', reject);
  });
}

http.createServer(async (req, res) => {
  if (req.method !== 'POST' || req.url !== '/api/v1/runtime/dispatch') {
    res.statusCode = 404;
    res.end('not found');
    return;
  }

  const body = await readJson(req);
  console.log('dispatch request', body.task_id, body.payload?.action_id || body.payload?.external_interface);

  res.setHeader('Content-Type', 'application/json');
  res.end(JSON.stringify({ status: 200, accepted: true, worker: 'game-http-server' }));
}).listen(port, () => {
  console.log(`game push receiver listening on ${port}`);
});
```

---

## 八、建议

当前推荐的工程边界是：

- 游戏侧通用、稳定、长期存在的接口，固化到 Engine `external_interfaces` 配置中
- 当前回合、当前 NPC、当前场景临时开放的接口，通过请求级 `dynamic_interfaces` 传入
- callback / resume / post-process 的最终执行语义，仍以 Engine 持久化的 task payload 快照为准
