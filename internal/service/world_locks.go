package service

import "sync"

// worldLocks 提供世界粒度的互斥锁，用于世界 fork/snapshot 等需要锁定世界的操作。
var (
	worldLocks sync.Map
)

// LockWorld 获取指定世界的写锁。同一时刻只有一个 goroutine 能持有该锁。
// 锁在调用方主动释放前保持有效。
func LockWorld(worldID string) {
	mu := &sync.Mutex{}
	if actual, loaded := worldLocks.LoadOrStore(worldID, mu); loaded {
		mu = actual.(*sync.Mutex)
	}
	mu.Lock()
}

// UnlockWorld 释放指定世界的写锁。
func UnlockWorld(worldID string) {
	if val, ok := worldLocks.Load(worldID); ok {
		val.(*sync.Mutex).Unlock()
	}
}
