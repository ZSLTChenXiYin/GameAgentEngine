package api

import (
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
)

// handleServiceError 将 service 层的分类错误映射到合适的 HTTP 状态码和错误码。
func handleServiceError(w http.ResponseWriter, err error) {
	serviceCode := service.ErrorCode(err)
	switch {
	// 通用无效请求
	case service.IsKind(err, service.ErrorInvalid):
		if serviceCode != "" {
			errorJSONCode(w, http.StatusBadRequest, serviceCode, err.Error())
			return
		}
		errorJSONCode(w, http.StatusBadRequest, "invalid_request", err.Error())

	// 领域级无效请求
	case service.IsKind(err, service.ErrorInvalidNodeType):
		errorJSONCode(w, http.StatusBadRequest, "invalid_node_type", err.Error())
	case service.IsKind(err, service.ErrorInvalidComponentType):
		errorJSONCode(w, http.StatusBadRequest, "invalid_component_type", err.Error())
	case service.IsKind(err, service.ErrorInvalidMemoryLevel):
		errorJSONCode(w, http.StatusBadRequest, "invalid_memory_level", err.Error())
	case service.IsKind(err, service.ErrorInvalidRelationType):
		errorJSONCode(w, http.StatusBadRequest, "invalid_relation_type", err.Error())
	case service.IsKind(err, service.ErrorCrossWorldRelation):
		errorJSONCode(w, http.StatusBadRequest, "cross_world_relation", err.Error())
	case service.IsKind(err, service.ErrorImportDuplicateName):
		errorJSONCode(w, http.StatusBadRequest, "import_duplicate_node_name", err.Error())
	case service.IsKind(err, service.ErrorWorldNodeConstraint):
		errorJSONCode(w, http.StatusBadRequest, "world_node_constraint", err.Error())
	case service.IsKind(err, service.ErrorNoUpdates):
		errorJSONCode(w, http.StatusBadRequest, "no_updates", err.Error())
	case service.IsKind(err, service.ErrorParentNotFound):
		errorJSONCode(w, http.StatusBadRequest, "parent_not_found", err.Error())

	// 通用未找到
	case service.IsKind(err, service.ErrorNotFound):
		errorJSONCode(w, http.StatusNotFound, "not_found", err.Error())

	// 领域级未找到
	case service.IsKind(err, service.ErrorNodeNotFound):
		errorJSONCode(w, http.StatusNotFound, "node_not_found", err.Error())
	case service.IsKind(err, service.ErrorComponentNotFound):
		errorJSONCode(w, http.StatusNotFound, "component_not_found", err.Error())
	case service.IsKind(err, service.ErrorMemoryNotFound):
		errorJSONCode(w, http.StatusNotFound, "memory_not_found", err.Error())
	case service.IsKind(err, service.ErrorRelationNotFound):
		errorJSONCode(w, http.StatusNotFound, "relation_not_found", err.Error())
	case service.IsKind(err, service.ErrorWorldNotFound):
		errorJSONCode(w, http.StatusNotFound, "world_not_found", err.Error())

	// 通用冲突
	case service.IsKind(err, service.ErrorConflict):
		if serviceCode != "" {
			errorJSONCode(w, http.StatusConflict, serviceCode, err.Error())
			return
		}
		errorJSONCode(w, http.StatusConflict, "conflict", err.Error())

	// 领域级冲突
	case service.IsKind(err, service.ErrorNodeHasChildren):
		errorJSONCode(w, http.StatusConflict, "node_has_children", err.Error())
	case service.IsKind(err, service.ErrorParentCycle):
		errorJSONCode(w, http.StatusConflict, "parent_cycle_detected", err.Error())

	default:
		errorJSONCode(w, http.StatusInternalServerError, "internal_error", err.Error())
	}
}
