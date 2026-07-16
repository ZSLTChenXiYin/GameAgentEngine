extends Node

func _ready() -> void:
	var client := GameAgentEngineClient.new("http://127.0.0.1:8080", "dev-key")
	print(client.health_path())
