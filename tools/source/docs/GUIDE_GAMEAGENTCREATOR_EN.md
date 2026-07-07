# GameAgentCreator Guide

[**中文**](./GUIDE_GAMEAGENTCREATOR.md) | **English**

GameAgentCreator is the browser-based visual editor bundled with GameAgentEngine.

---

## Open Creator

Open this file in a browser:

`web/GameAgentCreator/index.html`

You can also use the CLI `inspect` flow when available in your environment.

---

## Main Layout

- top bar: world selection, language switch, theme switch, config entry
- left tree: hierarchical node outline
- center area: world page, snapshots, settings, policy, logs, traces
- right side: node detail summary and attached data

---

## Supported Editing Flows

### World operations

- create a world
- rename the selected world from the world page
- create a working copy via `fork`
- save a snapshot
- validate, restore, and delete snapshots
- edit world settings
- edit world policy

### Node operations

- create node
- edit node
- delete leaf node
- copy node
- drag a node onto another node to reparent it
- drag a node to the root drop zone to clear its parent

Node copy currently duplicates:

- the selected node
- the subtree when subtree copy is enabled
- attached components
- attached memories
- relations that remain fully inside the copied subtree

### Runtime operations

- advance tick
- run autonomous behavior
- evaluate event impact
- scope advance
- timeline replan

### Observability

- inference logs
- debug traces
- configured / effective pipeline mode visibility
- round usage visibility

---

## Snapshot Page Notes

The Snapshots page is used for two related views:

- if a normal runnable world is selected, the page shows snapshots created from that world
- if a save snapshot world is selected, the page shows snapshot metadata for the current snapshot and lists all save snapshots from its source world

---

## Current Constraints

- Creator talks to the engine over HTTP and depends on a running server
- world rename and node copy require the newer API routes now exposed by the engine
- the packaged `web/GameAgentCreator` copy is the one intended for distribution and direct browser use
