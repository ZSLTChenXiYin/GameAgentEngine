# GameAgentEngine C++ SDK

This SDK is the C++ baseline for native-side GameAgentEngine integration.

## Current Status

This is now a minimal practical request-builder layer rather than a pure path-only scaffold.

It still does not provide a full HTTP transport implementation, but it now covers the main request shapes needed to integrate native code with Engine and Worker loops.

## Current Scope

- health / version / invoke request construction
- runtime task pending / get / claim / start request construction
- callback request construction
- minimal loop example for pull-task handling

## Included Examples

- `examples/health.cpp`
- `examples/invoke_dialogue.cpp`
- `examples/task_pull_once.cpp`
