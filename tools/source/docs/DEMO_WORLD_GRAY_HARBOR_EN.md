# Demo World: Gray Harbor Border

[**中文**](./DEMO_WORLD_GRAY_HARBOR.md) | **English**

"Gray Harbor Border" is a demo border-governance simulation world used to showcase the core features of GameAgentEngine v0.2.0.

---

## World Overview

Gray Harbor Border is a resource-scarce frontier town with multiple factions. The player serves as the council governor, navigating between various powers, responding to emergent events, and maintaining border stability.

### Node Structure

```
Gray Harbor Border (world)
├── Council Hall (location)
│   └── Round Table Chamber (location)
│       └── Gray Harbor Council (faction) — governing faction
│           ├── Speaker Elrin (npc) — calm, restrained, skilled at weighing options
│           └── Steward Brahm (npc) — focused on resource realities, dislikes empty talk
├── Mist Bay Mine (location)
│   └── Mist Bay Procurement Station (location)
│       └── Iron Tide Merchant Guild (faction)
├── North Gate Fortress (location)
│   └── North Gate Garrison (location)
│       └── Northern Garrison (faction)
│           └── Commander Cyllo (npc) — pragmatic, combat-focused, believes in strength
├── Riverside Market (location)
│   └── Riverside Warehouse (location)
│       └── Iron Tide Merchant Guild (faction)
│           └── Representative Mira (npc) — shrewd, diplomatic, profit-driven
```

### Resource States

The world tracks regional resources through custom components:

- `resource_state` (world-level): food=62, order=58, defense=49, morale=55, treasury=46
- `district_state` (region-level): stability, pressure, output

---

## Import Method

```bash
# Import using DevCli
GameAgentDevCli import demo-world.yaml --reset
```

The `--reset` flag clears the database before importing. If you omit `--reset`, the Demo world is appended to the existing database.