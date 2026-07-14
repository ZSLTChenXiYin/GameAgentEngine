package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DAGInstance manages one sub-task DAG for a single Execute call.
type DAGInstance struct {
	ID              string
	Tree            *TaskTree
	llmProvider     LLMProvider
	MaxRetries      int
	TimeoutDuration time.Duration
	order           []string
	tasks           map[string]*subTaskState
	completed       map[string]bool
	results         map[string]*InvokeResponse
	failed          map[string]string
}

type subTaskState struct {
	Decl   SubTaskDeclaration
	Status SubTaskStatus
}

// NewDAGInstance creates a sub-task DAG runtime.
func NewDAGInstance(tree *TaskTree, llmProvider LLMProvider, maxRetries int, timeout time.Duration) *DAGInstance {
	if maxRetries <= 0 {
		maxRetries = 2
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &DAGInstance{
		ID:              uuid.NewString(),
		Tree:            tree,
		llmProvider:     llmProvider,
		MaxRetries:      maxRetries,
		TimeoutDuration: timeout,
		order:           make([]string, 0, 8),
		tasks:           make(map[string]*subTaskState),
		completed:       make(map[string]bool),
		results:         make(map[string]*InvokeResponse),
		failed:          make(map[string]string),
	}
}

// Register adds one declared sub-task.
func (d *DAGInstance) Register(decl SubTaskDeclaration) error {
	if _, exists := d.tasks[decl.Label]; exists {
		return fmt.Errorf("sub-task %q already registered", decl.Label)
	}
	status := SubTaskReady
	if len(decl.DependsOn) > 0 {
		status = SubTaskPending
	}
	d.tasks[decl.Label] = &subTaskState{Decl: decl, Status: status}
	d.order = append(d.order, decl.Label)
	return nil
}

// ReadyTasks returns all ready sub-tasks.
func (d *DAGInstance) ReadyTasks() []SubTaskDeclaration {
	var ready []SubTaskDeclaration
	for _, label := range d.orderedTaskLabels() {
		st := d.tasks[label]
		if st == nil {
			continue
		}
		if st.Status == SubTaskReady {
			ready = append(ready, st.Decl)
		}
	}
	return ready
}

// HasReady reports whether any sub-task is ready.
func (d *DAGInstance) HasReady() bool {
	for _, label := range d.orderedTaskLabels() {
		st := d.tasks[label]
		if st == nil {
			continue
		}
		if st.Status == SubTaskReady {
			return true
		}
	}
	return false
}

// MarkRunning marks one sub-task as running.
func (d *DAGInstance) MarkRunning(label string) {
	if st, ok := d.tasks[label]; ok {
		st.Status = SubTaskRunning
	}
}

// OnTaskComplete marks one sub-task complete and unlocks dependents.
func (d *DAGInstance) OnTaskComplete(label string, resp *InvokeResponse) {
	if st, ok := d.tasks[label]; ok {
		st.Status = SubTaskCompleted
		d.completed[label] = true
		d.results[label] = resp

		for _, taskLabel := range d.orderedTaskLabels() {
			item := d.tasks[taskLabel]
			if item == nil {
				continue
			}
			if item.Status == SubTaskPending && allDepsMet(item.Decl.DependsOn, d.completed) {
				item.Status = SubTaskReady
			}
		}
	}
}

// OnTaskFailed marks one sub-task failed.
func (d *DAGInstance) OnTaskFailed(label string, err error) {
	if st, ok := d.tasks[label]; ok {
		st.Status = SubTaskFailed
		d.failed[label] = err.Error()
		for _, taskLabel := range d.orderedTaskLabels() {
			item := d.tasks[taskLabel]
			if item == nil {
				continue
			}
			if item.Status == SubTaskPending && allDepsMet(item.Decl.DependsOn, d.completed) {
				item.Status = SubTaskReady
			}
		}
	}
}

// AllDone reports whether every sub-task has reached a terminal state.
func (d *DAGInstance) AllDone() bool {
	for _, label := range d.orderedTaskLabels() {
		st := d.tasks[label]
		if st == nil {
			continue
		}
		if st.Status != SubTaskCompleted && st.Status != SubTaskFailed {
			return false
		}
	}
	return true
}

// MergeResults merges sub-task outputs according to each declaration's merge_mode.
func (d *DAGInstance) MergeResults() *InvokeResponse {
	merged := &InvokeResponse{}

	for _, label := range d.orderedTaskLabels() {
		st := d.tasks[label]
		resp := d.results[label]
		if st == nil || resp == nil {
			continue
		}

		switch st.Decl.MergeMode {
		case "override":
			merged = resp
		case "summarize":
			summary := d.summarizeResults()
			merged.Reply = mergeStrings(merged.Reply, summary, "\n")
			merged.ActionCalls = append(merged.ActionCalls, resp.ActionCalls...)
			merged.MemoryUpdates = append(merged.MemoryUpdates, resp.MemoryUpdates...)
		default:
			merged.ActionCalls = append(merged.ActionCalls, resp.ActionCalls...)
			merged.MemoryUpdates = append(merged.MemoryUpdates, resp.MemoryUpdates...)
			if merged.Reply == "" {
				merged.Reply = resp.Reply
			} else {
				merged.Reply = merged.Reply + "\n" + resp.Reply
			}
		}
	}

	if len(d.failed) > 0 {
		var errParts []string
		for label, errMsg := range d.failed {
			errParts = append(errParts, fmt.Sprintf("[子任务 %s 失败] %s", label, errMsg))
		}
		merged.Reply = mergeStrings(merged.Reply, strings.Join(errParts, "\n"), "\n")
	}

	return merged
}

// ExecuteWithRetry runs one sub-task with retry and timeout.
func (d *DAGInstance) ExecuteWithRetry(label string, fn func() (*InvokeResponse, error)) (*InvokeResponse, bool) {
	for attempt := 0; attempt <= d.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[dag] retrying sub-task %s (attempt %d/%d)", label, attempt, d.MaxRetries)
		}

		ctx, cancel := context.WithTimeout(context.Background(), d.TimeoutDuration)
		done := make(chan struct{}, 1)
		var resp *InvokeResponse
		var err error

		go func() {
			resp, err = fn()
			close(done)
		}()

		select {
		case <-done:
			cancel()
			if err != nil {
				log.Printf("[dag] sub-task %s attempt %d failed: %v", label, attempt+1, err)
				continue
			}
			return resp, true
		case <-ctx.Done():
			cancel()
			log.Printf("[dag] sub-task %s attempt %d timed out", label, attempt+1)
			continue
		}
	}
	return nil, false
}

