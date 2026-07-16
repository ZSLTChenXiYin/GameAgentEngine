const { GameAgentEngineClient } = require('../src/client');

function required(name, fallback) {
  const value = process.env[name] || fallback;
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

async function main() {
  const client = new GameAgentEngineClient(process.env.GAE_SERVER || 'http://127.0.0.1:8080', process.env.GAE_KEY || 'dev-key');
  const consumer = process.env.GAE_CONSUMER || 'game_client';
  const owner = process.env.GAE_OWNER || 'js-sdk-roundtrip';
  const resultStatus = process.env.GAE_CALLBACK_STATUS || 'success';

  const pending = await client.listPendingRuntimeTasks(consumer, 1);
  const task = pending && pending.tasks && pending.tasks[0];
  if (!task) {
    console.log(`No pending runtime task for consumer=${consumer}.`);
    return;
  }

  console.log(`Claiming task ${task.task_id} (${task.interface_name || 'unknown_interface'})`);
  const claimed = await client.claimRuntimeTask(task.task_id, consumer, owner);
  const claimedTask = claimed && claimed.task;
  if (!claimedTask || !claimedTask.lease_token) {
    throw new Error(`Task ${task.task_id} missing lease token after claim.`);
  }

  const started = await client.startRuntimeTask(task.task_id, claimedTask.lease_token);
  console.log(`Started task ${task.task_id} status=${started && started.task ? started.task.status : 'unknown'}`);

  const callbackId = required('GAE_CALLBACK_ID', (started && started.task && started.task.callback_id) || task.callback_id);
  const callbackRequestId = process.env.GAE_CALLBACK_REQUEST_ID || `js-sdk-${task.task_id}`;
  const result = {
    worker: 'js-sdk-example',
    source: 'worker_runtime_roundtrip',
    interface_name: task.interface_name,
    task_id: task.task_id,
    consumer,
  };

  const callback = await client.actionCallback(callbackId, resultStatus, result, callbackRequestId);
  console.log(JSON.stringify({
    task_id: task.task_id,
    callback_id: callbackId,
    callback_status: callback && callback.status,
    resume_execution_id: callback && callback.resume_execution_id,
    resumed: Boolean(callback && callback.resumed),
    post_process_applied: Boolean(callback && callback.post_process && callback.post_process.applied),
  }, null, 2));
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

