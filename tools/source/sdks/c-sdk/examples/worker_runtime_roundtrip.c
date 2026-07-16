#include <stdio.h>
#include "../src/game_agent_engine_client.h"

int main(void) {
    printf("%s\n", gae_pending_tasks_path("game_client", 1));
    printf("%s\n", gae_claim_runtime_task_payload("task-1", "game_client", "c-sdk-example"));
    printf("%s\n", gae_start_runtime_task_payload("task-1", "lease-token"));
    printf("%s\n", gae_heartbeat_runtime_task_payload("task-1", "lease-token"));
    printf("%s\n", gae_callback_payload("callback-1", "success", "{\"worker\":\"c-sdk-example\"}"));
    printf("%s\n", gae_release_runtime_task_payload("task-1", "lease-token", "manual release"));
    printf("%s\n", gae_requeue_runtime_task_payload("task-1", 1500, "manual requeue"));
    printf("%s\n", gae_runtime_task_stats_path());
    return 0;
}
