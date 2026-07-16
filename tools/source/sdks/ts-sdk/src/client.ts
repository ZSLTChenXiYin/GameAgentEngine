export type InvokeRequest = {
  world_id: string;
  task_type: string;
  node_id: string;
  session_id?: string;
  messages?: Array<{ role: string; content: string }>;
  context?: Record<string, unknown>;
};

export class GameAgentEngineClient {
  constructor(
    public readonly baseUrl: string,
    public readonly apiKey: string,
  ) {}

  private async request(path: string, init?: RequestInit) {
    const res = await fetch(this.baseUrl + path, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
        ...(init?.headers || {}),
      },
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(`HTTP ${res.status} ${path}: ${text}`);
    }
    const text = await res.text();
    return text ? JSON.parse(text) : null;
  }

  health() { return this.request('/health', { method: 'GET' }); }
  getVersion() { return this.request('/api/v1/version', { method: 'GET' }); }
  invoke(body: InvokeRequest) { return this.request('/api/v1/invoke', { method: 'POST', body: JSON.stringify(body) }); }
  listPendingRuntimeTasks(consumer: string, limit = 20) {
    const q = new URLSearchParams({ consumer, limit: String(limit) });
    return this.request('/api/v1/runtime/tasks/pending?' + q.toString(), { method: 'GET' });
  }
  claimRuntimeTask(taskId: string, consumer: string, owner: string) {
    return this.request('/api/v1/runtime/tasks/claim', { method: 'POST', body: JSON.stringify({ task_id: taskId, consumer, lease_owner: owner }) });
  }
  startRuntimeTask(taskId: string, leaseToken: string) {
    return this.request('/api/v1/runtime/tasks/start', { method: 'POST', body: JSON.stringify({ task_id: taskId, lease_token: leaseToken }) });
  }
  actionCallback(callbackId: string, status: string, result: unknown) {
    return this.request('/api/v1/actions/callback', { method: 'POST', body: JSON.stringify({ callback_id: callbackId, status, result }) });
  }
}
