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
    "processing_time_ms": 3980,
    "configured_pipeline_mode": "polling",
    "effective_pipeline_mode": "vertical",
    "max_analysis_rounds": 4,
    "rounds_used": 1
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

### POST /api/v1/worlds/{world_id}/timeline/replan

Regenerate the world future outline.

### POST /api/v1/worlds/{world_id}/scopes/{scope_id}/advance

Advance evolution within a specific scope.

### POST /api/v1/worlds/{world_id}/fork

Create a working-copy fork of a world. The request body may optionally include `name` and `lock_world`; when `lock_world` is `true`, the source world is locked during copying.

### POST /api/v1/worlds/{world_id}/snapshots

Create a save snapshot of a world. The request body may optionally include `name` and `lock_world`; when `lock_world` is `true`, the source world is locked during snapshotting.

### POST /api/v1/worlds/{world_id}/restore

Restore a saved snapshot into a new world. `world_id` should be a snapshot world ID. The request body may optionally include `name` and `lock_world`; when `lock_world` is `true`, the snapshot source world is locked during restore.

### DELETE /api/v1/worlds/{world_id}/snapshot

Delete a saved snapshot world together with its persisted snapshot metadata. `world_id` should be a snapshot world ID whose snapshot reason is `save_snapshot`.

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
  "propagation_max_depth": 0,
  "sub_task_max_retries": 0,
  "sub_task_timeout_secs": 0
}
```

Notes:

- `memory_limit`, `max_analysis_rounds`, and `max_context_depth` must be greater than `0` when provided.
- `propagation_max_depth`, `sub_task_max_retries`, and `sub_task_timeout_secs` may be set to `0` explicitly.
- `pipeline_mode` must be one of `vertical`, `polling`, or `full`.

### GET /api/v1/worlds/{world_id}/snapshot-validation

Validate whether a saved snapshot can still be safely restored. Returns a structured report describing version compatibility and drift checks.

### GET /api/v1/worlds/{world_id}/snapshot-metadata

Retrieve snapshot metadata for a copied world.

### GET /api/v1/worlds/{world_id}/snapshots

List all save snapshots created from a source world. Only `save_snapshot` records are returned.

### Invoke Response Metadata

`metadata` may include these pipeline observability fields in addition to the standard model/timing values:

- `configured_pipeline_mode`: the world settings pipeline mode before request-level overrides
- `effective_pipeline_mode`: the mode actually used for this execution
- `max_analysis_rounds`: the resolved round budget for this execution
- `rounds_used`: how many rounds the execution consumed

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
