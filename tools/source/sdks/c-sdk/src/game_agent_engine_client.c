#include "game_agent_engine_client.h"

#include <stdio.h>

static char gae_buffer[1024];

const char* gae_health_path(void) { return "/health"; }
const char* gae_version_path(void) { return "/api/v1/version"; }
const char* gae_invoke_path(void) { return "/api/v1/invoke"; }
const char* gae_pending_tasks_path(const char* consumer, int limit) {
    snprintf(gae_buffer, sizeof(gae_buffer), "/api/v1/runtime/tasks/pending?consumer=%s&limit=%d", consumer, limit);
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
const char* gae_callback_payload(const char* callback_id, const char* status, const char* result_json) {
    snprintf(gae_buffer, sizeof(gae_buffer), "{\"callback_id\":\"%s\",\"status\":\"%s\",\"result\":%s}", callback_id, status, result_json);
    return gae_buffer;
}
