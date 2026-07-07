# API Reference

[**中文**](./API_REFERENCE.md) | **English**

All engine API routes use the `X-API-Key` header unless otherwise noted.

Base prefix: `/api/v1/`

Many mutation endpoints also support an `Idempotency-Key` header so callers can safely retry long-running operations.

---

## Health and Version

### `GET /health`

Health check. No API key required.

### `GET /api/v1/version`

Returns the current engine version and minimum compatible version.

---

## Inference and Planning

### `POST /api/v1/invoke`

Unified inference entry point.

Typical request body:

```json
{
  "world_id": "world-id",
  "node_id": "node-id",
  "task_type": "npc_dialogue",
  "messages": [
    { "role": "user", "content": "Hello" }
  ],
  "context": {
    "pipeline_mode": "polling",
    "max_analysis_rounds": 4
  }
}
```

### `POST /api/v1/actions/callback`

Completes an async action callback.

### `GET /api/v1/plans/pending`

Lists plans waiting for manual review.

### `POST /api/v1/worlds/{world_id}/plan/approve`

Approves a pending world change plan.

### `POST /api/v1/worlds/{world_id}/plan/reject`

Rejects a pending world change plan.

---

## Nodes

### `GET /api/v1/nodes`

List nodes.

Query parameters:

- `world_id`
- `node_type`
- `limit`
- `offset`

### `GET /api/v1/nodes/{id}`

Returns a node detail payload with the node, children, components, memories, and relations.

### `POST /api/v1/nodes`

Create a node.

Typical request body:

```json
{
  "world_id": "world-id",
  "name": "Village Guard",
  "node_type": "npc",
  "parent_id": "optional-parent-id"
}
```

### `PUT /api/v1/nodes/{id}`

Update mutable node fields such as `name`, `node_type`, and `parent_id`.

### `POST /api/v1/nodes/{id}/copy`

Copy a node inside the same world.

Request body:

```json
{
  "name": "Copied Node",
  "parent_id": "optional-parent-id",
  "include_descendants": true
}
```

Behavior notes:

- world nodes are not copied through this route
- subtree copy is enabled by default
- internal subtree relations are copied when both ends remain inside the copied set

### `DELETE /api/v1/nodes/{id}`

Delete a leaf node.

### `GET /api/v1/nodes/{node_id}/autonomous`

Read autonomous configuration for a node.

### `PUT /api/v1/nodes/{node_id}/autonomous`

Create or update autonomous configuration for a node.

### `POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run`

Manually trigger one autonomous run for a node.

---

## Components

- `POST /api/v1/components`
- `GET /api/v1/components`
- `GET /api/v1/components/{id}`
- `PUT /api/v1/components/{id}`
- `DELETE /api/v1/components/{id}`

Component mutation routes now validate payload data by component type:

- `autonomous`: must be a valid JSON object and satisfy the Engine's current strong field checks
- `profile`: must be a valid JSON object, but field shape remains open
- other current built-in types: treated as free text unless structured rules are later added for that type

If the payload does not satisfy the component rules, the API returns `400` with error code `invalid_component_data`.

---

## Memories

- `POST /api/v1/memories`
- `GET /api/v1/memories`
- `GET /api/v1/memories/{id}`
- `PUT /api/v1/memories/{id}`
- `DELETE /api/v1/memories/{id}`
- `POST /api/v1/memories/propagate`

The propagation route lets callers explicitly drive memory distribution rules such as `mode`, `target_ids`, `tags`, `max_depth`, and `publish_up`.

---

## Relations

- `POST /api/v1/relations`
- `GET /api/v1/relations`
- `GET /api/v1/relations/{id}`
- `PUT /api/v1/relations/{id}`
- `DELETE /api/v1/relations/{id}`

---

## World Management

### `GET /api/v1/worlds`

List world root nodes.

### `PUT /api/v1/worlds/{world_id}`

Update mutable world fields.

Current request body:

```json
{
  "name": "Renamed World"
}
```

---

## World Runtime Operations

### `POST /api/v1/worlds/{world_id}/ticks/advance`

Advance one world tick.

Typical request body:

```json
{
  "tick_type": "scheduled",
  "game_time": "Day 3 - Evening",
  "autonomous_limit": 10
}
```

### `POST /api/v1/worlds/{world_id}/events/impact`

Evaluate how an event affects the world.

### `POST /api/v1/worlds/{world_id}/scopes/{scope_id}/advance`

Advance a non-world scope node.

### `POST /api/v1/worlds/{world_id}/timeline/replan`

Rebuild the future outline for a world.

---

## Snapshot and World Copy Flows

### `POST /api/v1/worlds/{world_id}/fork`

Create a runnable working copy.

### `GET /api/v1/worlds/{world_id}/snapshots`

List save snapshots created from a source world.

### `POST /api/v1/worlds/{world_id}/snapshots`

Create a save snapshot world.

### `DELETE /api/v1/worlds/{world_id}/snapshot`

Delete a save snapshot world and its metadata.

### `GET /api/v1/worlds/{world_id}/snapshot-metadata`

Read snapshot metadata for a copied world.

### `GET /api/v1/worlds/{world_id}/snapshot-validation`

Validate whether a snapshot remains restorable.

### `POST /api/v1/worlds/{world_id}/restore`

Restore a save snapshot into a fresh runnable world.

All copy routes accept an optional body like:

```json
{
  "name": "Optional Name",
  "lock_world": true
}
```

---

## World Settings and Policy

### `GET /api/v1/worlds/{world_id}/settings`
### `PUT /api/v1/worlds/{world_id}/settings`

Get or partially update world runtime settings.

Important fields:

- `memory_limit`
- `max_analysis_rounds`
- `max_context_depth`
- `auto_apply`
- `require_review_above`
- `pipeline_mode`
- `propagation_max_depth`
- `sub_task_max_retries`
- `sub_task_timeout_secs`
- `enable_propagation_machine`

### `GET /api/v1/worlds/{world_id}/policy`
### `PUT /api/v1/worlds/{world_id}/policy`

Get or update world action policy.

---

## Creator Import

### `POST /api/v1/creator/import`

Import world configuration from YAML or JSON payload text.

Typical request body:

```json
{
  "format": "yaml",
  "content": "...",
  "reset": false,
  "dry_run": false
}
```

---

## Logs and Debugging

### `GET /api/v1/logs`

Read inference logs. Supports `world_id`, `task_type`, `limit`, and `offset` query parameters.

### `GET /debug/traces`

Read debug traces.

---

## Response Metadata

Inference responses may include:

- `configured_pipeline_mode`
- `effective_pipeline_mode`
- `max_analysis_rounds`
- `rounds_used`

These fields are especially useful when different games choose different `pipeline_mode` levels.

---

## Error Model

Typical statuses:

- `400` invalid request
- `404` not found
- `409` conflict
- `500` internal error

Error payload shape:

```json
{
  "error": "message",
  "code": "machine_code"
}
```
