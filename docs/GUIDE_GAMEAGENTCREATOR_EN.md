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
- center area: world page, snapshots, plans, settings, policy, continuity, logs, traces
- right side: node detail summary and attached data

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

### Memory propagation

- explicitly trigger propagation from the memory list in node detail
- choose among `upward`, `tag_broadcast`, `targeted`, and `manual` modes
- fill tags, target node IDs, max depth, and `publish_up`

### Observability

- continuity aggregation page for the latest `world_tick` bundle
- continuity diff card for current vs previous tick facts and summaries
- inference logs
- debug traces
- configured / effective pipeline mode visibility
- round usage visibility

## Continuity Workflow

- open `Continuity` to load the latest world-oriented continuity bundle
- use the request filter to focus logs and traces for a single `request_id`
- compare `Latest Tick Summary`, `Previous Tick Summary`, and fact additions/removals in `Continuity Diff`
- move to `State` when you want to directly edit `world_state`, `story_state`, `story_history`, or `tick_policy`
- keep `state_snapshot` as a read-only engine checkpoint unless you are intentionally rebuilding generated state

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
