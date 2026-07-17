# Player Interaction Overview

[**中文**](./PLAYER_INTERACTION.md) | **English**

This page is the main design entrypoint for player natural-language input, structured interaction, group chat, and game-side authority validation.

## 1. Core Principles

- the player should exist as a formal node in Engine
- the player node carries identity, relationships, memory, and narrative position
- the player node does not own high-frequency truth state
- high-frequency truth state must remain in authoritative game-side data

## 2. How Natural-Language Input Should Be Handled

The recommended approach is not to let natural language directly mutate authoritative game truth. Instead:

1. interpret the input as player intent
2. extract the target, scene, object, preconditions, and potential risk
3. return to the game side for any truth validation that depends on authoritative state
4. only actions that pass game-side rule validation can actually land
5. then let Engine reason about NPC reaction, group-chat atmosphere, and narrative feedback

## 3. Responsibility Boundary Between Engine and the Game Side

- Engine: world modeling, NPC understanding, narrative organization, relationship reasoning, and memory reasoning
- game side: HP, inventory, money, quest stage, scene occupancy, immediate state, and other high-frequency truth

For the authority boundary in detail, see:

- [Game State Authority Boundary](./GAME_STATE_AUTHORITY.md)

## 4. Recommended Interaction Modes

Current interaction should at least distinguish these modes:

- `direct_dialogue`: the player speaks with a single NPC
- `group_chat`: the player shares one speaking space with a group of NPCs
- `gift_response`: an NPC reaction triggered after the player gives a gift or shows an item
- other high-risk actions: first validated by the game-side rules layer, then conditionally bridged into Engine reaction reasoning

## 5. Mapping to Play Mode

In the current `GameAgentWorker play` flow, REPL commands roughly map as follows:

- `/+talk <npc>`: set the direct dialogue target; subsequent plain text is sent to that target
- `/+say <message>`: speak publicly to the current room; the current primary group responder replies first
- `/+ask <npc> <message>`: nominate a specific NPC to reply within group-chat context
- `/+act <message>`: interpret player intent first, let the game side validate and land authoritative state, then bridge to Engine feedback when needed
- `/+gift <npc> <item>`: execute the gift on the game side first, then request Engine feedback
- `/+show_item <npc> <item>`: validate item presence first, then request target reaction

Legacy aliases such as `/talk`, `/say`, and `/ask` are still accepted for compatibility, but formal documentation now standardizes on the `/+cmd + args` style.

## 6. Current Constraints

- group chat still uses a single primary responder rather than parallel multi-NPC reasoning
- high-risk natural-language actions cannot bypass the game-side rules layer
- free-form player narration is suitable as intent input, but not as direct truth mutation

## 7. Supplemental Material

The following detailed design notes have moved to `docs/internal/`:

- [Player Input Pipeline](../internal/PLAYER_INPUT_PIPELINE.md)
- [Player Intent Schema](../internal/PLAYER_INTENT_SCHEMA.md)
- [Interaction API Draft](../internal/INTERACTION_API.md)
