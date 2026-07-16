import java.net.URI;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class GameAgentEngineClient {
    private final String baseUrl;
    private final String apiKey;
    private final HttpClient httpClient = HttpClient.newHttpClient();

    public GameAgentEngineClient(String baseUrl, String apiKey) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
    }

    public String get(String path) throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(baseUrl + path))
            .header("X-API-Key", apiKey)
            .header("Accept", "application/json")
            .GET()
            .build();
        return httpClient.send(request, HttpResponse.BodyHandlers.ofString()).body();
    }

    public String post(String path, String jsonBody) throws Exception {
        HttpRequest request = HttpRequest.newBuilder()
            .uri(URI.create(baseUrl + path))
            .header("X-API-Key", apiKey)
            .header("Accept", "application/json")
            .header("Content-Type", "application/json")
            .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
            .build();
        return httpClient.send(request, HttpResponse.BodyHandlers.ofString()).body();
    }

    public String health() throws Exception { return get("/health"); }
    public String version() throws Exception { return get("/api/v1/version"); }
    public String invoke(String jsonBody) throws Exception { return post("/api/v1/invoke", jsonBody); }
    public String listPendingRuntimeTasks(String consumer, int limit) throws Exception {
        return get("/api/v1/runtime/tasks/pending?consumer=" + URLEncoder.encode(consumer, StandardCharsets.UTF_8) + "&limit=" + limit);
    }
    public String claimRuntimeTask(String taskId, String consumer, String owner) throws Exception {
        return post("/api/v1/runtime/tasks/claim", "{\"task_id\":\"" + taskId + "\",\"consumer\":\"" + consumer + "\",\"lease_owner\":\"" + owner + "\"}");
    }
    public String startRuntimeTask(String taskId, String leaseToken) throws Exception {
        return post("/api/v1/runtime/tasks/start", "{\"task_id\":\"" + taskId + "\",\"lease_token\":\"" + leaseToken + "\"}");
    }
    public String actionCallback(String callbackId, String status, String resultJson) throws Exception {
        return post("/api/v1/actions/callback", "{\"callback_id\":\"" + callbackId + "\",\"status\":\"" + status + "\",\"result\":" + resultJson + "}");
    }
}
