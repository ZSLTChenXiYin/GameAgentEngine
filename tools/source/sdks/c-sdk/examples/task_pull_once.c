#include <stdio.h>
#include "../src/game_agent_engine_client.h"

int main(void) {
    printf("%s\n", gae_pending_tasks_path("game_client", 1));
    printf("%s\n", gae_claim_runtime_task_payload("task-1", "game_client", "c-sdk-example"));
    printf("%s\n", gae_start_runtime_task_payload("task-1", "lease-token"));
    return 0;
}

