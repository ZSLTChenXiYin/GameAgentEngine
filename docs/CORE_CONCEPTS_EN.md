# Core Concepts

[**中文**](./CORE_CONCEPTS.md) | **English**

GameAgentEngine v0.2.0 models game worlds using four fundamental abstractions: **Nodes**, **Components**, **Memories**, and **Relations**. Together they form a directed graph representing each entity, its attributes, history, and connections to other entities. The engine also provides an inference pipeline, a memory propagation system, and sub-task DAG orchestration.

---

## 1. Node

A Node is the unified entity abstraction. Everything in the game world is a Node.

### Node Types

| Type | Description |
|---|---|
| `world` | Root container for a game world |
| `faction` | An organized group or faction |
| `location` | A place or area |
| `npc` | A non-player character (Agent) |
| `item` | An object or item |
| `quest_line` | A quest or storyline |
| `event` | A scheduled or ongoing event |

### Node Structure

```json
{
  "id": "uuid-string",
  "world_id": "world-uuid",
  "name": "Speaker Elrin",
  "node_type": "npc",
  "parent_id": "faction-uuid",
  "created_at": "2026-07-03T...",
  "updated_at": "2026-07-03T..."
}
```

### Node Hierarchy

Nodes form a tree structure via `parent_id`:

- A `world` node has no parent
- A world node is a root node whose `world_id` equals its own `id`
- Non-world nodes must belong to a world via `world_id`
- Non-leaf nodes (nodes that have children) cannot be deleted

```
World: Gray Harbor Border
├── Faction: Gray Harbor Council
│   ├── NPC: Speaker Elrin (belongs_to Gray Harbor Council)
│   └── NPC: Steward Brahm (belongs_to Gray Harbor Council)
├── Faction: Iron Tide Merchant Guild
│   └── NPC: Representative Mira (belongs_to Iron Tide Merchant Guild)
├── Location: Council Hall
│   └── Location: Round Table Chamber
├── Location: Mist Bay Mine
│   └── Location: Mist Bay Procurement Station
└── Location: North Gate Fortress
    └── Location: North Gate Garrison
```

---

## 2. Component

Components are structured data attached to Nodes. They contain all descriptive information about an entity.

### Built-in Component Types

| Type | Purpose | Example Content |
|---|---|---|
| `profile` | Identity & role | `{"name":"Speaker Elrin","personality":"calm, restrained"}` |
| `lore` | World or faction background | Backstory summary |
| `rule` | Gameplay rules | Turn-based rule definitions |
| `timeline` | Current turn context | Round phase information |
| `action_policy` | Action preferences | AI behavior biases |
| `relations` | Relation summaries | Cached relationship overview |
| `prompt_profile` | LLM behavior constraints | Character tone and style |
| `autonomous` | Autonomous behavior config | Triggers, capability allowlist |
| `memory` | Legacy memory (deprecated) | Use the Memory entity instead |

### Custom Components

You can create custom component types matching the pattern `^[a-z][a-z0-9_:-]{1,63}$`:

Examples from the Demo world:

- `resource_state` — `{"food":62,"order":58,"defense":49,"morale":55,"treasury":46}`
- `district_state` — `{"stability":48,"pressure":68,"output":72}`
- `demo_state` — Tracks Demo UI game state

---

## 3. Memory

Memories store information that AI Agents (NPCs) can read. They are the engine's mechanism for persistent context.

### Memory Levels

| Level | Visibility | Description |
|---|---|---|
| `short_term` | This NPC only | Ephemeral; typically not retained after the current inference round |
| `long_term` | This NPC only | Persistent; the NPC's personal history |
| `shared` | Faction/Organization level | Visible to nodes in the same faction |
| `world` | All entities in the world | World-level public knowledge |

### Memory Propagation

The engine supports four memory propagation modes to help memories flow between node levels:

| Mode | Description |
|---|---|
| `upward` | Propagate up the parent chain (default); depth limited by `propagation_max_depth` |
| `tag_broadcast` | Spread to nodes matching given tags |
| `targeted` | Direct propagation to a specified list of nodes |
| `manual` | No automatic propagation; user triggers manually |

Propagation can be configured as a **state machine** (`enable_propagation_machine`), which automatically executes propagation actions according to preset rule chains, including content transformation (prefix addition, level promotion, tag appending).

Propagation rules are configured in the database via WorldSettings as user-level dynamic configuration.

### Memory Structure

```json
{
  "id": "uuid-string",
  "node_id": "npc-uuid",
  "content": "She remembers that the last border riot was caused by simultaneous pressure on taxes and military supplies.",
  "level": "long_term",
  "tags": "npc,history,war",
  "created_at": "2026-07-03T..."
}
```

---

## 4. Relation

Relations are directed edges between Nodes, forming a graph structure.

### Built-in Relation Types

| Type | Semantics | Example |
|---|---|---|
| `belongs_to` | Membership / belonging | An NPC belongs to a faction |
| `ally` | Alliance or cooperation | Alliance between factions |
| `enemy` | Hostile relationship | Hostility between factions |
| `subordinate` | Hierarchical subordination | A branch office reports to headquarters |
| `kinship` | Family or blood relation | Kinship between NPCs |
| `located_at` | Physical location | An NPC is located at a place |

