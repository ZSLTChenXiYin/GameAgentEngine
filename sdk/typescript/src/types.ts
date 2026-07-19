// Auto-generated from Go SDK types

export interface Node {
  id: string;
  world_id: string;
  name: string;
  node_type: string;
  parent_id?: string;
  created_at: string;
  updated_at: string;
}

export interface Component {
  id: string;
  node_id: string;
  component_type: string;
  data: string;
  created_at: string;
  updated_at: string;
}

export interface ComponentMeta {
  type: string;
  validation_mode: string;
  data_format: string;
  help_text?: string;
}

export interface Relation {
  id: string;
  world_id: string;
  source_id: string;
  target_id: string;
  relation_type: string;
  weight: number;
  properties?: string;
  created_at: string;
  updated_at: string;
}

export interface Memory {
  id: string;
  node_id: string;
  content: string;
  level: string;
  tags: string;
  created_at: string;
  updated_at: string;
}

export interface NodeDetail {
  node: Node;
  components?: Component[];
  memories?: Memory[];
}

export interface InteractionContext {
  mode?: string;
  speaker_node_id?: string;
  target_node_id?: string;
  scene_node_id?: string;
  room_id?: string;
  participant_node_ids?: string[];
  audience_scope?: string;
  turn_index?: number;
  event?: InteractionEvent;
}

export interface InteractionEvent {
  type: string;
  data?: Record<string, unknown>;
}

export interface InvokeContext {
  include_related_nodes?: boolean;
  max_analysis_rounds?: number;
  max_depth?: number;
  pipeline_mode?: string;
  dynamic_interfaces?: DynamicInterface[];
}

export interface DynamicInterface {
  interface_id: string;
  max_calls?: number;
}

export interface InvokeRequest {
  world_id: string;
  task_type: string;
  node_id: string;
  message?: string;
  participant_node_ids?: string[];
  mode?: string;
  audience_scope?: string;
  turn_index?: number;
  event?: InteractionEvent;
  context?: InvokeContext;
}

export interface InvokeResponse {
  request_id: string;
  world_id?: string;
  node_id?: string;
  task_type: string;
  reply: string;
  execution_mode: string;
  metadata?: ResponseMeta;
  action_calls?: ActionCallResult[];
  memory_updates?: MemoryUpdate[];
  advanced_ticks?: number;
}

export interface ResponseMeta {
  model_name: string;
  total_tokens: number;
  completion_tokens: number;
  duration_ms: number;
}

export interface ActionCallResult {
  action_id: string;
  api_name: string;
  args: Record<string, unknown>;
  result?: string;
  error?: string;
}

export interface MemoryUpdate {
  node_id: string;
  content: string;
  level: string;
  tags: string;
}

export interface TickResponse {
  request_id: string;
  reply: string;
  summary?: string;
  future_outline?: string;
  autonomous_results?: AutonomousRunResult[];
}

export interface AutonomousRunResult {
  node_id: string;
  response?: InvokeResponse;
  error?: string;
}

export interface AutonomousConfig {
  enabled: boolean;
  trigger: string;
  interval_seconds?: number;
  priority?: number;
  cooldown_seconds?: number;
  status?: string;
  capabilities?: AgentCapability[];
}

export interface AgentCapability {
  id: string;
  description?: string;
  schema?: Record<string, unknown>;
}

export interface PlayerInputInterpretRequest {
  world_id: string;
  target_node_id?: string;
  actor_node_id?: string;
  session_id?: string;
  message: string;
  participant_node_ids?: string[];
  context?: InvokeContext;
}

export interface InteractionExecuteRequest {
  world_id: string;
  actor_node_id: string;
  target_node_id: string;
  task_type: string;
  message: string;
  context?: InvokeContext;
}

export interface CreateNodeRequest {
  name: string;
  node_type: string;
  parent_id?: string;
}

export interface PolicyCheckRequest {
  world_id: string;
  actor_node_id: string;
  action: string;
  target_node_id?: string;
}

export interface PolicyCheckResult {
  allowed: boolean;
  reason?: string;
  constraint?: string;
}

export interface WorldChangePlan {
  impact_level?: string;
  summary?: string;
  world_events?: WorldEvent[];
  proposed_actions?: ProposedAction[];
}

export interface WorldEvent {
  event_type: string;
  scope?: string;
  description: string;
  confidence: number;
}

export interface ProposedAction {
  api_name: string;
  args: Record<string, unknown>;
}

export interface TickInfo {
  id: string;
  world_id: string;
  tick_number: number;
  summary: string;
  created_at: string;
}

export interface LogEntry {
  id: string;
  world_id: string;
  level: string;
  message: string;
  created_at: string;
}

export interface WorldSnapshotInfo {
  id: string;
  name?: string;
  size_bytes: number;
  created_at: string;
}

export interface SnapshotValidationResult {
  valid: boolean;
  errors?: string[];
}

export const RelationType = {
  BelongsTo: "belongs_to",
  Ally: "ally",
  Enemy: "enemy",
  Subordinate: "subordinate",
  Kinship: "kinship",
  LocatedAt: "located_at",
  ExternalParent: "external_parent",
} as const;

export const PipelineMode = {
  Vertical: "vertical",
  Polling: "polling",
  Full: "full",
} as const;

export const PropagationMode = {
  Upward: "upward",
  Environment: "environment_scope",
  Organization: "organization_scope",
  TagBroadcast: "tag_broadcast",
  Targeted: "targeted",
  Manual: "manual",
} as const;

export const TaskType = {
  NPCDialogue: "npc_dialogue",
  WorldTick: "world_tick",
  AutonomousAct: "autonomous_act",
  WorldEvent: "world_event",
  PlayerInputInterpret: "player_input_interpret",
  Custom: "custom",
} as const;

export const AutonomousTrigger = {
  Manual: "manual",
  WorldTickSync: "world_tick_sync",
  Scheduled: "scheduled",
} as const;

export const AutonomousStatus = {
  Idle: "idle",
  Running: "running",
  Completed: "completed",
  Failed: "failed",
} as const;
