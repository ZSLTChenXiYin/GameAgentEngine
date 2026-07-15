package api

import (
	"net/http"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/version"
)

func NewRouter(p *engine.Pipeline) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", Health)
	mux.HandleFunc("GET /api/v1/version", VersionHandler)

	mux.HandleFunc("POST /api/v1/invoke", MakeInvokeHandler(p))
	mux.HandleFunc("POST /api/v1/player/input/interpret", MakePlayerInputInterpretHandler(p))
	mux.HandleFunc("POST /api/v1/actions/callback", CallbackReplayMiddleware(MakeActionCallbackHandler(p)))
	mux.HandleFunc("GET /api/v1/runtime/tasks", MakeListRuntimeTasksHandler())
	mux.HandleFunc("GET /api/v1/runtime/tasks/stats", MakeRuntimeTaskStatsHandler())
	mux.HandleFunc("GET /api/v1/runtime/tasks/pending", MakeListPendingRuntimeTasksHandler())
	mux.HandleFunc("GET /api/v1/runtime/tasks/{task_id}", MakeGetRuntimeTaskHandler())
	mux.HandleFunc("POST /api/v1/runtime/tasks/claim", MakeClaimRuntimeTaskHandler())
	mux.HandleFunc("POST /api/v1/runtime/tasks/start", MakeStartRuntimeTaskHandler())
	mux.HandleFunc("POST /api/v1/runtime/tasks/heartbeat", MakeHeartbeatRuntimeTaskHandler())
	mux.HandleFunc("POST /api/v1/runtime/tasks/heartbeat-timeout/sweep", MakeSweepRuntimeTaskHeartbeatTimeoutHandler())
	mux.HandleFunc("POST /api/v1/runtime/tasks/heartbeat-timeout/requeue", MakeBatchRequeueHeartbeatTimeoutRuntimeTasksHandler())
	mux.HandleFunc("POST /api/v1/runtime/tasks/release", MakeReleaseRuntimeTaskHandler())
	mux.HandleFunc("POST /api/v1/runtime/tasks/requeue", MakeRequeueRuntimeTaskHandler())

	mux.HandleFunc("GET /api/v1/worlds/{world_id}/policy", GetWorldPolicyHandler)
	mux.HandleFunc("PUT /api/v1/worlds/{world_id}/policy", SetWorldPolicyHandler)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/settings", GetWorldSettingsHandler)
	mux.HandleFunc("PUT /api/v1/worlds/{world_id}/settings", SetWorldSettingsHandler)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/state-components", GetStateComponentsHandler)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/state-components/{component_type}", GetStateComponentHandler)
	mux.HandleFunc("PUT /api/v1/worlds/{world_id}/state-components/{component_type}", PutStateComponentHandler)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/timelines", GetTimelinesHandler)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/timelines/latest", GetLatestTimelineHandler)

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/ticks/advance", IdempotencyMiddleware(MakeTickAdvanceHandler(p)))
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/events/impact", IdempotencyMiddleware(MakeEventImpactHandler(p)))
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/scopes/{scope_id}/advance", IdempotencyMiddleware(MakeScopeAdvanceHandler(p)))
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/timeline/replan", IdempotencyMiddleware(MakeTimelineReplanHandler(p)))
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/plan/approve", IdempotencyMiddleware(MakePlanApproveHandler(p)))
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/plan/reject", IdempotencyMiddleware(MakePlanRejectHandler(p)))
	mux.HandleFunc("GET /api/v1/plans/pending", MakeListPendingPlansHandler(p))

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/nodes/{node_id}/autonomous/run", IdempotencyMiddleware(MakeAutonomousRunHandler(p)))
	mux.HandleFunc("POST /api/v1/creator/import", IdempotencyMiddleware(CreatorImportHandler))

	mux.HandleFunc("GET /api/v1/nodes", GetAllNodesHandler)
	mux.HandleFunc("GET /api/v1/nodes/{id}", GetNodeHandler)
	mux.HandleFunc("POST /api/v1/nodes", CreateNodeHandler)
	mux.HandleFunc("PUT /api/v1/nodes/{id}", UpdateNodeHandler)
	mux.HandleFunc("POST /api/v1/nodes/{id}/copy", IdempotencyMiddleware(CopyNodeHandler))
	mux.HandleFunc("DELETE /api/v1/nodes/{id}", DeleteNodeHandler)
	mux.HandleFunc("GET /api/v1/nodes/{node_id}/autonomous", MakeAutonomousConfigGetHandler(p))
	mux.HandleFunc("PUT /api/v1/nodes/{node_id}/autonomous", IdempotencyMiddleware(MakeAutonomousConfigPutHandler(p)))

	mux.HandleFunc("POST /api/v1/components", AddComponentHandler)
	mux.HandleFunc("GET /api/v1/components", GetComponentsHandler)
	mux.HandleFunc("GET /api/v1/components/{id}", GetComponentHandler)
	mux.HandleFunc("PUT /api/v1/components/{id}", UpdateComponentHandler)
	mux.HandleFunc("DELETE /api/v1/components/{id}", DeleteComponentHandler)
	mux.HandleFunc("POST /api/v1/memories", CreateMemoryHandler)
	mux.HandleFunc("GET /api/v1/memories", GetMemoriesHandler)
	mux.HandleFunc("GET /api/v1/memories/{id}", GetMemoryHandler)
	mux.HandleFunc("PUT /api/v1/memories/{id}", UpdateMemoryHandler)
	mux.HandleFunc("DELETE /api/v1/memories/{id}", DeleteMemoryHandler)

	mux.HandleFunc("GET /api/v1/relations", GetRelationsHandler)
	mux.HandleFunc("GET /api/v1/relations/{id}", GetRelationHandler)
	mux.HandleFunc("POST /api/v1/relations", CreateRelationHandler)
	mux.HandleFunc("PUT /api/v1/relations/{id}", UpdateRelationHandler)
	mux.HandleFunc("DELETE /api/v1/relations/{id}", DeleteRelationHandler)

	mux.HandleFunc("GET /api/v1/worlds", GetWorldsHandler)
	mux.HandleFunc("PUT /api/v1/worlds/{world_id}", UpdateWorldHandler)
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/fork", IdempotencyMiddleware(MakeForkWorldHandler(p)))
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/snapshots", MakeListWorldSnapshotsHandler(p))
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/snapshots", IdempotencyMiddleware(MakeCreateWorldSnapshotHandler(p)))
	mux.HandleFunc("DELETE /api/v1/worlds/{world_id}/snapshot", MakeDeleteWorldSnapshotHandler(p))
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/snapshot-metadata", MakeGetWorldSnapshotMetadataHandler(p))
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/snapshot-validation", MakeValidateWorldSnapshotHandler(p))
	mux.HandleFunc("POST /api/v1/worlds/{world_id}/restore", IdempotencyMiddleware(MakeRestoreWorldHandler(p)))

	mux.HandleFunc("GET /api/v1/logs", GetLogsHandler)
	mux.HandleFunc("GET /api/v1/pipeline/stats", GetPipelineStatsHandler)
	mux.HandleFunc("GET /debug/traces", MakeDebugTracesHandler(p))
	mux.HandleFunc("POST /api/v1/memories/propagate", MakePropagateMemoryHandler(p))
	return mux
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{
		"version":        version.Version,
		"min_compatible": version.MinCompatibleVersion,
	})
}
