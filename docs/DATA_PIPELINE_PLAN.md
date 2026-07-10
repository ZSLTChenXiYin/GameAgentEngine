# Data Pipeline Plan

This document is the durable source of truth for the database pipeline rollout.
It must preserve the full plan, stage boundaries, completed work, verification rules,
and next actions even if conversation context is compressed.

## Core Goals

1. Provide a unified data access pipeline for current and future databases.
2. Keep business code on shared read, write, transaction, batch, and lock entrypoints.
3. Separate driver strategy from business behavior.
4. Ensure SQLite is stable under write pressure.
5. Preserve concurrent read/write capability for MySQL and future server databases.
6. Make every phase testable, gray-release friendly, and reversible.

## Long-Term Architecture

1. Unified pipeline API.
   Reader, Writer, WriteTransaction, BatchSink, WorldLock, RetryPolicy, MigrationRunner.
2. Driver adapters.
   SQLiteAdapter, MySQLAdapter first; PostgreSQL, MariaDB, TiDB later.
3. Strategy layer.
   SQLite uses single-writer plus concurrent reads; MySQL-class databases use pooled concurrent transactions.
4. Repository and service stability.
   Business code should not embed driver-specific behavior.
5. Observability and operations.
   Track queue depth, retry counts, lock waits, slow SQL, transaction time.

## Full Development Plan

### Phase 0: Freeze Current Entry Rules

Scope:
1. Inventory all database read and write entrypoints.
2. Prevent new business writes from bypassing the store pipeline.
3. Record lock contention baseline and hotspot transactions.

Acceptance:
1. Direct business writes outside the unified store pipeline are treated as violations.
2. Existing hotspots are documented and traceable.

### Phase 1: Unified Pipeline Interface

Scope:
1. Define shared read, write, transaction, batch, migration, retry, and lock boundaries.
2. Keep the service and engine layers insulated from driver details.

Acceptance:
1. Service code uses stable shared entrypoints.
2. New database backends can be added in the adapter layer without changing business flows.

### Phase 2: Connection and Transaction Foundation

Scope:
1. Implement Reader, Writer, WriteTransaction, Close, and HealthCheck semantics.
2. Make pool strategy driver-aware.

Acceptance:
1. SQLite and MySQL can be initialized through the same top-level flow.

### Phase 3: SQLite Adapter

Scope:
1. Enable WAL, busy timeout, synchronous mode, and foreign keys.
2. Use a single write connection and concurrent reads.
3. Reduce implicit write transaction overhead where safe.

Acceptance:
1. Lock contention is reduced under common write pressure.

### Phase 4: MySQL Adapter

Scope:
1. Preserve concurrent read/write behavior.
2. Support connection pooling, deadlock retry, and timeout handling.
3. Keep room for future read/write splitting.

Acceptance:
1. MySQL is not degraded to a global single-writer model.

### Phase 5: Migrate Existing Writes to Unified Pipeline

Scope:
1. Route store-level writes through Writer and WriteTransaction.
2. Route service-level multi-step writes through WriteTransaction.

Acceptance:
1. Non-test business writes no longer bypass the shared pipeline.

### Phase 6: Log Batch Pipeline

Scope:
1. Buffer InferenceLog writes in memory.
2. Flush by size or time.
3. Provide safe fallback on queue pressure.
4. Guarantee a consistent view before log reads.
5. Close and flush safely when reinitializing or shutting down.

Acceptance:
1. High-frequency log writes are reduced from fragmented inserts to grouped persistence.

### Phase 7: Memory and Propagation Batch Pipeline

Scope:
1. Batch direct memory writes from pipeline responses.
2. Batch propagation inserts where possible.
3. Reduce repeated count-plus-insert round trips.
4. Keep behavior identical while shrinking write fragmentation.

Acceptance:
1. World tick and autonomous flows generate fewer write transactions.

### Phase 8: Business-Level Concurrency Control

Scope:
1. Add world-level mutual exclusion for critical operations.
2. Distinguish regular writes from exclusive heavy tasks.

Acceptance:
1. Conflicting heavy operations for the same world do not run concurrently.

### Phase 9: Retry and Recovery Layer

Scope:
1. Normalize retriable database errors.
2. Retry SQLite lock conflicts and server-database deadlocks where safe.

Acceptance:
1. Short-lived conflicts recover automatically when semantics allow.

### Phase 10: Unified Migration Control

Scope:
1. Centralize migrations.
2. Prepare for future backends that need stricter schema evolution.

Acceptance:
1. Schema management is not tied to ad hoc startup behavior.

### Phase 11: Add Next Database Backends

Scope:
1. Add PostgreSQL next to validate abstraction quality.
2. Evaluate MariaDB and TiDB reuse versus dedicated adapters.

Acceptance:
1. Business code remains unchanged while the backend varies.

### Phase 12: Observability and Load Testing

Scope:
1. Add queue, retry, lock, and transaction metrics.
2. Build cross-database load scenarios.

