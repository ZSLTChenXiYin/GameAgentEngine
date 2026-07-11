package store

import (
	"time"

	"gorm.io/gorm"
)

// NodeModel 是节点实体在数据库中的持久化结构。
type NodeModel struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID       string         `gorm:"uniqueIndex;size:36;not null" json:"id"`
	WorldID    int64          `gorm:"index;not null" json:"-"`
	WorldUUID  string         `gorm:"-" json:"world_id"`
	Name       string         `gorm:"size:255;not null" json:"name"`
	NodeType   string         `gorm:"size:50;not null;index" json:"node_type"`
	ParentID   *int64         `gorm:"index" json:"-"`
	ParentUUID *string        `gorm:"-" json:"parent_id,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (NodeModel) TableName() string { return "nodes" }

// ComponentModel 是节点组件的持久化结构。
type ComponentModel struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID          string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	NodeID        int64     `gorm:"index;not null" json:"-"`
	NodeUUID      string    `gorm:"-" json:"node_id"`
	ComponentType string    `gorm:"size:50;not null;index" json:"component_type"`
	Data          string    `gorm:"type:text;not null" json:"data"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (ComponentModel) TableName() string { return "components" }

// RelationModel 是节点之间有向关系的持久化结构。
type RelationModel struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID         string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	WorldID      int64     `gorm:"index;not null" json:"-"`
	WorldUUID    string    `gorm:"-" json:"world_id"`
	SourceID     int64     `gorm:"index;not null" json:"-"`
	SourceUUID   string    `gorm:"-" json:"source_id"`
	TargetID     int64     `gorm:"index;not null" json:"-"`
	TargetUUID   string    `gorm:"-" json:"target_id"`
	RelationType string    `gorm:"size:50;not null;index" json:"relation_type"`
	Weight       int       `gorm:"default:0" json:"weight"`
	Properties   string    `gorm:"type:text" json:"properties,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RelationModel) TableName() string { return "relations" }

// MemoryModel 是节点记忆的持久化结构。
type MemoryModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	NodeID    int64     `gorm:"index;not null" json:"-"`
	NodeUUID  string    `gorm:"-" json:"node_id"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Level     string    `gorm:"size:20;not null;default:long_term" json:"level"`
	Tags      string    `gorm:"size:500" json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (MemoryModel) TableName() string { return "memories" }

// TimelineModel 是世界时间线刻度的持久化结构。
type TimelineModel struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID          string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	WorldID       int64     `gorm:"index;not null;uniqueIndex:idx_world_tick" json:"-"`
	WorldUUID     string    `gorm:"-" json:"world_id"`
	TickNumber    int       `gorm:"not null;uniqueIndex:idx_world_tick" json:"tick_number"`
	TickType      string    `gorm:"size:20;not null" json:"tick_type"`
	GameTime      string    `gorm:"size:100" json:"game_time"`
	Summary       string    `gorm:"type:text" json:"summary,omitempty"`
	Data          string    `gorm:"type:text" json:"data,omitempty"`
	FutureOutline string    `gorm:"type:text" json:"future_outline,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

func (TimelineModel) TableName() string { return "timelines" }

