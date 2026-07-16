# GameAgentEngine Lua SDK

This SDK is the Lua baseline for lightweight script-side GameAgentEngine integration.

## Current Status

This is now a minimal practical request-builder layer rather than a pure path-only scaffold.

It does not include a built-in HTTP transport, but it now covers the main request and payload shapes needed for invoke, runtime task consumption, and callback loops.

## Current Scope

- health / version / invoke path and request helpers
- runtime task pending / get / claim / start helpers
- callback payload helper
- minimal loop example for pull-task wiring

## Included Examples

- `examples/health.lua`
- `examples/invoke_dialogue.lua`
- `examples/task_pull_once.lua`
