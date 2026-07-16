using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;

namespace GameAgentEngine.SDK;

public class GameAgentEngineClient
{
    private readonly HttpClient _http;
    private readonly string _baseUrl;

    public GameAgentEngineClient(string baseUrl, string apiKey, HttpClient? httpClient = null)
    {
        _baseUrl = baseUrl.TrimEnd('/');
        _http = httpClient ?? new HttpClient();
        _http.DefaultRequestHeaders.Remove("X-API-Key");
        _http.DefaultRequestHeaders.Add("X-API-Key", apiKey);
    }

    public Task<string> HealthAsync() => GetAsync("/health");
    public Task<string> GetVersionAsync() => GetAsync("/api/v1/version");
    public Task<string> InvokeAsync(object body) => PostAsync("/api/v1/invoke", body);
    public Task<string> ListPendingRuntimeTasksAsync(string consumer, int limit = 20) => GetAsync($"/api/v1/runtime/tasks/pending?consumer={consumer}&limit={limit}");
    public Task<string> ClaimRuntimeTaskAsync(string taskId, string consumer, string owner) => PostAsync("/api/v1/runtime/tasks/claim", new { task_id = taskId, consumer = consumer, lease_owner = owner });
    public Task<string> StartRuntimeTaskAsync(string taskId, string leaseToken) => PostAsync("/api/v1/runtime/tasks/start", new { task_id = taskId, lease_token = leaseToken });
    public Task<string> ActionCallbackAsync(string callbackId, string status, object result) => PostAsync("/api/v1/actions/callback", new { callback_id = callbackId, status = status, result = result });

    private async Task<string> GetAsync(string path)
    {
        var res = await _http.GetAsync(_baseUrl + path);
        var text = await res.Content.ReadAsStringAsync();
        if (!res.IsSuccessStatusCode) throw new HttpRequestException($"HTTP {(int)res.StatusCode} {path}: {text}");
        return text;
    }

    private async Task<string> PostAsync(string path, object body)
    {
        var json = JsonSerializer.Serialize(body);
        var res = await _http.PostAsync(_baseUrl + path, new StringContent(json, Encoding.UTF8, "application/json"));
        var text = await res.Content.ReadAsStringAsync();
        if (!res.IsSuccessStatusCode) throw new HttpRequestException($"HTTP {(int)res.StatusCode} {path}: {text}");
        return text;
    }
}
