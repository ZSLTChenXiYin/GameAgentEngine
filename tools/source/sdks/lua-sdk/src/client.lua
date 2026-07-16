local Client = {}
Client.__index = Client

function Client.new(base_url, api_key)
  return setmetatable({ base_url = base_url, api_key = api_key }, Client)
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

return Client
