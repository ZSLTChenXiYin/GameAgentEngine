# GameAgentEngine TypeScript SDK

This SDK is the TypeScript baseline client for GameAgentEngine.

## Current Scope

- baseline HTTP client wrapper
- health / version / invoke / runtime-task / callback entrypoints
- minimal examples for external integration flows
- aligned terminology with the Go SDK baseline

## Status

This directory currently provides a baseline scaffold, not full parity with the Go SDK yet.

## Recommended Use

Use this SDK when you need to:

- connect a TypeScript runtime to GameAgentEngine
- trigger `invoke`
- consume pull tasks
- callback task results
- inspect state / timelines / logs

## Notes

- best fit for Node.js services, tools, and strongly typed frontend-side helpers
- intended to be the primary non-Go SDK baseline

## Included Examples

- `examples/health.ts`
- `examples/invoke_dialogue.ts`
- `examples/task_pull_once.ts`

