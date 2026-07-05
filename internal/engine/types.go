package engine

import (
	"regexp"
	"slices"
	"time"
)

// NodeType 表示节点在实体层级中的类别。
type NodeType string

const (
	NodeTypeWorld     NodeType = "world"
	NodeTypeFaction   NodeType = "faction"
	NodeTypeLocation  NodeType = "location"
	NodeTypeNPC       NodeType = "npc"
	NodeTypeItem      NodeType = "item"
	NodeTypeQuestLine NodeType = "quest_line"
	NodeTypeEvent     NodeType = "event"
)

// ValidNodeTypes 返回当前支持的全部节点类型。
func ValidNodeTypes() []NodeType {
	return []NodeType{NodeTypeWorld, NodeTypeFaction, NodeTypeLocation, NodeTypeNPC, NodeTypeItem, NodeTypeQuestLine, NodeTypeEvent}
}

// ComponentType 表示挂载在节点上的组件类别。
type ComponentType string

const (
	CompProfile       ComponentType = "profile"
	CompMemory        ComponentType = "memory"
	CompRule          ComponentType = "rule"
	CompTimeline      ComponentType = "timeline"
	CompActionPolicy  ComponentType = "action_policy"
	CompRelations     ComponentType = "relations"
	CompPromptProfile ComponentType = "prompt_profile"
	CompLore          ComponentType = "lore"
	CompAutonomous    ComponentType = "autonomous"
)

// ValidComponentTypes 返回当前支持的全部组件类型。
func ValidComponentTypes() []ComponentType {
	return []ComponentType{CompProfile, CompMemory, CompRule, CompTimeline, CompActionPolicy, CompRelations, CompPromptProfile, CompLore, CompAutonomous}
}

// RelationType 表示两个节点之间的有向关系类型。
type RelationType string

const (
	RelBelongsTo   RelationType = "belongs_to"
	RelAlly        RelationType = "ally"
	RelEnemy       RelationType = "enemy"
	RelSubordinate RelationType = "subordinate"
	RelKinship     RelationType = "kinship"
	RelLocatedAt   RelationType = "located_at"
)

// ValidRelationTypes 返回当前支持的全部关系类型。
func ValidRelationTypes() []RelationType {
	return []RelationType{RelBelongsTo, RelAlly, RelEnemy, RelSubordinate, RelKinship, RelLocatedAt}
}

// MemoryLevel 表示记忆的可见范围和持久性。
type MemoryLevel string

const (
	MemShortTerm MemoryLevel = "short_term"
	MemLongTerm  MemoryLevel = "long_term"
	MemShared    MemoryLevel = "shared"
	MemWorld     MemoryLevel = "world"
)

// ValidMemoryLevels 返回当前支持的全部记忆层级。
func ValidMemoryLevels() []MemoryLevel {
	return []MemoryLevel{MemShortTerm, MemLongTerm, MemShared, MemWorld}
}

// TaskType 表示一次推理请求的任务类型。
type TaskType string

const (
	TaskNPCDialogue   TaskType = "npc_dialogue"
	TaskWorldTick     TaskType = "world_tick"
	TaskWorldEvent    TaskType = "world_event_impact"
	TaskAutonomousAct TaskType = "autonomous_act"
	TaskCustom        TaskType = "custom"
)

// ValidTaskTypes 返回当前支持的全部任务类型。
func ValidTaskTypes() []TaskType {
	return []TaskType{TaskNPCDialogue, TaskWorldTick, TaskWorldEvent, TaskAutonomousAct, TaskCustom}
}

const (
	AutonomousTriggerManual        = "manual"
	AutonomousTriggerWorldTickSync = "world_tick_sync"
	AutonomousTriggerScheduled     = "scheduled"
)

// AgentCapability 描述某个自主节点被显式授权调用的能力。
type AgentCapability struct {
	ID          string         `json:"id"`
	Mode        string         `json:"mode,omitempty"`
	Description string         `json:"description,omitempty"`
	Schema      map[string]any `json:"schema,omitempty"`
}

