#include <iostream>
#include "../src/game_agent_engine_client.hpp"

int main() {
    GameAgentEngineClient client("http://127.0.0.1:8080", "dev-key");
    std::cout << client.pendingTasksRequest("game_client", 1).path << std::endl;
    std::cout << client.claimRuntimeTaskRequest("task-1", "game_client", "cpp-sdk-example").body << std::endl;
    std::cout << client.startRuntimeTaskRequest("task-1", "lease-token").body << std::endl;
    std::cout << client.heartbeatRuntimeTaskRequest("task-1", "lease-token").body << std::endl;
    std::cout << client.actionCallbackRequest("callback-1", "success", "{\"worker\":\"cpp-sdk-example\"}").body << std::endl;
    std::cout << client.releaseRuntimeTaskRequest("task-1", "lease-token", "manual release").body << std::endl;
    std::cout << client.requeueRuntimeTaskRequest("task-1", 1500, "manual requeue").body << std::endl;
    std::cout << client.runtimeTaskStatsRequest().path << std::endl;
    return 0;
}
