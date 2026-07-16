#pragma once

#include <string>

struct GameAgentEngineRequest {
    std::string method;
    std::string path;
    std::string body;
};

class GameAgentEngineClient {
public:
    GameAgentEngineClient(std::string baseUrl, std::string apiKey);
    std::string healthPath() const;
    std::string versionPath() const;
    std::string invokePath() const;
    std::string pendingTasksPath(const std::string& consumer, int limit = 20) const;
    std::string runtimeTasksPath(const std::string& category = "", const std::string& status = "", int limit = 20) const;
    std::string runtimeTaskPath(const std::string& taskId) const;
    std::string claimRuntimeTaskPayload(const std::string& taskId, const std::string& consumer, const std::string& owner) const;
    std::string startRuntimeTaskPayload(const std::string& taskId, const std::string& leaseToken) const;
    std::string callbackPayload(const std::string& callbackId, const std::string& status, const std::string& resultJson) const;
    GameAgentEngineRequest healthRequest() const;
    GameAgentEngineRequest invokeRequest(const std::string& bodyJson) const;
    GameAgentEngineRequest pendingTasksRequest(const std::string& consumer, int limit = 20) const;
    GameAgentEngineRequest claimRuntimeTaskRequest(const std::string& taskId, const std::string& consumer, const std::string& owner) const;
    GameAgentEngineRequest startRuntimeTaskRequest(const std::string& taskId, const std::string& leaseToken) const;
    GameAgentEngineRequest actionCallbackRequest(const std::string& callbackId, const std::string& status, const std::string& resultJson) const;
    const std::string& baseUrl() const;
    const std::string& apiKey() const;
private:
    std::string base_url_;
    std::string api_key_;
};
