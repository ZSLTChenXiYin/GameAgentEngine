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

function Client:callback_payload(callback_id, status, result_json)
  return string.format('{"callback_id":"%s","status":"%s","result":%s}', callback_id, status, result_json)
end

function Client:health_request()
  return { method = "GET", path = "/health" }
end

function Client:invoke_request(body_json)
  return { method = "POST", path = "/api/v1/invoke", body = body_json }
end

function Client:pending_tasks_request(consumer, limit)
  return { method = "GET", path = string.format("/api/v1/runtime/tasks/pending?consumer=%s&limit=%d", consumer, limit or 20) }
end

function Client:claim_runtime_task_request(task_id, consumer, owner)
  return { method = "POST", path = "/api/v1/runtime/tasks/claim", body = self:claim_runtime_task_payload(task_id, consumer, owner) }
end

function Client:start_runtime_task_request(task_id, lease_token)
  return { method = "POST", path = "/api/v1/runtime/tasks/start", body = self:start_runtime_task_payload(task_id, lease_token) }
end

function Client:action_callback_request(callback_id, status, result_json)
  return { method = "POST", path = "/api/v1/actions/callback", body = self:callback_payload(callback_id, status, result_json) }
end

return Client
