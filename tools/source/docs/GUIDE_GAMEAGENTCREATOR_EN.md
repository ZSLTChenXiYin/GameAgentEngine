# GameAgentCreator Guide

[**中文**](./GUIDE_GAMEAGENTCREATOR.md) | **English**

GameAgentCreator is the web-based visual editor for GameAgentEngine v0.2.0. It communicates with the engine HTTP API through the browser, providing node tree browsing, component editing, memory viewing, world operations, and more.

---

## Launch Methods

```bash
# Method 1: Open via DevCli
GameAgentDevCli inspect

# Method 2: Open directly in the browser
# Open tools/source/web/GameAgentCreator/index.html
```

---

## Interface Layout

The Creator interface follows mainstream game engine editor layouts, divided into the following areas:

- **Top navigation bar**: world selection, language switch, configuration entry
- **Left node tree**: displays all nodes hierarchically with collapse and drag support
- **Center workspace**: shows component details of the selected node, supports key-value editing
- **Right inspector**: recursively displays all data of the selected node (including properties and sub-objects)
- **Bottom monitor**: runtime information, log output

All sub-panels can be freely resized and scrolled.

---

## Core Features

### Node Tree

- Displays world nodes in parent-child hierarchy
- Different node types are identified by colored rounded labels
- Selected node is highlighted, with ancestor path shown in a lighter shade
- Supports drag-and-drop to attach child nodes

### Workspace

- View and edit component content of the selected node
- Component name and action buttons are shown in the top row
- KV key-value pairs recursively display all nesting levels
- Supports adding, deleting, and modifying key-value pairs

### Inspector

- Recursively displays key-value pairs of node properties
- Hover tooltip shows full content
- Horizontal and vertical scrollbars ensure content doesn't overflow the screen

### World Operations

- **Create World**: create a new world from the navigation bar
- **Import Config**: import world configuration from a file
- **Clone World**: click "Clone World" on the world page, with a dialog asking whether to lock the source world
- **World Settings**: modify PipelineMode, memory limit, propagation parameters, and other dynamic configuration
- **Advance Tick**: advance the world timeline
- **Run Autonomous Behavior**: trigger node autonomous behavior
- **Event Impact**: evaluate an event's impact on the world
- **Scope Advance**: advance evolution within a specific scope
- **Replan**: regenerate the world outline