// AutonomousConfig 是挂载在节点 autonomous 组件里的自主行为配置。
// 该配置属于运行时动态配置，应通过数据库组件由 Creator/DevCli/SDK 管理。
type AutonomousConfig struct {
	Enabled         bool              `json:"enabled"`
	Trigger         string            `json:"trigger"`
	IntervalSeconds int               `json:"interval_seconds,omitempty"`
	Capabilities    []AgentCapability `json:"capabilities,omitempty"`
	LastRunAt       *time.Time        `json:"last_run_at,omitempty"`
	LastError       string            `json:"last_error,omitempty"`
}

// DataQuery 表示 LLM 在分析过程中对额外数据的查询需求。
// LLM 可在对话推理中通过 request_data 告知管线需要哪些额外数据。
type DataQuery struct {
	Type   string `json:"type"`              // 查询类型: "node_components" / "node_memories" / "node_relations" 等
	NodeID string `json:"node_id,omitempty"` // 要查询的游戏节点 ID
	Filter string `json:"filter,omitempty"`  // 过滤条件: component_type / relation_type / memory_level 等
	Limit  int    `json:"limit,omitempty"`   // 最大返回数量
}

// DataRequest 是 LLM 在响应中嵌入的数据请求，管线据此加载更多数据后继续推理。
type DataRequest struct {
	Label   string      `json:"label"`             // 数据标签，用于在任务树中标识该次查询
	Target  string      `json:"target"`            // "store"（本地存储）或 "game_client"（游戏客户端）
	Queries []DataQuery `json:"queries,omitempty"` // 要查询的数据列表
}

// ExecutionMode 表示世界变更计划的执行模式。
// ExecutionMode 表示世界变更计划的执行模式。
type ExecutionMode string

const (
	ModeDebug      ExecutionMode = "debug"
	ModeReview     ExecutionMode = "review"
	ModeProduction ExecutionMode = "production"
)
// PipelineMode 表示推理管线的运行模式。
type PipelineMode string

const (
	PipelineVertical PipelineMode = "vertical"   // 垂直管线：一次 LLM 调用，无轮询无任务树
	PipelinePolling  PipelineMode = "polling"    // 轮询管线：多轮 LLM 轮询，无任务树
	PipelineFull     PipelineMode = "full"       // 全功能管线：当前完整实现
)

