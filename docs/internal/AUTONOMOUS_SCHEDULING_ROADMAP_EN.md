# Autonomous Scheduling Roadmap

[**中文**](./AUTONOMOUS_SCHEDULING_ROADMAP.md) | **English**

This document records future work for evolving autonomous execution from basic trigger scanning into a bounded scheduling system.

## 1. Current Baseline

Current autonomous execution supports three trigger modes:

- `manual`
- `world_tick_sync`
- `scheduled`

Current behavior is intentionally simple:

- load autonomous components for a world
- scan them sequentially
- filter by enabled flag and trigger
- for `scheduled`, check due-state by interval
- stop at a fixed run limit

This is enough for a baseline, but it is not yet a real scheduling model.

## 2. Future Goal

The future goal is not to make autonomous behavior globally always-on.

The goal is:

- run the right nodes at the right time
- avoid full-world scanning as the primary mechanism
- support bounded prioritization and batching
- keep autonomous behavior compatible with external authority callbacks and resume flow

## 3. Why Future Scheduling Is Needed

The current model still has these limits:

- no priority
- no explicit queueing state
- no event-driven wake-up path
- no fair batching policy across heterogeneous nodes
- no distinction between “important but not due” and “due but low-value”

## 4. Future Scheduling Model

### 4.1 Dispatch Layers

Future autonomous dispatch should be split conceptually into:

- trigger admission
- wake-up / due-state evaluation
- priority scoring
- batch selection
- execution
- post-run state update

### 4.2 Trigger Admission

Trigger admission should continue to support current modes:

- manual
- world tick sync
- scheduled

But future work may also add event-driven wake-up without breaking these coarse modes.

## 5. Priority and Batching

### 5.1 Why Priority Is Separate from Trigger

Trigger answers “when can this node run”.

Priority answers “which runnable nodes should run first”.

These must remain separate fields.

### 5.2 Suggested Future Priority Signals

Priority may later consider:

- explicit autonomous priority in config
- relation to the current active world scope
- recent player interaction relevance
- recent world-event relevance
- wake-up cause severity
- starvation prevention bonus for long-unrun nodes

### 5.3 Suggested Batch Rules

Future batch selection should be bounded by:

- max nodes per world per dispatch
- optional per-trigger caps
- optional per-scope caps
- cooldown windows for repeated wake-ups

## 6. Event-Driven Wake-Up

### 6.1 Meaning

Event-driven wake-up means the system does not only discover runnable nodes by scanning every autonomous component on every cycle.

Instead, internal or external events explicitly enqueue or mark nodes as newly relevant.

### 6.2 Example Wake-Up Sources

Possible future wake-up sources:

- player dialogue directed at a node
- player action affecting a node or its scene
- scene-state or room-state change
- quest-state or item-ownership change
- authority callback completion
- world-tick-generated event touching a node or scope

### 6.3 Why This Does Not Replace Game-Side Driving

Game-side driving and event-driven wake-up are not the same thing.

Game-side driving means external systems can call or resume Engine work.

Event-driven wake-up means Engine / service keeps a more precise internal model of which autonomous nodes became worth running.

These mechanisms should coexist.

## 7. Autonomous Runtime State Machine

### 7.1 Recommended Scope

Future work should first add a scheduling lifecycle state machine, not a gameplay personality state machine.

Suggested runtime lifecycle states:

- `idle`
- `queued`
- `running`
- `waiting_external`
- `cooled_down`
- `blocked`
- `failed`

This state machine describes execution lifecycle only.

It should not be conflated with gameplay behavior states such as patrol, trade, fear, or combat.

### 7.2 Why It Matters

This helps with:

- queue visibility
- retry control
- duplicate wake-up suppression
- callback resume correctness
- future observability and fairness

## 8. Relationship to World Tick

Future autonomous scheduling should integrate with, but stay separate from, world-tick scope selection.

- world tick = summary and world progression
- autonomous = node-level follow-up execution

Useful future integration points:

- world tick can produce a shortlist of wake-worthy nodes
- `world_focus` and activity scoring can influence autonomous priority
- autonomous wake-up may feed back into later scope summaries

## 9. Implementation Order

Suggested future implementation order:

1. add explicit autonomous priority and cooldown semantics
2. add runnable-node scoring and bounded batching
3. add execution lifecycle state tracking
4. add event-driven wake-up queue or wake-mark model
5. integrate wake-up causes with world-tick outputs and interaction flows
6. only later evaluate whether richer behavior-state frameworks are needed
