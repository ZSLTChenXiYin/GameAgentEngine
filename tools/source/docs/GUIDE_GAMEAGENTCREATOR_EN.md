# GameAgentCreator Guide

[**中文**](./GUIDE_GAMEAGENTCREATOR.md) | **English**

GameAgentCreator is the browser-based visual editor bundled with GameAgentEngine.

---

## Open Creator

Open this file in a browser:

`tools/source/web/GameAgentCreator/index.html`

You can also use the CLI `inspect` flow when available in your environment.

---

## Main Layout

- top bar: world selection, language switch, theme switch, config entry
- left tree: hierarchical node outline
- center area: world page, snapshots, plans, settings, policy, continuity, state, timelines, logs, traces
- right side: node detail summary and attached data

### Page Overview

- `Worlds`: manage the current world, node tree, node detail, and common runtime entry points
- `Snapshots`: inspect snapshot lists for runnable worlds or source metadata and restore candidates for snapshot worlds
- `Plans`: review and approve pending world change plans
- `Policy`: edit world-level policy configuration
- `Settings`: edit world-level runtime settings
- `Continuity`: inspect the latest `world_tick` continuity bundle, diff, and request-scoped artifacts
- `State`: inspect and edit continuity-related state components including `world_state`, `story_state`, `story_history`, and `tick_policy`
- `Timelines`: inspect recent tick timeline archives, structured payloads, and result summaries
- `Logs`: inspect inference execution logs by request and event type
- `Traces`: inspect Debug-mode prompts, parse results, and pipeline details

---

## Supported Editing Flows

### World operations

- create a world
- rename the selected world from the world page
- create a working copy via `fork`
- save a snapshot
- validate, restore, and delete snapshots
- inspect and review pending world change plans
- edit world settings
- edit world policy

### Node operations

- create node
- edit node
- delete leaf node
- copy node
- create outgoing relations from node actions
- drag a node onto another node to reparent it
- drag a node to the root drop zone to clear its parent

Node copy currently duplicates:

- the selected node
- the subtree when subtree copy is enabled
- attached components
- attached memories
- relations that remain fully inside the copied subtree

The outgoing-relation action opens the same relation editor used for relation creation, so the current node can create any supported relation to another node instead of only serving the old external-parent flow.

### Component editing and validation

- Creator surfaces whether a component is strong, weak, or free-text through shared component metadata
- editing `autonomous` applies structured validation to required fields
- editing `world_state`, `story_state`, `story_history`, and `tick_policy` now applies field-aware structured validation
- editing `profile` requires a valid JSON object
- current built-in text-oriented component types can still be edited as plain text

### Runtime operations

- advance tick
- run autonomous behavior
- evaluate event impact
- scope advance
- timeline replan

World time inspection currently flows through:

- `world_time_state` in continuity state views
- timeline payloads that now include `advanced_ticks`, `previous_world_time_state`, and `world_time_state`

### Memory propagation

- explicitly trigger propagation from the memory list in node detail
- choose among `upward`, `tag_broadcast`, `targeted`, and `manual` modes
- fill tags, target node IDs, max depth, and `publish_up`

### Observability

- continuity aggregation page for the latest `world_tick` bundle
- continuity diff card for current vs previous tick facts and summaries
- `Logs` page for inspecting inference logs by event type and `request_id`
- `Traces` page for Debug-mode prompts, parse results, and pipeline details
- configured / effective pipeline mode visibility
- round usage visibility

## Continuity Workflow

- open `Continuity` to load the latest world-oriented continuity bundle
- use the request filter to focus logs and traces for a single `request_id`
- compare `Latest Tick Summary`, `Previous Tick Summary`, and fact additions/removals in `Continuity Diff`
- move to `State` when you want to directly edit `world_state`, `story_state`, `story_history`, or `tick_policy`
- keep `state_snapshot` as a read-only engine checkpoint unless you are intentionally rebuilding generated state

### Continuity State and Timelines

- `State` is used to inspect `world_state`, `story_state`, `story_history`, `tick_policy`, and `state_snapshot`
- `Timelines` is used to inspect recent tick archives and structured payloads
- `State` can edit continuity state components directly, except for `state_snapshot`
- `state_snapshot` remains read-only and works best as an engine-generated checkpoint view

When you need to investigate story-context drift, this reading order works well:

1. `Timelines` for the latest tick `reply` and `future_outline`
2. `State` for `world_state.canonical_facts` and `story_history.entries`
3. `Logs` for request / response / detail events
4. `Traces` for Debug-mode prompts and parse results

---

## Snapshot Page Notes

The Snapshots page is used for two related views:

- if a normal runnable world is selected, the page shows snapshots created from that world
- if a save snapshot world is selected, the page shows snapshot metadata for the current snapshot and lists all save snapshots from its source world

---

## Current Constraints

- Creator talks to the engine over HTTP and depends on a running server
- world rename and node copy require the newer API routes now exposed by the engine
- Creator's component validation hints come from the generated `js/component-meta.js` bundle artifact
- the packaged `tools/source/web/GameAgentCreator` copy is the one intended for distribution and direct browser use

## Current Semantic Guidance

### Tree hierarchy and explicit relations

- The World Outline only shows `Primary Parent`, which maps to `node.parent_id`.
- Dragging nodes, moving them to root, and `Add New Parent` only rewrite `parent_id`.
- `located_at` models the current environment position. Use it for NPCs, props, or groups that move over time instead of rewriting the stable hierarchy every time they relocate.
- `belongs_to` models stable affiliation or ownership, while `subordinate` models command/reporting chains. Neither relation means current location.
- `external_parent` is auxiliary DAG scope only. It is excluded from default context assembly and default propagation, so it should not replace the main hierarchy, current location, or primary organization modeling.
- `ally`, `enemy`, and `kinship` are social graph edges and stay out of the default identity/environment expansion path.
- The node detail view now surfaces `Relation Validation` so modeling drift such as multiple `located_at` edges, auxiliary `external_parent` usage, or NPCs missing a `located_at` edge can be spotted directly in the editor.
- The node detail view also surfaces `Graph Context Preview`, which summarizes the primary identity chain, environment chain, organization chain, and social-link summary that the engine will treat as the node's graph context.

### Memory propagation modes

- `upward` walks the stable `Primary Parent` chain and remains the default mode.
- `environment_scope` walks the environment chain rooted by `located_at`. `Publish Up` extends beyond that scoped graph only when explicitly enabled.
- `organization_scope` walks organization/control edges such as `belongs_to` and `subordinate`. `Publish Up` extends beyond that scoped graph only when explicitly enabled.
- `tag_broadcast` and `targeted` bypass the default structural walk and instead rely on tags or explicit target node IDs.
- `manual` records an operator-directed propagation request without relying on default graph traversal.
