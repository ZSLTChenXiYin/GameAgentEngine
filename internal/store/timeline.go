package store

import "gorm.io/gorm"

type InferenceLogQuery struct {
	WorldUUID     string
	NodeUUID      string
	TaskType      string
	Category      string
	EventName     string
	ExecutionMode string
	RequestID     string
	Round         int
	Limit         int
	Offset        int
}

// CreateTimelineTick 写入一条世界时间线刻度记录。
func CreateTimelineTick(m *TimelineModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	return Write(func(db *gorm.DB) error {
		return db.Create(m).Error
	})
}

// GetTimelineTicks 获取某个世界最近的时间线刻度。
func GetTimelineTicks(worldUUID string, limit int) ([]TimelineModel, error) {
	worldID := ResolveWorldUUID(worldUUID)
	var list []TimelineModel
	q := DB.Where("world_id = ?", worldID).Order("tick_number DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveTimelineWorldUUIDs(list)
	}
	return list, err
}

// GetLatestTick 返回某个世界最新的一条刻度记录。
func GetLatestTick(worldUUID string) (*TimelineModel, error) {
	worldID := ResolveWorldUUID(worldUUID)
	var m TimelineModel
	err := DB.Where("world_id = ?", worldID).Order("tick_number DESC").First(&m).Error
	if err == nil {
		l2 := []TimelineModel{m}
		resolveTimelineWorldUUIDs(l2)
		m.WorldUUID = l2[0].WorldUUID
	}
	return &m, err
}

// CreateInferenceLog 写入一次推理调用日志。
func CreateInferenceLog(m *InferenceLogModel) error {
	if m.UUID == "" {
		m.UUID = NewUUID()
	}
	if m.WorldID == 0 && m.WorldUUID != "" {
		m.WorldID = ResolveWorldUUID(m.WorldUUID)
	}
	if m.NodeID == nil && m.NodeUUID != "" {
		if nodeID := ResolveNodeUUID(m.NodeUUID); nodeID != 0 {
			m.NodeID = &nodeID
		}
	}
	return enqueueInferenceLog(m)
}

// GetInferenceLogs 获取推理日志列表，支持分页和按类型过滤。
// worldUUID 非空时按世界过滤；taskType 非空时按任务类型过滤。
// limit <= 0 表示不限制；offset < 0 时按 0 处理。
func GetInferenceLogs(worldUUID string, limit, offset int, taskType string) ([]InferenceLogModel, error) {
	return GetInferenceLogsByQuery(InferenceLogQuery{WorldUUID: worldUUID, Limit: limit, Offset: offset, TaskType: taskType})
}

// GetInferenceLogsByQuery 获取推理日志列表，支持按结构化字段组合查询。
func GetInferenceLogsByQuery(query InferenceLogQuery) ([]InferenceLogModel, error) {
	if err := FlushLogSink(); err != nil {
		return nil, err
	}
	var list []InferenceLogModel
	q := DB.Order("created_at DESC")
	if query.WorldUUID != "" {
		worldID := ResolveWorldUUID(query.WorldUUID)
		q = q.Where("world_id = ?", worldID)
	}
	if query.NodeUUID != "" {
		nodeID := ResolveNodeUUID(query.NodeUUID)
		q = q.Where("node_id = ?", nodeID)
	}
	if query.TaskType != "" {
		q = q.Where("task_type = ?", query.TaskType)
	}
	if query.Category != "" {
		q = q.Where("category = ?", query.Category)
	}
	if query.EventName != "" {
		q = q.Where("event_name = ?", query.EventName)
	}
	if query.ExecutionMode != "" {
		q = q.Where("execution_mode = ?", query.ExecutionMode)
	}
	if query.RequestID != "" {
		q = q.Where("request_id = ?", query.RequestID)
	}
	if query.Round > 0 {
		q = q.Where("round = ?", query.Round)
	}
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}
	err := q.Find(&list).Error
	if err == nil && len(list) > 0 {
		resolveLogNodeUUIDs(list)
	}
	return list, err
}
