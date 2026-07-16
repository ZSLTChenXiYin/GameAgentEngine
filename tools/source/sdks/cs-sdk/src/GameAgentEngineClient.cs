using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace GameAgentEngine.SDK;

public sealed class GameAgentEngineClient
{
    private readonly HttpClient _http;
    private readonly JsonSerializerOptions _json;
    private readonly string _baseUrl;

    public GameAgentEngineClient(string baseUrl, string apiKey, HttpClient? httpClient = null, JsonSerializerOptions? jsonSerializerOptions = null)
    {
        if (string.IsNullOrWhiteSpace(baseUrl))
        {
            throw new ArgumentException("baseUrl is required", nameof(baseUrl));
        }

        if (string.IsNullOrWhiteSpace(apiKey))
        {
            throw new ArgumentException("apiKey is required", nameof(apiKey));
        }

        _baseUrl = baseUrl.TrimEnd('/');
        _http = httpClient ?? new HttpClient();
        _json = jsonSerializerOptions ?? JsonDefaults.Create();

        _http.DefaultRequestHeaders.Remove("X-API-Key");
        _http.DefaultRequestHeaders.Add("X-API-Key", apiKey);
        if (!_http.DefaultRequestHeaders.Accept.Any(h => h.MediaType == "application/json"))
        {
            _http.DefaultRequestHeaders.Accept.Add(new MediaTypeWithQualityHeaderValue("application/json"));
        }
    }

    public Task<JsonElement?> HealthAsync(CancellationToken cancellationToken = default) =>
        SendAsync<JsonElement?>(HttpMethod.Get, "/health", null, null, cancellationToken);

    public Task<VersionInfo?> GetVersionAsync(CancellationToken cancellationToken = default) =>
        SendAsync<VersionInfo>(HttpMethod.Get, "/api/v1/version", null, null, cancellationToken);

    public Task<InvokeResponse?> InvokeAsync(InvokeRequest request, CancellationToken cancellationToken = default) =>
        SendAsync<InvokeResponse>(HttpMethod.Post, "/api/v1/invoke", request, null, cancellationToken);

    public Task<InvokeResponse?> InterpretPlayerInputAsync(PlayerInputInterpretRequest request, CancellationToken cancellationToken = default) =>
        SendAsync<InvokeResponse>(HttpMethod.Post, "/api/v1/player/input/interpret", request, null, cancellationToken);

