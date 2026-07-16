extends Node

func _ready() -> void:
	var client := GameAgentEngineClient.new("http://127.0.0.1:8080", "dev-key")
	var file := FileAccess.open("tools/source/tests/runtime_task_dynamic_interfaces.json", FileAccess.READ)
	var dynamic_interfaces = JSON.parse_string(file.get_as_text())
	var request := client.invoke_request({
		"world_id": "demo_world",
		"node_id": "innkeeper_001",
		"task_type": "npc_dialogue",
		"messages": [{
			"role": "user",
			"content": "Before answering, query the nearby scene state and then respond."
		}],
		"context": {
			"pipeline_mode": "full",
			"dynamic_interfaces": dynamic_interfaces,
		}
	})
	print(request)
	print(client.list_runtime_tasks_request(null, null, 20))
	print("After the Engine emits a runtime task, process it with GameAgentWorker pull-once --consumer game_client")
