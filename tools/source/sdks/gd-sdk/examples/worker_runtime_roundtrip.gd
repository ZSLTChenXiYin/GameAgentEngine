extends Node

func _ready() -> void:
	var client := GameAgentEngineClient.new("http://127.0.0.1:8080", "dev-key")
	print(client.pending_tasks_request("game_client", 1))
	print(client.claim_runtime_task_request("task-1", "game_client", "gd-sdk-example"))
	print(client.start_runtime_task_request("task-1", "lease-token"))
	print(client.heartbeat_runtime_task_request("task-1", "lease-token"))
	print(client.action_callback_request("callback-1", "success", {"worker": "gd-sdk-example"}, "gd-sdk-task-1"))
	print(client.release_runtime_task_request("task-1", "lease-token", "manual release"))
	print(client.requeue_runtime_task_request("task-1", 1500, "manual requeue"))
	print(client.runtime_task_stats_request())
