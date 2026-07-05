package store

import (
	"time"

	"gorm.io/gorm"
)

// NodeModel 是节点实体在数据库中的持久化结构。
type NodeModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID      string         `gorm:"uniqueIndex;size:36;not null" json:"id"`
	WorldID   int64          `gorm:"index;not null" json:"-"`
	WorldUUID string         `gorm:"-" json:"world_id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	NodeType  string         `gorm:"size:50;not null;index" json:"node_type"`
	ParentID  *int64         `gorm:"index" json:"-"`
	ParentUUID *string       `gorm:"-" json:"parent_id,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
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

// InferenceLogModel 是推理调用日志的持久化结构。
type InferenceLogModel struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"-"`
	UUID         string    `gorm:"uniqueIndex;size:36;not null" json:"id"`
	WorldID      int64     `gorm:"index;not null" json:"-"`
	WorldUUID    string    `gorm:"-" json:"world_id"`
	TaskType     string    `gorm:"size:50;not null" json:"task_type"`
	NodeID       *int64    `gorm:"index" json:"-"`
	NodeUUID     string    `gorm:"-" json:"node_id,omitempty"`
	RequestData  string    `gorm:"type:text" json:"request_data,omitempty"`
	ResponseData string    `gorm:"type:text" json:"response_data,omitempty"`
	LLMModel     string    `gorm:"size:100" json:"llm_model"`
	TokensUsed   int       `gorm:"default:0" json:"tokens_used"`
	DurationMs   int64     `gorm:"default:0" json:"duration_ms"`
	CreatedAt    time.Time `json:"created_at"`
}

func (InferenceLogModel) TableName() string { return "inference_logs" }

// IdempotencyKeyModel 存储幂等操作的已缓存结果。
type IdempotencyKeyModel struct {
	ID        string    `gorm:"primaryKey;size:64" json:"id"`
	Result    string    `gorm:"type:text;not null" json:"result"`
	CreatedAt time.Time `json:"created_at"`
}

func (IdempotencyKeyModel) TableName() string { return "idempotency_keys" }

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