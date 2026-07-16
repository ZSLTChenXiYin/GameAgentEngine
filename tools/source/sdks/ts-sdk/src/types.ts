export type JsonValue =
  | string
  | number
  | boolean
  | null
  | JsonValue[]
  | { [key: string]: JsonValue };

export type ChatMessage = {
  role: string;
  content: string;
};

export type DynamicInterface = {
  id: string;
  kind: string;
  external_interface: string;
  description?: string;
  query_types?: string[];
  args_schema?: Record<string, unknown>;
  max_queries?: number;
  max_calls?: number;
};

export type InteractionEvent = {
  type: string;
  item_id?: string;
  args?: Record<string, unknown>;
};

export type InteractionContext = {
  mode?: string;
  speaker_node_id?: string;
  target_node_id?: string;
  scene_node_id?: string;
  room_id?: string;
  participant_node_ids?: string[];
  audience_scope?: string;
  turn_index?: number;
  event?: InteractionEvent;
};

export type InvokeContext = {
  max_analysis_rounds?: number;
  max_depth?: number;
  memory_limit?: number;
  include_related_nodes?: boolean;
  pipeline_mode?: string;
  player_input_interpret?: boolean;
  dynamic_interfaces?: DynamicInterface[];
  interaction?: InteractionContext;
};

export type WorldEvent = {
  event_type: string;
  scope_id: string;
  description: string;
  severity: string;
};

export type InvokeRequest = {
  world_id: string;
  task_type: string;
  node_id: string;
  session_id?: string;
  messages?: ChatMessage[];
  context?: InvokeContext;
  event?: WorldEvent;
};

export type ActionCall = {
  action_id: string;
  args: Record<string, unknown>;
  mode?: string;
  callback_id?: string;
};

export type ResponseMeta = {
  llm_model?: string;
  tokens_used?: number;
  processing_time_ms?: number;
  configured_pipeline_mode?: string;
  effective_pipeline_mode?: string;
  max_analysis_rounds?: number;
  rounds_used?: number;
};

export type PlayerIntentPrecondition = {
  type: string;
  actor_node_id?: string;
  target_node_id?: string;
  scene_node_id?: string;
  item_id?: string;
  task_id?: string;
  expected?: string;
  args?: Record<string, unknown>;
};

export type PlayerIntentStep = {
  type: string;
  target_node_id?: string;
  scene_node_id?: string;
  item_id?: string;
  content?: string;
  args?: Record<string, unknown>;
  preconditions?: PlayerIntentPrecondition[];
};

export type PlayerIntent = {
  type: string;
  actor_node_id?: string;
  scene_node_id?: string;
  target_node_id?: string;
  summary?: string;
  risk_level?: string;
  confidence?: number;
  steps?: PlayerIntentStep[];
};

export type MissingFact = {
  type: string;
  node_id?: string;
  item_id?: string;
  task_id?: string;
  reason?: string;
};

export type SuggestedInteraction = {
  mode?: string;
  event_type?: string;
  audience_scope?: string;
  target_node_id?: string;
};

export type PlayerIntentInterpretation = {
  intent?: PlayerIntent;
  missing_facts?: MissingFact[];
  suggested_interaction?: SuggestedInteraction;
};

export type InvokeResponse = {
  request_id: string;
  task_type: string;
  execution_mode: string;
  reply?: string;
  advanced_ticks?: number;
  action_calls?: ActionCall[];
  memory_updates?: Array<Record<string, unknown>>;
  player_intent?: PlayerIntentInterpretation;
  sub_tasks?: Array<Record<string, unknown>>;
  metadata?: ResponseMeta;
  world_change_plan?: Record<string, unknown>;
};

export type PlayerInputInterpretRequest = {
  world_id: string;
  player_node_id: string;
  scene_node_id?: string;
  target_node_id?: string;
  session_id?: string;
  message: string;
  participant_node_ids?: string[];
  context?: InvokeContext;
};

export type RuntimeTask = {
  task_id: string;
  category?: string;
  interface_name?: string;
  delivery_mode?: string;
  consumer?: string;
  transport?: string;
  world_id?: string;
  node_id?: string;
  request_id?: string;
  callback_id?: string;
  resume_execution_id?: string;
  idempotency_key?: string;
  status: string;
  lease_token?: string;
  lease_owner?: string;
  attempt_count?: number;
  max_attempts?: number;
  priority?: number;
  payload_json?: string;
  result_json?: string;
  error_message?: string;
  dispatch_attempts?: number;
  last_dispatch_status_code?: number;
  last_dispatch_error?: string;
  last_dispatch_failure_class?: string;
  last_dispatch_decision?: string;
  fallback_from_transport?: string;
  last_transition_reason?: string;
  heartbeat_timeout_count?: number;
  available_at?: string;
  dispatched_at?: string;
  claimed_at?: string;
  last_heartbeat_at?: string;
  heartbeat_timeout_at?: string;
  completed_at?: string;
  created_at?: string;
  updated_at?: string;
};

