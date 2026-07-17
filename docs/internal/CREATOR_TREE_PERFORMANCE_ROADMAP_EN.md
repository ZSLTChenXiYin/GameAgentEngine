# Creator Large-Tree Performance Roadmap

[**中文**](./CREATOR_TREE_PERFORMANCE_ROADMAP.md) | **English**

This document locks down the performance roadmap for the left-side outline tree in GameAgentCreator under large-world conditions.

The goal is not vague “performance optimization.” The goal is explicit: when the outline reaches ten thousand nodes or more, Creator must remain browsable, searchable, expandable, and editable.

---

## 1. Background

The current Creator outline becomes visibly slow as node count grows. Based on the real implementation, the main symptoms are:

- first render time grows sharply;
- typing into the filter rebuilds the full tree repeatedly;
- expanding or collapsing a branch redraws the whole tree;
- selection, range selection, and drag interactions also rebuild the tree;
- each row has its own event bindings, so memory and initialization cost scale with node count.

This is not just one slow function. The current rendering model itself is not suitable for 10k-scale trees.

---

## 2. Confirmed Problems

### 2.1 The current outline uses full-tree rebuild rendering

Each `renderTree()` call currently:

- clears the entire container;
- rebuilds `nodeMap` and `childMap`;
- recursively recreates all visible DOM nodes;
- rebinds click, drag, and context-menu handlers per node.

That is acceptable at small scale, but it becomes the dominant bottleneck at 10k scale.

### 2.2 Filter, collapse, and selection lack local-update semantics

The current interaction model has no diff-style or partial-refresh capability. As a result:

- collapsing one branch rebuilds the entire tree;
- selecting one node rebuilds the entire tree;
- filter changes rescan and repaint the whole result set.

### 2.3 Off-screen nodes still become real DOM

There is no virtual scrolling yet. If a node is logically visible in the tree, it becomes real DOM.

That means the browser maintains thousands of rows even when the user can only see a few dozen.

### 2.4 Event listeners are bound per node

The current model attaches multiple handlers to each node. As node count grows:

- initialization cost increases;
- memory usage increases;
- rerender cost grows because handlers are rebound repeatedly.

### 2.5 Search does not yet use a large-tree strategy

Current search mostly filters the full node array directly. The problem is not correctness. The problem is that:

- there is no reusable index;
- there is no debounce/throttle path;
- there is no match-set-only refresh strategy.

---

## 3. Upgrade Principles

Future implementation should follow these principles:

1. treat visible-region rendering as the default, not full-tree DOM materialization;
2. separate data preparation from view rendering;
3. solve the major costs first: DOM volume, full-tree rerendering, and per-node event binding;
4. do not break existing semantics such as primary-parent hierarchy, alias selection, path highlighting, reparent drag-and-drop, and context menus;
5. require measurable validation at each stage rather than subjective “it feels faster.”

---

## 4. P0: Establish the Performance Baseline

### 4.1 Goal

Create a reproducible baseline before optimization starts.

### 4.2 Metrics to capture

- first tree render time;
- expand/collapse latency;
- single-node selection latency;
- filter refresh latency;
- scroll FPS / dropped-frame behavior;
- browser memory and DOM node count.

### 4.3 Baseline scales

- 1k nodes;
- 5k nodes;
- 10k nodes;
- optionally 20k if needed.

---

## 5. P0: Rebuild the Tree Data Layer

### 5.1 Goal

Turn the raw world-node array into reusable indexed tree state so UI interaction no longer forces full data reconstruction.

### 5.2 Suggested work

- maintain stable `nodeMap`;
- maintain stable `childMap`;
- maintain a stable flattened `visibleRows` result;
- treat filter, collapsed, selected, and dragging as light view-driving state rather than full-tree recomputation triggers.

---

## 6. P0: Flattened Visible-Row Model

### 6.1 Goal

Replace recursive DOM-tree rendering with a “tree semantics + row rendering” model.

### 6.2 Expected benefits

- virtual scrolling becomes straightforward;
- expand/collapse affects only one row range;
- selection and highlight can update locally;
- search-result navigation becomes easier.

---

## 7. P0: Virtual Scrolling

### 7.1 Goal

Only render rows inside the viewport plus a small buffer.

### 7.2 Requirements

- keep indentation, selection, highlight, expand arrows, and drag hit-testing correct;
- keep scrollbar length aligned with the true total row count;
- avoid flicker and jumpiness while scrolling.

### 7.3 Success condition

At 10k nodes, scrolling is no longer dominated by excessive DOM volume.

---

## 8. P1: Local Updates Instead of Full-Tree Refresh

### 8.1 Goal

Convert the most common interactions from full rerenders into local state updates.

### 8.2 Highest-priority interactions

- expand/collapse;
- single-node selection;
- range selection updates;
- ancestor-path highlighting;
- drop-target highlighting during drag.

---

## 9. P1: Event Delegation

### 9.1 Goal

Move to container-level click, context-menu, and drag hit handling instead of per-node binding.

### 9.2 Expected benefits

- lower initial construction cost;
- lower rerender cost;
- easier node reuse under virtual scrolling.

---

## 10. P1: Search Optimization and Large-Tree Degradation Strategy

### 10.1 Search optimization

- build indexed candidate paths for name/type lookup;
- debounce input;
- locate matching nodes first, then decide whether to expand ancestor paths;
- avoid “one search = one full-tree repaint” whenever possible.

### 10.2 Degradation strategy

As tree size grows further, consider:

- default-collapsing deep branches;
- expanding only root and active branches on first paint;
- prompting users to use search-first navigation;
- on-demand expansion for extremely large branches.

---

## 11. Acceptance Criteria

At minimum, the roadmap should land these outcomes:

- at 10k nodes, first paint remains acceptable and scrolling stays smooth with no obvious freezes;
- expanding or collapsing one branch no longer causes visible whole-tree stalls;
- filter input does not cause long main-thread blocking;
- single-node selection and context-menu opening remain near-instant;
- existing hierarchy semantics, drag-to-reparent, path highlight, and multi-select behavior do not regress.

---

## 12. Recommended Implementation Order

Recommended sequence:

1. establish the performance baseline;
2. rebuild the tree data layer;
3. introduce the flattened visible-row model;
4. add virtual scrolling;
5. convert to local updates;
6. switch to event delegation;
7. optimize search and large-tree degradation strategy;
8. add regression samples and fold them into future Creator acceptance.

This roadmap should remain the highest-priority future development item, ahead of new Creator interaction enhancements and general documentation work.
