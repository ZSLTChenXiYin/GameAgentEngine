const baseUrl = process.env.ENGINE_BASE_URL || 'http://127.0.0.1:8080';
const runtimeTaskToken = process.env.RUNTIME_TASK_TOKEN || 'dev-task-token';
const callbackToken = process.env.CALLBACK_TOKEN || 'dev-callback-token';
const consumer = process.env.RUNTIME_TASK_CONSUMER || 'game_client';
const leaseOwner = process.env.RUNTIME_TASK_OWNER || 'local-worker-1';

async function engineFetch(path, init = {}) {
  const headers = {
    'Content-Type': 'application/json',
    'X-Runtime-Task-Token': runtimeTaskToken,
    ...(init.headers || {}),
  };
  const resp = await fetch(baseUrl + path, { ...init, headers });
  if (!resp.ok) {
    throw new Error(`engine request failed: ${resp.status} ${await resp.text()}`);
  }
  return resp.json();
}

async function callbackFetch(body, requestId) {
  const resp = await fetch(baseUrl + '/api/v1/actions/callback', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-Callback-Token': callbackToken,
      'X-Callback-Request-Id': requestId,
    },
    body: JSON.stringify(body),
  });
  if (!resp.ok) {
    throw new Error(`callback failed: ${resp.status} ${await resp.text()}`);
  }
  return resp.json();
}

async function pollOnce() {
  const pending = await engineFetch(`/api/v1/runtime/tasks/pending?consumer=${encodeURIComponent(consumer)}&limit=1`);
  const task = pending.tasks?.[0];
  if (!task) {
    console.log('no pending task');
    return false;
  }

  const claimedBody = await engineFetch('/api/v1/runtime/tasks/claim', {
    method: 'POST',
    body: JSON.stringify({ task_id: task.task_id, consumer, lease_owner: leaseOwner }),
  });
  const claimed = claimedBody.task;

  await engineFetch('/api/v1/runtime/tasks/start', {
    method: 'POST',
    body: JSON.stringify({ task_id: claimed.task_id, lease_token: claimed.lease_token }),
  });

  const result = {
    source: 'runtime_task_pull_worker_example',
    observed_task: claimed.task_id,
  };

  const callback = await callbackFetch({
    callback_id: claimed.callback_id,
    status: 'success',
    result,
  }, `cb-${claimed.task_id}`);

  console.log(JSON.stringify({ task: claimed.task_id, callback }, null, 2));
  return true;
}

pollOnce().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
