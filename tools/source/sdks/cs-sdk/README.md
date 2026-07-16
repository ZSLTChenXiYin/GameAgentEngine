# GameAgentEngine C# SDK

This SDK is the C# baseline client for GameAgentEngine.

## Current Scope

- baseline HTTP client wrapper
- health / version / invoke / runtime-task / callback entrypoints
- minimal examples for external integration flows
- aligned terminology with the Go SDK baseline

## Status

This directory currently provides a baseline scaffold, not full parity with the Go SDK yet.

## Recommended Use

Use this SDK when you need to:

- connect a C# runtime to GameAgentEngine
- trigger `invoke`
- consume pull tasks
- callback task results
- inspect state / timelines / logs

## Notes

- intended as the Unity-friendly baseline client
- favors explicit POCO models and HttpClient-based usage

## Included Examples

- `examples/HealthExample.cs`
- `examples/InvokeDialogueExample.cs`
- `examples/TaskPullOnceExample.cs`

