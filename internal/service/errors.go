package service

import "fmt"

// ErrorKind 描述服务层错误的类别，便于入口层统一映射状态码。
type ErrorKind string

const (
	// 通用错误类别
	ErrorInvalid  ErrorKind = "invalid"
	ErrorNotFound ErrorKind = "not_found"
	ErrorConflict ErrorKind = "conflict"

	// 领域级无效请求错误
	ErrorInvalidNodeType      ErrorKind = "invalid_node_type"
	ErrorInvalidComponentType ErrorKind = "invalid_component_type"
	ErrorInvalidMemoryLevel   ErrorKind = "invalid_memory_level"
	ErrorInvalidRelationType  ErrorKind = "invalid_relation_type"
	ErrorCrossWorldRelation   ErrorKind = "cross_world_relation"
	ErrorImportDuplicateName  ErrorKind = "import_duplicate_node_name"
	ErrorWorldNodeConstraint  ErrorKind = "world_node_constraint"
	ErrorNoUpdates            ErrorKind = "no_updates"
	ErrorParentNotFound       ErrorKind = "parent_not_found"

	// 领域级未找到错误
	ErrorNodeNotFound      ErrorKind = "node_not_found"
	ErrorComponentNotFound ErrorKind = "component_not_found"
	ErrorMemoryNotFound    ErrorKind = "memory_not_found"
	ErrorRelationNotFound  ErrorKind = "relation_not_found"
	ErrorWorldNotFound     ErrorKind = "world_not_found"

	// 领域级冲突错误
	ErrorNodeHasChildren ErrorKind = "node_has_children"
	ErrorParentCycle     ErrorKind = "parent_cycle_detected"
)

// Error 表示一个可分类的领域错误。
type Error struct {
	Kind    ErrorKind
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// IsKind 判断错误是否属于指定类别。
func IsKind(err error, kind ErrorKind) bool {
	serviceErr, ok := err.(*Error)
	return ok && serviceErr.Kind == kind
}

// errorf 创建带指定分类的领域错误。
func errorf(kind ErrorKind, format string, args ...any) error {
	return &Error{Kind: kind, Message: fmt.Sprintf(format, args...)}
}

func invalidf(format string, args ...any) error {
	return errorf(ErrorInvalid, format, args...)
}

func notFoundf(format string, args ...any) error {
	return errorf(ErrorNotFound, format, args...)
}

func conflictf(format string, args ...any) error {
	return errorf(ErrorConflict, format, args...)
}
