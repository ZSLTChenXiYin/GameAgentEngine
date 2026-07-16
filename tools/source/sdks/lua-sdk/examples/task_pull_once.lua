local Client = require("../src/client")

local client = Client.new("http://127.0.0.1:8080", "dev-key")
print(client:pending_tasks_request("game_client", 1).path)
print(client:claim_runtime_task_payload("task-1", "game_client", "lua-sdk-example"))
print(client:start_runtime_task_payload("task-1", "lease-token"))

