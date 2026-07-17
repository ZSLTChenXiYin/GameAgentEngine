#include <stdio.h>
#include "../src/game_agent_engine_client.h"

int main(void) {
    const char* world_id = "demo_world";
    printf("== world settings ==\nGET %s\n", gae_world_settings_path(world_id));
    printf("== state components ==\nGET %s\n", gae_state_components_path(world_id));
    printf("== latest timeline ==\nGET %s\n", gae_latest_timeline_path(world_id));
    printf("== recent timelines ==\nGET %s\n", gae_timelines_path(world_id, 5));
    printf("== recent logs ==\nGET %s\n", gae_logs_path(world_id, 10, 0, NULL));
    printf("== recent debug traces ==\nGET %s\n", gae_debug_traces_path(world_id, 10));
    printf("== world policy ==\nGET %s\n", gae_world_policy_path(world_id));
    return 0;
}
