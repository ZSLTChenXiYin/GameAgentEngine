package engine

import (
	"log"
	"sync"
	"time"
)

const (
	maxWakeQueueSize = 1000
)

// WakeEventTTL controls how long wake events stay valid before being discarded.
var WakeEventTTL time.Duration = 10 * time.Minute

// WakeEvent represents an event-driven wake trigger for an autonomous node.
type WakeEvent struct {
	NodeID    string
	Reason    string
	Priority  int
	Timestamp time.Time
}

// wakeQueue holds pending wake events, keyed by world ID.
var (
	wakeQueue   = make(map[string][]WakeEvent)
	wakeMu      sync.Mutex
)

// EnqueueWake adds a wake event for the given node in the specified world.
func EnqueueWake(worldID, nodeID, reason string) {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	queue := wakeQueue[worldID]
	if len(queue) >= maxWakeQueueSize {
		log.Printf("[warn][wake] queue full for world %s, dropping wake for %s", worldID, nodeID)
		return
	}
	wakeQueue[worldID] = append(queue, WakeEvent{
		NodeID:    nodeID,
		Reason:    reason,
		Priority:  10,
		Timestamp: time.Now(),
	})
}

// EnqueueWakeWithPriority adds a wake event with a custom priority.
// EnqueueWakeWithPriority adds a wake event with a custom priority.
// This is a public API for external callers; within the engine,
// EnqueueWake is sufficient for most use cases.
func EnqueueWakeWithPriority(worldID, nodeID, reason string, priority int) {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	queue := wakeQueue[worldID]
	if len(queue) >= maxWakeQueueSize {
		log.Printf("[warn][wake] queue full for world %s, dropping wake for %s", worldID, nodeID)
		return
	}
	wakeQueue[worldID] = append(queue, WakeEvent{
		NodeID:    nodeID,
		Reason:    reason,
		Priority:  priority,
		Timestamp: time.Now(),
	})
}

// ConsumeWakeEvents returns all pending wake events for a world and clears them.
func ConsumeWakeEvents(worldID string) []WakeEvent {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	events := wakeQueue[worldID]
	delete(wakeQueue, worldID)

	// Filter out expired events
	now := time.Now()
	filtered := events[:0]
	for _, e := range events {
		if now.Sub(e.Timestamp) <= WakeEventTTL {
			filtered = append(filtered, e)
		}
	}
	if filtered == nil {
		return make([]WakeEvent, 0)
	}
	return filtered
}

// PendingWakeEventCount returns the number of pending wake events for a world.
// PendingWakeEventCount returns the number of pending wake events for a world.
// This is a public API for external callers to observe queue depth.
func PendingWakeEventCount(worldID string) int {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	return len(wakeQueue[worldID])
}