func allDepsMet(deps []string, completed map[string]bool) bool {
	for _, dep := range deps {
		if !completed[dep] {
			return false
		}
	}
	return true
}

func mergeStrings(a, b, sep string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + sep + b
}

func (d *DAGInstance) orderedTaskLabels() []string {
	if len(d.order) == 0 {
		return nil
	}
	labels := make([]string, 0, len(d.order))
	for _, label := range d.order {
		if _, ok := d.tasks[label]; ok {
			labels = append(labels, label)
		}
	}
	return labels
}

func (d *DAGInstance) summarizeBaseLines() []string {
	baseLines := []string{
		"You are merging multiple engine sub-task results into one final reply.",
		`Return JSON only. Use either {"reply":"..."} or {"reply":"...","request_data":{"label":"...","target":"store","queries":[...]}}.`,
		"This standalone DAG summarize helper only supports store-backed request_data continuation.",
		"If more engine-side data is required before concluding, request it with request_data targeting store.",
		"For game_client callbacks and pause/resume, use the pipeline-driven DAG summarize path instead of this helper.",
		"========== Sub-task Results ==========",
	}

	for _, label := range d.orderedTaskLabels() {
		resp := d.results[label]
		if resp == nil {
			continue
		}
		reply := strings.TrimSpace(resp.Reply)
		if len(reply) > 200 {
			reply = reply[:200] + "..."
		}
		baseLines = append(baseLines, "- "+label+": "+reply)
	}

	if len(d.failed) > 0 {
		baseLines = append(baseLines, "========== Failed Sub-tasks ==========")
		for _, label := range d.orderedTaskLabels() {
			if errMsg := strings.TrimSpace(d.failed[label]); errMsg != "" {
				baseLines = append(baseLines, fmt.Sprintf("- %s: %s", label, errMsg))
			}
		}
	}

	return baseLines
}

// summarizeResults asks the LLM to merge sub-task outcomes into one reply.
// It is a lightweight standalone helper: store request_data loops are supported,
// while pipeline-managed game_client pause/resume flows live in pipeline.go.
func (d *DAGInstance) summarizeResults() string {
	if d.llmProvider == nil {
		return "[summarize] 无 LLM 提供者"
	}

	pipeline := NewPipeline(d.llmProvider)
	baseLines := d.summarizeBaseLines()

	stateLines := make([]string, 0, 8)
	for round := 0; round < 4; round++ {
		roundLines := append([]string{}, baseLines...)
		if len(stateLines) > 0 {
			roundLines = append(roundLines, "", "========== Summarize Context ==========")
			roundLines = append(roundLines, stateLines...)
		}

		prompt := strings.Join(roundLines, "\n")
		resp, err := d.llmProvider.Chat(&LLMChatRequest{SystemPrompt: prompt})
		if err != nil {
			log.Printf("[dag] summarize round %d failed: %v", round+1, err)
			return "[summarize] LLM 摘要失败"
		}

		parsed := pipeline.parseLLMJSON(resp.Content)
		rawRequestData := strings.TrimSpace(parsed.RawRequestData)
		if rawRequestData != "" && rawRequestData != "null" {
			var dr DataRequest
			if err := json.Unmarshal([]byte(rawRequestData), &dr); err != nil {
				stateLines = append(stateLines, "[summarize request_data invalid] "+err.Error())
				continue
			}
			if strings.TrimSpace(dr.Target) == "" {
				dr.Target = "store"
			}
			if dr.Target != "store" {
				stateLines = append(stateLines, "[summarize request_data blocked] standalone DAG summarize only supports store target; use pipeline DAG summarize for game_client callbacks")
				continue
			}
			if len(dr.Queries) == 0 {
				stateLines = append(stateLines, "[summarize request_data blocked] no queries requested")
				continue
			}
			stateLines = append(stateLines, "[summarize request_data] "+firstNonEmpty(dr.Label, "store_query"))
			result := pipeline.handleDataRequest(nil, &dr)
			if strings.TrimSpace(result) == "" {
				result = "[no data returned]"
			}
			stateLines = append(stateLines, result)
			continue
		}

		if reply := strings.TrimSpace(parsed.Reply); reply != "" {
			return reply
		}
		if raw := strings.TrimSpace(resp.Content); raw != "" {
			return raw
		}
	}

	return "[summarize] LLM 摘要失败"
}
