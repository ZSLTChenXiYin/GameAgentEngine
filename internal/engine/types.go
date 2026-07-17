package engine

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/sdk"
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
	CompProfile  ComponentType = "profile"
	CompMemory   ComponentType = "memory"
	CompRule     ComponentType = "rule"
	CompTimeline ComponentType = "timeline"
	// CompRelations 仅表示与关系相关的辅助组件负载；权威关系数据仍以 relations 表中的结构化边为准。
	// 后续开发不得把它当成与 RelationModel 并列的第二套真相来源。
	CompActionPolicy   ComponentType = "action_policy"
	CompRelations      ComponentType = "relations"
	CompPromptProfile  ComponentType = "prompt_profile"
	CompLore           ComponentType = "lore"
	CompAutonomous     ComponentType = "autonomous"
	CompWorldState     ComponentType = "world_state"
	CompStoryState     ComponentType = "story_state"
	CompStoryHistory   ComponentType = "story_history"
	CompTickPolicy     ComponentType = "tick_policy"
	CompWorldTimeState ComponentType = "world_time_state"
	CompStateSnapshot  ComponentType = "state_snapshot"
	CompWorldFocus     ComponentType = "world_focus"
)

// ValidComponentTypes 返回当前支持的全部组件类型。
func ValidComponentTypes() []ComponentType {
	return []ComponentType{CompProfile, CompMemory, CompRule, CompTimeline, CompActionPolicy, CompRelations, CompPromptProfile, CompLore, CompAutonomous, CompWorldState, CompStoryState, CompStoryHistory, CompTickPolicy, CompWorldTimeState, CompStateSnapshot, CompWorldFocus}
}

// StateComponentTypes returns the engine-recognized world continuity component types.
func StateComponentTypes() []ComponentType {
	return []ComponentType{CompWorldState, CompStoryState, CompStoryHistory, CompTickPolicy, CompWorldTimeState, CompStateSnapshot, CompWorldFocus}
}

// IsStateComponentType reports whether the given component type is used for persistent world tick continuity.
func IsStateComponentType(componentType string) bool {
	for _, item := range StateComponentTypes() {
		if string(item) == componentType {
			return true
		}
	}
	return false
}

// RelationType 表示两个节点之间的有向关系类型。
//
// 语义总约束：
//  1. 所有关系边统一按 source -> target 解读，禁止在调用方再反向猜测语义。
//  2. parent 不是普通关系边；parent 是唯一主层级结构，负责稳定归属、默认祖先链和默认上行传播。
//  3. 显式关系边负责补充动态语义，不得随意替代 parent 的职责；尤其不能把当前位置、命令链和额外作用域
//     混写成同一种关系。
//  4. 各关系类型在默认上下文中的职责必须严格区分：稳定归属走 parent，动态环境走 located_at，组织/控制
//     关系走 belongs_to/subordinate，社会语义走 ally/enemy/kinship。
type RelationType string

const (
	// RelBelongsTo 表示 source 在归属、拥有、编制或资产依附意义上属于 target。
	// 它用于表达稳定的归属语义，但不替代 parent，不表示当前位置，也不自动构成默认祖先链。
	RelBelongsTo RelationType = "belongs_to"
	// RelAlly 表示 source 与 target 处于友好、合作或联盟关系。
	// 它属于社会语义边，不参与默认层级继承；如果业务上要求双向盟友，应由上层显式写双向边或做对称解释。
	RelAlly RelationType = "ally"
	// RelEnemy 表示 source 与 target 处于敌对、冲突或对抗关系。
	// 它属于社会语义边，不参与默认层级继承；如果业务上要求双向敌对，应由上层显式写双向边或做对称解释。
	RelEnemy RelationType = "enemy"
	// RelSubordinate 表示 source 受 target 指挥、汇报或控制。
	// 它强调命令链/汇报链，不等同于一般归属，不表示当前位置，也不替代主父节点。
	RelSubordinate RelationType = "subordinate"
	// RelKinship 表示 source 与 target 存在亲属、家族或婚姻关系。
	// 它属于人物背景/社会纽带关系，不参与默认层级继承；更细的亲属类型应写入 properties，而不是继续扩枚举。
	RelKinship RelationType = "kinship"
	// RelLocatedAt 表示 source 当前位于 target 所代表的地点/场景节点。
	// 它只表达动态环境位置，不表示稳定归属，不替代 parent；后续环境上下文必须优先沿这条关系装配。
	RelLocatedAt RelationType = "located_at"
	// RelExternalParent 表示 source 在主 parent 之外额外挂接到 target 作用域。
	// 它只用于额外作用域建模，禁止拿来表达当前位置、普通组织归属或社会关系；是否参与默认上下文/传播必须
	// 由引擎显式实现，不能靠调用方臆测。
	RelExternalParent RelationType = "external_parent"
)

