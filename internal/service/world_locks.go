package service

import "sync"

type worldLockEntry struct {
	mu   sync.Mutex
	refs int
}

// worldLocks 提供世界粒度的互斥锁，用于世界 tick、自主执行、fork/snapshot 等重操作。
var (
	worldLocksMu sync.Mutex
	worldLocks   = map[string]*worldLockEntry{}
)

// LockWorld 获取指定世界的写锁。同一时刻只有一个 goroutine 能持有该锁。
// 锁在调用方主动释放前保持有效。
func LockWorld(worldID string) {
	if worldID == "" {
		return
	}
	worldLocksMu.Lock()
	entry := worldLocks[worldID]
	if entry == nil {
		entry = &worldLockEntry{}
		worldLocks[worldID] = entry
	}
	entry.refs++
	worldLocksMu.Unlock()
	entry.mu.Lock()
}

// UnlockWorld 释放指定世界的写锁。
func UnlockWorld(worldID string) {
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
	worldLocksMu.Lock()
	entry.refs--
	if entry.refs <= 0 && worldLocks[worldID] == entry {
		delete(worldLocks, worldID)
	}
	worldLocksMu.Unlock()
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
