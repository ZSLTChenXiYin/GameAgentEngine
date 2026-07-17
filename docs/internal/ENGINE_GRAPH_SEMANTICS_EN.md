# Engine Graph Semantics

## 1. Core split: parent vs explicit relations

Engine now treats `parent` and explicit relation edges as two different semantic layers.

- `parent`
  - The only primary hierarchy in the engine.
  - Represents stable identity, ownership, or structural containment.
  - Drives default ancestor loading.
  - Drives default upward propagation.
  - Must not be repurposed to simulate current location, temporary command structure, or auxiliary scope links.

- Explicit relations
  - Supplement `parent` with dynamic or orthogonal semantics.
  - Must be interpreted strictly as `source -> target`.
  - Must not silently replace `parent` responsibilities.
  - Must be consumed by task-specific graph assembly or explicit propagation modes.

This split exists to avoid one relation carrying incompatible meanings. In practice, stable identity, dynamic environment, organization/control, and social semantics evolve at different speeds and must not share the same structural path by accident.

## 2. Relation type semantics

### `belongs_to`
- Meaning: stable affiliation, ownership, inventory-like attachment, roster, or institutional membership.
- Does not mean current location.
- Does not replace `parent`.
- Does not create the default ancestor chain.
- May be used by organization/control context supplements and organization-scope propagation.

### `ally`
- Meaning: friendly, cooperative, or alliance relationship.
- Social edge only.
- Not part of default hierarchy loading.
- Not part of default prompt expansion.
- If bidirectional alliance is needed, callers must write both directions or explicitly interpret symmetry.

### `enemy`
- Meaning: hostile, oppositional, or conflict relationship.
- Social edge only.
- Not part of default hierarchy loading.
- Not part of default prompt expansion.
- If bidirectional hostility is needed, callers must write both directions or explicitly interpret symmetry.

### `subordinate`
- Meaning: command, reporting, or control chain.
- Not equivalent to generic ownership.
- Does not mean current location.
- Does not replace `parent`.
- May be used by organization/control context supplements and organization-scope propagation.

### `kinship`
- Meaning: family, bloodline, or marriage tie.
- Social/background edge only.
- Not part of default hierarchy loading.
- Fine-grained subtype information should live in relation properties, not by exploding enum count.

### `located_at`
- Meaning: current environment position of `source` inside the place represented by `target`.
- Dynamic environment edge only.
- Does not mean stable ownership.
- Does not replace `parent`.
- Default environment context must be assembled through this edge.
- Environment-scope propagation must use this edge explicitly.

### `external_parent`
- Meaning: auxiliary scope attachment outside the primary `parent` chain.
- Reserved for explicit, non-default extra scope modeling.
- Must not be used to mean current location.
- Must not be used to mean ordinary organization ownership.
- Must not be used to mean social relation.
- Currently excluded from default prompt expansion.
- Currently excluded from default propagation.
- Still validated structurally for cycle safety.

## 3. Task-specific graph assembly

The engine must not build one global graph prompt for every task. Each task uses a different graph slice.

### `npc_dialogue`
- Base identity: current node plus `parent` ancestor chain.
- Base environment: `located_at` target plus environment ancestor chain.
- Optional related-node supplement: only controlled relation types such as `located_at`, `belongs_to`, `subordinate`.
- Social edges are excluded from default expansion.

### `autonomous_act`
- Base identity: current node plus `parent` ancestor chain.
- Base environment: `located_at` target plus environment ancestor chain.
- Optional control supplement: `belongs_to`, `subordinate` as bounded context additions.
- Social edges are excluded from default expansion unless a future task-specific selector explicitly asks for them.

### `world_event_impact`
- If scope node is a location/world node, use it directly as environment scope.
- If scope node is an NPC or other entity, resolve environment through `located_at`.
- Focus on local impact graph, not the whole world graph.

