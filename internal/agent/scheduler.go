package agent

import (
	"sync/atomic"
	"context"
	"log"
	"time"

	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/engine"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/service"
	"github.com/ZSLTChenXiYin/GameAgentEngine/internal/store"
)

// Scheduler periodically runs scheduled autonomous behavior for enabled nodes.
type Scheduler struct {
	started  atomic.Bool
	pipeline *engine.Pipeline
	interval time.Duration
	limit    int
}

// NewScheduler creates a background autonomous scheduler.
func NewScheduler(p *engine.Pipeline, interval time.Duration, limit int) *Scheduler {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	return &Scheduler{pipeline: p, interval: interval, limit: limit}
}

// Start runs the scheduler until ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	s.started.Store(true)
	s.runOnce()
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	s.started.Store(false)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runOnce()
		}
	}
}

func (s *Scheduler) runOnce() {
	worlds, err := store.GetWorlds()
	if err != nil {
		log.Printf("[autonomous:scheduler] load worlds: %v", err)
		return
	}
	limit := s.limit
	var wg sync.WaitGroup
	for _, w := range worlds {
		wg.Add(1)
		go func(world store.WorldModel) {
			defer wg.Done()
			runs := service.RunScheduledAutonomous(s.pipeline, world.UUID, &limit, time.Now())
			if len(runs) > 0 {
				log.Printf("[autonomous:scheduler] world=%s runs=%d", world.UUID, len(runs))
			}
		}(w)
	}
	wg.Wait()
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	s.started.Store(false)
}
