package engine

import (
	"log"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

type memoryWriteRequest struct {
	NodeUUID string
	NodeID   int64
	Content  string
	Level    MemoryLevel
	Tags     string
}

func persistMemoryBatch(items []memoryWriteRequest) error {
	if len(items) == 0 {
		return nil
	}
	models := make([]store.MemoryModel, 0, len(items))
	for _, item := range items {
		if item.NodeID == 0 || item.Content == "" {
			continue
		}
		models = append(models, store.MemoryModel{
			UUID:     store.NewUUID(),
			NodeID:   item.NodeID,
			NodeUUID: item.NodeUUID,
			Content:  item.Content,
			Level:    string(item.Level),
			Tags:     item.Tags,
		})
	}
	if len(models) == 0 {
		return nil
	}
	if err := store.CreateMemoriesBulk(models); err == nil {
		return nil
	}

	// Fall back to row-by-row writes, skipping bad rows instead of failing entirely.
	var lastErr error
	for i := range models {
		if err := store.CreateMemory(&models[i]); err != nil {
			logMemoryBatchFailure("persist row "+models[i].NodeUUID, err)
			lastErr = err
		}
	}
	return lastErr
}

func filterNewPropagationTargets(targets []memoryWriteRequest) []memoryWriteRequest {
	if len(targets) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(targets))
	filtered := make([]memoryWriteRequest, 0, len(targets))
	for _, target := range targets {
		if target.NodeUUID == "" || target.NodeID == 0 || target.Content == "" {
			continue
		}
		key := target.NodeUUID + "\x00" + target.Content + "\x00" + target.Tags + "\x00" + string(target.Level)
		if seen[key] {
			continue
		}
		seen[key] = true
		filtered = append(filtered, target)
	}
	return filtered
}

func logMemoryBatchFailure(prefix string, err error) {
	if err != nil {
		log.Printf("%s: %v", prefix, err)
	}
}
