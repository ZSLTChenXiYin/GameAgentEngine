using System.Text.Json;
using GameAgentEngine.SDK;

var client = new GameAgentEngineClient(
    Environment.GetEnvironmentVariable("GAE_SERVER") ?? "http://127.0.0.1:8080",
    Environment.GetEnvironmentVariable("GAE_KEY") ?? "dev-key");

var dynamicInterfacesFile = Environment.GetEnvironmentVariable("GAE_DYNAMIC_INTERFACES_FILE")
    ?? "tools/source/workerhome/fixtures/runtime_task_dynamic_interfaces.json";
var dynamicInterfaces = JsonSerializer.Deserialize<List<DynamicInterface>>(
    await File.ReadAllTextAsync(dynamicInterfacesFile),
    new JsonSerializerOptions
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
    }) ?? new List<DynamicInterface>();

var response = await client.InvokeAsync(new InvokeRequest
{
    WorldId = Environment.GetEnvironmentVariable("GAE_WORLD_ID") ?? "demo_world",
    NodeId = Environment.GetEnvironmentVariable("GAE_NODE_ID") ?? "innkeeper_001",
    TaskType = Environment.GetEnvironmentVariable("GAE_TASK_TYPE") ?? "npc_dialogue",
    Messages = new List<ChatMessage>
    {
        new()
        {
            Role = "user",
            Content = Environment.GetEnvironmentVariable("GAE_MESSAGE") ?? "Before answering, query the nearby scene state and then respond."
        }
    },
    Context = new InvokeContext
    {
        PipelineMode = Environment.GetEnvironmentVariable("GAE_PIPELINE_MODE") ?? "full",
        DynamicInterfaces = dynamicInterfaces,
    }
});

Console.WriteLine($"request_id={response?.RequestId} task_type={response?.TaskType} execution_mode={response?.ExecutionMode}");
Console.WriteLine($"reply={response?.Reply ?? "<no reply>"}");

var dataRequest = response?.ActionCalls?.FirstOrDefault(item =>
    string.Equals(item.ActionId, "data_request", StringComparison.OrdinalIgnoreCase) &&
    string.Equals(item.Mode, "async", StringComparison.OrdinalIgnoreCase));

if (string.IsNullOrWhiteSpace(dataRequest?.CallbackId))
{
    Console.WriteLine("No async data request was emitted.");
    return;
}

var tasks = await client.ListRuntimeTasksAsync(limit: 20);
var task = tasks?.Tasks?.FirstOrDefault(item => item.CallbackId == dataRequest.CallbackId);
if (task is null)
{
    throw new InvalidOperationException($"Runtime task not found for callback_id={dataRequest.CallbackId}");
}

Console.WriteLine($"runtime_task_id={task.TaskId} interface={task.InterfaceName} consumer={task.Consumer} status={task.Status} callback_id={task.CallbackId}");
Console.WriteLine($"Next step: GameAgentWorker pull-once --consumer {task.Consumer ?? "game_client"}");
