import java.nio.file.Files;
import java.nio.file.Path;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class WorkerAuthorityQueryExample {
    public static void main(String[] args) throws Exception {
        var client = new GameAgentEngineClient(
            System.getenv().getOrDefault("GAE_SERVER", "http://127.0.0.1:8080"),
            System.getenv().getOrDefault("GAE_KEY", "dev-key")
        );

        var dynamicInterfacesFile = System.getenv().getOrDefault(
            "GAE_DYNAMIC_INTERFACES_FILE",
            "tools/source/tests/runtime_task_dynamic_interfaces.json"
        );
        var dynamicInterfaces = Files.readString(Path.of(dynamicInterfacesFile));

        var worldId = System.getenv().getOrDefault("GAE_WORLD_ID", "demo_world");
        var nodeId = System.getenv().getOrDefault("GAE_NODE_ID", "innkeeper_001");
        var taskType = System.getenv().getOrDefault("GAE_TASK_TYPE", "npc_dialogue");
        var pipelineMode = System.getenv().getOrDefault("GAE_PIPELINE_MODE", "full");
        var message = System.getenv().getOrDefault(
            "GAE_MESSAGE",
            "Before answering, query the nearby scene state and then respond."
        );

        var requestJson = "{" +
            "\"world_id\":\"" + jsonEscape(worldId) + "\"," +
            "\"node_id\":\"" + jsonEscape(nodeId) + "\"," +
            "\"task_type\":\"" + jsonEscape(taskType) + "\"," +
            "\"messages\":[{" +
                "\"role\":\"user\"," +
                "\"content\":\"" + jsonEscape(message) + "\"" +
            "}]," +
            "\"context\":{" +
                "\"pipeline_mode\":\"" + jsonEscape(pipelineMode) + "\"," +
                "\"dynamic_interfaces\":" + dynamicInterfaces +
            "}" +
        "}";

        var response = client.invoke(requestJson);
        System.out.println(response);

        var callbackId = findFirstField(response, "callback_id");
        if (callbackId == null || callbackId.isBlank()) {
            System.out.println("No async data request was emitted.");
            return;
        }

        var tasksJson = client.listRuntimeTasks(null, null, 20);
        System.out.println(tasksJson);
        System.out.println("Find the runtime task with callback_id=" + callbackId);
        System.out.println("Next step: GameAgentWorker pull-once --consumer game_client");
    }

    private static String findFirstField(String json, String fieldName) {
        var pattern = Pattern.compile("\\\"" + Pattern.quote(fieldName) + "\\\"\\s*:\\s*\\\"([^\\\"]+)\\\"");
        Matcher matcher = pattern.matcher(json);
        if (matcher.find()) {
            return matcher.group(1);
        }
        return null;
    }

    private static String jsonEscape(String value) {
        return value
            .replace("\\", "\\\\")
            .replace("\"", "\\\"")
            .replace("\r", "\\r")
            .replace("\n", "\\n");
    }
}
