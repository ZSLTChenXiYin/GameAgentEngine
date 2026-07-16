#include <iostream>
#include "../src/game_agent_engine_client.hpp"

int main() {
    GameAgentEngineClient client("http://127.0.0.1:8080", "dev-key");
    std::cout << client.pendingTasksPath("game_client", 1) << std::endl;
    std::cout << client.claimRuntimeTaskPayload("task-1", "game_client", "cpp-sdk-example") << std::endl;
    std::cout << client.startRuntimeTaskPayload("task-1", "lease-token") << std::endl;
    return 0;
}

