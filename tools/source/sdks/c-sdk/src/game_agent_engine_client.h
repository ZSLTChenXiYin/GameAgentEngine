#ifndef GAME_AGENT_ENGINE_CLIENT_H
#define GAME_AGENT_ENGINE_CLIENT_H

const char* gae_health_path(void);
const char* gae_version_path(void);
const char* gae_invoke_path(void);
const char* gae_interpret_player_input_path(void);
const char* gae_advance_tick_path(const char* world_id);
const char* gae_world_settings_path(const char* world_id);
const char* gae_state_components_path(const char* world_id);
const char* gae_state_component_path(const char* world_id, const char* component_type);
const char* gae_timelines_path(const char* world_id, int limit);
const char* gae_latest_timeline_path(const char* world_id);
const char* gae_logs_path(const char* world_id, int limit, int offset, const char* task_type);
const char* gae_debug_traces_path(const char* world_id, int limit);
const char* gae_world_policy_path(const char* world_id);
const char* gae_pending_tasks_path(const char* consumer, int limit);
const char* gae_runtime_tasks_path(const char* category, const char* status, int limit);
const char* gae_runtime_task_path(const char* task_id);
const char* gae_claim_runtime_task_payload(const char* task_id, const char* consumer, const char* owner);
const char* gae_start_runtime_task_payload(const char* task_id, const char* lease_token);
const char* gae_heartbeat_runtime_task_payload(const char* task_id, const char* lease_token);
const char* gae_release_runtime_task_payload(const char* task_id, const char* lease_token, const char* error_message);
const char* gae_requeue_runtime_task_payload(const char* task_id, int retry_delay_ms, const char* error_message);
const char* gae_runtime_task_stats_path(void);
const char* gae_callback_payload(const char* callback_id, const char* status, const char* result_json);

#endif
