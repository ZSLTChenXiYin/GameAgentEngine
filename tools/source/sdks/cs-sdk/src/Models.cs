using System.Text.Json;
using System.Text.Json.Serialization;

namespace GameAgentEngine.SDK;

public sealed class VersionInfo
{
    [JsonPropertyName("version")]
    public string Version { get; set; } = string.Empty;

    [JsonPropertyName("min_compatible")]
    public string MinCompatible { get; set; } = string.Empty;
}

public sealed class ChatMessage
{
    [JsonPropertyName("role")]
    public string Role { get; set; } = string.Empty;

    [JsonPropertyName("content")]
    public string Content { get; set; } = string.Empty;
}

public sealed class DynamicInterface
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;

    [JsonPropertyName("kind")]
    public string Kind { get; set; } = string.Empty;

    [JsonPropertyName("external_interface")]
    public string ExternalInterface { get; set; } = string.Empty;

    [JsonPropertyName("description")]
    public string? Description { get; set; }

    [JsonPropertyName("query_types")]
    public List<string>? QueryTypes { get; set; }

    [JsonPropertyName("args_schema")]
    public Dictionary<string, JsonElement>? ArgsSchema { get; set; }

    [JsonPropertyName("max_queries")]
    public int? MaxQueries { get; set; }

    [JsonPropertyName("max_calls")]
    public int? MaxCalls { get; set; }
}

public sealed class InteractionEvent
{
    [JsonPropertyName("type")]
    public string Type { get; set; } = string.Empty;

    [JsonPropertyName("item_id")]
    public string? ItemId { get; set; }

    [JsonPropertyName("args")]
    public Dictionary<string, JsonElement>? Args { get; set; }
}

public sealed class InteractionContext
{
    [JsonPropertyName("mode")]
    public string? Mode { get; set; }

    [JsonPropertyName("speaker_node_id")]
    public string? SpeakerNodeId { get; set; }

    [JsonPropertyName("target_node_id")]
    public string? TargetNodeId { get; set; }

    [JsonPropertyName("scene_node_id")]
    public string? SceneNodeId { get; set; }

    [JsonPropertyName("room_id")]
    public string? RoomId { get; set; }

    [JsonPropertyName("participant_node_ids")]
    public List<string>? ParticipantNodeIds { get; set; }

    [JsonPropertyName("audience_scope")]
    public string? AudienceScope { get; set; }

    [JsonPropertyName("turn_index")]
    public int? TurnIndex { get; set; }

    [JsonPropertyName("event")]
    public InteractionEvent? Event { get; set; }
}

public sealed class InvokeContext
{
    [JsonPropertyName("max_analysis_rounds")]
    public int? MaxAnalysisRounds { get; set; }

    [JsonPropertyName("max_depth")]
    public int? MaxDepth { get; set; }

    [JsonPropertyName("memory_limit")]
    public int? MemoryLimit { get; set; }

    [JsonPropertyName("include_related_nodes")]
    public bool? IncludeRelatedNodes { get; set; }

    [JsonPropertyName("pipeline_mode")]
    public string? PipelineMode { get; set; }

    [JsonPropertyName("player_input_interpret")]
    public bool? PlayerInputInterpret { get; set; }

    [JsonPropertyName("dynamic_interfaces")]
    public List<DynamicInterface>? DynamicInterfaces { get; set; }

    [JsonPropertyName("interaction")]
    public InteractionContext? Interaction { get; set; }
}

public sealed class InvokeRequest
{
    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("task_type")]
    public string TaskType { get; set; } = string.Empty;

    [JsonPropertyName("node_id")]
    public string NodeId { get; set; } = string.Empty;

    [JsonPropertyName("session_id")]
    public string? SessionId { get; set; }

    [JsonPropertyName("messages")]
    public List<ChatMessage>? Messages { get; set; }

    [JsonPropertyName("context")]
    public InvokeContext? Context { get; set; }
}

public sealed class PlayerInputInterpretRequest
{
    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("player_node_id")]
    public string PlayerNodeId { get; set; } = string.Empty;

    [JsonPropertyName("scene_node_id")]
    public string? SceneNodeId { get; set; }

    [JsonPropertyName("target_node_id")]
    public string? TargetNodeId { get; set; }

    [JsonPropertyName("session_id")]
    public string? SessionId { get; set; }

    [JsonPropertyName("message")]
    public string Message { get; set; } = string.Empty;

    [JsonPropertyName("participant_node_ids")]
    public List<string>? ParticipantNodeIds { get; set; }

    [JsonPropertyName("context")]
    public InvokeContext? Context { get; set; }
}

