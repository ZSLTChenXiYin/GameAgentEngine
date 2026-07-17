public class ContinuityInspectExample {
    public static void main(String[] args) throws Exception {
        var client = new GameAgentEngineClient(
            System.getenv().getOrDefault("GAE_SERVER", "http://127.0.0.1:8080"),
            System.getenv().getOrDefault("GAE_KEY", "dev-key")
        );

        var worldId = System.getenv().getOrDefault("GAE_WORLD_ID", "demo_world");
        var logLimit = Integer.parseInt(System.getenv().getOrDefault("GAE_LOG_LIMIT", "10"));
        var traceLimit = Integer.parseInt(System.getenv().getOrDefault("GAE_TRACE_LIMIT", "10"));
        var timelineLimit = Integer.parseInt(System.getenv().getOrDefault("GAE_TIMELINE_LIMIT", "5"));

        System.out.println("== world settings ==");
        System.out.println(client.getWorldSettings(worldId));

        System.out.println("== state components ==");
        System.out.println(client.getStateComponents(worldId));

        System.out.println("== latest timeline ==");
        System.out.println(client.getLatestTimeline(worldId));

        System.out.println("== recent timelines ==");
        System.out.println(client.getTimelines(worldId, timelineLimit));

        System.out.println("== recent logs ==");
        System.out.println(client.getLogs(worldId, logLimit, 0, null));

        System.out.println("== recent debug traces ==");
        System.out.println(client.getDebugTraces(worldId, traceLimit));
    }
}