// InferenceLogModel 是统一运行日志的持久化结构。
type InferenceLogModel struct {
	ID                     int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID                   string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	WorldID                int64     `gorm:"index;not null" json:"-"`
	WorldUUID              string    `gorm:"-" json:"world_id"`
	TaskType               string    `gorm:"size:50;not null;index" json:"task_type"`
	NodeID                 *int64    `gorm:"index" json:"-"`
	NodeUUID               string    `gorm:"-" json:"node_id,omitempty"`
	Category               string    `gorm:"size:50;index" json:"category,omitempty"`
	EventName              string    `gorm:"size:100;index" json:"event_name,omitempty"`
	LogLevel               string    `gorm:"size:20;index" json:"log_level,omitempty"`
	Message                string    `gorm:"type:text" json:"message,omitempty"`
	RequestID              string    `gorm:"size:36;index" json:"request_id,omitempty"`
	ExecutionMode          string    `gorm:"size:20;index" json:"execution_mode,omitempty"`
	ConfiguredPipelineMode string    `gorm:"size:20" json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode  string    `gorm:"size:20" json:"effective_pipeline_mode,omitempty"`
	Round                  int       `gorm:"default:0" json:"round,omitempty"`
	RequestData            string    `gorm:"type:text" json:"request_data,omitempty"`
	ResponseData           string    `gorm:"type:text" json:"response_data,omitempty"`
	DetailData             string    `gorm:"type:text" json:"detail_data,omitempty"`
	LLMModel               string    `gorm:"size:100" json:"llm_model"`
	TokensUsed             int       `gorm:"default:0" json:"tokens_used"`
	DurationMs             int64     `gorm:"default:0" json:"duration_ms"`
	CreatedAt              time.Time `json:"created_at"`
}

func (InferenceLogModel) TableName() string { return "logs" }

// IdempotencyKeyModel 存储幂等操作的已缓存结果。
type IdempotencyKeyModel struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	Fingerprint string    `gorm:"size:64;not null" json:"fingerprint"`
	StatusCode  int       `gorm:"default:200" json:"status_code"`
	Result      string    `gorm:"type:text;not null" json:"result"`
	CreatedAt   time.Time `json:"created_at"`
}

func (IdempotencyKeyModel) TableName() string { return "idempotency_keys" }

// WorldSnapshotModel stores metadata for save-oriented world copy operations.
type WorldSnapshotModel struct {
	ID                   int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID                 string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	SourceWorldID        int64     `gorm:"index;not null" json:"-"`
	SourceWorldUUID      string    `gorm:"size:36;not null;index" json:"source_world_id"`
	SnapshotWorldID      int64     `gorm:"index;not null" json:"-"`
	SnapshotWorldUUID    string    `gorm:"size:36;not null;uniqueIndex" json:"snapshot_world_id"`
	SnapshotName         string    `gorm:"size:255;not null" json:"snapshot_name"`
	Reason               string    `gorm:"size:50;not null;default:fork_world" json:"reason"`
	EngineVersion        string    `gorm:"size:50;not null" json:"engine_version"`
	MinCompatibleVersion string    `gorm:"size:50;not null" json:"min_compatible_version"`
	SchemaVersion        string    `gorm:"size:50;not null" json:"schema_version"`
	NodeCount            int       `gorm:"default:0" json:"node_count"`
	ComponentCount       int       `gorm:"default:0" json:"component_count"`
	MemoryCount          int       `gorm:"default:0" json:"memory_count"`
	RelationCount        int       `gorm:"default:0" json:"relation_count"`
	ComponentTypesJSON   string    `gorm:"type:text;not null" json:"component_types"`
	SettingsHash         string    `gorm:"size:64;not null" json:"settings_hash"`
	PolicyHash           string    `gorm:"size:64;not null" json:"policy_hash"`
	PayloadHash          string    `gorm:"size:64;not null" json:"payload_hash"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (WorldSnapshotModel) TableName() string { return "world_snapshots" }

// WorldPolicyModel 是世界级动作策略的持久化结构。
type WorldPolicyModel struct {
	WorldID        int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	WorldUUID      string    `gorm:"uniqueIndex;size:36;not null" json:"world_id"`
	BlockedActions string    `gorm:"type:text" json:"blocked_actions"`
	SafeActions    string    `gorm:"type:text" json:"safe_actions"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (WorldPolicyModel) TableName() string { return "world_policies" }

// WorldSettingsModel 是世界级运行设置的持久化结构。
type WorldSettingsModel struct {
	WorldID                  int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	WorldUUID                string    `gorm:"uniqueIndex;size:36;not null" json:"world_id"`
	MemoryLimit              int       `gorm:"default:50" json:"memory_limit"`
	MaxAnalysisRounds        int       `gorm:"default:5" json:"max_analysis_rounds"`
	MaxContextDepth          int       `gorm:"default:3" json:"max_context_depth"`
	AutoApply                bool      `gorm:"default:true" json:"auto_apply"`
	RequireReviewAbove       string    `gorm:"size:20;default:critical" json:"require_review_above"`
	EnablePropagationMachine bool      `gorm:"default:false" json:"enable_propagation_machine"`
	PropagationMaxDepth      int       `gorm:"default:2" json:"propagation_max_depth"`
	SubTaskMaxRetries        int       `gorm:"default:2" json:"sub_task_max_retries"`
	SubTaskTimeoutSecs       int       `gorm:"default:60" json:"sub_task_timeout_secs"`
	PipelineMode             string    `gorm:"size:20;default:full" json:"pipeline_mode"`
	WorldTimeSettingsJSON    string    `gorm:"type:text" json:"world_time_settings,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func (WorldSettingsModel) TableName() string { return "world_settings" }

// PropagationChainModel 是标签传播规则链的持久化结构。
type PropagationChainModel struct {
	ID               int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID             string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	WorldID          int64     `gorm:"index;not null" json:"-"`
	WorldUUID        string    `gorm:"-" json:"world_id"`
	Name             string    `gorm:"size:255" json:"name"`
	Description      string    `gorm:"type:text" json:"description,omitempty"`
	TriggerTags      string    `gorm:"type:text;not null" json:"trigger_tags"`
	TriggerNodeTypes string    `gorm:"type:text" json:"trigger_node_types,omitempty"`
	Actions          string    `gorm:"type:text;not null" json:"actions"`
	Enabled          bool      `gorm:"default:true" json:"enabled"`
	MaxDepth         int       `gorm:"default:10" json:"max_depth"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (PropagationChainModel) TableName() string { return "propagation_chains" }

// AsyncCallbackRecordModel stores pending and completed async callback records.
type AsyncCallbackRecordModel struct {
	ID                int64      `gorm:"primaryKey;autoIncrement" json:"-"`
	CallbackID        string     `gorm:"uniqueIndex;size:64;not null" json:"callback_id"`
	ActionID          string     `gorm:"size:100;not null;index" json:"action_id"`
	Status            string     `gorm:"size:20;not null;index" json:"status"`
	NodeUUID          string     `gorm:"size:36;index" json:"node_id,omitempty"`
	WorldUUID         string     `gorm:"size:36;index" json:"world_id,omitempty"`
	RequestID         string     `gorm:"size:36;index" json:"request_id,omitempty"`
	ResumeExecutionID string     `gorm:"size:64;index" json:"resume_execution_id,omitempty"`
	ArgsJSON          string     `gorm:"type:text" json:"args_json,omitempty"`
	ResultJSON        string     `gorm:"type:text" json:"result_json,omitempty"`
	ErrorMessage      string     `gorm:"type:text" json:"error_message,omitempty"`
	PostProcessStatus string     `gorm:"size:20;index" json:"post_process_status,omitempty"`
	PostProcessResult string     `gorm:"type:text" json:"post_process_result,omitempty"`
	PostProcessError  string     `gorm:"type:text" json:"post_process_error,omitempty"`
	PostProcessedAt   *time.Time `json:"post_processed_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (AsyncCallbackRecordModel) TableName() string { return "async_callback_records" }

// PausedExecutionModel stores a resumable multi-turn execution snapshot.
type PausedExecutionModel struct {
	ID                     int64      `gorm:"primaryKey;autoIncrement" json:"-"`
	ExecutionID            string     `gorm:"uniqueIndex;size:64;not null" json:"execution_id"`
	RequestID              string     `gorm:"size:36;not null;index" json:"request_id"`
	WorldUUID              string     `gorm:"size:36;not null;index" json:"world_id"`
	NodeUUID               string     `gorm:"size:36;not null;index" json:"node_id"`
	TaskType               string     `gorm:"size:50;not null;index" json:"task_type"`
	ExecutionMode          string     `gorm:"size:20;not null" json:"execution_mode"`
	ConfiguredPipelineMode string     `gorm:"size:20" json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode  string     `gorm:"size:20" json:"effective_pipeline_mode,omitempty"`
	Status                 string     `gorm:"size:20;not null;index" json:"status"`
	PausedRound            int        `gorm:"not null" json:"paused_round"`
	MaxRounds              int        `gorm:"not null" json:"max_rounds"`
	TargetNodeID           string     `gorm:"size:36;not null" json:"target_node_id"`
	PauseReason            string     `gorm:"size:100;not null" json:"pause_reason"`
	CallbackID             string     `gorm:"size:64;index" json:"callback_id,omitempty"`
	OriginalRequestJSON    string     `gorm:"type:text;not null" json:"original_request_json"`
	BuiltContextJSON       string     `gorm:"type:text;not null" json:"built_context_json"`
	RuntimeJSON            string     `gorm:"type:text;not null" json:"runtime_json"`
	RoundStateJSON         string     `gorm:"type:text;not null" json:"round_state_json"`
	PendingDataRequestJSON string     `gorm:"type:text" json:"pending_data_request_json,omitempty"`
	ResumePayloadJSON      string     `gorm:"type:text" json:"resume_payload_json,omitempty"`
	LastError              string     `gorm:"type:text" json:"last_error,omitempty"`
	ResumedAt              *time.Time `json:"resumed_at,omitempty"`
	CompletedAt            *time.Time `json:"completed_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func (PausedExecutionModel) TableName() string { return "paused_executions" }

// RuntimeTaskModel stores pull-delivered external interaction tasks.
type RuntimeTaskModel struct {
	ID                       int64      `gorm:"primaryKey;autoIncrement" json:"-"`
	TaskID                   string     `gorm:"uniqueIndex;size:64;not null" json:"task_id"`
	Category                 string     `gorm:"size:50;not null;index" json:"category"`
	InterfaceName            string     `gorm:"size:100;not null;index" json:"interface_name"`
	DeliveryMode             string     `gorm:"size:20;not null;index" json:"delivery_mode"`
	Consumer                 string     `gorm:"size:50;index" json:"consumer,omitempty"`
	Transport                string     `gorm:"size:100;index" json:"transport,omitempty"`
	WorldUUID                string     `gorm:"size:36;index" json:"world_id,omitempty"`
	NodeUUID                 string     `gorm:"size:36;index" json:"node_id,omitempty"`
	RequestID                string     `gorm:"size:36;index" json:"request_id,omitempty"`
	CallbackID               string     `gorm:"size:64;index" json:"callback_id,omitempty"`
	ResumeExecutionID        string     `gorm:"size:64;index" json:"resume_execution_id,omitempty"`
	IdempotencyKey           string     `gorm:"size:128;index" json:"idempotency_key,omitempty"`
	Status                   string     `gorm:"size:20;not null;index" json:"status"`
	LeaseOwner               string     `gorm:"size:100;index" json:"lease_owner,omitempty"`
	LeaseToken               string     `gorm:"size:64;index" json:"lease_token,omitempty"`
	AttemptCount             int        `gorm:"default:0" json:"attempt_count"`
	MaxAttempts              int        `gorm:"default:0" json:"max_attempts"`
	Priority                 int        `gorm:"default:0;index" json:"priority"`
	PayloadJSON              string     `gorm:"type:text;not null" json:"payload_json"`
	ResultJSON               string     `gorm:"type:text" json:"result_json,omitempty"`
	ErrorMessage             string     `gorm:"type:text" json:"error_message,omitempty"`
	AvailableAt              *time.Time `gorm:"index" json:"available_at,omitempty"`
	DispatchedAt             *time.Time `json:"dispatched_at,omitempty"`
	LastDispatchAt           *time.Time `json:"last_dispatch_at,omitempty"`
	DispatchAttempts         int        `gorm:"default:0" json:"dispatch_attempts"`
	LastDispatchStatusCode   int        `gorm:"default:0" json:"last_dispatch_status_code,omitempty"`
	LastDispatchError        string     `gorm:"type:text" json:"last_dispatch_error,omitempty"`
	LastDispatchFailureClass string     `gorm:"size:50;index" json:"last_dispatch_failure_class,omitempty"`
	LastDispatchDecision     string     `gorm:"size:50;index" json:"last_dispatch_decision,omitempty"`
	FallbackFromTransport    string     `gorm:"size:100;index" json:"fallback_from_transport,omitempty"`
	LastTransitionReason     string     `gorm:"size:100;index" json:"last_transition_reason,omitempty"`
	ClaimedAt                *time.Time `json:"claimed_at,omitempty"`
	LastHeartbeatAt          *time.Time `json:"last_heartbeat_at,omitempty"`
	HeartbeatTimeoutAt       *time.Time `json:"heartbeat_timeout_at,omitempty"`
	HeartbeatTimeoutCount    int        `gorm:"default:0" json:"heartbeat_timeout_count"`
	CompletedAt              *time.Time `json:"completed_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

func (RuntimeTaskModel) TableName() string { return "runtime_tasks" }


// PendingPlanModel persists world change plans pending human approval.
// It survives server restarts so pending reviews are not lost.
type PendingPlanModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	PlanID    string    `gorm:"uniqueIndex;size:36;not null" json:"plan_id"`
	WorldUUID string    `gorm:"size:36;not null;index" json:"world_id"`
	TaskType  string    `gorm:"size:40" json:"task_type"`
	Status    string    `gorm:"size:20;not null;default:pending;index" json:"status"`
	DataJSON  string    `gorm:"type:text" json:"data"`
	TickNumber int      `json:"tick_number"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
