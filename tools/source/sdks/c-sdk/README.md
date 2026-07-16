# GameAgentEngine C SDK

This SDK is the C baseline for native-side GameAgentEngine integration.

## Current Status

This is now a minimal practical request-builder layer rather than a pure constant-path stub.

It does not include an HTTP transport, but it now provides the core path and payload builders needed for invoke, pending-task inspection, claim/start, and callback completion loops.

## Current Scope

- health / version / invoke path helpers
- pending task path helper
- claim / start / callback payload helpers
- minimal loop example for pull-task wiring

## Included Examples

- `examples/health.c`
- `examples/invoke_dialogue.c`
- `examples/task_pull_once.c`
