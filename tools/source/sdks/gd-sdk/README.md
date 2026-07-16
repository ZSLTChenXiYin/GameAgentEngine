# GameAgentEngine GDScript SDK

This SDK is the GDScript baseline client for GameAgentEngine.

## Current Scope

- baseline HTTP client wrapper
- health / version / invoke / runtime-task / callback entrypoints
- minimal examples for external integration flows
- aligned terminology with the Go SDK baseline

## Status

This directory currently provides a baseline scaffold, not full parity with the Go SDK yet.

## Recommended Use

Use this SDK when you need to:

- connect a GDScript runtime to GameAgentEngine
- trigger `invoke`
- consume pull tasks
- callback task results
- inspect state / timelines / logs

## Notes

- intended as the Godot-friendly baseline client
- focuses on lightweight HTTPRequest-based integration patterns

## Included Examples

- `examples/health.gd`
- `examples/invoke_dialogue.gd`
- `examples/task_pull_once.gd`

