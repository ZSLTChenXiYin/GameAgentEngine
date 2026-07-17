local Client = require('client')

local client = Client.new(os.getenv('GAE_SERVER') or 'http://127.0.0.1:8080', os.getenv('GAE_KEY') or 'dev-key')
local world_id = os.getenv('GAE_WORLD_ID') or 'demo_world'
local log_limit = tonumber(os.getenv('GAE_LOG_LIMIT') or '10')
local trace_limit = tonumber(os.getenv('GAE_TRACE_LIMIT') or '10')
local timeline_limit = tonumber(os.getenv('GAE_TIMELINE_LIMIT') or '5')

local function dump(label, req)
  print('== ' .. label .. ' ==')
  print(req.method .. ' ' .. req.path)
  if req.body then
    print(req.body)
  end
end

dump('world settings', client:world_settings_request(world_id))
dump('state components', client:state_components_request(world_id))
dump('latest timeline', client:latest_timeline_request(world_id))
dump('recent timelines', client:timelines_request(world_id, timeline_limit))
dump('recent logs', client:logs_request(world_id, log_limit, 0, nil))
dump('recent debug traces', client:debug_traces_request(world_id, trace_limit))
dump('world policy', client:world_policy_request(world_id))
