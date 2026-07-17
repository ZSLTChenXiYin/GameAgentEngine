import { readFile } from 'node:fs/promises';
import { GameAgentEngineClient } from '../src/client';
import type { DynamicInterface, InvokeRequest } from '../src/types';

function required(name: string, fallback?: string): string {
  const value = process.env[name] ?? fallback;
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

async function loadDynamicInterfaces(path: string): Promise<DynamicInterface[]> {
  const text = await readFile(path, 'utf8');
  return JSON.parse(text) as DynamicInterface[];
}

async function main() {
  const client = new GameAgentEngineClient(
    process.env.GAE_SERVER || 'http://127.0.0.1:8080',
    process.env.GAE_KEY || 'dev-key',
  );
  const dynamicInterfacesPath = process.env.GAE_DYNAMIC_INTERFACES_FILE || 'tools/source/workerhome/fixtures/runtime_task_dynamic_interfaces.json';
  const dynamicInterfaces = await loadDynamicInterfaces(dynamicInterfacesPath);

  const request: InvokeRequest = {
    world_id: required('GAE_WORLD_ID', 'demo_world'),
    node_id: required('GAE_NODE_ID', 'innkeeper_001'),
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
  };

  const response = await client.invoke(request);
  console.log(JSON.stringify({
    request_id: response.request_id,
    task_type: response.task_type,
    execution_mode: response.execution_mode,
    reply: response.reply,
    action_calls: response.action_calls || [],
  }, null, 2));

  const dataRequest = (response.action_calls || []).find(
    (item) => item.action_id === 'data_request' && item.mode === 'async',
  );

  if (!dataRequest?.callback_id) {
    console.log('No async data request was emitted.');
    return;
  }

  const pending = await client.listRuntimeTasks(undefined, undefined, 20);
  const task = pending.tasks.find((item) => item.callback_id === dataRequest.callback_id);
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

  console.log('Process it with GameAgentWorker, for example:');
  console.log(`GameAgentWorker pull-once --consumer ${task.consumer || 'game_client'}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});

