# World Tick Context Roadmap

**中文** | [**English**](./WORLD_TICK_CONTEXT_ROADMAP_EN.md)

This document records future Engine work for improving world-tick and scope-tick context selection without turning world inference into full-world graph stuffing.

## 1. Problem Statement

Current `world_tick` behavior is world-or-scope centric:

- one focus node is selected first
- context is assembled around that focus node
- world-tick summary logic only expands a constrained subset of relations and direct child summaries
- autonomous node execution happens as a separate post-tick phase

This keeps the kernel bounded, but it also creates clear future gaps:

- important deep descendants are easy to miss
- high-activity nodes are not ranked explicitly
- large scopes do not yet support staged scope refinement
- the current summary path is still mostly static rather than demand-shaped

## 2. Future Goal

The future goal is not “load more nodes by default”.

The goal is:

- keep the Engine prompt budget bounded
- improve awareness of structurally important descendants
- improve awareness of tick-relevant active nodes
- support staged expansion from coarse scope summaries into selected sub-scopes
- keep world tick suitable for embedding and kernelization

## 3. `world_focus` Component

### 3.1 Purpose

`world_focus` is a future context-selection component.

It is meant to allow selected descendants under the current focus subtree to participate in world-tick or scope-tick reasoning even when they are not direct children of the focus node.

It is not:

- a truth-state component
- an autonomous behavior component
- a persistence shortcut
- a UI bookmark feature

It is a reasoning-selection hint owned by Engine context assembly.

### 3.2 Intended Semantics

When a world tick or scope tick starts from focus node `F`:

1. Engine builds the normal base scope around `F`
2. Engine scans descendants under `F`
3. descendants carrying `world_focus` become candidates for promotion into this tick
4. the promoted nodes participate as explicit observation points
5. they are summarized under bounded rules rather than recursively exploding into full context

### 3.3 First-Version Constraints

The first implementation should stay conservative:

- only descendants of the current focus node are eligible
- only selected task types should honor it, starting with `world_tick`
- default behavior should be summary-first rather than full heavy-context inclusion
- hard limits should exist for scan depth, selected node count, and expansion cost

### 3.4 Suggested Payload Shape

Suggested future payload shape:

```json
{
  "enabled": true,
  "tasks": ["world_tick"],
  "priority": 80,
  "reason": "quest_hub",
  "max_parent_distance": 3,
  "summary_only": true,
  "include_children": 0,
  "include_relations": ["belongs_to", "subordinate", "located_at"]
}
```

The final shape may change, but these semantics should remain:

- explicit enablement
- task scoping
- selection priority
- bounded descendant distance
- bounded summary behavior

## 4. Active-Node Selection Model

`world_focus` does not remove the need for activity selection.

These are different concerns:

- `world_focus` = long-term structural importance
- active-node selection = current-tick relevance

### 4.1 Selection Principle

Future node selection should use a scored candidate model rather than a single boolean flag.

The Engine or service layer should build a candidate set first, then score and trim it.

### 4.2 Candidate Sources

Candidate sources should include:

- the current focus node
- direct child scopes
- event-impacted nodes
- nodes referenced by recent world-change output
- `world_focus` descendants
- due autonomous nodes
- recently updated nodes
- recently interacted player-facing nodes

### 4.3 Activity Signals

Suggested activity score signals:

- structural proximity to the current focus
- recent component / relation / memory change
- recent authority callback or runtime-task completion
- recent player interaction
- presence in pending or newly generated world events
- autonomous due-state or wake-state
- explicit `world_focus` priority bonus

### 4.4 Expected Output

The result should not be “all active nodes join the prompt”.

Instead, selection should yield:

- top-level scope nodes for summary
- promoted high-value observation nodes
- optionally a smaller deepening shortlist for later refinement

## 5. Scope Refinement

### 5.1 Meaning

“Scope refinement” means staged expansion:

1. summarize multiple scopes coarsely
2. decide which scope changed materially
3. refine only the selected child scope
4. optionally repeat one more level if budget still allows

This is a budget-control strategy, not a full recursive traversal strategy.

### 5.2 Why It Exists

Without staged refinement, the system has only two bad extremes:

- too shallow to notice important local dynamics
- too broad and too expensive if everything is expanded at once

### 5.3 Role of LLM-Guided Loading

Future refinement may allow LLM-guided selection, but not unconstrained self-expansion.

Preferred model:

- Engine prepares a bounded refinement candidate list
- LLM may choose which sub-scope deserves refinement and why
- Engine still owns the actual data-loading rules and safety limits

LLM should have selection power, not arbitrary graph-expansion power.

### 5.4 Suggested Phases

Suggested future phases:

1. coarse scope summary
2. active-scope selection
3. selective child-scope refinement
4. optional focused descendant promotion through `world_focus`
5. final world-tick synthesis

## 6. Boundaries

This roadmap should not collapse these responsibilities together:

- world tick summary planning
- autonomous per-node execution
- authority-state truth ownership
- play / Worker presentation logic

The Engine should own context-selection semantics, but not game-side truth mutation or Worker shell presentation.

## 7. Implementation Order

Suggested future implementation order:

1. add `world_focus` component contract
2. add candidate-node selection and scoring
3. add scope-refinement phases
4. connect world-tick summary generation to the refined scope model
5. only later evaluate whether more dynamic multi-stage loading is needed