public sealed class ActionCall
{
    [JsonPropertyName("action_id")]
    public string ActionId { get; set; } = string.Empty;

    [JsonPropertyName("args")]
    public Dictionary<string, JsonElement>? Args { get; set; }

    [JsonPropertyName("mode")]
    public string? Mode { get; set; }

    [JsonPropertyName("callback_id")]
    public string? CallbackId { get; set; }
}

public sealed class ResponseMeta
{
    [JsonPropertyName("llm_model")]
    public string? LlmModel { get; set; }

    [JsonPropertyName("tokens_used")]
    public int? TokensUsed { get; set; }

    [JsonPropertyName("processing_time_ms")]
    public long? ProcessingTimeMs { get; set; }

    [JsonPropertyName("configured_pipeline_mode")]
    public string? ConfiguredPipelineMode { get; set; }

    [JsonPropertyName("effective_pipeline_mode")]
    public string? EffectivePipelineMode { get; set; }

    [JsonPropertyName("max_analysis_rounds")]
    public int? MaxAnalysisRounds { get; set; }

    [JsonPropertyName("rounds_used")]
    public int? RoundsUsed { get; set; }
}

public sealed class PlayerIntentInterpretation
{
    [JsonPropertyName("intent")]
    public JsonElement? Intent { get; set; }

    [JsonPropertyName("missing_facts")]
    public List<JsonElement>? MissingFacts { get; set; }

    [JsonPropertyName("suggested_interaction")]
    public JsonElement? SuggestedInteraction { get; set; }
}

public sealed class InvokeResponse
{
    [JsonPropertyName("request_id")]
    public string RequestId { get; set; } = string.Empty;

    [JsonPropertyName("task_type")]
    public string TaskType { get; set; } = string.Empty;

    [JsonPropertyName("execution_mode")]
    public string ExecutionMode { get; set; } = string.Empty;

    [JsonPropertyName("reply")]
    public string? Reply { get; set; }

    [JsonPropertyName("advanced_ticks")]
    public int? AdvancedTicks { get; set; }

    [JsonPropertyName("action_calls")]
    public List<ActionCall>? ActionCalls { get; set; }

    [JsonPropertyName("memory_updates")]
    public List<JsonElement>? MemoryUpdates { get; set; }

    [JsonPropertyName("player_intent")]
    public PlayerIntentInterpretation? PlayerIntent { get; set; }

    [JsonPropertyName("metadata")]
    public ResponseMeta? Metadata { get; set; }
}

public sealed class TickAdvanceRequest
{
    [JsonPropertyName("tick_type")]
    public string TickType { get; set; } = string.Empty;

    [JsonPropertyName("game_time")]
    public string GameTime { get; set; } = string.Empty;

    [JsonPropertyName("requested_ticks")]
    public int? RequestedTicks { get; set; }

    [JsonPropertyName("autonomous_limit")]
    public int? AutonomousLimit { get; set; }
}

public sealed class TickResponse
{
    [JsonPropertyName("tick")]
    public TimelineTick? Tick { get; set; }

    [JsonPropertyName("invoke")]
    public InvokeResponse? Invoke { get; set; }

    [JsonPropertyName("advanced_ticks")]
    public int? AdvancedTicks { get; set; }

    [JsonPropertyName("world_time_state")]
    public JsonElement? WorldTimeState { get; set; }

    [JsonPropertyName("autonomous_runs")]
    public List<JsonElement>? AutonomousRuns { get; set; }
}

