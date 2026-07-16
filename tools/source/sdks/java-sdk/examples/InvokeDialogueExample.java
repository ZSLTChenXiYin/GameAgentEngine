public class InvokeDialogueExample {
    public static void main(String[] args) throws Exception {
        var client = new GameAgentEngineClient(
            System.getenv().getOrDefault("GAE_SERVER", "http://127.0.0.1:8080"),
            System.getenv().getOrDefault("GAE_KEY", "dev-key")
        );
        var requestJson = "{" +
            "\"world_id\":\"demo_world\"," +
            "\"node_id\":\"innkeeper_001\"," +
            "\"task_type\":\"npc_dialogue\"," +
            "\"messages\":[{" +
                "\"role\":\"user\"," +
                "\"content\":\"What happened at the south gate?\"" +
            "}]" +
        "}";
        System.out.println(client.invoke(requestJson));
    }
}
