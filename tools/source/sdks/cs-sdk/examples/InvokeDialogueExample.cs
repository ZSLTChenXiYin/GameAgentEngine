using System;
using GameAgentEngine.SDK;

var client = new GameAgentEngineClient(Environment.GetEnvironmentVariable("GAE_SERVER") ?? "http://127.0.0.1:8080", Environment.GetEnvironmentVariable("GAE_KEY") ?? "dev-key");
var response = await client.InvokeAsync(new {
    world_id = Environment.GetEnvironmentVariable("GAE_WORLD_ID") ?? "demo_world",
    node_id = Environment.GetEnvironmentVariable("GAE_NODE_ID") ?? "innkeeper_001",
    task_type = "npc_dialogue",
    messages = new [] { new { role = "user", content = "Is this place safe tonight?" } }
});
Console.WriteLine(response);
