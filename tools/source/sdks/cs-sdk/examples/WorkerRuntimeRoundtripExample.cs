using GameAgentEngine.SDK;

static string Required(string name, string? fallback = null)
{
    var value = Environment.GetEnvironmentVariable(name) ?? fallback;
    if (string.IsNullOrWhiteSpace(value))
    {
        throw new InvalidOperationException($"Missing required environment variable: {name}");
    }
    return value;
}

var client = new GameAgentEngineClient(
    Environment.GetEnvironmentVariable("GAE_SERVER") ?? "http://127.0.0.1:8080",
    Environment.GetEnvironmentVariable("GAE_KEY") ?? "dev-key");

var consumer = Environment.GetEnvironmentVariable("GAE_CONSUMER") ?? "game_client";
var owner = Environment.GetEnvironmentVariable("GAE_OWNER") ?? "cs-sdk-roundtrip";
var resultStatus = Environment.GetEnvironmentVariable("GAE_CALLBACK_STATUS") ?? "success";

var pending = await client.ListPendingRuntimeTasksAsync(consumer, 1);
var task = pending?.Tasks?.FirstOrDefault();
if (task is null)
{
    Console.WriteLine($"No pending runtime task for consumer={consumer}.");
    return;
}

Console.WriteLine($"Claiming task {task.TaskId} ({task.InterfaceName ?? "unknown_interface"})");
var claimed = await client.ClaimRuntimeTaskAsync(task.TaskId, consumer, owner);
var claimedTask = claimed?.Task;
if (string.IsNullOrWhiteSpace(claimedTask?.LeaseToken))
{
    throw new InvalidOperationException($"Task {task.TaskId} missing lease token after claim.");
}

var started = await client.StartRuntimeTaskAsync(task.TaskId, claimedTask.LeaseToken);
Console.WriteLine($"Started task {task.TaskId} status={started?.Task?.Status}");

var callbackId = Required("GAE_CALLBACK_ID", started?.Task?.CallbackId ?? task.CallbackId);
var callbackRequestId = Environment.GetEnvironmentVariable("GAE_CALLBACK_REQUEST_ID") ?? $"cs-sdk-{task.TaskId}";
var result = new
{
    worker = "cs-sdk-example",
    source = "worker_runtime_roundtrip",
    interface_name = task.InterfaceName,
    task_id = task.TaskId,
    consumer,
};

var callback = await client.ActionCallbackAsync(callbackId, resultStatus, result, callbackRequestId);
Console.WriteLine($"callback_status={callback?.Status} resume_execution_id={callback?.ResumeExecutionId} resumed={(callback?.Resumed is not null)} post_process_applied={(callback?.PostProcess?.Applied ?? false)}");

