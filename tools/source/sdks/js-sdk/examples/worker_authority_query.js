const fs = require('node:fs/promises');
const { GameAgentEngineClient } = require('../src/client');

async function main() {
  const client = new GameAgentEngineClient(process.env.GAE_SERVER || 'http://127.0.0.1:8080', process.env.GAE_KEY || 'dev-key');
  const dynamicInterfacesFile = process.env.GAE_DYNAMIC_INTERFACES_FILE || 'tools/source/workerhome/fixtures/runtime_task_dynamic_interfaces.json';
  const dynamicInterfaces = JSON.parse(await fs.readFile(dynamicInterfacesFile, 'utf8'));

  const response = await client.invoke({
    world_id: process.env.GAE_WORLD_ID || 'demo_world',
    node_id: process.env.GAE_NODE_ID || 'innkeeper_001',
    task_type: process.env.GAE_TASK_TYPE || 'npc_dialogue',
    messages: [
      {
        role: 'user',
        content: process.env.GAE_MESSAGE || 'Before answering, query the nearby scene state and then respond.',
      },
    ],
    context: {
      pipeline_mode: process.env.GAE_PIPELINE_MODE || 'full',
      dynamic_interfaces: dynamicInterfaces,
    },
  });

  console.log(JSON.stringify({
    request_id: response && response.request_id,
    task_type: response && response.task_type,
    execution_mode: response && response.execution_mode,
    reply: response && response.reply,
    action_calls: (response && response.action_calls) || [],
  }, null, 2));

  const dataRequest = ((response && response.action_calls) || []).find(
    (item) => item.action_id === 'data_request' && item.mode === 'async',
  );

  if (!dataRequest || !dataRequest.callback_id) {
    console.log('No async data request was emitted.');
    return;
  }

  const tasks = await client.listRuntimeTasks(undefined, undefined, 20);
  const task = (tasks && tasks.tasks || []).find((item) => item.callback_id === dataRequest.callback_id);
  if (!task) {
    throw new Error(`Runtime task not found for callback_id=${dataRequest.callback_id}`);
  }

  console.log(JSON.stringify({
    runtime_task_id: task.task_id,
    interface_name: task.interface_name,
    consumer: task.consumer,
    status: task.status,
    callback_id: task.callback_id,
  }, null, 2));

  console.log(`Next step: GameAgentWorker pull-once --consumer ${task.consumer || 'game_client'}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

