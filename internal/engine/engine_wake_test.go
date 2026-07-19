package engine

import (
	"testing"
	"time"
)

func TestEnqueueAndConsumeWakeEvents(t *testing.T) {
	worldID := "test_world_1"
	
	EnqueueWake(worldID, "node_1", "test_reason")
	EnqueueWake(worldID, "node_2", "another_reason")
	
	events := ConsumeWakeEvents(worldID)
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].NodeID != "node_1" || events[1].NodeID != "node_2" {
		t.Errorf("unexpected event order: %+v", events)
	}
	
	// After consume, should be empty
	remaining := ConsumeWakeEvents(worldID)
	if len(remaining) != 0 {
		t.Errorf("expected 0 remaining events, got %d", len(remaining))
	}
}

func TestWakeEventMultiWorldIsolation(t *testing.T) {
	EnqueueWake("world_a", "node_a1", "")
	EnqueueWake("world_b", "node_b1", "")
	
	eventsA := ConsumeWakeEvents("world_a")
	eventsB := ConsumeWakeEvents("world_b")
	
	if len(eventsA) != 1 || eventsA[0].NodeID != "node_a1" {
		t.Errorf("world_a: expected 1 event, got %d", len(eventsA))
	}
	if len(eventsB) != 1 || eventsB[0].NodeID != "node_b1" {
		t.Errorf("world_b: expected 1 event, got %d", len(eventsB))
	}
}

func TestPendingWakeEventCount(t *testing.T) {
	worldID := "test_count"
	if c := PendingWakeEventCount(worldID); c != 0 {
		t.Errorf("expected 0, got %d", c)
	}
	
	EnqueueWake(worldID, "n1", "")
	EnqueueWake(worldID, "n2", "")
	
	if c := PendingWakeEventCount(worldID); c != 2 {
		t.Errorf("expected 2, got %d", c)
	}
	
	ConsumeWakeEvents(worldID)
	if c := PendingWakeEventCount(worldID); c != 0 {
		t.Errorf("expected 0 after consume, got %d", c)
	}
}

func TestMaxWakeQueueSize(t *testing.T) {
	worldID := "test_max"
	size := maxWakeQueueSize
	
	for i := 0; i < size+10; i++ {
		EnqueueWake(worldID, "node_"+string(rune('0'+i%10)), "")
	}
	
	events := ConsumeWakeEvents(worldID)
	if len(events) > size {
		t.Errorf("events exceeded max size: %d > %d", len(events), size)
	}
	if len(events) < size-10 {
		t.Errorf("too few events: %d, expected near %d", len(events), size)
	}
}

func TestWakeEventTTL(t *testing.T) {
	worldID := "test_ttl"
	savedTTL := WakeEventTTL
	WakeEventTTL = 1 * time.Nanosecond
	defer func() { WakeEventTTL = savedTTL }()
	
	EnqueueWake(worldID, "n1", "")
	time.Sleep(10 * time.Millisecond)
	
	events := ConsumeWakeEvents(worldID)
	if len(events) != 0 {
		t.Errorf("expected 0 expired events, got %d", len(events))
	}
}

func TestWakeEventConcurrentAccess(t *testing.T) {
	worldID := "test_concurrent"
	done := make(chan struct{})
	
	go func() {
		for i := 0; i < 50; i++ {
			EnqueueWake(worldID, "cn", "")
		}
		done <- struct{}{}
	}()
	
	go func() {
		for i := 0; i < 50; i++ {
			EnqueueWake(worldID+ "_2", "cn", "")
		}
		done <- struct{}{}
	}()
	
	go func() {
		for i := 0; i < 20; i++ {
			ConsumeWakeEvents(worldID)
			PendingWakeEventCount(worldID)
		}
		done <- struct{}{}
	}()
	
	<-done
	<-done
	<-done
	
	// Should not deadlock or panic
	ConsumeWakeEvents(worldID)
	ConsumeWakeEvents(worldID + "_2")
}
