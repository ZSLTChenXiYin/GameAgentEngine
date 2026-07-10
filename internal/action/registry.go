package action

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// Registry 负责注册动作实现并跟踪异步回调状态。
type Registry struct {
	mu        sync.RWMutex
	syncActs  map[string]SyncAction
	asyncActs map[string]AsyncAction
	pending   map[string]*ActionCallRecord
}

// NewRegistry 创建一个空的动作注册表。
func NewRegistry() *Registry {
	return &Registry{
		syncActs:  make(map[string]SyncAction),
		asyncActs: make(map[string]AsyncAction),
		pending:   make(map[string]*ActionCallRecord),
	}
}

// RegisterSync 注册一个同步动作实现。
func (r *Registry) RegisterSync(a SyncAction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := a.ID()
	if _, exists := r.syncActs[id]; exists {
		return fmt.Errorf("sync action %s already registered", id)
	}
	r.syncActs[id] = a
	return nil
}

// RegisterAsync 注册一个异步动作实现。
func (r *Registry) RegisterAsync(a AsyncAction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	id := a.ID()
	if _, exists := r.asyncActs[id]; exists {
		return fmt.Errorf("async action %s already registered", id)
	}
	r.asyncActs[id] = a
	return nil
}

// IsSync 判断动作是否已注册为同步动作。
func (r *Registry) IsSync(actionID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.syncActs[actionID]
	return ok
}

// IsAsync 判断动作是否已注册为异步动作。
func (r *Registry) IsAsync(actionID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.asyncActs[actionID]
	return ok
}

// Exists 判断动作 ID 是否已存在于注册表中。
func (r *Registry) Exists(actionID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, sOk := r.syncActs[actionID]
	_, aOk := r.asyncActs[actionID]
	return sOk || aOk
}

// ExecuteSync 执行一个已注册的同步动作。
func (r *Registry) ExecuteSync(actionID string, args map[string]any) (any, error) {
	r.mu.RLock()
	a, ok := r.syncActs[actionID]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("sync action %s not found", actionID)
	}
	return a.Execute(args)
}

type CallbackMetadata struct {
	NodeID            string
	WorldID           string
	RequestID         string
	ResumeExecutionID string
}

// CreateCallback 为异步动作生成回调记录并返回回调 ID。
func (r *Registry) CreateCallback(actionID string, args map[string]any) string {
	return r.CreateCallbackWithMetadata(actionID, args, CallbackMetadata{})
}

// CreateCallbackWithMetadata generates a callback record and persists it.
func (r *Registry) CreateCallbackWithMetadata(actionID string, args map[string]any, meta CallbackMetadata) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	cb := &ActionCallRecord{
		CallbackID:        uuid.NewString(),
		ActionID:          actionID,
		NodeID:            meta.NodeID,
		WorldID:           meta.WorldID,
		RequestID:         meta.RequestID,
		ResumeExecutionID: meta.ResumeExecutionID,
		Args:              args,
		Status:            "pending",
	}
	r.pending[cb.CallbackID] = cb
	argsJSON := marshalJSON(args)
	if err := store.CreateAsyncCallbackRecord(&store.AsyncCallbackRecordModel{
		CallbackID:        cb.CallbackID,
		ActionID:          actionID,
		Status:            "pending",
		NodeUUID:          meta.NodeID,
		WorldUUID:         meta.WorldID,
		RequestID:         meta.RequestID,
		ResumeExecutionID: meta.ResumeExecutionID,
		ArgsJSON:          argsJSON,
	}); err != nil {
		// Keep the in-memory path alive even if persistence fails.
	}
	return cb.CallbackID
}

// HandleCallback 处理游戏侧上报的异步动作执行结果。
func (r *Registry) HandleCallback(callbackID string, status string, result any) (*ActionCallRecord, error) {
	r.mu.Lock()
	rec, ok := r.pending[callbackID]
	if ok {
		delete(r.pending, callbackID)
	}
	r.mu.Unlock()

	if !ok {
		model, err := store.GetAsyncCallbackRecord(callbackID)
		if err != nil {
			return nil, fmt.Errorf("callback %s not found", callbackID)
		}
		var args map[string]any
		if model.ArgsJSON != "" {
			_ = json.Unmarshal([]byte(model.ArgsJSON), &args)
		}
		rec = &ActionCallRecord{
			CallbackID:        model.CallbackID,
			ActionID:          model.ActionID,
			NodeID:            model.NodeUUID,
			WorldID:           model.WorldUUID,
			RequestID:         model.RequestID,
			ResumeExecutionID: model.ResumeExecutionID,
			Args:              args,
			Status:            model.Status,
		}
	}

	rec.Status = status
	rec.Result = result
	resultJSON := marshalJSON(result)
	errMsg := ""
	if status == "failed" {
		if s, ok := result.(string); ok {
			errMsg = s
		} else {
			errMsg = resultJSON
		}
	}
	if err := store.CompleteAsyncCallbackRecord(callbackID, status, resultJSON, errMsg); err != nil {
		return nil, err
	}

	r.mu.RLock()
	a, ok := r.asyncActs[rec.ActionID]
	r.mu.RUnlock()
	if ok {
		if err := a.OnResult(callbackID, status, result); err != nil {
			return nil, err
		}
	}
	return rec, nil
}

// List 返回当前注册表中所有动作的标识列表。
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var ids []string
	for id := range r.syncActs {
		ids = append(ids, id)
	}
	for id := range r.asyncActs {
		ids = append(ids, id+"(async)")
	}
	return ids
}

func marshalJSON(value any) string {
	if value == nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

