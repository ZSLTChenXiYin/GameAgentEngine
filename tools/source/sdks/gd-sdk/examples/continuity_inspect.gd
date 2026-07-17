extends Node

func _ready() -> void:
	var client := GameAgentEngineClient.new("http://127.0.0.1:8080", "dev-key")
	var world_id := "demo_world"
	print(client.world_settings_request(world_id))
	print(client.state_components_request(world_id))
	print(client.latest_timeline_request(world_id))
	print(client.timelines_request(world_id, 5))
	print(client.logs_request({"world_id": world_id, "limit": 10, "task_type": "tick"}))
	print(client.debug_traces_request(world_id, 10))
	print(client.world_policy_request(world_id))
	print(client.set_world_policy_request(world_id, {"mode": "balanced"}))
