#include <iostream>
#include "../src/game_agent_engine_client.hpp"

int main() {
    GameAgentEngineClient client("http://127.0.0.1:8080", "dev-key");
    std::cout << client.healthPath() << std::endl;
    return 0;
}