// Node 是节点实体的公开传输结构。
type Node struct {
	ID        string    `json:"id"`
	WorldID   string    `json:"world_id"`
	Name      string    `json:"name"`
	NodeType  NodeType  `json:"node_type"`
	ParentID  *string   `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Component 是组件实体的公开传输结构。
type Component struct {
	ID            string        `json:"id"`
	NodeID        string        `json:"node_id"`
	ComponentType ComponentType `json:"component_type"`
	Data          string        `json:"data"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// Relation 是节点关系的公开传输结构。
type Relation struct {
	ID           string       `json:"id"`
	WorldID      string       `json:"world_id"`
	SourceID     string       `json:"source_id"`
	TargetID     string       `json:"target_id"`
	RelationType RelationType `json:"relation_type"`
	Weight       int          `json:"weight"`
	Properties   string       `json:"properties,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
}

// Memory 是节点记忆的公开传输结构。
type Memory struct {
	ID        string      `json:"id"`
	NodeID    string      `json:"node_id"`
	Content   string      `json:"content"`
	Level     MemoryLevel `json:"level"`
	Tags      string      `json:"tags,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

// InvokeRequest 是统一的推理请求结构。
type InvokeRequest struct {
	WorldID   string         `json:"world_id"`
	TaskType  TaskType       `json:"task_type"`
	NodeID    string         `json:"node_id"`
	SessionID string         `json:"session_id,omitempty"`
	Messages  []ChatMessage  `json:"messages,omitempty"`
	Context   *InvokeContext `json:"context,omitempty"`
	Event     *WorldEvent    `json:"event,omitempty"`
}

// ChatMessage 表示一次对话中的单条消息。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// InvokeContext 表示调用方希望追加的上下文约束。
type InvokeContext struct {
	IncludeRelatedNodes bool `json:"include_related_nodes,omitempty"`
	MemoryLimit         int  `json:"memory_limit,omitempty"`
	MaxDepth            int  `json:"max_depth,omitempty"`
	MaxAnalysisRounds   int          `json:"max_analysis_rounds,omitempty"`
	PipelineMode        PipelineMode  `json:"pipeline_mode,omitempty"`
}

// InvokeResponse 是统一的推理响应结构。
type InvokeResponse struct {
	RequestID       string           `json:"request_id"`
	TaskType        TaskType         `json:"task_type"`
	ExecutionMode   ExecutionMode    `json:"execution_mode"`
	Reply           string           `json:"reply,omitempty"`
	ActionCalls     []ActionCall     `json:"action_calls,omitempty"`
	WorldChangePlan *WorldChangePlan `json:"world_change_plan,omitempty"`
	FutureOutline   string           `json:"future_outline,omitempty"`
	MemoryUpdates   []MemoryUpdate       `json:"memory_updates,omitempty"`
	DataRequest     *DataRequest         `json:"data_request,omitempty"`
	SubTasks        []SubTaskDeclaration `json:"sub_tasks,omitempty"`
	Metadata        *ResponseMeta        `json:"metadata,omitempty"`
}

// AutonomousRunResult 描述一次自主节点触发的执行结果。
type AutonomousRunResult struct {
	NodeID   string          `json:"node_id"`
	Response *InvokeResponse `json:"response,omitempty"`
	Error    string          `json:"error,omitempty"`
}

// ActionCall 描述一次需要执行的动作调用。
type ActionCall struct {
	ActionID   string         `json:"action_id"`
	Args       map[string]any `json:"args"`
	Mode       string         `json:"mode,omitempty"`
	CallbackID string         `json:"callback_id,omitempty"`
}

// PropagationMode 描述记忆传播模式。
type PropagationMode string

const (
	PropModeUpward       PropagationMode = "upward"          // 沿父链向上传播
	PropModeTagBroadcast PropagationMode = "tag_broadcast"  // 按 tags 匹配节点扩散
	PropModeTargeted     PropagationMode = "targeted"       // 定向传播到指定节点
	PropModeManual       PropagationMode = "manual"         // 不自动传播，用户手动触发
)

// PropagationRule 描述一条记忆的传播规则。
type PropagationRule struct {
	Mode          PropagationMode     `json:"mode,omitempty"`          // 传播模式，默认 upward
	TargetTags    []string            `json:"target_tags,omitempty"`   // tag_broadcast 模式：匹配这些 tag 的节点
	TargetNodeIDs []string            `json:"target_node_ids,omitempty"` // targeted 模式：目标节点 ID 列表
	MaxDepth      int                 `json:"max_depth,omitempty"`    // 0 = 全路径
	PublishUp     bool                `json:"publish_up,omitempty"`   // 是否上升到上层节点
}


// PropagateAction 描述规则链中触发后的一个传播动作。
type PropagateAction struct {
	Mode          PropagationMode     `json:"mode"`
	TargetTags    []string            `json:"target_tags,omitempty"`
	TargetNodeIDs []string            `json:"target_node_ids,omitempty"`
	MaxDepth      int                 `json:"max_depth,omitempty"`
	PublishUp     bool                `json:"publish_up,omitempty"`
	Transform     *TransformRule      `json:"transform,omitempty"`
	NextChainIDs  []string            `json:"next_chain_ids,omitempty"`
}

// TransformRule 描述传播时对记忆内容的转换规则。
type TransformRule struct {
	ContentPrefix string   `json:"content_prefix,omitempty"`
	LevelUp       bool     `json:"level_up,omitempty"`
	AppendTags    []string `json:"append_tags,omitempty"`
}
// MemoryUpdate 描述一次需要落库的记忆更新。
type MemoryUpdate struct {
	NodeID      string           `json:"node_id"`
	Content     string           `json:"content"`
	Level       MemoryLevel      `json:"level"`
	Tags        string           `json:"tags,omitempty"`
	Propagation *PropagationRule `json:"propagation,omitempty"` // 传播规则，nil 表示默认向上
}

// ResponseMeta 记录本次推理的元信息。
type ResponseMeta struct {
	LLMModel         string `json:"llm_model"`
	TokensUsed       int    `json:"tokens_used"`
	ProcessingTimeMs int64  `json:"processing_time_ms"`
}

// WorldEvent 描述一个待评估的世界事件。
type WorldEvent struct {
	EventType   string `json:"event_type"`
	ScopeID     string `json:"scope_id"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// WorldChangePlan 描述世界推进后的建议变更计划。
type WorldChangePlan struct {
	TimelineID      string           `json:"timeline_id"`
	ImpactLevel     string           `json:"impact_level"`
	Summary         string           `json:"summary"`
	WorldEvents     []PlanEvent      `json:"world_events"`
	ProposedActions []ProposedAction `json:"proposed_actions"`
}

// PlanEvent 描述世界变更计划中的单个事件。
type PlanEvent struct {
	EventType   string  `json:"event_type"`
	Scope       string  `json:"scope"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// ProposedAction 描述世界变更计划中的建议动作。
type ProposedAction struct {
	APIName string         `json:"api_name"`
	Args    map[string]any `json:"args"`
}

// WorldTimelineTick 描述一次世界时间线刻度信息。
type WorldTimelineTick struct {
	WorldID    string    `json:"world_id"`
	TickNumber int       `json:"tick_number"`
	TickType   string    `json:"tick_type"`
	GameTime   string    `json:"game_time"`
	CreatedAt  time.Time `json:"created_at"`
}

// LLMProvider 是引擎对大模型能力的统一抽象。
type LLMProvider interface {
	Chat(systemPrompt string, messages []ChatMessage) (*LLMResult, error)
	ModelName() string
}

// LLMResult 表示一次大模型调用的标准输出。
type LLMResult struct {
	Content string `json:"content"`
	Model   string `json:"model"`
	Tokens  int    `json:"tokens"`
}




// ==================== Sub-Task DAG Types ====================

// SubTaskStatus 表示子任务的执行状态。
type SubTaskStatus string

const (
	SubTaskPending   SubTaskStatus = "pending"
	SubTaskReady     SubTaskStatus = "ready"
	SubTaskRunning   SubTaskStatus = "running"
	SubTaskCompleted SubTaskStatus = "completed"
	SubTaskFailed    SubTaskStatus = "failed"
)

// SubTaskDeclaration 描述 LLM 声明的一个子任务。
type SubTaskDeclaration struct {
	Label     string   `json:"label"`
	TaskType  TaskType `json:"task_type"`
	NodeID    string   `json:"node_id"`
	DependsOn []string `json:"depends_on,omitempty"`
	MergeMode string   `json:"merge_mode,omitempty"`
}

var customComponentTypePattern = regexp.MustCompile(`^[a-z][a-z0-9_:-]{1,63}$`)

func IsValidNodeType(nodeType string) bool {
	return slices.Contains(ValidNodeTypes(), NodeType(nodeType))
}

func IsValidRelationType(relationType string) bool {
	return slices.Contains(ValidRelationTypes(), RelationType(relationType))
}

func IsValidMemoryLevel(level string) bool {
	return slices.Contains(ValidMemoryLevels(), MemoryLevel(level))
}

func IsValidComponentType(componentType string) bool {
	if slices.Contains(ValidComponentTypes(), ComponentType(componentType)) {
		return true
	}
	return customComponentTypePattern.MatchString(componentType)
}
