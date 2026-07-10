package store

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	driverMySQL "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type WriteRetryOptions struct {
	Enabled     bool
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

var (
	writeRetryMu sync.RWMutex
	writeRetry   = WriteRetryOptions{
		Enabled:     true,
		MaxAttempts: 3,
		BaseDelay:   40 * time.Millisecond,
		MaxDelay:    250 * time.Millisecond,
	}
	currentDriver string
)

func ConfigureWriteRetry(opts WriteRetryOptions) {
	current := currentWriteRetryOptions()
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = current.MaxAttempts
	}
	if opts.BaseDelay < 0 {
		opts.BaseDelay = 0
	}
	if opts.MaxDelay <= 0 {
		opts.MaxDelay = current.MaxDelay
	}
	if opts.BaseDelay > opts.MaxDelay {
		opts.MaxDelay = opts.BaseDelay
	}
	writeRetryMu.Lock()
	writeRetry = opts
	writeRetryMu.Unlock()
}

func currentWriteRetryOptions() WriteRetryOptions {
	writeRetryMu.RLock()
	defer writeRetryMu.RUnlock()
	return writeRetry
}

func setCurrentDriver(driver string) {
	writeRetryMu.Lock()
	currentDriver = normalizeDriver(driver)
	writeRetryMu.Unlock()
}

func currentDriverName() string {
	writeRetryMu.RLock()
	defer writeRetryMu.RUnlock()
	return currentDriver
}

func Write(fn func(db *gorm.DB) error) error {
	return withWriteRetry("write", func() error {
		return fn(Writer())
	})
}

func withWriteRetry(operation string, fn func() error) error {
	opts := currentWriteRetryOptions()
	attempts := 1
	if opts.Enabled {
		attempts = opts.MaxAttempts
		if attempts <= 0 {
			attempts = 1
		}
	}
	driver := currentDriverName()
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		err = fn()
		if err == nil {
			if attempt > 1 {
				log.Printf("[db-retry] operation=%s driver=%s recovered_after=%d", operation, driver, attempt)
			}
			return nil
		}
		if attempt >= attempts || !isRetriableWriteError(driver, err) {
			return err
		}
		delay := retryDelay(opts, attempt)
		log.Printf("[db-retry] operation=%s driver=%s attempt=%d/%d delay_ms=%d err=%v", operation, driver, attempt, attempts, delay.Milliseconds(), err)
		if delay > 0 {
			time.Sleep(delay)
		}
	}
	return err
}

func retryDelay(opts WriteRetryOptions, attempt int) time.Duration {
	if attempt <= 0 || opts.BaseDelay <= 0 {
		return 0
	}
	delay := opts.BaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
		if opts.MaxDelay > 0 && delay >= opts.MaxDelay {
			return opts.MaxDelay
		}
	}
	if opts.MaxDelay > 0 && delay > opts.MaxDelay {
		return opts.MaxDelay
	}
	return delay
}

func normalizeDriver(driver string) string {
	return strings.ToLower(strings.TrimSpace(driver))
}

func isRetriableWriteError(driver string, err error) bool {
	if err == nil {
		return false
	}
	driver = normalizeDriver(driver)
	message := strings.ToLower(err.Error())
	switch driver {
	case "mysql":
		var mysqlErr *driverMySQL.MySQLError
		if errors.As(err, &mysqlErr) {
			switch mysqlErr.Number {
			case 1205, 1213:
				return true
			}
		}
		return strings.Contains(message, "deadlock found") || strings.Contains(message, "lock wait timeout exceeded")
	case "sqlite", "":
		return strings.Contains(message, "database is locked") ||
			strings.Contains(message, "database table is locked") ||
			strings.Contains(message, "database schema is locked") ||
			strings.Contains(message, "sqlite_busy") ||
			strings.Contains(message, "sqlite_locked")
	default:
		return false
	}
}
