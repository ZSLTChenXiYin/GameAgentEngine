package engine

import (
	"github.com/google/uuid"
	"sync"
	"time"
)

// Trace 记录一次完整的 LLM 推理轨迹。
// 仅在 Debug 执行模式下启用。
type TraceStep struct {
	Name       string `json:"name"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

type Trace struct {
	ID                     string               `json:"id"`
	WorldID                string               `json:"world_id"`
	RequestID              string               `json:"request_id"`
	TaskType               TaskType             `json:"task_type"`
	NodeID                 string               `json:"node_id"`
	ConfiguredPipelineMode string               `json:"configured_pipeline_mode,omitempty"`
	EffectivePipelineMode  string               `json:"effective_pipeline_mode,omitempty"`
	MaxAnalysisRounds      int                  `json:"max_analysis_rounds,omitempty"`
	RoundsUsed             int                  `json:"rounds_used,omitempty"`
	Timestamp              time.Time            `json:"timestamp"`
	DurationMs             int64                `json:"duration_ms"`
	PromptTokens           int                  `json:"prompt_tokens"`
	CompletionTokens       int                  `json:"completion_tokens"`
	SystemPrompt           string               `json:"system_prompt,omitempty"`
	Messages               []ChatMessage        `json:"messages,omitempty"`
	RawLLMResponse         string               `json:"raw_llm_response,omitempty"`
	ParsedActions          []ActionCall         `json:"parsed_actions,omitempty"`
	ParsedMemories         []MemoryUpdate       `json:"parsed_memories,omitempty"`
	SubTasks               []SubTaskDeclaration `json:"sub_tasks,omitempty"`
	WorldChangePlan        *WorldChangePlan     `json:"world_change_plan,omitempty"`
	Round                  int                  `json:"round"`
	Steps                  []TraceStep          `json:"steps,omitempty"`
	Error                  string               `json:"error,omitempty"`
}

// TraceRing 是 Trace 的线程安全环形缓冲区。
type TraceRing struct {
	mu   sync.Mutex
	buf  []*Trace
	cap  int
	pos  int
	full bool
}

// NewTraceRing 创建一个容量为 capacity 的环形缓冲区。
func NewTraceRing(capacity int) *TraceRing {
	return &TraceRing{
		buf: make([]*Trace, capacity),
		cap: capacity,
	}
}

// Push 将一条 trace 加入环形缓冲区。
func (r *TraceRing) Push(t *Trace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.pos] = t
	r.pos++
	if r.pos >= r.cap {
		r.pos = 0
		r.full = true
	}
}

// List 返回最近最多 limit 条 trace，按时间倒序。
func (r *TraceRing) List(limit int) []*Trace {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*Trace
	n := r.cap
	if !r.full {
		n = r.pos
	}
	start := r.pos - 1
	if start < 0 {
		start = r.cap - 1
	}
	count := 0
	for i := 0; i < n && count < limit; i++ {
		idx := (start - i + r.cap) % r.cap
		if r.buf[idx] != nil {
			result = append(result, r.buf[idx])
			count++
		}
	}
	return result
}

// FilterByWorld 返回指定世界的 trace。
func (r *TraceRing) FilterByWorld(worldID string, limit int) []*Trace {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []*Trace
	n := r.cap
	if !r.full {
		n = r.pos
	}
	start := r.pos - 1
	if start < 0 {
		start = r.cap - 1
	}
	count := 0
	for i := 0; i < n && count < limit; i++ {
		idx := (start - i + r.cap) % r.cap
		t := r.buf[idx]
		if t != nil && t.WorldID == worldID {
			result = append(result, t)
			count++
		}
	}
	return result
}

// Len 返回当前存储的 trace 数量。
func (r *TraceRing) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.full {
		return r.cap
	}
	return r.pos
}

// GlobalTraceRing 是全局 trace 环形缓冲区，Debug 模式下使用。
var GlobalTraceRing = NewTraceRing(1000)

// buildDebugTrace 构建一条 Debug 模式下的 LLM 调用 trace。
func buildDebugTrace(
	worldID, requestID string,
	taskType TaskType,
	nodeID string,
	start, llmStart time.Time,
	systemPrompt string,
	messages []ChatMessage,
	rawLLMResponse string,
	resp *InvokeResponse,
	runtime *executionConfig,
	round int,
	errStr string,
) *Trace {
	steps := []TraceStep{
		{Name: "prompt_build", DurationMs: llmStart.Sub(start).Milliseconds()},
		{Name: "llm_call", DurationMs: time.Since(llmStart).Milliseconds()},
		{Name: "parse_execute", DurationMs: time.Since(start).Milliseconds() - llmStart.Sub(start).Milliseconds() - time.Since(llmStart).Milliseconds()},
	}
	return &Trace{
		ID:                     uuid.NewString()[:8],
		WorldID:                worldID,
		RequestID:              requestID,
		TaskType:               taskType,
		NodeID:                 nodeID,
		ConfiguredPipelineMode: configuredPipelineMode(runtime),
		EffectivePipelineMode:  effectivePipelineMode(runtime),
		MaxAnalysisRounds:      maxAnalysisRounds(runtime),
		RoundsUsed:             round + 1,
		Timestamp:              time.Now(),
		DurationMs:             time.Since(start).Milliseconds(),
		SystemPrompt:           truncateForContext(systemPrompt, 2000),
		Messages:               messages,
		RawLLMResponse:         truncateForContext(rawLLMResponse, 5000),
		ParsedActions:          resp.ActionCalls,
		ParsedMemories:         resp.MemoryUpdates,
		SubTasks:               resp.SubTasks,
		WorldChangePlan:        resp.WorldChangePlan,
		Error:                  errStr,
		Round:                  round,
		Steps:                  steps,
	}
}
