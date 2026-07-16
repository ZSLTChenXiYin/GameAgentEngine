public class TaskPullOnceExample {
    public static void main(String[] args) throws Exception {
        var client = new GameAgentEngineClient(
            System.getenv().getOrDefault("GAE_SERVER", "http://127.0.0.1:8080"),
            System.getenv().getOrDefault("GAE_KEY", "dev-key")
        );
        var consumer = System.getenv().getOrDefault("GAE_CONSUMER", "game_client");
        System.out.println(client.listPendingRuntimeTasks(consumer, 1));
    }
}

