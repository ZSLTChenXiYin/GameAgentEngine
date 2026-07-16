using System;
using GameAgentEngine.SDK;

var client = new GameAgentEngineClient(Environment.GetEnvironmentVariable("GAE_SERVER") ?? "http://127.0.0.1:8080", Environment.GetEnvironmentVariable("GAE_KEY") ?? "dev-key");
Console.WriteLine(await client.HealthAsync());
