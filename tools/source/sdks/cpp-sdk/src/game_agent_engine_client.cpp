#include "game_agent_engine_client.hpp"

#include <sstream>

GameAgentEngineClient::GameAgentEngineClient(std::string baseUrl, std::string apiKey)
    : base_url_(std::move(baseUrl)), api_key_(std::move(apiKey)) {}

std::string GameAgentEngineClient::healthPath() const { return base_url_ + "/health"; }
std::string GameAgentEngineClient::versionPath() const { return base_url_ + "/api/v1/version"; }
std::string GameAgentEngineClient::invokePath() const { return base_url_ + "/api/v1/invoke"; }
std::string GameAgentEngineClient::interpretPlayerInputPath() const { return base_url_ + "/api/v1/player/input/interpret"; }
std::string GameAgentEngineClient::advanceTickPath(const std::string& worldId) const {
    return base_url_ + "/api/v1/worlds/" + worldId + "/ticks/advance";
}

std::string GameAgentEngineClient::worldSettingsPath(const std::string& worldId) const {
    return base_url_ + "/api/v1/worlds/" + worldId + "/settings";
}

std::string GameAgentEngineClient::stateComponentsPath(const std::string& worldId) const {
    return base_url_ + "/api/v1/worlds/" + worldId + "/state-components";
}

std::string GameAgentEngineClient::stateComponentPath(const std::string& worldId, const std::string& componentType) const {
    return base_url_ + "/api/v1/worlds/" + worldId + "/state-components/" + componentType;
}

std::string GameAgentEngineClient::timelinesPath(const std::string& worldId, int limit) const {
    std::ostringstream path;
    path << base_url_ << "/api/v1/worlds/" << worldId << "/timelines";
    if (limit > 0) {
        path << "?limit=" << limit;
    }
    return path.str();
}

std::string GameAgentEngineClient::latestTimelinePath(const std::string& worldId) const {
    return base_url_ + "/api/v1/worlds/" + worldId + "/timelines/latest";
}

std::string GameAgentEngineClient::logsPath(const std::string& worldId, int limit, int offset, const std::string& taskType) const {
    std::ostringstream path;
    path << base_url_ << "/api/v1/logs?world_id=" << worldId;
    if (limit > 0) path << "&limit=" << limit;
    if (offset > 0) path << "&offset=" << offset;
    if (!taskType.empty()) path << "&task_type=" << taskType;
    return path.str();
}

std::string GameAgentEngineClient::debugTracesPath(const std::string& worldId, int limit) const {
    std::ostringstream path;
    path << base_url_ << "/debug/traces?world_id=" << worldId;
    if (limit > 0) path << "&limit=" << limit;
    return path.str();
}

std::string GameAgentEngineClient::worldPolicyPath(const std::string& worldId) const {
    return base_url_ + "/api/v1/worlds/" + worldId + "/policy";
}
std::string GameAgentEngineClient::pendingTasksPath(const std::string& consumer, int limit) const {
    return base_url_ + "/api/v1/runtime/tasks/pending?consumer=" + consumer + "&limit=" + std::to_string(limit);
}

std::string GameAgentEngineClient::runtimeTasksPath(const std::string& category, const std::string& status, int limit) const {
    std::ostringstream path;
    path << base_url_ << "/api/v1/runtime/tasks?limit=" << limit;
    if (!category.empty()) path << "&category=" << category;
    if (!status.empty()) path << "&status=" << status;
    return path.str();
}

std::string GameAgentEngineClient::runtimeTaskPath(const std::string& taskId) const {
    return base_url_ + "/api/v1/runtime/tasks/" + taskId;
}

std::string GameAgentEngineClient::claimRuntimeTaskPayload(const std::string& taskId, const std::string& consumer, const std::string& owner) const {
    return std::string("{\"task_id\":\"") + taskId + "\",\"consumer\":\"" + consumer + "\",\"lease_owner\":\"" + owner + "\"}";
}

std::string GameAgentEngineClient::startRuntimeTaskPayload(const std::string& taskId, const std::string& leaseToken) const {
    return std::string("{\"task_id\":\"") + taskId + "\",\"lease_token\":\"" + leaseToken + "\"}";
}

std::string GameAgentEngineClient::heartbeatRuntimeTaskPayload(const std::string& taskId, const std::string& leaseToken) const {
    return std::string("{\"task_id\":\"") + taskId + "\",\"lease_token\":\"" + leaseToken + "\"}";
}

std::string GameAgentEngineClient::releaseRuntimeTaskPayload(const std::string& taskId, const std::string& leaseToken, const std::string& errorMessage) const {
    return std::string("{\"task_id\":\"") + taskId + "\",\"lease_token\":\"" + leaseToken + "\",\"error_message\":\"" + errorMessage + "\"}";
}

std::string GameAgentEngineClient::requeueRuntimeTaskPayload(const std::string& taskId, int retryDelayMs, const std::string& errorMessage) const {
    return std::string("{\"task_id\":\"") + taskId + "\",\"retry_delay_ms\":" + std::to_string(retryDelayMs) + ",\"error_message\":\"" + errorMessage + "\"}";
}

std::string GameAgentEngineClient::callbackPayload(const std::string& callbackId, const std::string& status, const std::string& resultJson) const {
    return std::string("{\"callback_id\":\"") + callbackId + "\",\"status\":\"" + status + "\",\"result\":" + resultJson + "}";
}

