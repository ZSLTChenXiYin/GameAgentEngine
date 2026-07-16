local Client = {}
Client.__index = Client

function Client.new(base_url, api_key)
  return setmetatable({ base_url = string.gsub(base_url, "/$", ""), api_key = api_key }, Client)
end

function Client:health_path()
  return self.base_url .. "/health"
end

function Client:version_path()
  return self.base_url .. "/api/v1/version"
end

function Client:invoke_path()
  return self.base_url .. "/api/v1/invoke"
end

function Client:interpret_player_input_path()
  return self.base_url .. "/api/v1/player/input/interpret"
end

function Client:pending_tasks_path(consumer, limit)
  return string.format("%s/api/v1/runtime/tasks/pending?consumer=%s&limit=%d", self.base_url, consumer, limit or 20)
end

function Client:runtime_tasks_path(category, status, limit)
  local parts = { string.format("limit=%d", limit or 20) }
  if category and category ~= "" then table.insert(parts, "category=" .. category) end
  if status and status ~= "" then table.insert(parts, "status=" .. status) end
  return self.base_url .. "/api/v1/runtime/tasks?" .. table.concat(parts, "&")
end

function Client:runtime_task_path(task_id)
  return self.base_url .. "/api/v1/runtime/tasks/" .. task_id
end

function Client:claim_runtime_task_payload(task_id, consumer, owner)
  return string.format('{"task_id":"%s","consumer":"%s","lease_owner":"%s"}', task_id, consumer, owner)
end

function Client:start_runtime_task_payload(task_id, lease_token)
  return string.format('{"task_id":"%s","lease_token":"%s"}', task_id, lease_token)
end

function Client:heartbeat_runtime_task_payload(task_id, lease_token)
  return string.format('{"task_id":"%s","lease_token":"%s"}', task_id, lease_token)
end

function Client:release_runtime_task_payload(task_id, lease_token, error_message)
  return string.format('{"task_id":"%s","lease_token":"%s","error_message":"%s"}', task_id, lease_token, error_message or "")
end

function Client:requeue_runtime_task_payload(task_id, retry_delay_ms, error_message)
  return string.format('{"task_id":"%s","retry_delay_ms":%d,"error_message":"%s"}', task_id, retry_delay_ms or 0, error_message or "")
end

function Client:callback_payload(callback_id, status, result_json)
  return string.format('{"callback_id":"%s","status":"%s","result":%s}', callback_id, status, result_json)
end

function Client:health_request()
  return { method = "GET", path = "/health" }
end

function Client:invoke_request(body_json)
  return { method = "POST", path = "/api/v1/invoke", body = body_json }
end

function Client:interpret_player_input_request(body_json)
  return { method = "POST", path = "/api/v1/player/input/interpret", body = body_json }
end

function Client:runtime_tasks_request(category, status, limit)
  return { method = "GET", path = "/api/v1/runtime/tasks" .. self:build_query({ category = category, status = status, limit = limit or 20 }) }
end

function Client:pending_tasks_request(consumer, limit)
  return { method = "GET", path = string.format("/api/v1/runtime/tasks/pending?consumer=%s&limit=%d", consumer, limit or 20) }
end

function Client:runtime_task_request(task_id)
  return { method = "GET", path = "/api/v1/runtime/tasks/" .. task_id }
end

function Client:claim_runtime_task_request(task_id, consumer, owner)
  return { method = "POST", path = "/api/v1/runtime/tasks/claim", body = self:claim_runtime_task_payload(task_id, consumer, owner) }
end

function Client:start_runtime_task_request(task_id, lease_token)
  return { method = "POST", path = "/api/v1/runtime/tasks/start", body = self:start_runtime_task_payload(task_id, lease_token) }
end

function Client:heartbeat_runtime_task_request(task_id, lease_token)
  return { method = "POST", path = "/api/v1/runtime/tasks/heartbeat", body = self:heartbeat_runtime_task_payload(task_id, lease_token) }
end

function Client:release_runtime_task_request(task_id, lease_token, error_message)
  return { method = "POST", path = "/api/v1/runtime/tasks/release", body = self:release_runtime_task_payload(task_id, lease_token, error_message) }
end

function Client:requeue_runtime_task_request(task_id, retry_delay_ms, error_message)
  return { method = "POST", path = "/api/v1/runtime/tasks/requeue", body = self:requeue_runtime_task_payload(task_id, retry_delay_ms, error_message) }
end

function Client:runtime_task_stats_request()
  return { method = "GET", path = "/api/v1/runtime/tasks/stats" }
end

function Client:action_callback_request(callback_id, status, result_json)
  return { method = "POST", path = "/api/v1/actions/callback", body = self:callback_payload(callback_id, status, result_json) }
end

function Client:build_query(params)
  local parts = {}
  for key, value in pairs(params or {}) do
    if value ~= nil and tostring(value) ~= "" then
      table.insert(parts, tostring(key) .. "=" .. tostring(value))
    end
  end
  if #parts == 0 then
    return ""
  end
  return "?" .. table.concat(parts, "&")
end

return Client
