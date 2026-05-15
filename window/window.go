// Package window provides a sliding-window counter for rate-limiting
// or frequency tracking of jobs within a rolling time period.
package window

import (
	"sync"
	"time"
)

// Clock allows injecting a fake time source in tests.
type Clock func() time.Time

// Counter tracks how many events occurred within a sliding window.
type Counter struct {
	mu       sync.Mutex
	window   time.Duration
	clock    Clock
	buckets  []entry
}

type entry struct {
	at    time.Time
	count int
}

// Option configures a Counter.
type Option func(*Counter)

// WithClock injects a custom clock (useful in tests).
func WithClock(c Clock) Option {
	return func(w *Counter) { w.clock = c }
}

// New creates a Counter that tracks events within the given window duration.
func New(window time.Duration, opts ...Option) *Counter {
	c := &Counter{
		window: window,
		clock:  time.Now,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Add records n events at the current time.
func (c *Counter) Add(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	c.evict(now)
	c.buckets = append(c.buckets, entry{at: now, count: n})
}

// Count returns the total number of events within the current window.
func (c *Counter) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	c.evict(now)
	total := 0
	for _, b := range c.buckets {
		total += b.count
	}
	return total
}

// Reset clears all recorded events.
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buckets = c.buckets[:0]
}

// evict removes entries older than the window. Must be called with mu held.
func (c *Counter) evict(now time.Time) {
	cutoff := now.Add(-c.window)
	i := 0
	for i < len(c.buckets) && c.buckets[i].at.Before(cutoff) {
		i++
	}
	c.buckets = c.buckets[i:]
}