public sealed class WorldSettings
{
    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("memory_limit")]
    public int MemoryLimit { get; set; }

    [JsonPropertyName("max_analysis_rounds")]
    public int MaxAnalysisRounds { get; set; }

    [JsonPropertyName("max_context_depth")]
    public int MaxContextDepth { get; set; }

    [JsonPropertyName("auto_apply")]
    public bool AutoApply { get; set; }

    [JsonPropertyName("require_review_above")]
    public string RequireReviewAbove { get; set; } = string.Empty;

    [JsonPropertyName("propagation_max_depth")]
    public int PropagationMaxDepth { get; set; }

    [JsonPropertyName("enable_propagation_machine")]
    public bool EnablePropagationMachine { get; set; }

    [JsonPropertyName("sub_task_max_retries")]
    public int SubTaskMaxRetries { get; set; }

    [JsonPropertyName("sub_task_timeout_secs")]
    public int SubTaskTimeoutSecs { get; set; }

    [JsonPropertyName("pipeline_mode")]
    public string PipelineMode { get; set; } = string.Empty;

    [JsonPropertyName("world_time_settings")]
    public JsonElement? WorldTimeSettings { get; set; }
}

public sealed class StateComponentEnvelope
{
    [JsonPropertyName("component_type")]
    public string ComponentType { get; set; } = string.Empty;

    [JsonPropertyName("component")]
    public JsonElement? Component { get; set; }

    [JsonPropertyName("data")]
    public JsonElement? Data { get; set; }
}

public sealed class StateComponentsResponse
{
    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("components")]
    public List<StateComponentEnvelope>? Components { get; set; }
}

public sealed class StateComponentResponse
{
    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("state_component")]
    public StateComponentEnvelope? StateComponent { get; set; }
}

public sealed class TimelineTick
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;

    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("tick_number")]
    public int TickNumber { get; set; }

    [JsonPropertyName("tick_type")]
    public string TickType { get; set; } = string.Empty;

    [JsonPropertyName("game_time")]
    public string? GameTime { get; set; }

    [JsonPropertyName("summary")]
    public string? Summary { get; set; }

    [JsonPropertyName("data")]
    public string? Data { get; set; }

    [JsonPropertyName("future_outline")]
    public string? FutureOutline { get; set; }

    [JsonPropertyName("created_at")]
    public string CreatedAt { get; set; } = string.Empty;
}

public sealed class TimelineEnvelope
{
    [JsonPropertyName("tick_number")]
    public int TickNumber { get; set; }

    [JsonPropertyName("tick_type")]
    public string TickType { get; set; } = string.Empty;

    [JsonPropertyName("game_time")]
    public string? GameTime { get; set; }

    [JsonPropertyName("advanced_ticks")]
    public int? AdvancedTicks { get; set; }

    [JsonPropertyName("summary")]
    public string? Summary { get; set; }

    [JsonPropertyName("future_outline")]
    public string? FutureOutline { get; set; }

    [JsonPropertyName("timeline")]
    public TimelineTick? Timeline { get; set; }

    [JsonPropertyName("data")]
    public JsonElement? Data { get; set; }
}

public sealed class TimelinesResponse
{
    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("timelines")]
    public List<TimelineEnvelope>? Timelines { get; set; }
}

public sealed class LatestTimelineResponse
{
    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("timeline")]
    public TimelineEnvelope? Timeline { get; set; }
}

public sealed class InferenceLogQuery
{
    public string? WorldId { get; set; }
    public string? NodeId { get; set; }
    public string? TaskType { get; set; }
    public string? Category { get; set; }
    public string? EventName { get; set; }
    public string? ExecutionMode { get; set; }
    public string? RequestId { get; set; }
    public int? Round { get; set; }
    public int? Limit { get; set; }
    public int? Offset { get; set; }
}

public sealed class InferenceLog
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;

    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("task_type")]
    public string TaskType { get; set; } = string.Empty;

    [JsonPropertyName("node_id")]
    public string NodeId { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string? Category { get; set; }

    [JsonPropertyName("event_name")]
    public string? EventName { get; set; }

    [JsonPropertyName("message")]
    public string? Message { get; set; }

    [JsonPropertyName("request_id")]
    public string? RequestId { get; set; }

    [JsonPropertyName("execution_mode")]
    public string? ExecutionMode { get; set; }

    [JsonPropertyName("round")]
    public int? Round { get; set; }

    [JsonPropertyName("created_at")]
    public string? CreatedAt { get; set; }
}

