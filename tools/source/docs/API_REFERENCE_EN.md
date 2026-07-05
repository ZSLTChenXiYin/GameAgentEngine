# API Reference

[**中文**](./API_REFERENCE.md) | **English**

GameAgentEngine v0.2.0 provides a RESTful HTTP API. All API requests authenticate via the `X-API-Key` header.

---

## Base Path

All endpoints are prefixed with `/api/v1/`. The service listens on `0.0.0.0:8080` by default.

---

## Health & Status

### GET /health

Service health check.

```bash
curl http://127.0.0.1:8080/health
# {"status":"ok"}
```

### GET /api/v1/status

Service status and world list.

```bash
curl -H "X-API-Key: dev-key" http://127.0.0.1:8080/api/v1/status
```

---

## Inference

### POST /api/v1/invoke

Unified inference entry point. Supports all task types.

```json
// Request
{
  "world_id": "world-id",
  "node_id": "target-node-id",
  "task_type": "npc_dialogue",
  "context": {
    "messages": [{"role": "user", "content": "Hello"}]
  }
}
```

```json
// Response
{
  "request_id": "uuid",
  "task_type": "npc_dialogue",
  "reply": "Character response...",
  "action_calls": [],
  "memory_updates": [],
  "world_change_plan": {},
  "metadata": {
    "llm_model": "deepseek-chat",
    "processing_time_ms": 3980
  }
}
```

### POST /api/v1/actions/callback

Async action callback endpoint.

```json
{
  "callback_id": "callback-id",
  "status": "completed",
  "result": {}
}
```

---

## Nodes

### POST /api/v1/nodes

Create a node.

```json
{
  "world_id": "world-id",
  "name": "node-name",
  "node_type": "npc",
  "parent_id": "parent-node-id (optional)"
}
```

### GET /api/v1/nodes/{id}

Get node details (includes components, memories, relations).

### PUT /api/v1/nodes/{id}

Update a node.

### DELETE /api/v1/nodes/{id}

Delete a node (leaf nodes only).

### GET /api/v1/nodes

List nodes with pagination and filtering.

| Parameter | Description |
|---|---|
| `world_id` | Filter by world |
| `node_type` | Filter by type |
| `limit` | Max results |
| `offset` | Offset |

---

## Components

### POST /api/v1/components

Create a component.

```json
{
  "node_id": "node-id",
  "component_type": "profile",
  "data": "{\"name\":\"Elrin\"}"
}
```

### GET /api/v1/components

List components for a node. Parameter: `node_id`.

### GET /api/v1/components/{id}

Get a single component.

### PUT /api/v1/components/{id}

Update a component.

### DELETE /api/v1/components/{id}

Delete a component.

---

## Memories

### POST /api/v1/memories

Create a memory.

```json
{
  "node_id": "node-id",
  "content": "memory content",
  "level": "long_term",
  "tags": "tag1,tag2"
}
```

### GET /api/v1/memories

List memories for a node. Parameter: `node_id`.

### GET /api/v1/memories/{id}

Get a single memory.

### PUT /api/v1/memories/{id}

Update a memory.

### DELETE /api/v1/memories/{id}

Delete a memory.

### POST /api/v1/memories/propagate

Manually propagate a memory.

```json
{
  "memory_id": "memory-id",
  "target_node": "target-node-id",
  "mode": "upward",
  "tags": [],
  "target_ids": [],
  "max_depth": 0,
  "publish_up": false
}
```

---

## Relations

### POST /api/v1/relations

Create a relation.

```json
{
  "world_id": "world-id",
  "source_id": "source-node-id",
  "target_id": "target-node-id",
  "relation_type": "ally",
  "weight": 50,
  "properties": "{}"
}
```

### GET /api/v1/relations

List relations. Parameters: `world_id`, `limit`, `offset`.

### GET /api/v1/relations/{id}

Get a single relation.

### PUT /api/v1/relations/{id}

Update a relation.

### DELETE /api/v1/relations/{id}

Delete a relation.

---

## World Operations

### POST /api/v1/worlds/{world_id}/ticks/advance

Advance the world timeline (Tick).

```json
{
  "tick_type": "scheduled",
  "game_time": "Day 2 - Noon",
  "autonomous_limit": 10
}
```

Response includes the tick record, inference response, and autonomous behavior results.

### POST /api/v1/worlds/{world_id}/events/impact

Evaluate an event's impact.

```json
{
  "event_type": "diplomatic_crisis",
  "scope_id": "scope-node-id",
  "description": "Event description",
  "severity": "critical"
}
```

### POST /api/v1/worlds/{world_id}/ticks/replan

Regenerate the world future outline.

### POST /api/v1/worlds/{world_id}/scope/{scope_id}/advance

Advance evolution within a specific scope.

### POST /api/v1/worlds/{world_id}/clone

Clone a world with all its data. The request body may optionally include `lock_world` (bool) to lock the source world during cloning.

---

## World Settings

### GET /api/v1/worlds/{world_id}/settings

Get world runtime settings.

```json
{
  "world_id": "...",
  "memory_limit": 50,
  "max_analysis_rounds": 5,
  "max_context_depth": 3,
  "auto_apply": true,
  "require_review_above": "critical",
  "pipeline_mode": "full",
  "propagation_max_depth": 2,
  "sub_task_max_retries": 2,
  "sub_task_timeout_secs": 60,
  "enable_propagation_machine": false
}
```

### PUT /api/v1/worlds/{world_id}/settings

Update world runtime settings. Supports partial updates (only send fields to change).

```json
{
  "pipeline_mode": "polling",
  "propagation_max_depth": 3
}
```

---

## World Policy

### GET /api/v1/worlds/{world_id}/policy

Get world action policy.

### PUT /api/v1/worlds/{world_id}/policy

Update world action policy.

```json
{
  "blocked_actions": ["kill_character"],
  "safe_actions": ["add_memory"]
}
```

---

## Autonomous Behavior

### GET /api/v1/nodes/{node_id}/autonomous

Get a node's autonomous behavior configuration.

### PUT /api/v1/nodes/{node_id}/autonomous

Set a node's autonomous behavior configuration.

### POST /api/v1/nodes/{node_id}/autonomous/run

Manually trigger a node's autonomous behavior.

### POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run

(Compatibility path) Manually trigger autonomous behavior.

---

## Logs

### GET /api/v1/logs

Read inference logs.

| Parameter | Description |
|---|---|
| `world_id` | Filter by world |
| `limit` | Max results (default 10) |

---

## Import

### POST /api/v1/creator/import

Import world configuration (YAML or JSON).

```json
{
  "format": "yaml",
  "content": "YAML or JSON string",
  "reset": false,
  "dry_run": false
}
```

---

## Error Codes

| HTTP Status | Description |
|---|---|
| 400 | Invalid request parameters |
| 404 | Resource not found |
| 409 | Conflict (e.g., deleting a non-leaf node) |
| 500 | Internal server error |

Error response format:

```json
{
  "error": "Human-readable error message",
  "code": "Machine-readable error code"
}
```