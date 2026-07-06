package sdk

import "time"

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
	IncludeRelatedNodes bool   `json:"include_related_nodes,omitempty"` // 是否加载关联节点的数据
	PipelineMode        string `json:"pipeline_mode,omitempty"`         // 管线模式：vertical/polling/full
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
	Propagation *PropagationRule `json:"propagation,omitempty"`
}

// WorldChangePlan 描述一次世界刻推进产生的变更计划。
type WorldChangePlan struct {
	TimelineID      string           `json:"timeline_id"`
	ImpactLevel     string           `json:"impact_level"`
	Summary         string           `json:"summary"`
	WorldEvents     []PlanEvent      `json:"world_events"`
	ProposedActions []ProposedAction `json:"proposed_actions"`
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
	ID           string `json:"id"`
	WorldID      string `json:"world_id"`
	TaskType     string `json:"task_type"`
	NodeID       string `json:"node_id"`
	RequestData  string `json:"request_data,omitempty"`
	ResponseData string `json:"response_data,omitempty"`
	LLMModel     string `json:"llm_model"`
	TokensUsed   int    `json:"tokens_used"`
	DurationMs   int64  `json:"duration_ms"`
	CreatedAt    string `json:"created_at"`
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
	Mode          string   `json:"mode,omitempty"`
	TargetTags    []string `json:"target_tags,omitempty"`
	TargetNodeIDs []string `json:"target_node_ids,omitempty"`
	MaxDepth      int      `json:"max_depth,omitempty"`
	PublishUp     bool     `json:"publish_up,omitempty"`
}

// WorldSettings 表示世界的运行设置。
type WorldSettings struct {
	WorldID                  string `json:"world_id"`
	MemoryLimit              int    `json:"memory_limit"`
	MaxAnalysisRounds        int    `json:"max_analysis_rounds"`
	MaxContextDepth          int    `json:"max_context_depth"`
	AutoApply                bool   `json:"auto_apply"`
	RequireReviewAbove       string `json:"require_review_above"`
	PropagationMaxDepth      int    `json:"propagation_max_depth"`
	EnablePropagationMachine bool   `json:"enable_propagation_machine"`
	SubTaskMaxRetries        int    `json:"sub_task_max_retries"`
	SubTaskTimeoutSecs       int    `json:"sub_task_timeout_secs"`
	PipelineMode             string `json:"pipeline_mode"`
}

// WorldSettingsUpdate represents a partial world settings update request.
type WorldSettingsUpdate struct {
	MemoryLimit              *int    `json:"memory_limit,omitempty"`
	MaxAnalysisRounds        *int    `json:"max_analysis_rounds,omitempty"`
	MaxContextDepth          *int    `json:"max_context_depth,omitempty"`
	AutoApply                *bool   `json:"auto_apply,omitempty"`
	RequireReviewAbove       *string `json:"require_review_above,omitempty"`
	PropagationMaxDepth      *int    `json:"propagation_max_depth,omitempty"`
	EnablePropagationMachine *bool   `json:"enable_propagation_machine,omitempty"`
	SubTaskMaxRetries        *int    `json:"sub_task_max_retries,omitempty"`
	SubTaskTimeoutSecs       *int    `json:"sub_task_timeout_secs,omitempty"`
	PipelineMode             *string `json:"pipeline_mode,omitempty"`
}

// NodeDetail 表示节点详情接口返回的聚合结构。
type NodeDetail struct {
	Node       Node        `json:"node"`
	Components []Component `json:"components,omitempty"`
	Memories   []Memory    `json:"memories,omitempty"`
	Children   []Node      `json:"children,omitempty"`
	Relations  []Relation  `json:"relations,omitempty"`
}

// SubTaskDeclaration 描述 LLM 声明的一个子任务。
type SubTaskDeclaration struct {
	Label     string   `json:"label"`
	TaskType  string   `json:"task_type"`
	NodeID    string   `json:"node_id"`
	DependsOn []string `json:"depends_on,omitempty"`
	MergeMode string   `json:"merge_mode,omitempty"`
}
