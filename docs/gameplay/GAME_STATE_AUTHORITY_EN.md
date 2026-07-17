# Game State Authority Boundary

This page is the current primary specification for authority boundaries between Engine and the game side.

## 1. Core Principles

- The game side is the only authoritative source for high-frequency truth data.
- Engine is responsible for world modeling, NPC cognition, memory, relationships, narrative, and interaction reasoning.
- Engine does not directly own high-frequency truth data; it queries the game side asynchronously when needed.
- Player, NPC, and scene information stored in Engine may exist as semantic mirrors, but must not replace game-side truth.

## 2. Authoritative Game-Side Data

The following data must be treated as authoritative only in `gameagentworker` local state:

- player HP, stamina, currency, hunger, and status effects
- player inventory, equipment, item counts, durability, and availability
- exact NPC/player location and scene occupancy
- immediate scene state, such as whether a door is open, whether the counter is occupied, and current room members
- real quest stage, completion conditions, cooldowns, and reward delivery state
- trade stock, prices, combat outcomes, drops, movement blocking, and other high-frequency rule state

## 3. Engine-Side Modeled Data

The following data is suitable for long-term residency in Engine:

- world nodes, location nodes, NPC nodes, and player nodes
- organizations, factions, ownership, and social relationships
- long-term memory, shared memory, and world memory
- scene semantic descriptions, location hierarchy, and world-time semantic state
- player/NPC identity impressions, recent behavior summaries, and narrative context

## 4. Semantic Mirror Data

The following data may be synchronized into Engine as reasoning aids, but not as truth:

- `player_001` appears injured, wealthy, suspicious, or popular
- the player recently gave a gift to the innkeeper
- the scene recently contained an argument, disturbance, threat, or failed trade
- the current room theme, group-chat atmosphere, and visible emotional summary of room members

## 5. Data Engine May Query On Demand

The following data may be requested during Engine reasoning, but should not be injected into prompts by default:

- current HP or HP band of a player/NPC
- current player money
- whether the player holds a key item
- whether an NPC/player is present in the current scene
- current room members and scene occupancy
- real current quest stage
- immediate state of a scene

## 6. Player Node Principles

- The player should exist as a formal node in Engine, typically as `player` or `player_actor`.
- The player node carries identity, relationships, memory, and narrative position.
- The player node is not the source of authoritative player state.
- Authoritative player-state queries must return to the game side.

## 7. Execution Principles

- Dialogue-oriented natural-language input may enter Engine reasoning directly.
- High-risk actions must not change state directly from natural language; they must first pass game-side rule validation.
- Structured actions should first land as authoritative truth on the game side, then Engine may generate NPC reactions and narrative feedback.
