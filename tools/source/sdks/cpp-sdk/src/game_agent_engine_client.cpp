#include "game_agent_engine_client.hpp"

GameAgentEngineClient::GameAgentEngineClient(std::string baseUrl, std::string apiKey)
    : base_url_(std::move(baseUrl)), api_key_(std::move(apiKey)) {}

std::string GameAgentEngineClient::healthPath() const { return base_url_ + "/health"; }
std::string GameAgentEngineClient::versionPath() const { return base_url_ + "/api/v1/version"; }
std::string GameAgentEngineClient::invokePath() const { return base_url_ + "/api/v1/invoke"; }
std::string GameAgentEngineClient::pendingTasksPath(const std::string& consumer, int limit) const {
    return base_url_ + "/api/v1/runtime/tasks/pending?consumer=" + consumer + "&limit=" + std::to_string(limit);
}
