local Client = require("../src/client")

local client = Client.new("http://127.0.0.1:8080", "dev-key")
local file = assert(io.open("tools/source/tests/runtime_task_dynamic_interfaces.json", "r"))
local dynamic_interfaces = file:read("*a")
file:close()

local request = client:invoke_request('{"world_id":"demo_world","node_id":"innkeeper_001","task_type":"npc_dialogue","messages":[{"role":"user","content":"Before answering, query the nearby scene state and then respond."}],"context":{"pipeline_mode":"full","dynamic_interfaces":' .. dynamic_interfaces .. '}}')
print(request.method, request.path)
print(request.body)
print(client:runtime_tasks_request(nil, nil, 20).path)
print("After the Engine emits a runtime task, process it with GameAgentWorker pull-once --consumer game_client")
