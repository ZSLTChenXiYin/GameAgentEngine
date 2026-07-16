const { GameAgentEngineClient } = require('../src/client');

async function main() {
  const client = new GameAgentEngineClient(process.env.GAE_SERVER || 'http://127.0.0.1:8080', process.env.GAE_KEY || 'dev-key');
  const consumer = process.env.GAE_CONSUMER || 'game_client';
  const owner = process.env.GAE_OWNER || 'js-sdk-example';
  const pending = await client.listPendingRuntimeTasks(consumer, 1);
  const task = pending && pending.tasks && pending.tasks[0];
  if (!task) return console.log('No pending tasks.');
  const claimed = await client.claimRuntimeTask(task.task_id, consumer, owner);
  const lease = claimed && claimed.task && claimed.task.lease_token;
  if (!lease) throw new Error('missing lease token');
  const started = await client.startRuntimeTask(task.task_id, lease);
  console.log(JSON.stringify({
    task_id: task.task_id,
    interface_name: task.interface_name,
    claimed_status: claimed && claimed.task && claimed.task.status,
    started_status: started && started.task && started.task.status,
    lease_token_present: Boolean(lease),
    callback_id: (started && started.task && started.task.callback_id) || task.callback_id,
  }, null, 2));
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