Acceptance:
1. Each backend has measurable regression thresholds.

### Phase 13: Rollout and Fallback Strategy

Scope:
1. Add feature flags for each major capability.
2. Support fast downgrade without removing the shared pipeline.

Acceptance:
1. Operations can disable risky features without reverting business code.

## Driver Strategy Rules

### SQLite

1. Single writer.
2. Concurrent reads.
3. WAL plus timeout tuning.
4. Batch logs and memory aggressively.
5. Use world-level serialization for critical business paths when needed.

### MySQL and Future Server Databases

1. Concurrent pooled writes.
2. Concurrent reads.
3. Retry deadlocks and lock wait timeouts where safe.
4. Keep batch sinks for throughput efficiency, not just lock avoidance.
5. Use world-level locks only for business consistency on critical operations.

## Stage Status

### Completed

1. Unified SQLite and MySQL read/write entrypoints.
2. Writer and WriteTransaction plumbing through hotspot write paths.
3. SQLite writer strategy with WAL and tuned connection policy.
4. Log batch pipeline with queue, timed flush, size-based flush, fallback direct writes,
   explicit flush before reads, and sink close on reinit/shutdown.
5. Memory batch helpers for direct pipeline memory writes.
6. Batched propagation target persistence for environment and organization style fan-out paths.
7. World-level exclusion now guards heavy same-world service operations while allowing different worlds to proceed independently.
8. Shared write retry handling now covers SQLite lock conflicts and MySQL-style deadlock or lock wait failures on centralized write paths.
9. Database migrations now run through a dedicated shared runner instead of being embedded ad hoc in initialization flow.
10. PostgreSQL now initializes through the same adapter entrypoints, validating that the shared pipeline can absorb a new backend without business-layer changes.
11. Lightweight pipeline observability is now queryable through structured stats instead of relying only on ad hoc logs.

### In Progress

1. Feature flags and fallback controls.

### Pending After Current Phase

1. Durable plan tracking in repository.

## Verification Rules

Every phase must include:
1. Code changes.
2. Automated verification.
3. A git commit after the phase is complete.
4. Updated status in this document.

## Completed Phase Evidence

### Phase 5 Evidence

1. Store writes route through Writer.
2. Service transactions route through WriteTransaction.
3. Core tests passed after migration.

### Phase 6 Evidence

1. InferenceLog writes now use an internal sink.
2. Log reads flush buffered data before querying.
3. Reinitialization closes any previous sink.
4. Store, service, and engine regression tests passed.

### Phase 7 Evidence

1. Direct pipeline memory writes now go through a batch persistence helper.
2. Propagation target writes for grouped target sets now use batched inserts with dedupe filtering.
3. Behavior-preserving engine and service regression tests passed after the refactor.

### Phase 8 Evidence

1. `AdvanceWorldTickWithAutonomous`, `RunAutonomousNode`, `RunScheduledAutonomous`, world copy flows, and snapshot deletion now share world-level exclusion boundaries.
2. World copy and restore operations now enforce the same-world lock even when older callers omit the legacy `lock_world` flag.
3. World lock tests verify same-world serialization and different-world concurrency.
4. Store, service, and engine regression tests passed after the change.

### Phase 9 Evidence

1. Store writes and `WriteTransaction` now share a bounded retry layer instead of each callsite handling transient conflicts ad hoc.
2. SQLite `database is locked` style failures and MySQL deadlock or lock-wait timeout failures are normalized as retriable write conflicts.
3. Batched log persistence now retries through the same shared write retry layer.
4. Store retry tests plus store, service, and engine regression tests passed after the change.

### Phase 10 Evidence

1. Schema and data migration steps now run through `MigrationRunner` and `RunMigrations` instead of being hard-coded inline inside `Init()`.
2. The migration runner supports ordered reusable steps, which future adapters can share.
3. Migration runner tests verify ordered execution and step-name error reporting.
4. Store, service, and engine regression tests passed after the refactor.

### Phase 11 Evidence

1. PostgreSQL initializes through the same store adapter flow as SQLite and MySQL.
2. Shared retry logic now recognizes PostgreSQL deadlock and lock-not-available conflicts.
3. Config source and initialization template now advertise PostgreSQL as a supported driver.
4. Store, service, and engine regression tests passed after the adapter addition.

### Phase 12 Evidence

1. The store layer now exposes structured stats for write retries, transactions, and log sink queue or flush behavior.
2. World-level lock stats are exposed alongside store pipeline stats.
3. `GET /api/v1/pipeline/stats` now provides a stable diagnostics surface for load testing and troubleshooting.
4. Store, API, service, and engine regression tests passed after the addition.

## Current Next Action

Implement Phase 13.

Concrete targets:
1. Add feature flags for major pipeline capabilities.
2. Make risky enhancements degradable without reverting business-layer code.
3. Keep the safe defaults aligned with the completed rollout.
4. Commit the phase.
