const { GameAgentEngineClient } = require('../src/client');

async function main() {
  const client = new GameAgentEngineClient(process.env.GAE_SERVER || 'http://127.0.0.1:8080', process.env.GAE_KEY || 'dev-key');
  console.log(await client.health());
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
