package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DAGInstance 管理一次推理中声明的子任务 DAG。
// 生命周期限于单次 Execute() 调用，不做跨请求持久化。
type DAGInstance struct {
	ID              string                      // DAG 唯一标识
	Tree            *TaskTree                   // 关联的任务节点树
	llmProvider     LLMProvider                 // LLM 提供者（用于 summarize 模式）
	MaxRetries      int                         // 子任务最大重试次数
	TimeoutDuration time.Duration               // 子任务超时时间
	tasks           map[string]*subTaskState    // label -> 子任务状态
	completed       map[string]bool             // 已完成的子任务
	results         map[string]*InvokeResponse  // label -> 子任务结果
	failed          map[string]string           // label -> 错误信息
}

type subTaskState struct {
	Decl   SubTaskDeclaration
	Status SubTaskStatus
}

// NewDAGInstance 创建子任务 DAG 实例。
func NewDAGInstance(tree *TaskTree, llmProvider LLMProvider, maxRetries int, timeout time.Duration) *DAGInstance {
	if maxRetries <= 0 {
		maxRetries = 2
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &DAGInstance{
		llmProvider:     llmProvider,
		MaxRetries:      maxRetries,
		TimeoutDuration: timeout,
		ID:        uuid.NewString(),
		Tree:      tree,
		tasks:     make(map[string]*subTaskState),
		completed: make(map[string]bool),
		results:   make(map[string]*InvokeResponse),
		failed:    make(map[string]string),
	}
}

// Register 注册一个子任务声明。
func (d *DAGInstance) Register(decl SubTaskDeclaration) error {
	if _, exists := d.tasks[decl.Label]; exists {
		return fmt.Errorf("sub-task %q already registered", decl.Label)
	}
	status := SubTaskReady
	if len(decl.DependsOn) > 0 {
		status = SubTaskPending
	}
	d.tasks[decl.Label] = &subTaskState{
		Decl:   decl,
		Status: status,
	}
	return nil
}

// ReadyTasks 返回所有依赖已满足的待执行子任务。
func (d *DAGInstance) ReadyTasks() []SubTaskDeclaration {
	var ready []SubTaskDeclaration
	for _, st := range d.tasks {
		if st.Status == SubTaskReady {
			ready = append(ready, st.Decl)
		}
	}
	return ready
}

// HasReady 判断是否有就绪的子任务。
func (d *DAGInstance) HasReady() bool {
	for _, st := range d.tasks {
		if st.Status == SubTaskReady {
			return true
		}
	}
	return false
}

// MarkRunning 将子任务标记为运行中。
func (d *DAGInstance) MarkRunning(label string) {
	if st, ok := d.tasks[label]; ok {
		st.Status = SubTaskRunning
	}
}

// OnTaskComplete 标记子任务完成，并检查依赖是否满足以解锁新的就绪任务。
func (d *DAGInstance) OnTaskComplete(label string, resp *InvokeResponse) {
	if st, ok := d.tasks[label]; ok {
		st.Status = SubTaskCompleted
		d.completed[label] = true
		d.results[label] = resp

		// 检查是否有因本次完成而解锁的子任务
		for _, st2 := range d.tasks {
			if st2.Status == SubTaskPending && allDepsMet(st2.Decl.DependsOn, d.completed) {
				st2.Status = SubTaskReady
			}
		}
	}
}

// OnTaskFailed 标记子任务失败。
func (d *DAGInstance) OnTaskFailed(label string, err error) {
	if st, ok := d.tasks[label]; ok {
		st.Status = SubTaskFailed
		d.failed[label] = err.Error()
		// 失败不阻塞其他子任务
		for _, st2 := range d.tasks {
			if st2.Status == SubTaskPending && allDepsMet(st2.Decl.DependsOn, d.completed) {
				st2.Status = SubTaskReady
			}
		}
	}
}

// AllDone 返回是否所有子任务都已结束（完成或失败）。
func (d *DAGInstance) AllDone() bool {
	for _, st := range d.tasks {
		if st.Status != SubTaskCompleted && st.Status != SubTaskFailed {
			return false
		}
	}
	return true
}

// MergeResults 将所有子任务的结果按 MergeMode 汇聚。
func (d *DAGInstance) MergeResults() *InvokeResponse {
	merged := &InvokeResponse{}

	// 按注册顺序汇聚（map 遍历无序，先收集顺序）
	order := make([]string, 0, len(d.tasks))
	for label := range d.tasks {
		order = append(order, label)
	}

	for _, label := range order {
		st := d.tasks[label]
		resp := d.results[label]
		if resp == nil {
			continue
		}

		switch st.Decl.MergeMode {
		case "override":
			// 后完成的覆盖之前的结果
			merged = resp
		case "summarize":
			summary := d.summarizeResults()
			merged.Reply = mergeStrings(merged.Reply, summary, "\n")
			merged.ActionCalls = append(merged.ActionCalls, resp.ActionCalls...)
			merged.MemoryUpdates = append(merged.MemoryUpdates, resp.MemoryUpdates...)
		default:
			// "append" 或空值：追加模式
			merged.ActionCalls = append(merged.ActionCalls, resp.ActionCalls...)
			merged.MemoryUpdates = append(merged.MemoryUpdates, resp.MemoryUpdates...)
			if merged.Reply == "" {
				merged.Reply = resp.Reply
			} else {
				merged.Reply = merged.Reply + "\n" + resp.Reply
			}
		}
	}

	// 如果有失败的子任务，在 Reply 中附加错误信息
	if len(d.failed) > 0 {
		var errParts []string
		for label, errMsg := range d.failed {
			errParts = append(errParts, fmt.Sprintf("[子任务 %s 失败] %s", label, errMsg))
		}
		merged.Reply = merged.Reply + "\n" + strings.Join(errParts, "\n")
	}

	return merged
}

// allDepsMet 检查子任务的所有依赖是否都已标记为完成。
// ExecuteWithRetry 带重试和超时地执行一个子任务。
// 返回 (响应, 是否成功)。
func (d *DAGInstance) ExecuteWithRetry(label string, fn func() (*InvokeResponse, error)) (*InvokeResponse, bool) {
	for attempt := 0; attempt <= d.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("[dag] retrying sub-task %s (attempt %d/%d)", label, attempt, d.MaxRetries)
		}

		ctx, cancel := context.WithTimeout(context.Background(), d.TimeoutDuration)
		defer cancel()

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

// summarizeResults 调用 LLM 对所有子任务结果做语义摘要。
func (d *DAGInstance) summarizeResults() string {
	if d.llmProvider == nil {
		return "[summarize] 无 LLM 提供者"
	}
	var parts []string
	parts = append(parts, "以下是对多个并行子任务结果的摘要。子任务列表：")
	for label, resp := range d.results {
		if resp != nil {
			reply := resp.Reply
			if len(reply) > 200 {
				reply = reply[:200] + "..."
			}
			parts = append(parts, "- " + label + ": " + reply)
		}
	}
	if len(d.failed) > 0 {
		parts = append(parts, "失败子任务: " + fmt.Sprint(d.failed))
	}
	parts = append(parts, "请根据上述子任务结果生成一个统一的摘要。")

	prompt := strings.Join(parts, "\n")
	if resp, err := d.llmProvider.Chat(&LLMChatRequest{SystemPrompt: prompt}); err == nil {
		return resp.Content
	}
	return "[summarize] LLM 摘要失败"
}

