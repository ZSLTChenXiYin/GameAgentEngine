public class InvokeDialogueExample {
    public static void main(String[] args) {
        GameAgentEngineClient client = new GameAgentEngineClient("http://127.0.0.1:8080", "dev-key");
        System.out.println(client.invokePath());
        System.out.println("{\"world_id\":\"demo_world\",\"node_id\":\"innkeeper_001\",\"task_type\":\"npc_dialogue\"}");
    }
}
