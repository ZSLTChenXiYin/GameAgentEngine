# GameAgentEngine Java SDK

This SDK is the Java baseline client for GameAgentEngine.

## Current Scope

- baseline HTTP access layer
- health / version / invoke / runtime-task / callback entrypoints
- minimal integration examples

## Status

This directory currently provides a baseline scaffold, not full parity with the Go SDK yet.

## Notes

- intended as the JVM-side baseline for service bridges and tooling
- favors simple HttpClient + String payload patterns first

## Included Examples

- `examples/HealthExample.java`
- `examples/InvokeDialogueExample.java`

