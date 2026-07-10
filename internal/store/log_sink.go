package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"
)

type LogSinkOptions struct {
	Enabled       bool
	BatchSize     int
	FlushInterval time.Duration
	QueueSize     int
}

var (
	logSinkMu      sync.RWMutex
	configuredSink = LogSinkOptions{
		Enabled:       true,
		BatchSize:     32,
		FlushInterval: 750 * time.Millisecond,
		QueueSize:     1024,
	}
	activeLogSink *inferenceLogSink
	logSinkClosed bool
)

type inferenceLogSink struct {
	opts    LogSinkOptions
	queue   chan *InferenceLogModel
	stopCh  chan struct{}
	doneCh  chan struct{}
	flushCh chan chan error
	wg      sync.WaitGroup
}

func ConfigureLogSink(opts LogSinkOptions) {
	if opts.BatchSize <= 0 {
		opts.BatchSize = configuredSink.BatchSize
	}
	if opts.FlushInterval <= 0 {
		opts.FlushInterval = configuredSink.FlushInterval
	}
	if opts.QueueSize <= 0 {
		opts.QueueSize = configuredSink.QueueSize
	}
	logSinkMu.Lock()
	configuredSink = opts
	logSinkMu.Unlock()
}

func currentLogSinkOptions() LogSinkOptions {
	logSinkMu.RLock()
	defer logSinkMu.RUnlock()
	return configuredSink
}

func initLogSink() {
	logSinkMu.Lock()
	defer logSinkMu.Unlock()
	logSinkClosed = false
	if activeLogSink != nil {
		activeLogSink.shutdown()
		activeLogSink = nil
	}
	if !configuredSink.Enabled {
		return
	}
	activeLogSink = newInferenceLogSink(configuredSink)
}

func CloseLogSink() error {
	logSinkMu.Lock()
	defer logSinkMu.Unlock()
	logSinkClosed = true
	if activeLogSink == nil {
		return nil
	}
	err := activeLogSink.shutdown()
	activeLogSink = nil
	return err
}

func FlushLogSink() error {
	logSinkMu.RLock()
	sink := activeLogSink
	logSinkMu.RUnlock()
	if sink == nil {
		return nil
	}
	return sink.flushPending()
}

func enqueueInferenceLog(model *InferenceLogModel) error {
	logSinkMu.RLock()
	sink := activeLogSink
	closed := logSinkClosed
	logSinkMu.RUnlock()
	if sink == nil || closed {
		recordLogDirectWrite()
		return persistInferenceLog(model)
	}
	recordLogEnqueue()
	return sink.enqueue(model)
}

func persistInferenceLog(model *InferenceLogModel) error {
	if model == nil {
		return nil
	}
	return Write(func(db *gorm.DB) error {
		return db.Create(model).Error
	})
}

func newInferenceLogSink(opts LogSinkOptions) *inferenceLogSink {
	s := &inferenceLogSink{
		opts:    opts,
		queue:   make(chan *InferenceLogModel, opts.QueueSize),
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
		flushCh: make(chan chan error),
	}
	s.wg.Add(1)
	go s.run()
	return s
}

func (s *inferenceLogSink) enqueue(model *InferenceLogModel) error {
	if model == nil {
		return nil
	}
	select {
	case s.queue <- model:
		return nil
	default:
		recordLogFallbackWrite()
		return persistInferenceLog(model)
	}
}

func (s *inferenceLogSink) run() {
	defer s.wg.Done()
	defer close(s.doneCh)

	ticker := time.NewTicker(s.opts.FlushInterval)
	defer ticker.Stop()

	batch := make([]*InferenceLogModel, 0, s.opts.BatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = persistInferenceLogBatch(batch)
		batch = batch[:0]
	}

	for {
		select {
		case model := <-s.queue:
			if model != nil {
				batch = append(batch, model)
				if len(batch) >= s.opts.BatchSize {
					flush()
				}
			}
		case <-ticker.C:
			flush()
		case ack := <-s.flushCh:
			for {
				select {
				case model := <-s.queue:
					if model != nil {
						batch = append(batch, model)
					}
				default:
					ack <- persistInferenceLogBatch(batch)
					batch = batch[:0]
					close(ack)
					goto nextLoop
				}
			}
		case <-s.stopCh:
			for {
				select {
				case model := <-s.queue:
					if model != nil {
						batch = append(batch, model)
					}
				default:
					flush()
					return
				}
			}
		}
	nextLoop:
	}
}

func (s *inferenceLogSink) flushPending() error {
	ack := make(chan error, 1)
	s.flushCh <- ack
	if err, ok := <-ack; ok {
		return err
	}
	return nil
}

func (s *inferenceLogSink) shutdown() error {
	select {
	case <-s.doneCh:
		return nil
	default:
	}
	close(s.stopCh)
	s.wg.Wait()
	return nil
}

func persistInferenceLogBatch(batch []*InferenceLogModel) error {
	if len(batch) == 0 {
		return nil
	}
	rows := make([]InferenceLogModel, 0, len(batch))
	for _, item := range batch {
		if item == nil {
			continue
		}
		rows = append(rows, *item)
	}
	if len(rows) == 0 {
		return nil
	}
	err := withWriteRetry("log_batch", func() error {
		return Writer().Transaction(func(tx *gorm.DB) error {
			return tx.Create(&rows).Error
		})
	})
	if err == nil {
		recordLogBatchFlush(len(rows))
		return nil
	}
	recordLogFlushFailure(err)
	var result error
	for i := range rows {
		if singleErr := Write(func(db *gorm.DB) error {
			return db.Create(&rows[i]).Error
		}); singleErr != nil {
			recordLogFlushFailure(singleErr)
			result = errors.Join(result, fmt.Errorf("persist log %s: %w", rows[i].UUID, singleErr))
		}
	}
	return result
}
