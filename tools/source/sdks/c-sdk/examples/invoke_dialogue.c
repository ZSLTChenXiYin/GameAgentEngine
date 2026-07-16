#include <stdio.h>
#include "../src/game_agent_engine_client.h"

int main(void) {
    printf("%s\n", gae_invoke_path());
    printf("{\"world_id\":\"demo_world\",\"node_id\":\"innkeeper_001\",\"task_type\":\"npc_dialogue\"}\n");
    return 0;
}
