class GameAgentEngineError extends Error {
  constructor(message, status, path, body) {
    super(message);
    this.name = 'GameAgentEngineError';
    this.status = status;
    this.path = path;
    this.body = body;
  }
}

class GameAgentEngineClient {
  constructor(baseUrlOrOptions, apiKey, fetchImpl) {
    const options = typeof baseUrlOrOptions === 'string'
      ? { baseUrl: baseUrlOrOptions, apiKey, fetchImpl }
      : baseUrlOrOptions;

    if (!options || !options.baseUrl) {
      throw new Error('baseUrl is required');
    }
    if (!options.apiKey) {
      throw new Error('apiKey is required');
    }
    if (!options.fetchImpl && typeof fetch !== 'function') {
      throw new Error('fetch implementation is required in this runtime');
    }

    this.baseUrl = options.baseUrl.replace(/\/+$/, '');
    this.apiKey = options.apiKey;
    this.fetchImpl = options.fetchImpl || fetch;
    this.defaultHeaders = options.defaultHeaders || {};
  }

  async request(method, path, body, options) {
    const headers = {
      Accept: 'application/json',
      'X-API-Key': this.apiKey,
      ...this.defaultHeaders,
      ...((options && options.headers) || {}),
    };

    let payload;
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json';
      payload = JSON.stringify(body);
    }

    const response = await this.fetchImpl(this.baseUrl + path, {
      method,
      headers,
      body: payload,
      signal: options && options.signal,
    });

    const rawText = await response.text();
    const parsed = rawText ? safeParseJSON(rawText) : null;
    if (!response.ok) {
      throw new GameAgentEngineError(
        `HTTP ${response.status} ${method} ${path}`,
        response.status,
        path,
        parsed || rawText,
      );
    }
    return parsed;
  }

  buildQuery(params) {
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(params || {})) {
      if (value === undefined || value === null || value === '') {
        continue;
      }
      query.set(key, String(value));
    }
    const text = query.toString();
    return text ? `?${text}` : '';
  }

  health(options) { return this.request('GET', '/health', undefined, options); }
  getVersion(options) { return this.request('GET', '/api/v1/version', undefined, options); }
  invoke(body, options) { return this.request('POST', '/api/v1/invoke', body, options); }
  interpretPlayerInput(body, options) { return this.request('POST', '/api/v1/player/input/interpret', body, options); }
  advanceTick(worldId, body, options) { return this.request('POST', `/api/v1/worlds/${encodeURIComponent(worldId)}/ticks/advance`, body, options); }
  getWorldSettings(worldId, options) { return this.request('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/settings`, undefined, options); }
  setWorldSettings(worldId, settings, options) { return this.request('PUT', `/api/v1/worlds/${encodeURIComponent(worldId)}/settings`, settings, options); }
  getStateComponents(worldId, options) { return this.request('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/state-components`, undefined, options); }
  getStateComponent(worldId, componentType, options) { return this.request('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/state-components/${encodeURIComponent(componentType)}`, undefined, options); }
  putStateComponent(worldId, componentType, payload, options) { return this.request('PUT', `/api/v1/worlds/${encodeURIComponent(worldId)}/state-components/${encodeURIComponent(componentType)}`, payload, options); }
  getTimelines(worldId, limit, options) { return this.request('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/timelines${this.buildQuery({ limit })}`, undefined, options); }
  getLatestTimeline(worldId, options) { return this.request('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/timelines/latest`, undefined, options); }
  getLogs(query = {}, options) { return this.request('GET', `/api/v1/logs${this.buildQuery(query)}`, undefined, options); }
  getDebugTraces(worldId, limit = 20, options) { return this.request('GET', `/debug/traces${this.buildQuery({ world_id: worldId, limit })}`, undefined, options); }
  listRuntimeTasks(category, status, limit = 20, options) { return this.request('GET', `/api/v1/runtime/tasks${this.buildQuery({ category, status, limit })}`, undefined, options); }
  listPendingRuntimeTasks(consumer, limit = 20, options) { return this.request('GET', `/api/v1/runtime/tasks/pending${this.buildQuery({ consumer, limit })}`, undefined, options); }
  getRuntimeTask(taskId, options) { return this.request('GET', `/api/v1/runtime/tasks/${encodeURIComponent(taskId)}`, undefined, options); }
  claimRuntimeTask(taskId, consumer, owner, options) { return this.request('POST', '/api/v1/runtime/tasks/claim', { task_id: taskId, consumer, lease_owner: owner }, options); }
  startRuntimeTask(taskId, leaseToken, options) { return this.request('POST', '/api/v1/runtime/tasks/start', { task_id: taskId, lease_token: leaseToken }, options); }
  heartbeatRuntimeTask(taskId, leaseToken, options) { return this.request('POST', '/api/v1/runtime/tasks/heartbeat', { task_id: taskId, lease_token: leaseToken }, options); }
  releaseRuntimeTask(taskId, leaseToken, errorMessage = '', options) { return this.request('POST', '/api/v1/runtime/tasks/release', { task_id: taskId, lease_token: leaseToken, error_message: errorMessage }, options); }
  requeueRuntimeTask(taskId, retryDelayMs = 0, errorMessage = '', options) { return this.request('POST', '/api/v1/runtime/tasks/requeue', { task_id: taskId, retry_delay_ms: retryDelayMs, error_message: errorMessage }, options); }
  getRuntimeTaskStats(options) { return this.request('GET', '/api/v1/runtime/tasks/stats', undefined, options); }
  actionCallback(callbackId, status, result, requestId, options) {
    const headers = {
      ...((options && options.headers) || {}),
      ...(requestId ? { 'X-Callback-Request-Id': requestId } : {}),
    };
    return this.request('POST', '/api/v1/actions/callback', { callback_id: callbackId, status, result }, { ...(options || {}), headers });
  }
}

function safeParseJSON(text) {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

module.exports = { GameAgentEngineClient, GameAgentEngineError };
