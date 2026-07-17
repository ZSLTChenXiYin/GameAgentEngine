# Roleplay Interaction Roadmap

[**中文**](./ROLEPLAY_INTERACTION_ROADMAP.md) | **English**

This document records future development work around direct roleplay chat, room chat, player-intent bridging, and interaction-session modeling.

## 1. Scope

This roadmap covers:

- direct single-target dialogue
- room or group-chat interaction
- player natural-language control that must still pass game-side authority validation
- session-like interaction context that keeps actor / target / scene / room semantics stable

This roadmap does not redefine Engine as a UI shell.

## 2. Current Baseline

The current direction already established in the project is:

- the player should exist as a real node in Engine world modeling
- high-frequency authoritative data remains game-side
- direct dialogue and group-chat semantics are carried through interaction context
- `play` belongs to Worker rather than Engine core
- natural-language action input must be interpreted first, then authority-validated, then bridged back into interaction response

## 3. Future Goal

The future goal is not unrestricted free-form narration as immediate truth.

The goal is:

- let player natural language drive proposal and interaction flow naturally
- preserve authority-state correctness
- give NPCs enough scene, actor, and room semantics to respond coherently
- support room-based interaction without requiring unrestricted multi-agent graph explosion

## 4. Direct Single-Chat Roadmap

Future single-chat work should keep these semantics explicit:

- actor node
- target node
- scene node
- optional room identifier
- turn index
- event type

Future refinements should improve:

- stable session continuity
- target switching rules
- scene-aware prompt assembly
- item-showing / gift / threat / trade event semantics under the same interaction model

## 5. Group-Chat Roadmap

### 5.1 Current Principle

Current group-chat direction is intentionally constrained:

- room authority remains game-side
- Engine still plays one primary NPC per turn
- other participants enter prompt context as lighter room summary or participant summary

### 5.2 Future Decision Space

Future group-chat development should stay explicit about which model is being chosen:

- keep one-primary-responder as the stable default
- allow target nomination within group context
- optionally introduce staged multi-NPC response sequences later

It should not silently drift into unconstrained many-NPC simultaneous reasoning.

### 5.3 Future Staged Expansion

If deeper group-chat is later needed, preferred order is:

1. improve room-state authority and participant visibility rules
2. improve participant summary injection
3. allow bounded staged secondary responses
4. only then reconsider whether heavier multi-NPC reasoning is justified

## 6. Interaction Session Modeling

Future work should stabilize an interaction-session concept that keeps context coherent across turns.

Desired session semantics include:

- who is speaking
- who is currently targeted
- which room or scene the turn belongs to
- whether the turn is public / private / whisper-like
- which participants are formally present
- whether a turn was triggered by speech, gift, item show, threat, or another structured event

This session model should unify:

- plain interaction invoke
- Worker play direct chat
- Worker play room chat
- player-intent bridge outputs

## 7. Player Natural-Language Control Bridge

The future bridge should preserve this contract:

1. player language is interpreted through the player node
2. Engine outputs intent proposal, missing facts, and structured steps
3. game-side authority validates and executes what is legal
4. the execution result is bridged into direct chat, room chat, or structured NPC reaction

This keeps player natural language expressive without allowing unsupported narration to become truth automatically.

## 8. Relationship to World Tick and Autonomous

Roleplay interaction should benefit from later world-tick and autonomous improvements, but should not depend on turning them into the same subsystem.

- world tick improves world awareness and selected scope freshness
- autonomous improves node initiative and post-event behavior
- roleplay interaction governs actor-target-scene-room semantics and turn-by-turn player-facing flow

These systems should cooperate but remain conceptually separate.

## 9. Implementation Order

Suggested future implementation order:

1. keep interaction-session semantics stable across direct chat and room chat
2. improve room authority ownership and participant modeling
3. improve player-intent bridge quality and post-validation interaction handoff
4. refine one-primary-responder group-chat behavior until its limits are clearly measured
5. only then decide whether staged multi-NPC reply flows are worth adding
