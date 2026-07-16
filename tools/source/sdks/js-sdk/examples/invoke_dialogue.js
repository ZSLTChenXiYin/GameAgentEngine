const { GameAgentEngineClient } = require('../src/client');

async function main() {
  const client = new GameAgentEngineClient(process.env.GAE_SERVER || 'http://127.0.0.1:8080', process.env.GAE_KEY || 'dev-key');
  const resp = await client.invoke({ world_id: process.env.GAE_WORLD_ID || 'demo_world', node_id: process.env.GAE_NODE_ID || 'innkeeper_001', task_type: 'npc_dialogue', messages: [{ role: 'user', content: 'Did anyone suspicious come by tonight?' }] });
  console.log(JSON.stringify(resp, null, 2));
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
