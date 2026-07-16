using System;
using GameAgentEngine.SDK;

var client = new GameAgentEngineClient(Environment.GetEnvironmentVariable("GAE_SERVER") ?? "http://127.0.0.1:8080", Environment.GetEnvironmentVariable("GAE_KEY") ?? "dev-key");
var pending = await client.ListPendingRuntimeTasksAsync(Environment.GetEnvironmentVariable("GAE_CONSUMER") ?? "game_client", 1);
var task = pending?.Tasks?.FirstOrDefault();
if (task is null)
{
    Console.WriteLine("No pending tasks.");
    return;
}

Console.WriteLine($"task_id={task.TaskId} interface={task.InterfaceName} status={task.Status}");
