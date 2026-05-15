// Package quota provides per-key job processing limits over a rolling window.
// It allows callers to cap how many jobs a given key (e.g. tenant, user) may
// process within a configurable period.
package quota

import (
	"sync"
	"time"
)

// Clock allows time to be injected for testing.
type Clock func() time.Time

// entry tracks usage for a single key.
type entry struct {
	count    int
	windowAt time.Time
}

// Quota enforces per-key limits over a rolling window.
type Quota struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	clock    Clock
	entries  map[string]*entry
}

// Option is a functional option for Quota.
type Option func(*Quota)

// WithClock overrides the clock used by Quota (useful in tests).
func WithClock(c Clock) Option {
	return func(q *Quota) { q.clock = c }
}

// New creates a Quota that allows at most limit jobs per key per window.
func New(limit int, window time.Duration, opts ...Option) *Quota {
	q := &Quota{
		limit:   limit,
		window:  window,
		clock:   time.Now,
		entries: make(map[string]*entry),
	}
	for _, o := range opts {
		o(q)
	}
	return q
}

// Allow reports whether the given key may process another job.
// It increments the counter if allowed.
func (q *Quota) Allow(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clock()
	e, ok := q.entries[key]
	if !ok || now.After(e.windowAt.Add(q.window)) {
		q.entries[key] = &entry{count: 1, windowAt: now}
		return true
	}
	if e.count >= q.limit {
		return false
	}
	e.count++
	return true
}

// Remaining returns how many more jobs the key may process in the current window.
func (q *Quota) Remaining(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clock()
	e, ok := q.entries[key]
	if !ok || now.After(e.windowAt.Add(q.window)) {
		return q.limit
	}
	r := q.limit - e.count
	if r < 0 {
		return 0
	}
	return r
}

// Reset clears the usage record for the given key.
func (q *Quota) Reset(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.entries, key)
}