    public Task<TickResponse?> AdvanceTickAsync(string worldId, TickAdvanceRequest request, CancellationToken cancellationToken = default) =>
        SendAsync<TickResponse>(HttpMethod.Post, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/ticks/advance", request, null, cancellationToken);

    public Task<WorldSettings?> GetWorldSettingsAsync(string worldId, CancellationToken cancellationToken = default) =>
        SendAsync<WorldSettings>(HttpMethod.Get, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/settings", null, null, cancellationToken);

    public Task<WorldSettings?> SetWorldSettingsAsync(string worldId, object settings, CancellationToken cancellationToken = default) =>
        SendAsync<WorldSettings>(HttpMethod.Put, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/settings", settings, null, cancellationToken);

    public Task<StateComponentsResponse?> GetStateComponentsAsync(string worldId, CancellationToken cancellationToken = default) =>
        SendAsync<StateComponentsResponse>(HttpMethod.Get, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/state-components", null, null, cancellationToken);

    public Task<StateComponentResponse?> GetStateComponentAsync(string worldId, string componentType, CancellationToken cancellationToken = default) =>
        SendAsync<StateComponentResponse>(HttpMethod.Get, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/state-components/{Uri.EscapeDataString(componentType)}", null, null, cancellationToken);

    public Task<StateComponentResponse?> PutStateComponentAsync(string worldId, string componentType, object payload, CancellationToken cancellationToken = default) =>
        SendAsync<StateComponentResponse>(HttpMethod.Put, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/state-components/{Uri.EscapeDataString(componentType)}", payload, null, cancellationToken);

    public Task<TimelinesResponse?> GetTimelinesAsync(string worldId, int? limit = null, CancellationToken cancellationToken = default) =>
        SendAsync<TimelinesResponse>(HttpMethod.Get, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/timelines{BuildQuery(("limit", limit))}", null, null, cancellationToken);

    public Task<LatestTimelineResponse?> GetLatestTimelineAsync(string worldId, CancellationToken cancellationToken = default) =>
        SendAsync<LatestTimelineResponse>(HttpMethod.Get, $"/api/v1/worlds/{Uri.EscapeDataString(worldId)}/timelines/latest", null, null, cancellationToken);

    public Task<List<InferenceLog>?> GetLogsAsync(InferenceLogQuery? query = null, CancellationToken cancellationToken = default) =>
        SendAsync<List<InferenceLog>>(HttpMethod.Get, $"/api/v1/logs{BuildQuery(query)}", null, null, cancellationToken);

    public Task<DebugTraceList?> GetDebugTracesAsync(string worldId, int limit = 20, CancellationToken cancellationToken = default) =>
        SendAsync<DebugTraceList>(HttpMethod.Get, $"/debug/traces{BuildQuery(("world_id", worldId), ("limit", limit))}", null, null, cancellationToken);

    public Task<RuntimeTaskListResponse?> ListRuntimeTasksAsync(string? category = null, string? status = null, int limit = 20, CancellationToken cancellationToken = default) =>
        SendAsync<RuntimeTaskListResponse>(HttpMethod.Get, $"/api/v1/runtime/tasks{BuildQuery(("category", category), ("status", status), ("limit", limit))}", null, null, cancellationToken);

    public Task<RuntimeTaskListResponse?> ListPendingRuntimeTasksAsync(string consumer, int limit = 20, CancellationToken cancellationToken = default) =>
        SendAsync<RuntimeTaskListResponse>(HttpMethod.Get, $"/api/v1/runtime/tasks/pending{BuildQuery(("consumer", consumer), ("limit", limit))}", null, null, cancellationToken);

    public Task<RuntimeTaskEnvelope?> GetRuntimeTaskAsync(string taskId, CancellationToken cancellationToken = default) =>
        SendAsync<RuntimeTaskEnvelope>(HttpMethod.Get, $"/api/v1/runtime/tasks/{Uri.EscapeDataString(taskId)}", null, null, cancellationToken);

    public Task<RuntimeTaskEnvelope?> ClaimRuntimeTaskAsync(string taskId, string consumer, string leaseOwner, CancellationToken cancellationToken = default) =>
        SendAsync<RuntimeTaskEnvelope>(HttpMethod.Post, "/api/v1/runtime/tasks/claim", new { task_id = taskId, consumer, lease_owner = leaseOwner }, null, cancellationToken);

    public Task<RuntimeTaskEnvelope?> StartRuntimeTaskAsync(string taskId, string leaseToken, CancellationToken cancellationToken = default) =>
        SendAsync<RuntimeTaskEnvelope>(HttpMethod.Post, "/api/v1/runtime/tasks/start", new { task_id = taskId, lease_token = leaseToken }, null, cancellationToken);

    public Task<JsonElement?> HeartbeatRuntimeTaskAsync(string taskId, string leaseToken, CancellationToken cancellationToken = default) =>
        SendAsync<JsonElement?>(HttpMethod.Post, "/api/v1/runtime/tasks/heartbeat", new { task_id = taskId, lease_token = leaseToken }, null, cancellationToken);

    public Task<JsonElement?> ReleaseRuntimeTaskAsync(string taskId, string leaseToken, string? errorMessage = null, CancellationToken cancellationToken = default) =>
        SendAsync<JsonElement?>(HttpMethod.Post, "/api/v1/runtime/tasks/release", new { task_id = taskId, lease_token = leaseToken, error_message = errorMessage ?? string.Empty }, null, cancellationToken);

    public Task<RuntimeTaskEnvelope?> RequeueRuntimeTaskAsync(string taskId, int retryDelayMs = 0, string? errorMessage = null, CancellationToken cancellationToken = default) =>
        SendAsync<RuntimeTaskEnvelope>(HttpMethod.Post, "/api/v1/runtime/tasks/requeue", new { task_id = taskId, retry_delay_ms = retryDelayMs, error_message = errorMessage ?? string.Empty }, null, cancellationToken);

    public Task<RuntimeTaskStatsEnvelope?> GetRuntimeTaskStatsAsync(CancellationToken cancellationToken = default) =>
        SendAsync<RuntimeTaskStatsEnvelope>(HttpMethod.Get, "/api/v1/runtime/tasks/stats", null, null, cancellationToken);

    public Task<CallbackResponse?> ActionCallbackAsync(string callbackId, string status, object result, string? requestId = null, CancellationToken cancellationToken = default)
    {
        Dictionary<string, string>? headers = null;
        if (!string.IsNullOrWhiteSpace(requestId))
        {
            headers = new Dictionary<string, string>
            {
                ["X-Callback-Request-Id"] = requestId,
            };
        }

        return SendAsync<CallbackResponse>(HttpMethod.Post, "/api/v1/actions/callback", new { callback_id = callbackId, status, result }, headers, cancellationToken);
    }

    private async Task<T?> SendAsync<T>(HttpMethod method, string path, object? body, IDictionary<string, string>? extraHeaders, CancellationToken cancellationToken)
    {
        using var request = new HttpRequestMessage(method, _baseUrl + path);
        if (extraHeaders is not null)
        {
            foreach (var header in extraHeaders)
            {
                request.Headers.Remove(header.Key);
                request.Headers.Add(header.Key, header.Value);
            }
        }

        if (body is not null)
        {
            var json = JsonSerializer.Serialize(body, _json);
            request.Content = new StringContent(json, Encoding.UTF8, "application/json");
        }

        using var response = await _http.SendAsync(request, cancellationToken).ConfigureAwait(false);
        var text = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
        if (!response.IsSuccessStatusCode)
        {
            throw new GameAgentEngineException($"HTTP {(int)response.StatusCode} {method} {path}", (int)response.StatusCode, path, text);
        }

        if (string.IsNullOrWhiteSpace(text))
        {
            return default;
        }

        return JsonSerializer.Deserialize<T>(text, _json);
    }

    private static string BuildQuery(params (string Key, object? Value)[] pairs)
    {
        var items = new List<string>();
        foreach (var (key, value) in pairs)
        {
            if (value is null)
            {
                continue;
            }

            var text = value switch
            {
                string s when string.IsNullOrWhiteSpace(s) => null,
                string s => s,
                _ => Convert.ToString(value, System.Globalization.CultureInfo.InvariantCulture),
            };

            if (string.IsNullOrWhiteSpace(text))
            {
                continue;
            }

            items.Add($"{Uri.EscapeDataString(key)}={Uri.EscapeDataString(text)}");
        }

        return items.Count == 0 ? string.Empty : "?" + string.Join("&", items);
    }

    private static string BuildQuery(InferenceLogQuery? query)
    {
        if (query is null)
        {
            return string.Empty;
        }

        return BuildQuery(
            ("world_id", query.WorldId),
            ("node_id", query.NodeId),
            ("task_type", query.TaskType),
            ("category", query.Category),
            ("event_name", query.EventName),
            ("execution_mode", query.ExecutionMode),
            ("request_id", query.RequestId),
            ("round", query.Round),
            ("limit", query.Limit),
            ("offset", query.Offset));
    }
}

internal static class JsonDefaults
{
    public static JsonSerializerOptions Create() => new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,
        WriteIndented = false,
    };
}

public sealed class GameAgentEngineException : Exception
{
    public GameAgentEngineException(string message, int statusCode, string path, string? responseText) : base(message)
    {
        StatusCode = statusCode;
        Path = path;
        ResponseText = responseText;
    }

    public int StatusCode { get; }
    public string Path { get; }
    public string? ResponseText { get; }
}