public sealed class DebugTrace
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;

    [JsonPropertyName("world_id")]
    public string WorldId { get; set; } = string.Empty;

    [JsonPropertyName("request_id")]
    public string RequestId { get; set; } = string.Empty;

    [JsonPropertyName("task_type")]
    public string TaskType { get; set; } = string.Empty;

    [JsonPropertyName("node_id")]
    public string NodeId { get; set; } = string.Empty;

    [JsonPropertyName("timestamp")]
    public string? Timestamp { get; set; }

    [JsonPropertyName("duration_ms")]
    public long? DurationMs { get; set; }

    [JsonPropertyName("error")]
    public string? Error { get; set; }
}

public sealed class DebugTraceList
{
    [JsonPropertyName("traces")]
    public List<DebugTrace>? Traces { get; set; }

    [JsonPropertyName("count")]
    public int Count { get; set; }
}

public sealed class RuntimeTask
{
    [JsonPropertyName("task_id")]
    public string TaskId { get; set; } = string.Empty;

    [JsonPropertyName("category")]
    public string? Category { get; set; }

    [JsonPropertyName("interface_name")]
    public string? InterfaceName { get; set; }

    [JsonPropertyName("delivery_mode")]
    public string? DeliveryMode { get; set; }

    [JsonPropertyName("consumer")]
    public string? Consumer { get; set; }

    [JsonPropertyName("transport")]
    public string? Transport { get; set; }

    [JsonPropertyName("world_id")]
    public string? WorldId { get; set; }

    [JsonPropertyName("node_id")]
    public string? NodeId { get; set; }

    [JsonPropertyName("request_id")]
    public string? RequestId { get; set; }

    [JsonPropertyName("callback_id")]
    public string? CallbackId { get; set; }

    [JsonPropertyName("resume_execution_id")]
    public string? ResumeExecutionId { get; set; }

    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;

    [JsonPropertyName("lease_token")]
    public string? LeaseToken { get; set; }

    [JsonPropertyName("lease_owner")]
    public string? LeaseOwner { get; set; }

    [JsonPropertyName("payload_json")]
    public string? PayloadJson { get; set; }

    [JsonPropertyName("result_json")]
    public string? ResultJson { get; set; }

    [JsonPropertyName("error_message")]
    public string? ErrorMessage { get; set; }

    [JsonPropertyName("created_at")]
    public string? CreatedAt { get; set; }

    [JsonPropertyName("updated_at")]
    public string? UpdatedAt { get; set; }
}

public sealed class RuntimeTaskListResponse
{
    [JsonPropertyName("tasks")]
    public List<RuntimeTask>? Tasks { get; set; }
}

public sealed class RuntimeTaskEnvelope
{
    [JsonPropertyName("task")]
    public RuntimeTask? Task { get; set; }
}

public sealed class RuntimeTaskStats
{
    [JsonPropertyName("generated_at")]
    public string? GeneratedAt { get; set; }

    [JsonPropertyName("total")]
    public long? Total { get; set; }

    [JsonPropertyName("ready_pull")]
    public long? ReadyPull { get; set; }

    [JsonPropertyName("in_flight")]
    public long? InFlight { get; set; }

    [JsonPropertyName("terminal")]
    public long? Terminal { get; set; }

    [JsonPropertyName("heartbeat_timeout")]
    public long? HeartbeatTimeout { get; set; }

    [JsonPropertyName("by_status")]
    public Dictionary<string, long>? ByStatus { get; set; }

    [JsonPropertyName("by_category")]
    public Dictionary<string, long>? ByCategory { get; set; }

    [JsonPropertyName("by_consumer")]
    public Dictionary<string, long>? ByConsumer { get; set; }
}

public sealed class RuntimeTaskStatsEnvelope
{
    [JsonPropertyName("stats")]
    public RuntimeTaskStats? Stats { get; set; }
}

public sealed class CallbackPostProcess
{
    [JsonPropertyName("status")]
    public string? Status { get; set; }

    [JsonPropertyName("applied")]
    public bool? Applied { get; set; }

    [JsonPropertyName("details")]
    public JsonElement? Details { get; set; }
}

public sealed class CallbackResponse
{
    [JsonPropertyName("status")]
    public string Status { get; set; } = string.Empty;

    [JsonPropertyName("resume_execution_id")]
    public string? ResumeExecutionId { get; set; }

    [JsonPropertyName("post_process")]
    public CallbackPostProcess? PostProcess { get; set; }

    [JsonPropertyName("resumed")]
    public InvokeResponse? Resumed { get; set; }
}

