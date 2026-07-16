#include <iostream>
#include "../src/game_agent_engine_client.hpp"

int main() {
    GameAgentEngineClient client("http://127.0.0.1:8080", "dev-key");
    std::cout << client.invokePath() << std::endl;
    std::cout << R"({"world_id":"demo_world","node_id":"innkeeper_001","task_type":"npc_dialogue"})" << std::endl;
    return 0;
}
