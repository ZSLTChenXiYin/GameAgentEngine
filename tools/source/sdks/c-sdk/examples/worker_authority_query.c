#include <stdio.h>
#include "../src/game_agent_engine_client.h"

int main(void) {
    printf("POST %s\n", gae_invoke_path());
    printf("load dynamic interfaces from tools/source/workerhome/fixtures/runtime_task_dynamic_interfaces.json in your host transport\n");
    printf("then inspect callback_id from the response and list tasks via %s\n", gae_runtime_tasks_path("", "", 20));
    printf("next step: GameAgentWorker pull-once --consumer game_client\n");
    return 0;
}
