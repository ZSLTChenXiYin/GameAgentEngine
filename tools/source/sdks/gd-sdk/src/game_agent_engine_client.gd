class_name GameAgentEngineClient
extends RefCounted

var base_url: String
var api_key: String

func _init(p_base_url: String, p_api_key: String) -> void:
	base_url = p_base_url.trim_suffix("/")
	api_key = p_api_key

func build_headers(extra_headers: Dictionary = {}) -> PackedStringArray:
	var headers: PackedStringArray = [
		"Accept: application/json",
		"X-API-Key: %s" % api_key,
	]
	for key in extra_headers.keys():
		headers.append("%s: %s" % [str(key), str(extra_headers[key])])
	return headers

func build_query(params: Dictionary) -> String:
	var parts: Array[String] = []
	for key in params.keys():
		var value = params[key]
		if value == null:
			continue
		var text := str(value)
		if text.is_empty():
			continue
		parts.append("%s=%s" % [String(key).uri_encode(), text.uri_encode()])
	if parts.is_empty():
		return ""
	return "?" + "&".join(parts)

func request_path(path: String) -> String:
	return base_url + path

func create_request(method: int, path: String, body := null, extra_headers: Dictionary = {}) -> Dictionary:
	var request := {
		"url": request_path(path),
		"method": method,
		"headers": build_headers(extra_headers),
	}
	if body != null:
		request["body"] = JSON.stringify(body)
		request["headers"].append("Content-Type: application/json")
	return request

func parse_response(body: PackedByteArray) -> Variant:
	if body.is_empty():
		return null
	var text := body.get_string_from_utf8()
	if text.strip_edges().is_empty():
		return null
	var parsed := JSON.parse_string(text)
	if parsed == null:
		return text
	return parsed

func ensure_success(response_code: int, path: String, body: PackedByteArray) -> void:
	if response_code >= 400:
		var text := body.get_string_from_utf8()
		push_error("HTTP %d %s: %s" % [response_code, path, text])

func health_request() -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/health")

func version_request() -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/version")

func invoke_request(payload: Dictionary) -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/invoke", payload)

func interpret_player_input_request(payload: Dictionary) -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/player/input/interpret", payload)

func advance_tick_request(world_id: String, payload: Dictionary) -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/worlds/%s/ticks/advance" % world_id.uri_encode(), payload)

func world_settings_request(world_id: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/worlds/%s/settings" % world_id.uri_encode())

func set_world_settings_request(world_id: String, payload: Dictionary) -> Dictionary:
	return create_request(HTTPClient.METHOD_PUT, "/api/v1/worlds/%s/settings" % world_id.uri_encode(), payload)

func state_components_request(world_id: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/worlds/%s/state-components" % world_id.uri_encode())

func state_component_request(world_id: String, component_type: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/worlds/%s/state-components/%s" % [world_id.uri_encode(), component_type.uri_encode()])

func put_state_component_request(world_id: String, component_type: String, payload: Dictionary) -> Dictionary:
	return create_request(HTTPClient.METHOD_PUT, "/api/v1/worlds/%s/state-components/%s" % [world_id.uri_encode(), component_type.uri_encode()], payload)

func timelines_request(world_id: String, limit: int = 0) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/worlds/%s/timelines%s" % [world_id.uri_encode(), build_query({"limit": limit if limit > 0 else null})])

func latest_timeline_request(world_id: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/worlds/%s/timelines/latest" % world_id.uri_encode())

func logs_request(query: Dictionary = {}) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/logs%s" % build_query(query))

func debug_traces_request(world_id: String, limit: int = 20) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/debug/traces%s" % build_query({"world_id": world_id, "limit": limit}))

func list_runtime_tasks_request(category := null, status := null, limit: int = 20) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/runtime/tasks%s" % build_query({"category": category, "status": status, "limit": limit}))

func pending_tasks_request(consumer: String, limit: int = 20) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/runtime/tasks/pending%s" % build_query({"consumer": consumer, "limit": limit}))

func runtime_task_request(task_id: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/runtime/tasks/%s" % task_id.uri_encode())

func claim_runtime_task_request(task_id: String, consumer: String, lease_owner: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/runtime/tasks/claim", {
		"task_id": task_id,
		"consumer": consumer,
		"lease_owner": lease_owner,
	})

func start_runtime_task_request(task_id: String, lease_token: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/runtime/tasks/start", {
		"task_id": task_id,
		"lease_token": lease_token,
	})

func heartbeat_runtime_task_request(task_id: String, lease_token: String) -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/runtime/tasks/heartbeat", {
		"task_id": task_id,
		"lease_token": lease_token,
	})

func release_runtime_task_request(task_id: String, lease_token: String, error_message: String = "") -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/runtime/tasks/release", {
		"task_id": task_id,
		"lease_token": lease_token,
		"error_message": error_message,
	})

func requeue_runtime_task_request(task_id: String, retry_delay_ms: int = 0, error_message: String = "") -> Dictionary:
	return create_request(HTTPClient.METHOD_POST, "/api/v1/runtime/tasks/requeue", {
		"task_id": task_id,
		"retry_delay_ms": retry_delay_ms,
		"error_message": error_message,
	})

func runtime_task_stats_request() -> Dictionary:
	return create_request(HTTPClient.METHOD_GET, "/api/v1/runtime/tasks/stats")

func action_callback_request(callback_id: String, status: String, result: Dictionary, callback_request_id: String = "") -> Dictionary:
	var headers := {}
	if not callback_request_id.is_empty():
		headers["X-Callback-Request-Id"] = callback_request_id
	return create_request(HTTPClient.METHOD_POST, "/api/v1/actions/callback", {
		"callback_id": callback_id,
		"status": status,
		"result": result,
	}, headers)