### Relation Structure

```json
{
  "id": "uuid-string",
  "world_id": "world-uuid",
  "source_id": "source-node-uuid",
  "target_id": "target-node-uuid",
  "relation_type": "belongs_to",
  "weight": 92,
  "properties": "{\"role\":\"chair\"}",
  "created_at": "2026-07-03T..."
}
```

### Weight

`weight` is an integer representing the strength or sentiment of a relation. Positive values represent positive relations, negative values represent negative relations.

---

## 5. Task Types

The engine supports five inference task types:

| Task Type | Purpose |
|---|---|
| `npc_dialogue` | NPC dialogue — respond in character with full context |
| `world_tick` | Timeline advancement — generate world state changes and future outline |
| `world_event_impact` | Event impact assessment — evaluate how an event affects the world |
| `autonomous_act` | Node autonomous behavior — decide whether to act within capability allowlist |
| `custom` | User-defined — uses the raw system prompt from context |

---

## 6. PipelineMode

Each world can independently configure its inference pipeline mode, stored in the database WorldSettings and independent of ExecutionMode:

| Mode | Value | Behavior |
|---|---|---|
| Vertical | `vertical` | Single LLM call, no task node tree, no polling |
| Polling | `polling` | Multi-round LLM polling, supports request_data queries |
| Full | `full` (default) | Full features: multi-round polling + DAG sub-task orchestration |

---

## 7. Sub-task DAG

In `full` mode, the LLM can declare a `sub_tasks` array in its JSON response, and the engine orchestrates execution automatically:

- **Dependency resolution**: the `depends_on` field controls execution order
- **Retry & timeout**: each sub-task can be configured with retry count and timeout (via WorldSettings)
- **Merge modes**: `append` (concatenate), `override` (replace), `summarize` (LLM summary)

---

## 8. Actions

Actions are operations proposed by the LLM in its response. They can be executed synchronously (within the pipeline) or asynchronously (returning a callback to the game client).

### Built-in Actions

| Action | Mode | Description |
|---|---|---|
| `update_mood` | Sync | Update NPC mood state |
| `add_memory` | Sync | Write a memory directly to a node |
| `send_dialogue` | Sync | Record dialogue content |
| `adjust_relation` | Async | Request relation weight adjustment |
| `spawn_item` | Async | Request item creation |

### Action Policy

The database WorldPolicy configures `blocked_actions` and `safe_actions` to control the scope of actions the LLM can execute.

---

## 9. World Timeline

The World Timeline tracks game state changes between Ticks.

### Tick Lifecycle

1. Client requests a Tick advance
2. Pipeline executes a `world_tick` task according to PipelineMode
3. Policy engine evaluates the plan
4. Memories are written and propagated
5. `world_tick_sync` autonomous nodes are triggered
6. A TimelineModel record is created

---

## 10. World Fork and Save

The Engine now splits world-copy behavior into three formal semantics: working-copy fork (`ForkWorld`), save snapshot (`CreateWorldSnapshot`), and restore from snapshot (`RestoreWorld`). Each operation copies the full world data set, including nodes, components, memories, relations, world settings, and world policies, and is triggered as a server-side atomic API operation.

### API Endpoint

```
POST /api/v1/worlds/{world_id}/fork
POST /api/v1/worlds/{world_id}/snapshots
POST /api/v1/worlds/{world_id}/restore
```

### Optional Parameters

| Parameter | Type | Description |
|---|---|---|
| `lock_world` | bool | Lock the source world during cloning to prevent concurrent writes (optional, defaults to unlocked) |

### Locking Mechanism

When `lock_world` is `true`, the engine uses a world-granularity mutex to protect the source world or source snapshot world. The lock only affects the fork / snapshot / restore operation itself and does not prevent normal reads or writes to nodes, components, memories, or relations. The lock is held only for the duration of the copy or restore and is released immediately upon completion.

### Behavior

- The clone creates a complete copy of the source world, including all nodes, components, memories, relations, WorldSettings, and WorldPolicy
- The new world is named with the `name` parameter; if omitted, it is automatically named "original_name (copy)"
- Cloning is a synchronous operation and may take some time for large worlds

### Common Use Cases

- **Save/Load**: Create a snapshot of the game world as an archive
- **World Branching**: Fork an existing world and evolve different branches for testing
- **Parallel Authoring**: Multiple developers can fork or snapshot the same source world for isolated work

### State Assetization Boundary

The current Engine keeps three state-bearing assets deliberately separate:

- `ForkWorld`: a runnable working copy for editing, simulation, and branch evolution
- `CreateWorldSnapshot`: a save-oriented archive intended to preserve compatibility and restore safety
- `RestoreWorld`: a runnable world recreated from a validated save snapshot

Future partial-state archives should not overload these APIs. If the Engine later adds agent-runtime-only saves, scope-level saves, or structure-only exports, they should be introduced as distinct asset types with explicit compatibility rules.
- **A/B Testing**: Clone the same world and advance it under different configurations, comparing outcome differences
- **Parallel Development**: Multiple developers work independently from the same world snapshot
