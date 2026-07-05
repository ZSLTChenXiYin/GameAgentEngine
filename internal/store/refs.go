package store

// resolveComponentNodeUUIDs 批量填充 ComponentModel 的 NodeUUID。
func resolveComponentNodeUUIDs(list []ComponentModel) {
	if len(list) == 0 {
		return
	}
	ids := make([]int64, len(list))
	for i, c := range list {
		ids[i] = c.NodeID
	}
	var nodes []NodeModel
	DB.Select("id,uuid").Where("id IN ?", ids).Find(&nodes)
	m := map[int64]string{}
	for _, n := range nodes {
		m[n.ID] = n.UUID
	}
	for i, c := range list {
		if u, ok := m[c.NodeID]; ok {
			list[i].NodeUUID = u
		}
	}
}

// resolveRelationRefs 批量填充 RelationModel 的 UUID 引用。
func resolveRelationRefs(list []RelationModel) {
	if len(list) == 0 {
		return
	}
	// 收集所有 int64 ID
	wids := map[int64]bool{}
	sids := map[int64]bool{}
	tids := map[int64]bool{}
	for _, r := range list {
		wids[r.WorldID] = true
		sids[r.SourceID] = true
		tids[r.TargetID] = true
	}
	nodesByID := map[int64]string{}
	for _, m := range [](map[int64]bool){wids, sids, tids} {
		ids := make([]int64, 0, len(m))
		for id := range m {
			ids = append(ids, id)
		}
		var ns []NodeModel
		DB.Select("id,uuid").Where("id IN ?", ids).Find(&ns)
		for _, n := range ns {
			nodesByID[n.ID] = n.UUID
		}
	}
	for i, r := range list {
		if u, ok := nodesByID[r.WorldID]; ok {
			list[i].WorldUUID = u
		}
		if u, ok := nodesByID[r.SourceID]; ok {
			list[i].SourceUUID = u
		}
		if u, ok := nodesByID[r.TargetID]; ok {
			list[i].TargetUUID = u
		}
	}
}

// resolveMemoryNodeUUIDs 批量填充 MemoryModel 的 NodeUUID。
func resolveMemoryNodeUUIDs(list []MemoryModel) {
	if len(list) == 0 {
		return
	}
	ids := make([]int64, len(list))
	for i, m := range list {
		ids[i] = m.NodeID
	}
	var nodes []NodeModel
	DB.Select("id,uuid").Where("id IN ?", ids).Find(&nodes)
	mm := map[int64]string{}
	for _, n := range nodes {
		mm[n.ID] = n.UUID
	}
	for i, c := range list {
		if u, ok := mm[c.NodeID]; ok {
			list[i].NodeUUID = u
		}
	}
}

// resolveLogNodeUUIDs 批量填充 InferenceLogModel 的 NodeUUID / WorldUUID。
func resolveLogNodeUUIDs(list []InferenceLogModel) {
	if len(list) == 0 {
		return
	}
	nidSet := map[int64]bool{}
	widSet := map[int64]bool{}
	for _, l := range list {
		widSet[l.WorldID] = true
		if l.NodeID != nil {
			nidSet[*l.NodeID] = true
		}
	}
	all := map[int64]string{}
	for _, ids := range [](map[int64]bool){nidSet, widSet} {
		if len(ids) == 0 {
			continue
		}
		iids := make([]int64, 0, len(ids))
		for id := range ids {
			iids = append(iids, id)
		}
		var ns []NodeModel
		DB.Select("id,uuid").Where("id IN ?", iids).Find(&ns)
		for _, n := range ns {
			all[n.ID] = n.UUID
		}
	}
	for i, l := range list {
		if u, ok := all[l.WorldID]; ok {
			list[i].WorldUUID = u
		}
		if l.NodeID != nil {
			if u, ok := all[*l.NodeID]; ok {
				list[i].NodeUUID = u
			}
		}
	}
}

// resolvePropagationWorldUUIDs 批量填充 PropagationChainModel 的 WorldUUID。
func resolvePropagationWorldUUIDs(list []PropagationChainModel) {
	if len(list) == 0 {
		return
	}
	wids := make([]int64, len(list))
	for i, p := range list {
		wids[i] = p.WorldID
	}
	var nodes []NodeModel
	DB.Select("id,uuid").Where("id IN ?", wids).Find(&nodes)
	m := map[int64]string{}
	for _, n := range nodes {
		m[n.ID] = n.UUID
	}
	for i, p := range list {
		if u, ok := m[p.WorldID]; ok {
			list[i].WorldUUID = u
		}
	}
}

// resolveTimelineWorldUUIDs 批量填充 TimelineModel 的 WorldUUID。
func resolveTimelineWorldUUIDs(list []TimelineModel) {
	if len(list) == 0 {
		return
	}
	wids := make([]int64, len(list))
	for i, t := range list {
		wids[i] = t.WorldID
	}
	var nodes []NodeModel
	DB.Select("id,uuid").Where("id IN ?", wids).Find(&nodes)
	m := map[int64]string{}
	for _, n := range nodes {
		m[n.ID] = n.UUID
	}
	for i, t := range list {
		if u, ok := m[t.WorldID]; ok {
			list[i].WorldUUID = u
		}
	}
}

// resolveNodeParentUUIDs 批量填充 NodeModel 的 ParentUUID 和 WorldUUID。
func resolveNodeParentUUIDs(list []NodeModel) {
	if len(list) == 0 {
		return
	}
	// Collect parent IDs
	pidSet := map[int64]bool{}
	widSet := map[int64]bool{}
	for _, n := range list {
		widSet[n.WorldID] = true
		if n.ParentID != nil {
			pidSet[*n.ParentID] = true
		}
	}
	// Merge parent and world lookups
	allIDs := map[int64]string{}
	for _, ids := range [](map[int64]bool){pidSet, widSet} {
		if len(ids) == 0 { continue }
		iids := make([]int64, 0, len(ids))
		for id := range ids { iids = append(iids, id) }
		var nodes []NodeModel
		DB.Select("id,uuid").Where("id IN ?", iids).Find(&nodes)
		for _, n := range nodes { allIDs[n.ID] = n.UUID }
	}
	for i, n := range list {
		if n.ParentID != nil {
			if u, ok := allIDs[*n.ParentID]; ok {
				list[i].ParentUUID = &u
			}
		}
		if u, ok := allIDs[n.WorldID]; ok {
			list[i].WorldUUID = u
		}
	}
}
