# GameAgentEngine C SDK

This SDK is the C baseline client for GameAgentEngine.

## Current Scope

- baseline HTTP access layer
- health / version / invoke / runtime-task / callback entrypoints
- minimal integration examples

## Status

This directory currently provides a baseline scaffold, not full parity with the Go SDK yet.

## Notes

- intended as the lowest-level native baseline
- currently limited to path and request-shape helpers before full transport abstraction

## Included Examples

- `examples/health.c`
- `examples/invoke_dialogue.c`

