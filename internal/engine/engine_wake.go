package engine

import (
	"log"
	"sync"
	"time"
)

const (
	maxWakeQueueSize = 1000
	wakeEventTTL     = 10 * time.Minute
)

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
		if now.Sub(e.Timestamp) <= wakeEventTTL {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// PendingWakeEventCount returns the number of pending wake events for a world.
func PendingWakeEventCount(worldID string) int {
	wakeMu.Lock()
	defer wakeMu.Unlock()
	return len(wakeQueue[worldID])
}
