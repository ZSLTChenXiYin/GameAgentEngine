public class HealthExample {
    public static void main(String[] args) throws Exception {
        GameAgentEngineClient client = new GameAgentEngineClient("http://127.0.0.1:8080", "dev-key");
        System.out.println(client.health());
    }
}
