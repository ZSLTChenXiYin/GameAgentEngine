# GameAgentEngine C++ SDK

This SDK is the C++ baseline client for GameAgentEngine.

## Current Scope

- baseline HTTP access layer
- health / version / invoke / runtime-task / callback entrypoints
- minimal integration examples

## Status

This directory currently provides a baseline scaffold, not full parity with the Go SDK yet.

## Notes

- intended as the native-plugin and engine-bridge baseline
- favors explicit HTTP wrapper structure over framework-specific bindings

## Included Examples

- `examples/health.cpp`
- `examples/invoke_dialogue.cpp`

