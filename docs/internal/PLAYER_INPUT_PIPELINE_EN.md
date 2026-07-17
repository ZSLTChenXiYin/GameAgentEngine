# Player Natural-Language Input Pipeline

This document defines the execution boundary for player natural-language input inside GameAgentEngine.

## 1. Goal

The goal of player natural-language input is not to directly mutate world truth, but to use the player mirror node to produce:

- intent understanding
- missing-fact detection
- authoritative data query requests
- structured action proposals

Whether final state actually lands must still be decided by the authoritative game-side system.

## 2. Design Principles

### 2.1 The Player Node Proposes, But Does Not Decide

The player mirror node may reason about player input like a normal Agent, but it may only output:

- `intent`
- `preconditions`
- `missing_facts`
- `risk_level`

It must not treat any unvalidated action as already successful.

### 2.2 The Game Side Owns Authoritative Validation and Execution

The game side / worker is responsible for:

- scene adjacency validation
- item possession validation
- validation of money / HP / inventory / quest state
- truth-state mutation after successful execution
- bridging execution results back into interaction invoke

### 2.3 Reuse Invoke for the Interpretation Path

Player-input interpretation should not create a separate execution core.

Recommended approach:

- expose a friendly external entrypoint: `/api/v1/player/input/interpret`
- still construct an `InvokeRequest` internally
- `task_type=custom`
- `node_id=player mirror node`

### 2.4 Separate Dialogue Input From Action Input

To avoid mixing “speaking” and “acting” into one channel:

- normal dialogue still goes through `npc_dialogue` / interaction
- action-oriented natural language goes through player input interpret
- in play mode, `/act` or `/do` should explicitly enter the action-interpretation path

## 3. Overall Execution Chain

```text
player natural-language input
    -> player input interpret
    -> Engine produces PlayerIntent
    -> game-side validator
    -> game-side executor
    -> interaction bridge
    -> NPC / group chat / scene response
```

## 4. Responsibility Split

### 4.1 Engine

Responsible for:

- semantic decomposition of player input
- step-splitting for composite actions
- missing-fact detection
- request-data query requests
- structured intent-proposal output

Not responsible for:

- final truth-state landing
- authoritative rejection of illegal actions
- transaction rollback

### 4.2 Game Side / Worker

Responsible for:

- authoritative state reading
- validator
- executor
- interaction bridge

### 4.3 Play / REPL

Responsible for:

- developer input entrypoint
- debug output
- authority-state-file-driven playtesting

## 5. Input Categories

### 5.1 Dialogue-Type Input

Examples:

- “老板，今晚谁最后一个从码头回来？”
- “你刚才看见谁进门了？”

Characteristics:

- primarily speech
- usually does not directly trigger high-risk truth mutation

Recommended path:

- go through interaction invoke directly

### 5.2 Action-Type Input

Examples:

- “我把沾血的短刀拍在柜台上，问老板今晚有没有见过这把刀的主人”
- “我把银戒指塞给老板，试探她会不会松口”
- “我转身往后门走，同时示意守卫别出声”

Characteristics:

- contains action steps
- requires authoritative state validation
- may be composite actions

Recommended path:

- run player input interpret first

## 6. Composite Action Constraints

Composite actions must be split into structured `steps`.

For example:

`我把沾血的短刀拍在柜台上，问老板今晚有没有见过这把刀的主人`

This should not be treated as a narrative fact that has already succeeded. It should instead be split into:

1. `show_item`
2. `speech`

If the first step fails authoritative validation, then the second step must not proceed on the false premise that “the knife has already been shown.”

## 7. First-Version Support Scope

The first version should support at least these intent / step types:

- `speech`
- `show_item`
- `gift`
- `trade_request`
- `threaten`
- `move`
- `inspect`
- `use_item`
- `composite`

## 8. First-Version Unsupported Scope

The first version should not directly support:

- unbounded free-form world modification
- direct landing of unmodeled actions
- parallel chained reasoning across multiple NPCs
- treating player input as an already-established world fact

## 9. Relationship to Interaction

Player input interpret is not a replacement for interaction. It is a pre-layer in front of interaction.

Typical bridge mappings:

- `speech` -> `direct_dialogue` or `group_chat`
- `show_item` -> `event=show_item`
- `gift` -> `gift_response`
- `trade_request` -> `trade_dialogue`
- `threaten` -> `event=threaten`

## 10. Core Conclusion

The player node may carry “behavior reasoning for player natural language,” but it must not carry “final world-truth writeback.”

At the engineering level, the system must keep this separation:

- Engine proposes
- the game side decides
