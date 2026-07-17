#include <iostream>
#include "../src/game_agent_engine_client.hpp"

int main() {
    GameAgentEngineClient client("http://127.0.0.1:8080", "dev-key");
    const std::string world_id = "demo_world";

    std::cout << client.advanceTickRequest(world_id, "{}").method << " " << client.advanceTickPath(world_id) << std::endl;
    std::cout << client.worldSettingsRequest(world_id).path << std::endl;
    std::cout << client.setWorldSettingsRequest(world_id, "{\"tick_interval_ms\":500}").body << std::endl;
    std::cout << client.stateComponentsRequest(world_id).path << std::endl;
    std::cout << client.stateComponentRequest(world_id, "world_focus").path << std::endl;
    std::cout << client.putStateComponentRequest(world_id, "world_focus", "{\"enabled\":true}").body << std::endl;
    std::cout << client.timelinesRequest(world_id, 10).path << std::endl;
    std::cout << client.latestTimelineRequest(world_id).path << std::endl;
    std::cout << client.logsRequest(world_id, 10, 0, "tick").path << std::endl;
    std::cout << client.debugTracesRequest(world_id, 10).path << std::endl;
    std::cout << client.worldPolicyRequest(world_id).path << std::endl;
    std::cout << client.setWorldPolicyRequest(world_id, "{\"mode\":\"balanced\"}").body << std::endl;
    return 0;
}
