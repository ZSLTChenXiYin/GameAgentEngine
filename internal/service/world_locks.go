package service

import (
	"sync"
	"sync/atomic"
)

type worldLockEntry struct {
	mu   sync.Mutex
	refs int
}

// worldLocks 提供世界粒度的互斥锁，用于世界 tick、自主执行、fork/snapshot 等重操作。
var (
	worldLocksMu      sync.Mutex
	worldLocks        = map[string]*worldLockEntry{}
	worldLocksEnabled atomic.Bool
	worldLockStats    struct {
		acquires      atomic.Uint64
		contended     atomic.Uint64
		activeHolders atomic.Int64
		maxActive     atomic.Int64
		trackedWorlds atomic.Int64
	}
)

func init() {
	worldLocksEnabled.Store(true)
}

type WorldLockStats struct {
	Acquires      uint64 `json:"acquires"`
	Contended     uint64 `json:"contended"`
	ActiveHolders int64  `json:"active_holders"`
	MaxActive     int64  `json:"max_active"`
	TrackedWorlds int64  `json:"tracked_worlds"`
}

// LockWorld 获取指定世界的写锁。同一时刻只有一个 goroutine 能持有该锁。
// 锁在调用方主动释放前保持有效。
func LockWorld(worldID string) {
	if !worldLocksEnabled.Load() {
		return
	}
	if worldID == "" {
		return
	}
	worldLocksMu.Lock()
	entry := worldLocks[worldID]
	if entry == nil {
		entry = &worldLockEntry{}
		worldLocks[worldID] = entry
		worldLockStats.trackedWorlds.Store(int64(len(worldLocks)))
	}
	if entry.refs > 0 {
		worldLockStats.contended.Add(1)
	}
	entry.refs++
	worldLocksMu.Unlock()
	entry.mu.Lock()
	worldLockStats.acquires.Add(1)
	active := worldLockStats.activeHolders.Add(1)
	for {
		maxActive := worldLockStats.maxActive.Load()
		if active <= maxActive || worldLockStats.maxActive.CompareAndSwap(maxActive, active) {
			break
		}
	}
}

// UnlockWorld 释放指定世界的写锁。
func UnlockWorld(worldID string) {
	if !worldLocksEnabled.Load() {
		return
	}
	if worldID == "" {
		return
	}
	worldLocksMu.Lock()
	entry := worldLocks[worldID]
	worldLocksMu.Unlock()
	if entry == nil {
		return
	}
	entry.mu.Unlock()
	worldLockStats.activeHolders.Add(-1)
	worldLocksMu.Lock()
	entry.refs--
	if entry.refs <= 0 && worldLocks[worldID] == entry {
		delete(worldLocks, worldID)
		worldLockStats.trackedWorlds.Store(int64(len(worldLocks)))
	}
	worldLocksMu.Unlock()
}

func ConfigureWorldLocks(enabled bool) {
	worldLocksEnabled.Store(enabled)
}

func GetWorldLockStats() WorldLockStats {
	return WorldLockStats{
		Acquires:      worldLockStats.acquires.Load(),
		Contended:     worldLockStats.contended.Load(),
		ActiveHolders: worldLockStats.activeHolders.Load(),
		MaxActive:     worldLockStats.maxActive.Load(),
		TrackedWorlds: worldLockStats.trackedWorlds.Load(),
	}
}

func withWorldLock(worldID string, fn func() error) error {
	LockWorld(worldID)
	defer UnlockWorld(worldID)
	return fn()
}

func withWorldLockValue[T any](worldID string, fn func() (T, error)) (T, error) {
	LockWorld(worldID)
	defer UnlockWorld(worldID)
	return fn()
}