### `world_tick`
- Must use summary-style graph assembly, not raw graph dumping.
- Current implementation adds a high-value relation summary block.
- Summary focuses on:
  - child distribution of scope nodes
  - high-value child relation samples
  - only `located_at`, `belongs_to`, `subordinate`
- Social edges such as `ally`, `enemy`, `kinship` stay out of the default tick summary.

### `custom`
- Reuses the same base context discipline.
- Any broader graph expansion should come from explicit `request_data` or future custom selectors, not from silent default expansion.

## 4. Pipeline mode semantics

Pipeline mode is not only about number of rounds. It also defines graph assembly intensity.

### `vertical`
- Smallest graph slice.
- Single LLM call.
- Use only the minimum identity/environment context needed for the current task.
- Prefer `request_data` if more graph data is required.

### `polling`
- Medium graph slice.
- Multi-round.
- Allows bounded additional graph fetches through `request_data`.
- Still must not degrade into all-graph prompt stuffing.

### `full`
- Structured multi-round graph assembly.
- May use task tree and sub-task orchestration.
- Still requires explicit boundaries on graph expansion.
- Current implementation keeps base task context merged with task-tree context rather than replacing it.

## 5. Default context loading rules

### Identity chain
- Source: primary `parent` chain only.
- Purpose: stable identity, structural ownership, default ancestor memory/component loading.

### Environment chain
- Source: `located_at` target, then that node's `parent` ancestors.
- Purpose: current place and surrounding environment memory/component loading.

### Related-node supplement
- Controlled by `IncludeRelatedNodes`.
- Not a license to load all neighbor nodes.
- Current default allowlist is task-bounded and excludes social relations.
- `external_parent` is excluded by default.

## 6. Propagation model

Propagation also follows the same semantic split and must not overload one path with multiple meanings.

### Default propagation
- `nil` propagation rule means `upward`.
- `upward` walks only the primary `parent` chain.
- It does not include `located_at`, `belongs_to`, `subordinate`, or `external_parent`.

### Explicit propagation modes

#### `upward`
- Stable hierarchy propagation.
- Uses only `parent`.

#### `environment_scope`
- Dynamic environment propagation.
- Uses `located_at` target and then that target's `parent` ancestors.
- Does not inspect organization/control edges.

#### `organization_scope`
- Organization/control propagation.
- Uses `belongs_to` and `subordinate` targets and then each target's primary `parent` chain.
- Does not inspect `located_at`.

#### `tag_broadcast`
- Explicit broadcast by tag search.
- Not a structural hierarchy mode.

#### `targeted`
- Explicit point-to-point propagation.
- Not a structural hierarchy mode.

#### `manual`
- Disabled automatic propagation.
- External caller triggers propagation manually.

### `external_parent` in propagation
- Not part of default propagation.
- If a future product requirement needs auxiliary-scope propagation, it should be introduced as a new explicit mode, not by mutating `upward`.

## 7. Parsing and execution contract

- `memory_updates[].propagation` must be parsed from model output and preserved through engine execution.
- If a propagation mode is not parsed, the mode does not exist in practice even if the engine has an implementation.
- Current implementation now parses and executes explicit propagation rules from LLM output.

## 8. Why this design exists

This design keeps future development from drifting into ambiguous graph semantics.

Without these boundaries, the same node relationship can accidentally mean:
- who the entity is
- where the entity is
- who commands the entity
- which extra scope should observe the entity
- who the entity socially cares about

Once those meanings collapse into one path, prompt assembly, memory propagation, and world simulation become unstable and hard to reason about. The current engine policy is therefore:

- one stable primary hierarchy: `parent`
- one explicit environment edge: `located_at`
- one explicit organization/control family: `belongs_to`, `subordinate`
- one explicit social family: `ally`, `enemy`, `kinship`
- one reserved auxiliary scope edge: `external_parent`

Any future extension should preserve that separation unless the whole engine contract is intentionally redesigned.
