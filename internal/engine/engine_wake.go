package engine

import (
	"sync"
	"time"
)

// WakeEvent represents an event-driven wake trigger for an autonomous node.
type WakeEvent struct {
	NodeID    string
	WorldID   string
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
// When the autonomous scheduler runs, it will prioritize waking this node.
func EnqueueWake(worldID, nodeID, reason string) {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	wakeQueue[worldID] = append(wakeQueue[worldID], WakeEvent{
		NodeID:    nodeID,
		WorldID:   worldID,
		Reason:    reason,
		Priority:  10,
		Timestamp: time.Now(),
	})
}

// EnqueueWakeWithPriority adds a wake event with a custom priority.
func EnqueueWakeWithPriority(worldID, nodeID, reason string, priority int) {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	wakeQueue[worldID] = append(wakeQueue[worldID], WakeEvent{
		NodeID:    nodeID,
		WorldID:   worldID,
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
	return events
}

// PendingWakeEventCount returns the number of pending wake events for a world.
func PendingWakeEventCount(worldID string) int {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	return len(wakeQueue[worldID])
}