GameAgentEngineRequest GameAgentEngineClient::healthRequest() const { return {"GET", "/health", ""}; }
GameAgentEngineRequest GameAgentEngineClient::invokeRequest(const std::string& bodyJson) const { return {"POST", "/api/v1/invoke", bodyJson}; }
GameAgentEngineRequest GameAgentEngineClient::interpretPlayerInputRequest(const std::string& bodyJson) const { return {"POST", "/api/v1/player/input/interpret", bodyJson}; }
GameAgentEngineRequest GameAgentEngineClient::advanceTickRequest(const std::string& worldId, const std::string& bodyJson) const { return {"POST", "/api/v1/worlds/" + worldId + "/ticks/advance", bodyJson}; }
GameAgentEngineRequest GameAgentEngineClient::worldSettingsRequest(const std::string& worldId) const { return {"GET", "/api/v1/worlds/" + worldId + "/settings", ""}; }
GameAgentEngineRequest GameAgentEngineClient::setWorldSettingsRequest(const std::string& worldId, const std::string& bodyJson) const { return {"PUT", "/api/v1/worlds/" + worldId + "/settings", bodyJson}; }
GameAgentEngineRequest GameAgentEngineClient::stateComponentsRequest(const std::string& worldId) const { return {"GET", "/api/v1/worlds/" + worldId + "/state-components", ""}; }
GameAgentEngineRequest GameAgentEngineClient::stateComponentRequest(const std::string& worldId, const std::string& componentType) const { return {"GET", "/api/v1/worlds/" + worldId + "/state-components/" + componentType, ""}; }
GameAgentEngineRequest GameAgentEngineClient::putStateComponentRequest(const std::string& worldId, const std::string& componentType, const std::string& bodyJson) const { return {"PUT", "/api/v1/worlds/" + worldId + "/state-components/" + componentType, bodyJson}; }
GameAgentEngineRequest GameAgentEngineClient::timelinesRequest(const std::string& worldId, int limit) const { return {"GET", timelinesPath(worldId, limit).substr(base_url_.size()), ""}; }
GameAgentEngineRequest GameAgentEngineClient::latestTimelineRequest(const std::string& worldId) const { return {"GET", "/api/v1/worlds/" + worldId + "/timelines/latest", ""}; }
GameAgentEngineRequest GameAgentEngineClient::logsRequest(const std::string& worldId, int limit, int offset, const std::string& taskType) const { return {"GET", logsPath(worldId, limit, offset, taskType).substr(base_url_.size()), ""}; }
GameAgentEngineRequest GameAgentEngineClient::debugTracesRequest(const std::string& worldId, int limit) const { return {"GET", debugTracesPath(worldId, limit).substr(base_url_.size()), ""}; }
GameAgentEngineRequest GameAgentEngineClient::worldPolicyRequest(const std::string& worldId) const { return {"GET", "/api/v1/worlds/" + worldId + "/policy", ""}; }
GameAgentEngineRequest GameAgentEngineClient::setWorldPolicyRequest(const std::string& worldId, const std::string& bodyJson) const { return {"PUT", "/api/v1/worlds/" + worldId + "/policy", bodyJson}; }
GameAgentEngineRequest GameAgentEngineClient::pendingTasksRequest(const std::string& consumer, int limit) const { return {"GET", "/api/v1/runtime/tasks/pending?consumer=" + consumer + "&limit=" + std::to_string(limit), ""}; }
GameAgentEngineRequest GameAgentEngineClient::runtimeTasksRequest(const std::string& category, const std::string& status, int limit) const { return {"GET", runtimeTasksPath(category, status, limit).substr(base_url_.size()), ""}; }
GameAgentEngineRequest GameAgentEngineClient::runtimeTaskRequest(const std::string& taskId) const { return {"GET", "/api/v1/runtime/tasks/" + taskId, ""}; }
GameAgentEngineRequest GameAgentEngineClient::claimRuntimeTaskRequest(const std::string& taskId, const std::string& consumer, const std::string& owner) const { return {"POST", "/api/v1/runtime/tasks/claim", claimRuntimeTaskPayload(taskId, consumer, owner)}; }
GameAgentEngineRequest GameAgentEngineClient::startRuntimeTaskRequest(const std::string& taskId, const std::string& leaseToken) const { return {"POST", "/api/v1/runtime/tasks/start", startRuntimeTaskPayload(taskId, leaseToken)}; }
GameAgentEngineRequest GameAgentEngineClient::heartbeatRuntimeTaskRequest(const std::string& taskId, const std::string& leaseToken) const { return {"POST", "/api/v1/runtime/tasks/heartbeat", heartbeatRuntimeTaskPayload(taskId, leaseToken)}; }
GameAgentEngineRequest GameAgentEngineClient::releaseRuntimeTaskRequest(const std::string& taskId, const std::string& leaseToken, const std::string& errorMessage) const { return {"POST", "/api/v1/runtime/tasks/release", releaseRuntimeTaskPayload(taskId, leaseToken, errorMessage)}; }
GameAgentEngineRequest GameAgentEngineClient::requeueRuntimeTaskRequest(const std::string& taskId, int retryDelayMs, const std::string& errorMessage) const { return {"POST", "/api/v1/runtime/tasks/requeue", requeueRuntimeTaskPayload(taskId, retryDelayMs, errorMessage)}; }
GameAgentEngineRequest GameAgentEngineClient::runtimeTaskStatsRequest() const { return {"GET", "/api/v1/runtime/tasks/stats", ""}; }
GameAgentEngineRequest GameAgentEngineClient::actionCallbackRequest(const std::string& callbackId, const std::string& status, const std::string& resultJson) const { return {"POST", "/api/v1/actions/callback", callbackPayload(callbackId, status, resultJson)}; }
const std::string& GameAgentEngineClient::baseUrl() const { return base_url_; }
const std::string& GameAgentEngineClient::apiKey() const { return api_key_; }
