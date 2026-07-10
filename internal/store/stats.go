package store

import (
	"sync/atomic"
	"time"
)

type WriteRetryStats struct {
	Attempts   uint64 `json:"attempts"`
	Retries    uint64 `json:"retries"`
	Recoveries uint64 `json:"recoveries"`
	Failures   uint64 `json:"failures"`
}

type TransactionStats struct {
	Count           uint64 `json:"count"`
	Failures        uint64 `json:"failures"`
	TotalDurationMs uint64 `json:"total_duration_ms"`
}

type LogSinkStats struct {
	Enabled          bool   `json:"enabled"`
	QueueDepth       int    `json:"queue_depth"`
	QueueCapacity    int    `json:"queue_capacity"`
	Enqueued         uint64 `json:"enqueued"`
	DirectWrites     uint64 `json:"direct_writes"`
	BatchFlushes     uint64 `json:"batch_flushes"`
	BatchRows        uint64 `json:"batch_rows"`
	FallbackWrites   uint64 `json:"fallback_writes"`
	FlushFailures    uint64 `json:"flush_failures"`
	LastFlushUnixMs  int64  `json:"last_flush_unix_ms"`
	LastErrorUnixMs  int64  `json:"last_error_unix_ms"`
	LastErrorMessage string `json:"last_error_message,omitempty"`
}

type PipelineStats struct {
	Driver       string           `json:"driver"`
	WriteRetry   WriteRetryStats  `json:"write_retry"`
	Transactions TransactionStats `json:"transactions"`
	LogSink      LogSinkStats     `json:"log_sink"`
}

var pipelineStats struct {
	writeAttempts       atomic.Uint64
	writeRetries        atomic.Uint64
	writeRecoveries     atomic.Uint64
	writeFailures       atomic.Uint64
	transactions        atomic.Uint64
	transactionFailures atomic.Uint64
	transactionDuration atomic.Uint64
	logEnqueued         atomic.Uint64
	logDirectWrites     atomic.Uint64
	logBatchFlushes     atomic.Uint64
	logBatchRows        atomic.Uint64
	logFallbackWrites   atomic.Uint64
	logFlushFailures    atomic.Uint64
	logLastFlushUnixMs  atomic.Int64
	logLastErrorUnixMs  atomic.Int64
	logLastErrorMessage atomic.Value
}

func ResetPipelineStats() {
	pipelineStats.writeAttempts.Store(0)
	pipelineStats.writeRetries.Store(0)
	pipelineStats.writeRecoveries.Store(0)
	pipelineStats.writeFailures.Store(0)
	pipelineStats.transactions.Store(0)
	pipelineStats.transactionFailures.Store(0)
	pipelineStats.transactionDuration.Store(0)
	pipelineStats.logEnqueued.Store(0)
	pipelineStats.logDirectWrites.Store(0)
	pipelineStats.logBatchFlushes.Store(0)
	pipelineStats.logBatchRows.Store(0)
	pipelineStats.logFallbackWrites.Store(0)
	pipelineStats.logFlushFailures.Store(0)
	pipelineStats.logLastFlushUnixMs.Store(0)
	pipelineStats.logLastErrorUnixMs.Store(0)
	pipelineStats.logLastErrorMessage.Store("")
}

func GetPipelineStats() PipelineStats {
	stats := PipelineStats{
		Driver: currentDriverName(),
		WriteRetry: WriteRetryStats{
			Attempts:   pipelineStats.writeAttempts.Load(),
			Retries:    pipelineStats.writeRetries.Load(),
			Recoveries: pipelineStats.writeRecoveries.Load(),
			Failures:   pipelineStats.writeFailures.Load(),
		},
		Transactions: TransactionStats{
			Count:           pipelineStats.transactions.Load(),
			Failures:        pipelineStats.transactionFailures.Load(),
			TotalDurationMs: pipelineStats.transactionDuration.Load(),
		},
		LogSink: LogSinkStats{
			Enqueued:        pipelineStats.logEnqueued.Load(),
			DirectWrites:    pipelineStats.logDirectWrites.Load(),
			BatchFlushes:    pipelineStats.logBatchFlushes.Load(),
			BatchRows:       pipelineStats.logBatchRows.Load(),
			FallbackWrites:  pipelineStats.logFallbackWrites.Load(),
			FlushFailures:   pipelineStats.logFlushFailures.Load(),
			LastFlushUnixMs: pipelineStats.logLastFlushUnixMs.Load(),
			LastErrorUnixMs: pipelineStats.logLastErrorUnixMs.Load(),
		},
	}
	if value := pipelineStats.logLastErrorMessage.Load(); value != nil {
		stats.LogSink.LastErrorMessage, _ = value.(string)
	}
	logSinkMu.RLock()
	stats.LogSink.Enabled = activeLogSink != nil && !logSinkClosed
	if activeLogSink != nil {
		stats.LogSink.QueueDepth = len(activeLogSink.queue)
		stats.LogSink.QueueCapacity = cap(activeLogSink.queue)
		if !stats.LogSink.Enabled {
			stats.LogSink.Enabled = activeLogSink.opts.Enabled
		}
		if stats.LogSink.QueueCapacity == 0 {
			stats.LogSink.QueueCapacity = activeLogSink.opts.QueueSize
		}
	} else {
		stats.LogSink.Enabled = currentLogSinkOptions().Enabled && !logSinkClosed
		stats.LogSink.QueueCapacity = currentLogSinkOptions().QueueSize
	}
	logSinkMu.RUnlock()
	return stats
}

func recordWriteAttempt() {
	pipelineStats.writeAttempts.Add(1)
}

func recordWriteRetry() {
	pipelineStats.writeRetries.Add(1)
}

func recordWriteRecovery() {
	pipelineStats.writeRecoveries.Add(1)
}

func recordWriteFailure() {
	pipelineStats.writeFailures.Add(1)
}

func recordTransactionResult(duration time.Duration, err error) {
	pipelineStats.transactions.Add(1)
	pipelineStats.transactionDuration.Add(uint64(duration.Milliseconds()))
	if err != nil {
		pipelineStats.transactionFailures.Add(1)
	}
}

func recordLogEnqueue() {
	pipelineStats.logEnqueued.Add(1)
}

func recordLogDirectWrite() {
	pipelineStats.logDirectWrites.Add(1)
}

func recordLogBatchFlush(rows int) {
	pipelineStats.logBatchFlushes.Add(1)
	pipelineStats.logBatchRows.Add(uint64(rows))
	pipelineStats.logLastFlushUnixMs.Store(time.Now().UnixMilli())
}

func recordLogFallbackWrite() {
	pipelineStats.logFallbackWrites.Add(1)
}

func recordLogFlushFailure(err error) {
	pipelineStats.logFlushFailures.Add(1)
	pipelineStats.logLastErrorUnixMs.Store(time.Now().UnixMilli())
	if err == nil {
		pipelineStats.logLastErrorMessage.Store("")
		return
	}
	pipelineStats.logLastErrorMessage.Store(err.Error())
}
