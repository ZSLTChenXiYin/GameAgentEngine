# GameAgentEngine Java SDK

This SDK is the Java baseline for GameAgentEngine integration.

## Current Status

This is now a minimal practical HTTP client rather than a read-only placeholder.

It still does not provide full typed model coverage, but it now supports the main outer loop: health, invoke, pending tasks, claim/start, and callback.

## Current Scope

- health and version
- invoke
- pending runtime task list
- claim / start runtime task
- callback completion

## Included Examples

- `examples/HealthExample.java`
- `examples/InvokeDialogueExample.java`
- `examples/TaskPullOnceExample.java`
