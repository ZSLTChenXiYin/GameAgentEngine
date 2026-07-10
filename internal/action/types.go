package action

// Action 是所有游戏动作的基础接口。
type Action interface {
	ID() string
	Validate(args map[string]any) error
}

// SyncAction 表示在引擎管线内立即执行的同步动作。
type SyncAction interface {
	Action
	Execute(args map[string]any) (any, error)
}

// AsyncAction 表示返回给游戏侧异步执行的动作。
type AsyncAction interface {
	Action
	// OnResult 在游戏侧通过回调上报结果时被触发。
	OnResult(callbackID string, status string, result any) error
}

// Result 描述一次动作执行的通用结果。
type Result struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ActionCallRecord 记录待回调的异步动作调用信息。
type ActionCallRecord struct {
	CallbackID        string
	ActionID          string
	NodeID            string
	WorldID           string
	RequestID         string
	ResumeExecutionID string
	Args              map[string]any
	Status            string // pending, success, failed
	Result            any
}
