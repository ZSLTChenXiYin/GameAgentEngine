package sdk

import "time"

const (
	RelationBelongsTo      = "belongs_to"
	RelationAlly           = "ally"
	RelationEnemy          = "enemy"
	RelationSubordinate    = "subordinate"
	RelationKinship        = "kinship"
	RelationLocatedAt      = "located_at"
	RelationExternalParent = "external_parent"
)

const (
	PipelineModeVertical = "vertical"
	PipelineModePolling  = "polling"
	PipelineModeFull     = "full"
)

const (
	PropagationModeUpward       = "upward"
	PropagationModeEnvironment  = "environment_scope"
	PropagationModeOrganization = "organization_scope"
	PropagationModeTagBroadcast = "tag_broadcast"
	PropagationModeTargeted     = "targeted"
	PropagationModeManual       = "manual"
)

// Node 表示 Agent API 返回的节点实体。
type Node struct {
	ID        string    `json:"id"`
	WorldID   string    `json:"world_id"`
	Name      string    `json:"name"`
	NodeType  string    `json:"node_type"`
	ParentID  *string   `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Component 表示挂载在节点上的组件。
type Component struct {
	ID            string `json:"id"`
	NodeID        string `json:"node_id"`
	ComponentType string `json:"component_type"`
	Data          string `json:"data"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Relation 表示两个节点之间的有向关系。
// 所有关系都严格按 source -> target 解读；SDK 调用方不应自行反向猜语义。
type Relation struct {
	ID           string `json:"id"`
	WorldID      string `json:"world_id"`
	SourceID     string `json:"source_id"`
	TargetID     string `json:"target_id"`
	RelationType string `json:"relation_type"`
	Weight       int    `json:"weight"`
	Properties   string `json:"properties,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// Memory 表示节点的一条持久化记忆。
type Memory struct {
	ID        string `json:"id"`
	NodeID    string `json:"node_id"`
	Content   string `json:"content"`
	Level     string `json:"level"`
	Tags      string `json:"tags,omitempty"`
	CreatedAt string `json:"created_at"`
}

// ChatMessage 表示对话中的一条消息。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// InvokeContext 表示调用方希望追加的上下文约束，可在请求层面覆盖服务端配置。
type InvokeContext struct {
	MaxAnalysisRounds   int    `json:"max_analysis_rounds,omitempty"`   // LLM 内部轮询最大次数（0 表示使用服务端配置）
	MaxDepth            int    `json:"max_depth,omitempty"`             // 上下文向上追溯的最大深度（0 表示使用服务端配置）
	MemoryLimit         int    `json:"memory_limit,omitempty"`          // 每次推理最多加载的记忆数量（0 表示使用服务端配置）
	IncludeRelatedNodes bool   `json:"include_related_nodes,omitempty"` // 是否启用受控关系补充；这不是“把所有邻接关系节点全部塞进上下文”的开关。
	PipelineMode        string `json:"pipeline_mode,omitempty"`         // 管线模式：vertical/polling/full；也决定关系图谱装配强度。
}

// InvokeRequest 是 SDK 侧的统一推理请求结构。
type InvokeRequest struct {
	Context   *InvokeContext `json:"context,omitempty"`
	WorldID   string         `json:"world_id"`
	TaskType  string         `json:"task_type"`
	NodeID    string         `json:"node_id"`
	SessionID string         `json:"session_id,omitempty"`
	Messages  []ChatMessage  `json:"messages,omitempty"`
	Event     *WorldEvent    `json:"event,omitempty"`
}

// WorldEvent 表示待评估的世界事件。
type WorldEvent struct {
	EventType   string `json:"event_type"`
	ScopeID     string `json:"scope_id"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// InvokeResponse 是 SDK 侧的统一推理响应结构。
type InvokeResponse struct {
	RequestID       string               `json:"request_id"`
	TaskType        string               `json:"task_type"`
	ExecutionMode   string               `json:"execution_mode"`
	Reply           string               `json:"reply,omitempty"`
	AdvancedTicks   int                  `json:"advanced_ticks,omitempty"`
	ActionCalls     []ActionCall         `json:"action_calls,omitempty"`
	WorldChangePlan *WorldChangePlan     `json:"world_change_plan,omitempty"`
	MemoryUpdates   []MemoryUpdate       `json:"memory_updates,omitempty"`
	SubTasks        []SubTaskDeclaration `json:"sub_tasks,omitempty"`
	Metadata        *ResponseMeta        `json:"metadata,omitempty"`
}

// AgentCapability 描述某个自主节点被显式授权调用的能力。
type AgentCapability struct {
	ID          string         `json:"id"`
	Mode        string         `json:"mode,omitempty"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
}

// AutonomousConfig 描述挂载到节点 autonomous 组件中的自主行为配置。
type AutonomousConfig struct {
	Enabled         bool              `json:"enabled"`
	Trigger         string            `json:"trigger"`
	IntervalSeconds int               `json:"interval_seconds,omitempty"`
	Capabilities    []AgentCapability `json:"capabilities,omitempty"`
	LastRunAt       string            `json:"last_run_at,omitempty"`
	LastError       string            `json:"last_error,omitempty"`
}

// AutonomousConfigResponse 是自主行为配置 API 的返回结构。
type AutonomousConfigResponse struct {
	Component *Component        `json:"component,omitempty"`
	Config    *AutonomousConfig `json:"config,omitempty"`
}

// AutonomousRunResult 描述一次自主节点触发的执行结果。
type AutonomousRunResult struct {
	NodeID   string          `json:"node_id"`
	Response *InvokeResponse `json:"response,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// ActionCall 表示一条待执行动作调用。
type ActionCall struct {
	ActionID   string         `json:"action_id"`
	Args       map[string]any `json:"args"`
	Mode       string         `json:"mode,omitempty"`
	CallbackID string         `json:"callback_id,omitempty"`
}

// MemoryUpdate 表示一条记忆更新结果。
type MemoryUpdate struct {
	NodeID      string           `json:"node_id"`
	Content     string           `json:"content"`
	Level       string           `json:"level"`
	Tags        string           `json:"tags,omitempty"`
	Propagation *PropagationRule `json:"propagation,omitempty"` // nil 表示默认 upward；显式环境/组织传播必须通过 propagation 指定。
}

// WorldChangePlan 描述一次世界刻推进产生的变更计划。
type WorldChangePlan struct {
	TimelineID      string           `json:"timeline_id"`
	ImpactLevel     string           `json:"impact_level"`
	Summary         string           `json:"summary"`
	WorldEvents     []PlanEvent      `json:"world_events"`
	ProposedActions []ProposedAction `json:"proposed_actions"`
}

// PendingPlan 表示一条待审批的世界变更计划。
type PendingPlan struct {
	PlanID          string           `json:"plan_id"`
	WorldID         string           `json:"world_id"`
	TickNumber      int              `json:"tick_number"`
	TaskType        string           `json:"task_type"`
	WorldChangePlan *WorldChangePlan `json:"world_change_plan"`
	ActionCalls     []ActionCall     `json:"action_calls,omitempty"`
	MemoryUpdates   []MemoryUpdate   `json:"memory_updates,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	Status          string           `json:"status"`
}

// PlanDecisionResponse 表示计划审批接口返回结果。
type PlanDecisionResponse struct {
	Status string       `json:"status"`
	PlanID string       `json:"plan_id"`
	Plan   *PendingPlan `json:"plan,omitempty"`
}

// PlanEvent 描述计划中的世界事件。
type PlanEvent struct {
	EventType   string  `json:"event_type"`
	Scope       string  `json:"scope"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// ProposedAction 描述计划中的建议动作。
type ProposedAction struct {
	APIName string         `json:"api_name"`
	Args    map[string]any `json:"args"`
}

// ResponseMeta 描述一次推理调用的元信息。
type ResponseMeta struct {
	LLMModel               string `json:"llm_model"`
	TokensUsed             int    `json:"tokens_used"`
	ProcessingTimeMs       int64  `json:"processing_time_ms"`
	ConfiguredPipelineMode string `json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode  string `json:"effective_pipeline_mode,omitempty"`
	MaxAnalysisRounds      int    `json:"max_analysis_rounds,omitempty"`
	RoundsUsed             int    `json:"rounds_used,omitempty"`
}

// InferenceLog 表示一次推理调用的服务端日志记录。
type InferenceLog struct {
	ID                     string `json:"id"`
	WorldID                string `json:"world_id"`
	TaskType               string `json:"task_type"`
	NodeID                 string `json:"node_id"`
	Category               string `json:"category,omitempty"`
	EventName              string `json:"event_name,omitempty"`
	LogLevel               string `json:"log_level,omitempty"`
	Message                string `json:"message,omitempty"`
	RequestID              string `json:"request_id,omitempty"`
	ExecutionMode          string `json:"execution_mode,omitempty"`
	ConfiguredPipelineMode string `json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode  string `json:"effective_pipeline_mode,omitempty"`
	Round                  int    `json:"round,omitempty"`
	RequestData            string `json:"request_data,omitempty"`
	ResponseData           string `json:"response_data,omitempty"`
	DetailData             string `json:"detail_data,omitempty"`
	LLMModel               string `json:"llm_model"`
	TokensUsed             int    `json:"tokens_used"`
	DurationMs             int64  `json:"duration_ms"`
	CreatedAt              string `json:"created_at"`
}

// InferenceLogQuery describes server-side log filters.
type InferenceLogQuery struct {
	WorldID       string
	NodeID        string
	TaskType      string
	Category      string
	EventName     string
	ExecutionMode string
	RequestID     string
	Round         int
	Limit         int
	Offset        int
}

// DebugTrace 表示一条引擎调试轨迹。
type DebugTrace struct {
	ID                     string `json:"id"`
	WorldID                string `json:"world_id"`
	RequestID              string `json:"request_id"`
	TaskType               string `json:"task_type"`
	NodeID                 string `json:"node_id"`
	ConfiguredPipelineMode string `json:"configured_pipeline_mode"`
	EffectivePipelineMode  string `json:"effective_pipeline_mode"`
	MaxAnalysisRounds      int    `json:"max_analysis_rounds"`
	RoundsUsed             int    `json:"rounds_used"`
	Timestamp              string `json:"timestamp"`
	DurationMs             int64  `json:"duration_ms"`
	Error                  string `json:"error"`
}

// DebugTraceList 表示调试轨迹列表接口返回结果。
type DebugTraceList struct {
	Traces []DebugTrace `json:"traces"`
	Count  int          `json:"count"`
}

// ImportResult 表示 creator/import 接口返回的导入或纯校验摘要。
type ImportResult struct {
	WorldID        string `json:"world_id,omitempty"`
	WorldName      string `json:"world_name"`
	DryRun         bool   `json:"dry_run"`
	NodeCount      int    `json:"node_count"`
	ComponentCount int    `json:"component_count"`
	MemoryCount    int    `json:"memory_count"`
	RelationCount  int    `json:"relation_count"`
}

// WorldPolicy 表示世界的动作策略。
type SnapshotValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SnapshotValidationResult struct {
	SnapshotWorldID             string                    `json:"snapshot_world_id"`
	SourceWorldID               string                    `json:"source_world_id"`
	SnapshotName                string                    `json:"snapshot_name"`
	Reason                      string                    `json:"reason"`
	Valid                       bool                      `json:"valid"`
	SchemaVersion               string                    `json:"schema_version"`
	EngineVersion               string                    `json:"engine_version"`
	MinCompatibleVersion        string                    `json:"min_compatible_version"`
	CurrentEngineVersion        string                    `json:"current_engine_version"`
	CurrentMinCompatibleVersion string                    `json:"current_min_compatible_version"`
	NodeCount                   int                       `json:"node_count"`
	ComponentCount              int                       `json:"component_count"`
	MemoryCount                 int                       `json:"memory_count"`
	RelationCount               int                       `json:"relation_count"`
	SavedComponentTypes         []string                  `json:"saved_component_types"`
	CurrentComponentTypes       []string                  `json:"current_component_types"`
	SavedSettingsHash           string                    `json:"saved_settings_hash"`
	CurrentSettingsHash         string                    `json:"current_settings_hash"`
	SavedPolicyHash             string                    `json:"saved_policy_hash"`
	CurrentPolicyHash           string                    `json:"current_policy_hash"`
	Issues                      []SnapshotValidationIssue `json:"issues,omitempty"`
}

type WorldSnapshotInfo struct {
	ID                   string   `json:"id"`
	SourceWorldID        string   `json:"source_world_id"`
	SnapshotWorldID      string   `json:"snapshot_world_id"`
	SnapshotName         string   `json:"snapshot_name"`
	Reason               string   `json:"reason"`
	Status               string   `json:"status"`
	Restorable           bool     `json:"restorable"`
	EngineVersion        string   `json:"engine_version"`
	MinCompatibleVersion string   `json:"min_compatible_version"`
	SchemaVersion        string   `json:"schema_version"`
	NodeCount            int      `json:"node_count"`
	ComponentCount       int      `json:"component_count"`
	MemoryCount          int      `json:"memory_count"`
	RelationCount        int      `json:"relation_count"`
	ComponentTypes       []string `json:"component_types"`
	SettingsHash         string   `json:"settings_hash"`
	PolicyHash           string   `json:"policy_hash"`
	PayloadHash          string   `json:"payload_hash"`
	ValidationIssues     []string `json:"validation_issues,omitempty"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
}

type WorldPolicy struct {
	WorldID        string   `json:"world_id"`
	BlockedActions []string `json:"blocked_actions"`
	SafeActions    []string `json:"safe_actions"`
}

// PropagationRule 描述记忆传播规则。
type PropagationRule struct {
	Mode          string   `json:"mode,omitempty"`            // upward / environment_scope / organization_scope / tag_broadcast / targeted / manual
	TargetTags    []string `json:"target_tags,omitempty"`     // tag_broadcast 模式：按标签扩散
	TargetNodeIDs []string `json:"target_node_ids,omitempty"` // targeted 模式：定向目标节点
	MaxDepth      int      `json:"max_depth,omitempty"`       // 对结构传播模式表示目标节点起始后的祖先展开深度
	PublishUp     bool     `json:"publish_up,omitempty"`      // 仅影响 upward 主父链传播的更高层发布行为，不改变模式语义
}

func ValidRelationTypes() []string {
	return []string{RelationBelongsTo, RelationAlly, RelationEnemy, RelationSubordinate, RelationKinship, RelationLocatedAt, RelationExternalParent}
}

func ValidPropagationModes() []string {
	return []string{PropagationModeUpward, PropagationModeEnvironment, PropagationModeOrganization, PropagationModeTagBroadcast, PropagationModeTargeted, PropagationModeManual}
}

func ValidPipelineModes() []string {
	return []string{PipelineModeVertical, PipelineModePolling, PipelineModeFull}
}

// WorldSettings 表示世界的运行设置。
type WorldSettings struct {
	WorldID                  string             `json:"world_id"`
	MemoryLimit              int                `json:"memory_limit"`
	MaxAnalysisRounds        int                `json:"max_analysis_rounds"`
	MaxContextDepth          int                `json:"max_context_depth"`
	AutoApply                bool               `json:"auto_apply"`
	RequireReviewAbove       string             `json:"require_review_above"`
	PropagationMaxDepth      int                `json:"propagation_max_depth"`
	EnablePropagationMachine bool               `json:"enable_propagation_machine"`
	SubTaskMaxRetries        int                `json:"sub_task_max_retries"`
	SubTaskTimeoutSecs       int                `json:"sub_task_timeout_secs"`
	PipelineMode             string             `json:"pipeline_mode"`
	WorldTimeSettings        *WorldTimeSettings `json:"world_time_settings,omitempty"`
}

// WorldSettingsUpdate represents a partial world settings update request.
type WorldSettingsUpdate struct {
	MemoryLimit              *int               `json:"memory_limit,omitempty"`
	MaxAnalysisRounds        *int               `json:"max_analysis_rounds,omitempty"`
	MaxContextDepth          *int               `json:"max_context_depth,omitempty"`
	AutoApply                *bool              `json:"auto_apply,omitempty"`
	RequireReviewAbove       *string            `json:"require_review_above,omitempty"`
	PropagationMaxDepth      *int               `json:"propagation_max_depth,omitempty"`
	EnablePropagationMachine *bool              `json:"enable_propagation_machine,omitempty"`
	SubTaskMaxRetries        *int               `json:"sub_task_max_retries,omitempty"`
	SubTaskTimeoutSecs       *int               `json:"sub_task_timeout_secs,omitempty"`
	PipelineMode             *string            `json:"pipeline_mode,omitempty"`
	WorldTimeSettings        *WorldTimeSettings `json:"world_time_settings,omitempty"`
}

// WorldTimeSettings describes the engine-enforced world time configuration.
type WorldTimeSettings struct {
	TickScaleMode     string                  `json:"tick_scale_mode,omitempty"`
	TickMinUnit       string                  `json:"tick_min_unit,omitempty"`
	TickStep          int                     `json:"tick_step,omitempty"`
	TickUnits         []string                `json:"tick_units,omitempty"`
	TimeScaleCarry    []WorldTimeCarryRule    `json:"time_scale_carry,omitempty"`
	TimeCalendar      *WorldTimeCalendar      `json:"time_calendar,omitempty"`
	UnitValueSequence []WorldTimeUnitSequence `json:"unit_value_sequences,omitempty"`
}

// WorldTimeCarryRule describes one adjacent carry relationship between configured time units.
type WorldTimeCarryRule struct {
	From string `json:"from"`
	To   string `json:"to"`
	Base int    `json:"base"`
}

// WorldTimeCalendar describes the optional named calendar template for one world.
type WorldTimeCalendar struct {
	Enabled      bool                    `json:"enabled"`
	CalendarName string                  `json:"calendar_name,omitempty"`
	Units        []WorldTimeCalendarUnit `json:"units,omitempty"`
}

// WorldTimeCalendarUnit stores one configured calendar unit and its current value.
type WorldTimeCalendarUnit struct {
	Unit  string `json:"unit"`
	Value string `json:"value,omitempty"`
}

// WorldTimeUnitSequence stores ordered symbolic values for a unit such as 时辰.
type WorldTimeUnitSequence struct {
	Unit   string   `json:"unit"`
	Values []string `json:"values,omitempty"`
}

// WorldTimeState describes the persisted runtime world time state after one world tick.
type WorldTimeState struct {
	TickScaleMode     string                  `json:"tick_scale_mode,omitempty"`
	TickMinUnit       string                  `json:"tick_min_unit,omitempty"`
	TickStep          int                     `json:"tick_step,omitempty"`
	TickUnits         []string                `json:"tick_units,omitempty"`
	CalendarName      string                  `json:"calendar_name,omitempty"`
	CurrentUnits      []WorldTimeCalendarUnit `json:"current_units,omitempty"`
	CurrentTimeLabel  string                  `json:"current_time_label,omitempty"`
	TotalTicks        int                     `json:"total_ticks,omitempty"`
	LastTickNumber    int                     `json:"last_tick_number,omitempty"`
	LastTickType      string                  `json:"last_tick_type,omitempty"`
	LastAdvancedTicks int                     `json:"last_advanced_ticks,omitempty"`
	Metadata          map[string]any          `json:"metadata,omitempty"`
}

// StateComponentEnvelope describes one engine-recognized continuity component.
type StateComponentEnvelope struct {
	ComponentType string     `json:"component_type"`
	Component     *Component `json:"component,omitempty"`
	Data          any        `json:"data,omitempty"`
}

// StateComponentsResponse is the list response for world continuity components.
type StateComponentsResponse struct {
	WorldID    string                   `json:"world_id"`
	Components []StateComponentEnvelope `json:"components"`
}

// StateComponentResponse is the single-component response for world continuity state.
type StateComponentResponse struct {
	WorldID        string                 `json:"world_id"`
	StateComponent StateComponentEnvelope `json:"state_component"`
}

// TimelineTick describes one persisted world tick archive entry.
type TimelineTick struct {
	ID            string `json:"id"`
	WorldID       string `json:"world_id"`
	TickNumber    int    `json:"tick_number"`
	TickType      string `json:"tick_type"`
	GameTime      string `json:"game_time,omitempty"`
	Summary       string `json:"summary,omitempty"`
	Data          string `json:"data,omitempty"`
	FutureOutline string `json:"future_outline,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// TimelineEnvelope provides parsed timeline payload together with the raw row.
type TimelineEnvelope struct {
	TickNumber    int          `json:"tick_number"`
	TickType      string       `json:"tick_type"`
	GameTime      string       `json:"game_time,omitempty"`
	AdvancedTicks int          `json:"advanced_ticks,omitempty"`
	Summary       string       `json:"summary,omitempty"`
	FutureOutline string       `json:"future_outline,omitempty"`
	Timeline      TimelineTick `json:"timeline"`
	Data          any          `json:"data,omitempty"`
}

// TimelinesResponse is the list response for world tick archives.
type TimelinesResponse struct {
	WorldID   string             `json:"world_id"`
	Timelines []TimelineEnvelope `json:"timelines"`
}

// LatestTimelineResponse is the latest-entry response for world tick archives.
type LatestTimelineResponse struct {
	WorldID  string           `json:"world_id"`
	Timeline TimelineEnvelope `json:"timeline"`
}

// ContinuityBundleOptions controls how much continuity debugging context to load.
type ContinuityBundleOptions struct {
	TimelineLimit int
	LogLimit      int
	TraceLimit    int
	SkipLogs      bool
	SkipTraces    bool
	LogQuery      *InferenceLogQuery
}

// ContinuityBundle aggregates the core artifacts used to inspect world tick continuity.
type ContinuityBundle struct {
	WorldID         string                   `json:"world_id"`
	LatestTimeline  *TimelineEnvelope        `json:"latest_timeline,omitempty"`
	Timelines       []TimelineEnvelope       `json:"timelines,omitempty"`
	StateComponents []StateComponentEnvelope `json:"state_components,omitempty"`
	Logs            []InferenceLog           `json:"logs,omitempty"`
	Traces          []DebugTrace             `json:"traces,omitempty"`
}

// NodeDetail 表示节点详情接口返回的聚合结构。
type NodeDetail struct {
	Node                     Node                      `json:"node"`
	Components               []Component               `json:"components,omitempty"`
	Memories                 []Memory                  `json:"memories,omitempty"`
	Children                 []Node                    `json:"children,omitempty"`
	Relations                []Relation                `json:"relations,omitempty"`
	RelationValidationIssues []RelationValidationIssue `json:"relation_validation_issues,omitempty"`
	GraphContextPreview      *GraphContextPreview      `json:"graph_context_preview,omitempty"`
}

type RelationValidationIssue struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type GraphContextPreview struct {
	PrimaryParentChain []string `json:"primary_parent_chain,omitempty"`
	EnvironmentChain   []string `json:"environment_chain,omitempty"`
	OrganizationChains []string `json:"organization_chains,omitempty"`
	SocialLinks        []string `json:"social_links,omitempty"`
	AuxiliaryScopes    []string `json:"auxiliary_scopes,omitempty"`
	Summary            []string `json:"summary,omitempty"`
}

// SubTaskDeclaration 描述 LLM 声明的一个子任务。
type SubTaskDeclaration struct {
	Label     string   `json:"label"`
	TaskType  string   `json:"task_type"`
	NodeID    string   `json:"node_id"`
	DependsOn []string `json:"depends_on,omitempty"`
	MergeMode string   `json:"merge_mode,omitempty"`
}

// RuntimeTask describes a runtime external interaction task.
// Used in Pull mode for game clients to claim and track tasks.
type RuntimeTask struct {
	TaskID       string `json:"task_id"`
	Category     string `json:"category,omitempty"`
	Status       string `json:"status"`
	Consumer     string `json:"consumer,omitempty"`
	WorldID      string `json:"world_id,omitempty"`
	NodeID       string `json:"node_id,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
	CallbackID   string `json:"callback_id,omitempty"`
	LeaseToken   string `json:"lease_token,omitempty"`
	LeaseOwner   string `json:"lease_owner,omitempty"`
	AttemptCount int    `json:"attempt_count,omitempty"`
	MaxAttempts  int    `json:"max_attempts,omitempty"`
	Priority     int    `json:"priority,omitempty"`
	PayloadJSON  string `json:"payload_json,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}