// ValidRelationTypes 返回当前支持的全部关系类型。
func ValidRelationTypes() []RelationType {
	return []RelationType{RelBelongsTo, RelAlly, RelEnemy, RelSubordinate, RelKinship, RelLocatedAt, RelExternalParent}
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

type DynamicInterfaceKind string

const (
	DynamicInterfaceDataRequest DynamicInterfaceKind = "data_request"
	DynamicInterfaceAction      DynamicInterfaceKind = "action"
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
	Priority        int               `json:"priority,omitempty"`
	CooldownSeconds int               `json:"cooldown_seconds,omitempty"`
	Capabilities    []AgentCapability `json:"capabilities,omitempty"`
	LastRunAt       *time.Time        `json:"last_run_at,omitempty"`
	LastError       string            `json:"last_error,omitempty"`
}

// DataQuery 表示 LLM 在分析过程中对额外数据的查询需求。
// LLM 可在对话推理中通过 request_data 告知管线需要哪些额外数据。
//
// Filter 的语义依赖于 Type：
// 1. node_components 使用 component_type。
// 2. node_relations 使用 relation_type。
// 3. node_memories 使用 memory_level。
// 任何实现都必须让这里的注释与实际行为保持一致，避免调用方以为支持过滤但实际返回全量数据。
type DataQuery struct {
	Type   string `json:"type"`              // 查询类型: "node_components" / "node_memories" / "node_relations" 等
	NodeID string `json:"node_id,omitempty"` // 要查询的游戏节点 ID
	Filter string `json:"filter,omitempty"`  // 过滤条件: component_type / relation_type / memory_level 等
	Limit  int    `json:"limit,omitempty"`   // 最大返回数量
}

// DataRequest 是 LLM 在响应中嵌入的数据请求，管线据此加载更多数据后继续推理。
type DataRequest struct {
	Label             string      `json:"label"`                        // 数据标签，用于在任务树中标识该次查询
	Target            string      `json:"target"`                       // "store"（本地存储）或 "game_client"（游戏客户端）
	ExternalInterface string      `json:"external_interface,omitempty"` // 可选: 正式接口配置名
	DeliveryMode      string      `json:"delivery_mode,omitempty"`      // 可选: push / pull / hybrid
	PrimaryTransport  string      `json:"primary_transport,omitempty"`  // 可选: 例如 game_http
	Consumer          string      `json:"consumer,omitempty"`           // 可选: 例如 game_client / bridge
	MaxAttempts       int         `json:"max_attempts,omitempty"`       // 可选: pull/hybrid 最大领取尝试次数
	TimeoutMs         int         `json:"timeout_ms,omitempty"`         // 可选: 外部调用超时
	Queries           []DataQuery `json:"queries,omitempty"`            // 要查询的数据列表
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
//
// 运行模式不仅决定轮次数量，也决定关系图谱的装配强度：
// 1. vertical: 使用最小闭环关系子图。只装配当前任务必需的身份/环境关系，缺数据时优先让模型发 request_data。
// 2. polling: 使用中等关系子图。允许在多轮推理中按需查询更多关系或节点数据，但默认仍应限制扩图范围。
// 3. full: 使用结构化关系子图。允许 task tree / sub-task 在显式约束下继续扩图，但不能退回无差别全图拼接。
type PipelineMode string

const (
	PipelineVertical PipelineMode = "vertical" // 垂直管线：一次 LLM 调用，最小关系子图，无轮询无任务树。
	PipelinePolling  PipelineMode = "polling"  // 轮询管线：多轮 LLM 轮询，中等关系子图，无任务树。
	PipelineFull     PipelineMode = "full"     // 全功能管线：多轮/任务树/子任务协作下的结构化关系子图。
)

func IsValidPipelineMode(mode string) bool {
	switch PipelineMode(mode) {
	case "", PipelineVertical, PipelinePolling, PipelineFull:
		return true
	default:
		return false
	}
}

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
// DynamicInterface describes one request-scoped external capability exposed to the model.
// It only carries this invocation's business contract. Delivery, routing, retry, and
// callback governance remain owned by external_interfaces config.
type DynamicInterface struct {
	ID                string               `json:"id"`
	Kind              DynamicInterfaceKind `json:"kind"`
	ExternalInterface string               `json:"external_interface"`
	Description       string               `json:"description,omitempty"`
	QueryTypes        []string             `json:"query_types,omitempty"`
	ArgsSchema        map[string]any       `json:"args_schema,omitempty"`
	MaxQueries        int                  `json:"max_queries,omitempty"`
	MaxCalls          int                  `json:"max_calls,omitempty"`
}

type InteractionEvent struct {
	Type   string         `json:"type"`
	ItemID string         `json:"item_id,omitempty"`
	Args   map[string]any `json:"args,omitempty"`
}

type PlayerIntentPrecondition struct {
	Type         string         `json:"type"`
	ActorNodeID  string         `json:"actor_node_id,omitempty"`
	TargetNodeID string         `json:"target_node_id,omitempty"`
	SceneNodeID  string         `json:"scene_node_id,omitempty"`
	ItemID       string         `json:"item_id,omitempty"`
	TaskID       string         `json:"task_id,omitempty"`
	Expected     string         `json:"expected,omitempty"`
	Args         map[string]any `json:"args,omitempty"`
}

type PlayerIntentStep struct {
	Type          string                     `json:"type"`
	TargetNodeID  string                     `json:"target_node_id,omitempty"`
	SceneNodeID   string                     `json:"scene_node_id,omitempty"`
	ItemID        string                     `json:"item_id,omitempty"`
	Content       string                     `json:"content,omitempty"`
	Args          map[string]any             `json:"args,omitempty"`
	Preconditions []PlayerIntentPrecondition `json:"preconditions,omitempty"`
}

type PlayerIntent struct {
	Type         string             `json:"type"`
	ActorNodeID  string             `json:"actor_node_id,omitempty"`
	SceneNodeID  string             `json:"scene_node_id,omitempty"`
	TargetNodeID string             `json:"target_node_id,omitempty"`
	Summary      string             `json:"summary,omitempty"`
	RiskLevel    string             `json:"risk_level,omitempty"`
	Confidence   float64            `json:"confidence,omitempty"`
	Steps        []PlayerIntentStep `json:"steps,omitempty"`
}

type MissingFact struct {
	Type   string `json:"type"`
	NodeID string `json:"node_id,omitempty"`
	ItemID string `json:"item_id,omitempty"`
	TaskID string `json:"task_id,omitempty"`
	Reason string `json:"reason,omitempty"`
}

type SuggestedInteraction struct {
	Mode          string `json:"mode,omitempty"`
	EventType     string `json:"event_type,omitempty"`
	AudienceScope string `json:"audience_scope,omitempty"`
	TargetNodeID  string `json:"target_node_id,omitempty"`
}

type PlayerIntentInterpretation struct {
	Intent               *PlayerIntent         `json:"intent,omitempty"`
	MissingFacts         []MissingFact         `json:"missing_facts,omitempty"`
	SuggestedInteraction *SuggestedInteraction `json:"suggested_interaction,omitempty"`
}

type InteractionContext struct {
	Mode               string            `json:"mode,omitempty"`
	SpeakerNodeID      string            `json:"speaker_node_id,omitempty"`
	TargetNodeID       string            `json:"target_node_id,omitempty"`
	SceneNodeID        string            `json:"scene_node_id,omitempty"`
	RoomID             string            `json:"room_id,omitempty"`
	ParticipantNodeIDs []string          `json:"participant_node_ids,omitempty"`
	AudienceScope      string            `json:"audience_scope,omitempty"`
	TurnIndex          int               `json:"turn_index,omitempty"`
	Event              *InteractionEvent `json:"event,omitempty"`
}

type InvokeContext struct {
	// IncludeRelatedNodes 是受控的关系补充开关，不是“把所有关系边另一端节点全部展开”的许可。
	// 后续实现应允许它按任务/关系类型/方向/hop 数受限扩图，避免关系噪音淹没主上下文。
	IncludeRelatedNodes  bool                `json:"include_related_nodes,omitempty"`
	MemoryLimit          int                 `json:"memory_limit,omitempty"`
	MaxDepth             int                 `json:"max_depth,omitempty"`
	MaxAnalysisRounds    int                 `json:"max_analysis_rounds,omitempty"`
	PipelineMode         PipelineMode        `json:"pipeline_mode,omitempty"`
	PlayerInputInterpret bool                `json:"player_input_interpret,omitempty"`
	DynamicInterfaces    []DynamicInterface  `json:"dynamic_interfaces,omitempty"`
	Interaction          *InteractionContext `json:"interaction,omitempty"`
}

// InvokeResponse 是统一的推理响应结构。
type InvokeResponse struct {
	RequestID       string                      `json:"request_id"`
	TraceID         string                      `json:"trace_id,omitempty"`
	TaskType        TaskType                    `json:"task_type"`
	ExecutionMode   ExecutionMode               `json:"execution_mode"`
	Reply           string                      `json:"reply,omitempty"`
	AdvancedTicks   int                         `json:"advanced_ticks,omitempty"`
	ActionCalls     []ActionCall                `json:"action_calls,omitempty"`
	WorldChangePlan *WorldChangePlan            `json:"world_change_plan,omitempty"`
	FutureOutline   string                      `json:"future_outline,omitempty"`
	MemoryUpdates   []MemoryUpdate              `json:"memory_updates,omitempty"`
	DataRequest     *DataRequest                `json:"data_request,omitempty"`
	PlayerIntent    *PlayerIntentInterpretation `json:"player_intent,omitempty"`
	SubTasks        []SubTaskDeclaration        `json:"sub_tasks,omitempty"`
	Metadata        *ResponseMeta               `json:"metadata,omitempty"`
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
	PropModeUpward       PropagationMode = "upward"             // 只沿主 parent 链向上传播，不包含 located_at、belongs_to、subordinate 或 external_parent。
	PropModeEnvironment  PropagationMode = "environment_scope"  // 沿 located_at 指向的当前环境节点及其 parent 场景祖先传播，用于环境作用域，不替代主父链。
	PropModeOrganization PropagationMode = "organization_scope" // 沿 belongs_to/subordinate 指向的组织或控制节点及其主 parent 链传播，用于组织/控制作用域。
	PropModeTagBroadcast PropagationMode = "tag_broadcast"      // 按 tags 匹配节点扩散，适合显式广播，不表达层级、环境或控制链语义。
	PropModeTargeted     PropagationMode = "targeted"           // 定向传播到指定节点，适合业务显式点名，不表达默认结构语义。
	PropModeManual       PropagationMode = "manual"             // 不自动传播，交由外部显式触发。
)

// IsValidPropagationMode reports whether mode is one of the engine's
// explicitly supported propagation strategies.
func IsValidPropagationMode(mode PropagationMode) bool {
	switch mode {
	case PropModeUpward, PropModeEnvironment, PropModeOrganization, PropModeTagBroadcast, PropModeTargeted, PropModeManual:
		return true
	default:
		return false
	}
}

// PropagationRule 描述一条记忆的传播规则。
type PropagationRule struct {
	Mode          PropagationMode `json:"mode,omitempty"`            // 传播模式；空值默认 upward。environment_scope 与 organization_scope 必须走显式关系语义，不能偷用 upward。
	TargetTags    []string        `json:"target_tags,omitempty"`     // tag_broadcast 模式：匹配这些 tag 的节点
	TargetNodeIDs []string        `json:"target_node_ids,omitempty"` // targeted 模式：目标节点 ID 列表
	MaxDepth      int             `json:"max_depth,omitempty"`       // 0 = 全路径；对 environment_scope / organization_scope 表示目标节点起始后的祖先展开深度
	PublishUp     bool            `json:"publish_up,omitempty"`      // 是否在沿主 parent 链继续传播时发布到更高层公共节点；不改变模式语义归属
}

// PropagateAction 描述规则链中触发后的一个传播动作。
type PropagateAction struct {
	Mode          PropagationMode `json:"mode"`
	TargetTags    []string        `json:"target_tags,omitempty"`
	TargetNodeIDs []string        `json:"target_node_ids,omitempty"`
	MaxDepth      int             `json:"max_depth,omitempty"`
	PublishUp     bool            `json:"publish_up,omitempty"`
	Transform     *TransformRule  `json:"transform,omitempty"`
	NextChainIDs  []string        `json:"next_chain_ids,omitempty"`
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
	Propagation *PropagationRule `json:"propagation,omitempty"` // 传播规则；nil 表示默认 upward。显式环境/组织传播必须写在这里，不能靠节点树结构猜测。
}

// ResponseMeta 记录本次推理的元信息。
type ResponseMeta struct {
	LLMModel               string `json:"llm_model"`
	TokensUsed             int    `json:"tokens_used"`
	ProcessingTimeMs       int64  `json:"processing_time_ms"`
	ConfiguredPipelineMode string `json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode  string `json:"effective_pipeline_mode,omitempty"`
	MaxAnalysisRounds      int    `json:"max_analysis_rounds,omitempty"`
	RoundsUsed             int    `json:"rounds_used,omitempty"`
}

// LLMToolDefinition describes one structured callable tool exposed to the provider.
type LLMToolInvocationKind string

const (
	LLMToolInvocationAction      LLMToolInvocationKind = "action"
	LLMToolInvocationDataRequest LLMToolInvocationKind = "data_request"
)

// LLMDataRequestTemplate defines fixed request_data fields that a provider should
// merge with structured tool arguments before returning normalized JSON.
type LLMDataRequestTemplate struct {
	Label             string `json:"label,omitempty"`
	Target            string `json:"target,omitempty"`
	ExternalInterface string `json:"external_interface,omitempty"`
	DeliveryMode      string `json:"delivery_mode,omitempty"`
	PrimaryTransport  string `json:"primary_transport,omitempty"`
	Consumer          string `json:"consumer,omitempty"`
	TimeoutMs         int    `json:"timeout_ms,omitempty"`
}

// LLMToolDefinition describes one structured callable tool exposed to the provider.
type LLMToolDefinition struct {
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	Parameters  map[string]any          `json:"parameters,omitempty"`
	Invocation  LLMToolInvocationKind   `json:"-"`
	ActionID    string                  `json:"-"`
	DataRequest *LLMDataRequestTemplate `json:"-"`
}

// LLMChatRequest carries prompt, messages, and optional structured tool definitions.
type LLMChatRequest struct {
	SystemPrompt string              `json:"system_prompt"`
	Messages     []ChatMessage       `json:"messages,omitempty"`
	Tools        []LLMToolDefinition `json:"tools,omitempty"`
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
	Chat(req *LLMChatRequest) (*LLMResult, error)
	ModelName() string
}

// LLMStructuredToolProvider is an optional provider capability contract.
// Providers that do not implement it are treated as tool-capable for backward compatibility.
type LLMStructuredToolProvider interface {
	SupportsStructuredTools() bool
}

// LLMResult 表示一次大模型调用的标准输出。
type LLMResult struct {
	Content  string         `json:"content"`
	Model    string         `json:"model"`
	Tokens   int            `json:"tokens"`
	Metadata map[string]any `json:"metadata,omitempty"`
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
var dynamicInterfaceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_:-]{1,63}$`)

var validInteractionModes = []string{sdk.InteractionModeDirectDialogue, sdk.InteractionModeGroupChat, sdk.InteractionModeGiftResponse, sdk.InteractionModeTradeDialogue}
var validInteractionEventTypes = []string{sdk.InteractionEventSpeech, sdk.InteractionEventGift, sdk.InteractionEventShowItem, sdk.InteractionEventTradeRequest, sdk.InteractionEventThreaten}
var validInteractionAudienceScopes = []string{sdk.InteractionAudiencePublic, sdk.InteractionAudiencePrivate, sdk.InteractionAudienceWhisper}
var validSuggestedInteractionEventTypes = []string{sdk.InteractionEventSpeech, sdk.InteractionEventGift, sdk.InteractionEventShowItem, sdk.InteractionEventTradeRequest, sdk.InteractionEventThreaten}
var validPlayerIntentTypes = sdk.ValidPlayerIntentTypes()
var validPlayerIntentStepTypes = sdk.ValidPlayerIntentStepTypes()
var validPlayerIntentRiskLevels = sdk.ValidPlayerIntentRiskLevels()
var validPlayerIntentPreconditionTypes = sdk.ValidPlayerIntentPreconditionTypes()
var validMissingFactTypes = sdk.ValidMissingFactTypes()

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

func IsValidDynamicInterfaceKind(kind string) bool {
	switch DynamicInterfaceKind(kind) {
	case DynamicInterfaceDataRequest, DynamicInterfaceAction:
		return true
	default:
		return false
	}
}

func ValidateInteractionContext(ctx *InteractionContext) error {
	if ctx == nil {
		return nil
	}
	mode := strings.TrimSpace(ctx.Mode)
	if mode == "" {
		return fmt.Errorf("interaction.mode required")
	}
	if !slices.Contains(validInteractionModes, mode) {
		return fmt.Errorf("interaction.mode must be one of: %s", strings.Join(validInteractionModes, ", "))
	}
	if strings.TrimSpace(ctx.SpeakerNodeID) == "" {
		return fmt.Errorf("interaction.speaker_node_id required")
	}
	if strings.TrimSpace(ctx.TargetNodeID) == "" {
		return fmt.Errorf("interaction.target_node_id required")
	}
	if ctx.TurnIndex < 0 {
		return fmt.Errorf("interaction.turn_index must be >= 0")
	}
	if scope := strings.TrimSpace(ctx.AudienceScope); scope != "" && !slices.Contains(validInteractionAudienceScopes, scope) {
		return fmt.Errorf("interaction.audience_scope must be one of: %s", strings.Join(validInteractionAudienceScopes, ", "))
	}
	seenParticipants := make(map[string]struct{}, len(ctx.ParticipantNodeIDs))
	for i, participant := range ctx.ParticipantNodeIDs {
		id := strings.TrimSpace(participant)
		if id == "" {
			return fmt.Errorf("interaction.participant_node_ids[%d] required", i)
		}
		if _, exists := seenParticipants[id]; exists {
			return fmt.Errorf("interaction.participant_node_ids[%d] duplicated: %s", i, id)
		}
		seenParticipants[id] = struct{}{}
	}
	if ctx.Event != nil {
		eventType := strings.TrimSpace(ctx.Event.Type)
		if eventType == "" {
			return fmt.Errorf("interaction.event.type required")
		}
		if !slices.Contains(validInteractionEventTypes, eventType) {
			return fmt.Errorf("interaction.event.type must be one of: %s", strings.Join(validInteractionEventTypes, ", "))
		}
	}
	return nil
}

func ValidatePlayerIntentInterpretation(payload *PlayerIntentInterpretation) error {
	if payload == nil {
		return nil
	}
	if payload.Intent == nil {
		return fmt.Errorf("player_intent.intent required")
	}
	if err := ValidatePlayerIntent(payload.Intent); err != nil {
		return err
	}
	for i, item := range payload.MissingFacts {
		if trimmed := strings.TrimSpace(item.Type); trimmed == "" {
			return fmt.Errorf("player_intent.missing_facts[%d].type required", i)
		} else if !slices.Contains(validMissingFactTypes, trimmed) {
			return fmt.Errorf("player_intent.missing_facts[%d].type must be one of: %s", i, strings.Join(validMissingFactTypes, ", "))
		}
	}
	if payload.SuggestedInteraction != nil {
		if mode := strings.TrimSpace(payload.SuggestedInteraction.Mode); mode != "" && !slices.Contains(validInteractionModes, mode) {
			return fmt.Errorf("player_intent.suggested_interaction.mode must be one of: %s", strings.Join(validInteractionModes, ", "))
		}
		if eventType := strings.TrimSpace(payload.SuggestedInteraction.EventType); eventType != "" && !slices.Contains(validSuggestedInteractionEventTypes, eventType) {
			return fmt.Errorf("player_intent.suggested_interaction.event_type must be one of: %s", strings.Join(validSuggestedInteractionEventTypes, ", "))
		}
		if scope := strings.TrimSpace(payload.SuggestedInteraction.AudienceScope); scope != "" && !slices.Contains(validInteractionAudienceScopes, scope) {
			return fmt.Errorf("player_intent.suggested_interaction.audience_scope must be one of: %s", strings.Join(validInteractionAudienceScopes, ", "))
		}
	}
	return nil
}

func ValidatePlayerIntent(intent *PlayerIntent) error {
	if intent == nil {
		return fmt.Errorf("player_intent.intent required")
	}
	intentType := strings.TrimSpace(intent.Type)
	if intentType == "" {
		return fmt.Errorf("player_intent.intent.type required")
	}
	if !slices.Contains(validPlayerIntentTypes, intentType) {
		return fmt.Errorf("player_intent.intent.type must be one of: %s", strings.Join(validPlayerIntentTypes, ", "))
	}
	if strings.TrimSpace(intent.ActorNodeID) == "" {
		return fmt.Errorf("player_intent.intent.actor_node_id required")
	}
	if risk := strings.TrimSpace(intent.RiskLevel); risk != "" && !slices.Contains(validPlayerIntentRiskLevels, risk) {
		return fmt.Errorf("player_intent.intent.risk_level must be one of: %s", strings.Join(validPlayerIntentRiskLevels, ", "))
	}
	if intent.Confidence < 0 || intent.Confidence > 1 {
		return fmt.Errorf("player_intent.intent.confidence must be between 0 and 1")
	}
	if len(intent.Steps) == 0 {
		return fmt.Errorf("player_intent.intent.steps required")
	}
	if intentType != sdk.PlayerIntentTypeComposite && len(intent.Steps) != 1 {
		return fmt.Errorf("player_intent.intent.steps must contain exactly 1 step for non-composite intent")
	}
	for i, step := range intent.Steps {
		if intentType != sdk.PlayerIntentTypeComposite && strings.TrimSpace(step.Type) != intentType {
			return fmt.Errorf("player_intent.intent.steps[%d]: type must match intent.type %s", i, intentType)
		}
		if err := ValidatePlayerIntentStep(step); err != nil {
			return fmt.Errorf("player_intent.intent.steps[%d]: %w", i, err)
		}
	}
	return nil
}

func ValidatePlayerIntentStep(step PlayerIntentStep) error {
	stepType := strings.TrimSpace(step.Type)
	if stepType == "" {
		return fmt.Errorf("type required")
	}
	if !slices.Contains(validPlayerIntentStepTypes, stepType) {
		return fmt.Errorf("type must be one of: %s", strings.Join(validPlayerIntentStepTypes, ", "))
	}
	switch stepType {
	case sdk.PlayerIntentTypeSpeech:
		if strings.TrimSpace(step.Content) == "" {
			return fmt.Errorf("content required")
		}
	case sdk.PlayerIntentTypeShowItem, sdk.PlayerIntentTypeGift:
		if strings.TrimSpace(step.TargetNodeID) == "" {
			return fmt.Errorf("target_node_id required")
		}
		if strings.TrimSpace(step.ItemID) == "" {
			return fmt.Errorf("item_id required")
		}
	case sdk.PlayerIntentTypeTradeRequest, sdk.PlayerIntentTypeThreaten:
		if strings.TrimSpace(step.TargetNodeID) == "" {
			return fmt.Errorf("target_node_id required")
		}
	case sdk.PlayerIntentTypeMove:
		if strings.TrimSpace(step.TargetNodeID) == "" {
			if destination, _ := step.Args["destination_scene_id"].(string); strings.TrimSpace(destination) == "" {
				return fmt.Errorf("target_node_id or args.destination_scene_id required")
			}
		}
	case sdk.PlayerIntentTypeInspect:
		if strings.TrimSpace(step.TargetNodeID) == "" && strings.TrimSpace(step.ItemID) == "" {
			return fmt.Errorf("target_node_id or item_id required")
		}
	case sdk.PlayerIntentTypeUseItem:
		if strings.TrimSpace(step.ItemID) == "" {
			return fmt.Errorf("item_id required")
		}
	}
	for i, item := range step.Preconditions {
		if err := ValidatePlayerIntentPrecondition(item); err != nil {
			return fmt.Errorf("preconditions[%d]: %w", i, err)
		}
	}
	return nil
}

func ValidatePlayerIntentPrecondition(item PlayerIntentPrecondition) error {
	preType := strings.TrimSpace(item.Type)
	if preType == "" {
		return fmt.Errorf("type required")
	}
	if !slices.Contains(validPlayerIntentPreconditionTypes, preType) {
		return fmt.Errorf("type must be one of: %s", strings.Join(validPlayerIntentPreconditionTypes, ", "))
	}
	return nil
}

func ValidateDynamicInterfaces(items []DynamicInterface) error {
	seenIDs := make(map[string]struct{}, len(items))
	for i, item := range items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			return fmt.Errorf("dynamic_interfaces[%d].id required", i)
		}
		if !dynamicInterfaceIDPattern.MatchString(id) {
			return fmt.Errorf("dynamic_interfaces[%d].id must match %s", i, dynamicInterfaceIDPattern.String())
		}
		if _, exists := seenIDs[id]; exists {
			return fmt.Errorf("dynamic_interfaces[%d].id duplicated: %s", i, id)
		}
		seenIDs[id] = struct{}{}

		kind := strings.TrimSpace(string(item.Kind))
		if !IsValidDynamicInterfaceKind(kind) {
			return fmt.Errorf("dynamic_interfaces[%d].kind must be one of: data_request, action", i)
		}
		if strings.TrimSpace(item.ExternalInterface) == "" {
			return fmt.Errorf("dynamic_interfaces[%d].external_interface required", i)
		}
		if item.MaxQueries < 0 {
			return fmt.Errorf("dynamic_interfaces[%d].max_queries must be >= 0", i)
		}
		if item.MaxCalls < 0 {
			return fmt.Errorf("dynamic_interfaces[%d].max_calls must be >= 0", i)
		}

		switch item.Kind {
		case DynamicInterfaceDataRequest:
			if len(item.QueryTypes) == 0 {
				return fmt.Errorf("dynamic_interfaces[%d].query_types required for data_request", i)
			}
			if item.MaxCalls > 0 {
				return fmt.Errorf("dynamic_interfaces[%d].max_calls only applies to action interfaces", i)
			}
			if len(item.ArgsSchema) > 0 {
				return fmt.Errorf("dynamic_interfaces[%d].args_schema only applies to action interfaces", i)
			}
		case DynamicInterfaceAction:
			if item.MaxQueries > 0 {
				return fmt.Errorf("dynamic_interfaces[%d].max_queries only applies to data_request interfaces", i)
			}
			if len(item.QueryTypes) > 0 {
				return fmt.Errorf("dynamic_interfaces[%d].query_types only applies to data_request interfaces", i)
			}
			if len(item.ArgsSchema) > 0 {
				schemaType, _ := item.ArgsSchema["type"].(string)
				if strings.TrimSpace(schemaType) != "object" {
					return fmt.Errorf("dynamic_interfaces[%d].args_schema.type must be object when provided", i)
				}
			}
		}

		seenQueryTypes := make(map[string]struct{}, len(item.QueryTypes))
		for j, queryType := range item.QueryTypes {
			trimmed := strings.TrimSpace(queryType)
			if trimmed == "" {
				return fmt.Errorf("dynamic_interfaces[%d].query_types[%d] required", i, j)
			}
			if _, exists := seenQueryTypes[trimmed]; exists {
				return fmt.Errorf("dynamic_interfaces[%d].query_types[%d] duplicated: %s", i, j, trimmed)
			}
			seenQueryTypes[trimmed] = struct{}{}
		}
	}
	return nil
}
