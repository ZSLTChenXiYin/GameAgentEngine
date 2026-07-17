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
    std::string interpretPlayerInputPath() const;
    std::string advanceTickPath(const std::string& worldId) const;
    std::string worldSettingsPath(const std::string& worldId) const;
    std::string stateComponentsPath(const std::string& worldId) const;
    std::string stateComponentPath(const std::string& worldId, const std::string& componentType) const;
    std::string timelinesPath(const std::string& worldId, int limit = 20) const;
    std::string latestTimelinePath(const std::string& worldId) const;
    std::string logsPath(const std::string& worldId, int limit = 20, int offset = 0, const std::string& taskType = "") const;
    std::string debugTracesPath(const std::string& worldId, int limit = 20) const;
    std::string worldPolicyPath(const std::string& worldId) const;
    std::string pendingTasksPath(const std::string& consumer, int limit = 20) const;
    std::string runtimeTasksPath(const std::string& category = "", const std::string& status = "", int limit = 20) const;
    std::string runtimeTaskPath(const std::string& taskId) const;
    std::string claimRuntimeTaskPayload(const std::string& taskId, const std::string& consumer, const std::string& owner) const;
    std::string startRuntimeTaskPayload(const std::string& taskId, const std::string& leaseToken) const;
    std::string heartbeatRuntimeTaskPayload(const std::string& taskId, const std::string& leaseToken) const;
    std::string releaseRuntimeTaskPayload(const std::string& taskId, const std::string& leaseToken, const std::string& errorMessage = "") const;
    std::string requeueRuntimeTaskPayload(const std::string& taskId, int retryDelayMs = 0, const std::string& errorMessage = "") const;
    std::string callbackPayload(const std::string& callbackId, const std::string& status, const std::string& resultJson) const;
    GameAgentEngineRequest healthRequest() const;
    GameAgentEngineRequest invokeRequest(const std::string& bodyJson) const;
    GameAgentEngineRequest interpretPlayerInputRequest(const std::string& bodyJson) const;
    GameAgentEngineRequest advanceTickRequest(const std::string& worldId, const std::string& bodyJson) const;
    GameAgentEngineRequest worldSettingsRequest(const std::string& worldId) const;
    GameAgentEngineRequest setWorldSettingsRequest(const std::string& worldId, const std::string& bodyJson) const;
    GameAgentEngineRequest stateComponentsRequest(const std::string& worldId) const;
    GameAgentEngineRequest stateComponentRequest(const std::string& worldId, const std::string& componentType) const;
    GameAgentEngineRequest putStateComponentRequest(const std::string& worldId, const std::string& componentType, const std::string& bodyJson) const;
    GameAgentEngineRequest timelinesRequest(const std::string& worldId, int limit = 20) const;
    GameAgentEngineRequest latestTimelineRequest(const std::string& worldId) const;
    GameAgentEngineRequest logsRequest(const std::string& worldId, int limit = 20, int offset = 0, const std::string& taskType = "") const;
    GameAgentEngineRequest debugTracesRequest(const std::string& worldId, int limit = 20) const;
    GameAgentEngineRequest worldPolicyRequest(const std::string& worldId) const;
    GameAgentEngineRequest setWorldPolicyRequest(const std::string& worldId, const std::string& bodyJson) const;
    GameAgentEngineRequest pendingTasksRequest(const std::string& consumer, int limit = 20) const;
    GameAgentEngineRequest runtimeTasksRequest(const std::string& category = "", const std::string& status = "", int limit = 20) const;
    GameAgentEngineRequest runtimeTaskRequest(const std::string& taskId) const;
    GameAgentEngineRequest claimRuntimeTaskRequest(const std::string& taskId, const std::string& consumer, const std::string& owner) const;
    GameAgentEngineRequest startRuntimeTaskRequest(const std::string& taskId, const std::string& leaseToken) const;
    GameAgentEngineRequest heartbeatRuntimeTaskRequest(const std::string& taskId, const std::string& leaseToken) const;
    GameAgentEngineRequest releaseRuntimeTaskRequest(const std::string& taskId, const std::string& leaseToken, const std::string& errorMessage = "") const;
    GameAgentEngineRequest requeueRuntimeTaskRequest(const std::string& taskId, int retryDelayMs = 0, const std::string& errorMessage = "") const;
    GameAgentEngineRequest runtimeTaskStatsRequest() const;
    GameAgentEngineRequest actionCallbackRequest(const std::string& callbackId, const std::string& status, const std::string& resultJson) const;
    const std::string& baseUrl() const;
    const std::string& apiKey() const;
private:
    std::string base_url_;
    std::string api_key_;
};
