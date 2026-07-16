extends Node

func _ready() -> void:
	var client := GameAgentEngineClient.new("http://127.0.0.1:8080", "dev-key")
	print(client.invoke_request({
		"world_id": "demo_world",
		"node_id": "innkeeper_001",
		"task_type": "npc_dialogue",
		"messages": [{"role": "user", "content": "Did anyone cause trouble tonight?"}]
	}))
