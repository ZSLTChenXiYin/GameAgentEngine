# GameAgentEngine JavaScript SDK

This SDK is the JavaScript baseline client for GameAgentEngine.

## Current Scope

- baseline HTTP client wrapper
- health / version / invoke / runtime-task / callback entrypoints
- minimal examples for external integration flows
- aligned terminology with the Go SDK baseline

## Status

This directory currently provides a baseline scaffold, not full parity with the Go SDK yet.

## Recommended Use

Use this SDK when you need to:

- connect a JavaScript runtime to GameAgentEngine
- trigger `invoke`
- consume pull tasks
- callback task results
- inspect state / timelines / logs

## Notes

- best fit for lightweight scripts, Node.js tools, and bridge processes
- keeps the API surface close to the TypeScript baseline without TS-only types

## Included Examples

- `examples/health.js`
- `examples/invoke_dialogue.js`
- `examples/task_pull_once.js`

