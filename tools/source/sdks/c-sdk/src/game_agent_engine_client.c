#include "game_agent_engine_client.h"

#include <stdio.h>

static char gae_buffer[1024];

const char* gae_health_path(void) { return "/health"; }
const char* gae_version_path(void) { return "/api/v1/version"; }
const char* gae_invoke_path(void) { return "/api/v1/invoke"; }
const char* gae_interpret_player_input_path(void) { return "/api/v1/player/input/interpret"; }
const char* gae_advance_tick_path(const char* world_id) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/worlds/%s/ticks/advance", world_id);
    return gae_buffer;
}
const char* gae_world_settings_path(const char* world_id) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/worlds/%s/settings", world_id);
    return gae_buffer;
}
const char* gae_state_components_path(const char* world_id) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/worlds/%s/state-components", world_id);
    return gae_buffer;
}
const char* gae_state_component_path(const char* world_id, const char* component_type) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/worlds/%s/state-components/%s", world_id, component_type);
    return gae_buffer;
}
const char* gae_timelines_path(const char* world_id, int limit) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/worlds/%s/timelines?limit=%d", world_id, limit);
    return gae_buffer;
}
const char* gae_latest_timeline_path(const char* world_id) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/worlds/%s/timelines/latest", world_id);
    return gae_buffer;
}
const char* gae_logs_path(const char* world_id, int limit, int offset, const char* task_type) {
    snprintf(
        gae_buffer,
        sizeof(gae_buffer),
        "/api/v1/logs?world_id=%s&limit=%d&offset=%d%s%s",
        world_id,
        limit,
        offset,
        (task_type && task_type[0]) ? "&task_type=" : "",
        (task_type && task_type[0]) ? task_type : ""
    );
    return gae_buffer;
}
const char* gae_debug_traces_path(const char* world_id, int limit) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/debug/traces?world_id=%s&limit=%d", world_id, limit);
    return gae_buffer;
}
const char* gae_world_policy_path(const char* world_id) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/worlds/%s/policy", world_id);
    return gae_buffer;
}
const char* gae_pending_tasks_path(const char* consumer, int limit) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/runtime/tasks/pending?consumer=%s&limit=%d", consumer, limit);
    return gae_buffer;
}
const char* gae_runtime_tasks_path(const char* category, const char* status, int limit) {
    snprintf(
        gae_buffer,
        sizeof(gae_buffer),
        "/api/v1/runtime/tasks?limit=%d%s%s%s%s",
        limit,
        (category && category[0]) ? "&category=" : "",
        (category && category[0]) ? category : "",
        (status && status[0]) ? "&status=" : "",
        (status && status[0]) ? status : ""
    );
    return gae_buffer;
}
const char* gae_runtime_task_path(const char* task_id) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/runtime/tasks/%s", task_id);
    return gae_buffer;
}
const char* gae_claim_runtime_task_payload(const char* task_id, const char* consumer, const char* owner) {
    snprintf(gae_buffer, sizeof(gae_buffer), "{\"task_id\":\"%s\",\"consumer\":\"%s\",\"lease_owner\":\"%s\"}", task_id, consumer, owner);
    return gae_buffer;
}
const char* gae_start_runtime_task_payload(const char* task_id, const char* lease_token) {
    snprintf(gae_buffer, sizeof(gae_buffer), "{\"task_id\":\"%s\",\"lease_token\":\"%s\"}", task_id, lease_token);
    return gae_buffer;
}
const char* gae_heartbeat_runtime_task_payload(const char* task_id, const char* lease_token) {
    snprintf(gae_buffer, sizeof(gae_buffer), "{\"task_id\":\"%s\",\"lease_token\":\"%s\"}", task_id, lease_token);
    return gae_buffer;
}
const char* gae_release_runtime_task_payload(const char* task_id, const char* lease_token, const char* error_message) {
    snprintf(gae_buffer, sizeof(gae_buffer), "{\"task_id\":\"%s\",\"lease_token\":\"%s\",\"error_message\":\"%s\"}", task_id, lease_token, error_message ? error_message : "");
    return gae_buffer;
}
const char* gae_requeue_runtime_task_payload(const char* task_id, int retry_delay_ms, const char* error_message) {
    snprintf(gae_buffer, sizeof(gae_buffer), "{\"task_id\":\"%s\",\"retry_delay_ms\":%d,\"error_message\":\"%s\"}", task_id, retry_delay_ms, error_message ? error_message : "");
    return gae_buffer;
}
const char* gae_runtime_task_stats_path(void) {
    return "/api/v1/runtime/tasks/stats";
}
const char* gae_callback_payload(const char* callback_id, const char* status, const char* result_json) {
    snprintf(gae_buffer, sizeof(gae_buffer), "{\"callback_id\":\"%s\",\"status\":\"%s\",\"result\":%s}", callback_id, status, result_json);
    return gae_buffer;
}
