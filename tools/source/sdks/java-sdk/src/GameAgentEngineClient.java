import java.net.URI;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.ArrayList;
import java.util.List;

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
        return post(path, jsonBody, null);
    }

    public String post(String path, String jsonBody, String callbackRequestId) throws Exception {
        var builder = HttpRequest.newBuilder()
            .uri(URI.create(baseUrl + path))
            .header("X-API-Key", apiKey)
            .header("Accept", "application/json")
            .header("Content-Type", "application/json");
        if (callbackRequestId != null && !callbackRequestId.isBlank()) {
            builder.header("X-Callback-Request-Id", callbackRequestId);
        }
        HttpRequest request = builder
            .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
            .build();
        return httpClient.send(request, HttpResponse.BodyHandlers.ofString()).body();
    }

    public String postOrPut(String method, String path, String jsonBody) throws Exception {
        var builder = HttpRequest.newBuilder()
            .uri(URI.create(baseUrl + path))
            .header("X-API-Key", apiKey)
            .header("Accept", "application/json")
            .header("Content-Type", "application/json");
        HttpRequest request = builder
            .method(method, HttpRequest.BodyPublishers.ofString(jsonBody))
            .build();
        return httpClient.send(request, HttpResponse.BodyHandlers.ofString()).body();
    }

    private String buildQuery(String[][] pairs) {
        List<String> items = new ArrayList<>();
        for (String[] pair : pairs) {
            if (pair == null || pair.length < 2) {
                continue;
            }
            var key = pair[0];
            var value = pair[1];
            if (key == null || value == null || value.isBlank()) {
                continue;
            }
            items.add(URLEncoder.encode(key, StandardCharsets.UTF_8) + "=" + URLEncoder.encode(value, StandardCharsets.UTF_8));
        }
        if (items.isEmpty()) {
            return "";
        }
        return "?" + String.join("&", items);
    }

    public String health() throws Exception { return get("/health"); }
    public String version() throws Exception { return get("/api/v1/version"); }
    public String invoke(String jsonBody) throws Exception { return post("/api/v1/invoke", jsonBody); }
    public String interpretPlayerInput(String jsonBody) throws Exception { return post("/api/v1/player/input/interpret", jsonBody); }
    public String advanceTick(String worldId, String jsonBody) throws Exception {
        return post("/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/ticks/advance", jsonBody);
    }
    public String getWorldSettings(String worldId) throws Exception {
        return get("/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/settings");
    }
    public String setWorldSettings(String worldId, String jsonBody) throws Exception {
        return postOrPut("PUT", "/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/settings", jsonBody);
    }
    public String getStateComponents(String worldId) throws Exception {
        return get("/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/state-components");
    }
    public String getStateComponent(String worldId, String componentType) throws Exception {
        return get("/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/state-components/" + URLEncoder.encode(componentType, StandardCharsets.UTF_8));
    }
    public String putStateComponent(String worldId, String componentType, String jsonBody) throws Exception {
        return postOrPut("PUT", "/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/state-components/" + URLEncoder.encode(componentType, StandardCharsets.UTF_8), jsonBody);
    }
    public String getTimelines(String worldId, int limit) throws Exception {
        return get("/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/timelines" + buildQuery(new String[][]{
            {"limit", limit > 0 ? Integer.toString(limit) : null},
        }));
    }
    public String getLatestTimeline(String worldId) throws Exception {
        return get("/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/timelines/latest");
    }
    public String getLogs(String worldId, int limit, int offset, String taskType) throws Exception {
        return get("/api/v1/logs" + buildQuery(new String[][]{
            {"world_id", worldId},
            {"limit", limit > 0 ? Integer.toString(limit) : null},
            {"offset", offset > 0 ? Integer.toString(offset) : null},
            {"task_type", taskType},
        }));
    }
    public String getDebugTraces(String worldId, int limit) throws Exception {
        return get("/debug/traces" + buildQuery(new String[][]{
            {"world_id", worldId},
            {"limit", limit > 0 ? Integer.toString(limit) : null},
        }));
    }
    public String getWorldPolicy(String worldId) throws Exception {
        return get("/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/policy");
    }
    public String setWorldPolicy(String worldId, String jsonBody) throws Exception {
        return postOrPut("PUT", "/api/v1/worlds/" + URLEncoder.encode(worldId, StandardCharsets.UTF_8) + "/policy", jsonBody);
    }
    public String listPendingRuntimeTasks(String consumer, int limit) throws Exception {
        return get("/api/v1/runtime/tasks/pending?consumer=" + URLEncoder.encode(consumer, StandardCharsets.UTF_8) + "&limit=" + limit);
    }
    public String listRuntimeTasks(String category, String status, int limit) throws Exception {
        return get("/api/v1/runtime/tasks" + buildQuery(new String[][]{
            {"category", category},
            {"status", status},
            {"limit", Integer.toString(limit)},
        }));
    }
    public String getRuntimeTask(String taskId) throws Exception {
        return get("/api/v1/runtime/tasks/" + URLEncoder.encode(taskId, StandardCharsets.UTF_8));
    }
    public String claimRuntimeTask(String taskId, String consumer, String owner) throws Exception {
        return post("/api/v1/runtime/tasks/claim", "{\"task_id\":\"" + taskId + "\",\"consumer\":\"" + consumer + "\",\"lease_owner\":\"" + owner + "\"}");
    }
    public String startRuntimeTask(String taskId, String leaseToken) throws Exception {
        return post("/api/v1/runtime/tasks/start", "{\"task_id\":\"" + taskId + "\",\"lease_token\":\"" + leaseToken + "\"}");
    }
    public String heartbeatRuntimeTask(String taskId, String leaseToken) throws Exception {
        return post("/api/v1/runtime/tasks/heartbeat", "{\"task_id\":\"" + taskId + "\",\"lease_token\":\"" + leaseToken + "\"}");
    }
    public String releaseRuntimeTask(String taskId, String leaseToken, String errorMessage) throws Exception {
        return post("/api/v1/runtime/tasks/release", "{\"task_id\":\"" + taskId + "\",\"lease_token\":\"" + leaseToken + "\",\"error_message\":\"" + jsonEscape(errorMessage) + "\"}");
    }
    public String requeueRuntimeTask(String taskId, int retryDelayMs, String errorMessage) throws Exception {
        return post("/api/v1/runtime/tasks/requeue", "{\"task_id\":\"" + taskId + "\",\"retry_delay_ms\":" + retryDelayMs + ",\"error_message\":\"" + jsonEscape(errorMessage) + "\"}");
    }
    public String getRuntimeTaskStats() throws Exception {
        return get("/api/v1/runtime/tasks/stats");
    }
    public String actionCallback(String callbackId, String status, String resultJson) throws Exception {
        return actionCallback(callbackId, status, resultJson, null);
    }
    public String actionCallback(String callbackId, String status, String resultJson, String callbackRequestId) throws Exception {
        return post("/api/v1/actions/callback", "{\"callback_id\":\"" + callbackId + "\",\"status\":\"" + status + "\",\"result\":" + resultJson + "}", callbackRequestId);
    }

    private String jsonEscape(String value) {
        if (value == null) {
            return "";
        }
        return value
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("\r", "\\r")
            .replace("\n", "\\n");
    }
}
