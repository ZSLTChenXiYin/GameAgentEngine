class GameAgentEngineClient {
  constructor(baseUrl, apiKey) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }

  async request(path, init) {
    const res = await fetch(this.baseUrl + path, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
        ...(init && init.headers ? init.headers : {}),
      },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status} ${path}: ${await res.text()}`);
    const text = await res.text();
    return text ? JSON.parse(text) : null;
  }

  health() { return this.request('/health', { method: 'GET' }); }
  getVersion() { return this.request('/api/v1/version', { method: 'GET' }); }
  invoke(body) { return this.request('/api/v1/invoke', { method: 'POST', body: JSON.stringify(body) }); }
  listPendingRuntimeTasks(consumer, limit = 20) {
    const q = new URLSearchParams({ consumer, limit: String(limit) });
    return this.request('/api/v1/runtime/tasks/pending?' + q.toString(), { method: 'GET' });
  }
  claimRuntimeTask(taskId, consumer, owner) {
    return this.request('/api/v1/runtime/tasks/claim', { method: 'POST', body: JSON.stringify({ task_id: taskId, consumer, lease_owner: owner }) });
  }
  startRuntimeTask(taskId, leaseToken) {
    return this.request('/api/v1/runtime/tasks/start', { method: 'POST', body: JSON.stringify({ task_id: taskId, lease_token: leaseToken }) });
  }
  actionCallback(callbackId, status, result) {
    return this.request('/api/v1/actions/callback', { method: 'POST', body: JSON.stringify({ callback_id: callbackId, status, result }) });
  }
}

module.exports = { GameAgentEngineClient };
