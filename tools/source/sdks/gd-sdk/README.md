# GameAgentEngine GDScript SDK

This SDK is the Godot-friendly baseline client surface for GameAgentEngine.

## Current Status

This is now a practical first version for request construction and integration wiring rather than a pure path-only scaffold.

It still does not provide full Go SDK parity, but it already covers the main request shapes needed for Godot-side Engine / Worker integration.

## Current Capability Scope

- health and version request builders
- invoke and player input interpretation request builders
- runtime task list / pending / get / claim / start / heartbeat / release / requeue / stats request builders
- callback request builder
- world settings / state components / timelines / logs / debug traces request builders
- world tick advance request builder

## Notes

- intended as the Godot-friendly baseline client
- focuses on lightweight `HTTPRequest` / `HTTPClient` integration patterns
- returns request dictionaries so Godot projects can route them through their own async HTTP layer

## Included Examples

- `examples/health.gd`
- `examples/invoke_dialogue.gd`
- `examples/task_pull_once.gd`

## Not Yet Included

Not yet at Go SDK parity:

- built-in async HTTP execution wrapper around `HTTPRequest`
- full worlds / nodes / components / memories / relations CRUD surface
- higher-level Godot gameplay helpers for `play`-style flows
