# Configuration

[**中文**](./CONFIGURATION.md) | **English**

Packaged-build configuration focus:

- static config file: local `gameagentengine.conf.yaml`
- dynamic world config: `world_settings`
- time rules: `world_time_settings`
- time result: `world_time_state`

---

## Current Packaged Template Highlights

```yaml
database:
  driver: "sqlite"
  dsn: "gameagentengine.db"
  migrations_enabled: true
  write_retry_enabled: true
  log_batch_enabled: true

engine:
  execution_mode: "debug"
  world_lock_enabled: true
  autonomous_scheduler_enabled: false
```

This template is optimized for local demos and packaged onboarding. If fields are omitted, the engine still falls back to code-level defaults.

---

## Additional Notes

- `database.driver` supports `sqlite`, `mysql`, and `postgres`
- `GET /api/v1/pipeline/stats` is useful for lock contention, batching, and retry diagnosis
- if `world_time_settings` is missing, world-time-dependent flows are intentionally blocked
