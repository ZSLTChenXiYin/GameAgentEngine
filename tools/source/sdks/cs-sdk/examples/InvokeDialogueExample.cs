using System;
using GameAgentEngine.SDK;

var client = new GameAgentEngineClient(Environment.GetEnvironmentVariable("GAE_SERVER") ?? "http://127.0.0.1:8080", Environment.GetEnvironmentVariable("GAE_KEY") ?? "dev-key");
var response = await client.InvokeAsync(new InvokeRequest
{
    WorldId = Environment.GetEnvironmentVariable("GAE_WORLD_ID") ?? "demo_world",
    NodeId = Environment.GetEnvironmentVariable("GAE_NODE_ID") ?? "innkeeper_001",
    TaskType = "npc_dialogue",
    Messages = new List<ChatMessage>
    {
        new() { Role = "user", Content = "Is this place safe tonight?" }
    }
});
Console.WriteLine(response?.Reply ?? "<no reply>");
