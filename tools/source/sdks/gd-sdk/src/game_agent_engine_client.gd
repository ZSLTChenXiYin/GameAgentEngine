class_name GameAgentEngineClient
extends RefCounted

var base_url: String
var api_key: String

func _init(p_base_url: String, p_api_key: String) -> void:
	base_url = p_base_url.trim_suffix("/")
	api_key = p_api_key

func build_headers() -> PackedStringArray:
	return ["Content-Type: application/json", "X-API-Key: %s" % api_key]

func health_path() -> String:
	return base_url + "/health"

func version_path() -> String:
	return base_url + "/api/v1/version"

func invoke_path() -> String:
	return base_url + "/api/v1/invoke"

func pending_tasks_path(consumer: String, limit: int = 20) -> String:
	return "%s/api/v1/runtime/tasks/pending?consumer=%s&limit=%d" % [base_url, consumer.uri_encode(), limit]
