import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class WorkerRuntimeRoundtripExample {
    public static void main(String[] args) throws Exception {
        var client = new GameAgentEngineClient(
            System.getenv().getOrDefault("GAE_SERVER", "http://127.0.0.1:8080"),
            System.getenv().getOrDefault("GAE_KEY", "dev-key")
        );

        var consumer = System.getenv().getOrDefault("GAE_CONSUMER", "game_client");
        var owner = System.getenv().getOrDefault("GAE_OWNER", "java-sdk-roundtrip");
        var callbackStatus = System.getenv().getOrDefault("GAE_CALLBACK_STATUS", "success");

        var pendingJson = client.listPendingRuntimeTasks(consumer, 1);
        var taskId = findFirstField(pendingJson, "task_id");
        if (taskId == null || taskId.isBlank()) {
            System.out.println("No pending runtime task for consumer=" + consumer + ".");
            return;
        }

        var interfaceName = defaultString(findFirstField(pendingJson, "interface_name"), "unknown_interface");
        System.out.println("Claiming task " + taskId + " (" + interfaceName + ")");

        var claimedJson = client.claimRuntimeTask(taskId, consumer, owner);
        var leaseToken = findFirstField(claimedJson, "lease_token");
        if (leaseToken == null || leaseToken.isBlank()) {
            throw new IllegalStateException("Task " + taskId + " missing lease token after claim.");
        }

        var startedJson = client.startRuntimeTask(taskId, leaseToken);
        System.out.println(startedJson);

        var callbackId = System.getenv().getOrDefault(
            "GAE_CALLBACK_ID",
            defaultString(findFirstField(startedJson, "callback_id"), findFirstField(pendingJson, "callback_id"))
        );
        if (callbackId == null || callbackId.isBlank()) {
            throw new IllegalStateException("Task " + taskId + " missing callback_id.");
        }

        var callbackRequestId = System.getenv().getOrDefault("GAE_CALLBACK_REQUEST_ID", "java-sdk-" + taskId);
        var resultJson = "{" +
            "\"worker\":\"java-sdk-example\"," +
            "\"source\":\"worker_runtime_roundtrip\"," +
            "\"interface_name\":\"" + jsonEscape(interfaceName) + "\"," +
            "\"task_id\":\"" + jsonEscape(taskId) + "\"," +
            "\"consumer\":\"" + jsonEscape(consumer) + "\"" +
        "}";

        var callbackJson = client.actionCallback(callbackId, callbackStatus, resultJson, callbackRequestId);
        System.out.println(callbackJson);
    }

    private static String findFirstField(String json, String fieldName) {
        var pattern = Pattern.compile("\\\"" + Pattern.quote(fieldName) + "\\\"\\s*:\\s*\\\"([^\\\"]+)\\\"");
        Matcher matcher = pattern.matcher(json);
        if (matcher.find()) {
            return matcher.group(1);
        }
        return null;
    }

    private static String defaultString(String value, String fallback) {
        if (value == null || value.isBlank()) {
            return fallback;
        }
        return value;
    }

    private static String jsonEscape(String value) {
        return value
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("\r", "\\r")
            .replace("\n", "\\n");
    }
}
