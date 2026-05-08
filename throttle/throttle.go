// Package throttle provides a middleware that limits the rate at which
// jobs are processed by introducing a minimum interval between executions.
package throttle

import (
	"sync"
	"time"

	"github.com/yourorg/retryq/job"
)

// Clock abstracts time for testing.
type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}

type realClock struct{}

func (realClock) Now() time.Time          { return time.Now() }
func (realClock) Sleep(d time.Duration)   { time.Sleep(d) }

// Throttle enforces a minimum interval between successive job executions.
type Throttle struct {
	mu       sync.Mutex
	last     time.Time
	interval time.Duration
	clock    Clock
}

// Option configures a Throttle.
type Option func(*Throttle)

// WithClock replaces the default clock (useful in tests).
func WithClock(c Clock) Option {
	return func(t *Throttle) { t.clock = c }
}

// New creates a Throttle that enforces at least interval between job executions.
func New(interval time.Duration, opts ...Option) *Throttle {
	t := &Throttle{
		interval: interval,
		clock:    realClock{},
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

// Handler is the function signature for job processors.
type Handler func(j *job.Job) error

// Wrap returns a Handler that waits until the throttle interval has elapsed
// since the last execution before invoking next.
func (t *Throttle) Wrap(next Handler) Handler {
	return func(j *job.Job) error {
		t.mu.Lock()
		now := t.clock.Now()
		if !t.last.IsZero() {
			if elapsed := now.Sub(t.last); elapsed < t.interval {
				wait := t.interval - elapsed
				t.mu.Unlock()
				t.clock.Sleep(wait)
				t.mu.Lock()
			}
		}
		t.last = t.clock.Now()
		t.mu.Unlock()
		return next(j)
	}
}
