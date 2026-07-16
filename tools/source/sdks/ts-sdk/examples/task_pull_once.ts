import { GameAgentEngineClient } from '../src/client';

async function main() {
  const client = new GameAgentEngineClient(process.env.GAE_SERVER || 'http://127.0.0.1:8080', process.env.GAE_KEY || 'dev-key');
  const consumer = process.env.GAE_CONSUMER || 'game_client';
  const owner = process.env.GAE_OWNER || 'ts-sdk-example';
  const pending = await client.listPendingRuntimeTasks(consumer, 1);
  const task = pending?.tasks?.[0];
  if (!task) return console.log('No pending tasks.');
  const claimed = await client.claimRuntimeTask(task.task_id, consumer, owner);
  const lease = claimed?.task?.lease_token;
  if (!lease) throw new Error('missing lease token');
  await client.startRuntimeTask(task.task_id, lease);
  console.log('Started task', task.task_id);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
