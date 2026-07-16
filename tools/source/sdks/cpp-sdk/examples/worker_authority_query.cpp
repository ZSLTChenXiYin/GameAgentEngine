#include <fstream>
#include <iostream>
#include <sstream>
#include "../src/game_agent_engine_client.hpp"

static std::string read_text(const std::string& path) {
    std::ifstream in(path);
    std::ostringstream buffer;
    buffer << in.rdbuf();
    return buffer.str();
}

int main() {
    GameAgentEngineClient client("http://127.0.0.1:8080", "dev-key");
    const auto dynamic_interfaces = read_text("tools/source/tests/runtime_task_dynamic_interfaces.json");
    const std::string invoke_body =
        std::string("{\"world_id\":\"demo_world\",\"node_id\":\"innkeeper_001\",\"task_type\":\"npc_dialogue\",\"messages\":[{\"role\":\"user\",\"content\":\"Before answering, query the nearby scene state and then respond.\"}],\"context\":{\"pipeline_mode\":\"full\",\"dynamic_interfaces\":")
        + dynamic_interfaces + "}}";
    const auto request = client.invokeRequest(invoke_body);
    std::cout << request.method << " " << request.path << std::endl;
    std::cout << request.body << std::endl;
    std::cout << "Execute the request in your host transport, then inspect callback_id in the response." << std::endl;
    std::cout << client.runtimeTasksRequest("", "", 20).path << std::endl;
    std::cout << "After the Engine emits a runtime task, process it with GameAgentWorker pull-once --consumer game_client" << std::endl;
    return 0;
}
