import type {
  CallbackResponse,
  DebugTraceList,
  InferenceLog,
  InferenceLogQuery,
  InvokeRequest,
  InvokeResponse,
  LatestTimelineResponse,
  PlayerInputInterpretRequest,
  RequestOptions,
  RuntimeTask,
  RuntimeTaskStats,
  StateComponentResponse,
  StateComponentsResponse,
  TickResponse,
  TimelinesResponse,
  VersionInfo,
  WorldSettings,
} from './types';

export class GameAgentEngineError extends Error {
  constructor(
    message: string,
    public readonly status?: number,
    public readonly path?: string,
    public readonly body?: unknown,
  ) {
    super(message);
    this.name = 'GameAgentEngineError';
  }
}

export type ClientOptions = {
  baseUrl: string;
  apiKey: string;
  fetchImpl?: typeof fetch;
  defaultHeaders?: Record<string, string>;
};

export class GameAgentEngineClient {
  public readonly baseUrl: string;
  public readonly apiKey: string;
  private readonly fetchImpl: typeof fetch;
  private readonly defaultHeaders: Record<string, string>;

  constructor(baseUrl: string, apiKey: string, fetchImpl?: typeof fetch);
  constructor(options: ClientOptions);
  constructor(
    baseUrlOrOptions: string | ClientOptions,
    apiKey?: string,
    fetchImpl?: typeof fetch,
  ) {
    const options =
      typeof baseUrlOrOptions === 'string'
        ? {
            baseUrl: baseUrlOrOptions,
            apiKey: apiKey ?? '',
            fetchImpl,
          }
        : baseUrlOrOptions;

    if (!options.baseUrl) {
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
    this.fetchImpl = options.fetchImpl ?? fetch;
    this.defaultHeaders = options.defaultHeaders ?? {};
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
    options?: RequestOptions,
  ): Promise<T> {
    const headers: Record<string, string> = {
      Accept: 'application/json',
      'X-API-Key': this.apiKey,
      ...this.defaultHeaders,
      ...(options?.headers ?? {}),
    };

    let payload: BodyInit | undefined;
    if (body !== undefined) {
      headers['Content-Type'] = 'application/json';
      payload = JSON.stringify(body);
    }

    const response = await this.fetchImpl(this.baseUrl + path, {
      method,
      headers,
      body: payload,
      signal: options?.signal,
    });

    const rawText = await response.text();
    const parsed = rawText ? safeParseJSON(rawText) : null;

    if (!response.ok) {
      throw new GameAgentEngineError(
        `HTTP ${response.status} ${method} ${path}`,
        response.status,
        path,
        parsed ?? rawText,
      );
    }

    return parsed as T;
  }

  private buildQuery(params: Record<string, string | number | undefined | null>) {
    const query = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value === undefined || value === null || value === '') {
        continue;
      }
      query.set(key, String(value));
    }
    const text = query.toString();
    return text ? `?${text}` : '';
  }

  health(options?: RequestOptions) {
    return this.request<unknown>('GET', '/health', undefined, options);
  }

  getVersion(options?: RequestOptions) {
    return this.request<VersionInfo>('GET', '/api/v1/version', undefined, options);
  }

  invoke(body: InvokeRequest, options?: RequestOptions) {
    return this.request<InvokeResponse>('POST', '/api/v1/invoke', body, options);
  }

  interpretPlayerInput(body: PlayerInputInterpretRequest, options?: RequestOptions) {
    return this.request<InvokeResponse>('POST', '/api/v1/player/input/interpret', body, options);
  }

  advanceTick(
    worldId: string,
    body: {
      tick_type: string;
      game_time: string;
      requested_ticks?: number;
      autonomous_limit?: number;
    },
    options?: RequestOptions,
  ) {
    return this.request<TickResponse>('POST', `/api/v1/worlds/${encodeURIComponent(worldId)}/ticks/advance`, body, options);
  }

  getWorldSettings(worldId: string, options?: RequestOptions) {
    return this.request<WorldSettings>('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/settings`, undefined, options);
  }

  setWorldSettings(worldId: string, settings: Partial<WorldSettings>, options?: RequestOptions) {
    return this.request<WorldSettings>('PUT', `/api/v1/worlds/${encodeURIComponent(worldId)}/settings`, settings, options);
  }

  getStateComponents(worldId: string, options?: RequestOptions) {
    return this.request<StateComponentsResponse>('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/state-components`, undefined, options);
  }

  getStateComponent(worldId: string, componentType: string, options?: RequestOptions) {
    return this.request<StateComponentResponse>(
      'GET',
      `/api/v1/worlds/${encodeURIComponent(worldId)}/state-components/${encodeURIComponent(componentType)}`,
      undefined,
      options,
    );
  }

  putStateComponent(worldId: string, componentType: string, payload: unknown, options?: RequestOptions) {
    return this.request<StateComponentResponse>(
      'PUT',
      `/api/v1/worlds/${encodeURIComponent(worldId)}/state-components/${encodeURIComponent(componentType)}`,
      payload,
      options,
    );
  }

  getTimelines(worldId: string, limit?: number, options?: RequestOptions) {
    const suffix = this.buildQuery({ limit });
    return this.request<TimelinesResponse>('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/timelines${suffix}`, undefined, options);
  }

  getLatestTimeline(worldId: string, options?: RequestOptions) {
    return this.request<LatestTimelineResponse>('GET', `/api/v1/worlds/${encodeURIComponent(worldId)}/timelines/latest`, undefined, options);
  }

  getLogs(query: InferenceLogQuery = {}, options?: RequestOptions) {
    const suffix = this.buildQuery(query);
    return this.request<InferenceLog[]>('GET', `/api/v1/logs${suffix}`, undefined, options);
  }

  getDebugTraces(worldId: string, limit = 20, options?: RequestOptions) {
    const suffix = this.buildQuery({ world_id: worldId, limit });
    return this.request<DebugTraceList>('GET', `/debug/traces${suffix}`, undefined, options);
  }

  listRuntimeTasks(category?: string, status?: string, limit = 20, options?: RequestOptions) {
    const suffix = this.buildQuery({ category, status, limit });
    return this.request<{ tasks: RuntimeTask[] }>('GET', `/api/v1/runtime/tasks${suffix}`, undefined, options);
  }

  listPendingRuntimeTasks(consumer: string, limit = 20, options?: RequestOptions) {
    const suffix = this.buildQuery({ consumer, limit });
    return this.request<{ tasks: RuntimeTask[] }>('GET', `/api/v1/runtime/tasks/pending${suffix}`, undefined, options);
  }

  getRuntimeTask(taskId: string, options?: RequestOptions) {
    return this.request<{ task: RuntimeTask }>('GET', `/api/v1/runtime/tasks/${encodeURIComponent(taskId)}`, undefined, options);
  }

  claimRuntimeTask(taskId: string, consumer: string, leaseOwner: string, options?: RequestOptions) {
    return this.request<{ task: RuntimeTask }>(
      'POST',
      '/api/v1/runtime/tasks/claim',
      { task_id: taskId, consumer, lease_owner: leaseOwner },
      options,
    );
  }

  startRuntimeTask(taskId: string, leaseToken: string, options?: RequestOptions) {
    return this.request<{ task: RuntimeTask }>(
      'POST',
      '/api/v1/runtime/tasks/start',
      { task_id: taskId, lease_token: leaseToken },
      options,
    );
  }

  heartbeatRuntimeTask(taskId: string, leaseToken: string, options?: RequestOptions) {
    return this.request<unknown>(
      'POST',
      '/api/v1/runtime/tasks/heartbeat',
      { task_id: taskId, lease_token: leaseToken },
      options,
    );
  }

  releaseRuntimeTask(taskId: string, leaseToken: string, errorMessage = '', options?: RequestOptions) {
    return this.request<unknown>(
      'POST',
      '/api/v1/runtime/tasks/release',
      { task_id: taskId, lease_token: leaseToken, error_message: errorMessage },
      options,
    );
  }

  requeueRuntimeTask(taskId: string, retryDelayMs = 0, errorMessage = '', options?: RequestOptions) {
    return this.request<{ task: RuntimeTask }>(
      'POST',
      '/api/v1/runtime/tasks/requeue',
      { task_id: taskId, retry_delay_ms: retryDelayMs, error_message: errorMessage },
      options,
    );
  }

  getRuntimeTaskStats(options?: RequestOptions) {
    return this.request<{ stats: RuntimeTaskStats }>('GET', '/api/v1/runtime/tasks/stats', undefined, options);
  }

  actionCallback(
    callbackId: string,
    status: string,
    result: unknown,
    requestId?: string,
    options?: RequestOptions,
  ) {
    const headers = {
      ...(options?.headers ?? {}),
      ...(requestId ? { 'X-Callback-Request-Id': requestId } : {}),
    };
    return this.request<CallbackResponse>(
      'POST',
      '/api/v1/actions/callback',
      { callback_id: callbackId, status, result },
      { ...options, headers },
    );
  }
}

function safeParseJSON(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