export type RuntimeTaskStats = {
  generated_at?: string;
  total?: number;
  ready_pull?: number;
  in_flight?: number;
  terminal?: number;
  heartbeat_timeout?: number;
  dispatch_error_tasks?: number;
  retry_exhausted_tasks?: number;
  dispatched_without_callback?: number;
  repeated_heartbeat_timeouts?: number;
  oldest_dispatched_age_secs?: number;
  oldest_ready_task_age_secs?: number;
  by_status?: Record<string, number>;
  by_category?: Record<string, number>;
  by_consumer?: Record<string, number>;
  by_delivery_mode?: Record<string, number>;
  by_transport?: Record<string, number>;
  by_interface?: Record<string, number>;
  by_dispatch_failure_class?: Record<string, number>;
  by_dispatch_decision?: Record<string, number>;
  by_heartbeat_timeout_count?: Record<string, number>;
};

export type CallbackPostProcess = {
  status?: string;
  applied?: boolean;
  details?: Record<string, unknown>;
};

export type CallbackResponse = {
  status: string;
  resume_execution_id?: string;
  post_process?: CallbackPostProcess;
  resumed?: InvokeResponse;
};

export type InferenceLog = {
  id: string;
  world_id: string;
  task_type: string;
  node_id: string;
  category?: string;
  event_name?: string;
  log_level?: string;
  message?: string;
  request_id?: string;
  execution_mode?: string;
  configured_pipeline_mode?: string;
  effective_pipeline_mode?: string;
  round?: number;
  request_data?: string;
  response_data?: string;
  detail_data?: string;
  llm_model?: string;
  tokens_used?: number;
  duration_ms?: number;
  created_at?: string;
};

export type InferenceLogQuery = {
  world_id?: string;
  node_id?: string;
  task_type?: string;
  category?: string;
  event_name?: string;
  execution_mode?: string;
  request_id?: string;
  round?: number;
  limit?: number;
  offset?: number;
};

export type DebugTrace = {
  id: string;
  world_id: string;
  request_id: string;
  task_type: string;
  node_id: string;
  configured_pipeline_mode?: string;
  effective_pipeline_mode?: string;
  max_analysis_rounds?: number;
  rounds_used?: number;
  timestamp?: string;
  duration_ms?: number;
  error?: string;
};

export type DebugTraceList = {
  traces: DebugTrace[];
  count: number;
};

export type WorldTimeCarryRule = {
  from: string;
  to: string;
  base: number;
};

export type WorldTimeCalendarUnit = {
  unit: string;
  value?: string;
};

export type WorldTimeCalendar = {
  enabled: boolean;
  calendar_name?: string;
  units?: WorldTimeCalendarUnit[];
};

export type WorldTimeUnitSequence = {
  unit: string;
  values?: string[];
};

export type WorldTimeSettings = {
  tick_scale_mode?: string;
  tick_min_unit?: string;
  tick_step?: number;
  tick_units?: string[];
  time_scale_carry?: WorldTimeCarryRule[];
  time_calendar?: WorldTimeCalendar;
  unit_value_sequences?: WorldTimeUnitSequence[];
};

export type WorldSettings = {
  world_id: string;
  memory_limit: number;
  max_analysis_rounds: number;
  max_context_depth: number;
  auto_apply: boolean;
  require_review_above: string;
  propagation_max_depth: number;
  enable_propagation_machine: boolean;
  sub_task_max_retries: number;
  sub_task_timeout_secs: number;
  pipeline_mode: string;
  world_time_settings?: WorldTimeSettings;
};

export type StateComponentEnvelope = {
  component_type: string;
  component?: Record<string, unknown>;
  data?: unknown;
};

export type StateComponentsResponse = {
  world_id: string;
  components: StateComponentEnvelope[];
};

export type StateComponentResponse = {
  world_id: string;
  state_component: StateComponentEnvelope;
};

export type TimelineTick = {
  id: string;
  world_id: string;
  tick_number: number;
  tick_type: string;
  game_time?: string;
  summary?: string;
  data?: string;
  future_outline?: string;
  created_at: string;
};

export type TimelineEnvelope = {
  tick_number: number;
  tick_type: string;
  game_time?: string;
  advanced_ticks?: number;
  summary?: string;
  future_outline?: string;
  timeline: TimelineTick;
  data?: unknown;
};

export type TimelinesResponse = {
  world_id: string;
  timelines: TimelineEnvelope[];
};

export type LatestTimelineResponse = {
  world_id: string;
  timeline: TimelineEnvelope;
};

export type TickResponse = {
  tick?: TimelineTick;
  invoke?: InvokeResponse;
  advanced_ticks?: number;
  world_time_state?: Record<string, unknown>;
  autonomous_runs?: Array<Record<string, unknown>>;
};

export type VersionInfo = {
  version: string;
  min_compatible: string;
};

export type RequestOptions = {
  headers?: Record<string, string>;
  signal?: AbortSignal;
};

