# GameAgentEngine TypeScript SDK

TypeScript client SDK for GameAgentEngine.

## Installation

```bash
npm install @gameagentengine/sdk
```

## Usage

```typescript
import { Client, TaskType } from "@gameagentengine/sdk";

const client = new Client("http://127.0.0.1:8080", "dev-key");

// Create a world
const world = await client.createNode("", "My World", "world");

// Get all worlds
const worlds = await client.getWorlds();

// Create a node
const npc = await client.createNode(world.id, "Innkeeper", "npc");

// Add a component
await client.addComponent(npc.id, "profile", JSON.stringify({
  name: "Innkeeper",
  description: "A friendly innkeeper"
}));

// Execute an interaction
const resp = await client.executeInteraction({
  world_id: world.id,
  actor_node_id: npc.id,
  target_node_id: "player_001",
  task_type: TaskType.NPCDialogue,
  message: "Hello! Welcome to the inn.",
});

// Advance a world tick
const tick = await client.advanceTick(world.id);
```

## API

See the Go SDK documentation in `docs/sdk/` for the full API reference.

## Building

```bash
npm run build
```
