# Play Mode

`GameAgentWorker play` provides a local text-game shell for developer-side engine experience and integration validation.

Current scope:

- one player actor controlled from CLI
- direct dialogue with `/talk <npc>` plus plain text input
- structured commands such as `/gift`, `/show_item`, `/trade`, `/threaten`
- first-pass room chat with `/say` and `/ask`
- authoritative HP / inventory / money / quest / scene state loaded from a YAML/JSON state file
- embedded pull worker support so Engine `game_client_request_data` callbacks can resolve during play

## Quick Start

Start Engine first, then import the demo world:

```bash
GameAgentEngine serve
GameAgentDevCli import tools/source/demo-world.yaml
```

Open play mode against the matching authority state:

```bash
GameAgentWorker play --state-file tools/source/demo-state.yaml --world-id demo_world --player-node-id player_001
```

## Recommended First Commands

```text
/look
/who
/talk "Innkeeper Mara"
老板，今晚有人见过这把刀的主人吗？
/show_item "Innkeeper Mara" "Bloody Short Knife"
/gift "Innkeeper Mara" "Silver Ring"
/room
/say 今晚谁最后一个从码头回来？
/ask "Guard Han" 你刚才守的是哪扇门？
```

## Design Boundaries

- high-frequency truth remains game-side / worker-side authority state
- play mode is not a raw engine shell; it is a constrained text-game façade
- group chat still selects one primary responder per turn
- state-changing commands should be structured and authority-validated before invoking Engine
