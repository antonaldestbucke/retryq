// Package scheduler provides periodic job scheduling for retryq.
// It allows jobs to be enqueued automatically at a fixed interval.
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/example/retryq/job"
	"github.com/example/retryq/queue"
)

// JobFactory is a function that produces a new Job to be enqueued on each tick.
type JobFactory func() *job.Job

// Scheduler periodically creates and enqueues jobs using a JobFactory.
type Scheduler struct {
	queue    *queue.Queue
	factory  JobFactory
	interval time.Duration
	mu       sync.Mutex
	running  bool
	cancel   context.CancelFunc
}

// New creates a new Scheduler that enqueues jobs produced by factory
// into q at the given interval.
func New(q *queue.Queue, factory JobFactory, interval time.Duration) *Scheduler {
	return &Scheduler{
		queue:    q,
		factory:  factory,
		interval: interval,
	}
}

// Start begins the scheduling loop. It is safe to call Start only once;
// subsequent calls while the scheduler is running are no-ops.
func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	s.running = true
	ctx, s.cancel = context.WithCancel(ctx)

	go s.loop(ctx)
}

// Stop halts the scheduling loop. It blocks until the loop has exited.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.cancel()
	s.running = false
}

// Running reports whether the scheduler is currently active.
func (s *Scheduler) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *Scheduler) loop(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j := s.factory()
			if j != nil {
				_ = s.queue.Enqueue(j)
			}
		case <-ctx.Done():
			return
		}
	}
}
