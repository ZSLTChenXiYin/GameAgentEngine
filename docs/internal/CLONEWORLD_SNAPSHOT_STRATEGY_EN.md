# CloneWorld Snapshot Strategy

[**中文**](./CLONEWORLD_SNAPSHOT_STRATEGY.md) | **English**

> Historical note: this document is retained as design background. The current engine API has moved to `ForkWorld`, `CreateWorldSnapshot`, and `RestoreWorld`, and no longer exposes `CloneWorld` as a compatibility interface.

This document describes the recommended evolution path for `CloneWorld` when it is used as a save-game mechanism.

The current `CloneWorld` implementation is closer to "duplicate a world and all of its business data".
If it is meant to support game saves, the optimization target should include more than copy speed:

- snapshot compatibility
- restore reliability
- version migration support
- large-world copy performance
- save lifecycle management

---

## 1. Goal

Under save-game semantics, `CloneWorld` should aim to:

1. preserve a complete, restorable Agent world state at a point in time
2. let future Engine versions safely identify and restore old saves
3. reduce the risk that runtime schema changes make older saves unusable
4. keep acceptable save latency for medium and large worlds

That means the internal design should gradually move from plain "world clone" toward "world snapshot".

---

## 2. Recommended semantic split

Long term, it is better to separate two related but distinct needs.

### A. CloneWorld

Use cases:

- branch simulation
- debug copies
- world template duplication
- alternate timeline forks

Characteristics:

- preserve the current API style
- prioritize copy throughput
- tolerate lightweight metadata only

### B. SnapshotWorld / RestoreWorld

Use cases:

- save games
- autosave
- pre-event checkpoints
- rollback / restore

Characteristics:

- prioritize compatibility and recovery safety
- must include version metadata
- may run schema or data migrations before restore

If short-term API expansion is not desired, `CloneWorld` can still start adopting snapshot-oriented metadata internally.

---

## 3. Minimum snapshot metadata

Each save should eventually carry a dedicated metadata header with at least:

| Field | Purpose |
|---|---|
| `snapshot_id` | unique snapshot identifier |
| `source_world_id` | source world UUID |
| `source_world_name` | source world name |
| `created_at` | snapshot creation time |
| `engine_version` | Engine version that created the snapshot |
| `schema_version` | data/schema version at save time |
| `content_version` | optional domain content version |
| `reason` | manual save / autosave / checkpoint / test clone |
| `node_count` | node count |
| `component_count` | component count |
| `memory_count` | memory count |
| `relation_count` | relation count |
| `payload_hash` | snapshot payload checksum |

These fields make it possible to:

- validate restore compatibility
- detect corruption
- show save slots meaningfully
- support migration tooling later

---

## 4. Recommended data shape

The snapshot logic should be split into two layers.

### Snapshot Header

Stores version, source, statistics, and checksum metadata.

### Snapshot Payload

Stores the actual world state:

- world node
- nodes
- components
- memories
- relations
- world_settings
- world_policy
- propagation chains (if enabled)
- future extensible entities

Benefits:

- save lists do not need the full payload
- restore can validate the header first
- payload encoding can evolve later without breaking the outer structure

---

## 5. Compatibility strategy

This is the most important layer for save-game design.

### 5.1 Always record versions

Copying business tables without `engine_version` and `schema_version` makes reliable restore decisions impossible later.

### 5.2 Add a restore gate before loading data

Suggested restore flow:

1. read snapshot header
2. check `schema_version`
3. restore directly if it matches the current runtime
4. run migration if it is older than current
5. reject or mark read-only if it is newer than current

### 5.3 Prefer logical migrations over database-file coupling

If the Engine needs to support SQLite, MySQL, and future schema evolution, logical world snapshots are safer than raw database file copies.

Logical snapshots are better for:

- cross-database restore
- field-level migration
- long-term version evolution

---

## 6. CloneWorld performance roadmap

Without changing external behavior, the best optimization path is:

### 6.1 Build mapping tables once

The biggest current cost is repeated lookup churn.

Prebuild in-memory maps such as:

- `oldNodeID -> oldUUID`
- `oldUUID -> newUUID`
- `newUUID -> newNodeID`

Then let component, memory, and relation copying use memory lookups rather than repeated database reads.

### 6.2 Batch reads and batch writes

For larger worlds:

- batch-load nodes
- batch-load components
- batch-load memories
- batch-load relations
- batch-insert by entity type

This will significantly reduce SQL traffic inside the transaction.

### 6.3 Remove repeated resolution work inside the transaction

Examples:

- do not repeatedly resolve UUID to int64 IDs
- do not repeatedly resolve parent UUIDs per row
- do not reload world settings or world policy inside row loops

### 6.4 Keep explicit source-world locking semantics for saves

If `CloneWorld` is used for saves, locking semantics should remain explicit:

- is a strongly consistent snapshot required?
- are in-flight writes visible?
- is an approximate snapshot acceptable for speed?

Default recommendation: use strongly consistent snapshots for save-game flows.

---

## 7. Suggested restore flow

If `RestoreWorld` is introduced later, a safe flow would be:

1. read snapshot header
2. verify `payload_hash`
3. validate `schema_version`
4. run migration if needed
5. create a new world or restore into an explicit target
6. rebuild nodes / components / memories / relations in batches
7. restore settings / policy / propagation data
8. write restore logs

Direct overwrite of a live production world should be avoided unless the API explicitly requests it and proper locking is applied.

---

## 8. Suggested short-term rollout order

### Phase 1: keep the API, improve internals

- enrich clone results with metadata
- introduce batch mapping in the copy flow
- reduce transaction-time N+1 queries

### Phase 2: add explicit snapshot metadata

- add a snapshot header table or logical object
- write version and statistics metadata for each save/clone

### Phase 3: introduce restore support

- version gate before restore
- migration hooks
- full restore path

### Phase 4: formally split save snapshots from plain clones

- keep `CloneWorld`
- add `SnapshotWorld` / `RestoreWorld`

---

## 9. Conclusion

If `CloneWorld` mainly serves as a save-game mechanism, the right direction is not just "make copying faster".

It should evolve into a world snapshot that is:

- checksum-verifiable
- version-aware
- migration-friendly
- restorable

and then be optimized for throughput on top of that.

The next best steps are:

1. design the snapshot metadata model
2. refactor CloneWorld internals to use batch mapping + batch writes

Once those two are in place, a future restore design becomes much more straightforward.
