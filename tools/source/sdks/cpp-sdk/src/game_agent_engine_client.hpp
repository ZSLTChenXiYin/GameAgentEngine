#pragma once

#include <string>

class GameAgentEngineClient {
public:
    GameAgentEngineClient(std::string baseUrl, std::string apiKey);
    std::string healthPath() const;
    std::string versionPath() const;
    std::string invokePath() const;
    std::string pendingTasksPath(const std::string& consumer, int limit = 20) const;
private:
    std::string base_url_;
    std::string api_key_;
};
